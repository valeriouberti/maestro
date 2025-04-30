package kafka_client

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/google/uuid"
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
func (kc *KafkaClient) GetConsumerGroupDetails(ctx context.Context, groupID string) (*domain.ConsumerGroupDetails, error) {
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

	groupInfo := &domain.ConsumerGroupDetails{
		GroupID:     groupID,
		State:       string(rune(group.State)),
		Coordinator: coordinator,
		Members:     members,
		Topics:      topics,
	}

	return groupInfo, nil
}

// GetTopicMessages retrieves messages from a specified topic and partition
// GetTopicMessages retrieves messages from a specified topic and partition
func (kc *KafkaClient) GetTopicMessages(ctx context.Context, topicName string, partition int32, offset int64, limit int) ([]domain.TopicMessage, error) {
	// Create a context with extended timeout for this operation specifically
	ctx, cancel := context.WithTimeout(ctx, kc.Timeout*2) // Double the timeout for message retrieval
	defer cancel()

	if topicName == "" {
		return nil, fmt.Errorf("topic name cannot be empty")
	}

	// Validate topic exists
	metadata, err := kc.AdminClient.GetMetadata(&topicName, false, int(kc.Timeout.Milliseconds()))
	if err != nil {
		return nil, fmt.Errorf("failed to check if topic exists: %w", err)
	}

	topicMetadata, exists := metadata.Topics[topicName]
	if !exists {
		return nil, fmt.Errorf("topic '%s' not found", topicName)
	}

	// Validate partition exists
	partitionExists := false
	for partID := range topicMetadata.Partitions {
		if int32(partID) == partition {
			partitionExists = true
			break
		}
	}
	if !partitionExists && partition != -1 {
		return nil, fmt.Errorf("partition %d does not exist for topic '%s'", partition, topicName)
	}

	// For "latest" offset, first get the current high watermark to use as starting point
	if offset == int64(kafka.OffsetEnd) {
		// We'll use the admin client to check topic offsets first
		offsets, err := kc.getPartitionOffsets(ctx, topicName, partition)
		if err != nil {
			return nil, fmt.Errorf("failed to get offset information: %w", err)
		}

		// If partition is empty, return empty results immediately
		if offsets.high <= offsets.low {
			return []domain.TopicMessage{}, nil
		}

		// For "latest" offset, we'll take a much smaller window to ensure we get newer messages
		var startOffset int64

		// If limit is very large, use a smaller value to ensure we get recent messages
		adjustedLimit := int64(limit)
		if adjustedLimit > 100 {
			adjustedLimit = 100
		}

		// Calculate starting offset to get newest messages (use a smaller window)
		startOffset = offsets.high - adjustedLimit
		if startOffset < offsets.low {
			startOffset = offsets.low
		}

		// Use this calculated offset instead of "latest"
		offset = startOffset
	}

	// Create a consumer configuration with more robust settings
	config := &kafka.ConfigMap{
		"bootstrap.servers":         strings.Join(kc.Brokers, ","),
		"group.id":                  "maestro-message-reader-" + uuid.New().String(),
		"auto.offset.reset":         "earliest", // Use earliest as the default
		"enable.auto.commit":        false,
		"socket.keepalive.enable":   true,
		"session.timeout.ms":        10000,   // 10 seconds
		"max.poll.interval.ms":      30000,   // 30 seconds
		"socket.timeout.ms":         10000,   // 10 seconds
		"message.max.bytes":         1048576, // 1MB
		"fetch.max.bytes":           5242880, // 5MB (must be >= message.max.bytes)
		"receive.message.max.bytes": 5243392, // 5MB + 512 (must be >= fetch.max.bytes + 512)
	}

	consumer, err := kafka.NewConsumer(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka consumer: %w", err)
	}
	defer func() {
		if err := consumer.Close(); err != nil {
			fmt.Printf("Error closing Kafka consumer: %v", err)
		}
	}()

	// Assign partition with the explicit offset (which may have been calculated above)
	err = consumer.Assign([]kafka.TopicPartition{
		{
			Topic:     &topicName,
			Partition: partition,
			Offset:    kafka.Offset(offset),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to assign partition: %w", err)
	}

	messages := make([]domain.TopicMessage, 0, limit)
	deadline := time.Now().Add(kc.Timeout)
	messageCount := 0
	emptyPollCount := 0
	maxEmptyPolls := 5 // Be more aggressive in returning

	for messageCount < limit && time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return messages, ctx.Err()
		default:
			ev := consumer.Poll(200) // Increased poll timeout for better throughput
			if ev == nil {
				emptyPollCount++
				// If we've had several empty polls and already have some messages, return them
				if emptyPollCount >= maxEmptyPolls && messageCount > 0 {
					return messages, nil
				}
				continue
			}

			// Reset empty poll counter when we get an event
			emptyPollCount = 0

			switch e := ev.(type) {
			case *kafka.Message:
				var messageKey, messageValue string

				// Safely handle message key and value
				if e.Key != nil {
					messageKey = string(e.Key)
				}

				if e.Value != nil {
					messageValue = string(e.Value)
				}

				message := domain.TopicMessage{
					Topic:     *e.TopicPartition.Topic,
					Partition: e.TopicPartition.Partition,
					Offset:    int64(e.TopicPartition.Offset),
					Timestamp: e.Timestamp,
					Key:       messageKey,
					Value:     messageValue,
					Headers:   make(map[string]string),
				}

				// Extract headers if any
				for _, header := range e.Headers {
					message.Headers[header.Key] = string(header.Value)
				}

				messages = append(messages, message)
				messageCount++

				if messageCount >= limit {
					return messages, nil
				}
			case kafka.Error:
				// Don't fail immediately on timeouts or transient errors
				kafkaErr := e.Code()
				if kafkaErr == kafka.ErrTimedOut ||
					kafkaErr == kafka.ErrTransport ||
					kafkaErr == kafka.ErrBrokerNotAvailable {
					// Log but continue
					fmt.Printf("Recoverable Kafka error: %v\n", e)
					continue
				}

				return messages, fmt.Errorf("consumer error: %v", e)
			}
		}
	}

	// Return what we have if we reach here (timeout or done)
	return messages, nil
}

