package api

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/gin-gonic/gin"

	"github.com/valeriouberti/maestro/internal/kafka_client"
	"github.com/valeriouberti/maestro/pkg/domain"
)

// ErrorResponse represents a standardized error response structure for API endpoints.
// It contains HTTP status code, a general error message, and optional detailed information
// about the error that can be omitted when empty.
type ErrorResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Detail  string `json:"detail,omitempty"`
}

// TopicCreationRequest contains the parameters needed to create a new Kafka topic.
// It validates that essential fields are provided through JSON binding tags.
//
// Fields:
//   - Name: The name of the Kafka topic to create (required)
//   - NumPartitions: The number of partitions for the topic, must be at least 1 (required)
//   - ReplicationFactor: The replication factor for the topic, must be at least 1 (required)
//   - Config: Optional map of configuration parameters for the topic with string keys and values
type TopicCreationRequest struct {
	Name              string            `json:"name" binding:"required"`
	NumPartitions     int32             `json:"numPartitions" binding:"required,min=1"`
	ReplicationFactor int16             `json:"replicationFactor" binding:"required,min=1"`
	Config            map[string]string `json:"config,omitempty"`
}

// TopicConfigUpdateRequest represents a request to update configuration for a topic.
// It contains a map of configuration key-value pairs that should be applied to the topic.
// The configuration map is required for the request to be valid.
type TopicConfigUpdateRequest struct {
	Config map[string]string `json:"config" binding:"required"`
}

