package config

import "os"

type Config struct {
	DatabaseURL string
	RedisURL    string
	NatsURL     string
	GRPCPort    string
}

func Load() *Config {
	return &Config{
		DatabaseURL: getEnv("DATABASE_URL", "postgres://dormos:dormos123@localhost:5432/dormos?sslmode=disable"),
		RedisURL:    getEnv("REDIS_URL", "redis://:redis123@localhost:6379/0"),
		NatsURL:     getEnv("NATS_URL", "nats://localhost:4222"),
		GRPCPort:    getEnv("ISSUE_SERVICE_PORT", "50052"),
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
