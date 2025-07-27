# Softex Labs - Landing Page

## 🚀 Mejoras Implementadas

### ✅ Seguridad y Validación
- **Rate limiting mejorado** con bloqueo temporal por IP (3 requests/5min)
- **Validación robusta** con límites de caracteres y sanitización
- **CORS configurado** de forma segura
- **Sanitización de entrada** para prevenir XSS
- **Validación de email** con regex mejorada
- **Logging de seguridad** para monitoreo

### ✅ Experiencia de Usuario (UX)
- **Validación en tiempo real** del formulario
- **Contadores de caracteres** con indicadores visuales
- **Alertas mejoradas** con botones de cierre
- **Estados de carga** con spinner animado
- **Mensajes de error/éxito** más informativos
- **Accesibilidad mejorada** con ARIA labels
- **Navegación móvil** optimizada

### ✅ Performance y Optimización
- **CSS separado** del HTML para mejor mantenimiento
- **Service Worker** para cache offline
- **PWA capabilities** con manifest
- **Lazy loading** para imágenes
- **Preload de recursos** críticos
- **Animaciones optimizadas** con CSS

### ✅ Mantenibilidad del Código
- **Código modular** con funciones separadas
- **Variables CSS** para consistencia
- **Comentarios detallados** en el código
- **Estructura de archivos** organizada
- **Tests unitarios** para el backend
- **Logging estructurado** con niveles

## 📁 Estructura de Archivos

```
├── README.md                 # Documentación principal
├── index.html               # Landing page mejorada
├── styles.css               # Estilos CSS separados
├── app.js                   # JavaScript mejorado
├── sw.js                    # Service Worker para PWA
├── site.webmanifest         # Manifest para PWA
├── vercel.json              # Configuración de Vercel
├── .vercelignore            # Archivos a ignorar en deployment
├── go.mod                   # Dependencias Go
├── package.json             # Dependencias Node.js
├── api/
│   ├── contact.go           # Handler principal mejorado
│   ├── contact_test.go      # Tests unitarios
│   └── health.go            # Health check endpoint
└── .github/
    └── workflows/
        └── ci-cd.yml        # Pipeline CI/CD
```

## 🔧 Variables de Entorno

### Obligatorias para Vercel
```bash
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=tu-email@gmail.com
SMTP_PASS=tu-app-password
TO_EMAIL=contacto@softex-labs.xyz
ALLOWED_ORIGIN=https://tu-dominio.vercel.app
```

## 🚀 Despliegue en Vercel

### Paso 1: Preparar el Repositorio
1. Asegúrate de que todos los archivos estén en tu repositorio Git
2. Haz commit y push de todos los cambios

### Paso 2: Conectar a Vercel
1. Ve a [vercel.com](https://vercel.com) e inicia sesión
2. Haz clic en "New Project"
3. Importa tu repositorio desde GitHub/GitLab/Bitbucket

### Paso 3: Configurar Variables de Entorno
En el dashboard de Vercel, ve a `Settings` > `Environment Variables` y añade:

| Variable | Valor | Descripción |
|----------|-------|-------------|
| `SMTP_HOST` | `smtp.gmail.com` | Servidor SMTP |
| `SMTP_PORT` | `587` | Puerto SMTP |
| `SMTP_USER` | `tu-email@gmail.com` | Tu email |
| `SMTP_PASS` | `tu-app-password` | Contraseña de aplicación |
| `TO_EMAIL` | `contacto@softex-labs.xyz` | Email destino |
| `ALLOWED_ORIGIN` | `https://tu-dominio.vercel.app` | Dominio permitido |

### Paso 4: Deploy
1. Haz clic en "Deploy"
2. Vercel detectará automáticamente la configuración Go
3. El deployment debería completarse sin errores

## 🔍 Verificar el Deployment

### Health Check
```bash
curl https://tu-dominio.vercel.app/api/health
```

Respuesta esperada:
```json
{
  "status": "healthy",
  "timestamp": "2024-01-01T12:00:00Z",
  "version": "2.0.0",
  "service": "softex-labs-contact-api"
}
```

### Test del Formulario
```bash
curl -X POST https://tu-dominio.vercel.app/api/contact \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test User",
    "email": "test@example.com", 
    "message": "Este es un mensaje de prueba"
  }'
```

## 🧪 Testing Local

```bash
# Ejecutar tests
go test ./...

# Tests con coverage
go test -v -coverprofile=coverage.out ./...

# Ver coverage en HTML
go tool cover -html=coverage.out

# Servidor local para desarrollo
npm run dev
```

## 🛡️ Seguridad

- **Rate limiting**: 3 requests por 5 minutos por IP
- **Validación de entrada**: Sanitización automática
- **CORS**: Configurado para dominios específicos
- **TLS**: Conexión segura para SMTP
- **Headers de seguridad**: Implementados automáticamente por Vercel

## 📱 PWA Features

- **Offline capability** con Service Worker
- **Installable** en dispositivos móviles
- **Cache inteligente** de recursos estáticos
- **Manifest** configurado para app-like experience

## 🎨 Personalización

### Colores
Modifica las variables CSS en `styles.css`:
```css
:root {
  --primary-color: #4f46e5;
  --secondary-color: #06b6d4;
  --accent-color: #f59e0b;
}
```

### Contenido
Edita `index.html` para cambiar textos, servicios y información de contacto.

## 🔧 Troubleshooting

### Error: "Origen no permitido"
- Verifica que `ALLOWED_ORIGIN` coincida exactamente con tu dominio
- Para desarrollo local, usa `ALLOWED_ORIGIN=*`

### Error: "Faltan variables de entorno SMTP"
- Asegúrate de configurar todas las variables SMTP en Vercel
- Verifica que no haya espacios extra en los valores

### Error: "Rate limit excedido"
- Espera 5 minutos antes de intentar nuevamente
- Para desarrollo, puedes reiniciar la función

### Error de deployment
- Verifica que `vercel.json` tenga formato JSON válido
- Asegúrate de que `go.mod` esté en la raíz del proyecto

## 📞 Soporte

Para soporte técnico o consultas sobre las mejoras implementadas:
- Email: contacto@softex-labs.xyz
- Revisa los logs en el dashboard de Vercel
- Consulta la documentación de Vercel para Go

---

**Softex Labs** - Transformando negocios con tecnología innovadora 🚀