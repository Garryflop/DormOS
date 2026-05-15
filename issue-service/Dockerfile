FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy shared pkg module
COPY pkg/ ./pkg/

# Copy issue-service
COPY issue-service/ ./issue-service/

WORKDIR /app/issue-service

RUN go mod download
RUN go build -o issue-service ./cmd/main.go

FROM alpine:3.19

WORKDIR /app

COPY --from=builder /app/issue-service/issue-service .
COPY --from=builder /app/issue-service/migrations ./migrations

EXPOSE 50052

CMD ["./issue-service"]