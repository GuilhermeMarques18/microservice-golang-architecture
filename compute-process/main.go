package main

import (
	"context"
	"crypto/sha256"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	pb "github.com/GuilhermeMarques18/microservice-golang-architecture/gen/go/process/v1"
	_ "go.uber.org/automaxprocs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
)

const maxComplexity = 100_000

type server struct {
	pb.UnimplementedComputerServiceServer
	workerPool chan struct{}
	logger     *slog.Logger
}

func NewServer(logger *slog.Logger) *server {
	n := runtime.GOMAXPROCS(0)
	return &server{
		workerPool: make(chan struct{}, n),
		logger:     logger,
	}
}

func performChainedHash(ctx context.Context, payload []byte, complexity int32) ([]byte, error) {
	hash := sha256.Sum256(payload)
	currentHash := hash[:]
	for i := int32(1); i < complexity; i++ {
		if i%1000 == 0 {
			select {
			case <-ctx.Done():
				return nil, status.Error(codes.Canceled, "Processamento interrompido")
			default:
			}
		}
		nextHash := sha256.Sum256(currentHash)
		currentHash = nextHash[:]
	}
	return currentHash, nil
}

func (s *server) ProcessData(ctx context.Context, req *pb.ProcessDataRequest) (*pb.ProcessDataResponse, error) {
	if req.Complexity <= 0 || req.Complexity > maxComplexity {
		return nil, status.Errorf(codes.InvalidArgument, "Complexity deve estar entre 1 e %d", maxComplexity)
	}
	if len(req.Payload) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Payload não pode ser vazio")
	}

	select {
	case s.workerPool <- struct{}{}:
		defer func() { <-s.workerPool }()
	case <-ctx.Done():
		return nil, status.Error(codes.Canceled, "Requisição cancelada pelo cliente")
	default:
		s.logger.Warn("Carga rejeitada: pool de workers saturado", "request_id", req.RequestId)
		return nil, status.Error(codes.ResourceExhausted, "Limite de CPU atingido")
	}

	startTime := time.Now()
	result, err := performChainedHash(ctx, req.Payload, req.Complexity)
	if err != nil {
		s.logger.Error("Falha no processamento", "request_id", req.RequestId, "error", err)
		return nil, err
	}
	processingTime := time.Since(startTime).Milliseconds()

	s.logger.Info("Processamento concluído",
		"Request_id", req.RequestId,
		"Complexity", req.Complexity,
		"Processing_time_ms", processingTime,
	)

	return &pb.ProcessDataResponse{
		RequestId:        req.RequestId,
		ResultHash:       result,
		ProcessingTimeMs: processingTime,
	}, nil
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	port := ":50051"
	lis, err := net.Listen("tcp", port)
	if err != nil {
		logger.Error("Falha ao escutar porta", "Port", port, "Error", err)
		os.Exit(1)
	}

	grpcServer := grpc.NewServer(
		grpc.MaxRecvMsgSize(4 << 20),
	)

	computeSrv := NewServer(logger)
	pb.RegisterComputerServiceServer(grpcServer, computeSrv)

	healthSrv := health.NewServer()
	healthpb.RegisterHealthServer(grpcServer, healthSrv)
	healthSrv.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	go func() {
		logger.Info("Compute-processor rodando", "port", port, "gomaxprocs", runtime.GOMAXPROCS(0))
		if err := grpcServer.Serve(lis); err != nil {
			logger.Error("Falha ao iniciar servidor gRPC", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	logger.Info("Sinal de encerramento recebido, iniciando graceful shutdown")

	healthSrv.SetServingStatus("", healthpb.HealthCheckResponse_NOT_SERVING)

	grpcServer.GracefulStop()
	logger.Info("Servidor gRPC encerrado")
}
