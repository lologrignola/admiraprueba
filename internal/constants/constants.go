package constants

const (
	// Default values
	DefaultPort     = "8080"
	DefaultLogLevel = "info"
	
	// HTTP timeouts
	DefaultHTTPTimeout = 30
	DefaultMaxRetries  = 3
	DefaultRetryDelay  = 1
	
	// Pagination
	DefaultLimit  = 100
	MaxLimit      = 1000
	DefaultOffset = 0
	
	// Date format
	DateFormat = "2006-01-02"
	
	// Lead estimation
	LeadConversionRate = 0.1 // 10% of clicks become leads
	
	// API versions
	APIVersion = "v1"
	
	// Health check
	HealthStatusHealthy = "healthy"
	HealthStatusReady   = "ready"
	HealthStatusUnhealthy = "unhealthy"
	
	// Opportunity stages
	StageClosedWon = "closed_won"
	StageProposal  = "proposal"
	StageQualified = "qualified"
	StageLead      = "lead"
)
