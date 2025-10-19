package event

import (
	"context"
	"encoding/json"
	"net"
	"strconv"
	"time"

	"github.com/segmentio/kafka-go"
)

type KafkaClient struct {
	host   string
	port   string
	writer *kafka.Writer
}

func NewKafkaClient(host string, port string, topicID string) (*KafkaClient, error) {
	this := &KafkaClient{
		host: host,
		port: port,
		writer: kafka.NewWriter(
			kafka.WriterConfig{
				Brokers: []string{
					net.JoinHostPort(host, port),
				},
				Topic:        topicID,
				BatchTimeout: 1 * time.Millisecond,
			},
		),
	}
	err := this.CreateTopic(topicID)
	if err != nil {
		return nil, err
	}

	return this, nil
}

func (this *KafkaClient) CreateTopic(topicID string) error {
	brokerConnection, err := kafka.Dial("tcp", net.JoinHostPort(this.host, this.port))
	if err != nil {
		return err
	}
	defer brokerConnection.Close()

	controller, err := brokerConnection.Controller()
	if err != nil {
		return err
	}

	controllerConnection, err := kafka.Dial("tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	if err != nil {
		return err
	}
	defer controllerConnection.Close()

	err = controllerConnection.CreateTopics(
		[]kafka.TopicConfig{
			{
				Topic:             topicID,
				NumPartitions:     3,
				ReplicationFactor: 1,
			},
		}...,
	)
	if err != nil {
		return err
	}

	return nil
}

func (this *KafkaClient) WriteMessage(context context.Context, key string, value any) error {
	message, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return this.writer.WriteMessages(
		context,
		kafka.Message{
			Key:   []byte(key),
			Value: message,
		},
	)
}
