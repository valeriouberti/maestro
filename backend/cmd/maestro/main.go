package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/valeriouberti/maestro/internal/config"
	"github.com/valeriouberti/maestro/pkg/api"
	"github.com/valeriouberti/maestro/pkg/kafka"
)

func main() {
	r := gin.Default()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
		os.Exit(1)
	}

	kClient, err := kafka.NewKafkaClient(cfg.KafkaBrokers)
	if err != nil {
		log.Fatalf("Error creating Kafka client: %v", err)
		os.Exit(1)
	}
	defer kClient.Close() // Ensure client is closed on exit

	apiGroup := r.Group("/api/v1")
	{
		apiGroup.GET("health", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"status": "ok",
			})
		})
		apiGroup.GET("/clusters", api.GetClustersHandler(kClient))
		apiGroup.GET("/topics", api.ListTopicsHandler)
		apiGroup.GET("/topics/:topicName", api.GetTopicHandler)
		apiGroup.POST("/topics", api.CreateTopicHandler)
		apiGroup.DELETE("/topics/:topicName", api.DeleteTopicHandler)
		apiGroup.PUT("/topics/:topicName/config", api.UpdateTopicConfigHandler)
		apiGroup.GET("/consumergroups", api.ListConsumerGroupsHandler)
		apiGroup.GET("/consumergroups/:groupId", api.GetConsumerGroupHandler)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("Server listening on port %s", port)
		if err := r.Run(":" + port); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error starting server: %v", err)
		}
	}()

	<-quit // Wait for interrupt signal

	log.Println("Shutting down server...")

}
