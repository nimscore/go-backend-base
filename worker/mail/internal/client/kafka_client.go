package client

import (
	"context"
	"fmt"
	"time"

	kafka "github.com/segmentio/kafka-go"
)

type KafkaClient struct {
	reader *kafka.Reader
}

func NewKafkaClient(host string, port string, topicID string, groupID string) *KafkaClient {
	return &KafkaClient{
		reader: kafka.NewReader(
			kafka.ReaderConfig{
				Brokers: []string{
					fmt.Sprintf("%s:%s", host, port),
				},
				GroupID:  groupID,
				Topic:    topicID,
				MaxWait:  500 * time.Millisecond,
				MinBytes: 1,
				MaxBytes: 1024 * 1024,
			},
		),
	}
}

func (this *KafkaClient) Read(context context.Context) (string, string, error) {
	message, err := this.reader.ReadMessage(context)
	if err != nil {
		return "", "", err
	}

	return string(message.Key), string(message.Value), nil
}
