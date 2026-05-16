package main

import (
	"context"
	"log"
	"net"
	"os"

	issuev1 "github.com/Garryflop/DormOS-gen-go/issue/v1"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"

	issuegrpc "github.com/Garryflop/DormManage/issue-service/internal/grpc"
	issuenats "github.com/Garryflop/DormManage/issue-service/internal/nats"
	"github.com/Garryflop/DormManage/issue-service/internal/repository"
	"github.com/Garryflop/DormManage/issue-service/internal/service"
)

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func main() {
	ctx := context.Background()

	// --- Postgres ---
	pgURL := getEnv("DATABASE_URL", "postgres://dormos:dormos123@127.0.0.1:5432/dormos?sslmode=disable")
	db, err := pgxpool.New(ctx, pgURL)
	if err != nil {
		log.Fatalf("pgxpool: %v", err)
	}
	defer db.Close()
	if err := db.Ping(ctx); err != nil {
		log.Fatalf("postgres ping: %v", err)
	}
	log.Println("✅ Postgres connected")

	// --- Redis ---
	rdb := redis.NewClient(&redis.Options{
		Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
		Password: getEnv("REDIS_PASSWORD", "redis123"),
	})
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("redis ping: %v", err)
	}
	log.Println("✅ Redis connected")

	// --- NATS ---
	publisher, err := issuenats.NewPublisher(getEnv("NATS_URL", "nats://localhost:4222"))
	if err != nil {
		log.Fatalf("nats: %v", err)
	}
	defer publisher.Close()
	log.Println("✅ NATS connected")

	// --- Repos & Service ---
	svc := service.New(
		repository.NewIssueRepo(db),
		repository.NewCommentRepo(db),
		repository.NewWorkerRepo(db),
		repository.NewCategoryRepo(db),
		publisher,
		rdb,
	)

	// --- gRPC server ---
	grpcPort := getEnv("GRPC_PORT", "50052")
	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	server := grpc.NewServer()
	issuev1.RegisterIssueServiceServer(server, issuegrpc.NewIssueServer(svc))

	log.Printf("🚀 Issue Service gRPC on :%s\n", grpcPort)
	if err := server.Serve(lis); err != nil {
		log.Fatalf("serve: %v", err)
	}
}
