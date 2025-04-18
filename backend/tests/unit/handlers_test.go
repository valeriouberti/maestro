// package unit

// import (
// 	"bytes"
// 	"context"
// 	"encoding/json"
// 	"errors"
// 	"net/http"
// 	"net/http/httptest"
// 	"testing"

// 	"github.com/gin-gonic/gin"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/mock"
// 	"github.com/valeriouberti/maestro/pkg/api"
// 	"github.com/valeriouberti/maestro/pkg/domain"
// )

// // MockKafkaClient is a mock implementation of the KafkaClient interface
// type MockKafkaClient struct {
// 	mock.Mock
// }

// func (m *MockKafkaClient) GetBrokers(ctx context.Context) ([]domain.BrokerInfo, error) {
// 	args := m.Called(ctx)
// 	brokers, _ := args.Get(0).([]domain.BrokerInfo)
// 	err, _ := args.Get(1).(error)
// 	return brokers, err
// }

// func (m *MockKafkaClient) ListTopics(ctx context.Context) ([]string, error) {
// 	args := m.Called(ctx)
// 	topics, _ := args.Get(0).([]string)
// 	err, _ := args.Get(1).(error)
// 	return topics, err
// }

// func (m *MockKafkaClient) GetTopicDetails(ctx context.Context, topicName string) (*domain.TopicInfo, error) {
// 	args := m.Called(ctx, topicName)
// 	topic, _ := args.Get(0).(*domain.TopicInfo)
// 	err, _ := args.Get(1).(error)
// 	return topic, err
// }

// func (m *MockKafkaClient) CreateTopic(ctx context.Context, topic domain.TopicInfo) error {
// 	args := m.Called(ctx, topic)
// 	return args.Error(0)
// }

// func (m *MockKafkaClient) DeleteTopic(ctx context.Context, topicName string) error {
// 	args := m.Called(ctx, topicName)
// 	return args.Error(0)
// }

// func (m *MockKafkaClient) UpdateTopicConfig(ctx context.Context, topicName string, config map[string]string) error {
// 	args := m.Called(ctx, topicName, config)
// 	return args.Error(0)
// }

// func (m *MockKafkaClient) ListConsumerGroups(ctx context.Context) ([]string, error) {
// 	args := m.Called(ctx)
// 	groups, _ := args.Get(0).([]string)
// 	err, _ := args.Get(1).(error)
// 	return groups, err
// }

// func (m *MockKafkaClient) GetConsumerGroupDetails(ctx context.Context, groupID string) (*domain.ConsumerGroupDetails, error) {
// 	args := m.Called(ctx, groupID)
// 	group, _ := args.Get(0).(*domain.ConsumerGroupDetails)
// 	err, _ := args.Get(1).(error)
// 	return group, err
// }

// func (m *MockKafkaClient) Close() {
// 	m.Called()
// }

// // setupRouter creates a test router with the given handler
// func setupRouter(handler gin.HandlerFunc, method, path string) *gin.Engine {
// 	gin.SetMode(gin.TestMode)
// 	router := gin.New()
// 	router.Handle(method, path, handler)
// 	return router
// }

// func TestGetClustersHandler(t *testing.T) {
// 	tests := []struct {
// 		name           string
// 		brokers        []domain.BrokerInfo
// 		err            error
// 		expectedStatus int
// 		expectedBody   map[string]interface{}
// 	}{
// 		{
// 			name: "successful_get_brokers",
// 			brokers: []domain.BrokerInfo{
// 				{ID: 1, Host: "broker1", Port: 9092},
// 				{ID: 2, Host: "broker2", Port: 9092},
// 			},
// 			err:            nil,
// 			expectedStatus: http.StatusOK,
// 			expectedBody: map[string]interface{}{
// 				"brokers": []interface{}{
// 					map[string]interface{}{"id": float64(1), "host": "broker1", "port": float64(9092)},
// 					map[string]interface{}{"id": float64(2), "host": "broker2", "port": float64(9092)},
// 				},
// 			},
// 		},
// 		{
// 			name:           "error_getting_brokers",
// 			brokers:        []domain.BrokerInfo{},
// 			err:            errors.New("failed to connect to Kafka"),
// 			expectedStatus: http.StatusInternalServerError,
// 			expectedBody: map[string]interface{}{
// 				"status":  float64(http.StatusInternalServerError),
// 				"message": "Failed to get broker metadata",
// 				"detail":  "failed to connect to Kafka",
// 			},
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			mockClient := new(MockKafkaClient)
// 			mockClient.On("GetBrokers", mock.Anything).Return(tt.brokers, tt.err)

