package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
)

var (
	receivedPayloads []map[string]interface{}
	payloadsMutex    sync.Mutex
)

func webhookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var payload map[string]interface{}
	err = json.Unmarshal(body, &payload)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	payloadsMutex.Lock()
	receivedPayloads = append(receivedPayloads, payload)
	payloadsMutex.Unlock()

	fmt.Fprintf(w, `{"status": "received", "count": %d}`, len(receivedPayloads))
	fmt.Printf("✓ Webhook recebido! Total: %d\n", len(receivedPayloads))
	fmt.Printf("  Payload: %s\n", string(body))
}

func main() {
	http.HandleFunc("/webhook", webhookHandler)
	fmt.Println("Servidor de webhook rodando em http://localhost:8888/webhook")
	if err := http.ListenAndServe(":8888", nil); err != nil {
		fmt.Printf("Erro ao iniciar o servidor: %v\n", err)
	}
}
