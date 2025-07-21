// Package response_time provides in-memory storage and percentile calculation for HTTP response times.
// Intended for import as: github.com/anandpulakala/InMemoryService/internal/response_time
package response_time

import (
	"errors"
	"sort"
	"sync"
	"time"
)

// ResponseTimeEntry holds a timestamp and duration for a request.
type ResponseTimeEntry struct {
	Timestamp time.Time
	Duration  time.Duration
}

// Service provides in-memory storage and percentile calculation for response times.
type Service struct {
	data []ResponseTimeEntry
	mu   sync.Mutex // For future concurrency, not strictly needed for single-threaded use
}

// NewService creates a new Service instance.
func NewService() *Service {
	return &Service{data: make([]ResponseTimeEntry, 0)}
}

// StoreResponseTime stores a response time entry in memory.
func (s *Service) StoreResponseTime(ts time.Time, duration time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	// In real-world, could fail due to OOM, etc. Here, always succeeds unless slice append fails.
	entry := ResponseTimeEntry{Timestamp: ts, Duration: duration}
	s.data = append(s.data, entry)
	return nil
}

// GetResponseTime returns the response time at the given percentile (e.g., 90 for 90th percentile).
func (s *Service) GetResponseTime(percentile float64) (time.Duration, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.data) == 0 {
		return 0, errors.New("no data available")
	}
	// Copy durations for sorting
	durations := make([]int64, len(s.data))
	for i, entry := range s.data {
		durations[i] = int64(entry.Duration)
	}
	sort.Slice(durations, func(i, j int) bool { return durations[i] < durations[j] })
	if percentile < 0 || percentile > 100 {
		return 0, errors.New("percentile must be between 0 and 100")
	}
	// Nearest-rank method
	idx := int((percentile/100.0)*float64(len(durations))) - 1
	if idx < 0 {
		idx = 0
	}
	if idx >= len(durations) {
		idx = len(durations) - 1
	}
	return time.Duration(durations[idx]), nil
}
