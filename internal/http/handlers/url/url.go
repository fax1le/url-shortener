package url

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"url-shortener/internal/config"
	"url-shortener/internal/kafka"
	"url-shortener/internal/logger"
	"url-shortener/internal/metrics"
	"url-shortener/internal/middleware/ratelimiter"
	"url-shortener/internal/middleware/requests"
	"url-shortener/internal/models"
	"url-shortener/internal/storage"
	"url-shortener/internal/utils/url"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type UrlHandler struct {
	Cfg      *config.Config
	Logger   logger.Logger
	DB       storage.Database
	Cache    storage.Cache
	Producer *kafka.KafkaProducer
}

var (
	ErrSlugExists = errors.New("Slug exists")
)

func AddUrlRoutes(r *gin.Engine, db storage.Database, cache storage.Cache, logger logger.Logger, producer *kafka.KafkaProducer, cfg *config.Config) {
	u := UrlHandler{
		DB:       db,
		Cache:    cache,
		Logger:   logger,
		Producer: producer,
		Cfg:      cfg,
	}

	redirectMetric, _ := metrics.NewHttpMetric("redirect")
	createMetric, _ := metrics.NewHttpMetric("shorten")
	qrMetric, _ := metrics.NewHttpMetric("qrcode")

	r.GET("/:slug", requests.LoggingMiddleware(u.Logger, *redirectMetric), ratelimiter.RateLimiter(10000, 100), u.RedirectHandler)
	r.POST("/shorten", requests.LoggingMiddleware(u.Logger, *createMetric), ratelimiter.RateLimiter(100, 100), u.CreateUrlHandler)
	r.GET("/qr/:slug", requests.LoggingMiddleware(u.Logger, *qrMetric), ratelimiter.RateLimiter(1000, 100), u.QrCodeHandler)
}

func (u *UrlHandler) RedirectHandler(c *gin.Context) {
	slug := c.Param("slug")

	err := url_utils.ValidateSlug(slug)

	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	cacheCtx, cacheCancel := context.WithTimeout(context.Background(), u.Cfg.Cache.Timeout)
	defer cacheCancel()

	longUrl, err := u.Cache.GetUrl(cacheCtx, slug)
	cacheCancel()

	if err == redis.Nil {
		metrics.CacheMissesCounter.Add(1)
		u.Logger.Warn("Cache missed", slug)
	} else if err != nil {
		u.Logger.Error("Cache error:", err)
		c.Status(http.StatusInternalServerError)
		return
	} else {
		select {
		case u.Producer.EventChan <- slug:
		//	u.Logger.Info("Event sent successfully")
		default:
			u.Logger.Warn("Dropping event, channel was full TO PRODUCER:", slug)
		}

		metrics.CacheHitsCounter.Add(1)

		u.Logger.Info("Redirect", longUrl, slug)
		c.Redirect(http.StatusPermanentRedirect, longUrl)

		return
	}

	dbCtx, dbCancel := context.WithTimeout(context.Background(), u.Cfg.DB.Timeout)
	defer dbCancel()

	longUrl, err = u.DB.GetUrl(dbCtx, slug)
	dbCancel()

	if err == sql.ErrNoRows {
		c.Status(http.StatusNotFound)
		return
	}

	if err != nil {
		u.Logger.Error("Postgres error:", err)
		c.Status(http.StatusInternalServerError)
		return
	}

	cacheCtx, cacheCancel = context.WithTimeout(context.Background(), u.Cfg.Cache.Timeout)
	defer cacheCancel()

	err = u.Cache.StoreUrl(cacheCtx, slug, longUrl)
	cacheCancel()

	if err != nil {
		u.Logger.Warn("Failed to cache, allowing to continue", longUrl, slug, err)
	}

	select {
	case u.Producer.EventChan <- slug:
	//	u.Logger.Info("Event sent successfully")
	default:
		u.Logger.Warn("Dropping event, channel was full:", slug)
	}

	u.Logger.Info("Redirect", longUrl, slug)
	c.Redirect(http.StatusPermanentRedirect, longUrl)
}

func (u *UrlHandler) CreateUrlHandler(c *gin.Context) {
	var newUrl models.Url

	err := c.BindJSON(&newUrl)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "bad request",
			"error":   err.Error(),
		})
		return
	}

	err = url_utils.ValidateUrl(newUrl.LongUrl)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "validation fail",
			"error":   err.Error(),
		})
		return
	}

	var slug string

	if newUrl.CustomAlias != "" {
		err = url_utils.ValidateSlug(newUrl.CustomAlias)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "invalid alias",
				"error":   err.Error(),
			})
			return
		}

		slug = newUrl.CustomAlias
	} else {
		slug, err = url_utils.GenerateSlug(7)

		if err != nil {
			u.Logger.Error("Short url generation error:", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "generation fail",
				"error":   err.Error(),
			})

			return
		}

	}

	dbCtx, dbCancel := context.WithTimeout(context.Background(), u.Cfg.DB.Timeout)
	defer dbCancel()

	err = u.DB.StoreUrl(dbCtx, newUrl.LongUrl, slug, u.Cfg.DB.UrlExpiration)
	dbCancel()

	if err != nil {
		u.Logger.Error("Postgres error:", err)

		status := 500

		if errors.Is(err, ErrSlugExists) {
			status = 400
		}

		c.JSON(status, gin.H{
			"message": "store fail",
			"error":   err.Error(),
		})
		return
	}

	shortUrl := "http://localhost:8080/" + slug
	qrUrl := "http://localhost:8080/qr/" + slug

	c.JSON(http.StatusCreated, gin.H{
		"message":   "successfully created",
		"short_url": shortUrl,
		"slug":      slug,
		"qr":        qrUrl,
	})

	u.Logger.Info("Short url created", newUrl.LongUrl, shortUrl)
}

func (u *UrlHandler) QrCodeHandler(c *gin.Context) {
	slug := c.Param("slug")

	shortUrl := "http://localhost:8080/" + slug

	data, err := url_utils.GenerateQrCode(shortUrl)

	if err != nil {
		c.Status(http.StatusInternalServerError)
		u.Logger.Error("QR code generation error:", err)
		return
	}

	c.Data(http.StatusOK, "image/png", data)
}
