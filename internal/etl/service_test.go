package etl

import (
	"testing"
	"time"

	"admira-etl/internal/config"
	"admira-etl/internal/models"
	"admira-etl/internal/storage"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransformData(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Suppress logs during tests

	cfg := &config.Config{}
	store := storage.NewInMemoryStorage()
	service := NewService(cfg, store, logger)

	tests := []struct {
		name      string
		adsData   *models.AdsData
		crmData   *models.CRMData
		sinceTime time.Time
		expected  []models.TransformedData
	}{
		{
			name: "basic transformation with matching UTM",
			adsData: &models.AdsData{
				Performance: []models.AdsPerformance{
					{
						Date:         "2025-01-01",
						CampaignID:   "C-1001",
						Channel:      "google_ads",
						Clicks:       1000,
						Impressions:  50000,
						Cost:         250.0,
						UTMCampaign:  "back_to_school",
						UTMSource:    "google",
						UTMMedium:    "cpc",
					},
				},
			},
			crmData: &models.CRMData{
				Opportunities: []models.Opportunity{
					{
						OpportunityID: "O-9001",
						ContactEmail:  "test@example.com",
						Stage:         "closed_won",
						Amount:        5000.0,
						CreatedAt:     time.Now(),
						UTMCampaign:   "back_to_school",
						UTMSource:     "google",
						UTMMedium:     "cpc",
					},
					{
						OpportunityID: "O-9002",
						ContactEmail:  "test2@example.com",
						Stage:         "proposal",
						Amount:        3000.0,
						CreatedAt:     time.Now(),
						UTMCampaign:   "back_to_school",
						UTMSource:     "google",
						UTMMedium:     "cpc",
					},
				},
			},
			sinceTime: time.Time{},
			expected: []models.TransformedData{
				{
					Date:         "2025-01-01",
					Channel:      "google_ads",
					CampaignID:   "C-1001",
					Clicks:       1000,
					Impressions:  50000,
					Cost:         250.0,
					Leads:        100, // 10% of clicks
					Opportunities: 2,
					ClosedWon:    1,
					Revenue:      5000.0,
					CPC:          0.25, // 250 / 1000
					CPA:          2.5,  // 250 / 100
					CVRLeadToOpp: 0.02, // 2 / 100
					CVROppToWon:  0.5,  // 1 / 2
					ROAS:         20.0, // 5000 / 250
				},
			},
		},
		{
			name: "transformation with no matching CRM data",
			adsData: &models.AdsData{
				Performance: []models.AdsPerformance{
					{
						Date:         "2025-01-01",
						CampaignID:   "C-1001",
						Channel:      "google_ads",
						Clicks:       1000,
						Impressions:  50000,
						Cost:         250.0,
						UTMCampaign:  "back_to_school",
						UTMSource:    "google",
						UTMMedium:    "cpc",
					},
				},
			},
			crmData: &models.CRMData{
				Opportunities: []models.Opportunity{},
			},
			sinceTime: time.Time{},
			expected: []models.TransformedData{
				{
					Date:         "2025-01-01",
					Channel:      "google_ads",
					CampaignID:   "C-1001",
					Clicks:       1000,
					Impressions:  50000,
					Cost:         250.0,
					Leads:        100,
					Opportunities: 0,
					ClosedWon:    0,
					Revenue:      0.0,
					CPC:          0.25,
					CPA:          2.5,
					CVRLeadToOpp: 0.0,
					CVROppToWon:  0.0,
					ROAS:         0.0,
				},
			},
		},
		{
			name: "transformation with zero clicks",
			adsData: &models.AdsData{
				Performance: []models.AdsPerformance{
					{
						Date:         "2025-01-01",
						CampaignID:   "C-1001",
						Channel:      "google_ads",
						Clicks:       0,
						Impressions:  50000,
						Cost:         250.0,
						UTMCampaign:  "back_to_school",
						UTMSource:    "google",
						UTMMedium:    "cpc",
					},
				},
			},
			crmData: &models.CRMData{
				Opportunities: []models.Opportunity{},
			},
			sinceTime: time.Time{},
			expected: []models.TransformedData{
				{
					Date:         "2025-01-01",
					Channel:      "google_ads",
					CampaignID:   "C-1001",
					Clicks:       0,
					Impressions:  50000,
					Cost:         250.0,
					Leads:        0,
					Opportunities: 0,
					ClosedWon:    0,
					Revenue:      0.0,
					CPC:          0.0, // Division by zero protection
					CPA:          0.0, // Division by zero protection
					CVRLeadToOpp: 0.0,
					CVROppToWon:  0.0,
					ROAS:         0.0,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.transformData(tt.adsData, tt.crmData, tt.sinceTime)
			require.NoError(t, err)
			require.Len(t, result, len(tt.expected))

			for i, expected := range tt.expected {
				actual := result[i]
				assert.Equal(t, expected.Date, actual.Date)
				assert.Equal(t, expected.Channel, actual.Channel)
				assert.Equal(t, expected.CampaignID, actual.CampaignID)
				assert.Equal(t, expected.Clicks, actual.Clicks)
				assert.Equal(t, expected.Impressions, actual.Impressions)
				assert.Equal(t, expected.Cost, actual.Cost)
				assert.Equal(t, expected.Leads, actual.Leads)
				assert.Equal(t, expected.Opportunities, actual.Opportunities)
				assert.Equal(t, expected.ClosedWon, actual.ClosedWon)
				assert.Equal(t, expected.Revenue, actual.Revenue)
				assert.InDelta(t, expected.CPC, actual.CPC, 0.001)
				assert.InDelta(t, expected.CPA, actual.CPA, 0.001)
				assert.InDelta(t, expected.CVRLeadToOpp, actual.CVRLeadToOpp, 0.001)
				assert.InDelta(t, expected.CVROppToWon, actual.CVROppToWon, 0.001)
				assert.InDelta(t, expected.ROAS, actual.ROAS, 0.001)
			}
		})
	}
}

func TestBuildCRMLookup(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	cfg := &config.Config{}
	store := storage.NewInMemoryStorage()
	service := NewService(cfg, store, logger)

	opportunities := []models.Opportunity{
		{
			OpportunityID: "O-9001",
			UTMCampaign:   "back_to_school",
			UTMSource:     "google",
			UTMMedium:     "cpc",
		},
		{
			OpportunityID: "O-9002",
			UTMCampaign:   "back_to_school",
			UTMSource:     "google",
			UTMMedium:     "cpc",
		},
		{
			OpportunityID: "O-9003",
			UTMCampaign:   "summer_sale",
			UTMSource:     "facebook",
			UTMMedium:     "cpc",
		},
	}

	lookup := service.buildCRMLookup(opportunities)

	// Test exact match
	exactKey := CRMLookupKey{
		UTMCampaign: "back_to_school",
		UTMSource:   "google",
		UTMMedium:   "cpc",
	}
	assert.Len(t, lookup[exactKey], 2)

	// Test different campaign
	differentKey := CRMLookupKey{
		UTMCampaign: "summer_sale",
		UTMSource:   "facebook",
		UTMMedium:   "cpc",
	}
	assert.Len(t, lookup[differentKey], 1)
}

func TestFindMatchingOpportunities(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	cfg := &config.Config{}
	store := storage.NewInMemoryStorage()
	service := NewService(cfg, store, logger)

	// Setup CRM lookup
	crmLookup := map[CRMLookupKey][]models.Opportunity{
		{
			UTMCampaign: "back_to_school",
			UTMSource:   "google",
			UTMMedium:   "cpc",
		}: {
			{OpportunityID: "O-9001"},
			{OpportunityID: "O-9002"},
		},
		{
			UTMCampaign: "back_to_school",
			UTMSource:   "",
			UTMMedium:   "",
		}: {
			{OpportunityID: "O-9003"},
		},
	}

	tests := []struct {
		name         string
		ad           models.AdsPerformance
		expectedLen  int
		expectedIDs  []string
	}{
		{
			name: "exact match",
			ad: models.AdsPerformance{
				UTMCampaign: "back_to_school",
				UTMSource:   "google",
				UTMMedium:   "cpc",
			},
			expectedLen: 2,
			expectedIDs: []string{"O-9001", "O-9002"},
		},
		{
			name: "fallback match",
			ad: models.AdsPerformance{
				UTMCampaign: "back_to_school",
				UTMSource:   "facebook",
				UTMMedium:   "cpc",
			},
			expectedLen: 1,
			expectedIDs: []string{"O-9003"},
		},
		{
			name: "no match",
			ad: models.AdsPerformance{
				UTMCampaign: "winter_sale",
				UTMSource:   "twitter",
				UTMMedium:   "cpc",
			},
			expectedLen: 0,
			expectedIDs: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.findMatchingOpportunities(tt.ad, crmLookup)
			assert.Len(t, result, tt.expectedLen)

			for i, expectedID := range tt.expectedIDs {
				assert.Equal(t, expectedID, result[i].OpportunityID)
			}
		})
	}
}

func TestCalculateMetrics(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	cfg := &config.Config{}
	store := storage.NewInMemoryStorage()
	service := NewService(cfg, store, logger)

	tests := []struct {
		name         string
		ad           models.AdsPerformance
		opportunities []models.Opportunity
		expected     Metrics
	}{
		{
			name: "normal metrics calculation",
			ad: models.AdsPerformance{
				Clicks: 1000,
				Cost:   250.0,
			},
			opportunities: []models.Opportunity{
				{Stage: "closed_won", Amount: 5000.0},
				{Stage: "closed_won", Amount: 3000.0},
				{Stage: "proposal", Amount: 2000.0},
			},
			expected: Metrics{
				Leads:         100, // 10% of clicks
				Opportunities: 3,
				ClosedWon:     2,
				Revenue:       8000.0,
				CPC:           0.25, // 250 / 1000
				CPA:           2.5,  // 250 / 100
				CVRLeadToOpp:  0.03, // 3 / 100
				CVROppToWon:   0.6667, // 2 / 3
				ROAS:          32.0, // 8000 / 250
			},
		},
		{
			name: "zero clicks protection",
			ad: models.AdsPerformance{
				Clicks: 0,
				Cost:   250.0,
			},
			opportunities: []models.Opportunity{},
			expected: Metrics{
				Leads:         0,
				Opportunities: 0,
				ClosedWon:     0,
				Revenue:       0.0,
				CPC:           0.0, // Division by zero protection
				CPA:           0.0, // Division by zero protection
				CVRLeadToOpp:  0.0,
				CVROppToWon:   0.0,
				ROAS:          0.0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.calculateMetrics(tt.ad, tt.opportunities)

			assert.Equal(t, tt.expected.Leads, result.Leads)
			assert.Equal(t, tt.expected.Opportunities, result.Opportunities)
			assert.Equal(t, tt.expected.ClosedWon, result.ClosedWon)
			assert.Equal(t, tt.expected.Revenue, result.Revenue)
			assert.InDelta(t, tt.expected.CPC, result.CPC, 0.001)
			assert.InDelta(t, tt.expected.CPA, result.CPA, 0.001)
			assert.InDelta(t, tt.expected.CVRLeadToOpp, result.CVRLeadToOpp, 0.001)
			assert.InDelta(t, tt.expected.CVROppToWon, result.CVROppToWon, 0.001)
			assert.InDelta(t, tt.expected.ROAS, result.ROAS, 0.001)
		})
	}
}

func TestNormalizeUTM(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	cfg := &config.Config{}
	store := storage.NewInMemoryStorage()
	service := NewService(cfg, store, logger)

	tests := []struct {
		input    string
		expected string
	}{
		{"Back_To_School", "back_to_school"},
		{"  GOOGLE  ", "google"},
		{"CPC", "cpc"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := service.normalizeUTM(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