// 			router := setupRouter(api.GetClustersHandler(mockClient), http.MethodGet, "/clusters")
// 			w := httptest.NewRecorder()
// 			req, _ := http.NewRequest(http.MethodGet, "/clusters", nil)
// 			router.ServeHTTP(w, req)

// 			assert.Equal(t, tt.expectedStatus, w.Code)

// 			var response map[string]interface{}
// 			err := json.Unmarshal(w.Body.Bytes(), &response)
// 			assert.NoError(t, err)
// 			assert.Equal(t, tt.expectedBody, response)

// 			mockClient.AssertExpectations(t)
// 		})
// 	}
// }

// func TestListTopicsHandler(t *testing.T) {
// 	tests := []struct {
// 		name           string
// 		topics         []string
// 		err            error
// 		expectedStatus int
// 		expectedBody   map[string]interface{}
// 	}{
// 		{
// 			name:           "successful_list_topics",
// 			topics:         []string{"topic1", "topic2", "topic3"},
// 			err:            nil,
// 			expectedStatus: http.StatusOK,
// 			expectedBody: map[string]interface{}{
// 				"topics": []interface{}{"topic1", "topic2", "topic3"},
// 			},
// 		},
// 		{
// 			name:           "error_listing_topics",
// 			topics:         []string{},
// 			err:            errors.New("failed to list topics"),
// 			expectedStatus: http.StatusInternalServerError,
// 			expectedBody: map[string]interface{}{
// 				"status":  float64(http.StatusInternalServerError),
// 				"message": "Failed to list topics",
// 				"detail":  "failed to list topics",
// 			},
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			mockClient := new(MockKafkaClient)
// 			mockClient.On("ListTopics", mock.Anything).Return(tt.topics, tt.err)

// 			router := setupRouter(api.ListTopicsHandler(mockClient), http.MethodGet, "/topics")
// 			w := httptest.NewRecorder()
// 			req, _ := http.NewRequest(http.MethodGet, "/topics", nil)
// 			router.ServeHTTP(w, req)

// 			assert.Equal(t, tt.expectedStatus, w.Code)

// 			var response map[string]interface{}
// 			err := json.Unmarshal(w.Body.Bytes(), &response)
// 			assert.NoError(t, err)
// 			assert.Equal(t, tt.expectedBody, response)

// 			mockClient.AssertExpectations(t)
// 		})
// 	}
// }

// func TestGetTopicHandler(t *testing.T) {
// 	tests := []struct {
// 		name           string
// 		topicName      string
// 		topic          *domain.TopicInfo
// 		err            error
// 		expectedStatus int
// 		expectedBody   map[string]interface{}
// 	}{
// 		{
// 			name:      "successful_get_topic",
// 			topicName: "test-topic",
// 			topic: &domain.TopicInfo{
// 				Name:              "test-topic",
// 				NumPartitions:     3,
// 				ReplicationFactor: 2,
// 				Config: map[string]string{
// 					"retention.ms": "86400000",
// 				},
// 				Partitions: []domain.PartitionInfo{
// 					{ID: 0, Leader: 1, Replicas: []int32{1, 2}, ISR: []int32{1, 2}},
// 				},
// 			},
// 			err:            nil,
// 			expectedStatus: http.StatusOK,
// 			expectedBody: map[string]interface{}{
// 				"topic": map[string]interface{}{
// 					"name":              "test-topic",
// 					"numPartitions":     float64(3),
// 					"replicationFactor": float64(2),
// 					"config": map[string]interface{}{
// 						"retention.ms": "86400000",
// 					},
// 					"partitions": []interface{}{
// 						map[string]interface{}{
// 							"id":       float64(0),
// 							"leader":   float64(1),
// 							"replicas": []interface{}{float64(1), float64(2)},
// 							"isr":      []interface{}{float64(1), float64(2)},
// 						},
// 					},
// 				},
// 			},
// 		},
// 		{
// 			name:           "empty_topic_name",
// 			topicName:      "",
// 			topic:          nil,
// 			err:            nil,
// 			expectedStatus: http.StatusBadRequest,
// 			expectedBody: map[string]interface{}{
// 				"status":  float64(http.StatusBadRequest),
// 				"message": "Topic name is required",
// 			},
// 		},
// 		{
// 			name:           "topic_not_found",
// 			topicName:      "nonexistent-topic",
// 			topic:          nil,
// 			err:            errors.New("topic not found"),
// 			expectedStatus: http.StatusInternalServerError,
// 			expectedBody: map[string]interface{}{
// 				"status":  float64(http.StatusInternalServerError),
// 				"message": "Failed to get topic details",
// 				"detail":  "topic not found",
// 			},
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			mockClient := new(MockKafkaClient)

