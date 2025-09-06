package handler

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"html"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode"
)

// ContactData representa los datos del formulario de contacto
type ContactData struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Message string `json:"message"`
}

// SmtpConfig contiene la configuración SMTP
type SmtpConfig struct {
	Host    string
	Port    string
	User    string
	Pass    string
	ToEmail string
}

// APIResponse representa la respuesta de la API
type APIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// Rate limiting
type requestInfo struct {
	count     int
	firstSeen time.Time
}

var (
	visitors = make(map[string]*requestInfo)
	mu       sync.Mutex
)

const (
	maxRequests = 3
	timeWindow  = 5 * time.Minute
)

// Función para obtener la configuración SMTP desde variables de entorno
func newSmtpConfig() (SmtpConfig, error) {
	config := SmtpConfig{
		Host:    os.Getenv("SMTP_HOST"),
		Port:    os.Getenv("SMTP_PORT"),
		User:    os.Getenv("SMTP_USER"),
		Pass:    os.Getenv("SMTP_PASS"),
		ToEmail: os.Getenv("TO_EMAIL"),
	}

	if config.Host == "" || config.Port == "" || config.User == "" || config.Pass == "" {
		return config, fmt.Errorf("faltan variables de entorno SMTP requeridas")
	}

	if config.ToEmail == "" {
		config.ToEmail = "info@softexlab.com"
	}

	return config, nil
}

// Función para validar los datos del formulario
func validateContactData(data *ContactData) error {
	// Sanitizar datos
	data.Name = sanitizeInput(data.Name)
	data.Email = sanitizeInput(data.Email)
	data.Message = sanitizeInput(data.Message)

	// Validar nombre
	if strings.TrimSpace(data.Name) == "" {
		return fmt.Errorf("el nombre es obligatorio")
	}
	if len(data.Name) > 100 {
		return fmt.Errorf("el nombre no puede exceder 100 caracteres")
	}
	// Validar que el nombre solo contenga letras y espacios
	nameRegex := regexp.MustCompile(`^[a-zA-ZáéíóúÁÉÍÓÚñÑ\s]+$`)
	if !nameRegex.MatchString(data.Name) {
		return fmt.Errorf("el nombre solo puede contener letras y espacios")
	}

	// Validar email
	if strings.TrimSpace(data.Email) == "" {
		return fmt.Errorf("el correo electrónico es obligatorio")
	}
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(data.Email) {
		return fmt.Errorf("el formato del correo electrónico no es válido")
	}

	// Validar mensaje
	if strings.TrimSpace(data.Message) == "" {
		return fmt.Errorf("el mensaje es obligatorio")
	}
	if len(data.Message) < 10 {
		return fmt.Errorf("el mensaje debe tener al menos 10 caracteres")
	}
	if len(data.Message) > 1000 {
		return fmt.Errorf("el mensaje no puede exceder 1000 caracteres")
	}

	return nil
}

// Función para sanitizar entrada
func sanitizeInput(input string) string {
	// Remover caracteres de control
	cleaned := strings.Map(func(r rune) rune {
		if unicode.IsControl(r) && r != '\n' && r != '\r' && r != '\t' {
			return -1
		}
		return r
	}, input)

	// Escapar HTML
	cleaned = html.EscapeString(cleaned)

	// Trim espacios
	return strings.TrimSpace(cleaned)
}

// Función para verificar rate limiting
func checkRateLimit(clientIP string) error {
	mu.Lock()
	defer mu.Unlock()

	now := time.Now()

	// Limpiar entradas antiguas
	for ip, info := range visitors {
		if now.Sub(info.firstSeen) > timeWindow {
			delete(visitors, ip)
		}
	}

	// Verificar límite para esta IP
	if info, exists := visitors[clientIP]; exists {
		if info.count >= maxRequests {
			return fmt.Errorf("demasiadas solicitudes. Intenta de nuevo en %v", timeWindow-now.Sub(info.firstSeen))
		}
		info.count++
	} else {
		visitors[clientIP] = &requestInfo{
			count:     1,
			firstSeen: now,
		}
	}

	return nil
}

