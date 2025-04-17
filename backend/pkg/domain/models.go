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
