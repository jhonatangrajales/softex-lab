package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"regexp"
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
var emailRegex = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)

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
		config.ToEmail = "grajajhon9@gmail.com" // Fallback si no se configura TO_EMAIL
		log.Println("Advertencia: TO_EMAIL no configurado. Usando el valor por defecto:", config.ToEmail)
	}

	return config, nil
}

func formatEmailBody(data ContactData) string {
	// Usamos comillas invertidas (backticks) para un string multilínea en Go.
	emailTemplate := `<!DOCTYPE html>
<html lang="es">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Nuevo Mensaje de Contacto - Softex Labs</title>
    <style>
        body {
            font-family: 'Montserrat', sans-serif;
            background-color: #f4f4f4;
            margin: 0;
            padding: 0;
            color: #333;
        }
        .container {
            width: 100%;
            max-width: 600px;
            margin: 20px auto;
            background-color: #ffffff;
            padding: 30px;
            border-radius: 8px;
            box-shadow: 0 0 10px rgba(0, 0, 0, 0.1);
        }
        .header {
            text-align: center;
            margin-bottom: 20px;
        }
        .logo {
            width: 150px;
            height: auto;
        }
        h1 {
            color: #4f46e5;
        }
        .message {
            margin-top: 20px;
        }
        .field {
            margin-bottom: 10px;
        }
        .label {
            font-weight: bold;
        }
        .footer {
            margin-top: 30px;
            text-align: center;
            font-size: 0.8em;
            color: #777;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Nuevo Mensaje de Contacto</h1>
        </div>
        <div class="message">
            <p>Has recibido un nuevo mensaje desde tu sitio web:</p>
            <div class="field">
                <span class="label">Nombre:</span> ` + data.Name + `
            </div>
            <div class="field">
                <span class="label">Email de Contacto:</span> ` + data.Email + `
            </div>
            <div class="field">
                <span class="label">Mensaje:</span>
                <p>` + data.Message + `</p>
            </div>
        </div>
        <div class="footer">
            <p>&copy; ` + fmt.Sprintf("%d", 2024) + ` Softex Labs. Todos los derechos reservados.</p>
        </div>
    </div>
</body>
</html>`

	return emailTemplate
}

// sendEmail construye y envía el correo electrónico.
func sendEmail(config SmtpConfig, data ContactData) error {
	auth := smtp.PlainAuth("", config.User, config.Pass, config.Host)

	// Content-Type ahora es text/html para renderizar el HTML.
	headers := "MIME-version: 1.0;\r\nContent-Type: text/html; charset=\"UTF-8\";\r\n"
	fromHeader := fmt.Sprintf("From: Softex Labs Contacto <%s>\r\n", config.User)
	toHeader := fmt.Sprintf("To: %s\r\n", config.ToEmail)
	subjectHeader := "Subject: Nuevo Mensaje de Contacto - Softex Labs\r\n"

	// Usamos la función para formatear el cuerpo del correo con la plantilla HTML.
	emailBody := fromHeader + toHeader + subjectHeader + headers + "\r\n" + formatEmailBody(data)

	smtpAddr := fmt.Sprintf("%s:%s", config.Host, config.Port)

	err := smtp.SendMail(smtpAddr, auth, config.User, []string{config.ToEmail}, []byte(emailBody))
	if err != nil {
		log.Printf("Error al enviar el correo: %v", err)
		return errors.New("hubo un error interno al intentar enviar el correo. Por favor, inténtelo de nuevo más tarde")
	}

	return nil
}

// parseAndValidateRequest extrae y valida los datos del formulario de la solicitud HTTP.
func parseAndValidateRequest(r *http.Request) (ContactData, error) {
	var data ContactData

	// Ahora solo aceptamos JSON, lo que simplifica y estandariza el código.
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

// Handler es la función que Vercel ejecutará.
func Handler(w http.ResponseWriter, r *http.Request) {
	// Log para confirmar que la función fue invocada. Esto es clave para el diagnóstico.
	log.Printf("Invocada función de contacto con método: %s", r.Method)

	// Configurar cabeceras CORS para todas las respuestas.
	// Esto es crucial para que los navegadores permitan las solicitudes desde tu frontend.
	w.Header().Set("Access-Control-Allow-Origin", "*") // En producción, considera restringir esto a tu dominio: "https://softex-labs.vercel.app"
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Manejar la solicitud pre-vuelo (preflight) de CORS que envía el navegador.
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

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
