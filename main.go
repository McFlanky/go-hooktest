package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/gliderlabs/ssh"
	"github.com/teris-io/shortid"
	gossh "golang.org/x/crypto/ssh"
)

var clients sync.Map

type HTTPHandler struct {
}

func (h *HTTPHandler) handleWebhook(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	ch, ok := clients.Load(id)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("client id not found"))
		return
	}
	b, err := io.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}
	defer r.Body.Close()
	ch.(chan string) <- string(b)
}

func startHTTPServer() error {
	httpPort := ":5000"
	router := http.NewServeMux()

	handler := &HTTPHandler{}
	router.HandleFunc("/{id}/", handler.handleWebhook)
	return http.ListenAndServe(httpPort, router)
}

func startSSHServer() error {
	sshPort := ":2222"

	handler := &SSHHandler{}
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
	log.Fatal(server.ListenAndServe())
	return server.ListenAndServe()
}

func main() {
	go startSSHServer()
	startHTTPServer()
}

type SSHHandler struct {
}

func (h *SSHHandler) handleSSHSession(session ssh.Session) {
	id := shortid.MustGenerate()
	webhookURL := "http://webhooker.com/" + id
	session.Write([]byte(webhookURL))

	respCh := make(chan string)
	clients.Store(id, respCh)

	for data := range respCh {
		session.Write([]byte(data + "\n"))
	}
}
