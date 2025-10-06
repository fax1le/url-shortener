package kafka

import (
	"context"
	"url-shortener/internal/config"
	"url-shortener/internal/logger"

	"github.com/segmentio/kafka-go"
)

type KafkaProducer struct {
	Writer    *kafka.Writer
	EventChan chan string
	Cfg       *config.Config
}

func NewProducer(cfg *config.Config) *KafkaProducer {
	return &KafkaProducer{
		Writer: kafka.NewWriter(kafka.WriterConfig{
			Brokers:      cfg.Kafka.Brokers,
			Topic:        cfg.Kafka.Topic,
			BatchSize:    cfg.Kafka.BatchSize,
			BatchTimeout: cfg.Kafka.BatchTimeout,
			RequiredAcks: 1,
			Async:        true,
		}),
		EventChan: make(chan string, cfg.Kafka.ProducerChannelSize),
		Cfg:       cfg,
	}
}

func (k *KafkaProducer) Write(stop context.Context, logger logger.Logger) {
	defer close(k.EventChan)

	var msgCount int64 = 0

	for {
		select {
		case <-stop.Done():
			logger.Info("Producer worker stopped")
			return
		case slug := <-k.EventChan:
			partition := msgCount % 6
			msgCount++

			kafkaCtx, kafkaCancel := context.WithTimeout(context.Background(), k.Cfg.Kafka.ProducerTimeout)

			err := k.Writer.WriteMessages(kafkaCtx, kafka.Message{
				Value:     []byte(slug),
				Partition: int(partition),
			})

			kafkaCancel()

			if err != nil {
				logger.Error("Kafka Producer error:", err)
			}
		}
	}
}

func (k *KafkaProducer) Close(logger logger.Logger) {
	logger.Info("Closing producer")
	k.Writer.Close()
}
