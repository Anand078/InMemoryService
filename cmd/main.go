// Package main provides an HTTP server for the in-memory response time service.
// It exposes RESTful endpoints to store response times and retrieve percentiles.
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/anandpulakala/InMemoryService/internal/response_time"
)

// Global service instance for handling response time operations
var svc = response_time.NewService()

// storeRequest represents the JSON payload for storing a response time.
// The timestamp should be in RFC3339 format, and duration_ms should be
// the response time in milliseconds.
type storeRequest struct {
	// Timestamp is when the HTTP request was received (RFC3339 format)
	Timestamp string `json:"timestamp"`
	// DurationMs is the response time in milliseconds
	DurationMs int64 `json:"duration_ms"`
}

// storeResponse represents the JSON response for a successful store operation.
type storeResponse struct {
	// Status indicates the result of the store operation
	Status string `json:"status"`
}

// percentileResponse represents the JSON response for a percentile query.
type percentileResponse struct {
	// Percentile is the requested percentile value
	Percentile float64 `json:"percentile"`
	// ResponseTimeMs is the response time at the given percentile in milliseconds
	ResponseTimeMs int64 `json:"response_time_ms"`
}

// storeHandler handles POST requests to store response time entries.
// It expects a JSON payload with timestamp and duration_ms fields.
// The timestamp must be in RFC3339 format (e.g., "2024-01-15T10:30:00Z").
// The duration_ms should be the response time in milliseconds.
//
// Example request:
//
//	POST /store
//	Content-Type: application/json
//	{
//	  "timestamp": "2024-01-15T10:30:00Z",
//	  "duration_ms": 150
//	}
//
// Example response:
//
//	{
//	  "status": "ok"
//	}
func storeHandler(w http.ResponseWriter, r *http.Request) {
	var req storeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	ts, err := time.Parse(time.RFC3339, req.Timestamp)
	if err != nil {
		http.Error(w, "invalid timestamp", http.StatusBadRequest)
		return
	}

	dur := time.Duration(req.DurationMs) * time.Millisecond
	if err := svc.StoreResponseTime(ts, dur); err != nil {
		http.Error(w, "failed to store", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(storeResponse{Status: "ok"})
}

// percentileHandler handles GET requests to retrieve response time percentiles.
// It expects a query parameter 'percentile' with a value between 0 and 100.
// The function returns the response time at the specified percentile.
//
// Example request:
//
//	GET /percentile?percentile=90
//
// Example response:
//
//	{
//	  "percentile": 90,
//	  "response_time_ms": 200
//	}
//
// Common percentiles:
//   - 50: median response time
//   - 90: 90th percentile (90% of requests are faster)
//   - 99: 99th percentile (99% of requests are faster)
func percentileHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("percentile")
	if q == "" {
		http.Error(w, "missing percentile", http.StatusBadRequest)
		return
	}

	p, err := strconv.ParseFloat(q, 64)
	if err != nil {
		http.Error(w, "invalid percentile", http.StatusBadRequest)
		return
	}

	dur, err := svc.GetResponseTime(p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(percentileResponse{
		Percentile:     p,
		ResponseTimeMs: dur.Milliseconds(),
	})
}

// main starts the HTTP server and sets up the API endpoints.
// The server listens on port 8080 and provides two endpoints:
//   - POST /store: Store a response time entry
//   - GET /percentile: Retrieve response time at a given percentile
//
// The server uses the global service instance to handle all operations.
// All data is stored in memory and will be lost when the server restarts.
func main() {
	http.HandleFunc("/store", storeHandler)
	http.HandleFunc("/percentile", percentileHandler)
	log.Println("Listening on :8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
