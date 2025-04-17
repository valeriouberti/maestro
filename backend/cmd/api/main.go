package main

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/valeriouberti/maestro/pkg/api"
)

func main() {
	r := gin.Default()

	r.GET("/api/v1/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	apiGroup := r.Group("/api/v1")
	{
		apiGroup.GET("/api/health", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"status": "ok",
			})
		})
		apiGroup.GET("/clusters", api.GetClustersHandler)
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
	r.Run(":" + port)
}
