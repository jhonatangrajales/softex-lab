# Softex Labs - Landing Page

## 🚨 SOLUCIÓN PARA PROBLEMAS DE DEPLOYMENT

### ✅ **Problemas Resueltos:**
- ❌ Error: "No Output Directory named 'public' found" → ✅ **SOLUCIONADO**
- ❌ CI/CD Pipeline fallando → ✅ **SIMPLIFICADO Y CORREGIDO**
- ❌ Funciones Go no desplegando → ✅ **CONFIGURACIÓN CORREGIDA**
- ❌ Build process fallando → ✅ **SCRIPTS MEJORADOS**

## 🚀 **PASOS URGENTES PARA DEPLOYMENT**

### 1. **Verificar Build Local** ✅
```bash
# Limpiar y construir
npm run clean
npm run build

# Verificar que se creó public/ con todos los archivos
ls -la public/
```

### 2. **Verificar Funciones Go** ✅
```bash
# Verificar módulos Go
go mod tidy
go vet ./...

# Verificar estructura de archivos
ls -la api/contact/
ls -la api/health/
```

### 3. **Commit y Push** 🔥
```bash
git add .
git commit -m "fix: Resolve Vercel deployment issues with proper build configuration"
git push origin main
```

### 4. **Configurar Variables en Vercel** 🔧
En el dashboard de Vercel → Settings → Environment Variables:

| Variable | Valor | Ejemplo |
|----------|-------|---------|
| `SMTP_HOST` | `smtp.gmail.com` | smtp.gmail.com |
| `SMTP_PORT` | `587` | 587 |
| `SMTP_USER` | Tu email | contacto@softex-labs.xyz |
| `SMTP_PASS` | App password | abcd efgh ijkl mnop |
| `TO_EMAIL` | Email destino | contacto@softex-labs.xyz |
| `ALLOWED_ORIGIN` | Tu dominio | https://tu-proyecto.vercel.app |

### 5. **Redeploy en Vercel** 🚀
1. Ve a tu proyecto en Vercel
2. Haz clic en "Redeploy"
3. Selecciona "Use existing Build Cache" = NO
4. Haz clic en "Redeploy"

## 📁 **Estructura Final Correcta**

```
├── README.md                 # Documentación
├── package.json             # Scripts de build corregidos
├── vercel.json              # Configuración Vercel v2
├── .vercelignore            # Archivos a ignorar
├── .gitignore               # Git ignore
├── go.mod                   # Módulos Go
├── index.html               # Archivo fuente
├── styles.css               # Estilos fuente
├── app.js                   # JavaScript fuente
├── sw.js                    # Service Worker fuente
├── site.webmanifest         # PWA manifest fuente
├── public/                  # 📂 DIRECTORIO DE BUILD
│   ├── index.html           # ✅ Generado por build
│   ├── styles.css           # ✅ Generado por build
│   ├── app.js               # ✅ Generado por build
│   ├── sw.js                # ✅ Generado por build
│   └── site.webmanifest     # ✅ Generado por build
├── api/                     # 📂 FUNCIONES SERVERLESS
│   ├── contact/
│   │   └── index.go         # ✅ Handler contacto
│   └── health/
│       └── index.go         # ✅ Handler health check
└── .github/
    └── workflows/
        └── ci-cd.yml        # ✅ Pipeline simplificado
```

## 🔧 **Configuración Vercel Corregida**

### `vercel.json` - Configuración v2:
```json
{
  "version": 2,
  "buildCommand": "npm run build",
  "outputDirectory": "public",
  "functions": {
    "api/contact/index.go": {
      "runtime": "@vercel/go@3.0.0"
    },
    "api/health/index.go": {
      "runtime": "@vercel/go@3.0.0"
    }
  },
  "routes": [
    {
      "src": "/api/contact",
      "dest": "/api/contact/index.go"
    },
    {
      "src": "/api/health", 
      "dest": "/api/health/index.go"
    }
  ]
}
```