// 			if tt.topicName != "" {
// 				mockClient.On("GetTopicDetails", mock.Anything, tt.topicName).Return(tt.topic, tt.err)
// 			}

// 			router := setupRouter(api.GetTopicHandler(mockClient), http.MethodGet, "/topics/:topicName")
// 			w := httptest.NewRecorder()

// 			var req *http.Request
// 			if tt.topicName == "" {
// 				req, _ = http.NewRequest(http.MethodGet, "/topics/", nil)
// 			} else {
// 				req, _ = http.NewRequest(http.MethodGet, "/topics/"+tt.topicName, nil)
// 			}

// 			router.ServeHTTP(w, req)

// 			assert.Equal(t, tt.expectedStatus, w.Code)

// 			var response map[string]interface{}
// 			err := json.Unmarshal(w.Body.Bytes(), &response)
// 			assert.NoError(t, err)
// 			assert.Equal(t, tt.expectedBody, response)

// 			mockClient.AssertExpectations(t)
// 		})
// 	}
// }

// func TestCreateTopicHandler(t *testing.T) {
// 	tests := []struct {
// 		name            string
// 		requestBody     map[string]interface{}
// 		createErr       error
// 		getDetailsErr   error
// 		topicDetails    *domain.TopicInfo
// 		expectedStatus  int
// 		expectedBodyKey string
// 	}{
// 		{
// 			name: "successful_create_topic",
// 			requestBody: map[string]interface{}{
// 				"name":              "new-topic",
// 				"numPartitions":     float64(3),
// 				"replicationFactor": float64(2),
// 				"config": map[string]string{
// 					"retention.ms": "86400000",
// 				},
// 			},
// 			createErr:     nil,
// 			getDetailsErr: nil,
// 			topicDetails: &domain.TopicInfo{
// 				Name:              "new-topic",
// 				NumPartitions:     3,
// 				ReplicationFactor: 2,
// 				Config: map[string]string{
// 					"retention.ms": "86400000",
// 				},
// 			},
// 			expectedStatus:  http.StatusCreated,
// 			expectedBodyKey: "message",
// 		},
// 		{
// 			name: "topic_already_exists",
// 			requestBody: map[string]interface{}{
// 				"name":              "existing-topic",
// 				"numPartitions":     float64(3),
// 				"replicationFactor": float64(2),
// 			},
// 			createErr:       errors.New("topic already exists"),
// 			getDetailsErr:   nil,
// 			topicDetails:    nil,
// 			expectedStatus:  http.StatusConflict,
// 			expectedBodyKey: "error",
// 		},
// 		{
// 			name: "invalid_request",
// 			requestBody: map[string]interface{}{
// 				"name": "invalid-topic",
// 				// Missing required fields
// 			},
// 			createErr:       nil,
// 			getDetailsErr:   nil,
// 			topicDetails:    nil,
// 			expectedStatus:  http.StatusBadRequest,
// 			expectedBodyKey: "error",
// 		},
// 		{
// 			name: "create_topic_error",
// 			requestBody: map[string]interface{}{
// 				"name":              "error-topic",
// 				"numPartitions":     float64(3),
// 				"replicationFactor": float64(2),
// 			},
// 			createErr:       errors.New("failed to create topic"),
// 			getDetailsErr:   nil,
// 			topicDetails:    nil,
// 			expectedStatus:  http.StatusInternalServerError,
// 			expectedBodyKey: "error",
// 		},
// 		{
// 			name: "get_details_error_after_create",
// 			requestBody: map[string]interface{}{
// 				"name":              "detail-error-topic",
// 				"numPartitions":     float64(3),
// 				"replicationFactor": float64(2),
// 			},
// 			createErr:       nil,
// 			getDetailsErr:   errors.New("failed to get details"),
// 			topicDetails:    nil,
// 			expectedStatus:  http.StatusCreated, // Should still succeed
// 			expectedBodyKey: "message",
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			mockClient := new(MockKafkaClient)

