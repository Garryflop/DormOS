package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/Garryflop/DormManage/document-service/internal/config"
	grpcserver "github.com/Garryflop/DormManage/document-service/internal/grpc"
	"github.com/Garryflop/DormManage/document-service/internal/nats"
	"github.com/Garryflop/DormManage/document-service/internal/repository"
	"github.com/Garryflop/DormManage/document-service/internal/service"
	documentv1 "github.com/Garryflop/DormOS-gen-go/document/v1"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	ctx := context.Background()

	log.Println("connecting to PostgreSQL...")
	db, err := pgxpool.New(ctx, cfg.DSN())
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		log.Fatalf("database ping failed: %v", err)
	}
	log.Println("PostgreSQL connected")

	log.Println("connecting to NATS...")
	publisher, err := nats.NewPublisher(cfg.NATSUrl)
	if err != nil {
		log.Fatalf("failed to connect to NATS: %v", err)
	}
	defer publisher.Close()
	log.Println("NATS connected")

	workflowRepo := repository.NewWorkflowRepository(db)
	requestRepo := repository.NewRequestRepository(db)
	stepRepo := repository.NewStepRepository(db)
	docSvc := service.NewDocumentService(workflowRepo, requestRepo, stepRepo, publisher)
	grpcSrv := grpcserver.NewServer(docSvc)

	addr := fmt.Sprintf(":%s", cfg.GRPCPort)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen on %s: %v", addr, err)
	}

	server := grpc.NewServer()
	documentv1.RegisterDocumentServiceServer(server, grpcSrv)
	reflection.Register(server)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("document-service gRPC server listening on %s", addr)
		if err := server.Serve(lis); err != nil {
			log.Fatalf("gRPC server error: %v", err)
		}
	}()

	<-quit
	log.Println("shutting down document-service...")
	server.GracefulStop()
	log.Println("document-service stopped")
}
