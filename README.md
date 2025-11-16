# Hotel Aggregator

A high-performance HTTP service that aggregates hotel search results from multiple providers concurrently, with built-in caching, rate limiting, and deduplication.

## Features

- **Concurrent Provider Querying**: Searches multiple hotel providers in parallel using errgroup
- **Intelligent Deduplication**: Automatically selects the best price for duplicate hotels
- **In-Memory Caching**: Configurable TTL cache to reduce provider load and improve response times (default: 30s)
- **Rate Limiting**: Configurable request limit per IP address (default: 10 requests per minute)
- **Structured Logging**: JSON logging with request IDs for full request tracing using log/slog
- **Graceful Shutdown**: Properly handles SIGTERM and SIGINT signals
- **Health & Metrics**: Built-in health check and Prometheus-compatible metrics endpoints
- **Environment-Based Configuration**: Easily configure via environment variables

## Quick Start

### Prerequisites

- Go 1.25.0 or higher (for local development)
- Docker and Docker Compose (for containerized deployment)

### Running with Docker (Recommended)

The easiest way to run the hotel aggregator is using Docker Compose:

1. **Copy the example environment file and customize if needed:**

```bash
cp .env.example .env
# Edit .env with your preferred values (optional - defaults work out of the box)
```

2. **Build and run the container:**

```bash
docker-compose up --build
```

The server will start on `http://localhost:8080`

3. **Run in detached mode (background):**

```bash
docker-compose up -d
```

4. **View logs:**

```bash
docker-compose logs -f
```

5. **Stop the service:**

```bash
docker-compose down
```

**Configuration:**

All configuration is managed through the [.env](.env) file. You can customize any of the following variables:

- `SERVER_PORT`: Port on which the server listens (default: 8080)
- `CACHE_TTL_SECONDS`: Cache time-to-live in seconds (default: 30)
- `RATE_LIMIT_MAX_REQUESTS`: Max requests per IP (default: 10)
- `RATE_LIMIT_WINDOW_MINUTES`: Rate limit window in minutes (default: 1)
- `PROVIDER_TIMEOUT_SECONDS`: Provider API timeout in seconds (default: 2)

After modifying [.env](.env), restart the container with `docker-compose restart`.

### Running Locally (Development)

```bash
go run cmd/server/main.go
```

The server will start on `http://localhost:8080`

### Example Search Request

```bash
curl "http://localhost:8080/search?city=London&checkin=2025-12-01&nights=3&adults=2"
```

Example response:

```json
{
  "search": {
    "city": "London",
    "checkin": "2025-12-01",
    "nights": 3,
    "adults": 2
  },
  "stats": {
    "providers_total": 3,
    "providers_succeeded": 3,
    "providers_failed": 0,
    "cache": "miss",
    "duration_ms": 245
  },
  "hotels": [
    {
      "hotel_id": "hotel_123",
      "name": "Grand Plaza Hotel",
      "currency": "USD",
      "price": 150.00
    }
  ]
}
```

## Configuration

The application can be configured using environment variables. All configuration values have sensible defaults, so the application works out of the box without any configuration.

### Environment Variables

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `SERVER_PORT` | Port on which the HTTP server listens | `8080` | `3000` |
| `CACHE_TTL_SECONDS` | Time-to-live for cached search results in seconds | `30` | `60` |
| `RATE_LIMIT_MAX_REQUESTS` | Maximum number of requests allowed per IP address within the rate limit window | `10` | `20` |
| `RATE_LIMIT_WINDOW_MINUTES` | Duration of the rate limit window in minutes | `1` | `5` |
| `PROVIDER_TIMEOUT_SECONDS` | Timeout for provider API calls in seconds | `2` | `5` |

### Using Environment Variables

**Option 1: Set environment variables directly**

```bash
export SERVER_PORT=3000
export CACHE_TTL_SECONDS=60
export RATE_LIMIT_MAX_REQUESTS=20
go run cmd/server/main.go
```

**Option 2: Use a .env file**