// 			if _, ok := tt.requestBody["name"]; ok && len(tt.requestBody) >= 3 {
// 				// Only set up expectations for valid requests
// 				topicName := tt.requestBody["name"].(string)
// 				numPartitions := int32(tt.requestBody["numPartitions"].(float64))
// 				replicationFactor := int(tt.requestBody["replicationFactor"].(float64))

// 				var config map[string]string
// 				if configVal, exists := tt.requestBody["config"]; exists {
// 					config = configVal.(map[string]string)
// 				}

// 				topicInfo := domain.TopicInfo{
// 					Name:              topicName,
// 					NumPartitions:     numPartitions,
// 					ReplicationFactor: replicationFactor,
// 					Config:            config,
// 				}

// 				mockClient.On("CreateTopic", mock.Anything, mock.MatchedBy(func(t domain.TopicInfo) bool {
// 					return t.Name == topicInfo.Name &&
// 						t.NumPartitions == topicInfo.NumPartitions &&
// 						t.ReplicationFactor == topicInfo.ReplicationFactor
// 				})).Return(tt.createErr)

// 				if tt.createErr == nil {
// 					mockClient.On("GetTopicDetails", mock.Anything, topicName).Return(tt.topicDetails, tt.getDetailsErr)
// 				}
// 			}

// 			router := setupRouter(api.CreateTopicHandler(mockClient), http.MethodPost, "/topics")
// 			w := httptest.NewRecorder()

// 			jsonBody, _ := json.Marshal(tt.requestBody)
// 			req, _ := http.NewRequest(http.MethodPost, "/topics", bytes.NewBuffer(jsonBody))
// 			req.Header.Set("Content-Type", "application/json")

// 			router.ServeHTTP(w, req)

// 			assert.Equal(t, tt.expectedStatus, w.Code)

// 			var response map[string]interface{}
// 			err := json.Unmarshal(w.Body.Bytes(), &response)
// 			assert.NoError(t, err)
// 			assert.Contains(t, response, tt.expectedBodyKey)

// 			mockClient.AssertExpectations(t)
// 		})
// 	}
// }

// func TestDeleteTopicHandler(t *testing.T) {
// 	tests := []struct {
// 		name           string
// 		topicName      string
// 		err            error
// 		expectedStatus int
// 		expectedBody   map[string]interface{}
// 	}{
// 		{
// 			name:           "successful_delete",
// 			topicName:      "topic-to-delete",
// 			err:            nil,
// 			expectedStatus: http.StatusOK,
// 			expectedBody: map[string]interface{}{
// 				"message": "Topic deleted successfully",
// 				"topic":   "topic-to-delete",
// 			},
// 		},
// 		{
// 			name:           "empty_topic_name",
// 			topicName:      "",
// 			err:            nil,
// 			expectedStatus: http.StatusBadRequest,
// 			expectedBody: map[string]interface{}{
// 				"status":  float64(http.StatusBadRequest),
// 				"message": "Topic name is required",
// 			},
// 		},
// 		{
// 			name:           "topic_not_found",
// 			topicName:      "nonexistent-topic",
// 			err:            errors.New("topic not found"),
// 			expectedStatus: http.StatusNotFound,
// 			expectedBody: map[string]interface{}{
// 				"status":  float64(http.StatusNotFound),
// 				"message": "Topic not found",
// 				"detail":  "topic not found",
// 			},
// 		},
// 		{
// 			name:           "delete_error",
// 			topicName:      "error-topic",
// 			err:            errors.New("failed to delete topic"),
// 			expectedStatus: http.StatusInternalServerError,
// 			expectedBody: map[string]interface{}{
// 				"status":  float64(http.StatusInternalServerError),
// 				"message": "Failed to delete topic",
// 				"detail":  "failed to delete topic",
// 			},
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			mockClient := new(MockKafkaClient)

