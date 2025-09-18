package models

import "time"

// External API Response Structures
type ExternalResponse struct {
	External ExternalData `json:"external"`
}

type ExternalData struct {
	Ads *AdsData `json:"ads,omitempty"`
	CRM *CRMData `json:"crm,omitempty"`
}

// Ads Data Models
type AdsData struct {
	Performance []AdsPerformance `json:"performance"`
}

type AdsPerformance struct {
	Date         string  `json:"date"`
	CampaignID   string  `json:"campaign_id"`
	Channel      string  `json:"channel"`
	Clicks       int     `json:"clicks"`
	Impressions  int     `json:"impressions"`
	Cost         float64 `json:"cost"`
	UTMCampaign  string  `json:"utm_campaign"`
	UTMSource    string  `json:"utm_source"`
	UTMMedium    string  `json:"utm_medium"`
}

// CRM Data Models
type CRMData struct {
	Opportunities []Opportunity `json:"opportunities"`
}

type Opportunity struct {
	OpportunityID string    `json:"opportunity_id"`
	ContactEmail  string    `json:"contact_email"`
	Stage         string    `json:"stage"`
	Amount        float64   `json:"amount"`
	CreatedAt     time.Time `json:"created_at"`
	UTMCampaign   string    `json:"utm_campaign"`
	UTMSource     string    `json:"utm_source"`
	UTMMedium     string    `json:"utm_medium"`
}

// Transformed Data Models
type TransformedData struct {
	Date         string  `json:"date"`
	Channel      string  `json:"channel"`
	CampaignID   string  `json:"campaign_id"`
	Clicks       int     `json:"clicks"`
	Impressions  int     `json:"impressions"`
	Cost         float64 `json:"cost"`
	Leads        int     `json:"leads"`
	Opportunities int    `json:"opportunities"`
	ClosedWon    int     `json:"closed_won"`
	Revenue      float64 `json:"revenue"`
	CPC          float64 `json:"cpc"`
	CPA          float64 `json:"cpa"`
	CVRLeadToOpp float64 `json:"cvr_lead_to_opp"`
	CVROppToWon  float64 `json:"cvr_opp_to_won"`
	ROAS         float64 `json:"roas"`
}

// API Request/Response Models
type IngestRequest struct {
	Since string `form:"since"`
}

type MetricsChannelRequest struct {
	From    string `form:"from" binding:"required"`
	To      string `form:"to" binding:"required"`
	Channel string `form:"channel" binding:"required"`
	Limit   int    `form:"limit"`
	Offset  int    `form:"offset"`
}

type MetricsFunnelRequest struct {
	From        string `form:"from" binding:"required"`
	To          string `form:"to" binding:"required"`
	UTMCampaign string `form:"utm_campaign" binding:"required"`
	Limit       int    `form:"limit"`
	Offset      int    `form:"offset"`
}

type ExportRequest struct {
	Date string `form:"date" binding:"required"`
}

type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	Version   string `json:"version"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

