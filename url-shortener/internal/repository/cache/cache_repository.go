package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"
	"url-shortener/internal/models"

	"github.com/redis/go-redis/v9"
)

type CacheRepository struct {
	client *redis.Client
}

func NewCacheRepository(client *redis.Client) *CacheRepository {
	return &CacheRepository{client: client}
}

func (r *CacheRepository) GetURL(ctx context.Context, shortCode string) (*models.URL, error) {
	key := fmt.Sprintf("url:%s", shortCode)

	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get from cache: %w", err)
	}

	var url models.URL
	if err := json.Unmarshal([]byte(data), &url); err != nil {
		return nil, fmt.Errorf("failed to unmarshal url: %w", err)
	}

	return &url, nil
}

func (r *CacheRepository) SetURL(ctx context.Context, url *models.URL) error {
	key := fmt.Sprintf("url:%s", url.ShortCode)

	data, err := json.Marshal(url)
	if err != nil {
		return fmt.Errorf("failed to marshal url: %w", err)
	}

	if err := r.client.Set(ctx, key, data, time.Hour).Err(); err != nil {
		return fmt.Errorf("failed to set cache: %w", err)
	}

	return nil
}

func (r *CacheRepository) DeleteURL(ctx context.Context, shortCode string) error {
	key := fmt.Sprintf("url:%s", shortCode)

	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete from cache: %w", err)
	}

	return nil
}
