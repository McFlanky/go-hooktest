package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
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
	session     ssh.Session
	destination string
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
	session := value.(Session)
	req, err := http.NewRequest(r.Method, session.destination, r.Body)
	if err != nil {
		log.Fatal(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	defer r.Body.Close()
	io.Copy(w, resp.Body)

}

func startHTTPServer() error {
	httpPort := ":5000"
	router := http.NewServeMux()

	handler := &HTTPHandler{}
	router.HandleFunc("/{id}", handler.handleWebhook)
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
	term.Write([]byte(startMessage()))
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
		destination, err := url.Parse(input)
		if err != nil {
			log.Fatal(err)
		}
		host := destination.Host
		// path := destination.Path
		internalSession := Session{
			session:     session,
			destination: destination.String(),
		}
		clients.Store(id, internalSession)

		webhookURL := fmt.Sprintf("http://localhost:5000/%s", id)
		command := fmt.Sprintf("\nGenerated Webhook: %s\n\nCopy & Run Command:\nssh -R 127.0.0.1:%d:%s localhost -p 2222 tunnel\n\n", webhookURL, generatedPort, host)
		term.Write([]byte(command))
		return
	}
}

func randomPort() int {
	min := 49152
	max := 65535
	return min + rand.Intn(max-min+1)
}

func startMessage() string {
	return `


	██╗  ██╗ ██████╗  ██████╗ ██╗  ██╗████████╗███████╗███████╗████████╗
	██║  ██║██╔═══██╗██╔═══██╗██║ ██╔╝╚══██╔══╝██╔════╝██╔════╝╚══██╔══╝
	███████║██║   ██║██║   ██║█████╔╝    ██║   █████╗  ███████╗   ██║   
	██╔══██║██║   ██║██║   ██║██╔═██╗    ██║   ██╔══╝  ╚════██║   ██║   
	██║  ██║╚██████╔╝╚██████╔╝██║  ██╗   ██║   ███████╗███████║   ██║
	╚═╝  ╚═╝ ╚═════╝  ╚═════╝ ╚═╝  ╚═╝   ╚═╝   ╚══════╝╚══════╝   ╚═╝         
	
	

Commands:
   setup    - Setup a new webhook
   tunnel   - Get tunnel config for existing webhook
   list     - List all webhooks
   active   - List all active ssh tunnels
   help	    - Show this message
`
}
