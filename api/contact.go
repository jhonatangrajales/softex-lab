package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"html/template"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

// ContactData representa los datos del formulario con validaciones mejoradas.
type ContactData struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Message string `json:"message"`
}

// SmtpConfig contiene la configuración para el servidor SMTP.
type SmtpConfig struct {
	Host    string
	Port    string
	User    string
	Pass    string
	ToEmail string
}

// Constantes de validación
const (
	MaxNameLength    = 100
	MaxEmailLength   = 254
	MaxMessageLength = 2000
	MinMessageLength = 10
)

// emailRegex es una expresión regular más robusta para validar el formato del correo electrónico.
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// --- Rate Limiter Mejorado ---

type requestInfo struct {
	count    int
	lastSeen time.Time
	blocked  bool
}

var (
	visitors = make(map[string]*requestInfo)
	mu       sync.RWMutex
)

const (
	rateLimit     = 3                // Límite de solicitudes por ventana de tiempo (más restrictivo)
	timeWindow    = 5 * time.Minute  // Ventana de tiempo más larga
	blockDuration = 15 * time.Minute // Tiempo de bloqueo por exceder límite
)

// init se ejecuta una vez cuando el paquete es cargado.
func init() {
	// Inicia una goroutine para limpiar visitantes antiguos y prevenir fugas de memoria.
	go cleanupVisitors()
}

func cleanupVisitors() {
	for {
		time.Sleep(10 * time.Minute) // Frecuencia de limpieza
		mu.Lock()
		for ip, v := range visitors {
			if time.Since(v.lastSeen) > timeWindow*2 {
				delete(visitors, ip)
			}
		}
		mu.Unlock()
	}
}

// checkRateLimit verifica si una IP ha excedido el límite de solicitudes
func checkRateLimit(clientIP string) error {
	mu.Lock()
	defer mu.Unlock()

	now := time.Now()
	
	if v, exists := visitors[clientIP]; exists {
		// Si está bloqueado, verificar si el bloqueo ha expirado
		if v.blocked && time.Since(v.lastSeen) < blockDuration {
			return errors.New("IP bloqueada temporalmente por exceder el límite de solicitudes")
		}
		
		// Si ha pasado la ventana de tiempo, resetear contador
		if time.Since(v.lastSeen) > timeWindow {
			v.count = 1
			v.blocked = false
		} else {
			v.count++
		}
		
		v.lastSeen = now
		
		// Si excede el límite, bloquear
		if v.count > rateLimit {
			v.blocked = true
			return errors.New("has excedido el límite de solicitudes. Intenta de nuevo en 15 minutos")
		}
	} else {
		visitors[clientIP] = &requestInfo{count: 1, lastSeen: now, blocked: false}
	}
	
	return nil
}

// --- Fin del Rate Limiter ---

func sendJSONError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	
	response := map[string]interface{}{
		"error":     message,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"status":    status,
	}
	
	json.NewEncoder(w).Encode(response)
}

func sendJSONSuccess(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	response := map[string]interface{}{
		"message":   message,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"status":    http.StatusOK,
	}
	
	json.NewEncoder(w).Encode(response)
}

// newSmtpConfig crea una configuración SMTP a partir de variables de entorno.
func newSmtpConfig() (SmtpConfig, error) {
	config := SmtpConfig{
		Host:    os.Getenv("SMTP_HOST"),
		Port:    os.Getenv("SMTP_PORT"),
		User:    os.Getenv("SMTP_USER"),
		Pass:    os.Getenv("SMTP_PASS"),
		ToEmail: os.Getenv("TO_EMAIL"),
	}

	if config.Host == "" || config.Port == "" || config.User == "" || config.Pass == "" {
		log.Println("Error: Configuración SMTP incompleta en las variables de entorno de Vercel.")
		return SmtpConfig{}, errors.New("error de configuración del servidor para enviar el correo. Contacte al administrador")
	}

	if config.ToEmail == "" {
		config.ToEmail = "contacto@softex-labs.xyz" // Fallback si no se configura TO_EMAIL
		log.Println("Advertencia: TO_EMAIL no configurado. Usando el valor por defecto:", config.ToEmail)
	}

	return config, nil
}

// sanitizeInput limpia y sanitiza la entrada del usuario
func sanitizeInput(input string) string {
	// Escapar HTML para prevenir XSS
	sanitized := html.EscapeString(input)
	// Remover espacios en blanco excesivos
	sanitized = strings.TrimSpace(sanitized)
	// Remover caracteres de control
	sanitized = regexp.MustCompile(`[\x00-\x1f\x7f]`).ReplaceAllString(sanitized, "")
	return sanitized
}

