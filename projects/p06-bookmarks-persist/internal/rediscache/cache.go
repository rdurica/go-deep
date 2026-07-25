// Package rediscache je Redis implementace app.Cache.
package rediscache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/rdurica/go-deep/projects/p06-bookmarks-persist/internal/bookmark"
)

// Cache ukládá záložky pod klíčem bookmark:{id}.
type Cache struct {
	client *redis.Client
	ttl    time.Duration
}

// Open připojí Redis klienta z REDIS_URL.
func Open(ctx context.Context, redisURL string, ttl time.Duration) (*Cache, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("redis: parse url: %w", err)
	}
	client := redis.NewClient(opts)
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("redis: ping: %w", err)
	}
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}
	return &Cache{client: client, ttl: ttl}, nil
}

// Close uzavře klienta.
func (c *Cache) Close() error {
	if c == nil || c.client == nil {
		return nil
	}
	return c.client.Close()
}

func key(id string) string {
	return "bookmark:" + id
}

type cacheDTO struct {
	ID        string    `json:"id"`
	URL       string    `json:"url"`
	Title     string    `json:"title"`
	Tags      []string  `json:"tags"`
	CreatedAt time.Time `json:"created_at"`
}

// Get načte záložku z cache. ok=false znamená miss.
func (c *Cache) Get(ctx context.Context, id string) (bookmark.Bookmark, bool, error) {
	raw, err := c.client.Get(ctx, key(id)).Bytes()
	if err == redis.Nil {
		return bookmark.Bookmark{}, false, nil
	}
	if err != nil {
		return bookmark.Bookmark{}, false, fmt.Errorf("redis: get: %w", err)
	}
	var dto cacheDTO
	if err := json.Unmarshal(raw, &dto); err != nil {
		_ = c.client.Del(ctx, key(id)).Err()
		return bookmark.Bookmark{}, false, nil
	}
	tags := dto.Tags
	if tags == nil {
		tags = []string{}
	}
	return bookmark.Bookmark{
		ID:        dto.ID,
		URL:       dto.URL,
		Title:     dto.Title,
		Tags:      tags,
		CreatedAt: dto.CreatedAt.UTC(),
	}, true, nil
}

// Set uloží záložku do cache s TTL.
func (c *Cache) Set(ctx context.Context, b bookmark.Bookmark) error {
	dto := cacheDTO{
		ID:        b.ID,
		URL:       b.URL,
		Title:     b.Title,
		Tags:      b.Tags,
		CreatedAt: b.CreatedAt.UTC(),
	}
	if dto.Tags == nil {
		dto.Tags = []string{}
	}
	raw, err := json.Marshal(dto)
	if err != nil {
		return fmt.Errorf("redis: marshal: %w", err)
	}
	if err := c.client.Set(ctx, key(b.ID), raw, c.ttl).Err(); err != nil {
		return fmt.Errorf("redis: set: %w", err)
	}
	return nil
}

// Delete odstraní klíč z cache.
func (c *Cache) Delete(ctx context.Context, id string) error {
	if err := c.client.Del(ctx, key(id)).Err(); err != nil {
		return fmt.Errorf("redis: del: %w", err)
	}
	return nil
}

// Ready pingne Redis.
func (c *Cache) Ready(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	if err := c.client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis: not ready: %w", err)
	}
	return nil
}