// Función para obtener la IP del cliente
func getClientIP(r *http.Request) string {
	// Verificar headers de proxy
	if ip := r.Header.Get("CF-Connecting-IP"); ip != "" {
		return ip
	}
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return strings.Split(ip, ",")[0]
	}
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}

	// Fallback a RemoteAddr
	ip := r.RemoteAddr
	if colon := strings.LastIndex(ip, ":"); colon != -1 {
		ip = ip[:colon]
	}
	return ip
}

// Función para formatear el cuerpo del email
func formatEmailBody(data ContactData, clientIP string) (string, error) {
	template := `
<!DOCTYPE html>
<html lang="es">
<head>
    <meta charset="UTF-8">
    <title>Nuevo mensaje de contacto - Softex Labs</title>
</head>
<body style="font-family: Arial, sans-serif; background-color: #f4f4f4; padding: 20px;">
    <div style="max-width: 600px; margin: 0 auto; background: white; padding: 30px; border-radius: 8px; box-shadow: 0 0 10px rgba(0,0,0,0.1);">
        <h2 style="color: #4f46e5; border-bottom: 2px solid #4f46e5; padding-bottom: 10px;">
            📧 Nuevo Mensaje de Contacto
        </h2>
        
        <div style="margin: 20px 0;">
            <h3 style="color: #333; margin-bottom: 15px;">Información del Contacto:</h3>
            <table style="width: 100%%; border-collapse: collapse;">
                <tr>
                    <td style="padding: 8px; background: #f9f9f9; border: 1px solid #ddd; font-weight: bold; width: 30%%;">Nombre:</td>
                    <td style="padding: 8px; border: 1px solid #ddd;">%s</td>
                </tr>
                <tr>
                    <td style="padding: 8px; background: #f9f9f9; border: 1px solid #ddd; font-weight: bold;">Email:</td>
                    <td style="padding: 8px; border: 1px solid #ddd;">%s</td>
                </tr>
                <tr>
                    <td style="padding: 8px; background: #f9f9f9; border: 1px solid #ddd; font-weight: bold;">IP:</td>
                    <td style="padding: 8px; border: 1px solid #ddd;">%s</td>
                </tr>
                <tr>
                    <td style="padding: 8px; background: #f9f9f9; border: 1px solid #ddd; font-weight: bold;">Fecha:</td>
                    <td style="padding: 8px; border: 1px solid #ddd;">%s</td>
                </tr>
            </table>
        </div>
        
        <div style="margin: 20px 0;">
            <h3 style="color: #333; margin-bottom: 15px;">Mensaje:</h3>
            <div style="background: #f9f9f9; padding: 15px; border-left: 4px solid #4f46e5; border-radius: 4px;">
                %s
            </div>
        </div>
        
        <div style="margin-top: 30px; padding-top: 20px; border-top: 1px solid #ddd; color: #666; font-size: 12px;">
            <p>Este mensaje fue enviado desde el formulario de contacto de Softex Labs.</p>
            <p>Responde directamente a este email para contactar al usuario.</p>
        </div>
    </div>
</body>
</html>`

	return fmt.Sprintf(template,
		data.Name,
		data.Email,
		clientIP,
		time.Now().Format("2006-01-02 15:04:05"),
		strings.ReplaceAll(data.Message, "\n", "<br>")), nil
}

