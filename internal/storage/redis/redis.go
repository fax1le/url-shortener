package redis

import (
	"context"
	"url-shortener/internal/config"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	rdb *redis.Client
	Cfg *config.Config
}

func StartRedis(cfg *config.Config) (*RedisCache, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:" + cfg.Cache.Host,
		Password: cfg.Cache.Password,
		DB:       cfg.Cache.DB,
		Protocol: 2,
		PoolSize: 100,
		MinIdleConns: 20,
	})

	err := rdb.Ping(context.Background()).Err()

	return &RedisCache{
		rdb: rdb,
		Cfg: cfg,
	}, err
}

func (r *RedisCache) Close() error {
	err := r.rdb.Close()
	return err
}
