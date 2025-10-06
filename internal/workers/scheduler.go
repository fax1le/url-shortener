package workers

import (
	"context"
	"fmt"
	"time"
	"url-shortener/internal/config"
	"url-shortener/internal/logger"
	"url-shortener/internal/metrics"
	"url-shortener/internal/storage"
	"url-shortener/internal/utils/url"
)

func Scheduler(stop context.Context, db storage.Database, cache storage.Cache, logger logger.Logger, cfg *config.Config, cachemetric *metrics.CacheMetric) {
	logger.Info("Scheduler started")

	dbCleanup := time.NewTicker(cfg.Scheduler.DBCleanupTimeout)
	defer dbCleanup.Stop()

	dbFlushClicks := time.NewTicker(cfg.Scheduler.DBFlushTimeout)
	defer dbFlushClicks.Stop()

	cacheFlushMetrics := time.NewTicker(cfg.Scheduler.MetricsFlushTimeout)
	defer cacheFlushMetrics.Stop()

	for {
		select {
		case <-stop.Done():
			logger.Info("Scheduler stopped")
			return
		case <-dbCleanup.C:
			logger.Info("Scheduler triggered cleanup")

			dbCtx, dbCancel := context.WithTimeout(context.Background(), cfg.DB.Timeout)
			defer dbCancel()

			RowsAffected, err := db.ExpireUrls(dbCtx)
			dbCancel()

			if err != nil {
				logger.Error("Scheduler cleanup error:", err)
			} else {
				logger.Info("Scheduler expired:", RowsAffected, "urls")
			}
		case <-dbFlushClicks.C:
			logger.Info("Scheduler triggered flushing clicks")

			timestamp := time.Now().Unix()
			newKey := fmt.Sprintf("clicks:processing:%v", timestamp)

			cacheCtx, cacheCancel := context.WithTimeout(context.Background(), cfg.Cache.Timeout)

			err := cache.RenameSetTTL(cacheCtx, "clicks", newKey, cfg.Cache.UrlExpiration)
			cacheCancel()

			if err != nil {
				logger.Error("Scheduler failed to proccess clicks:", err)
				continue
			}

			cacheCtx, cacheCancel = context.WithTimeout(context.Background(), cfg.Cache.Timeout)

			mp, err := cache.HashGetAll(cacheCtx, newKey)
			cacheCancel()

			if err != nil {
				logger.Error("Scheduler failed to get data from cache:", err)
				continue
			}

			clicks, err := url_utils.ConvertToInt64(mp)

			query, args := url_utils.BuildClicksArgs(clicks)

			dbCtx, dbCancel := context.WithTimeout(context.Background(), cfg.DB.Timeout)

			err = db.StoreClicks(dbCtx, query, args...)
			dbCancel()

			if err != nil {
				logger.Error("Scheduler failed to store clicks: key", newKey, "error:", err)
				continue
			}

			cacheCtx, cacheCancel = context.WithTimeout(context.Background(), cfg.Cache.Timeout)

			err = cache.Delete(cacheCtx, newKey)
			cacheCancel()

			if err != nil {
				logger.Error("Scheduler failed to delete cached clicks:", err)
			} else {
				logger.Info("Scheduler successfully flushed clicks")
			}
		case <-cacheFlushMetrics.C:
			cacheHits := metrics.CacheHitsCounter.Swap(0)
			cacheMisses := metrics.CacheMissesCounter.Swap(0)

			if cacheHits > 0 {
				cachemetric.TotalCacheHits.Add(float64(cacheHits))
			}

			if cacheMisses > 0 {
				cachemetric.TotalCacheMisses.Add(float64(cacheMisses))
			}
		}
	}
}
