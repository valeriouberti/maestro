package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/valeriouberti/maestro/internal/kafka"
	"github.com/valeriouberti/maestro/pkg/domain"
)

// ErrorResponse standardizes API error responses
type ErrorResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Detail  string `json:"detail,omitempty"`
}

// GetClustersHandler returns information about Kafka brokers
func GetClustersHandler(k *kafka.KafkaClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		brokerMetadata, err := k.GetBrokers(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Status:  http.StatusInternalServerError,
				Message: "Failed to get broker metadata",
				Detail:  err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{"brokers": brokerMetadata})
	}
}

// ListTopicsHandler returns a list of all Kafka topics
func ListTopicsHandler(k *kafka.KafkaClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		topics, err := k.ListTopics(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Status:  http.StatusInternalServerError,
				Message: "Failed to list topics",
				Detail:  err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{"topics": topics})
	}
}

// GetTopicHandler returns details for a specific topic
func GetTopicHandler(k *kafka.KafkaClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		topicName := c.Param("topicName")
		if topicName == "" {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Status:  http.StatusBadRequest,
				Message: "Topic name is required",
			})
			return
		}

		topic, err := k.GetTopicDetails(c.Request.Context(), topicName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Status:  http.StatusInternalServerError,
				Message: "Failed to get topic details",
				Detail:  err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{"topic": topic})
	}
}

// TopicCreationRequest defines the structure for creating a new topic
type TopicCreationRequest struct {
	Name              string            `json:"name" binding:"required"`
	NumPartitions     int32             `json:"numPartitions" binding:"required,min=1"`
	ReplicationFactor int16             `json:"replicationFactor" binding:"required,min=1"`
	Config            map[string]string `json:"config,omitempty"`
}

// CreateTopicHandler handles requests to create a new Kafka topic
func CreateTopicHandler(k *kafka.KafkaClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request TopicCreationRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request: " + err.Error(),
			})
			return
		}

		topicInfo := domain.TopicInfo{
			Name:              request.Name,
			NumPartitions:     request.NumPartitions,
			ReplicationFactor: int(request.ReplicationFactor),
			Config:            request.Config,
		}

		err := k.CreateTopic(c.Request.Context(), topicInfo)
		if err != nil {
			if strings.Contains(err.Error(), "already exists") {
				c.JSON(http.StatusConflict, gin.H{
					"error": "Topic already exists: " + err.Error(),
				})
				return
			}

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to create topic: " + err.Error(),
			})
			return
		}

		topic, err := k.GetTopicDetails(c.Request.Context(), request.Name)
		if err != nil {
			c.JSON(http.StatusCreated, gin.H{
				"message": "Topic created successfully",
				"topic": gin.H{
					"name":              request.Name,
					"numPartitions":     request.NumPartitions,
					"replicationFactor": request.ReplicationFactor,
				},
			})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "Topic created successfully",
			"topic":   topic,
		})
	}
}

// DeleteTopicHandler deletes a Kafka topic
func DeleteTopicHandler(k *kafka.KafkaClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		topicName := c.Param("topicName")
		if topicName == "" {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Status:  http.StatusBadRequest,
				Message: "Topic name is required",
			})
			return
		}

		err := k.DeleteTopic(c.Request.Context(), topicName)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				c.JSON(http.StatusNotFound, ErrorResponse{
					Status:  http.StatusNotFound,
					Message: "Topic not found",
					Detail:  err.Error(),
				})
				return
			}

			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Status:  http.StatusInternalServerError,
				Message: "Failed to delete topic",
				Detail:  err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Topic deleted successfully",
			"topic":   topicName,
		})
	}
}

// TopicConfigUpdateRequest defines the structure for updating a topic's configuration
type TopicConfigUpdateRequest struct {
	Config map[string]string `json:"config" binding:"required"`
}

// UpdateTopicConfigHandler updates a topic's configuration
func UpdateTopicConfigHandler(k *kafka.KafkaClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		topicName := c.Param("topicName")
		if topicName == "" {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Status:  http.StatusBadRequest,
				Message: "Topic name is required",
			})
			return
		}

		var request TopicConfigUpdateRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Status:  http.StatusBadRequest,
				Message: "Invalid configuration update request",
				Detail:  err.Error(),
			})
			return
		}

		// Validate that config is not empty
		if len(request.Config) == 0 {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Status:  http.StatusBadRequest,
				Message: "Configuration cannot be empty",
			})
			return
		}

		// Update the topic configuration
		err := k.UpdateTopicConfig(c.Request.Context(), topicName, request.Config)
		if err != nil {
			// Check for specific error types
			if strings.Contains(err.Error(), "not found") {
				c.JSON(http.StatusNotFound, ErrorResponse{
					Status:  http.StatusNotFound,
					Message: "Topic not found",
					Detail:  err.Error(),
				})
				return
			}

			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Status:  http.StatusInternalServerError,
				Message: "Failed to update topic configuration",
				Detail:  err.Error(),
			})
			return
		}

		// Get updated topic details to return in the response
		topic, err := k.GetTopicDetails(c.Request.Context(), topicName)
		if err != nil {
			// Still return success even if we can't retrieve the updated details
			c.JSON(http.StatusOK, gin.H{
				"message": "Topic configuration updated successfully",
				"topic": gin.H{
					"name":   topicName,
					"config": request.Config,
				},
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Topic configuration updated successfully",
			"topic":   topic,
		})
	}
}

// ListConsumerGroupsHandler returns a list of consumer groups
func ListConsumerGroupsHandler(k *kafka.KafkaClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		groups, err := k.ListConsumerGroups(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Status:  http.StatusInternalServerError,
				Message: "Failed to list consumer groups",
				Detail:  err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"groups": groups,
		})
	}
}

// GetConsumerGroupHandler returns details for a specific consumer group
func GetConsumerGroupHandler(k *kafka.KafkaClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		groupID := c.Param("groupId")
		if groupID == "" {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Status:  http.StatusBadRequest,
				Message: "Consumer group ID is required",
			})
			return
		}

		group, err := k.GetConsumerGroupDetails(c.Request.Context(), groupID)
		if err != nil {
			// Check for common errors
			if strings.Contains(err.Error(), "not found") {
				c.JSON(http.StatusNotFound, ErrorResponse{
					Status:  http.StatusNotFound,
					Message: "Consumer group not found",
					Detail:  err.Error(),
				})
				return
			}

			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Status:  http.StatusInternalServerError,
				Message: "Failed to get consumer group details",
				Detail:  err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"group": group,
		})
	}
}
