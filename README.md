# -SOFTEX-LABS-

Este es el repositorio para el landing page de Softex Labs, una aplicación web estática con un backend serverless en Go para el formulario de contacto, desplegada en Vercel.

## Despliegue y Configuración

El proyecto está diseñado para ser desplegado en Vercel. Para que el formulario de contacto funcione correctamente, es necesario configurar las siguientes variables de entorno en el panel de Vercel.

### Variables de Entorno

Ve a `Settings` > `Environment Variables` en tu proyecto de Vercel y añade las siguientes variables:

-   `SMTP_HOST`: El host del servidor SMTP (ej. `smtp.gmail.com`).
-   `SMTP_PORT`: El puerto del servidor SMTP (ej. `587`).
-   `SMTP_USER`: nombre de usuario SMTP.
-   `SMTP_PASS`: La contraseña de aplicación para cuenta de correo.
-   `TO_EMAIL`: El correo electrónico donde se recibirán los mensajes del formulario. Si no se especifica, se usará un valor por defecto.
-   `ALLOWED_ORIGIN`: El dominio de tu frontend para la configuración de CORS. Es crucial para la seguridad.
    -   **Ejemplo en producción:** `https://www.softex-labs.com` 
    -   **Ejemplo en desarrollo:** ` https://softex-labs.vercel.app` 

Sin estas variables, la función serverless (`/api/contact`) fallará.