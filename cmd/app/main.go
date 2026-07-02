package main

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/sainakuo/subscription-service/internal/config"
	database "github.com/sainakuo/subscription-service/internal/db"
	"github.com/sainakuo/subscription-service/internal/handler"
	"github.com/sainakuo/subscription-service/internal/repository"
	"github.com/sainakuo/subscription-service/internal/service"
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

	subscriptionRepository := repository.NewSubscriptionRepository(dbPool)
	subscriptionService := service.NewSubscriptionService(subscriptionRepository)
	subscriptionHandler := handler.NewSubscriptionHandler(subscriptionService)

	router := gin.Default()
	router.GET("/health", handler.HealthCheck)

	router.POST("/subscriptions", subscriptionHandler.Create)

	router.GET("/subscriptions", subscriptionHandler.List)

	router.GET("/subscriptions/total", subscriptionHandler.CalculateTotalCost)

	router.GET("/subscriptions/:id", subscriptionHandler.GetByID)

	router.PUT("/subscriptions/:id", subscriptionHandler.Update)

	router.DELETE("/subscriptions/:id", subscriptionHandler.Delete)

	address := ":" + cfg.AppPort
	log.Printf("Subscription service started on port %v", cfg.AppPort)

	err = router.Run(address)

	if err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
