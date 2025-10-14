package client

import (
	"context"
	"time"

	"github.com/segmentio/kafka-go"
)

type KafkaClient struct {
	writer *kafka.Writer
}

func NewKafkaClient(server string, topicID string) *KafkaClient {
	return &KafkaClient{
		writer: kafka.NewWriter(
			kafka.WriterConfig{
				Brokers: []string{
					server,
				},
				Topic:        topicID,
				BatchTimeout: 1 * time.Millisecond,
			},
		),
	}
}

func (this *KafkaClient) Write(context context.Context, key string, value string) error {
	return this.writer.WriteMessages(
		context,
		kafka.Message{
			Key:   []byte(key),
			Value: []byte(value),
		},
	)
}
