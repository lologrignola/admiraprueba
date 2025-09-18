# System Design Document - Admira ETL Service

## 🎯 Overview

This document outlines the architectural decisions, design patterns, and trade-offs made in the Admira ETL service implementation.

## 🏗️ Architecture Decisions

### 1. Idempotencia & Reprocesamiento

**Decision**: Implemented ingestion tracking with date-based idempotency keys.

**Implementation**:
- Each ingestion run tracks processed dates in `storage.ingestionTimes`
- Duplicate ingestion requests for the same date are detected and skipped
- Uses `since` parameter to control reprocessing scope

**Trade-offs**:
- ✅ Prevents duplicate data processing
- ✅ Enables safe retry mechanisms
- ❌ Requires additional storage for tracking
- ❌ Limited to date-level granularity

**Future Considerations**:
- Implement hash-based content deduplication for finer granularity
- Add batch-level idempotency keys for partial reprocessing

### 2. Particionamiento & Retención

**Decision**: In-memory storage with date-based partitioning.

**Implementation**:
- Data stored in chronological order with date filtering
- Simple pagination with `limit`/`offset` parameters
- No explicit retention policy (demo implementation)

**Trade-offs**:
- ✅ Simple implementation and fast access
- ✅ Easy to understand and debug
- ❌ Limited by available memory
- ❌ No persistence across restarts

**Future Considerations**:
- Implement time-based partitioning (daily/monthly buckets)
- Add configurable retention policies
- Consider persistent storage (PostgreSQL, ClickHouse)
- Implement data archival strategies

### 3. Concurrencia & Throughput

**Decision**: Goroutine-based concurrent processing with worker pools.

**Implementation**:
- Concurrent HTTP client calls to external APIs
- Sequential data transformation (maintains data consistency)
- In-memory storage with mutex protection

**Trade-offs**:
- ✅ Leverages Go's concurrency model effectively
- ✅ Simple synchronization with mutexes
- ❌ Limited by single-threaded transformation
- ❌ No horizontal scaling

**Future Considerations**:
- Implement worker pools for transformation pipeline
- Add message queues (Kafka, RabbitMQ) for async processing
- Consider distributed processing with multiple service instances
- Implement backpressure mechanisms

### 4. Calidad de Datos (UTMs Ausentes y Fallbacks)

**Decision**: Multi-level UTM matching with graceful fallbacks.

**Implementation**:
1. **Exact Match**: `utm_campaign` + `utm_source` + `utm_medium`
2. **Campaign Fallback**: `utm_campaign` only
3. **Source Fallback**: `utm_source` only
4. **Normalization**: Case-insensitive, trimmed matching

**Trade-offs**:
- ✅ Handles incomplete UTM data gracefully
- ✅ Flexible matching strategy
- ✅ Maintains data relationships
- ❌ May create false positives with broad fallbacks
- ❌ No confidence scoring for matches

**Future Considerations**:
- Implement confidence scoring for UTM matches
- Add fuzzy matching for typos and variations
- Create UTM validation rules and alerts
- Implement UTM enrichment from external sources

### 5. Observabilidad (Logs y Métricas Útiles)

**Decision**: Structured JSON logging with request correlation.

**Implementation**:
- JSON-formatted logs with consistent fields
- Request ID correlation across service calls
- Configurable log levels
- Health check endpoints (`/healthz`, `/readyz`)

**Trade-offs**:
- ✅ Machine-readable logs for analysis
- ✅ Easy correlation of related events
- ✅ Standard health check patterns
- ❌ No metrics collection (Prometheus)
- ❌ No distributed tracing

**Future Considerations**:
- Add Prometheus metrics (request count, latency, error rates)
- Implement distributed tracing (Jaeger, Zipkin)
- Add business metrics (ingestion volume, match rates)
- Create alerting rules for critical failures

### 6. Evolución en el Ecosistema Admira

**Decision**: Modular architecture with clear separation of concerns.

**Implementation**:
- Clean separation between HTTP, ETL, and storage layers
- Interface-based design for easy testing and replacement
- Configuration-driven external API integration

