package handler

import (
	"encoding/json"
	"net/http"
	"time"
)

// HealthResponse representa la respuesta del health check
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
	Service   string    `json:"service"`
}

// Health - Handler para health check (nombre de función exportada)
func Health(w http.ResponseWriter, r *http.Request) {
	// CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	health := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().UTC(),
		Version:   "2.0.0",
		Service:   "softex-labs-contact-api",
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache")

	if err := json.NewEncoder(w).Encode(health); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
