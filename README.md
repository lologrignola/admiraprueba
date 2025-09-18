# Admira ETL Service

A Go-based ETL service that consumes Ads and CRM data, transforms it with UTM matching, and exposes marketing metrics via REST API.

## 🚀 Quick Start

### Prerequisites

- Go 1.21+
- Docker & Docker Compose (optional)

### Environment Setup

1. Copy the environment template:
```bash
cp env.example .env
```

2. Update `.env` with your Mocky API URLs:
```bash
ADS_API_URL=https://api.mocki.io/v2/YOUR-ADS-UUID
CRM_API_URL=https://api.mocki.io/v2/YOUR-CRM-UUID
SINK_URL=https://api.mocki.io/v2/YOUR-SINK-UUID
SINK_SECRET=admira_secret_example
PORT=8080
```

### Running the Service

#### Option 1: Direct Go Execution
```bash
# Install dependencies
make deps

# Run tests
make test

# Build and run
make run
```

#### Option 2: Docker Compose (Recommended)
```bash
# Build and run with Docker Compose
make docker-run

# Or run in background
make docker-run-bg

# View logs
make docker-logs

# Stop containers
make docker-stop
```

## 📡 API Endpoints

### Health Checks
- `GET /healthz` - Health check endpoint
- `GET /readyz` - Readiness check endpoint

### Data Ingestion
- `POST /api/v1/ingest/run?since=YYYY-MM-DD` - Run ETL process

**Example:**
```bash
curl -X POST "http://localhost:8080/api/v1/ingest/run?since=2025-01-01"
```

### Metrics Retrieval

#### Channel Metrics
- `GET /api/v1/metrics/channel?from=YYYY-MM-DD&to=YYYY-MM-DD&channel=google_ads&limit=100&offset=0`

**Example:**
```bash
curl "http://localhost:8080/api/v1/metrics/channel?from=2025-01-01&to=2025-01-31&channel=google_ads&limit=50"
```

**Response:**
```json
{
  "data": [
    {
      "date": "2025-01-01",
      "channel": "google_ads",
      "campaign_id": "C-1001",
      "clicks": 1200,
      "impressions": 45000,
      "cost": 350.75,
      "leads": 120,
      "opportunities": 8,
      "closed_won": 3,
      "revenue": 5000.0,
      "cpc": 0.292,
      "cpa": 14.03,
      "cvr_lead_to_opp": 0.32,
      "cvr_opp_to_won": 0.375,
      "roas": 14.25
    }
  ],
  "count": 1,
  "limit": 50,
  "offset": 0
}
```

#### Funnel Metrics
- `GET /api/v1/metrics/funnel?from=YYYY-MM-DD&to=YYYY-MM-DD&utm_campaign=back_to_school&limit=100&offset=0`

**Example:**
```bash
curl "http://localhost:8080/api/v1/metrics/funnel?from=2025-01-01&to=2025-01-31&utm_campaign=back_to_school"
```

### Data Export
- `POST /api/v1/export/run?date=YYYY-MM-DD` - Export consolidated data

**Example:**
```bash
curl -X POST "http://localhost:8080/api/v1/export/run?date=2025-01-01"
```

## 🧪 Complete API Examples

### Example 1: Health Check
```bash
curl http://localhost:8080/healthz
```

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2025-01-18T18:59:45Z",
  "version": "1.0.0"
}
```

### Example 2: Data Ingestion
```bash
curl -X POST "http://localhost:8080/api/v1/ingest/run?since=2025-01-01"
```

**Response:**
```json
{
  "message": "Ingestion completed successfully",
  "since": "2025-01-01"
}
```

### Example 3: Export Data
```bash
curl -X POST "http://localhost:8080/api/v1/export/run?date=2025-01-01"
```

**Response:**
```json
{
  "message": "Export completed successfully",
  "date": "2025-01-01"
}
```

## 🔧 Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `ADS_API_URL` | External Ads API URL | Required |
| `CRM_API_URL` | External CRM API URL | Required |
| `SINK_URL` | Export sink URL | Optional |
| `SINK_SECRET` | HMAC secret for export | Optional |
| `PORT` | Server port | 8080 |
| `LOG_LEVEL` | Logging level (debug, info, warn, error) | info |

### Data Sources

The service expects the following data formats:

#### Ads Data Format
```json
{
  "external": {
    "ads": {
      "performance": [
        {
          "date": "2025-01-01",
          "campaign_id": "C-1001",
          "channel": "google_ads",
          "clicks": 1200,
          "impressions": 45000,
          "cost": 350.75,
          "utm_campaign": "back_to_school",
          "utm_source": "google",
          "utm_medium": "cpc"
        }
      ]
    }
  }
}
```

#### CRM Data Format
```json
{
  "external": {
    "crm": {
      "opportunities": [
        {
          "opportunity_id": "O-9001",
          "contact_email": "ana@example.com",
          "stage": "closed_won",
          "amount": 5000.0,
          "created_at": "2025-01-05T10:22:00Z",
          "utm_campaign": "back_to_school",
          "utm_source": "google",
          "utm_medium": "cpc"
        }
      ]
    }
  }
}
```

## 🧮 Metrics Calculation

The service calculates the following marketing metrics:

- **CPC (Cost Per Click)**: `cost / clicks`
- **CPA (Cost Per Acquisition)**: `cost / leads`
- **CVR Lead→Opportunity**: `opportunities / leads`
- **CVR Opportunity→Won**: `closed_won / opportunities`
- **ROAS (Return on Ad Spend)**: `revenue / cost`

### UTM Matching Strategy

1. **Exact Match**: Match by `utm_campaign`, `utm_source`, and `utm_medium`
2. **Campaign Fallback**: Match by `utm_campaign` only
3. **Source Fallback**: Match by `utm_source` only

## 🧪 Testing

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run specific test package
go test -v ./internal/etl/

# Run specific test
go test -v -run TestTransformData ./internal/etl/
```

