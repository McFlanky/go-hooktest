package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gliderlabs/ssh"
	"github.com/teris-io/shortid"
	gossh "golang.org/x/crypto/ssh"
	"golang.org/x/term"

	"math/rand"
)

type Session struct {
	session ssh.Session
	port    int
}

var clients sync.Map

type HTTPHandler struct {
}

func (h *HTTPHandler) handleWebhook(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	value, ok := clients.Load(id)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("client id not found"))
		return
	}
	fmt.Println("this is the id", id)
	session := value.(Session)
	dest := fmt.Sprintf("http://localhost:%d", session.port)
	_, err := http.Post(dest, "application/json", r.Body)
	if err != nil {
		log.Fatal(err)
	}
	defer r.Body.Close()
}

func startHTTPServer() error {
	httpPort := ":5000"
	router := http.NewServeMux()

	handler := &HTTPHandler{}
	router.HandleFunc("/{id}/*", handler.handleWebhook)
	return http.ListenAndServe(httpPort, router)
}

func startSSHServer() error {
	sshPort := ":2222"
	handler := NewSSHHandler()

	fwHandler := &ssh.ForwardedTCPHandler{}
	server := ssh.Server{
		Addr:    sshPort,
		Handler: handler.handleSSHSession,
		ServerConfigCallback: func(ctx ssh.Context) *gossh.ServerConfig {
			cfg := &gossh.ServerConfig{
				ServerVersion: "SSH-2.0-sendit",
			}
			cfg.Ciphers = []string{"chacha20-poly1305@openssh.com"}
			return cfg
		},
		PublicKeyHandler: func(ctx ssh.Context, key ssh.PublicKey) bool {
			return true
		},
		LocalPortForwardingCallback: ssh.LocalPortForwardingCallback(func(ctx ssh.Context, dhost string, dport uint32) bool {
			log.Println("Accepted forward", dhost, dport)
			// todo: auth validation
			return true
		}),
		ReversePortForwardingCallback: ssh.ReversePortForwardingCallback(func(ctx ssh.Context, host string, port uint32) bool {
			log.Println("attempt to bind", host, port, "granted")
			// todo: auth validation
			return true
		}),
		RequestHandlers: map[string]ssh.RequestHandler{
			"tcpip-forward":        fwHandler.HandleSSHRequest,
			"cancel-tcpip-forward": fwHandler.HandleSSHRequest,
		},
	}
	b, err := os.ReadFile("keys/privatekey.pub")
	if err != nil {
		log.Fatal(err)
	}
	privateKey, err := gossh.ParsePrivateKey(b)
	if err != nil {
		log.Fatal("Failed to parse private key: ", err)
	}
	server.AddHostKey(privateKey)
	return server.ListenAndServe()
}

func main() {
	go startSSHServer()
	startHTTPServer()
}

type SSHHandler struct {
}

func NewSSHHandler() *SSHHandler {
	return &SSHHandler{}

}

func (h *SSHHandler) handleSSHSession(session ssh.Session) {
	if session.RawCommand() == "tunnel" {
		session.Write([]byte("Tunneling traffic...\n"))
		<-session.Context().Done()
		return
	}

	term := term.NewTerminal(session, "$ ")
	msg := fmt.Sprintf("%s\n\nWelcome to HookTest!\n\nEnter your Webhook destination:\n", banner)
	term.Write([]byte(msg))
	for {
		input, err := term.ReadLine()
		if err != nil {
			log.Fatal(err)
		}

		if strings.Contains(input, "ssh -R") {
			for {
				time.Sleep(time.Second)
			}
		}

		generatedPort := randomPort()
		id := shortid.MustGenerate()
		internalSession := Session{
			session: session,
			port:    generatedPort,
		}
		clients.Store(id, internalSession)

		webhookURL := fmt.Sprintf("http://localhost:5000/%s", id)
		command := fmt.Sprintf("\nGenerated Webhook: %s\n\nCopy & Run Command:\nssh -R 127.0.0.1:%d:%s localhost -p 2222 tunnel\n", webhookURL, generatedPort, input)
		term.Write([]byte(command))
		return
	}
}

func randomPort() int {
	min := 49152
	max := 65535
	return min + rand.Intn(max-min+1)
}

var banner = `
██╗  ██╗ ██████╗  ██████╗ ██╗  ██╗████████╗███████╗███████╗████████╗
██║  ██║██╔═══██╗██╔═══██╗██║ ██╔╝╚══██╔══╝██╔════╝██╔════╝╚══██╔══╝
███████║██║   ██║██║   ██║█████╔╝    ██║   █████╗  ███████╗   ██║   
██╔══██║██║   ██║██║   ██║██╔═██╗    ██║   ██╔══╝  ╚════██║   ██║   
██║  ██║╚██████╔╝╚██████╔╝██║  ██╗   ██║   ███████╗███████║   ██║
╚═╝  ╚═╝ ╚═════╝  ╚═════╝ ╚═╝  ╚═╝   ╚═╝   ╚══════╝╚══════╝   ╚═╝                                                               
`
