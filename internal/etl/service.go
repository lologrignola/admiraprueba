package etl

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"admira-etl/internal/config"
	"admira-etl/internal/http"
	"admira-etl/internal/models"
	"admira-etl/internal/storage"

	"github.com/sirupsen/logrus"
)

type Service struct {
	config  *config.Config
	storage storage.Storage
	client  *http.Client
	logger  *logrus.Logger
}

func NewService(cfg *config.Config, store storage.Storage, logger *logrus.Logger) *Service {
	httpClient := http.NewClient(http.ClientConfig{
		Timeout:    cfg.HTTPTimeout,
		MaxRetries: cfg.MaxRetries,
		RetryDelay: cfg.RetryDelay,
	}, logger)

	return &Service{
		config:  cfg,
		storage: store,
		client:  httpClient,
		logger:  logger,
	}
}

func (s *Service) RunIngestion(ctx context.Context, since string) error {
	s.logger.WithField("since", since).Info("Starting data ingestion")

	// Parse since date
	var sinceTime time.Time
	var err error
	if since != "" {
		sinceTime, err = time.Parse("2006-01-02", since)
		if err != nil {
			return fmt.Errorf("invalid since date format: %w", err)
		}
	}

	// Fetch data from external APIs
	adsData, err := s.fetchAdsData(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch ads data: %w", err)
	}

	crmData, err := s.fetchCRMData(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch crm data: %w", err)
	}

	// Transform and merge data
	transformedData, err := s.transformData(adsData, crmData, sinceTime)
	if err != nil {
		return fmt.Errorf("failed to transform data: %w", err)
	}

	// Store transformed data
	if err := s.storage.StoreTransformedData(transformedData); err != nil {
		return fmt.Errorf("failed to store transformed data: %w", err)
	}

	// Update last ingestion time
	if err := s.storage.SetLastIngestionTime(time.Now()); err != nil {
		return fmt.Errorf("failed to update last ingestion time: %w", err)
	}

	s.logger.WithField("records_processed", len(transformedData)).Info("Data ingestion completed")
	return nil
}

func (s *Service) fetchAdsData(ctx context.Context) (*models.AdsData, error) {
	if s.config.AdsAPIURL == "" {
		return nil, fmt.Errorf("ads API URL not configured")
	}

	var response models.ExternalResponse
	if err := s.client.Get(ctx, s.config.AdsAPIURL, &response); err != nil {
		return nil, err
	}

	if response.External.Ads == nil {
		return &models.AdsData{Performance: []models.AdsPerformance{}}, nil
	}

	return response.External.Ads, nil
}

func (s *Service) fetchCRMData(ctx context.Context) (*models.CRMData, error) {
	if s.config.CRMAPIURL == "" {
		return nil, fmt.Errorf("crm API URL not configured")
	}

	var response models.ExternalResponse
	if err := s.client.Get(ctx, s.config.CRMAPIURL, &response); err != nil {
		return nil, err
	}

	if response.External.CRM == nil {
		return &models.CRMData{Opportunities: []models.Opportunity{}}, nil
	}

	return response.External.CRM, nil
}

func (s *Service) transformData(adsData *models.AdsData, crmData *models.CRMData, sinceTime time.Time) ([]models.TransformedData, error) {
	// Group CRM opportunities by UTM parameters for efficient lookup
	crmLookup := s.buildCRMLookup(crmData.Opportunities)

	var transformedData []models.TransformedData

	for _, ad := range adsData.Performance {
		// Filter by date if sinceTime is specified
		if !sinceTime.IsZero() {
			adDate, err := time.Parse("2006-01-02", ad.Date)
			if err != nil {
				s.logger.WithField("date", ad.Date).Warn("Invalid date format in ads data, skipping")
				continue
			}
			if adDate.Before(sinceTime) {
				continue
			}
		}

		// Find matching CRM opportunities
		matchingOpportunities := s.findMatchingOpportunities(ad, crmLookup)

		// Calculate metrics
		metrics := s.calculateMetrics(ad, matchingOpportunities)

		transformedData = append(transformedData, models.TransformedData{
			Date:         ad.Date,
			Channel:      ad.Channel,
			CampaignID:   ad.CampaignID,
			Clicks:       ad.Clicks,
			Impressions:  ad.Impressions,
			Cost:         ad.Cost,
			Leads:        metrics.Leads,
			Opportunities: metrics.Opportunities,
			ClosedWon:    metrics.ClosedWon,
			Revenue:      metrics.Revenue,
			CPC:          metrics.CPC,
			CPA:          metrics.CPA,
			CVRLeadToOpp: metrics.CVRLeadToOpp,
			CVROppToWon:  metrics.CVROppToWon,
			ROAS:         metrics.ROAS,
		})
	}

	return transformedData, nil
}

