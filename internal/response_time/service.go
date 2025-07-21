// Package response_time provides in-memory storage and percentile calculation for HTTP response times.
// Intended for import as: github.com/anandpulakala/InMemoryService/internal/response_time
package response_time

import (
	"errors"
	"sort"
	"sync"
	"time"
)

// ResponseTimeEntry represents a single HTTP response time measurement.
// It contains the timestamp when the request was received and the duration
// it took to complete the response.
type ResponseTimeEntry struct {
	// Timestamp is when the HTTP request was received
	Timestamp time.Time
	// Duration is how long the request took to respond
	Duration time.Duration
}

// Service provides in-memory storage and percentile calculation for response times.
// It is thread-safe and can handle concurrent access through mutex protection.
// The service stores response time entries in memory and provides methods
// to calculate percentiles of the stored data.
type Service struct {
	// data holds all stored response time entries
	data []ResponseTimeEntry
	// mu protects concurrent access to the data slice
	mu sync.Mutex
}

// NewService creates and returns a new Service instance.
// The returned service is ready to store response times and calculate percentiles.
func NewService() *Service {
	return &Service{data: make([]ResponseTimeEntry, 0)}
}

// StoreResponseTime stores a response time entry in memory.
// It takes a timestamp (when the request was received) and a duration
// (how long the request took to respond) and adds it to the internal storage.
//
// The function is thread-safe and can be called concurrently.
// In a real-world scenario, this could fail due to out-of-memory conditions,
// but for simplicity, it always succeeds unless the slice append fails.
//
// Example:
//
//	service := NewService()
//	err := service.StoreResponseTime(time.Now(), 150*time.Millisecond)
//	if err != nil {
//	    // handle error
//	}
func (s *Service) StoreResponseTime(ts time.Time, duration time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry := ResponseTimeEntry{Timestamp: ts, Duration: duration}
	s.data = append(s.data, entry)
	return nil
}

// GetResponseTime returns the response time at the given percentile.
// The percentile must be between 0 and 100 (inclusive).
// For example, 90 returns the 90th percentile response time.
//
// The function uses the nearest-rank method for percentile calculation.
// It sorts the stored durations in ascending order and returns the value
// at the appropriate position based on the percentile.
//
// Returns an error if:
//   - No data has been stored (empty dataset)
//   - The percentile is outside the valid range (0-100)
//
// Example:
//
//	p90, err := service.GetResponseTime(90)
//	if err != nil {
//	    // handle error
//	}
//	fmt.Printf("90th percentile: %v\n", p90)
func (s *Service) GetResponseTime(percentile float64) (time.Duration, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.data) == 0 {
		return 0, errors.New("no data available")
	}

	if percentile < 0 || percentile > 100 {
		return 0, errors.New("percentile must be between 0 and 100")
	}

	// Copy durations for sorting to avoid modifying the original data
	durations := make([]int64, len(s.data))
	for i, entry := range s.data {
		durations[i] = int64(entry.Duration)
	}

	// Sort durations in ascending order
	sort.Slice(durations, func(i, j int) bool { return durations[i] < durations[j] })

	// Calculate the index for the nearest-rank method
	idx := int((percentile/100.0)*float64(len(durations))) - 1
	if idx < 0 {
		idx = 0
	}
	if idx >= len(durations) {
		idx = len(durations) - 1
	}

	return time.Duration(durations[idx]), nil
}