## 🏗️ Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   External APIs  │    │   ETL Service   │    │   REST API      │
│                 │    │                 │    │                 │
│ • Ads API       │───▶│ • Data Fetch    │───▶│ • Metrics       │
│ • CRM API       │    │ • Transform     │    │ • Health        │
│                 │    │ • Store         │    │ • Export        │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │
                                ▼
                       ┌─────────────────┐
                       │   Storage       │
                       │                 │
                       │ • In-Memory     │
                       │ • Transformed   │
                       │   Data          │
                       └─────────────────┘
```

## 📊 Key Features

- **Idempotent Ingestion**: Prevents duplicate data processing
- **Retry Logic**: Exponential backoff for external API calls
- **UTM Matching**: Flexible matching with fallback strategies
- **Metric Calculations**: Comprehensive marketing metrics
- **Health Monitoring**: Health and readiness endpoints
- **Structured Logging**: JSON-formatted logs with correlation IDs
- **Docker Support**: Containerized deployment
- **Comprehensive Testing**: Unit tests with coverage reporting

## 🚨 Error Handling

The service handles various error scenarios:

- **Network Timeouts**: Configurable timeouts with retry logic
- **API Errors**: Proper HTTP status code handling
- **Data Validation**: Input validation and sanitization
- **Division by Zero**: Protected metric calculations
- **Missing UTMs**: Graceful fallback matching

## 🔍 Monitoring

### Health Endpoints
- `/healthz`: Basic health check
- `/readyz`: Readiness check (validates external API connectivity)

### Logging
- Structured JSON logging
- Request correlation IDs
- Configurable log levels
- Error context and stack traces

## 📈 Performance Considerations

- **In-Memory Storage**: Fast data access for demo purposes
- **Concurrent Processing**: Goroutine-based concurrent API calls
- **Pagination Support**: Efficient data retrieval
- **Connection Pooling**: HTTP client connection reuse

## ⚠️ Assumptions & Limitations

### Technical Assumptions
- **Lead Estimation**: Assumes 10% of clicks become leads (simplified model)
- **UTM Matching**: Uses exact string matching with fallbacks
- **Data Format**: Assumes consistent date format (YYYY-MM-DD)

### Architecture Limitations
- **Storage**: In-memory storage (data lost on restart)
- **Scaling**: Single-instance deployment only
- **Processing**: Sequential data transformation

### Data Quality Limitations
- **UTM Fallbacks**: May create false positives
- **Currency**: No multi-currency support
- **Timezone**: No timezone handling

### Security Limitations
- **Authentication**: No API authentication
- **HMAC**: Simplified signature implementation

### Business Logic Limitations
- **Attribution**: First-touch attribution only
- **Funnel**: Linear conversion model
- **Multi-touch**: No multi-touch attribution support

## 🛠️ Development

### Project Structure
```
├── main.go                 # Application entry point
├── internal/
│   ├── api/               # HTTP handlers and routes
│   ├── config/            # Configuration management
│   ├── etl/              # ETL service and transformation logic
│   ├── http/             # HTTP client with retry logic
│   ├── models/           # Data models and structures
│   └── storage/          # Data storage interface
├── Dockerfile            # Container configuration
├── docker-compose.yml    # Multi-container setup
├── Makefile             # Build and run commands
└── README.md            # This file
```

### Adding New Metrics

1. Add metric calculation in `internal/etl/service.go`
2. Update `models.TransformedData` struct
3. Add corresponding tests
4. Update API documentation

## 📫 Submission

- **Repository**: [GitHub Link] (replace with your actual repository URL)
- **Live Demo**: Service running on localhost:8080
- **Documentation**: Complete API documentation in this README
- **Testing**: Comprehensive unit tests with coverage reporting
- **Examples**: 3+ cURL examples provided above
- **Limitations**: Detailed assumptions and limitations documented

## 📄 License

This project is part of the Admira technical assessment.

