package app

import (
	"context"
	"errors"
	"net/http"
	_ "net/http/pprof"
	"os/signal"
	"runtime"
	"url-shortener/internal/config"
	base "url-shortener/internal/http/handlers"
	"url-shortener/internal/kafka"
	"url-shortener/internal/logger"
	"url-shortener/internal/metrics"
	"url-shortener/internal/storage"
	"url-shortener/internal/storage/postgres"
	redis_ "url-shortener/internal/storage/redis"
	"url-shortener/internal/workers"
)

type App struct {
	Cfg      *config.Config
	Server   *http.Server
	Metrics  *http.Server
	DB       storage.Database
	Cache    storage.Cache
	Logger   logger.Logger
	Producer *kafka.KafkaProducer
}

func New(cfg config.Config, logger logger.Logger) *App {
	return &App{
		Cfg:      &cfg,
		Server:   nil,
		Metrics:  nil,
		DB:       nil,
		Cache:    nil,
		Logger:   logger,
		Producer: nil,
	}
}

func (a *App) Run() {
	ctx, stop := signal.NotifyContext(context.Background())
	defer stop()

	go func() {
		a.Logger.Info("App server started on", a.Cfg.Server.Addr)

		if err := a.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.Logger.Fatal("App server error:", err)
		}
	}()

	go func() {
		a.Logger.Info("Metrics server started on", a.Cfg.Metrics.Addr)

		if err := a.Metrics.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.Logger.Fatal("Metrics server error:", err)
		}
	}()

	cachemetric, err := metrics.NewCacheMetric()

	if err != nil {
		a.Logger.Fatal("Cache metrics error:", err)
	}

	go workers.Scheduler(ctx, a.DB, a.Cache, a.Logger, a.Cfg, cachemetric)

	go a.Producer.Write(ctx, a.Logger)

	consumers := make([]*kafka.KafkaConsumer, 6)

	for i := range consumers {
		consumers[i] = kafka.NewConsumer(a.Cfg, a.Cache, i)
		go consumers[i].Read(ctx, a.Logger, i)
	}

	var m runtime.MemStats

	go metrics.StatsChecker(ctx, a.Logger, m, a.Cfg.Metrics.StatsTimeout)

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), a.Cfg.Server.ShutdownTimeout)
	defer cancel()

	a.Logger.Info("Gracefully shutting down")

	if err := a.Server.Shutdown(shutdownCtx); err != nil {
		a.Logger.Error("App server shutdown error:", err)
	}

	if err := a.Metrics.Shutdown(context.Background()); err != nil {
		a.Logger.Error("Metrics server shutdown error:", err)
	}

	a.Producer.Close(a.Logger)

	for i := range consumers {
		consumers[i].Close(a.Logger)
	}

	if err := a.Cache.Close(); err != nil {
		a.Logger.Error("Cache shutdown error:", err)
	}

	if err := a.DB.Close(); err != nil {
		a.Logger.Error("DB shutdown error:", err)
	}

	runtime.ReadMemStats(&m)
	a.Logger.Printf("INFO", "Number of goroutines: %d, Memory: %d KB, Objects: %d, NumGC : %d, NextGC: %d\n",
		runtime.NumGoroutine(),
		m.Alloc/1024,
		m.HeapObjects,
		m.NumGC,
		m.NextGC)

	runtime.GC()
	runtime.ReadMemStats(&m)
	a.Logger.Printf("INFO", "Goroutines: %d, Alloc: %d KB, HeapObjects: %d, NumGC: %d\n",
		runtime.NumGoroutine(), m.Alloc/1024, m.HeapObjects, m.NumGC)

	a.Logger.Close()
}

func (a *App) Init() {
	var err error

	err = a.InitStorage()

	if err != nil {
		a.Logger.Fatal("Storage initialization failed:", err)
	}

	a.Logger.Info("Storage initialization finished")

	err = a.InitCleanUp()

	if err != nil {
		a.Logger.Fatal("Storage cleanup failed:", err)
	}

	a.Logger.Info("Inital cleanup finished")

	a.Producer = kafka.NewProducer(a.Cfg)

	router := base.SetupRoutes(a.DB, a.Cache, a.Logger, a.Producer, a.Cfg)

	a.Logger.Info("Routes are set")

	a.Server = &http.Server{
		Addr:         a.Cfg.Server.Addr,
		Handler:      router,
		ReadTimeout:  a.Cfg.Server.ReadTimeout,
		WriteTimeout: a.Cfg.Server.WriteTimeout,
		IdleTimeout:  a.Cfg.Server.IdleTimeout,
	}

	metricsMux := http.NewServeMux()

	metrics.Expose(metricsMux)

	a.Metrics = &http.Server{
		Addr:    a.Cfg.Metrics.Addr,
		Handler: metricsMux,
	}

}

func (a *App) InitStorage() error {
	var err error

	a.Cache, err = redis_.StartRedis(a.Cfg)

	if err != nil {
		return errors.New("Redis connection failed: " + err.Error())
	}

	a.DB, err = postgres.StartDB(a.Cfg)

	if err != nil {
		return errors.New("Postgres connection failed: " + err.Error())
	}

	return nil
}

func (a *App) InitCleanUp() error {
	cacheCtx, cacheCancel := context.WithTimeout(context.Background(), a.Cfg.Cache.Timeout)
	defer cacheCancel()

	err := a.Cache.CleanUp(cacheCtx)
	cacheCancel()

	if err != nil {
		return errors.New("Cache cleanup error: " + err.Error())
	}

	dbCtx, dbCancel := context.WithTimeout(context.Background(), a.Cfg.DB.Timeout)
	defer dbCancel()

	err = a.DB.CleanUp(dbCtx)
	dbCancel()

	if err != nil {
		return errors.New("DB cleanup error: " + err.Error())
	}

	return nil
}
