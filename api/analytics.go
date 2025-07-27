package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

// Analytics representa datos de analytics del formulario
type Analytics struct {
	TotalSubmissions int64                    `json:"total_submissions"`
	SuccessRate      float64                  `json:"success_rate"`
	TopCountries     map[string]int           `json:"top_countries"`
	HourlyStats      map[string]int           `json:"hourly_stats"`
	LastUpdated      time.Time                `json:"last_updated"`
	ErrorStats       map[string]int           `json:"error_stats"`
}

// FormSubmission representa una submisión del formulario para analytics
type FormSubmission struct {
	Timestamp time.Time `json:"timestamp"`
	Success   bool      `json:"success"`
	Country   string    `json:"country"`
	Error     string    `json:"error,omitempty"`
}

var (
	submissions []FormSubmission
	analytics   Analytics
)

// Función para registrar una submisión
func recordSubmission(success bool, country, errorMsg string) {
	submission := FormSubmission{
		Timestamp: time.Now(),
		Success:   success,
		Country:   country,
		Error:     errorMsg,
	}
	
	submissions = append(submissions, submission)
	updateAnalytics()
}

// Actualizar estadísticas de analytics
func updateAnalytics() {
	if len(submissions) == 0 {
		return
	}

	analytics.TotalSubmissions = int64(len(submissions))
	analytics.LastUpdated = time.Now()
	
	// Calcular tasa de éxito
	successCount := 0
	countryCount := make(map[string]int)
	hourlyCount := make(map[string]int)
	errorCount := make(map[string]int)
	
	for _, sub := range submissions {
		if sub.Success {
			successCount++
		} else if sub.Error != "" {
			errorCount[sub.Error]++
		}
		
		countryCount[sub.Country]++
		hour := sub.Timestamp.Format("15")
		hourlyCount[hour]++
	}
	
	analytics.SuccessRate = float64(successCount) / float64(len(submissions)) * 100
	analytics.TopCountries = countryCount
	analytics.HourlyStats = hourlyCount
	analytics.ErrorStats = errorCount
}

// Handler para obtener analytics (solo para admin)
func AnalyticsHandler(w http.ResponseWriter, r *http.Request) {
	// Verificar autenticación básica (en producción usar JWT)
	adminKey := os.Getenv("ADMIN_KEY")
	if adminKey == "" || r.Header.Get("X-Admin-Key") != adminKey {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	
	if err := json.NewEncoder(w).Encode(analytics); err != nil {
		log.Printf("Error encoding analytics: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// Función para obtener país desde IP (simulado)
func getCountryFromIP(ip string) string {
	// En producción, usar un servicio como MaxMind GeoIP
	if ip == "127.0.0.1" || ip == "::1" {
		return "Local"
	}
	return "Unknown"
}

// Modificar el Handler principal para incluir analytics
func HandlerWithAnalytics(w http.ResponseWriter, r *http.Request) {
	clientIP := getClientIP(r)
	country := getCountryFromIP(clientIP)
	
	// Llamar al handler original
	originalHandler := http.HandlerFunc(Handler)
	
	// Crear un ResponseWriter personalizado para capturar el status
	rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
	originalHandler.ServeHTTP(rw, r)
	
	// Registrar la submisión para analytics
	success := rw.statusCode == http.StatusOK
	errorMsg := ""
	if !success {
		errorMsg = fmt.Sprintf("HTTP %d", rw.statusCode)
	}
	
	recordSubmission(success, country, errorMsg)
}

// ResponseWriter personalizado para capturar status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}