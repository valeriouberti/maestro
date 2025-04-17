package kafka

import (
	"context"
	"fmt"
	"strings"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type KafkaClient struct {
	AdminClient *kafka.AdminClient
	Brokers     []string
}

func NewKafkaClient(brokers []string) (*KafkaClient, error) {

	configMap := &kafka.ConfigMap{
		"bootstrap.servers": strings.Join(brokers, ","),
	}

	adminClient, err := kafka.NewAdminClient(configMap)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka admin client: %v", err)
	}

	return &KafkaClient{
		AdminClient: adminClient,
		Brokers:     brokers,
	}, nil
}

func (kc *KafkaClient) Close() {
	if kc.AdminClient != nil {
		kc.AdminClient.Close()
	}
}

func (kc *KafkaClient) GetBrokers() ([]kafka.BrokerMetadata, error) {

	_, cancel := context.WithTimeout(context.Background(), 5)
	defer cancel()

	metadata, err := kc.AdminClient.GetMetadata(nil, true, 5000)
	if err != nil {
		return nil, fmt.Errorf("failed to get metadata: %v", err)
	}

	brokerList := make([]kafka.BrokerMetadata, 0, len(metadata.Brokers))
	brokerList = append(brokerList, metadata.Brokers...)

	return brokerList, nil
}