### `package.json` - Scripts mejorados:
```json
{
  "scripts": {
    "build": "rm -rf public && mkdir -p public && cp index.html styles.css app.js sw.js site.webmanifest public/ && echo 'Build completed successfully'",
    "prebuild": "echo 'Starting build process...'",
    "postbuild": "echo 'Build process finished. Files in public:' && ls -la public/",
    "clean": "rm -rf public",
    "dev": "npx http-server public -p 3000 -o"
  }
}
```

## 🧪 **Testing y Verificación**

### Build Local:
```bash
npm run build
# Debe mostrar: "Build completed successfully"
# Debe crear directorio public/ con 5 archivos
```

### Funciones Go:
```bash
go mod verify
go vet ./...
# No debe mostrar errores
```

### Test Endpoints (después del deployment):
```bash
# Health check
curl https://tu-dominio.vercel.app/api/health

# Test formulario
curl -X POST https://tu-dominio.vercel.app/api/contact \
  -H "Content-Type: application/json" \
  -H "Origin: https://tu-dominio.vercel.app" \
  -d '{
    "name": "Test User",
    "email": "test@example.com",
    "message": "Mensaje de prueba desde curl"
  }'
```

## 🛡️ **Funcionalidades Implementadas**

### ✅ **Seguridad:**
- Rate limiting: 3 requests/5min por IP
- Validación robusta de entrada
- Sanitización contra XSS
- CORS configurado correctamente
- TLS para SMTP

### ✅ **UX/UI:**
- Validación en tiempo real
- Contadores de caracteres
- Estados de carga con spinner
- Alertas mejoradas
- Responsive design
- Accesibilidad (ARIA labels)

### ✅ **Performance:**
- Service Worker para cache offline
- PWA capabilities
- CSS y JS optimizados
- Lazy loading
- Preload de recursos críticos

### ✅ **Mantenibilidad:**
- Código modular y comentado
- Variables CSS para consistencia
- Estructura de archivos organizada
- Logging estructurado
- Pipeline CI/CD simplificado

## 🔍 **Troubleshooting Específico**

### Error: "Build failed"
```bash
# Limpiar cache y rebuilds
npm run clean
rm -rf node_modules package-lock.json
npm install
npm run build
```

### Error: "Function not found"
- Verificar que `api/contact/index.go` y `api/health/index.go` existen
- Verificar que ambos usan `package handler`
- Verificar que ambos tienen función `Handler(w http.ResponseWriter, r *http.Request)`

### Error: "CORS"
- Configurar `ALLOWED_ORIGIN` en variables de entorno de Vercel
- Para testing usar `ALLOWED_ORIGIN=*`
- Para producción usar tu dominio exacto

### Error: "SMTP"
- Verificar todas las variables SMTP en Vercel
- Usar contraseña de aplicación, no contraseña normal
- Verificar que no hay espacios extra en las variables

## 📞 **Soporte Urgente**

Si sigues teniendo problemas:

1. **Verifica logs en Vercel**: Dashboard → Functions → View Function Logs
2. **Verifica build logs**: Dashboard → Deployments → Click en deployment → View Build Logs
3. **Verifica variables**: Dashboard → Settings → Environment Variables

## 🎯 **Checklist Final**

- [ ] ✅ `npm run build` funciona sin errores
- [ ] ✅ Directorio `public/` se crea con 5 archivos
- [ ] ✅ `go vet ./...` no muestra errores
- [ ] ✅ Variables de entorno configuradas en Vercel
- [ ] ✅ Código pusheado a repositorio
- [ ] ✅ Redeploy ejecutado en Vercel
- [ ] ✅ `/api/health` responde correctamente
- [ ] ✅ Formulario de contacto funciona

---

**🚀 Con esta configuración, el deployment debería funcionar perfectamente en Vercel.**

**Softex Labs** - Transformando negocios con tecnología innovadora