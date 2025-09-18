package storage

import (
	"testing"
	"time"

	"admira-etl/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInMemoryStorage_StoreTransformedData(t *testing.T) {
	storage := NewInMemoryStorage()

	data := []models.TransformedData{
		{
			Date:        "2025-01-01",
			Channel:     "google_ads",
			CampaignID:  "C-1001",
			Clicks:      1000,
			Impressions: 50000,
			Cost:        250.0,
		},
		{
			Date:        "2025-01-02",
			Channel:     "facebook_ads",
			CampaignID:  "C-1002",
			Clicks:      800,
			Impressions: 40000,
			Cost:        200.0,
		},
	}

	err := storage.StoreTransformedData(data)
	require.NoError(t, err)

	// Verify data was stored
	from, _ := time.Parse("2006-01-01", "2025-01-01")
	to, _ := time.Parse("2006-01-01", "2025-01-02")
	retrieved, err := storage.GetTransformedData(from, to, map[string]string{}, 0, 0)
	require.NoError(t, err)
	assert.Len(t, retrieved, 2)
}

func TestInMemoryStorage_GetTransformedData(t *testing.T) {
	storage := NewInMemoryStorage()

	// Store test data
	data := []models.TransformedData{
		{
			Date:        "2025-01-01",
			Channel:     "google_ads",
			CampaignID:  "C-1001",
			Clicks:      1000,
			Impressions: 50000,
			Cost:        250.0,
		},
		{
			Date:        "2025-01-02",
			Channel:     "facebook_ads",
			CampaignID:  "C-1002",
			Clicks:      800,
			Impressions: 40000,
			Cost:        200.0,
		},
		{
			Date:        "2025-01-03",
			Channel:     "google_ads",
			CampaignID:  "C-1003",
			Clicks:      1200,
			Impressions: 60000,
			Cost:        300.0,
		},
	}

	err := storage.StoreTransformedData(data)
	require.NoError(t, err)

	tests := []struct {
		name     string
		from     string
		to       string
		filters  map[string]string
		limit    int
		offset   int
		expected int
	}{
		{
			name:     "get all data",
			from:     "2025-01-01",
			to:       "2025-01-03",
			filters:  map[string]string{},
			limit:    0,
			offset:   0,
			expected: 3,
		},
		{
			name:     "filter by channel",
			from:     "2025-01-01",
			to:       "2025-01-03",
			filters:  map[string]string{"channel": "google_ads"},
			limit:    0,
			offset:   0,
			expected: 2,
		},
		{
			name:     "filter by campaign",
			from:     "2025-01-01",
			to:       "2025-01-03",
			filters:  map[string]string{"campaign_id": "C-1001"},
			limit:    0,
			offset:   0,
			expected: 1,
		},
		{
			name:     "date range filter",
			from:     "2025-01-01",
			to:       "2025-01-02",
			filters:  map[string]string{},
			limit:    0,
			offset:   0,
			expected: 2,
		},
		{
			name:     "pagination",
			from:     "2025-01-01",
			to:       "2025-01-03",
			filters:  map[string]string{},
			limit:    2,
			offset:   0,
			expected: 2,
		},
		{
			name:     "pagination with offset",
			from:     "2025-01-01",
			to:       "2025-01-03",
			filters:  map[string]string{},
			limit:    2,
			offset:   1,
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			from, _ := time.Parse("2006-01-01", tt.from)
			to, _ := time.Parse("2006-01-01", tt.to)
			
			result, err := storage.GetTransformedData(from, to, tt.filters, tt.limit, tt.offset)
			require.NoError(t, err)
			assert.Len(t, result, tt.expected)
		})
	}
}

func TestInMemoryStorage_IngestionTime(t *testing.T) {
	storage := NewInMemoryStorage()

	// Test initial state
	lastTime, err := storage.GetLastIngestionTime()
	require.NoError(t, err)
	assert.True(t, lastTime.IsZero())

	// Set ingestion time
	now := time.Now()
	err = storage.SetLastIngestionTime(now)
	require.NoError(t, err)

	// Verify it was set
	retrievedTime, err := storage.GetLastIngestionTime()
	require.NoError(t, err)
	assert.Equal(t, now.Unix(), retrievedTime.Unix())
}

func TestInMemoryStorage_HasBeenIngested(t *testing.T) {
	storage := NewInMemoryStorage()

	// Test initial state
	assert.False(t, storage.HasBeenIngested("2025-01-01"))

	// Store data
	data := []models.TransformedData{
		{
			Date: "2025-01-01",
		},
	}

	err := storage.StoreTransformedData(data)
	require.NoError(t, err)

	// Verify ingestion tracking
	assert.True(t, storage.HasBeenIngested("2025-01-01"))
	assert.False(t, storage.HasBeenIngested("2025-01-02"))
}

