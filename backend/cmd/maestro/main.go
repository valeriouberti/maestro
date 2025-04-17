package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/valeriouberti/maestro/internal/config"
	"github.com/valeriouberti/maestro/internal/kafka"
	"github.com/valeriouberti/maestro/pkg/api"
)

func main() {
	// Setup logger
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Starting Maestro Kafka Management Service")

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Set Gin mode based on environment
	if cfg.EnvironmentName == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize router
	r := gin.Default()

	// Initialize Kafka client
	kClient, err := kafka.NewKafkaClient(cfg.KafkaBrokers)
	if err != nil {
		log.Fatalf("Failed to create Kafka client: %v", err)
	}
	defer kClient.Close()

	// Setup API routes
	setupRoutes(r, kClient)

	// Configure HTTP server
	srv := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      r,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server listening on port %s", cfg.ServerPort)
		if cfg.EnableTLS {
			if err := srv.ListenAndServeTLS(cfg.CertFile, cfg.KeyFile); err != nil && err != http.ErrServerClosed {
				log.Fatalf("Failed to start server: %v", err)
			}
		} else {
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("Failed to start server: %v", err)
			}
		}
	}()

	// Setup graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown the server
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited successfully")
}

// setupRoutes configures all API routes
func setupRoutes(r *gin.Engine, kClient *kafka.KafkaClient) {
	// Add basic middleware
	r.Use(gin.Recovery())
	r.Use(corsMiddleware())

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// API routes
	apiGroup := r.Group("/api/v1")
	{
		apiGroup.GET("/clusters", api.GetClustersHandler(kClient))
		apiGroup.GET("/topics", api.ListTopicsHandler(kClient))
		apiGroup.GET("/topics/:topicName", api.GetTopicHandler(kClient))
		apiGroup.POST("/topics", api.CreateTopicHandler(kClient))
		apiGroup.DELETE("/topics/:topicName", api.DeleteTopicHandler(kClient))
		apiGroup.PUT("/topics/:topicName/config", api.UpdateTopicConfigHandler(kClient))
		apiGroup.GET("/consumergroups", api.ListConsumerGroupsHandler(kClient))
		apiGroup.GET("/consumergroups/:groupId", api.GetConsumerGroupHandler(kClient))
	}
}

// corsMiddleware handles CORS for the API
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
