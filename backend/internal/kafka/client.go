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
		"bootstrap.servers": strings.Join(brokers, ","),
		"client.id":         "maestro-client",
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

	partitions := make([]domain.PartitionInfo, 0, len(topicMetadata.Partitions))
	for _, partition := range topicMetadata.Partitions {
		partitions = append(partitions, domain.PartitionInfo{
			ID:       partition.ID,
			Leader:   partition.Leader,
			Replicas: partition.Replicas,
			ISR:      partition.Isrs,
		})
	}

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

	if topic.Name == "" {
		return fmt.Errorf("topic name cannot be empty")
	}
	if topic.NumPartitions <= 0 {
		return fmt.Errorf("number of partitions must be greater than 0")
	}
	if topic.ReplicationFactor <= 0 {
		return fmt.Errorf("replication factor must be greater than 0")
	}

	topicSpec := kafka.TopicSpecification{
		Topic:             topic.Name,
		NumPartitions:     int(topic.NumPartitions),
		ReplicationFactor: int(topic.ReplicationFactor),
		Config:            topic.Config,
	}

	topicResults, err := kc.AdminClient.CreateTopics(
		ctx,
		[]kafka.TopicSpecification{topicSpec},
		kafka.SetAdminOperationTimeout(kc.Timeout),
	)

	if err != nil {
		return fmt.Errorf("failed to create topic: %w", err)
	}

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

	if topicName == "" {
		return fmt.Errorf("topic name cannot be empty")
	}

	metadata, err := kc.AdminClient.GetMetadata(&topicName, false, int(kc.Timeout.Milliseconds()))
	if err != nil {
		return fmt.Errorf("failed to check if topic exists: %w", err)
	}

	if _, exists := metadata.Topics[topicName]; !exists {
		return fmt.Errorf("topic '%s' not found", topicName)
	}

	topicResults, err := kc.AdminClient.DeleteTopics(
		ctx,
		[]string{topicName},
		kafka.SetAdminOperationTimeout(kc.Timeout),
	)

	if err != nil {
		return fmt.Errorf("failed to delete topic: %w", err)
	}

	if len(topicResults) > 0 {
		if topicResults[0].Error.Code() != kafka.ErrNoError {
			return fmt.Errorf("failed to delete topic '%s': %s",
				topicName, topicResults[0].Error.String())
		}
	}

	return nil
}

// UpdateTopicConfig updates the configuration of an existing Kafka topic
func (kc *KafkaClient) UpdateTopicConfig(ctx context.Context, topicName string, config map[string]string) error {
	ctx, cancel := context.WithTimeout(ctx, kc.Timeout)
	defer cancel()

	if topicName == "" {
		return fmt.Errorf("topic name cannot be empty")
	}
	if len(config) == 0 {
		return fmt.Errorf("no configuration provided")
	}

	metadata, err := kc.AdminClient.GetMetadata(&topicName, false, int(kc.Timeout.Milliseconds()))
	if err != nil {
		return fmt.Errorf("failed to check if topic exists: %w", err)
	}

	if _, exists := metadata.Topics[topicName]; !exists {
		return fmt.Errorf("topic '%s' not found", topicName)
	}

	configEntries := make([]kafka.ConfigEntry, 0, len(config))
	for key, value := range config {
		configEntries = append(configEntries, kafka.ConfigEntry{
			Name:  key,
			Value: value,
		})
	}

	configResource := kafka.ConfigResource{
		Type:   kafka.ResourceTopic,
		Name:   topicName,
		Config: configEntries,
	}

	result, err := kc.AdminClient.AlterConfigs(
		ctx,
		[]kafka.ConfigResource{configResource},
	)

	if err != nil {
		return fmt.Errorf("failed to update topic configuration: %w", err)
	}

	if len(result) > 0 {
		if result[0].Error.Code() != kafka.ErrNoError {
			return fmt.Errorf("failed to update topic configuration: %s", result[0].Error.String())
		}
	}

	return nil
}

// ListConsumerGroups retrieves a list of all consumer groups in the Kafka cluster
func (kc *KafkaClient) ListConsumerGroups(ctx context.Context) ([]domain.ConsumerGroupInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, kc.Timeout)
	defer cancel()

	groupList, err := kc.AdminClient.ListConsumerGroups(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list consumer groups: %w", err)
	}

	groups := make([]domain.ConsumerGroupInfo, 0)
	for _, group := range groupList.Valid {
		groups = append(groups, domain.ConsumerGroupInfo{
			GroupID: group.GroupID,
			State:   "", // Basic listing doesn't include state
		})
	}

	sort.Slice(groups, func(i, j int) bool {
		return groups[i].GroupID < groups[j].GroupID
	})

	return groups, nil
}

// GetConsumerGroupDetails retrieves detailed information about a specific consumer group
func (kc *KafkaClient) GetConsumerGroupDetails(ctx context.Context, groupID string) (*domain.ConsumerGroupDatails, error) {
	ctx, cancel := context.WithTimeout(ctx, kc.Timeout)
	defer cancel()

	if groupID == "" {
		return nil, fmt.Errorf("consumer group ID cannot be empty")
	}

	groups, err := kc.AdminClient.DescribeConsumerGroups(
		ctx,
		[]string{groupID},
		// Using context timeout instead of operation-specific timeout
	)
	if err != nil {
		return nil, fmt.Errorf("failed to describe consumer group: %w", err)
	}

	if len(groups.ConsumerGroupDescriptions) == 0 {
		return nil, fmt.Errorf("consumer group '%s' not found", groupID)
	}

	group := groups.ConsumerGroupDescriptions[0]
	if group.Error.Code() != kafka.ErrNoError {
		if group.Error.Code() == kafka.ErrGroupIDNotFound {
			return nil, fmt.Errorf("consumer group '%s' not found", groupID)
		}
		return nil, fmt.Errorf("failed to get consumer group details: %s", group.Error.String())
	}

	members := make([]domain.ConsumerGroupMemberInfo, 0, len(group.Members))
	subscribedTopics := make(map[string]bool)

	for _, member := range group.Members {
		assignments := make([]domain.TopicPartitionAssignment, 0, len(member.Assignment.TopicPartitions))

		for _, assignment := range member.Assignment.TopicPartitions {
			assignments = append(assignments, domain.TopicPartitionAssignment{
				Topic:     *assignment.Topic,
				Partition: assignment.Partition,
			})
		}

		members = append(members, domain.ConsumerGroupMemberInfo{
			ClientID:    member.ClientID,
			ConsumerID:  member.ConsumerID,
			Host:        member.Host,
			Assignments: assignments,
		})
	}

	topics := make([]string, 0, len(subscribedTopics))
	for topic := range subscribedTopics {
		topics = append(topics, topic)
	}
	sort.Strings(topics)

	coordinator := domain.BrokerInfo{
		ID:   int32(group.Coordinator.ID),
		Host: group.Coordinator.Host,
		Port: group.Coordinator.Port,
	}

	groupInfo := &domain.ConsumerGroupDatails{
		GroupID:     groupID,
		State:       string(group.State),
		Coordinator: coordinator,
		Members:     members,
		Topics:      topics,
	}

	return groupInfo, nil
}
