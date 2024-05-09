package main

import (
	"fmt"
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
	router.HandleFunc("/{id}/*", handler.handleWebhook)
	return http.ListenAndServe(httpPort, router)
}

func startSSHServer() error {
	sshPort := ":2222"
	handler := NewSSHHandler()
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
	return server.ListenAndServe()
}

func main() {
	go startSSHServer()
	startHTTPServer()
}

type SSHHandler struct {
	channels map[string]chan string
}

func NewSSHHandler() *SSHHandler {
	return &SSHHandler{
		channels: make(map[string]chan string),
	}
}

func (h *SSHHandler) handleSSHSession(session ssh.Session) {
	cmd := session.RawCommand()
	if cmd == "init" {
		id := shortid.MustGenerate()
		fmt.Println("new init id channel", id)
		webhookURL := "http://localhost:5000/" + id + "\n"
		resp := fmt.Sprintf("webhook url %s\nssh localhost -p 2222 -i /home/coletj/workspace/github.com/McFlanky/ssh-webhook-app/keys/privatekey.pub %s | curl -X POST -d @- http://localhost:3000/payment/webhook \n", webhookURL, id)
		//resp := fmt.Sprintf(`%s ssh localhost -p 2222 -i /home/coletj/workspace/github.com/McFlanky/ssh-webhook-app/keys/privatekey.pub %s | while IFS=read -r line; do echo "$line" | curl -X POST -d @- http://localhost:3000/payment/webhook; done`, webhookURL, id)
		session.Write([]byte(resp))
		respCh := make(chan string, 1)
		h.channels[id] = respCh
		clients.Store(id, respCh)
	}
	if len(cmd) > 0 && cmd != "init" {
		respCh, ok := h.channels[cmd]
		if !ok {
			session.Write([]byte("invalid webhook id\n"))
			return
		}
		for data := range respCh {
			session.Write([]byte(data + "\n"))
		}
	}
}