Copy the example file and customize it:

```bash
cp .env.example .env
# Edit .env with your preferred values
```

Then load it before running:

```bash
source .env
go run cmd/server/main.go
```

**Option 3: Inline with the command**

```bash
SERVER_PORT=3000 CACHE_TTL_SECONDS=60 go run cmd/server/main.go
```

### Configuration Validation

The application validates all configuration values on startup:

- All duration and count values must be positive integers
- Invalid values will cause the application to fail with a clear error message
- Configuration values are logged on startup for verification

Example startup log output:

```
2025-11-16 10:00:00 Starting hotel aggregator server...
2025-11-16 10:00:00 Loading configuration...
2025-11-16 10:00:00 Configuration:
2025-11-16 10:00:00   Server Port: 8080
2025-11-16 10:00:00   Cache TTL: 30s
2025-11-16 10:00:00   Rate Limit Max Requests: 10
2025-11-16 10:00:00   Rate Limit Window: 1m0s
2025-11-16 10:00:00   Provider Timeout: 2s
```

## API Reference

### GET /search

Search for hotels across all providers.

**Query Parameters:**

| Parameter | Type   | Required | Description                          |
|-----------|--------|----------|--------------------------------------|
| city      | string | Yes      | City name (e.g., "London")          |
| checkin   | string | Yes      | Check-in date (format: YYYY-MM-DD)  |
| nights    | int    | Yes      | Number of nights (positive integer) |
| adults    | int    | Yes      | Number of adults (positive integer) |

**Response:**

```json
{
  "search": {
    "city": "string",
    "checkin": "string",
    "nights": 0,
    "adults": 0
  },
  "stats": {
    "providers_total": 0,
    "providers_succeeded": 0,
    "providers_failed": 0,
    "cache": "hit|miss",
    "duration_ms": 0
  },
  "hotels": [
    {
      "hotel_id": "string",
      "name": "string",
      "currency": "string",
      "price": 0.00
    }
  ]
}
```

**Status Codes:**

- `200 OK`: Successful search
- `400 Bad Request`: Invalid parameters
- `429 Too Many Requests`: Rate limit exceeded
- `500 Internal Server Error`: Server error

### GET /healthz

Health check endpoint.

**Response:**

```json
{
  "status": "ok"
}
```

### GET /metrics

Get service metrics including per-provider statistics. Supports both JSON and Prometheus formats.

**Query Parameters:**

| Parameter | Type   | Required | Description                          |
|-----------|--------|----------|--------------------------------------|
| format    | string | No       | Output format: "prometheus" for Prometheus text format, omit for JSON (default) |

**JSON Format (Default):**

```bash
curl "http://localhost:8080/metrics"
```

Response:

```json
{
  "requests_total": 1234,
  "cache_hits": 450,
  "cache_misses": 784,
  "provider_errors": {
    "Mock1": 12,
    "Mock2": 8,
    "Mock3": 15
  },
  "provider_successes": {
    "Mock1": 772,
    "Mock2": 776,
    "Mock3": 769
  }
}
```

**Prometheus Format:**

```bash
curl "http://localhost:8080/metrics?format=prometheus"
```

Response (Prometheus text exposition format):

```
# HELP hostaggr_requests_total Total number of search requests received
# TYPE hostaggr_requests_total counter
hostaggr_requests_total 1234

# HELP hostaggr_cache_hits_total Number of requests served from cache
# TYPE hostaggr_cache_hits_total counter
hostaggr_cache_hits_total 450

# HELP hostaggr_cache_misses_total Number of requests that required provider queries
# TYPE hostaggr_cache_misses_total counter
hostaggr_cache_misses_total 784

# HELP hostaggr_provider_requests_total Total number of requests per provider by status
# TYPE hostaggr_provider_requests_total counter
hostaggr_provider_requests_total{provider="Mock1",status="success"} 772
hostaggr_provider_requests_total{provider="Mock2",status="success"} 776
hostaggr_provider_requests_total{provider="Mock3",status="success"} 769
hostaggr_provider_requests_total{provider="Mock1",status="error"} 12
hostaggr_provider_requests_total{provider="Mock2",status="error"} 8
hostaggr_provider_requests_total{provider="Mock3",status="error"} 15
```

