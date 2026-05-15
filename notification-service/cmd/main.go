package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	activityv1 "github.com/Garryflop/DormOS-gen-go/activity/v1"

	"github.com/redis/go-redis/v9"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	natsclient "github.com/nats-io/nats.go"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pgRepo "github.com/dormos/notification-service/internal/repository/postgres"

	grpcHandler "github.com/dormos/notification-service/internal/transport/grpc"
	natsTransport "github.com/dormos/notification-service/internal/transport/nats"

	"github.com/dormos/notification-service/internal/usecase"
)

func main() {

	// Load .env
	if err := godotenv.Load(); err != nil {
		log.Println("[WARN] .env file not found")
	}

	// PostgreSQL
	dbURL := mustEnv("DATABASE_URL")

	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf(
			"[FATAL] postgres connect: %v",
			err,
		)
	}

	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf(
			"[FATAL] postgres ping: %v",
			err,
		)
	}

	log.Println("[OK] PostgreSQL connected")

	// Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     mustEnv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	if _, err := rdb.Ping(context.Background()).Result(); err != nil {
		log.Fatalf(
			"[FATAL] redis ping: %v",
			err,
		)
	}

	log.Println("[OK] Redis connected")

	// NATS
	nc, err := natsclient.Connect(
		mustEnv("NATS_URL"),
	)

	if err != nil {
		log.Fatalf(
			"[FATAL] nats connect: %v",
			err,
		)
	}

	defer nc.Drain()

	log.Println("[OK] NATS connected")

	// Repositories
	eventRepo := pgRepo.NewEventRepository(pool)
	notifRepo := pgRepo.NewNotificationRepository(pool)

	// SMTP
	smtpCfg := usecase.SMTPConfig{
		Host:     mustEnv("SMTP_HOST"),
		Port:     mustEnv("SMTP_PORT"),
		Username: mustEnv("SMTP_USERNAME"),
		Password: mustEnv("SMTP_PASSWORD"),
		From:     mustEnv("SMTP_FROM"),
	}

	// Publisher
	publisher := &natsPublisher{
		nc: nc,
	}

	// UseCases
	activityUC := usecase.NewActivityUseCase(
		eventRepo,
		rdb,
		publisher,
	)

	notifUC := usecase.NewNotificationUseCase(
		notifRepo,
		smtpCfg,
	)

	// NATS Subscriber
	sub := natsTransport.NewSubscriber(
		nc,
		notifUC,
	)

	if err := sub.Subscribe(); err != nil {
		log.Fatalf(
			"[FATAL] nats subscribe: %v",
			err,
		)
	}

	log.Println("[OK] NATS subscribers registered")

	// gRPC
	grpcPort := getEnv("GRPC_PORT", "50054")

	lis, err := net.Listen(
		"tcp",
		fmt.Sprintf(":%s", grpcPort),
	)

	if err != nil {
		log.Fatalf(
			"[FATAL] listen: %v",
			err,
		)
	}

	grpcServer := grpc.NewServer()

	handler := grpcHandler.NewHandler(
		activityUC,
		notifUC,
	)

	activityv1.RegisterActivityServiceServer(
		grpcServer,
		handler,
	)

	reflection.Register(grpcServer)

	// Graceful shutdown
	quit := make(chan os.Signal, 1)

	signal.Notify(
		quit,
		syscall.SIGINT,
		syscall.SIGTERM,
	)

	go func() {

		log.Printf(
			"[OK] notification-service listening on :%s",
			grpcPort,
		)

		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf(
				"[FATAL] grpc serve: %v",
				err,
			)
		}
	}()

	<-quit

	log.Println("[SHUTDOWN] stopping grpc server")

	grpcServer.GracefulStop()

	log.Println("[SHUTDOWN] done")
}

// Helpers

func mustEnv(key string) string {

	v := os.Getenv(key)

	if v == "" {
		log.Fatalf(
			"[FATAL] %s is required",
			key,
		)
	}

	return v
}

func getEnv(key, fallback string) string {

	v := os.Getenv(key)

	if v == "" {
		return fallback
	}

	return v
}

// NATS Publisher

type natsPublisher struct {
	nc *natsclient.Conn
}

func (p *natsPublisher) Publish(
	subject string,
	data []byte,
) error {

	return p.nc.Publish(
		subject,
		data,
	)
}
