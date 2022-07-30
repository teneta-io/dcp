package redis

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/teneta-io/dcp/internal/config"
)

func New(ctx context.Context, cfg *config.RedisConfig) (client *redis.Client, err error) {
	client = redis.NewClient(&redis.Options{
		Addr:     cfg.DSN,
		Password: cfg.Password,
	})

	if _, err = client.Ping(ctx).Result(); err != nil {
		return nil, err
	}

	return client, nil
}
