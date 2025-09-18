package main

import (
	"log"
	"os"

	"admira-etl/internal/api"
	"admira-etl/internal/config"
	"admira-etl/internal/etl"
	"admira-etl/internal/storage"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Initialize configuration
	cfg := config.Load()

	// Initialize logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)

	// Initialize storage
	store := storage.NewInMemoryStorage()

	// Initialize ETL service
	etlService := etl.NewService(cfg, store, logger)

	// Initialize API handlers
	handlers := api.NewHandlers(etlService, logger)

	// Setup router
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Add request ID middleware
	router.Use(func(c *gin.Context) {
		c.Header("X-Request-ID", c.GetHeader("X-Request-ID"))
		c.Next()
	})

	// Setup routes
	api.SetupRoutes(router, handlers)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logger.WithField("port", port).Info("Starting server")
	if err := router.Run(":" + port); err != nil {
		logger.WithError(err).Fatal("Failed to start server")
	}
}

