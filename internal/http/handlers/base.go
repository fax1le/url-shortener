package base

import (
	"net/http"
	"url-shortener/internal/config"
	"url-shortener/internal/http/handlers/url"
	"url-shortener/internal/kafka"
	"url-shortener/internal/logger"
	"url-shortener/internal/storage"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(db storage.Database, cache storage.Cache, logger logger.Logger, producer *kafka.KafkaProducer, cfg *config.Config) *gin.Engine {
	r := gin.New()

	r.Use(gin.Recovery())

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, "pong")
	})

	url.AddUrlRoutes(r, db, cache, logger, producer, cfg)

	return r
}