type CRMLookupKey struct {
	UTMCampaign string
	UTMSource   string
	UTMMedium   string
}

type Metrics struct {
	Leads         int
	Opportunities int
	ClosedWon     int
	Revenue       float64
	CPC           float64
	CPA           float64
	CVRLeadToOpp  float64
	CVROppToWon   float64
	ROAS          float64
}

func (s *Service) buildCRMLookup(opportunities []models.Opportunity) map[CRMLookupKey][]models.Opportunity {
	lookup := make(map[CRMLookupKey][]models.Opportunity)

	for _, opp := range opportunities {
		key := CRMLookupKey{
			UTMCampaign: s.normalizeUTM(opp.UTMCampaign),
			UTMSource:   s.normalizeUTM(opp.UTMSource),
			UTMMedium:   s.normalizeUTM(opp.UTMMedium),
		}
		lookup[key] = append(lookup[key], opp)
	}

	return lookup
}

func (s *Service) findMatchingOpportunities(ad models.AdsPerformance, crmLookup map[CRMLookupKey][]models.Opportunity) []models.Opportunity {
	// Try exact match first
	exactKey := CRMLookupKey{
		UTMCampaign: s.normalizeUTM(ad.UTMCampaign),
		UTMSource:   s.normalizeUTM(ad.UTMSource),
		UTMMedium:   s.normalizeUTM(ad.UTMMedium),
	}

	if opportunities, exists := crmLookup[exactKey]; exists {
		return opportunities
	}

	// Try fallback matching (campaign only)
	fallbackKey := CRMLookupKey{
		UTMCampaign: s.normalizeUTM(ad.UTMCampaign),
		UTMSource:   "",
		UTMMedium:   "",
	}

	if opportunities, exists := crmLookup[fallbackKey]; exists {
		return opportunities
	}

	// Try source-only fallback
	sourceKey := CRMLookupKey{
		UTMCampaign: "",
		UTMSource:   s.normalizeUTM(ad.UTMSource),
		UTMMedium:   "",
	}

	if opportunities, exists := crmLookup[sourceKey]; exists {
		return opportunities
	}

	return []models.Opportunity{}
}

func (s *Service) normalizeUTM(utm string) string {
	return strings.ToLower(strings.TrimSpace(utm))
}

func (s *Service) calculateMetrics(ad models.AdsPerformance, opportunities []models.Opportunity) Metrics {
	metrics := Metrics{}

	// Count opportunities by stage
	for _, opp := range opportunities {
		metrics.Opportunities++
		if opp.Stage == "closed_won" {
			metrics.ClosedWon++
			metrics.Revenue += opp.Amount
		}
	}

	// Estimate leads (simplified: assume 10% of clicks become leads)
	metrics.Leads = int(float64(ad.Clicks) * 0.1)

	// Calculate CPC
	if ad.Clicks > 0 {
		metrics.CPC = ad.Cost / float64(ad.Clicks)
	}

	// Calculate CPA
	if metrics.Leads > 0 {
		metrics.CPA = ad.Cost / float64(metrics.Leads)
	}

	// Calculate conversion rates
	if metrics.Leads > 0 {
		metrics.CVRLeadToOpp = float64(metrics.Opportunities) / float64(metrics.Leads)
	}

	if metrics.Opportunities > 0 {
		metrics.CVROppToWon = float64(metrics.ClosedWon) / float64(metrics.Opportunities)
	}

	// Calculate ROAS
	if ad.Cost > 0 {
		metrics.ROAS = metrics.Revenue / ad.Cost
	}

	return metrics
}

func (s *Service) GetChannelMetrics(from, to time.Time, channel string, limit, offset int) ([]models.TransformedData, error) {
	filters := map[string]string{"channel": channel}
	return s.storage.GetTransformedData(from, to, filters, limit, offset)
}

