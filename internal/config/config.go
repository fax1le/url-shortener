package config

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server    ServerConfig
	DB        DBConfig
	Cache     CacheConfig
	Kafka     KafkaConfig
	Scheduler SchedulerConfig
	Metrics   MetricsConfig
}

type ServerConfig struct {
	Addr            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

type DBConfig struct {
	Host          string
	User          string
	Password      string
	Name          string
	Timeout       time.Duration
	UrlExpiration time.Duration
}

type CacheConfig struct {
	Host          string
	Password      string
	DB            int
	Timeout       time.Duration
	UrlExpiration time.Duration
	IpExpiration  time.Duration
}

type KafkaConfig struct {
	Brokers             []string
	Topic               string
	ProducerChannelSize int
	ProducerTimeout     time.Duration
	ConsumerTimeout     time.Duration
	ConsumerChannelSize int
	BatchSize           int
	BatchTimeout        time.Duration
	ReadBatchTimeout    time.Duration
	CommitTimeout       time.Duration
	CommitBatchSize     int
	CommitBatchTimeout  time.Duration
}

type SchedulerConfig struct {
	DBCleanupTimeout    time.Duration
	DBFlushTimeout      time.Duration
	CacheFlushTimeout   time.Duration
	MetricsFlushTimeout time.Duration
}

type MetricsConfig struct {
	Addr         string
	StatsTimeout time.Duration
}

func Load() Config {
	err := godotenv.Load()

	if err != nil {
		log.Fatal("Failed to load .env: ", err)
	}

	return Config{
		Server: ServerConfig{
			Addr:            getString("ADDR"),
			ReadTimeout:     getTime("READ_TIMEOUT"),
			WriteTimeout:    getTime("WRITE_TIMEOUT"),
			IdleTimeout:     getTime("IDLE_TIMEOUT"),
			ShutdownTimeout: getTime("SHUTDOWN_TIMEOUT"),
		},
		DB: DBConfig{
			Host:          getString("DB_HOST"),
			User:          getString("DB_USER"),
			Password:      getString("DB_PASSWORD"),
			Name:          getString("DB_NAME"),
			Timeout:       getTime("DB_TIMEOUT"),
			UrlExpiration: getTime("DB_URL_EXPIRATION"),
		},
		Cache: CacheConfig{
			Host:          getString("CACHE_HOST"),
			Password:      os.Getenv("CACHE_PASSWORD"),
			DB:            getInt("CACHE_DB"),
			Timeout:       getTime("CACHE_TIMEOUT"),
			UrlExpiration: getTime("CACHE_URL_EXPIRATION"),
			IpExpiration:  getTime("CACHE_IP_EXPIRATION"),
		},
		Kafka: KafkaConfig{
			Brokers:             getSliceString("KAFKA_BROKERS"),
			Topic:               getString("KAFKA_TOPIC"),
			ProducerChannelSize: getInt("KAFKA_PRODUCER_CHANNEL_SIZE"),
			ProducerTimeout:     getTime("KAFKA_PRODUCER_TIMEOUT"),
			ConsumerTimeout:     getTime("KAFKA_CONSUMER_TIMEOUT"),
			ConsumerChannelSize: getInt("KAFKA_CONSUMER_CHANNEL_SIZE"),
			BatchSize:           getInt("KAFKA_BATCH_SIZE"),
			BatchTimeout:        getTime("KAFKA_BATCH_TIMEOUT"),
			ReadBatchTimeout:    getTime("KAFKA_READ_BATCH_TIMEOUT"),
			CommitTimeout:       getTime("KAFKA_COMMIT_TIMEOUT"),
			CommitBatchSize:     getInt("KAFKA_COMMIT_BATCH_SIZE"),
			CommitBatchTimeout:  getTime("KAFKA_COMMIT_BATCH_TIMEOUT"),
		},
		Scheduler: SchedulerConfig{
			DBCleanupTimeout:    getTime("SCHEDULER_DB_CLEANUP_TIMEOUT"),
			DBFlushTimeout:      getTime("SCHEDULER_DB_FLUSH_TIMEOUT"),
			CacheFlushTimeout:   getTime("SCHEDULER_CACHE_FLUSH_TIMEOUT"),
			MetricsFlushTimeout: getTime("SCHEDULER_METRICS_FLUSH_TIMEOUT"),
		},
		Metrics: MetricsConfig{
			Addr:         getString("METRICS_ADDR"),
			StatsTimeout: getTime("METRICS_STATS_TIMEOUT"),
		},
	}
}

func getString(key string) string {
	val := os.Getenv(key)

	if val == "" {
		log.Fatal("Failed to load .env")
	}

	return val
}

func getInt(key string) int {
	val := os.Getenv(key)

	if val == "" {
		log.Fatal("Failed to load .env")
	}

	num, err := strconv.Atoi(val)

	if err != nil {
		log.Fatal("Failed to load .env: ", err)
	}

	return num
}

func getFloat(key string) float64 {
	val := os.Getenv(key)

	if val == "" {
		log.Fatal("Failed to load .env")
	}

	num, err := strconv.ParseFloat(val, 64)

	if err != nil {
		log.Fatal("Failed to load .env: ", err)
	}

	return num
}

func getTime(key string) time.Duration {
	val := os.Getenv(key)

	if val == "" {
		log.Fatal("Failed to load .env")
	}

	duration, err := time.ParseDuration(val)

	if err != nil {
		log.Fatal("Failed to load .env: ", err)
	}

	return duration
}

func getSliceString(key string) []string {
	val := os.Getenv(key)

	if val == "" {
		log.Fatal("Failed to load .env")
	}

	elems := strings.Split(val, ",")

	return elems
}
