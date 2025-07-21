# In-Memory Response Time Service

A lightweight, high-performance in-memory service written in Go for storing HTTP response times and computing percentiles in real-time.

## Features

- **Store Response Times**: Capture HTTP request timestamps and response durations
- **Percentile Calculations**: Compute 50th, 90th, 99th, and custom percentiles
- **In-Memory Storage**: Fast access using Go slices (no external database required)
- **HTTP API**: RESTful endpoints for easy integration
- **Thread-Safe**: Mutex-protected operations for future concurrency support
- **Error Handling**: Comprehensive error handling for edge cases

## Project Structure

```
InMemoryService/
├── cmd/
│   └── main.go          # HTTP server and API endpoints
├── internal/
│   └── response_time/
│       └── service.go   # Core business logic
├── go.mod               # Go module definition
└── README.md           # This file
```

## API Endpoints

### Store Response Time
**POST** `/store`

Stores a response time entry with timestamp and duration.

**Request Body:**
```json
{
  "timestamp": "2024-01-15T10:30:00Z",
  "duration_ms": 150
}
```

**Response:**
```json
{
  "status": "ok"
}
```

### Get Percentile
**GET** `/percentile?percentile=90`

Returns the response time at the specified percentile.

**Response:**
```json
{
  "percentile": 90,
  "response_time_ms": 200
}
```

## Usage Examples

### Using curl

```bash
# Store a response time
curl -X POST http://localhost:8080/store \
  -H "Content-Type: application/json" \
  -d '{"timestamp": "2024-01-15T10:30:00Z", "duration_ms": 150}'

# Get 90th percentile
curl "http://localhost:8080/percentile?percentile=90"

# Get 50th percentile (median)
curl "http://localhost:8080/percentile?percentile=50"

# Get 99th percentile
curl "http://localhost:8080/percentile?percentile=99"
```

### Using Go

```go
package main

import (
    "fmt"
    "time"
    "github.com/anandpulakala/InMemoryService/internal/response_time"
)

func main() {
    service := response_time.NewService()
    
    // Store some response times
    service.StoreResponseTime(time.Now(), 100*time.Millisecond)
    service.StoreResponseTime(time.Now(), 200*time.Millisecond)
    service.StoreResponseTime(time.Now(), 150*time.Millisecond)
    
    // Get 90th percentile
    p90, err := service.GetResponseTime(90)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    fmt.Printf("90th percentile: %v\n", p90)
}
```

## Installation & Running

1. **Clone the repository:**
   ```bash
   git clone https://github.com/anandpulakala/InMemoryService.git
   cd InMemoryService
   ```

2. **Run the service:**
   ```bash
   go run cmd/main.go
   ```

3. **The service will start on port 8080:**
   ```
   Listening on :8080...
   ```

## Technical Details

### Data Storage
- Uses Go slices for in-memory storage
- Each entry contains timestamp and duration
- Data is kept unsorted for efficient writes
- Sorting is performed on-demand for percentile calculations

### Percentile Calculation
- Uses the nearest-rank method for percentile calculation
- Sorts data in ascending order before computing percentiles
- Handles edge cases (empty data, invalid percentiles)

### Performance Considerations
- **Write-heavy scenarios**: Efficient O(1) append operations
- **Read-heavy scenarios**: O(n log n) sorting on each percentile request
- **Memory usage**: Linear growth with number of stored entries

### Trade-offs Discussed
1. **Sorted vs Unsorted Storage**:
   - **Unsorted (current)**: Fast writes, slower reads
   - **Sorted**: Slower writes, faster reads
   - Chosen unsorted for simplicity and write efficiency

2. **Concurrency**:
   - Mutex included for future multi-threaded access
   - Currently single-threaded as per requirements

3. **Error Handling**:
   - Invalid percentiles (0-100 range)
   - Empty data sets
   - Malformed requests

## Error Handling

The service handles various error conditions:

- **Invalid percentiles**: Must be between 0 and 100
- **Empty data**: Returns error when no data is available
- **Malformed requests**: Invalid JSON or missing required fields
- **Invalid timestamps**: Must be in RFC3339 format

## Scaling Considerations

### Current Limitations
- In-memory storage (data lost on restart)
- Single instance (no clustering)
- No persistence layer

### Future Enhancements
- **Persistence**: Add database backend (Redis, PostgreSQL)
- **Clustering**: Multiple service instances with load balancing
- **Caching**: Keep sorted data for faster percentile queries
- **Compression**: For large datasets
- **Monitoring**: Metrics and health checks

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

This project is open source and available under the [MIT License](LICENSE).

## Author

Anand Pulakala - [GitHub](https://github.com/anandpulakala) 