// 			if tt.topicName != "" {
// 				mockClient.On("DeleteTopic", mock.Anything, tt.topicName).Return(tt.err)
// 			}

// 			router := setupRouter(api.DeleteTopicHandler(mockClient), http.MethodDelete, "/topics/:topicName")
// 			w := httptest.NewRecorder()

// 			var req *http.Request
// 			if tt.topicName == "" {
// 				req, _ = http.NewRequest(http.MethodDelete, "/topics/", nil)
// 			} else {
// 				req, _ = http.NewRequest(http.MethodDelete, "/topics/"+tt.topicName, nil)
// 			}

// 			router.ServeHTTP(w, req)

// 			assert.Equal(t, tt.expectedStatus, w.Code)

// 			var response map[string]interface{}
// 			err := json.Unmarshal(w.Body.Bytes(), &response)
// 			assert.NoError(t, err)
// 			assert.Equal(t, tt.expectedBody, response)

// 			mockClient.AssertExpectations(t)
// 		})
// 	}
// }

// func TestUpdateTopicConfigHandler(t *testing.T) {
// 	tests := []struct {
// 		name            string
// 		topicName       string
// 		requestBody     map[string]interface{}
// 		updateErr       error
// 		getDetailsErr   error
// 		topicDetails    *domain.TopicInfo
// 		expectedStatus  int
// 		expectedBodyKey string
// 	}{
// 		{
// 			name:      "successful_update",
// 			topicName: "topic-to-update",
// 			requestBody: map[string]interface{}{
// 				"config": map[string]string{
// 					"retention.ms": "172800000",
// 				},
// 			},
// 			updateErr:     nil,
// 			getDetailsErr: nil,
// 			topicDetails: &domain.TopicInfo{
// 				Name: "topic-to-update",
// 				Config: map[string]string{
// 					"retention.ms": "172800000",
// 				},
// 			},
// 			expectedStatus:  http.StatusOK,
// 			expectedBodyKey: "message",
// 		},
// 		{
// 			name:            "empty_topic_name",
// 			topicName:       "",
// 			requestBody:     nil,
// 			updateErr:       nil,
// 			getDetailsErr:   nil,
// 			topicDetails:    nil,
// 			expectedStatus:  http.StatusBadRequest,
// 			expectedBodyKey: "status",
// 		},
// 		{
// 			name:      "empty_config",
// 			topicName: "topic-to-update",
// 			requestBody: map[string]interface{}{
// 				"config": map[string]string{},
// 			},
// 			updateErr:       nil,
// 			getDetailsErr:   nil,
// 			topicDetails:    nil,
// 			expectedStatus:  http.StatusBadRequest,
// 			expectedBodyKey: "status",
// 		},
// 		{
// 			name:      "topic_not_found",
// 			topicName: "nonexistent-topic",
// 			requestBody: map[string]interface{}{
// 				"config": map[string]string{
// 					"retention.ms": "172800000",
// 				},
// 			},
// 			updateErr:       errors.New("topic not found"),
// 			getDetailsErr:   nil,
// 			topicDetails:    nil,
// 			expectedStatus:  http.StatusNotFound,
// 			expectedBodyKey: "status",
// 		},
// 		{
// 			name:      "update_error",
// 			topicName: "error-topic",
// 			requestBody: map[string]interface{}{
// 				"config": map[string]string{
// 					"retention.ms": "172800000",
// 				},
// 			},
// 			updateErr:       errors.New("failed to update config"),
// 			getDetailsErr:   nil,
// 			topicDetails:    nil,
// 			expectedStatus:  http.StatusInternalServerError,
// 			expectedBodyKey: "status",
// 		},
// 		{
// 			name:      "get_details_error_after_update",
// 			topicName: "detail-error-topic",
// 			requestBody: map[string]interface{}{
// 				"config": map[string]string{
// 					"retention.ms": "172800000",
// 				},
// 			},
// 			updateErr:       nil,
// 			getDetailsErr:   errors.New("failed to get details"),
// 			topicDetails:    nil,
// 			expectedStatus:  http.StatusOK, // Should still succeed
// 			expectedBodyKey: "message",
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			mockClient := new(MockKafkaClient)

