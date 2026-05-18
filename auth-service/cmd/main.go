package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/Garryflop/DormManage/auth-service/internal/config"
	"github.com/Garryflop/DormManage/auth-service/internal/repository/postgres"
	transport "github.com/Garryflop/DormManage/auth-service/internal/transport/grpc"
	"github.com/Garryflop/DormManage/auth-service/internal/usecase"
	authv1 "github.com/Garryflop/DormOS-gen-go/auth/v1"
	"github.com/jackc/pgx/v5/pgxpool"
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
	shutdown, err := config.InitOpenTelemetry(context.Background(), "auth-service", otelEndpoint)
	if err != nil {
		log.Printf("⚠ Failed to init OpenTelemetry: %v", err)
	} else {
		defer shutdown(context.Background())
	}

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
	defer rdb.Close()

	// ── Clean Architecture Wiring ──────────────────────────────
	userRepo := postgres.NewUserRepository(db)
	sessionRepo := postgres.NewSessionRepository(db)
	authUC := usecase.NewAuthUseCase(userRepo, sessionRepo, cfg.JWTSecret)
	handler := transport.NewAuthHandler(authUC)

	// ── gRPC Server ────────────────────────────────────────────
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.GRPCPort))
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", cfg.GRPCPort, err)
	}

	grpcServer := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(loggingInterceptor),
	)

	authv1.RegisterAuthServiceServer(grpcServer, handler)
	reflection.Register(grpcServer)

	// ── Session cleanup goroutine ──────────────────────────────
	go func() {
		for range time.Tick(1 * time.Hour) {
			if err := sessionRepo.DeleteExpired(context.Background()); err != nil {
				log.Printf("session cleanup error: %v", err)
			}
		}
	}()

	log.Printf("✓ Auth Service gRPC listening on :%s", cfg.GRPCPort)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

func loggingInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	log.Printf("[auth-service] → %s", info.FullMethod)
	resp, err := handler(ctx, req)
	if err != nil {
		log.Printf("[auth-service] ✗ %s: %v", info.FullMethod, err)
	}
	return resp, err
}
