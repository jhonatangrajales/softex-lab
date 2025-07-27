package api

import (
	"encoding/json"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"strings"
	"time"
)

// NotificationConfig representa la configuración de notificaciones
type NotificationConfig struct {
	WebhookURL   string `json:"webhook_url"`
	SlackChannel string `json:"slack_channel"`
	EmailEnabled bool   `json:"email_enabled"`
	SlackEnabled bool   `json:"slack_enabled"`
}

// SlackMessage representa un mensaje de Slack
type SlackMessage struct {
	Channel     string            `json:"channel"`
	Username    string            `json:"username"`
	IconEmoji   string            `json:"icon_emoji"`
	Attachments []SlackAttachment `json:"attachments"`
}

// SlackAttachment representa un attachment de Slack
type SlackAttachment struct {
	Color     string       `json:"color"`
	Title     string       `json:"title"`
	Text      string       `json:"text"`
	Fields    []SlackField `json:"fields"`
	Timestamp int64        `json:"ts"`
}

// SlackField representa un campo de Slack
type SlackField struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

// Enviar notificación a Slack
func sendSlackNotification(data ContactData, clientIP string) error {
	webhookURL := os.Getenv("SLACK_WEBHOOK_URL")
	if webhookURL == "" {
		return nil // No configurado, no es error
	}

	message := SlackMessage{
		Channel:   "#contacto",
		Username:  "Softex Labs Bot",
		IconEmoji: ":email:",
		Attachments: []SlackAttachment{
			{
				Color: "good",
				Title: "Nuevo mensaje de contacto",
				Text:  "Se ha recibido un nuevo mensaje desde el sitio web",
				Fields: []SlackField{
					{Title: "Nombre", Value: data.Name, Short: true},
					{Title: "Email", Value: data.Email, Short: true},
					{Title: "IP", Value: clientIP, Short: true},
					{Title: "Fecha", Value: time.Now().Format("2006-01-02 15:04:05"), Short: true},
					{Title: "Mensaje", Value: data.Message, Short: false},
				},
				Timestamp: time.Now().Unix(),
			},
		},
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		return err
	}

	resp, err := http.Post(webhookURL, "application/json", strings.NewReader(string(jsonData)))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err
	}

	log.Println("Notificación de Slack enviada exitosamente")
	return nil
}

// Enviar auto-respuesta al usuario
func sendAutoResponse(config SmtpConfig, data ContactData) error {
	autoResponseEnabled := os.Getenv("AUTO_RESPONSE_ENABLED")
	if autoResponseEnabled != "true" {
		return nil
	}

	subject := "Gracias por contactar con Softex Labs"

	body := `<!DOCTYPE html>
<html lang="es">
<head>
    <meta charset="UTF-8">
    <title>Gracias por contactarnos</title>
</head>
<body style="font-family: Arial, sans-serif; background-color: #f4f4f4; padding: 20px;">
    <div style="max-width: 600px; margin: 0 auto; background: white; padding: 30px; border-radius: 8px;">
        <h1 style="color: #4f46e5;">¡Gracias por contactarnos!</h1>
        <p>Hola <strong>` + data.Name + `</strong>,</p>
        <p>Hemos recibido tu mensaje y te responderemos pronto.</p>
        <div style="background: #f9f9f9; padding: 15px; border-left: 4px solid #4f46e5; margin: 15px 0;">
            ` + data.Message + `
        </div>
        <p>Saludos,<br>Equipo Softex Labs</p>
    </div>
</body>
</html>`

	msg := "From: " + config.User + "\r\n" +
		"To: " + data.Email + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=UTF-8\r\n\r\n" +
		body

	auth := smtp.PlainAuth("", config.User, config.Pass, config.Host)
	addr := config.Host + ":" + config.Port

	err := smtp.SendMail(addr, auth, config.User, []string{data.Email}, []byte(msg))
	if err != nil {
		log.Printf("Error enviando auto-respuesta: %v", err)
		return err
	}

	log.Println("Auto-respuesta enviada exitosamente")
	return nil
}

// Handler mejorado con notificaciones
func HandlerWithNotifications(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	clientIP := getClientIP(r)

	log.Printf("Solicitud recibida - Método: %s, IP: %s", r.Method, clientIP)

	// Configuración de CORS
	allowedOrigin := os.Getenv("ALLOWED_ORIGIN")
	if allowedOrigin == "" {
		allowedOrigin = "https://softex-labs.xyz"
	}

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
		sendJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Enviar email principal
	err = sendEmail(config, data, clientIP)
	if err != nil {
		log.Printf("Error al enviar correo para IP %s: %v", clientIP, err)
		sendJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Enviar notificaciones adicionales (no bloquear si fallan)
	go func() {
		if err := sendSlackNotification(data, clientIP); err != nil {
			log.Printf("Error enviando notificación Slack: %v", err)
		}

		if err := sendAutoResponse(config, data); err != nil {
			log.Printf("Error enviando auto-respuesta: %v", err)
		}
	}()

	duration := time.Since(startTime)
	log.Printf("Correo enviado exitosamente - IP: %s, Duración: %v", clientIP, duration)

	sendJSONSuccess(w, "¡Mensaje enviado con éxito! Te responderemos pronto.")
}
