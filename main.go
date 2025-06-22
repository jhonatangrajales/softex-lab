package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/smtp"
	"os"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"golang.org/x/time/rate"
)

func sendJSONError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// --- Rate Limiter Implementation ---

// visitor struct holds a rate limiter and the last time it was seen.
type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// Use a mutex to protect the visitors map from concurrent access.
var visitors = make(map[string]*visitor)
var mu sync.Mutex

// getVisitorLimiter retrieves or creates a rate limiter for a given IP.
func getVisitorLimiter(ip string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	v, exists := visitors[ip]
	if !exists {
		// Allow 1 request every 10 seconds, with a burst of 3.
		limiter := rate.NewLimiter(rate.Every(10*time.Second), 3)
		visitors[ip] = &visitor{limiter, time.Now()}
		return limiter
	}

	v.lastSeen = time.Now()
	return v.limiter
}

func contactHandler(w http.ResponseWriter, r *http.Request) {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		sendJSONError(w, "Error al identificar la dirección IP.", http.StatusInternalServerError)
		return
	}

	if r.Method != http.MethodPost {
		sendJSONError(w, "Método no permitido. Solo se acepta POST.", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		sendJSONError(w, "Error al procesar el formulario.", http.StatusBadRequest)
		return
	}

	// Check the rate limiter for the current IP.
	limiter := getVisitorLimiter(ip)
	if !limiter.Allow() {
		sendJSONError(w, "Has enviado demasiadas solicitudes. Por favor, espera un momento.", http.StatusTooManyRequests)
		return
	}

	name := r.FormValue("name")
	email := r.FormValue("email")
	message := r.FormValue("message")

	if name == "" || email == "" || message == "" {
		sendJSONError(w, "Todos los campos (nombre, email, mensaje) son obligatorios.", http.StatusBadRequest)
		return
	}

	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")

	toEmail := "grajajhon9@gmail.com" // Correo de destino

	if smtpHost == "" || smtpPort == "" || smtpUser == "" || smtpPass == "" {
		log.Println("Error: Configuración SMTP incompleta. Define las variables de entorno SMTP_HOST, SMTP_PORT, SMTP_USER, SMTP_PASS.")
		sendJSONError(w, "Error de configuración del servidor para enviar el correo.", http.StatusInternalServerError)
		return
	}

	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)

	// Construir un cuerpo de correo más robusto con cabeceras MIME para evitar filtros de spam.
	headers := "MIME-version: 1.0;\nContent-Type: text/plain; charset=\"UTF-8\";\n"
	fromHeader := fmt.Sprintf("From: Softex Labs Contacto <%s>\r\n", smtpUser)
	toHeader := fmt.Sprintf("To: %s\r\n", toEmail)
	subjectHeader := "Subject: Nuevo Mensaje de Contacto - Softex Labs\r\n"

	msgBody := fmt.Sprintf("Has recibido un nuevo mensaje desde tu sitio web:\n\n"+
		"Nombre: %s\n"+
		"Email de Contacto: %s\n\n"+
		"Mensaje:\n%s\n", name, email, message)

	emailBody := fromHeader + toHeader + subjectHeader + headers + "\r\n" + msgBody

	smtpAddr := fmt.Sprintf("%s:%s", smtpHost, smtpPort)

	err = smtp.SendMail(smtpAddr, auth, smtpUser, []string{toEmail}, []byte(emailBody))
	if err != nil {
		log.Printf("Error al enviar el correo: %v", err)
		sendJSONError(w, "Hubo un error interno al intentar enviar el correo.", http.StatusInternalServerError)
		return
	}

	log.Println("Correo del formulario de contacto enviado exitosamente.")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "¡Mensaje enviado con éxito!"})
}

// Periodically clean up old entries from the visitors map.
func cleanupVisitors() {
	for {
		time.Sleep(1 * time.Minute)
		mu.Lock()
		for ip, v := range visitors {
			if time.Since(v.lastSeen) > 3*time.Minute {
				delete(visitors, ip)
			}
		}
		mu.Unlock()
	}
}

func main() {
	// Cargar variables de entorno desde .env para desarrollo local.
	// En un servidor de producción, estas se configuran directamente en el entorno.
	err := godotenv.Load()
	if err != nil {
		log.Println("Advertencia: No se pudo cargar el archivo .env. Se usarán las variables de entorno del sistema si existen.")
	}

	go cleanupVisitors()

	// Primero el endpoint para el formulario de contacto.
	http.HandleFunc("/api/contact", contactHandler)

	// Luego el servidor de archivos estáticos.
	fs := http.FileServer(http.Dir("."))
	http.Handle("/", fs)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Puerto por defecto
	}

	log.Printf("Servidor iniciado en http://localhost:%s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("No se pudo iniciar el servidor: %s\n", err.Error())
	}
}
