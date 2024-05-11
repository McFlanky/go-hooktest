package main

import (
	"fmt"
	"io"
	"net/http"
)

func main() {
	router := http.NewServeMux()

	router.HandleFunc("POST /payment/webhook", handlePaymentWebhook)

	http.ListenAndServe(":3000", router)
}

type WebhookRequest struct {
	Amount  int    `json:"amount"`
	Message string `json:"message"`
}

func handlePaymentWebhook(w http.ResponseWriter, r *http.Request) {
	b, _ := io.ReadAll(r.Body)
	fmt.Println(string(b))
	// fmt.Println(r.Header.Get("Content-Type"))
	//
	//	var req WebhookRequest
	//	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
	//		log.Fatal(err)
	//	}
	//	fmt.Println("we got our webhook data!", req)
}
