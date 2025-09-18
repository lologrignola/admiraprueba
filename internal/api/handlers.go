package api

import (
	"net/http"
	"time"

	"admira-etl/internal/etl"
	"admira-etl/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Handlers struct {
	etlService *etl.Service
	logger     *logrus.Logger
}

func NewHandlers(etlService *etl.Service, logger *logrus.Logger) *Handlers {
	return &Handlers{
		etlService: etlService,
		logger:     logger,
	}
}

func (h *Handlers) RunIngestion(c *gin.Context) {
	var req models.IngestRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.WithError(err).Error("Invalid ingestion request")
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid request parameters",
			Message: err.Error(),
		})
		return
	}

	h.logger.WithField("since", req.Since).Info("Starting ingestion")

	if err := h.etlService.RunIngestion(c.Request.Context(), req.Since); err != nil {
		h.logger.WithError(err).Error("Ingestion failed")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Ingestion failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Ingestion completed successfully",
		"since":   req.Since,
	})
}

func (h *Handlers) GetChannelMetrics(c *gin.Context) {
	var req models.MetricsChannelRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.WithError(err).Error("Invalid channel metrics request")
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid request parameters",
			Message: err.Error(),
		})
		return
	}

	// Parse dates
	from, err := time.Parse("2006-01-02", req.From)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid from date format",
			Message: "Expected YYYY-MM-DD format",
		})
		return
	}

	to, err := time.Parse("2006-01-02", req.To)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid to date format",
			Message: "Expected YYYY-MM-DD format",
		})
		return
	}

	// Set default pagination
	if req.Limit <= 0 {
		req.Limit = 100
	}

	data, err := h.etlService.GetChannelMetrics(from, to, req.Channel, req.Limit, req.Offset)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get channel metrics")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to retrieve metrics",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   data,
		"count":  len(data),
		"limit":  req.Limit,
		"offset": req.Offset,
	})
}

func (h *Handlers) GetFunnelMetrics(c *gin.Context) {
	var req models.MetricsFunnelRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.WithError(err).Error("Invalid funnel metrics request")
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid request parameters",
			Message: err.Error(),
		})
		return
	}

	// Parse dates
	from, err := time.Parse("2006-01-02", req.From)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid from date format",
			Message: "Expected YYYY-MM-DD format",
		})
		return
	}

	to, err := time.Parse("2006-01-02", req.To)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid to date format",
			Message: "Expected YYYY-MM-DD format",
		})
		return
	}

	// Set default pagination
	if req.Limit <= 0 {
		req.Limit = 100
	}

	data, err := h.etlService.GetFunnelMetrics(from, to, req.UTMCampaign, req.Limit, req.Offset)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get funnel metrics")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to retrieve metrics",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   data,
		"count":  len(data),
		"limit":  req.Limit,
		"offset": req.Offset,
	})
}

func (h *Handlers) ExportData(c *gin.Context) {
	var req models.ExportRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.WithError(err).Error("Invalid export request")
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid request parameters",
			Message: err.Error(),
		})
		return
	}

	h.logger.WithField("date", req.Date).Info("Starting data export")

	if err := h.etlService.ExportData(c.Request.Context(), req.Date); err != nil {
		h.logger.WithError(err).Error("Export failed")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Export failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Export completed successfully",
		"date":    req.Date,
	})
}

func (h *Handlers) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, models.HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().Format(time.RFC3339),
		Version:   "1.0.0",
	})
}

func (h *Handlers) ReadinessCheck(c *gin.Context) {
	// Check if external APIs are accessible
	// For simplicity, we'll just return ready if the service is running
	c.JSON(http.StatusOK, models.HealthResponse{
		Status:    "ready",
		Timestamp: time.Now().Format(time.RFC3339),
		Version:   "1.0.0",
	})
}

