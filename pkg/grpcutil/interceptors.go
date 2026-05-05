package grpcutil

import (
	"context"
	"log"

	"google.golang.org/grpc"
)

// LoggingInterceptor logs incoming gRPC requests.
func LoggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	log.Printf("Received request for %s", info.FullMethod)
	return handler(ctx, req)
}
