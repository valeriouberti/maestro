package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/valeriouberti/maestro/pkg/kafka"
)

type BrokerInfo struct {
	ID   int32  `json:"id"`
	Host string `json:"host"`
	Port int    `json:"port"`
}

// GetClustersHandler returns cluster information.
func GetClustersHandler(k *kafka.KafkaClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		brokerMetadata, err := k.GetBrokers()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get broker metadata"})
			return
		}

		brokers := make([]BrokerInfo, len(brokerMetadata))
		for i, broker := range brokerMetadata {
			brokers[i] = BrokerInfo{
				ID:   broker.ID,
				Host: broker.Host,
				Port: broker.Port,
			}
		}
		c.JSON(http.StatusOK, gin.H{"brokers": brokers})
	}
}

// ListTopicsHandler returns a list of topics.
func ListTopicsHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "List Topics - Not implemented yet"})
}

// GetTopicHandler returns details for a specific topic.
func GetTopicHandler(c *gin.Context) {
	topicName := c.Param("topicName")
	c.JSON(http.StatusOK, gin.H{"message": "Get Topic: " + topicName + " - Not implemented yet"})
}

// CreateTopicHandler creates a new topic.
func CreateTopicHandler(c *gin.Context) {
	c.JSON(http.StatusCreated, gin.H{"message": "Create Topic - Not implemented yet"})
}

// DeleteTopicHandler deletes a topic.
func DeleteTopicHandler(c *gin.Context) {
	topicName := c.Param("topicName")
	c.JSON(http.StatusOK, gin.H{"message": "Delete Topic: " + topicName + " - Not implemented yet"})
}

// UpdateTopicConfigHandler updates a topic's configuration.
func UpdateTopicConfigHandler(c *gin.Context) {
	topicName := c.Param("topicName")
	c.JSON(http.StatusOK, gin.H{"message": "Update Topic Config: " + topicName + " - Not implemented yet"})
}

// ListConsumerGroupsHandler returns a list of consumer groups.
func ListConsumerGroupsHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "List Consumer Groups - Not implemented yet"})
}

// GetConsumerGroupHandler returns details for a specific consumer group.
func GetConsumerGroupHandler(c *gin.Context) {
	groupID := c.Param("groupId")
	c.JSON(http.StatusOK, gin.H{"message": "Get Consumer Group: " + groupID + " - Not implemented yet"})
}