**Metrics Explained:**

- `requests_total`: Total number of search requests received
- `cache_hits`: Number of requests served from cache
- `cache_misses`: Number of requests that required provider queries
- `provider_errors`: Per-provider count of failed queries (timeouts, errors)
- `provider_successes`: Per-provider count of successful queries

Use these metrics to monitor:
- Overall system load (`requests_total`)
- Cache effectiveness (hit rate = `cache_hits` / `requests_total`)
- Individual provider health (success rate per provider)
- Provider reliability comparison

**Prometheus Scraping Configuration:**

Add this to your `prometheus.yml` to scrape metrics:

```yaml
scrape_configs:
  - job_name: 'hostaggr'
    scrape_interval: 15s
    static_configs:
      - targets: ['localhost:8080']
    params:
      format: ['prometheus']
```

## Architecture

### Concurrent Provider Querying

The aggregator uses [golang.org/x/sync/errgroup](https://pkg.go.dev/golang.org/x/sync/errgroup) to query all hotel providers concurrently:

- Each provider is queried in a separate goroutine
- All queries run in parallel with a 2-second timeout
- Provider failures don't block successful responses
- Statistics track succeeded vs failed providers

See [internal/search/aggregator.go:104-138](internal/search/aggregator.go#L104-L138) for implementation details.

### Deduplication Strategy

When multiple providers return the same hotel:

1. Hotels are identified by `hotel_id`
2. For duplicates, the **lowest price** is selected
3. Results are sorted by price (ascending)

This ensures users always see the best available price. See [internal/search/aggregator.go:159-183](internal/search/aggregator.go#L159-L183).

### Caching Approach

In-memory cache with the following characteristics:

- **TTL**: 30 seconds (configurable)
- **Key**: Hash of search parameters (city, checkin, nights, adults)
- **Thread-safe**: Uses RWMutex for concurrent access
- **Automatic cleanup**: Background goroutine evicts expired entries every 10 seconds

Cache hits are tracked in the response stats and metrics endpoint. See [internal/search/cache.go](internal/search/cache.go).

### Rate Limiting

Token bucket algorithm implementation:

- **Limit**: 10 requests per minute per IP address
- **Granular**: Per-IP tracking using remote address
- **Thread-safe**: Mutex-protected bucket map
- **Auto-cleanup**: Removes unused buckets after 5 minutes of inactivity

Returns `429 Too Many Requests` when limit is exceeded. See [internal/search/ratelimit.go](internal/search/ratelimit.go).

### Structured Logging

The application uses Go's standard library `log/slog` for structured JSON logging with request tracing:

**Features:**

- **JSON Format**: All logs are output in JSON format for easy parsing
- **Request IDs**: Every request gets a unique ID that propagates through the entire request lifecycle
- **Request Tracing**: Follow a request through HTTP handler → Aggregator → Provider queries
- **X-Request-ID Header**: Request ID is returned in response headers
- **Comprehensive Events**: Logs key events including cache hits/misses, rate limits, provider success/failures

**Example Log Output:**

```json
{"time":"2025-11-16T10:30:45.123Z","level":"INFO","msg":"request started","method":"GET","path":"/search","remote_addr":"192.168.1.1:54321","request_id":"a1b2c3d4e5f6g7h8"}
{"time":"2025-11-16T10:30:45.234Z","level":"INFO","msg":"cache miss","request_id":"a1b2c3d4e5f6g7h8","city":"Tokyo"}
{"time":"2025-11-16T10:30:45.345Z","level":"INFO","msg":"provider query succeeded","request_id":"a1b2c3d4e5f6g7h8","provider_name":"Mock1","duration_ms":98,"hotel_count":4}
{"time":"2025-11-16T10:30:45.356Z","level":"WARN","msg":"provider query failed","request_id":"a1b2c3d4e5f6g7h8","provider_name":"Mock2","duration_ms":102,"error":"Mock2: random provider failure"}
{"time":"2025-11-16T10:30:45.445Z","level":"INFO","msg":"provider query succeeded","request_id":"a1b2c3d4e5f6g7h8","provider_name":"Mock3","duration_ms":187,"hotel_count":4}
{"time":"2025-11-16T10:30:45.567Z","level":"INFO","msg":"request completed","status":200,"duration_ms":444,"request_id":"a1b2c3d4e5f6g7h8"}
```

**Filtering Logs by Request ID:**

Using `jq` to filter logs for a specific request:

```bash
# Follow a specific request through all log entries
cat logs.json | jq 'select(.request_id=="a1b2c3d4e5f6g7h8")'

# Extract just the message and timestamp
cat logs.json | jq -r 'select(.request_id=="a1b2c3d4e5f6g7h8") | "\(.time) \(.msg)"'

# Find all provider errors
cat logs.json | jq 'select(.msg=="provider query failed")'

# Calculate average request duration
cat logs.json | jq -s 'map(select(.msg=="request completed")) | map(.duration_ms) | add/length'
```

**Log Events:**

| Event | Level | Fields | Description |
|-------|-------|--------|-------------|
| `request started` | INFO | method, path, remote_addr, request_id | HTTP request received |
| `request completed` | INFO | status, duration_ms, request_id | HTTP request finished |
| `cache hit` | INFO | request_id, city, hotel_count | Search result served from cache |
| `cache miss` | INFO | request_id, city | Search required provider queries |
| `provider query succeeded` | INFO | request_id, provider_name, duration_ms, hotel_count | Provider returned results successfully |
| `provider query failed` | WARN | request_id, provider_name, duration_ms, error | Provider query failed or timed out |
| `rate limit exceeded` | WARN | request_id, ip | Client exceeded rate limit |
| `search failed` | ERROR | request_id, error | Internal error during search |

**Request ID Propagation:**

The request ID flows through the entire application stack:

1. **Middleware**: Generates unique ID and adds to context
2. **HTTP Response**: Returns `X-Request-ID` header
3. **Handlers**: Extract from context for logging
4. **Aggregator**: Passes through context to provider queries
5. **Providers**: Available in context for detailed logging

This allows correlation of all log entries for a single request, making debugging significantly easier.

## Running Tests

### Run all tests with race detection:

```bash
go test -race ./...
```

### Run tests with verbose output:

```bash
go test -v -race ./...
```

### Run specific test:

```bash
go test -v ./tests -run TestAggregator
```

### Test Coverage:

The test suite includes:

- **Aggregator tests**: Concurrent querying, deduplication, validation
- **Cache tests**: TTL behavior, concurrent access, cleanup
- **Rate limiter tests**: Bucket refill, per-IP limiting, cleanup

## Project Structure

```
hostaggr/
├── cmd/
│   └── server/
│       └── main.go           # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go         # Configuration management
│   ├── http/
│   │   └── handlers.go       # HTTP handlers
│   ├── models/
│   │   ├── hotel.go          # Hotel data structures
│   │   ├── search_request.go # Search request model
│   │   └── search_result.go  # Search response model
│   ├── obs/
│   │   └── metrics.go        # Metrics collection
│   ├── providers/
│   │   ├── provider.go       # Provider interface
│   │   ├── mock1.go          # Mock provider 1
│   │   ├── mock2.go          # Mock provider 2
│   │   └── mock3.go          # Mock provider 3
│   └── search/
│       ├── aggregator.go     # Core aggregation logic
│       ├── cache.go          # In-memory cache
│       └── ratelimit.go      # Rate limiting
├── tests/
│   ├── aggregator_test.go    # Aggregator tests
│   ├── cache_test.go         # Cache tests
│   └── ratelimit_test.go     # Rate limiter tests
└── .env.example              # Example environment configuration
```
