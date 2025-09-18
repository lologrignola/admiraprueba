package api

import (
	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine, handlers *Handlers) {
	// Health check endpoints
	router.GET("/healthz", handlers.HealthCheck)
	router.GET("/readyz", handlers.ReadinessCheck)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Ingestion endpoints
		v1.POST("/ingest/run", handlers.RunIngestion)

		// Metrics endpoints
		v1.GET("/metrics/channel", handlers.GetChannelMetrics)
		v1.GET("/metrics/funnel", handlers.GetFunnelMetrics)

		// Export endpoints
		v1.POST("/export/run", handlers.ExportData)
	}
}

