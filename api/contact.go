package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"os"
)

// No necesitamos el rate limiter en memoria para el modelo serverless,
// ya que cada invocación puede ser una instancia nueva.
// Vercel tiene su propia protección contra ataques DDoS.

func sendJSONError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*") // CORS para Vercel
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// Handler es la función que Vercel ejecutará.
func Handler(w http.ResponseWriter, r *http.Request) {
	// Vercel maneja las rutas, pero es buena práctica verificar el método.
	if r.Method != http.MethodPost {
		sendJSONError(w, "Método no permitido. Solo se acepta POST.", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		sendJSONError(w, "Error al procesar el formulario.", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	email := r.FormValue("email")
	message := r.FormValue("message")

	if name == "" || email == "" || message == "" {
		sendJSONError(w, "Todos los campos (nombre, email, mensaje) son obligatorios.", http.StatusBadRequest)
		return
	}

	// En Vercel, las variables de entorno se configuran en el dashboard del proyecto.
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")

	toEmail := "grajajhon9@gmail.com"

	if smtpHost == "" || smtpPort == "" || smtpUser == "" || smtpPass == "" {
		log.Println("Error: Configuración SMTP incompleta en las variables de entorno de Vercel.")
		sendJSONError(w, "Error de configuración del servidor para enviar el correo.", http.StatusInternalServerError)
		return
	}

	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)

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

	err := smtp.SendMail(smtpAddr, auth, smtpUser, []string{toEmail}, []byte(emailBody))
	if err != nil {
		log.Printf("Error al enviar el correo: %v", err)
		sendJSONError(w, "Hubo un error interno al intentar enviar el correo.", http.StatusInternalServerError)
		return
	}

	log.Println("Correo del formulario de contacto enviado exitosamente.")
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*") // CORS para Vercel
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "¡Mensaje enviado con éxito!"})
}
