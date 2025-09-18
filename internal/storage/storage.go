package storage

import (
	"sync"
	"time"

	"admira-etl/internal/models"
)

type Storage interface {
	StoreTransformedData(data []models.TransformedData) error
	GetTransformedData(from, to time.Time, filters map[string]string, limit, offset int) ([]models.TransformedData, error)
	GetLastIngestionTime() (time.Time, error)
	SetLastIngestionTime(t time.Time) error
}

type InMemoryStorage struct {
	mu              sync.RWMutex
	data            []models.TransformedData
	lastIngestion   time.Time
	ingestionTimes  map[string]time.Time // Track ingestion times by date for idempotency
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		data:           make([]models.TransformedData, 0),
		ingestionTimes: make(map[string]time.Time),
	}
}

func (s *InMemoryStorage) StoreTransformedData(data []models.TransformedData) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Append new data
	s.data = append(s.data, data...)

	// Update ingestion times for idempotency
	for _, item := range data {
		s.ingestionTimes[item.Date] = time.Now()
	}

	return nil
}

func (s *InMemoryStorage) GetTransformedData(from, to time.Time, filters map[string]string, limit, offset int) ([]models.TransformedData, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var filtered []models.TransformedData

	for _, item := range s.data {
		itemDate, err := time.Parse("2006-01-02", item.Date)
		if err != nil {
			continue
		}

		// Filter by date range
		if itemDate.Before(from) || itemDate.After(to) {
			continue
		}

		// Apply additional filters
		if !s.matchesFilters(item, filters) {
			continue
		}

		filtered = append(filtered, item)
	}

	// Apply pagination
	start := offset
	end := offset + limit
	if limit <= 0 {
		end = len(filtered)
	}
	if start >= len(filtered) {
		return []models.TransformedData{}, nil
	}
	if end > len(filtered) {
		end = len(filtered)
	}

	return filtered[start:end], nil
}

func (s *InMemoryStorage) matchesFilters(item models.TransformedData, filters map[string]string) bool {
	for key, value := range filters {
		switch key {
		case "channel":
			if item.Channel != value {
				return false
			}
		case "campaign_id":
			if item.CampaignID != value {
				return false
			}
		case "utm_campaign":
			// This would need to be stored in the transformed data
			// For now, we'll skip this filter
		}
	}
	return true
}

func (s *InMemoryStorage) GetLastIngestionTime() (time.Time, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lastIngestion, nil
}

func (s *InMemoryStorage) SetLastIngestionTime(t time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastIngestion = t
	return nil
}

func (s *InMemoryStorage) HasBeenIngested(date string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, exists := s.ingestionTimes[date]
	return exists
}

