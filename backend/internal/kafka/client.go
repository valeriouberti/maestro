package kafka

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/valeriouberti/maestro/pkg/domain"
)

// KafkaClient manages interactions with Kafka cluster
type KafkaClient struct {
	AdminClient *kafka.AdminClient
	Brokers     []string
	Timeout     time.Duration // Default timeout for operations
}

// NewKafkaClient creates a new Kafka client with the provided broker addresses
func NewKafkaClient(brokers []string) (*KafkaClient, error) {
	if len(brokers) == 0 {
		return nil, fmt.Errorf("no Kafka brokers provided")
	}

	configMap := &kafka.ConfigMap{
		"bootstrap.servers":     strings.Join(brokers, ","),
		"client.id":             "maestro-client",
		"broker.address.family": "v4", // Force IPv4 resolution,

		"socket.keepalive.enable":  true,
		"reconnect.backoff.ms":     1000,
		"reconnect.backoff.max.ms": 10000,
	}

	adminClient, err := kafka.NewAdminClient(configMap)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka admin client: %w", err)
	}

	return &KafkaClient{
		AdminClient: adminClient,
		Brokers:     brokers,
		Timeout:     10 * time.Second, // Default timeout
	}, nil
}

// Close releases resources used by the Kafka client
func (kc *KafkaClient) Close() {
	if kc.AdminClient != nil {
		kc.AdminClient.Close()
	}
}

// GetBrokers retrieves information about all brokers in the Kafka cluster
func (kc *KafkaClient) GetBrokers(ctx context.Context) ([]domain.BrokerInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, kc.Timeout)
	defer cancel()

	metadata, err := kc.AdminClient.GetMetadata(nil, true, int(kc.Timeout.Milliseconds()))
	if err != nil {
		return nil, fmt.Errorf("failed to get broker metadata: %w", err)
	}

	brokerList := make([]domain.BrokerInfo, 0, len(metadata.Brokers))
	for _, broker := range metadata.Brokers {
		brokerList = append(brokerList, domain.BrokerInfo{
			ID:   broker.ID,
			Host: broker.Host,
			Port: broker.Port,
		})
	}

	return brokerList, nil
}

// ListTopics retrieves information about all topics in the Kafka cluster
func (kc *KafkaClient) ListTopics(ctx context.Context) ([]domain.TopicInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, kc.Timeout)
	defer cancel()

	metadata, err := kc.AdminClient.GetMetadata(nil, true, int(kc.Timeout.Milliseconds()))
	if err != nil {
		return nil, fmt.Errorf("failed to get topic metadata: %w", err)
	}

	topics := make([]domain.TopicInfo, 0, len(metadata.Topics))
	topicNames := make([]string, 0, len(metadata.Topics))

	for topicName := range metadata.Topics {
		topicNames = append(topicNames, topicName)
	}

	sort.Strings(topicNames)

	for _, topicName := range topicNames {
		topicMetadata := metadata.Topics[topicName]
		partitions := make([]domain.PartitionInfo, 0, len(topicMetadata.Partitions))

		replicationFactor := 0
		if len(topicMetadata.Partitions) > 0 {
			for _, partition := range topicMetadata.Partitions {
				partitions = append(partitions, domain.PartitionInfo{
					ID:       partition.ID,
					Leader:   partition.Leader,
					Replicas: partition.Replicas,
					ISR:      partition.Isrs,
				})
			}

			if len(topicMetadata.Partitions[0].Replicas) > 0 {
				replicationFactor = len(topicMetadata.Partitions[0].Replicas)
			}
		}

		topics = append(topics, domain.TopicInfo{
			Name:              topicName,
			NumPartitions:     int32(len(topicMetadata.Partitions)),
			ReplicationFactor: replicationFactor,
			Partitions:        partitions,
		})
	}

	return topics, nil
}

