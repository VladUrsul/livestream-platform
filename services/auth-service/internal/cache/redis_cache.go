package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// AuthCache defines caching operations for the auth service.
type AuthCache interface {
	BlacklistToken(ctx context.Context, tokenID string, expiry time.Duration) error
	IsBlacklisted(ctx context.Context, tokenID string) (bool, error)
	StoreRefreshToken(ctx context.Context, userID, tokenHash string, expiry time.Duration) error
	GetRefreshToken(ctx context.Context, userID string) (string, error)
	DeleteRefreshToken(ctx context.Context, userID string) error
}

type redisAuthCache struct {
	client *redis.Client
}

// NewRedisAuthCache creates a Redis-backed auth cache.
func NewRedisAuthCache(client *redis.Client) AuthCache {
	return &redisAuthCache{client: client}
}

func (c *redisAuthCache) BlacklistToken(ctx context.Context, tokenID string, expiry time.Duration) error {
	return c.client.Set(ctx, blacklistKey(tokenID), "1", expiry).Err()
}

func (c *redisAuthCache) IsBlacklisted(ctx context.Context, tokenID string) (bool, error) {
	result, err := c.client.Exists(ctx, blacklistKey(tokenID)).Result()
	if err != nil {
		return false, fmt.Errorf("check blacklist: %w", err)
	}
	return result > 0, nil
}

func (c *redisAuthCache) StoreRefreshToken(ctx context.Context, userID, tokenHash string, expiry time.Duration) error {
	return c.client.Set(ctx, refreshKey(userID), tokenHash, expiry).Err()
}

func (c *redisAuthCache) GetRefreshToken(ctx context.Context, userID string) (string, error) {
	val, err := c.client.Get(ctx, refreshKey(userID)).Result()
	if err == redis.Nil {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("get refresh token: %w", err)
	}
	return val, nil
}

func (c *redisAuthCache) DeleteRefreshToken(ctx context.Context, userID string) error {
	return c.client.Del(ctx, refreshKey(userID)).Err()
}

func blacklistKey(tokenID string) string {
	return fmt.Sprintf("auth:blacklist:%s", tokenID)
}

func refreshKey(userID string) string {
	return fmt.Sprintf("auth:refresh:%s", userID)
}
