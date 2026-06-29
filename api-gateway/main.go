package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type ProcessRequestDTO struct {
	RequestID  string `json:"request_id"`
	Payload    []byte `json:"payload"`
	Complexity int32  `json:"complexity"`
}

func HandlerProtect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método inválido", http.StatusMethodNotAllowed)
		return
	}

	const maxBytes = 2 << 20
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	var dto ProcessRequestDTO
	if err := dec.Decode(&dto); err != nil {
		http.Error(w, "JSON inválido: "+err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Printf(" RequestID: %s | Complexidade: %d | Payload(bytes): %d\n",
		dto.RequestID, dto.Complexity, len(dto.Payload))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)

}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/process", HandlerProtect)

	s := &http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadTimeout:       5 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       100 * time.Second,
	}

	fmt.Println("Servidor iniciado na porta 8080")
	if err := s.ListenAndServe(); err != nil {
		fmt.Printf("Erro ao iniciar o servidor: %v\n", err)
	}

}
