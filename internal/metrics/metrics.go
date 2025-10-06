package metrics

import (
	"context"
	"net/http"
	"runtime"
	"sync/atomic"
	"time"
	"url-shortener/internal/logger"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	CacheHitsCounter   atomic.Int64
	CacheMissesCounter atomic.Int64
)

type HttpMetric struct {
	TotalRequests   prometheus.CounterVec
	LatencyRequests prometheus.Histogram
}

type CacheMetric struct {
	TotalCacheHits   prometheus.Counter
	TotalCacheMisses prometheus.Counter
}

func NewHttpMetric(name string) (*HttpMetric, error) {
	Total := *prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: name + "_total",
		Help: "Total requests on " + name,
	}, []string{"method", "status"})

	Latency := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    name + "_latency",
		Help:    "Request latency on " + name,
		Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2, 5, 10},
	})

	err := prometheus.Register(Total)

	if err != nil {
		return nil, err
	}

	err = prometheus.Register(Latency)

	if err != nil {
		return nil, err
	}

	return &HttpMetric{
		TotalRequests:   Total,
		LatencyRequests: Latency,
	}, nil
}

func NewCacheMetric() (*CacheMetric, error) {
	TotalCacheHits := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "cache_hits_total",
		Help: "Total cache hits",
	})

	TotalCacheMisses := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "cache_misses_total",
		Help: "Total cache misses",
	})

	err := prometheus.Register(TotalCacheHits)

	if err != nil {
		return nil, err
	}

	err = prometheus.Register(TotalCacheMisses)

	if err != nil {
		return nil, err
	}

	return &CacheMetric{
		TotalCacheHits:   TotalCacheHits,
		TotalCacheMisses: TotalCacheMisses,
	}, nil
}

func (h *HttpMetric) Export(method string, status string, latency time.Duration) {
	h.TotalRequests.WithLabelValues(method, status).Inc()
	h.LatencyRequests.Observe(latency.Seconds())
}

func Expose(mux *http.ServeMux) {
	mux.Handle("/metrics", promhttp.Handler())
}

func StatsChecker(stop context.Context, logger logger.Logger, m runtime.MemStats, timeout time.Duration) {
	ticker := time.NewTicker(timeout)
	defer ticker.Stop()

	for {
		select {
		case <-stop.Done():
			logger.Info("Stopping stats checker")
			return
		case <-ticker.C:
			runtime.ReadMemStats(&m)
			logger.Printf("INFO", "Number of goroutines: %d, Memory: %d KB, Objects: %d, NumGC : %d, NextGC: %d\n",
				runtime.NumGoroutine(),
				m.Alloc/1024,
				m.HeapObjects,
				m.NumGC,
				m.NextGC)
		}
	}

}