// validateContactData valida los datos del formulario con reglas más estrictas
func validateContactData(data *ContactData) error {
	// Sanitizar datos
	data.Name = sanitizeInput(data.Name)
	data.Email = sanitizeInput(data.Email)
	data.Message = sanitizeInput(data.Message)

	// Validar nombre
	if data.Name == "" {
		return errors.New("el nombre es obligatorio")
	}
	if len(data.Name) > MaxNameLength {
		return fmt.Errorf("el nombre no puede exceder %d caracteres", MaxNameLength)
	}
	if matched, _ := regexp.MatchString(`^[a-zA-ZáéíóúÁÉÍÓÚñÑ\s]+$`, data.Name); !matched {
		return errors.New("el nombre solo puede contener letras y espacios")
	}

	// Validar email
	if data.Email == "" {
		return errors.New("el correo electrónico es obligatorio")
	}
	if len(data.Email) > MaxEmailLength {
		return fmt.Errorf("el correo electrónico no puede exceder %d caracteres", MaxEmailLength)
	}
	if !emailRegex.MatchString(data.Email) {
		return errors.New("el formato del correo electrónico no es válido")
	}

	// Validar mensaje
	if data.Message == "" {
		return errors.New("el mensaje es obligatorio")
	}
	if len(data.Message) < MinMessageLength {
		return fmt.Errorf("el mensaje debe tener al menos %d caracteres", MinMessageLength)
	}
	if len(data.Message) > MaxMessageLength {
		return fmt.Errorf("el mensaje no puede exceder %d caracteres", MaxMessageLength)
	}

	return nil
}

// emailTemplate es una plantilla HTML parseada para el correo.
var emailTemplate = template.Must(template.New("contactEmail").Parse(`<!DOCTYPE html>
<html lang="es" style="font-family: 'Montserrat', sans-serif; margin: 0; padding: 0;">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Nuevo Mensaje de Contacto - Softex Labs</title>
</head>
<body style="background-color: #f4f4f4; color: #333; padding: 20px;">
    <table width="100%" border="0" cellspacing="0" cellpadding="0">
        <tr>
            <td align="center">
                <table width="600" border="0" cellspacing="0" cellpadding="0" style="background-color: #ffffff; border-radius: 8px; box-shadow: 0 0 10px rgba(0, 0, 0, 0.1); padding: 30px;">
                    <!-- Header -->
                    <tr>
                        <td align="center" style="padding-bottom: 20px; border-bottom: 1px solid #eee;">
                            <h1 style="color: #4f46e5; margin: 10px 0 0 0;">Nuevo Mensaje de Contacto</h1>
                            <p style="color: #666; margin: 5px 0 0 0; font-size: 14px;">Recibido el {{.Timestamp}}</p>
                        </td>
                    </tr>
                    <!-- Body -->
                    <tr>
                        <td style="padding: 20px 0;">
                            <p>Has recibido un nuevo mensaje desde tu sitio web:</p>
                            <table width="100%" style="margin: 20px 0;">
                                <tr>
                                    <td style="padding: 10px; background-color: #f8f9fa; border-left: 4px solid #4f46e5;">
                                        <p style="margin: 0 0 10px 0;"><strong>Nombre:</strong> {{.Name}}</p>
                                        <p style="margin: 0 0 10px 0;"><strong>Email:</strong> {{.Email}}</p>
                                        <p style="margin: 0;"><strong>IP:</strong> {{.ClientIP}}</p>
                                    </td>
                                </tr>
                            </table>
                            <p><strong>Mensaje:</strong></p>
                            <div style="background-color: #f9f9f9; border-left: 4px solid #4f46e5; padding: 15px; margin: 10px 0; white-space: pre-wrap;">{{.Message}}</div>
                        </td>
                    </tr>
                    <!-- Footer -->
                    <tr>
                        <td align="center" style="padding-top: 20px; border-top: 1px solid #eee; font-size: 0.8em; color: #777;">
                            <p>&copy; {{.Year}} Softex Labs. Todos los derechos reservados.</p>
                            <p style="margin: 5px 0 0 0;">Este mensaje fue enviado desde softex-labs.xyz</p>
                        </td>
                    </tr>
                </table>
            </td>
        </tr>
    </table>
</body>
</html>`))

