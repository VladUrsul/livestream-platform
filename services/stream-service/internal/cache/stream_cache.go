package cache

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type StreamCache interface {
	SetLive(ctx context.Context, username string, streamID string) error
	SetOffline(ctx context.Context, username string) error
	IsLive(ctx context.Context, username string) (bool, error)
	IncrementViewers(ctx context.Context, username string) (int, error)
	DecrementViewers(ctx context.Context, username string) (int, error)
	GetViewerCount(ctx context.Context, username string) (int, error)
	GetStreamID(ctx context.Context, username string) (string, error)
}

type redisStreamCache struct {
	client *redis.Client
}

func NewRedisStreamCache(client *redis.Client) StreamCache {
	return &redisStreamCache{client: client}
}

func (c *redisStreamCache) SetLive(ctx context.Context, username, streamID string) error {
	pipe := c.client.Pipeline()
	pipe.Set(ctx, liveKey(username), streamID, 24*time.Hour)
	pipe.Set(ctx, viewerKey(username), 0, 24*time.Hour)
	_, err := pipe.Exec(ctx)
	return err
}

func (c *redisStreamCache) SetOffline(ctx context.Context, username string) error {
	return c.client.Del(ctx, liveKey(username), viewerKey(username)).Err()
}

func (c *redisStreamCache) IsLive(ctx context.Context, username string) (bool, error) {
	result, err := c.client.Exists(ctx, liveKey(username)).Result()
	if err != nil {
		return false, fmt.Errorf("check live: %w", err)
	}
	return result > 0, nil
}

func (c *redisStreamCache) IncrementViewers(ctx context.Context, username string) (int, error) {
	count, err := c.client.Incr(ctx, viewerKey(username)).Result()
	return int(count), err
}

func (c *redisStreamCache) DecrementViewers(ctx context.Context, username string) (int, error) {
	count, err := c.client.Decr(ctx, viewerKey(username)).Result()
	if count < 0 {
		c.client.Set(ctx, viewerKey(username), 0, 0)
		return 0, nil
	}
	return int(count), err
}

func (c *redisStreamCache) GetViewerCount(ctx context.Context, username string) (int, error) {
	val, err := c.client.Get(ctx, viewerKey(username)).Result()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(val)
}

func (c *redisStreamCache) GetStreamID(ctx context.Context, username string) (string, error) {
	val, err := c.client.Get(ctx, liveKey(username)).Result()
	if err == redis.Nil {
		return "", nil
	}
	return val, err
}

func liveKey(username string) string   { return fmt.Sprintf("stream:live:%s", username) }
func viewerKey(username string) string { return fmt.Sprintf("stream:viewers:%s", username) }
