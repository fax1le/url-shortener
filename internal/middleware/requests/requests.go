package requests

import (
	"strconv"
	"time"
	"url-shortener/internal/logger"
	"url-shortener/internal/metrics"

	"github.com/gin-gonic/gin"
)

func LoggingMiddleware(logger logger.Logger, metric metrics.HttpMetric) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		method := c.Request.Method
		status := c.Writer.Status()
		latency := time.Since(start)
		ip := c.ClientIP()
		path := c.Request.URL.Path

		statusStr := strconv.Itoa(status)

		metric.Export(method, statusStr, latency)

		switch {
		case status >= 500:
			logger.Error("Server error", "|", status, "|", latency, "|", ip, "|", method, "|", "\""+path+"\"")
		case status >= 400:
			logger.Warn("Client error", "|", status, "|", latency, "|", ip, "|", method, "|", "\""+path+"\"")
		default:
			logger.Info("|", status, "|", latency, "|", ip, "|", method, "|", "\""+path+"\"")
		}
	}
}