// 			if tt.topicName != "" && tt.requestBody != nil {
// 				if configVal, exists := tt.requestBody["config"]; exists {
// 					config := configVal.(map[string]string)
// 					if len(config) > 0 {
// 						mockClient.On("UpdateTopicConfig", mock.Anything, tt.topicName, mock.MatchedBy(func(c map[string]string) bool {
// 							return len(c) == len(config)
// 						})).Return(tt.updateErr)

// 						if tt.updateErr == nil {
// 							mockClient.On("GetTopicDetails", mock.Anything, tt.topicName).Return(tt.topicDetails, tt.getDetailsErr)
// 						}
// 					}
// 				}
// 			}

// 			router := setupRouter(api.UpdateTopicConfigHandler(mockClient), http.MethodPut, "/topics/:topicName/config")
// 			w := httptest.NewRecorder()

// 			var req *http.Request
// 			if tt.requestBody != nil {
// 				jsonBody, _ := json.Marshal(tt.requestBody)
// 				if tt.topicName == "" {
// 					req, _ = http.NewRequest(http.MethodPut, "/topics//config", bytes.NewBuffer(jsonBody))
// 				} else {
// 					req, _ = http.NewRequest(http.MethodPut, "/topics/"+tt.topicName+"/config", bytes.NewBuffer(jsonBody))
// 				}
// 				req.Header.Set("Content-Type", "application/json")
// 			} else {
// 				if tt.topicName == "" {
// 					req, _ = http.NewRequest(http.MethodPut, "/topics//config", nil)
// 				} else {
// 					req, _ = http.NewRequest(http.MethodPut, "/topics/"+tt.topicName+"/config", nil)
// 				}
// 			}

// 			router.ServeHTTP(w, req)

// 			assert.Equal(t, tt.expectedStatus, w.Code)

// 			var response map[string]interface{}
// 			err := json.Unmarshal(w.Body.Bytes(), &response)
// 			assert.NoError(t, err)
// 			assert.Contains(t, response, tt.expectedBodyKey)

// 			mockClient.AssertExpectations(t)
// 		})
// 	}
// }

// func TestListConsumerGroupsHandler(t *testing.T) {
// 	tests := []struct {
// 		name           string
// 		groups         []string
// 		err            error
// 		expectedStatus int
// 		expectedBody   map[string]interface{}
// 	}{
// 		{
// 			name:           "successful_list_groups",
// 			groups:         []string{"group1", "group2", "group3"},
// 			err:            nil,
// 			expectedStatus: http.StatusOK,
// 			expectedBody: map[string]interface{}{
// 				"groups": []interface{}{"group1", "group2", "group3"},
// 			},
// 		},
// 		{
// 			name:           "error_listing_groups",
// 			groups:         []string{},
// 			err:            errors.New("failed to list groups"),
// 			expectedStatus: http.StatusInternalServerError,
// 			expectedBody: map[string]interface{}{
// 				"status":  float64(http.StatusInternalServerError),
// 				"message": "Failed to list consumer groups",
// 				"detail":  "failed to list groups",
// 			},
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			mockClient := new(MockKafkaClient)
// 			mockClient.On("ListConsumerGroups", mock.Anything).Return(tt.groups, tt.err)

// 			router := setupRouter(api.ListConsumerGroupsHandler(mockClient), http.MethodGet, "/consumer-groups")
// 			w := httptest.NewRecorder()
// 			req, _ := http.NewRequest(http.MethodGet, "/consumer-groups", nil)
// 			router.ServeHTTP(w, req)

