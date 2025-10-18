package event

import (
	"context"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

type KafkaClient struct {
	writer *kafka.Writer
}

func NewKafkaClient(host string, port string, topicID string) *KafkaClient {
	return &KafkaClient{
		writer: kafka.NewWriter(
			kafka.WriterConfig{
				Brokers: []string{
					fmt.Sprintf("%s:%s", host, port),
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