**Trade-offs**:
- ✅ Easy to extend and modify
- ✅ Testable components
- ✅ Clear boundaries between concerns
- ❌ No plugin architecture
- ❌ Limited to single data source types

**Future Considerations**:

#### Data Lake Integration
- **Current**: Single service with in-memory storage
- **Future**: 
  - Implement data lake connectors (S3, GCS, Azure Blob)
  - Add data format support (Parquet, Avro, Delta Lake)
  - Implement data versioning and schema evolution
  - Add data quality monitoring and validation

#### ETL Pipeline Evolution
- **Current**: Synchronous processing with simple transformations
- **Future**:
  - Implement streaming ETL with Apache Kafka
  - Add complex transformation rules engine
  - Implement data lineage tracking
  - Add real-time vs batch processing modes

#### API Contract Management
- **Current**: REST API with versioned endpoints
- **Future**:
  - Implement OpenAPI/Swagger specifications
  - Add API versioning strategy
  - Implement contract testing
  - Add API rate limiting and authentication

#### Scalability Patterns
- **Current**: Single-instance deployment
- **Future**:
  - Implement horizontal scaling with load balancers
  - Add database clustering and replication
  - Implement caching layers (Redis, Memcached)
  - Add auto-scaling based on metrics

## 🔧 Technical Implementation Details

### Data Flow Architecture

```
External APIs → HTTP Client → ETL Service → Storage → REST API
     ↓              ↓            ↓           ↓         ↓
  Retry Logic   Timeout      Transform   In-Memory   JSON
  Backoff       Handling     UTM Match   Storage     Response
```

### Error Handling Strategy

1. **Network Errors**: Retry with exponential backoff
2. **Data Errors**: Log and skip invalid records
3. **Business Logic Errors**: Return meaningful error messages
4. **System Errors**: Graceful degradation with health checks

### Security Considerations

- **Input Validation**: All API inputs validated and sanitized
- **Error Information**: Limited error details in responses
- **HMAC Signatures**: Export functionality includes signature verification
- **Environment Variables**: Sensitive data in environment variables

## 📊 Performance Characteristics

### Current Limitations
- **Memory Bound**: Limited by available RAM
- **Single Thread**: Transformation processing is sequential
- **No Caching**: Every request hits storage
- **No Compression**: Raw data storage without compression

### Optimization Opportunities
- **Batch Processing**: Process multiple records together
- **Data Compression**: Compress stored data
- **Indexing**: Add indexes for common query patterns
- **Connection Pooling**: Reuse HTTP connections

## 🚀 Deployment Considerations

### Current Deployment
- **Docker**: Single container deployment
- **Environment**: Configuration via environment variables
- **Health Checks**: Basic health and readiness endpoints
- **Logging**: Structured JSON logs to stdout

### Production Readiness Gaps
- **Secrets Management**: No integration with secret stores
- **Configuration Management**: No external config service
- **Service Discovery**: No service registry integration
- **Load Balancing**: No load balancer configuration

## 🔮 Future Roadmap

### Phase 1: Production Hardening
- Add persistent storage (PostgreSQL)
- Implement comprehensive monitoring
- Add authentication and authorization
- Create deployment automation

### Phase 2: Scale & Performance
- Implement horizontal scaling
- Add caching layers
- Optimize data processing pipeline
- Add real-time processing capabilities

### Phase 3: Advanced Features
- Add machine learning for UTM matching
- Implement data quality scoring
- Add advanced analytics and reporting
- Create data marketplace integration

## 📝 Decision Log

| Date | Decision | Rationale | Impact |
|------|----------|-----------|---------|
| 2025-01-XX | In-memory storage | Simple implementation for demo | Fast access, limited scale |
| 2025-01-XX | Sequential transformation | Data consistency | Simple logic, limited concurrency |
| 2025-01-XX | Multi-level UTM matching | Handle incomplete data | Flexible matching, potential false positives |
| 2025-01-XX | JSON structured logging | Observability | Machine-readable, no metrics |

This system design provides a solid foundation for the Admira ETL service while maintaining flexibility for future evolution and scaling requirements.