// 			assert.Equal(t, tt.expectedStatus, w.Code)

// 			var response map[string]interface{}
// 			err := json.Unmarshal(w.Body.Bytes(), &response)
// 			assert.NoError(t, err)
// 			assert.Equal(t, tt.expectedBody, response)

// 			mockClient.AssertExpectations(t)
// 		})
// 	}
// }

// func TestGetConsumerGroupHandler(t *testing.T) {
// 	tests := []struct {
// 		name           string
// 		groupID        string
// 		group          *domain.ConsumerGroupDetails
// 		err            error
// 		expectedStatus int
// 		expectedBody   map[string]interface{}
// 	}{
// 		{
// 			name:    "successful_get_group",
// 			groupID: "test-group",
// 			group: &domain.ConsumerGroupDetails{
// 				GroupID: "test-group",
// 				State:   "Stable",
// 				Members: []domain.ConsumerGroupMemberInfo{
// 					{
// 						ClientID:   "client-1",
// 						ConsumerID: "member-1",
// 						Host:       "host-1",
// 						Assignments: []domain.TopicPartitionAssignment{
// 							{Topic: "topic1", Partition: 0},
// 							{Topic: "topic2", Partition: 1},
// 						},
// 					},
// 				},
// 			},
// 			err:            nil,
// 			expectedStatus: http.StatusOK,
// 			expectedBody: map[string]interface{}{
// 				"group": map[string]interface{}{
// 					"groupId": "test-group",
// 					"state":   "Stable",
// 					"members": []interface{}{
// 						map[string]interface{}{
// 							"id":       "member-1",
// 							"clientId": "client-1",
// 							"topics":   []interface{}{"topic1", "topic2"},
// 						},
// 					},
// 				},
// 			},
// 		},
// 		{
// 			name:           "empty_group_id",
// 			groupID:        "",
// 			group:          nil,
// 			err:            nil,
// 			expectedStatus: http.StatusBadRequest,
// 			expectedBody: map[string]interface{}{
// 				"status":  float64(http.StatusBadRequest),
// 				"message": "Consumer group ID is required",
// 			},
// 		},
// 		{
// 			name:           "group_not_found",
// 			groupID:        "nonexistent-group",
// 			group:          nil,
// 			err:            errors.New("group not found"),
// 			expectedStatus: http.StatusNotFound,
// 			expectedBody: map[string]interface{}{
// 				"status":  float64(http.StatusNotFound),
// 				"message": "Consumer group not found",
// 				"detail":  "group not found",
// 			},
// 		},
// 		{
// 			name:           "error_getting_group",
// 			groupID:        "error-group",
// 			group:          nil,
// 			err:            errors.New("failed to get group details"),
// 			expectedStatus: http.StatusInternalServerError,
// 			expectedBody: map[string]interface{}{
// 				"status":  float64(http.StatusInternalServerError),
// 				"message": "Failed to get consumer group details",
// 				"detail":  "failed to get group details",
// 			},
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			mockClient := new(MockKafkaClient)

// 			if tt.groupID != "" {
// 				mockClient.On("GetConsumerGroupDetails", mock.Anything, tt.groupID).Return(tt.group, tt.err)
// 			}

// 			router := setupRouter(api.GetConsumerGroupHandler(mockClient), http.MethodGet, "/consumer-groups/:groupId")
// 			w := httptest.NewRecorder()

// 			var req *http.Request
// 			if tt.groupID == "" {
// 				req, _ = http.NewRequest(http.MethodGet, "/consumer-groups/", nil)
// 			} else {
// 				req, _ = http.NewRequest(http.MethodGet, "/consumer-groups/"+tt.groupID, nil)
// 			}

// 			router.ServeHTTP(w, req)

// 			assert.Equal(t, tt.expectedStatus, w.Code)

// 			var response map[string]interface{}
// 			err := json.Unmarshal(w.Body.Bytes(), &response)
// 			assert.NoError(t, err)
// 			assert.Equal(t, tt.expectedBody, response)

// 			mockClient.AssertExpectations(t)
// 		})
// 	}
// }
