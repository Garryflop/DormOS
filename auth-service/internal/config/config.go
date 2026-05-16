package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	GRPCPort             string
	DBHost               string
	DBPort               string
	DBUser               string
	DBPassword           string
	DBName               string
	RedisAddr            string
	RedisPassword        string
	RedisDB              int
	JWTSecret            string
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
	NATSUrl              string
}

func Load() (*Config, error) {
	redisDB, err := strconv.Atoi(getEnv("REDIS_DB", "0"))
	if err != nil {
		return nil, fmt.Errorf("invalid REDIS_DB: %w", err)
	}

	accessDuration, err := time.ParseDuration(getEnv("JWT_ACCESS_DURATION", "15m"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_ACCESS_DURATION: %w", err)
	}

	refreshDuration, err := time.ParseDuration(getEnv("JWT_REFRESH_DURATION", "168h"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_REFRESH_DURATION: %w", err)
	}

	jwtSecret := getEnv("JWT_SECRET", "")
	if jwtSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET must be set")
	}

	return &Config{
		GRPCPort:             getEnv("AUTH_GRPC_PORT", "50051"),
		DBHost:               getEnv("DB_HOST", "localhost"),
		DBPort:               getEnv("DB_PORT", "5432"),
		DBUser:               getEnv("DB_USER", "postgres"),
		DBPassword:           getEnv("DB_PASSWORD", "postgres"),
		DBName:               getEnv("DB_NAME", "auth_db"),
		RedisAddr:            getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:        getEnv("REDIS_PASSWORD", ""),
		RedisDB:              redisDB,
		JWTSecret:            jwtSecret,
		AccessTokenDuration:  accessDuration,
		RefreshTokenDuration: refreshDuration,
		NATSUrl:              getEnv("NATS_URL", "nats://localhost:4222"),
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
