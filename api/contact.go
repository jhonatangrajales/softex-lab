package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
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

// ContactData representa los datos del formulario.
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

// emailRegex es una expresión regular simple para validar el formato del correo electrónico.
var emailRegex = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

// --- Rate Limiter ---

type requestInfo struct {
	count    int
	lastSeen time.Time
}

var (
	visitors = make(map[string]*requestInfo)
	mu       sync.Mutex
)

const (
	rateLimit  = 5               // Límite de solicitudes por ventana de tiempo
	timeWindow = 1 * time.Minute // Ventana de tiempo
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
			if time.Since(v.lastSeen) > timeWindow {
				delete(visitors, ip)
			}
		}
		mu.Unlock()
	}
}

// --- Fin del Rate Limiter ---

func sendJSONError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	// Las cabeceras CORS ahora se establecen en el Handler principal.
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
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

// emailTemplate es una plantilla HTML parseada para el correo.
// Usar html/template escapa automáticamente la entrada del usuario, previniendo ataques XSS.
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
                        </td>
                    </tr>
                    <!-- Body -->
                    <tr>
                        <td style="padding: 20px 0;">
                            <p>Has recibido un nuevo mensaje desde tu sitio web:</p>
                            <p style="margin-bottom: 15px;"><strong>Nombre:</strong> {{.Name}}</p>
                            <p style="margin-bottom: 15px;"><strong>Email de Contacto:</strong> {{.Email}}</p>
                            <p><strong>Mensaje:</strong></p>
                            <p style="background-color: #f9f9f9; border-left: 4px solid #4f46e5; padding: 15px; margin: 0;">{{.Message}}</p>
                        </td>
                    </tr>
                    <!-- Footer -->
                    <tr>
                        <td align="center" style="padding-top: 20px; border-top: 1px solid #eee; font-size: 0.8em; color: #777;">
                            <p>&copy; {{.Year}} Softex Labs. Todos los derechos reservados.</p>
                        </td>
                    </tr>
                </table>
            </td>
        </tr>
    </table>
</body>
</html>`))

func formatEmailBody(data ContactData) (string, error) {
	var body bytes.Buffer
	templateData := struct {
		Name    string
		Email   string
		Message string
		Year    int
	}{
		Name:    data.Name,
		Email:   data.Email,
		Message: data.Message,
		Year:    time.Now().Year(),
	}

	if err := emailTemplate.Execute(&body, templateData); err != nil {
		log.Printf("Error al ejecutar la plantilla de correo: %v", err)
		return "", errors.New("error interno al generar el cuerpo del correo")
	}
	return body.String(), nil
}

func sendEmail(config SmtpConfig, data ContactData) error {
	auth := smtp.PlainAuth("", config.User, config.Pass, config.Host)

	headers := "MIME-version: 1.0;\r\nContent-Type: text/html; charset=\"UTF-8\";\r\n"
	fromHeader := fmt.Sprintf("From: Softex Labs Contacto <%s>\r\n", config.User)
	toHeader := fmt.Sprintf("To: %s\r\n", config.ToEmail)
	replyToHeader := fmt.Sprintf("Reply-To: %s\r\n", data.Email)
	subjectHeader := "Subject: Nuevo Mensaje de Contacto - Softex Labs\r\n"

	msgBody, err := formatEmailBody(data)
	if err != nil {
		return err
	}

	// Combina todos los headers y el cuerpo del mensaje.
	emailBody := fromHeader + toHeader + replyToHeader + subjectHeader + headers + "\r\n" + msgBody

	smtpAddr := fmt.Sprintf("%s:%s", config.Host, config.Port)

	err = smtp.SendMail(smtpAddr, auth, config.User, []string{config.ToEmail}, []byte(emailBody))
	if err != nil {
		log.Printf("Error al enviar el correo: %v", err)
		return errors.New("hubo un error interno al intentar enviar el correo. Por favor, inténtelo de nuevo más tarde")
	}

	return nil
}

// parseAndValidateRequest extrae y valida los datos del formulario de la solicitud HTTP.
func parseAndValidateRequest(r *http.Request) (ContactData, error) {
	var data ContactData

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		return ContactData{}, errors.New("error al decodificar la solicitud JSON. Asegúrese de que el formato sea correcto")
	}

	if data.Name == "" || data.Email == "" || data.Message == "" {
		return ContactData{}, errors.New("todos los campos (nombre, email, mensaje) son obligatorios")
	}

	if !emailRegex.MatchString(data.Email) {
		return ContactData{}, errors.New("el formato del correo electrónico no es válido")
	}

	return data, nil
}

func getClientIP(r *http.Request) string {
	ip := r.Header.Get("X-Forwarded-For")
	if ip != "" {
		return strings.Split(ip, ",")[0]
	}
	// Fallback para desarrollo local.
	return r.RemoteAddr
}

// Handler  función que Vercel ejecutará.
func Handler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Invocada función de contacto con método: %s", r.Method) //para pruebas

	// Configuración de CORS más segura y flexible a través de variables de entorno.
	allowedOrigin := os.Getenv("ALLOWED_ORIGIN")
	if allowedOrigin == "" {
		// Fallback para desarrollo local si no está definida. En producción,
		// se debe establecer la variable de entorno al dominio del frontend.
		allowedOrigin = "*"
	}
	w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Aplicar el límite de tasa (Rate Limiting)
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
			sendJSONError(w, "Has excedido el límite de solicitudes. Por favor, inténtalo de nuevo más tarde.", http.StatusTooManyRequests)
			return
		}
	} else {
		visitors[clientIP] = &requestInfo{count: 1, lastSeen: time.Now()}
	}
	mu.Unlock()

	if r.Method != http.MethodPost {
		sendJSONError(w, "Método no permitido. Solo se acepta POST.", http.StatusMethodNotAllowed)
		return
	}

	// 1. Parsear y validar la solicitud
	data, err := parseAndValidateRequest(r)
	if err != nil {
		sendJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 2. Cargar la configuración
	config, err := newSmtpConfig()
	if err != nil {
		sendJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 3. Enviar el correo
	err = sendEmail(config, data)
	if err != nil {
		sendJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Println("Correo del formulario de contacto enviado exitosamente.")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "¡Mensaje enviado con éxito!"})
}
