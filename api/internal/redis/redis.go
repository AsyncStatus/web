package redis

import (
	"api/config"
	"context"

	"github.com/nitishm/go-rejson/v4"
	goRedis "github.com/redis/go-redis/v9"
)

type Client struct {
	*goRedis.Client
	Rh *rejson.Handler
}

func MustNewClient(ctx context.Context, cfg *config.Config) *Client {
	opts, err := goRedis.ParseURL(cfg.RedisURL)
	if err != nil {
		panic(err)
	}

	redisClient := goRedis.NewClient(opts)
	if err = redisClient.Ping(ctx).Err(); err != nil {
		panic(err)
	}

	rh := rejson.NewReJSONHandler()
	rh.SetGoRedisClientWithContext(ctx, redisClient)

	return &Client{Client: redisClient, Rh: rh}
}