// GetTopicDetails retrieves detailed information about a specific topic
func (kc *KafkaClient) GetTopicDetails(ctx context.Context, topicName string) (*domain.TopicInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, kc.Timeout)
	defer cancel()

	metadata, err := kc.AdminClient.GetMetadata(&topicName, false, int(kc.Timeout.Milliseconds()))
	if err != nil {
		return nil, fmt.Errorf("failed to get topic details: %w", err)
	}

	topicMetadata, exists := metadata.Topics[topicName]
	if !exists {
		return nil, fmt.Errorf("topic '%s' not found", topicName)
	}

	// Get topic config
	configResources := []kafka.ConfigResource{
		{
			Type: kafka.ResourceTopic,
			Name: topicName,
		},
	}

	configResult, err := kc.AdminClient.DescribeConfigs(ctx, configResources)
	if err != nil {
		return nil, fmt.Errorf("failed to get topic configuration: %w", err)
	}

	config := make(map[string]string)
	if len(configResult) > 0 && configResult[0].Config != nil {
		for _, entry := range configResult[0].Config {
			if !entry.IsDefault {
				config[entry.Name] = entry.Value
			}
		}
	}

	// Build partition information
	partitions := make([]domain.PartitionInfo, 0, len(topicMetadata.Partitions))
	for _, partition := range topicMetadata.Partitions {
		partitions = append(partitions, domain.PartitionInfo{
			ID:       partition.ID,
			Leader:   partition.Leader,
			Replicas: partition.Replicas,
			ISR:      partition.Isrs,
		})
	}

	// Determine replication factor from first partition
	replicationFactor := 0
	if len(topicMetadata.Partitions) > 0 && len(topicMetadata.Partitions[0].Replicas) > 0 {
		replicationFactor = len(topicMetadata.Partitions[0].Replicas)
	}

	return &domain.TopicInfo{
		Name:              topicName,
		NumPartitions:     int32(len(topicMetadata.Partitions)),
		ReplicationFactor: replicationFactor,
		Config:            config,
		Partitions:        partitions,
	}, nil
}

// CreateTopic creates a new Kafka topic with the specified configuration
func (kc *KafkaClient) CreateTopic(ctx context.Context, topic domain.TopicInfo) error {
	ctx, cancel := context.WithTimeout(ctx, kc.Timeout)
	defer cancel()

	// Validate input
	if topic.Name == "" {
		return fmt.Errorf("topic name cannot be empty")
	}
	if topic.NumPartitions <= 0 {
		return fmt.Errorf("number of partitions must be greater than 0")
	}
	if topic.ReplicationFactor <= 0 {
		return fmt.Errorf("replication factor must be greater than 0")
	}

	// Create the topic specification
	topicSpec := kafka.TopicSpecification{
		Topic:             topic.Name,
		NumPartitions:     int(topic.NumPartitions),
		ReplicationFactor: int(topic.ReplicationFactor),
		Config:            topic.Config,
	}

	// Create the topic
	topicResults, err := kc.AdminClient.CreateTopics(
		ctx,
		[]kafka.TopicSpecification{topicSpec},
		kafka.SetAdminOperationTimeout(kc.Timeout),
	)

	if err != nil {
		return fmt.Errorf("failed to create topic: %w", err)
	}

	// Check for per-topic errors
	if len(topicResults) > 0 {
		if topicResults[0].Error.Code() != kafka.ErrNoError {
			return fmt.Errorf("failed to create topic '%s': %s",
				topic.Name, topicResults[0].Error.String())
		}
	}

	return nil
}

// DeleteTopic deletes a Kafka topic
func (kc *KafkaClient) DeleteTopic(ctx context.Context, topicName string) error {
	ctx, cancel := context.WithTimeout(ctx, kc.Timeout)
	defer cancel()

	// Validate input
	if topicName == "" {
		return fmt.Errorf("topic name cannot be empty")
	}

	// Check if topic exists before attempting to delete it
	metadata, err := kc.AdminClient.GetMetadata(&topicName, false, int(kc.Timeout.Milliseconds()))
	if err != nil {
		return fmt.Errorf("failed to check if topic exists: %w", err)
	}

	if _, exists := metadata.Topics[topicName]; !exists {
		return fmt.Errorf("topic '%s' not found", topicName)
	}

	// Delete the topic
	topicResults, err := kc.AdminClient.DeleteTopics(
		ctx,
		[]string{topicName},
		kafka.SetAdminOperationTimeout(kc.Timeout),
	)

	if err != nil {
		return fmt.Errorf("failed to delete topic: %w", err)
	}

	// Check for per-topic errors
	if len(topicResults) > 0 {
		if topicResults[0].Error.Code() != kafka.ErrNoError {
			return fmt.Errorf("failed to delete topic '%s': %s",
				topicName, topicResults[0].Error.String())
		}
	}

	return nil
}