// GetClustersHandler returns a Gin HTTP handler that retrieves Kafka broker metadata.
// It takes a Kafka client as input and returns a handler function that:
//   - Fetches broker metadata from Kafka
//   - Returns the metadata as JSON on success
//   - Returns an appropriate error response if the operation fails
//
// Parameters:
//   - k: A pointer to a KafkaClient that provides access to Kafka cluster information
//
// Returns:
//   - A Gin handler function that handles HTTP requests for Kafka cluster information
func GetClustersHandler(k *kafka_client.KafkaClient) gin.HandlerFunc {
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

// ListTopicsHandler creates a gin HTTP handler for retrieving Kafka topics.
// It takes a Kafka client and returns a handler function that:
//   - Fetches all available topics from Kafka
//   - Returns the topics as JSON with a 200 OK status on success
//   - Returns a 500 Internal Server Error with error details if the operation fails
//
// Parameters:
//   - k: A pointer to a kafka.KafkaClient used to interact with Kafka
//
// Returns:
//   - A gin.HandlerFunc that handles HTTP requests for listing Kafka topics
func ListTopicsHandler(k *kafka_client.KafkaClient) gin.HandlerFunc {
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

// GetTopicHandler returns a HTTP handler function that retrieves details for a specific Kafka topic.
//
// It creates a gin.HandlerFunc that:
// - Extracts the topic name from URL parameters
// - Validates that the topic name is provided
// - Fetches topic details using the provided Kafka client
// - Returns the topic details as JSON on success
// - Returns appropriate error responses when the topic name is missing or when the fetch operation fails
//
// Parameters:
//   - k: A pointer to a Kafka client used to retrieve topic information
//
// Returns:
//   - A gin.HandlerFunc that handles HTTP requests for Kafka topic details
func GetTopicHandler(k *kafka_client.KafkaClient) gin.HandlerFunc {
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

// CreateTopicHandler returns a Gin handler function that processes HTTP requests for Kafka topic creation.
//
// The handler accepts JSON requests containing topic details (name, partition count, replication factor, and optional config),
// validates the input, and attempts to create the topic via the provided Kafka client.
//
// It returns appropriate HTTP responses based on the operation result:
// - 201 Created: When the topic is successfully created, including the topic details
// - 400 Bad Request: When the request JSON is invalid or malformed
// - 409 Conflict: When the topic already exists
// - 500 Internal Server Error: When the topic creation fails for other reasons
//
// Parameters:
//   - k: A pointer to a Kafka client that handles the actual topic creation
//
// Returns:
//   - A Gin handler function that processes the HTTP request and generates the appropriate response
func CreateTopicHandler(k *kafka_client.KafkaClient) gin.HandlerFunc {
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

// DeleteTopicHandler creates a Gin handler for deleting Kafka topics.
// It takes a Kafka client and returns a handler function that:
// - Extracts the topic name from the URL path parameter
// - Validates the topic name is not empty
// - Attempts to delete the topic using the Kafka client
// - Returns appropriate HTTP responses:
//   - 200 OK with success message when topic is deleted successfully
//   - 400 Bad Request when topic name is missing
//   - 404 Not Found when the topic doesn't exist
//   - 500 Internal Server Error for other failures
//
// Parameters:
//   - k: A pointer to a kafka.KafkaClient instance used for topic operations
//
// Returns:
//   - A Gin handler function for the DELETE topic endpoint
func DeleteTopicHandler(k *kafka_client.KafkaClient) gin.HandlerFunc {
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

// UpdateTopicConfigHandler creates a gin HTTP handler for updating Kafka topic configurations.
// It accepts a KafkaClient instance to interact with the Kafka cluster.
//
// The handler expects a topic name as a URL parameter and a JSON request body with
// the following structure:
//
//	{
//	    "config": {
//	        "key1": "value1",
//	        "key2": "value2",
//	        ...
//	    }
//	}
//
// HTTP Responses:
// - 200 OK: Configuration updated successfully, returns the updated topic details
// - 400 Bad Request: Missing topic name, invalid request format, or empty configuration
// - 404 Not Found: Topic doesn't exist in the Kafka cluster
// - 500 Internal Server Error: Failed to update topic configuration
//
// If the update succeeds but retrieving updated details fails, it still returns 200 OK
// with a success message and the requested configuration changes.
func UpdateTopicConfigHandler(k *kafka_client.KafkaClient) gin.HandlerFunc {
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

// ListConsumerGroupsHandler returns a Gin HTTP handler function that lists all Kafka consumer groups.
// It takes a Kafka client as input and when the handler is called, it queries for all consumer groups
// from the Kafka cluster.
//
// If successful, it returns a JSON response with HTTP 200 status code containing the list of consumer groups.
// If an error occurs during the operation, it returns a JSON error response with HTTP 500 status code
// along with the error details.
//
// Parameters:
//   - k: A pointer to a KafkaClient instance used to interact with the Kafka cluster
//
// Returns:
//   - A Gin handler function that processes HTTP requests for listing consumer groups
func ListConsumerGroupsHandler(k *kafka_client.KafkaClient) gin.HandlerFunc {
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

// GetConsumerGroupHandler creates an HTTP handler for retrieving details of a Kafka consumer group.
// It accepts a Kafka client instance and returns a Gin handler function that:
//   - Extracts the consumer group ID from the URL path parameter "groupId"
//   - Validates that the group ID is provided
//   - Fetches the consumer group details from Kafka
//   - Returns appropriate HTTP responses based on the result:
//   - 200 OK with group details on success
//   - 400 Bad Request if group ID is missing
//   - 404 Not Found if the consumer group doesn't exist
//   - 500 Internal Server Error for other failures
//
// Parameters:
//   - k: A pointer to a Kafka client used to interact with Kafka
//
// Returns:
//   - A Gin HTTP handler function
func GetConsumerGroupHandler(k *kafka_client.KafkaClient) gin.HandlerFunc {
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

// GetTopicMessagesHandler returns a HTTP handler function that retrieves messages from a specific Kafka topic.
//
// It extracts parameters from the request including:
// - topicName: From the URL path
// - partition: Query parameter for the partition to consume from (defaults to 0)
// - offset: Query parameter for the starting offset (defaults to -1 for newest)
// - limit: Query parameter for the maximum number of messages to retrieve (defaults to 100)
//
// Returns:
// - 200 OK with the messages on success
// - 400 Bad Request if the topic name is missing or parameters are invalid
// - 404 Not Found if the topic doesn't exist
// - 500 Internal Server Error for other failures
func GetTopicMessagesHandler(k *kafka_client.KafkaClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		topicName := c.Param("topicName")
		if topicName == "" {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Status:  http.StatusBadRequest,
				Message: "Topic name is required",
			})
			return
		}

		// Parse optional query parameters with defaults
		partition := int32(0)
		if partitionStr := c.Query("partition"); partitionStr != "" {
			partitionInt, err := strconv.ParseInt(partitionStr, 10, 32)
			if err != nil {
				c.JSON(http.StatusBadRequest, ErrorResponse{
					Status:  http.StatusBadRequest,
					Message: "Invalid partition parameter",
					Detail:  err.Error(),
				})
				return
			}
			partition = int32(partitionInt)
		}

		offset := kafka.OffsetBeginning
		if offsetStr := c.Query("offset"); offsetStr != "" {
			// Special case for "latest" offset
			if offsetStr == "latest" {
				offset = kafka.OffsetEnd
			} else {
				offsetInt, err := strconv.ParseInt(offsetStr, 10, 64)
				if err != nil {
					c.JSON(http.StatusBadRequest, ErrorResponse{
						Status:  http.StatusBadRequest,
						Message: "Invalid offset parameter",
						Detail:  err.Error(),
					})
					return
				}
				offset = kafka.Offset(offsetInt)
			}
		}

		limit := 100 // Default limit
		if limitStr := c.Query("limit"); limitStr != "" {
			limitInt, err := strconv.Atoi(limitStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, ErrorResponse{
					Status:  http.StatusBadRequest,
					Message: "Invalid limit parameter",
					Detail:  err.Error(),
				})
				return
			}
			if limitInt <= 0 {
				c.JSON(http.StatusBadRequest, ErrorResponse{
					Status:  http.StatusBadRequest,
					Message: "Limit must be a positive integer",
				})
				return
			}
			limit = limitInt
		}

		messages, err := k.GetTopicMessages(c.Request.Context(), topicName, partition, int64(offset), limit)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				c.JSON(http.StatusNotFound, ErrorResponse{
					Status:  http.StatusNotFound,
					Message: "Topic or partition not found",
					Detail:  err.Error(),
				})
				return
			}

			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Status:  http.StatusInternalServerError,
				Message: "Failed to retrieve messages",
				Detail:  err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"topic":     topicName,
			"partition": partition,
			"offset":    offset,
			"count":     len(messages),
			"messages":  messages,
		})
	}
}
