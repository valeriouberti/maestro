package domain

// BrokerInfo represents the information about a Kafka broker to be returned in the API response.
type BrokerInfo struct {
	ID   int32  `json:"id"`
	Host string `json:"host"`
	Port int    `json:"port"`
}

// TopicInfo represents the information about a Kafka topic to be returned in the API response.
type TopicInfo struct {
	Name              string            `json:"name"`
	NumPartitions     int32             `json:"numPartitions"`
	ReplicationFactor int               `json:"replicationFactor"` // Default replication factor
	Config            map[string]string `json:"config,omitempty"`  // Configuration overrides
	Partitions        []PartitionInfo   `json:"partitions,omitempty"`
}

// PartitionInfo represents information about a specific partition within a topic.
type PartitionInfo struct {
	ID       int32   `json:"id"`
	Leader   int32   `json:"leader"`
	Replicas []int32 `json:"replicas"`
	ISR      []int32 `json:"isr"`
}

// ConsumerGroupInfo represents basic information about a consumer group.
type ConsumerGroupInfo struct {
	GroupID string `json:"groupId"`
	State   string `json:"state,omitempty"` // Example: Stable, Empty, PreparingRebalance
}

// ConsumerGroupInfo represents information about a Kafka consumer group
type ConsumerGroupDetails struct {
	GroupID     string                    `json:"groupId"`
	State       string                    `json:"state"`
	Coordinator BrokerInfo                `json:"coordinator"`
	Members     []ConsumerGroupMemberInfo `json:"members,omitempty"`
	Topics      []string                  `json:"topics,omitempty"`
}

// ConsumerGroupMemberInfo represents a member of a consumer group
type ConsumerGroupMemberInfo struct {
	ClientID    string                     `json:"clientId"`
	ConsumerID  string                     `json:"consumerId"`
	Host        string                     `json:"host"`
	Assignments []TopicPartitionAssignment `json:"assignments,omitempty"`
}

// TopicPartitionAssignment represents a topic-partition assignment to a consumer
type TopicPartitionAssignment struct {
	Topic     string `json:"topic"`
	Partition int32  `json:"partition"`
}
