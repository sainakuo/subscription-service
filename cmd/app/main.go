package main

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/sainakuo/subscription-service/internal/config"
	database "github.com/sainakuo/subscription-service/internal/db"
	"github.com/sainakuo/subscription-service/internal/handler"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	dbPool, err := database.NewPostgresPool(ctx, cfg)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer dbPool.Close()

	log.Println("Database connection established")

	router := gin.Default()
	router.GET("/health", handler.HealthCheck)
	address := ":" + cfg.AppPort
	log.Printf("Subscription service started on port %v", cfg.AppPort)

	err = router.Run(address)

	if err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
