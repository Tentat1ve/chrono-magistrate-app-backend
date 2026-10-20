package config

import (
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

type Config struct {
	ServiceHost string
	ServicePort string

	MinioEndpoint  string // адрес S3 API для сервера
	MinioAccessKey string
	MinioSecretKey string
	MinioBucket    string
	MinioPublicURL string // публичный адрес бакета для клиента

	JWT   JWTConfig
	Redis RedisConfig
}

// JWTConfig — настройки токенов авторизации
type JWTConfig struct {
	Secret    string        // ключ подписи токена
	ExpiresIn time.Duration // время жизни токена
}

// RedisConfig — подключение к Redis (чёрный список токенов)
type RedisConfig struct {
	Host     string
	Port     string
	Password string
}

func NewConfig() *Config {
	if err := godotenv.Load(); err != nil {
		logrus.Warn("файл .env не найден, используются переменные окружения")
	}

	expiresIn, err := time.ParseDuration(getEnv("JWT_EXPIRES_IN", "1h"))
	if err != nil {
		logrus.Warnf("некорректный JWT_EXPIRES_IN, используется 1h: %v", err)
		expiresIn = time.Hour
	}

	return &Config{
		ServiceHost:    getEnv("SERVICE_HOST", "localhost"),
		ServicePort:    getEnv("SERVICE_PORT", "8080"),
		MinioEndpoint:  getEnv("MINIO_ENDPOINT", "localhost:9000"),
		MinioAccessKey: getEnv("MINIO_ACCESS_KEY", "minioadmin"),
		MinioSecretKey: getEnv("MINIO_SECRET_KEY", "minioadminpassword"),
		MinioBucket:    getEnv("MINIO_BUCKET", "dignitaries"),
		MinioPublicURL: getEnv("MINIO_PUBLIC_URL", "http://localhost:9000/dignitaries"),
		JWT: JWTConfig{
			Secret:    getEnv("JWT_SECRET", "change-me-in-env"),
			ExpiresIn: expiresIn,
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
		},
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
