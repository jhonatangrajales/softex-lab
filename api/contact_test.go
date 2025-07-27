package api

import (
	"strings"
	"testing"
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
		t.Setenv("SMTP_HOST", "")
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
		t.Setenv("TO_EMAIL", "")

		config, err := newSmtpConfig()
		if err != nil {
			t.Fatalf("Se esperaba que no hubiera error, se obtuvo: %v", err)
		}
		if config.ToEmail != "contacto@softex-labs.xyz" {
			t.Errorf("TO_EMAIL incorrecto: se esperaba 'contacto@softex-labs.xyz', se obtuvo '%s'", config.ToEmail)
		}
	})
}

func TestValidateContactData(t *testing.T) {
	tests := []struct {
		name    string
		data    ContactData
		wantErr bool
		errMsg  string
	}{
		{
			name: "datos válidos",
			data: ContactData{
				Name:    "Juan Pérez",
				Email:   "juan@example.com",
				Message: "Este es un mensaje de prueba válido",
			},
			wantErr: false,
		},
		{
			name: "nombre vacío",
			data: ContactData{
				Name:    "",
				Email:   "juan@example.com",
				Message: "Mensaje válido",
			},
			wantErr: true,
			errMsg:  "el nombre es obligatorio",
		},
		{
			name: "email inválido",
			data: ContactData{
				Name:    "Juan Pérez",
				Email:   "email-invalido",
				Message: "Mensaje válido",
			},
			wantErr: true,
			errMsg:  "el formato del correo electrónico no es válido",
		},
		{
			name: "mensaje muy corto",
			data: ContactData{
				Name:    "Juan Pérez",
				Email:   "juan@example.com",
				Message: "Corto",
			},
			wantErr: true,
			errMsg:  "el mensaje debe tener al menos 10 caracteres",
		},
		{
			name: "nombre con caracteres inválidos",
			data: ContactData{
				Name:    "Juan123",
				Email:   "juan@example.com",
				Message: "Mensaje válido para prueba",
			},
			wantErr: true,
			errMsg:  "el nombre solo puede contener letras y espacios",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateContactData(&tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateContactData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("validateContactData() error = %v, expected to contain %v", err, tt.errMsg)
			}
		})
	}
}

func TestFormatEmailBody(t *testing.T) {
	data := ContactData{
		Name:    "Juan Pérez",
		Email:   "juan@example.com",
		Message: "Este es un mensaje de prueba",
	}
	clientIP := "192.168.1.1"

	body, err := formatEmailBody(data, clientIP)
	if err != nil {
		t.Fatalf("formatEmailBody() error = %v", err)
	}

	if !strings.Contains(body, data.Name) {
		t.Errorf("El cuerpo del email no contiene el nombre: %s", data.Name)
	}
	if !strings.Contains(body, data.Email) {
		t.Errorf("El cuerpo del email no contiene el email: %s", data.Email)
	}
	if !strings.Contains(body, data.Message) {
		t.Errorf("El cuerpo del email no contiene el mensaje: %s", data.Message)
	}
	if !strings.Contains(body, clientIP) {
		t.Errorf("El cuerpo del email no contiene la IP del cliente: %s", clientIP)
	}
}

func TestRateLimit(t *testing.T) {
	resetRateLimiter()

	clientIP := "192.168.1.100"

	for i := 0; i < 3; i++ {
		err := checkRateLimit(clientIP)
		if err != nil {
			t.Errorf("Solicitud %d debería haber pasado, pero obtuvo error: %v", i+1, err)
		}
	}

	err := checkRateLimit(clientIP)
	if err == nil {
		t.Error("La 4ta solicitud debería haber fallado por rate limit")
	}
}

func TestSanitizeInput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "texto normal",
			input:    "Juan Pérez",
			expected: "Juan Pérez",
		},
		{
			name:     "con espacios extra",
			input:    "  Juan Pérez  ",
			expected: "Juan Pérez",
		},
		{
			name:     "con HTML",
			input:    "<script>alert('xss')</script>Juan",
			expected: "&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;Juan",
		},
		{
			name:     "con caracteres de control",
			input:    "Juan\x00\x1fPérez",
			expected: "JuanPérez",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeInput(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizeInput() = %v, expected %v", result, tt.expected)
			}
		})
	}
}