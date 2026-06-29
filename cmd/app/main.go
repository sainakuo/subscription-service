package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/sainakuo/subscription-service/internal/handler"
)

func main() {
	router := gin.Default()
	router.GET("/health", handler.HealthCheck)
	log.Println("Subscription service started on port 8080")

	err := router.Run(":8080")

	if err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
