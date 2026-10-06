package config

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

type Config struct {
	ServiceHost    string
	ServicePort    string
	MinioPublicURL string // публичный адрес бакета Minio для браузера
}

func NewConfig() *Config {
	if err := godotenv.Load(); err != nil {
		logrus.Warn("файл .env не найден, используются переменные окружения")
	}

	return &Config{
		ServiceHost:    getEnv("SERVICE_HOST", "localhost"),
		ServicePort:    getEnv("SERVICE_PORT", "8080"),
		MinioPublicURL: getEnv("MINIO_PUBLIC_URL", "http://localhost:9000/dignitaries"),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