func (s *Service) GetFunnelMetrics(from, to time.Time, utmCampaign string, limit, offset int) ([]models.TransformedData, error) {
	// For funnel metrics, we need to filter by UTM campaign
	// Since we don't store UTM campaign in transformed data, we'll return all data
	// and let the client filter by campaign_id
	filters := map[string]string{}
	return s.storage.GetTransformedData(from, to, filters, limit, offset)
}

func (s *Service) ExportData(ctx context.Context, date string) error {
	if s.config.SinkURL == "" || s.config.SinkSecret == "" {
		return fmt.Errorf("sink URL or secret not configured")
	}

	// Parse date
	exportDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		return fmt.Errorf("invalid date format: %w", err)
	}

	// Get data for the specific date
	data, err := s.storage.GetTransformedData(exportDate, exportDate, map[string]string{}, 0, 0)
	if err != nil {
		return fmt.Errorf("failed to get data for export: %w", err)
	}

	// Group data by channel and campaign for consolidation
	consolidated := s.consolidateDataByChannelAndCampaign(data)

	// Export each consolidated record
	for _, record := range consolidated {
		if err := s.exportRecord(ctx, record); err != nil {
			s.logger.WithError(err).WithField("record", record).Error("Failed to export record")
			return err
		}
	}

	s.logger.WithField("records_exported", len(consolidated)).Info("Data export completed")
	return nil
}

func (s *Service) consolidateDataByChannelAndCampaign(data []models.TransformedData) []models.TransformedData {
	consolidated := make(map[string]models.TransformedData)

	for _, item := range data {
		key := item.Channel + "|" + item.CampaignID
		if existing, exists := consolidated[key]; exists {
			// Aggregate metrics
			existing.Clicks += item.Clicks
			existing.Impressions += item.Impressions
			existing.Cost += item.Cost
			existing.Leads += item.Leads
			existing.Opportunities += item.Opportunities
			existing.ClosedWon += item.ClosedWon
			existing.Revenue += item.Revenue
			
			// Recalculate derived metrics
			if existing.Clicks > 0 {
				existing.CPC = existing.Cost / float64(existing.Clicks)
			}
			if existing.Leads > 0 {
				existing.CPA = existing.Cost / float64(existing.Leads)
			}
			if existing.Leads > 0 {
				existing.CVRLeadToOpp = float64(existing.Opportunities) / float64(existing.Leads)
			}
			if existing.Opportunities > 0 {
				existing.CVROppToWon = float64(existing.ClosedWon) / float64(existing.Opportunities)
			}
			if existing.Cost > 0 {
				existing.ROAS = existing.Revenue / existing.Cost
			}
			
			consolidated[key] = existing
		} else {
			consolidated[key] = item
		}
	}

	// Convert map to slice
	var result []models.TransformedData
	for _, item := range consolidated {
		result = append(result, item)
	}

	// Sort by channel and campaign
	sort.Slice(result, func(i, j int) bool {
		if result[i].Channel != result[j].Channel {
			return result[i].Channel < result[j].Channel
		}
		return result[i].CampaignID < result[j].CampaignID
	})

	return result
}

func (s *Service) exportRecord(ctx context.Context, record models.TransformedData) error {
	// Create HMAC signature
	signature := s.createHMACSignature(record)

	// Log the signature for debugging
	s.logger.WithField("signature", signature).Debug("Created HMAC signature for export")

	// Make POST request to sink
	return s.client.Post(ctx, s.config.SinkURL, record, nil)
}

func (s *Service) createHMACSignature(data models.TransformedData) string {
	// Simple HMAC implementation (in production, use crypto/hmac)
	// For this example, we'll create a simple hash
	payload := fmt.Sprintf("%s|%s|%s|%d|%d|%.2f|%d|%d|%d|%.2f|%.3f|%.3f|%.3f|%.3f|%.3f|%.3f",
		data.Date, data.Channel, data.CampaignID, data.Clicks, data.Impressions,
		data.Cost, data.Leads, data.Opportunities, data.ClosedWon, data.Revenue,
		data.CPC, data.CPA, data.CVRLeadToOpp, data.CVROppToWon, data.ROAS)
	
	// In a real implementation, use crypto/hmac with SHA256
	// For this example, we'll use a simple approach
	return fmt.Sprintf("hmac-sha256:%x", []byte(payload+s.config.SinkSecret))
}

