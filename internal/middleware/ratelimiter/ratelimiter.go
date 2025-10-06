package ratelimiter

import (
	"context"
	"net/http"
	"strconv"
	"sync"
	"time"
	"url-shortener/internal/config"
	"url-shortener/internal/storage"
	"url-shortener/internal/utils/url"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

func RateLimiter(rps rate.Limit, burst int) gin.HandlerFunc {
	ipMap := make(map[string]*rate.Limiter)
	mu := sync.Mutex{}

	return func(c *gin.Context) {
		ip := url_utils.GetIP(c.Request)

		mu.Lock()
		limiter, exists := ipMap[ip]
		if !exists {
			limiter = rate.NewLimiter(rps, burst)
			ipMap[ip] = limiter
		}
		mu.Unlock()

		if !limiter.Allow() {
			c.AbortWithStatus(http.StatusTooManyRequests)
			return
		}

		c.Next()
	}
}

func RateLimiterMiddleware(cache storage.Cache, rps, burst float64, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := url_utils.GetIP(c.Request)

		allow, err := Allow(cache, ip, rps, burst, cfg)

		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		if !allow {
			c.AbortWithStatus(http.StatusTooManyRequests)
			return
		}

		c.Next()
	}
}

func Allow(cache storage.Cache, ip string, rps, burst float64, cfg *config.Config) (bool, error) {
	cacheCtx, cacheCancel := context.WithTimeout(context.Background(), cfg.Cache.Timeout)
	defer cacheCancel()

	vals, err := cache.GetIP(cacheCtx, ip)

	if err != nil {
		return false, err
	}

	if _, ok := vals["tokens"]; !ok {
		cacheCtx, cacheCancel := context.WithTimeout(context.Background(), cfg.Cache.Timeout)
		defer cacheCancel()

		err := cache.StoreIPLimit(cacheCtx, ip, rps, burst-1)

		if err != nil {
			return false, err
		}

		return true, nil
	}

	tokens, err := strconv.ParseFloat(vals["tokens"], 64)

	if err != nil {
		return false, err
	}

	refilled_at, err := strconv.ParseInt(vals["refilled_at"], 10, 64)

	if err != nil {
		return false, err
	}

	now := time.Now().UnixNano()

	elapsed := float64(now-refilled_at) / 1e9

	new_tokens := rps * elapsed

	tokens = min(burst, tokens+new_tokens)

	cacheCtx, cacheCancel = context.WithTimeout(context.Background(), cfg.Cache.Timeout)
	defer cacheCancel()

	if tokens >= 1 {
		err = cache.StoreIPLimit(cacheCtx, ip, rps, tokens-1)
		return true, err
	} else {
		err = cache.StoreIPLimit(cacheCtx, ip, rps, tokens)
		return false, nil
	}
}
