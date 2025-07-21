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
	mu sync.RWMutex
	// sortedCache holds sorted durations for faster percentile queries
	sortedCache []int64
	// cacheValid indicates if the sorted cache is up to date
	cacheValid bool
}

// NewService creates and returns a new Service instance.
// The returned service is ready to store response times and calculate percentiles.
func NewService() *Service {
	return &Service{
		data:        make([]ResponseTimeEntry, 0, 1000), // Pre-allocate for better performance
		sortedCache: make([]int64, 0, 1000),
		cacheValid:  true,
	}
}

// StoreResponseTime stores a response time entry in memory.
// It takes a timestamp (when the request was received) and a duration
// (how long the request took to respond) and adds it to the internal storage.
//
// The function is thread-safe and can be called concurrently.
// In a real-world scenario, this could fail due to out-of-memory conditions,
// but for simplicity, it always succeeds unless the slice append fails.
//
// Performance: O(1) amortized time complexity
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

	// Invalidate cache since we added new data
	s.cacheValid = false

	return nil
}

// GetResponseTime returns the response time at the given percentile.
// The percentile must be between 0 and 100 (inclusive).
// For example, 90 returns the 90th percentile response time.
//
// The function uses the nearest-rank method for percentile calculation.
// It maintains a sorted cache for better performance on repeated queries.
//
// Performance: O(n log n) for first query, O(1) for subsequent queries with cache
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
	s.mu.RLock()

	if len(s.data) == 0 {
		s.mu.RUnlock()
		return 0, errors.New("no data available")
	}

	if percentile < 0 || percentile > 100 {
		s.mu.RUnlock()
		return 0, errors.New("percentile must be between 0 and 100")
	}

	// Use cache if valid, otherwise rebuild it
	if !s.cacheValid {
		s.mu.RUnlock()
		s.rebuildCache()
		s.mu.RLock()
	}

	// Calculate the index for the nearest-rank method
	idx := int((percentile/100.0)*float64(len(s.sortedCache))) - 1
	if idx < 0 {
		idx = 0
	}
	if idx >= len(s.sortedCache) {
		idx = len(s.sortedCache) - 1
	}

	result := time.Duration(s.sortedCache[idx])
	s.mu.RUnlock()

	return result, nil
}

// rebuildCache rebuilds the sorted cache for faster percentile queries.
// This is called when new data is added and the cache becomes invalid.
func (s *Service) rebuildCache() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Reallocate cache if needed
	if cap(s.sortedCache) < len(s.data) {
		s.sortedCache = make([]int64, len(s.data))
	} else {
		s.sortedCache = s.sortedCache[:len(s.data)]
	}

	// Copy durations to cache
	for i, entry := range s.data {
		s.sortedCache[i] = int64(entry.Duration)
	}

	// Sort the cache
	sort.Slice(s.sortedCache, func(i, j int) bool {
		return s.sortedCache[i] < s.sortedCache[j]
	})

	s.cacheValid = true
}

// GetStats returns basic statistics about the stored data.
// This is useful for monitoring and debugging.
func (s *Service) GetStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return map[string]interface{}{
		"total_entries": len(s.data),
		"cache_valid":   s.cacheValid,
		"cache_size":    len(s.sortedCache),
	}
}
