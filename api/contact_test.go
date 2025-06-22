package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// Helper para resetear el estado del rate limiter entre tests.
func resetRateLimiter() {
	mu.Lock()
	defer mu.Unlock()
	visitors = make(map[string]*requestInfo)
}

func TestNewSmtpConfig(t *testing.T) {
	t.Run("éxito con todas las variables de entorno", func(t *testing.T) {
		t.Setenv("SMTP_HOST", "smtp.example.com")
		t.Setenv("SMTP_PORT", "587")
		t.Setenv("SMTP_USER", "user")
		t.Setenv("SMTP_PASS", "pass")
		t.Setenv("TO_EMAIL", "to@example.com")

		config, err := newSmtpConfig()
		if err != nil {
			t.Fatalf("Se esperaba que no hubiera error, se obtuvo: %v", err)
		}
		if config.Host != "smtp.example.com" {
			t.Errorf("Host incorrecto: se esperaba 'smtp.example.com', se obtuvo '%s'", config.Host)
		}
	})

	t.Run("falla por variables de entorno faltantes", func(t *testing.T) {
		// No se establecen variables de entorno
		t.Setenv("SMTP_HOST", "") // Asegurarse de que esté vacío
		_, err := newSmtpConfig()
		if err == nil {
			t.Fatal("Se esperaba un error, pero no se obtuvo ninguno")
		}
	})

	t.Run("fallback para TO_EMAIL", func(t *testing.T) {
		t.Setenv("SMTP_HOST", "smtp.example.com")
		t.Setenv("SMTP_PORT", "587")
		t.Setenv("SMTP_USER", "user")
		t.Setenv("SMTP_PASS", "pass")
		t.Setenv("TO_EMAIL", "") // TO_EMAIL está vacío

		config, err := newSmtpConfig()
		if err != nil {
			t.Fatalf("No se esperaba un error, se obtuvo: %v", err)
		}
		if config.ToEmail != "grajajhon9@gmail.com" {
			t.Errorf("Fallback de TO_EMAIL incorrecto: se esperaba 'grajajhon9@gmail.com', se obtuvo '%s'", config.ToEmail)
		}
	})
}

func TestParseAndValidateRequest(t *testing.T) {
	t.Run("solicitud válida", func(t *testing.T) {
		validData := ContactData{Name: "John Doe", Email: "john@example.com", Message: "Este es un mensaje de prueba."}
		body, _ := json.Marshal(validData)
		req := httptest.NewRequest("POST", "/api/contact", bytes.NewReader(body))

		data, err := parseAndValidateRequest(req)
		if err != nil {
			t.Fatalf("Se esperaba que no hubiera error, se obtuvo: %v", err)
		}
		if data.Name != validData.Name {
			t.Errorf("Nombre incorrecto: se esperaba '%s', se obtuvo '%s'", validData.Name, data.Name)
		}
	})

	t.Run("campos faltantes", func(t *testing.T) {
		invalidData := ContactData{Name: "John Doe", Email: "john@example.com"} // Falta el mensaje
		body, _ := json.Marshal(invalidData)
		req := httptest.NewRequest("POST", "/api/contact", bytes.NewReader(body))

		_, err := parseAndValidateRequest(req)
		if err == nil {
			t.Fatal("Se esperaba un error por campos faltantes, pero no se obtuvo ninguno")
		}
	})

	t.Run("email inválido", func(t *testing.T) {
		invalidData := ContactData{Name: "John Doe", Email: "invalid-email", Message: "Mensaje válido."}
		body, _ := json.Marshal(invalidData)
		req := httptest.NewRequest("POST", "/api/contact", bytes.NewReader(body))

		_, err := parseAndValidateRequest(req)
		if err == nil {
			t.Fatal("Se esperaba un error por email inválido, pero no se obtuvo ninguno")
		}
	})
}

func TestFormatEmailBody(t *testing.T) {
	data := ContactData{
		Name:    "Jane Doe",
		Email:   "jane@example.com",
		Message: "Hola, este es un mensaje de prueba.",
	}
	body, err := formatEmailBody(data)
	if err != nil {
		t.Fatalf("Error al formatear el cuerpo del email: %v", err)
	}

	if !strings.Contains(body, "Jane Doe") {
		t.Error("El cuerpo del email no contiene el nombre del remitente")
	}
	if !strings.Contains(body, "jane@example.com") {
		t.Error("El cuerpo del email no contiene el email del remitente")
	}
	if !strings.Contains(body, "Hola, este es un mensaje de prueba.") {
		t.Error("El cuerpo del email no contiene el mensaje")
	}
}

func TestRateLimiter(t *testing.T) {
	resetRateLimiter()
	t.Cleanup(resetRateLimiter) // Asegura que el mapa se limpie después del test

	// Simular el handler para probar el rate limiter
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Copiamos la lógica del rate limiter del Handler original
		clientIP := getClientIP(r)
		mu.Lock()
		if v, exists := visitors[clientIP]; exists {
			if time.Since(v.lastSeen) > timeWindow {
				v.count = 1
				v.lastSeen = time.Now()
			} else {
				v.count++
			}
			if v.count > rateLimit {
				mu.Unlock()
				sendJSONError(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}
		} else {
			visitors[clientIP] = &requestInfo{count: 1, lastSeen: time.Now()}
		}
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	})

	// Hacemos `rateLimit` solicitudes, todas deberían pasar
	for i := 0; i < rateLimit; i++ {
		req := httptest.NewRequest("POST", "/api/contact", nil)
		req.Header.Set("X-Forwarded-For", "192.0.2.1")
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("Solicitud %d: se esperaba el código 200, se obtuvo %d", i+1, rr.Code)
		}
	}

	// La siguiente solicitud (`rateLimit` + 1) debería fallar
	req := httptest.NewRequest("POST", "/api/contact", nil)
	req.Header.Set("X-Forwarded-For", "192.0.2.1")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTooManyRequests {
		t.Fatalf("Se esperaba el código 429, se obtuvo %d", rr.Code)
	}
}

func TestGetClientIP(t *testing.T) {
	t.Run("con X-Forwarded-For", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("X-Forwarded-For", "192.0.2.1, 198.51.100.10")
		ip := getClientIP(req)
		if ip != "192.0.2.1" {
			t.Errorf("Se esperaba '192.0.2.1', se obtuvo '%s'", ip)
		}
	})

	t.Run("sin X-Forwarded-For", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "127.0.0.1:12345"
		ip := getClientIP(req)
		if ip != "127.0.0.1:12345" {
			t.Errorf("Se esperaba '127.0.0.1:12345', se obtuvo '%s'", ip)
		}
	})
}
