# -SOFTEX-LABS-

Este es el repositorio para el landing page de Softex Labs, una aplicación web estática con un backend serverless en Go para el formulario de contacto, desplegada en Vercel.

## Despliegue y Configuración

El proyecto está diseñado para ser desplegado en Vercel. Para que el formulario de contacto funcione correctamente, es necesario configurar las siguientes variables de entorno en el panel de Vercel.

### Variables de Entorno

Ve a `Settings` > `Environment Variables` en tu proyecto de Vercel y añade las siguientes variables:

-   `SMTP_HOST`: El host de tu servidor SMTP (ej. `smtp.gmail.com`).
-   `SMTP_PORT`: El puerto de tu servidor SMTP (ej. `587`).
-   `SMTP_USER`: Tu nombre de usuario SMTP (tu correo electrónico).
-   `SMTP_PASS`: La contraseña de aplicación para tu cuenta de correo.
-   `TO_EMAIL`: El correo electrónico donde se recibirán los mensajes del formulario. Si no se especifica, se usará un valor por defecto.
-   `ALLOWED_ORIGIN`: El dominio de tu frontend para la configuración de CORS. Es crucial para la seguridad.
    -   **Ejemplo en producción:** `https://www.softex-labs.com`

Sin estas variables, la función serverless (`/api/contact`) fallará.