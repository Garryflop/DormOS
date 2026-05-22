FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy everything including the root vendor folder
COPY . .

WORKDIR /app/issue-service

# Build the service using the root vendor directory (completely offline, instant!)
RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor -o issue-service ./cmd/main.go

FROM alpine:3.19

WORKDIR /app

COPY --from=builder /app/issue-service/issue-service .
COPY --from=builder /app/issue-service/migrations ./migrations

EXPOSE 50052

CMD ["./issue-service"]