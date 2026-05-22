FROM golang:1.25-alpine AS builder

WORKDIR /app

# We use Go workspaces, so we copy everything, including the vendor directory
COPY . .

# Pass the service name as a build argument
ARG SERVICE_NAME

# Build the specific service using the local vendor directory (completely offline, instant build!)
RUN cd ${SERVICE_NAME} && CGO_ENABLED=0 GOOS=linux go build -mod=vendor -o /app/service ./cmd/main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/service .

EXPOSE 8080 50051 50052 50053 50054 50055
CMD ["./service"]
