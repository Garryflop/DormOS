package config

import "os"

type Config struct {
	DatabaseURL     string
	RedisURL        string
	GRPCPort        string
	MinioEndpoint   string
	MinioAccessKey  string
	MinioSecretKey  string
	MinioBucket     string
	MinioUseSSL     bool
}

func Load() *Config {
	return &Config{
		DatabaseURL:    getEnv("DATABASE_URL", "postgres://dormos:dormos123@localhost:5432/dormos?sslmode=disable"),
		RedisURL:       getEnv("REDIS_URL", "redis://:redis123@localhost:6379/0"),
		GRPCPort:       getEnv("FILE_SERVICE_PORT", "50053"),
		MinioEndpoint:  getEnv("MINIO_ENDPOINT", "localhost:9000"),
		MinioAccessKey: getEnv("MINIO_ACCESS_KEY", "admin"),
		MinioSecretKey: getEnv("MINIO_SECRET_KEY", "admin123"),
		MinioBucket:    getEnv("MINIO_BUCKET", "dormos-files"),
		MinioUseSSL:    getEnv("MINIO_USE_SSL", "false") == "true",
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