func formatEmailBody(data ContactData, clientIP string) (string, error) {
	var body bytes.Buffer
	templateData := struct {
		ContactData
		Year      int
		Timestamp string
		ClientIP  string
	}{
		ContactData: data,
		Year:        time.Now().Year(),
		Timestamp:   time.Now().Format("2006-01-02 15:04:05 UTC"),
		ClientIP:    clientIP,
	}

	if err := emailTemplate.Execute(&body, templateData); err != nil {
		return "", fmt.Errorf("error al generar el cuerpo del correo: %v", err)
	}

	return body.String(), nil
}

func sendEmail(config SmtpConfig, data ContactData, clientIP string) error {
	body, err := formatEmailBody(data, clientIP)
	if err != nil {
		return err
	}

	subject := fmt.Sprintf("Nuevo mensaje de contacto de %s - Softex Labs", data.Name)
	
	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
		config.User, config.ToEmail, subject, body)

	auth := smtp.PlainAuth("", config.User, config.Pass, config.Host)
	addr := fmt.Sprintf("%s:%s", config.Host, config.Port)

	err = smtp.SendMail(addr, auth, config.User, []string{config.ToEmail}, []byte(msg))
	if err != nil {
		log.Printf("Error al enviar correo: %v", err)
		return errors.New("no se pudo enviar el correo. Por favor, inténtalo de nuevo más tarde")
	}

	return nil
}

func parseAndValidateRequest(r *http.Request) (ContactData, error) {
	var data ContactData

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		return ContactData{}, errors.New("error al decodificar la solicitud JSON. Asegúrese de que el formato sea correcto")
	}

	if err := validateContactData(&data); err != nil {
		return ContactData{}, err
	}

	return data, nil
}

func getClientIP(r *http.Request) string {
	// Verificar headers de proxy en orden de prioridad
	headers := []string{
		"CF-Connecting-IP",    // Cloudflare
		"X-Forwarded-For",     // Estándar
		"X-Real-IP",           // Nginx
		"X-Client-IP",         // Apache
	}
	
	for _, header := range headers {
		ip := r.Header.Get(header)
		if ip != "" {
			// Tomar solo la primera IP si hay múltiples
			return strings.Split(strings.TrimSpace(ip), ",")[0]
		}
	}
	
	// Fallback para desarrollo local
	return strings.Split(r.RemoteAddr, ":")[0]
}

// Handler función que Vercel ejecutará.
func Handler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	clientIP := getClientIP(r)
	
	log.Printf("Solicitud recibida - Método: %s, IP: %s", r.Method, clientIP)

	// Configuración de CORS más segura
	allowedOrigin := os.Getenv("ALLOWED_ORIGIN")
	if allowedOrigin == "" {
		allowedOrigin = "https://softex-labs.xyz" // Dominio específico en producción
	}
	
	// Verificar origen para requests que no sean OPTIONS
	if r.Method != http.MethodOptions {
		origin := r.Header.Get("Origin")
		if origin != "" && origin != allowedOrigin && allowedOrigin != "*" {
			sendJSONError(w, "Origen no permitido", http.StatusForbidden)
			return
		}
	}
	
	w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Max-Age", "86400") // Cache preflight por 24 horas

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Aplicar el límite de tasa mejorado
	if err := checkRateLimit(clientIP); err != nil {
		log.Printf("Rate limit excedido para IP %s: %v", clientIP, err)
		sendJSONError(w, err.Error(), http.StatusTooManyRequests)
		return
	}

	if r.Method != http.MethodPost {
		sendJSONError(w, "Método no permitido. Solo se acepta POST.", http.StatusMethodNotAllowed)
		return
	}

	// 1. Parsear y validar la solicitud
	data, err := parseAndValidateRequest(r)
	if err != nil {
		log.Printf("Error de validación para IP %s: %v", clientIP, err)
		sendJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 2. Cargar la configuración
	config, err := newSmtpConfig()
	if err != nil {
		log.Printf("Error de configuración SMTP: %v", err)
		sendJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 3. Enviar el correo
	err = sendEmail(config, data, clientIP)
	if err != nil {
		log.Printf("Error al enviar correo para IP %s: %v", clientIP, err)
		sendJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	duration := time.Since(startTime)
	log.Printf("Correo enviado exitosamente - IP: %s, Duración: %v", clientIP, duration)
	
	sendJSONSuccess(w, "¡Mensaje enviado con éxito! Te responderemos pronto.")
}