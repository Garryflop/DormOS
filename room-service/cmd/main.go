package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/Garryflop/DormManage/room-service/internal/config"
	"github.com/Garryflop/DormManage/room-service/internal/handler"
	"github.com/Garryflop/DormManage/room-service/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	cfg := config.Load()

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
		// Non-fatal: log warning and continue without NATS
		log.Printf("⚠ NATS connection failed (events disabled): %v", err)
	} else {
		log.Println("✓ Connected to NATS")
		defer nc.Close()
	}

	// ── Wire up layers ─────────────────────────────────────────
	repo := repository.NewRoomRepository(db)
	h := handler.NewRoomHandler(repo, rdb, nc)

	// ── gRPC Server ────────────────────────────────────────────
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.GRPCPort))
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", cfg.GRPCPort, err)
	}

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(loggingInterceptor),
	)

	// TODO: Register generated RoomService server here once DormOS-gen-go is imported
	// roomv1.RegisterRoomServiceServer(grpcServer, h)
	_ = h // suppress unused warning until generated code is available

	// Enable gRPC server reflection (useful for grpcurl testing)
	reflection.Register(grpcServer)

	log.Printf("✓ Room Service gRPC listening on :%s", cfg.GRPCPort)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}

	_ = os.Getenv // ensure os is used
}

func loggingInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	log.Printf("[room-service] → %s", info.FullMethod)
	resp, err := handler(ctx, req)
	if err != nil {
		log.Printf("[room-service] ✗ %s: %v", info.FullMethod, err)
	}
	return resp, err
}