// Función para enviar email
func sendEmail(config SmtpConfig, data ContactData, clientIP string) error {
	subject := fmt.Sprintf("Nuevo mensaje de contacto de %s", data.Name)

	body, err := formatEmailBody(data, clientIP)
	if err != nil {
		return fmt.Errorf("error al formatear el cuerpo del email: %v", err)
	}

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
		config.User, config.ToEmail, subject, body)

	auth := smtp.PlainAuth("", config.User, config.Pass, config.Host)
	addr := fmt.Sprintf("%s:%s", config.Host, config.Port)

	// Configurar TLS
	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         config.Host,
	}

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("error al conectar con TLS: %v", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, config.Host)
	if err != nil {
		return fmt.Errorf("error al crear cliente SMTP: %v", err)
	}
	defer func() {
		if err := client.Quit(); err != nil {
			log.Printf("Error al cerrar cliente SMTP: %v", err)
		}
	}()

	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("error de autenticación SMTP: %v", err)
	}

	if err = client.Mail(config.User); err != nil {
		return fmt.Errorf("error al establecer remitente: %v", err)
	}

	if err = client.Rcpt(config.ToEmail); err != nil {
		return fmt.Errorf("error al establecer destinatario: %v", err)
	}

	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("error al iniciar datos: %v", err)
	}

	_, err = writer.Write([]byte(msg))
	if err != nil {
		return fmt.Errorf("error al escribir mensaje: %v", err)
	}

	err = writer.Close()
	if err != nil {
		return fmt.Errorf("error al cerrar escritor: %v", err)
	}

	return nil
}

// Función para parsear y validar la request
func parseAndValidateRequest(r *http.Request) (ContactData, error) {
	var data ContactData

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		return data, fmt.Errorf("error al decodificar JSON: %v", err)
	}

	if err := validateContactData(&data); err != nil {
		return data, err
	}

	return data, nil
}

// Función para enviar respuesta JSON de error
func sendJSONError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(APIResponse{
		Success: false,
		Message: message,
	})
}

// Función para enviar respuesta JSON de éxito
func sendJSONSuccess(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(APIResponse{
		Success: true,
		Message: message,
	})
}

func Contact(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	clientIP := getClientIP(r)

	log.Printf("Solicitud recibida - Método: %s, IP: %s", r.Method, clientIP)

	// Configuración de CORS más flexible para manejar www y sin www
	// Permitir todos los orígenes temporalmente para resolver el problema
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Max-Age", "86400")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Rate limiting
	if err := checkRateLimit(clientIP); err != nil {
		log.Printf("Rate limit excedido para IP %s: %v", clientIP, err)
		sendJSONError(w, err.Error(), http.StatusTooManyRequests)
		return
	}

	if r.Method != http.MethodPost {
		sendJSONError(w, "Método no permitido. Solo se acepta POST.", http.StatusMethodNotAllowed)
		return
	}

	// Parsear y validar
	data, err := parseAndValidateRequest(r)
	if err != nil {
		log.Printf("Error de validación para IP %s: %v", clientIP, err)
		sendJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Configuración SMTP
	config, err := newSmtpConfig()
	if err != nil {
		log.Printf("Error de configuración SMTP: %v", err)
		sendJSONError(w, "Error de configuración del servidor. Por favor, contacta al administrador.", http.StatusInternalServerError)
		return
	}

	// Enviar email
	err = sendEmail(config, data, clientIP)
	if err != nil {
		log.Printf("Error al enviar correo para IP %s: %v", clientIP, err)
		// Mensaje más amigable para el usuario
		if strings.Contains(err.Error(), "TLS handshake") || strings.Contains(err.Error(), "connection") {
			   sendJSONError(w, "Error de conexión con el servidor de correo. Por favor, intenta de nuevo en unos minutos o contacta directamente a info@softexlab.com", http.StatusServiceUnavailable)
		} else if strings.Contains(err.Error(), "authentication") {
			sendJSONError(w, "Error de configuración del servidor. Por favor, contacta al administrador.", http.StatusInternalServerError)
		} else {
			   sendJSONError(w, "Error al enviar el mensaje. Por favor, intenta de nuevo o contacta directamente a info@softexlab.com\nDetalle del error: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	duration := time.Since(startTime)
	log.Printf("Correo enviado exitosamente - IP: %s, Duración: %v", clientIP, duration)

	sendJSONSuccess(w, "¡Mensaje enviado con éxito! Te responderemos pronto.")
}
