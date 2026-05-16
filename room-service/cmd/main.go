package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/Garryflop/DormManage/room-service/internal/config"
	"github.com/Garryflop/DormManage/room-service/internal/repository/postgres"
	transport "github.com/Garryflop/DormManage/room-service/internal/transport/grpc"
	"github.com/Garryflop/DormManage/room-service/internal/usecase"
	roomv1 "github.com/Garryflop/DormOS-gen-go/room/v1"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	cfg := config.Load()

	// ── OpenTelemetry ──────────────────────────────────────────
	otelEndpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if otelEndpoint == "" {
		otelEndpoint = "otel-collector:4317"
	}
	shutdown, err := config.InitOpenTelemetry(context.Background(), "room-service", otelEndpoint)
	if err != nil {
		log.Printf("⚠ Failed to init OpenTelemetry: %v", err)
	} else {
		defer shutdown(context.Background())
	}

	// ── Postgres ──────────────────────────────────────────────
	db, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to Postgres: %v", err)
	}
	defer db.Close()

	if err := db.Ping(context.Background()); err != nil {
		log.Fatalf("Postgres ping failed: %v", err)
	}
	log.Println("✓ Connected to Postgres")

	// ── Redis ──────────────────────────────────────────────────
	opts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		log.Fatalf("Failed to parse Redis URL: %v", err)
	}
	rdb := redis.NewClient(opts)
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("Redis ping failed: %v", err)
	}
	log.Println("✓ Connected to Redis")

	// ── NATS ───────────────────────────────────────────────────
	nc, err := nats.Connect(cfg.NatsURL)
	if err != nil {
		log.Printf("⚠ NATS connection failed (events disabled): %v", err)
	} else {
		log.Println("✓ Connected to NATS")
		defer nc.Close()
	}

	// ── Clean Architecture Wiring ──────────────────────────────
	repo := postgres.NewRoomRepository(db)
	uc := usecase.NewRoomUseCase(repo, rdb, nc)
	h := transport.NewRoomHandler(uc)

	// ── gRPC Server ────────────────────────────────────────────
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.GRPCPort))
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", cfg.GRPCPort, err)
	}

	grpcServer := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(loggingInterceptor),
	)

	// Register generated RoomService server
	roomv1.RegisterRoomServiceServer(grpcServer, h)

	// Enable gRPC server reflection
	reflection.Register(grpcServer)

	log.Printf("✓ Room Service gRPC listening on :%s", cfg.GRPCPort)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

func loggingInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	log.Printf("[room-service] → %s", info.FullMethod)
	resp, err := handler(ctx, req)
	if err != nil {
		log.Printf("[room-service] ✗ %s: %v", info.FullMethod, err)
	}
	return resp, err
}
