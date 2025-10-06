package kafka

import (
	"context"
	"errors"
	"time"
	"url-shortener/internal/config"
	"url-shortener/internal/logger"
	"url-shortener/internal/storage"

	"github.com/segmentio/kafka-go"
)

type KafkaConsumer struct {
	Reader   *kafka.Reader
	Messages map[string]int64
	MsgChan  chan string
	Cache    storage.Cache
	Cfg      *config.Config
}

func NewConsumer(cfg *config.Config, cache storage.Cache, num int) *KafkaConsumer {
	return &KafkaConsumer{
		Reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:          cfg.Kafka.Brokers,
			Topic:            cfg.Kafka.Topic,
			GroupID:          "click-consumers",
			ReadBatchTimeout: cfg.Kafka.ReadBatchTimeout,
			StartOffset:      kafka.LastOffset,
		}),
		Messages: make(map[string]int64),
		MsgChan:  make(chan string, cfg.Kafka.ConsumerChannelSize),
		Cache:    cache,
		Cfg:      cfg,
	}
}

func (k *KafkaConsumer) Read(stop context.Context, logger logger.Logger, num int) {
	ticker := time.NewTicker(k.Cfg.Scheduler.CacheFlushTimeout)
	defer ticker.Stop()

	st := time.Now()

	go k.Fetcher(stop, k.MsgChan, logger)

	for {
		select {
		case <-stop.Done():
			logger.Info("Consumer worker stopped:", num)
			k.Messages = nil
			return
		case <-ticker.C:
			if len(k.Messages) > 0 {
				cacheCtx, cacheCancel := context.WithTimeout(context.Background(), k.Cfg.Cache.Timeout)

				err := k.Cache.IncrementBatch(cacheCtx, "clicks", k.Messages, num)
				cacheCancel()

				if err != nil {
					logger.Error("Failed to cache url clicks:", err)
				} else {
					clear(k.Messages)
				}

			}
		case msg := <-k.MsgChan:
			k.Messages[msg]++

			logger.Info("Consumer read message in:", time.Since(st))
			st = time.Now()

			if len(k.Messages) >= 1000 {
				cacheCtx, cacheCancel := context.WithTimeout(context.Background(), k.Cfg.Cache.Timeout)

				err := k.Cache.IncrementBatch(cacheCtx, "clicks", k.Messages, num)
				cacheCancel()

				if err != nil {
					logger.Error("Failed to cache url clicks:", err)
				} else {
					clear(k.Messages)
				}
			}
		}
	}
}

func (k *KafkaConsumer) Fetcher(ctx context.Context, msgChan chan string, logger logger.Logger) {
	defer close(msgChan)

	commitBatch := time.NewTicker(k.Cfg.Kafka.CommitBatchTimeout)
	defer commitBatch.Stop()

	batchMap := make(map[int][]kafka.Message, k.Cfg.Kafka.CommitBatchSize)

	for {
		start := time.Now()
		msg, err := k.Reader.FetchMessage(ctx)

		logger.Info("Fetching message took:", time.Since(start))

		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			logger.Info("Fetcher stopped")

			for _, batch := range batchMap {
				if len(batch) > 0 {
					commitCtx, commitCancel := context.WithTimeout(context.Background(), k.Cfg.Kafka.CommitTimeout)

					err = k.Reader.CommitMessages(commitCtx, batch...)
					commitCancel()

					if err != nil {
						logger.Warn("Consumer failed to commit message:", err)
					}
				}
			}

			return
		}

		if err != nil {
			logger.Error("Consumer failed to read message:", err)
			continue
		}

		select {
		case msgChan <- string(msg.Value):
			batchMap[msg.Partition] = append(batchMap[msg.Partition], msg)

			if len(batchMap[msg.Partition]) == cap(batchMap[msg.Partition]) {
				commitCtx, commitCancel := context.WithTimeout(context.Background(), k.Cfg.Kafka.CommitTimeout)

				err := k.Reader.CommitMessages(commitCtx, batchMap[msg.Partition]...)
				commitCancel()

				if err != nil {
					logger.Warn("Consumer failed to commit message:", err)
				} else {
					batchMap[msg.Partition] = batchMap[msg.Partition][:0]
				}
			}
		case <-ctx.Done():
			logger.Info("Fetcher stopped")
			return
		case <-commitBatch.C:
			for partition, batch := range batchMap {
				if len(batch) > 0 {
					commitCtx, commitCancel := context.WithTimeout(context.Background(), k.Cfg.Kafka.CommitTimeout)

					err = k.Reader.CommitMessages(commitCtx, batch...)
					commitCancel()

					if err != nil {
						logger.Warn("Consumer failed to commit message:", err)
					} else {
						batchMap[partition] = batchMap[partition][:0]
					}
				}
			}
		default:
			logger.Info("Channel is full, dropping the message")
		}
	}
}

func (k *KafkaConsumer) Close(logger logger.Logger) {
	logger.Info("Closing consumer")
	k.Reader.Close()
}
