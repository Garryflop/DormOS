package main

import (
	"context"
	"fmt"
	"log"
	"net"

	filev1 "github.com/Garryflop/DormOS-gen-go/file/v1"
	"github.com/Garryflop/DormManage/file-service/internal/config"
	"github.com/Garryflop/DormManage/file-service/internal/handler"
	"github.com/Garryflop/DormManage/file-service/internal/repository"
	"github.com/Garryflop/DormManage/file-service/internal/storage"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	cfg := config.Load()

	// ── Postgres ───────────────────────────────────────────────
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

	// ── MinIO ──────────────────────────────────────────────────
	minioStorage, err := storage.NewMinioStorage(
		cfg.MinioEndpoint,
		cfg.MinioAccessKey,
		cfg.MinioSecretKey,
		cfg.MinioBucket,
		cfg.MinioUseSSL,
	)
	if err != nil {
		log.Fatalf("Failed to connect to MinIO: %v", err)
	}
	log.Println("✓ Connected to MinIO")

	// ── Wire up layers ─────────────────────────────────────────
	repo := repository.NewFileRepository(db)
	_ = rdb

	fileServer := handler.NewFileGRPCServer(repo, minioStorage)

	// ── gRPC Server ────────────────────────────────────────────
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.GRPCPort))
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", cfg.GRPCPort, err)
	}

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(loggingInterceptor),
	)

	filev1.RegisterFileServiceServer(grpcServer, fileServer)
	reflection.Register(grpcServer)

	log.Printf("✓ File Service gRPC listening on :%s", cfg.GRPCPort)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

func loggingInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	log.Printf("[file-service] → %s", info.FullMethod)
	return handler(ctx, req)
}
