package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/GuilhermeMarques18/microservice-golang-architecture/api-gateway/dto"
	pb "github.com/GuilhermeMarques18/microservice-golang-architecture/gen/go/process/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
)

type Gateway struct {
	client pb.ComputerServiceClient
	logger *slog.Logger
}

func (g *Gateway) HandlerProtect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método inválido", http.StatusMethodNotAllowed)
		return
	}

	const maxBytes = 2 << 20
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	var reqDTO dto.ProcessRequestDTO
	if err := dec.Decode(&reqDTO); err != nil {
		http.Error(w, "JSON inválido: "+err.Error(), http.StatusBadRequest)
		return
	}

	if err := reqDTO.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	g.logger.Info("Request recebido", "Request_id", reqDTO.RequestID, "Complexity", reqDTO.Complexity, "Payload_bytes", len(reqDTO.Payload))

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	response, err := g.client.ProcessData(ctx, &pb.ProcessDataRequest{
		RequestId:  reqDTO.RequestID,
		Complexity: reqDTO.Complexity,
		Payload:    reqDTO.Payload,
	})

	if err != nil {
		st, _ := status.FromError(err)
		switch st.Code() {
		case codes.DeadlineExceeded:
			http.Error(w, "Tempo limite de processamento excedido", http.StatusGatewayTimeout)
		default:
			http.Error(w, "Erro interno do servidor", http.StatusBadGateway)
		}
		return
	}

	b, err := protojson.Marshal(response)
	if err != nil {
		g.logger.Error("falha ao serializar resposta", "request_id", reqDTO.RequestID, "error", err)
		http.Error(w, "erro interno", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(b)

}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	processorAddr := os.Getenv("PROCESSOR_ADDR")
	if processorAddr == "" {
		processorAddr = "compute-processor:50051"
	}

	conn, err := grpc.NewClient(
		"dns:///"+processorAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(`{"loadBalancingConfig": [{"round_robin":{}}]}`),
	)

	if err != nil {
		logger.Error("falha ao conectar no processor", "error", err)
		os.Exit(1)
	}
	defer conn.Close()

	gw := &Gateway{
		client: pb.NewComputerServiceClient(conn),
		logger: logger,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/process", gw.HandlerProtect)
	mux.HandleFunc("/health", healthHandler)

	s := &http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadTimeout:       5 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       100 * time.Second,
	}

	go func() {
		logger.Info("servidor iniciado", "port", 8080)
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("erro ao iniciar servidor", "error", err)
			os.Exit(1)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()
	<-ctx.Done()

	logger.Info("sinal de shutdown recebido, encerrando graciosamente...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.Shutdown(shutdownCtx); err != nil {
		logger.Error("erro no shutdown", "error", err)
	}
}
