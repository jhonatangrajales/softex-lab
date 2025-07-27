# Softex Labs - Landing Page

## 🚀 Mejoras Implementadas

### ✅ Seguridad y Validación
- **Rate limiting mejorado** con bloqueo temporal por IP
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
- **Lazy loading** para imágenes
- **Preload de recursos** críticos
- **Animaciones optimizadas** con CSS
- **Responsive design** mejorado

### ✅ Mantenibilidad del Código
- **Código modular** con funciones separadas
- **Variables CSS** para consistencia
- **Comentarios detallados** en el código
- **Estructura de archivos** organizada
- **Tests unitarios** para el backend
- **Logging estructurado** con niveles

### ✅ Funcionalidades Adicionales
- **Sistema de analytics** para métricas del formulario
- **Notificaciones Slack** para nuevos mensajes
- **Auto-respuesta** automática a usuarios
- **PWA capabilities** con manifest
- **Health checks** para monitoreo
- **CI/CD pipeline** con GitHub Actions

### ✅ Testing y Calidad
- **Tests unitarios** completos para el backend
- **Validación de entrada** exhaustiva
- **Rate limiting** probado
- **Sanitización** de datos testada
- **Pipeline CI/CD** automatizado

## 📁 Estructura de Archivos

```
├── README.md                 # Documentación principal
├── index.html               # Landing page mejorada
├── styles.css               # Estilos CSS separados
├── app.js                   # JavaScript mejorado
├── sw.js                    # Service Worker para PWA
├── site.webmanifest         # Manifest para PWA
├── go.mod                   # Dependencias Go
├── go.sum                   # Checksums de dependencias
├── package.json             # Dependencias Node.js
├── api/
│   ├── contact.go           # Handler principal mejorado
│   ├── contact_test.go      # Tests unitarios
│   ├── analytics.go         # Sistema de analytics
│   └── notifications.go     # Sistema de notificaciones
└── .github/
    └── workflows/
        └── ci-cd.yml        # Pipeline CI/CD
```

## 🔧 Variables de Entorno

### Obligatorias
```bash
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=tu-email@gmail.com
SMTP_PASS=tu-app-password
TO_EMAIL=contacto@softex-labs.xyz
ALLOWED_ORIGIN=https://softex-labs.xyz
```

### Opcionales
```bash
# Notificaciones Slack
SLACK_WEBHOOK_URL=https://hooks.slack.com/services/...

# Auto-respuesta
AUTO_RESPONSE_ENABLED=true

# Analytics
ADMIN_KEY=tu-clave-secreta-admin

# Versión de la app
APP_VERSION=1.0.0
```

## 🚀 Despliegue

### Vercel (Recomendado)
1. Conecta tu repositorio a Vercel
2. Configura las variables de entorno en el dashboard
3. El despliegue es automático con cada push

### Variables de Entorno en Vercel
Ve a `Settings` > `Environment Variables` y añade:
- `SMTP_HOST`, `SMTP_PORT`, `SMTP_USER`, `SMTP_PASS`
- `TO_EMAIL`, `ALLOWED_ORIGIN`
- Opcionalmente: `SLACK_WEBHOOK_URL`, `AUTO_RESPONSE_ENABLED`, `ADMIN_KEY`

## 🧪 Testing

```bash
# Ejecutar tests
go test ./...

# Tests con coverage
go test -v -coverprofile=coverage.out ./...

# Ver coverage en HTML
go tool cover -html=coverage.out
```

## 📊 Analytics

Accede a las métricas del formulario:
```bash
curl -H "X-Admin-Key: tu-clave-admin" https://tu-dominio.com/api/analytics
```

## 🔔 Notificaciones

### Slack
1. Crea un webhook en tu workspace de Slack
2. Configura `SLACK_WEBHOOK_URL` en las variables de entorno
3. Los mensajes aparecerán automáticamente en el canal configurado

### Auto-respuesta
1. Configura `AUTO_RESPONSE_ENABLED=true`
2. Los usuarios recibirán una confirmación automática

## 🛡️ Seguridad

- **Rate limiting**: 3 requests por 5 minutos por IP
- **Validación de entrada**: Sanitización automática
- **CORS**: Configurado para dominios específicos
- **Headers de seguridad**: Implementados automáticamente

## 📱 PWA Features

- **Offline capability** con Service Worker
- **Installable** en dispositivos móviles
- **Cache inteligente** de recursos estáticos
- **Manifest** configurado para app-like experience

## 🔍 Monitoreo

### Health Check
```bash
curl https://tu-dominio.com/health
```

### Métricas
```bash
curl https://tu-dominio.com/metrics
```

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

## 🚀 Próximas Mejoras

- [ ] Dashboard administrativo
- [ ] Integración con CRM
- [ ] Múltiples idiomas (i18n)
- [ ] Chat en vivo
- [ ] A/B testing
- [ ] Métricas avanzadas con Google Analytics 4

## 📞 Soporte

Para soporte técnico o consultas sobre las mejoras implementadas, contacta al equipo de desarrollo.

---

**Softex Labs** - Transformando negocios con tecnología innovadora 🚀