// Helper method to get partition offsets without creating a consumer
type partitionOffsets struct {
	low  int64
	high int64
}

func (kc *KafkaClient) getPartitionOffsets(ctx context.Context, topicName string, partition int32) (partitionOffsets, error) {
	result := partitionOffsets{}

	// Create a lightweight consumer just to get offsets
	config := &kafka.ConfigMap{
		"bootstrap.servers": strings.Join(kc.Brokers, ","),
		"group.id":          "maestro-offset-checker",
	}

	c, err := kafka.NewConsumer(config)
	if err != nil {
		return result, fmt.Errorf("failed to create offset checker consumer: %w", err)
	}
	defer c.Close()

	// Get low and high watermarks
	low, high, err := c.GetWatermarkOffsets(topicName, partition)
	if err != nil {
		return result, fmt.Errorf("failed to get watermark offsets: %w", err)
	}

	result.low = low
	result.high = high
	return result, nil
}

// PublishMessage publishes a message to a specified Kafka topic
func (kc *KafkaClient) PublishMessage(ctx context.Context, topicName string, partition int32, key string, value string, headers map[string]string) error {
	ctx, cancel := context.WithTimeout(ctx, kc.Timeout)
	defer cancel()

	if topicName == "" {
		return fmt.Errorf("topic name cannot be empty")
	}

	// Validate topic exists
	metadata, err := kc.AdminClient.GetMetadata(&topicName, false, int(kc.Timeout.Milliseconds()))
	if err != nil {
		return fmt.Errorf("failed to check if topic exists: %w", err)
	}

	topicMetadata, exists := metadata.Topics[topicName]
	if !exists {
		return fmt.Errorf("topic '%s' not found", topicName)
	}

	// Validate partition exists if specified
	if partition >= 0 {
		partitionExists := false
		for partID := range topicMetadata.Partitions {
			if int32(partID) == partition {
				partitionExists = true
				break
			}
		}
		if !partitionExists {
			return fmt.Errorf("partition %d does not exist for topic '%s'", partition, topicName)
		}
	}

	// Create producer
	config := &kafka.ConfigMap{
		"bootstrap.servers": strings.Join(kc.Brokers, ","),
		"group.id":          "maestro-message-producer",
		"acks":              "all", // Wait for all replicas
	}

	producer, err := kafka.NewProducer(config)
	if err != nil {
		return fmt.Errorf("failed to create Kafka producer: %w", err)
	}
	defer producer.Close()

	// Create message headers if any
	var kafkaHeaders []kafka.Header
	if len(headers) > 0 {
		kafkaHeaders = make([]kafka.Header, 0, len(headers))
		for k, v := range headers {
			kafkaHeaders = append(kafkaHeaders, kafka.Header{
				Key:   k,
				Value: []byte(v),
			})
		}
	}

	// Prepare the message
	message := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &topicName,
			Partition: partition,
		},
		Key:     []byte(key),
		Value:   []byte(value),
		Headers: kafkaHeaders,
	}

	// Set up delivery channel to receive delivery reports
	deliveryChan := make(chan kafka.Event, 1)
	defer close(deliveryChan)

	// Produce the message
	err = producer.Produce(message, deliveryChan)
	if err != nil {
		return fmt.Errorf("failed to produce message: %w", err)
	}

	// Wait for delivery report or context cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	case e := <-deliveryChan:
		m := e.(*kafka.Message)
		if m.TopicPartition.Error != nil {
			return fmt.Errorf("message delivery failed: %v", m.TopicPartition.Error)
		}
		// Return the partition and offset where the message was stored
		return nil
	}
}
