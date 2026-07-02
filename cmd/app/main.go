package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/sainakuo/subscription-service/internal/config"
	database "github.com/sainakuo/subscription-service/internal/db"
	"github.com/sainakuo/subscription-service/internal/handler"
	"github.com/sainakuo/subscription-service/internal/logger"
	"github.com/sainakuo/subscription-service/internal/middleware"
	"github.com/sainakuo/subscription-service/internal/repository"
	"github.com/sainakuo/subscription-service/internal/service"
)

func main() {
	cfg := config.Load()

	log := logger.New(cfg.LogLevel)
	slog.SetDefault(log)

	log.Info("configuration loaded",
		"app_port", cfg.AppPort,
		"log_level", cfg.LogLevel,
		"db_host", cfg.DBHost,
		"db_port", cfg.DBPort,
		"db_name", cfg.DBName,
	)

	ctx := context.Background()

	dbPool, err := database.NewPostgresPool(ctx, cfg)
	if err != nil {
		log.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer dbPool.Close()

	log.Info("Database connection established")

	subscriptionRepository := repository.NewSubscriptionRepository(dbPool)
	subscriptionService := service.NewSubscriptionService(subscriptionRepository)
	subscriptionHandler := handler.NewSubscriptionHandler(subscriptionService, log)

	router := gin.New()

	router.Use(gin.Recovery())
	router.Use(middleware.RequestLogger(log))

	router.GET("/health", handler.HealthCheck)

	router.POST("/subscriptions", subscriptionHandler.Create)

	router.GET("/subscriptions", subscriptionHandler.List)

	router.GET("/subscriptions/total", subscriptionHandler.CalculateTotalCost)

	router.GET("/subscriptions/:id", subscriptionHandler.GetByID)

	router.PUT("/subscriptions/:id", subscriptionHandler.Update)

	router.DELETE("/subscriptions/:id", subscriptionHandler.Delete)

	address := ":" + cfg.AppPort
	log.Info("Subscription service started", "address", address)

	err = router.Run(address)

	if err != nil {
		log.Error("failed to start server", "error", err)
		os.Exit(1)
	}
}
