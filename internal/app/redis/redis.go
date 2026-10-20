// Package redis — клиент Redis для чёрного списка токенов (logout)
package redis

import (
	"context"
	"fmt"
	"time"

	"awesomeProject/internal/app/config"

	"github.com/redis/go-redis/v9"
)

// servicePrefix — префикс ключей сервиса, чтобы значения разных сервисов в одном Redis не пересекались
const servicePrefix = "dignitarie_lists."

const blacklistPrefix = servicePrefix + "jwt.blacklist."

type Client struct {
	client *redis.Client
}

func New(ctx context.Context, cfg config.RedisConfig) (*Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:        fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password:    cfg.Password,
		DialTimeout: 10 * time.Second,
		ReadTimeout: 10 * time.Second,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("не удалось подключиться к Redis: %w", err)
	}
	return &Client{client: client}, nil
}

// AddToBlacklist помещает идентификатор токена в чёрный список на оставшееся время жизни токена
func (c *Client) AddToBlacklist(ctx context.Context, tokenID string, ttl time.Duration) error {
	return c.client.Set(ctx, blacklistPrefix+tokenID, "logged_out", ttl).Err()
}

// InBlacklist проверяет, что токен был отозван через logout
func (c *Client) InBlacklist(ctx context.Context, tokenID string) (bool, error) {
	n, err := c.client.Exists(ctx, blacklistPrefix+tokenID).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (c *Client) Close() error {
	return c.client.Close()
}
