package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	
	// Set log level from environment
	if level, err := logrus.ParseLevel(cfg.LogLevel); err == nil {
		logger.SetLevel(level)
	} else {
		logger.SetLevel(logrus.InfoLevel)
	}

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

	// Start server with graceful shutdown
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Create server with graceful shutdown
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		logger.WithField("port", port).Info("Starting server")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Fatal("Failed to start server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	// Give outstanding requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.WithError(err).Fatal("Server forced to shutdown")
	}

	logger.Info("Server exited")
}

