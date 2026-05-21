FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY . .

ARG SERVICE_NAME

RUN cd ${SERVICE_NAME} && go mod download
RUN cd ${SERVICE_NAME} && CGO_ENABLED=0 GOOS=linux go build -o /app/service ./cmd/main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/service .

EXPOSE 8080 50051 50052 50053 50054 50055
CMD ["./service"]
