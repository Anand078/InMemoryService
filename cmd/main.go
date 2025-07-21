package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/anandpulakala/InMemoryService/internal/response_time"
)

var svc = response_time.NewService()

type storeRequest struct {
	Timestamp  string `json:"timestamp"` // RFC3339
	DurationMs int64  `json:"duration_ms"`
}

type storeResponse struct {
	Status string `json:"status"`
}

type percentileResponse struct {
	Percentile     float64 `json:"percentile"`
	ResponseTimeMs int64   `json:"response_time_ms"`
}

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

func main() {
	http.HandleFunc("/store", storeHandler)
	http.HandleFunc("/percentile", percentileHandler)
	log.Println("Listening on :8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
