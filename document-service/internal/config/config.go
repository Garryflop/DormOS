package config

import (
	"fmt"
	"os"
)

type Config struct {
	GRPCPort        string
	DBHost          string
	DBPort          string
	DBUser          string
	DBPassword      string
	DBName          string
	NATSUrl         string
	AuthServiceAddr string
}

func Load() (*Config, error) {
	return &Config{
		GRPCPort:        getEnv("DOCUMENT_GRPC_PORT", "50052"),
		DBHost:          getEnv("DB_HOST", "localhost"),
		DBPort:          getEnv("DB_PORT", "5432"),
		DBUser:          getEnv("DB_USER", "postgres"),
		DBPassword:      getEnv("DB_PASSWORD", "postgres"),
		DBName:          getEnv("DB_NAME", "document_db"),
		NATSUrl:         getEnv("NATS_URL", "nats://localhost:4222"),
		AuthServiceAddr: getEnv("AUTH_SERVICE_ADDR", "localhost:50051"),
	}, nil
}

func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName,
	)
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
