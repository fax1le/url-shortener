package redis

import (
	"context"
	"log"
	"time"
)

func (r *RedisCache) GetUrl(ctx context.Context, slug string) (string, error) {
	longUrl, err := r.rdb.Get(ctx, "url:"+slug).Result()
	return longUrl, err
}

func (r *RedisCache) StoreUrl(ctx context.Context, slug string, longUrl string) error {
	err := r.rdb.Set(ctx, "url:"+slug, longUrl, r.Cfg.Cache.UrlExpiration).Err()
	return err
}

func (r *RedisCache) CleanUp(ctx context.Context) error {
	err := r.rdb.FlushAll(ctx).Err()
	return err
}

func (r *RedisCache) GetIP(ctx context.Context, ip string) (map[string]string, error) {
	res, err := r.rdb.HGetAll(ctx, "ip:"+ip).Result()
	return res, err
}

func (r *RedisCache) StoreIPLimit(ctx context.Context, ip string, rps, tokens float64) error {
	txpipe := r.rdb.TxPipeline()

	txpipe.HSet(ctx, "ip:"+ip, "tokens", tokens, "rps", rps, "refilled_at", time.Now().UnixNano())

	txpipe.Expire(ctx, "ip:"+ip, r.Cfg.Cache.IpExpiration)

	_, err := txpipe.Exec(ctx)

	return err
}

func (r *RedisCache) Increment(ctx context.Context, key string, val int64) error {
	err := r.rdb.IncrBy(ctx, key, val).Err()
	return err
}

func (r *RedisCache) IncrementBatch(ctx context.Context, key string, slugs map[string]int64, num int) error {
	pipe := r.rdb.Pipeline()

	flushed := 0

	for k, v := range slugs {
		pipe.HIncrBy(ctx, key, k, v)
		flushed += (int)(v)
	}

	start := time.Now()
	_, err := pipe.Exec(ctx)
	log.Println("Called by:", num, "Flushing clicks to redis took:", time.Since(start), "Flushed:", flushed)

	return err
}

func (r *RedisCache) HashGetAll(ctx context.Context, key string) (map[string]string, error) {
	mp, err := r.rdb.HGetAll(ctx, key).Result()
	return mp, err
}

func (r *RedisCache) Delete(ctx context.Context, key string) error {
	err := r.rdb.Del(ctx, key).Err()
	return err
}

func (r *RedisCache) RenameSetTTL(ctx context.Context, oldkey string, newkey string, ttl time.Duration) error {
	txpipe := r.rdb.TxPipeline()

	txpipe.Rename(ctx, oldkey, newkey)
	txpipe.Expire(ctx, newkey, ttl)

	_, err := txpipe.Exec(ctx)

	return err
}
