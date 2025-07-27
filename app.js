document.addEventListener('alpine:init', () => {
    Alpine.data('contactForm', () => ({
        formData: {
            name: '',
            email: '',
            message: ''
        },
        errors: {
            name: '',
            email: '',
            message: ''
        },
        loading: false,
        successMessage: '',
        errorMessage: '',
        showSuccess: false,
        showError: false,
        charCounts: {
            name: 0,
            message: 0
        },
        maxLengths: {
            name: 100,
            email: 254,
            message: 2000
        },

        // Inicialización
        init() {
            this.updateCharCounts();
        },

        // Actualizar contadores de caracteres
        updateCharCounts() {
            this.charCounts.name = this.formData.name.length;
            this.charCounts.message = this.formData.message.length;
        },

        // Validación mejorada con más reglas
        validateField(field) {
            let isValid = true;
            this.errors[field] = '';

            switch (field) {
                case 'name':
                    const name = this.formData.name.trim();
                    if (!name) {
                        this.errors.name = 'El nombre es obligatorio.';
                        isValid = false;
                    } else if (name.length > this.maxLengths.name) {
                        this.errors.name = `El nombre no puede exceder ${this.maxLengths.name} caracteres.`;
                        isValid = false;
                    } else if (!/^[a-zA-ZáéíóúÁÉÍÓÚñÑ\s]+$/.test(name)) {
                        this.errors.name = 'El nombre solo puede contener letras y espacios.';
                        isValid = false;
                    }
                    break;

                case 'email':
                    const email = this.formData.email.trim();
                    if (!email) {
                        this.errors.email = 'El correo electrónico es obligatorio.';
                        isValid = false;
                    } else if (email.length > this.maxLengths.email) {
                        this.errors.email = `El correo no puede exceder ${this.maxLengths.email} caracteres.`;
                        isValid = false;
                    } else if (!/^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/.test(email)) {
                        this.errors.email = 'Por favor, introduce un correo electrónico válido.';
                        isValid = false;
                    }
                    break;

                case 'message':
                    const message = this.formData.message.trim();
                    if (!message) {
                        this.errors.message = 'El mensaje es obligatorio.';
                        isValid = false;
                    } else if (message.length < 10) {
                        this.errors.message = 'El mensaje debe tener al menos 10 caracteres.';
                        isValid = false;
                    } else if (message.length > this.maxLengths.message) {
                        this.errors.message = `El mensaje no puede exceder ${this.maxLengths.message} caracteres.`;
                        isValid = false;
                    }
                    break;
            }
            return isValid;
        },

        // Validación en tiempo real
        onInput(field) {
            this.updateCharCounts();
            // Validar solo si el campo ya tiene contenido o ya se mostró un error
            if (this.formData[field] || this.errors[field]) {
                this.validateField(field);
            }
        },

        // Validar todos los campos
        validate() {
            const isNameValid = this.validateField('name');
            const isEmailValid = this.validateField('email');
            const isMessageValid = this.validateField('message');
            return isNameValid && isEmailValid && isMessageValid;
        },

        // Obtener clase CSS para el campo
        getFieldClass(field) {
            if (this.errors[field]) return 'form-input error';
            if (this.formData[field] && !this.errors[field]) return 'form-input success';
            return 'form-input';
        },

        // Obtener clase para contador de caracteres
        getCharCounterClass(field) {
            const count = this.charCounts[field];
            const max = this.maxLengths[field];
            const percentage = (count / max) * 100;
            
            if (percentage >= 100) return 'char-counter error';
            if (percentage >= 80) return 'char-counter warning';
            return 'char-counter';
        },

        // Cerrar alertas
        closeSuccess() {
            this.showSuccess = false;
            this.successMessage = '';
        },

        closeError() {
            this.showError = false;
            this.errorMessage = '';
        },

        // Envío del formulario con mejoras
        async submitData() {
            this.closeSuccess();
            this.closeError();

            if (!this.validate()) {
                this.showError = true;
                this.errorMessage = 'Por favor, corrige los errores en el formulario.';
                return;
            }

            this.loading = true;

            try {
                const response = await fetch('/api/contact', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify(this.formData)
                });

                const data = await response.json();

                if (!response.ok) {
                    throw new Error(data.error || `Error del servidor: ${response.status}`);
                }

                this.showSuccess = true;
                this.successMessage = data.message || '¡Mensaje enviado con éxito! Te responderemos pronto.';
                
                // Limpiar formulario
                this.formData = { name: '', email: '', message: '' };
                this.errors = { name: '', email: '', message: '' };
                this.updateCharCounts();
                
                // Scroll al mensaje de éxito
                setTimeout(() => {
                    document.querySelector('.alert-success')?.scrollIntoView({ 
                        behavior: 'smooth', 
                        block: 'center' 
                    });
                }, 100);

            } catch (error) {
                this.showError = true;
                this.errorMessage = error.message || 'No se pudo enviar el mensaje. Inténtalo de nuevo más tarde.';
                
                // Scroll al mensaje de error
                setTimeout(() => {
                    document.querySelector('.alert-error')?.scrollIntoView({ 
                        behavior: 'smooth', 
                        block: 'center' 
                    });
                }, 100);
            } finally {
                this.loading = false;
            }
        }
    }));

    // Componente para navegación móvil mejorada
    Alpine.data('navigation', () => ({
        mobileMenuOpen: false,
        
        toggleMobileMenu() {
            this.mobileMenuOpen = !this.mobileMenuOpen;
            // Prevenir scroll del body cuando el menú está abierto
            document.body.style.overflow = this.mobileMenuOpen ? 'hidden' : '';
        },

        closeMobileMenu() {
            this.mobileMenuOpen = false;
            document.body.style.overflow = '';
        },

        // Cerrar menú al hacer clic en un enlace
        handleLinkClick() {
            this.closeMobileMenu();
        }
    }));
});

// Lógica que se ejecuta después de que el DOM esté listo
document.addEventListener('DOMContentLoaded', () => {
    // Establecer año actual en el footer
    const yearSpan = document.getElementById('year');
    if (yearSpan) {
        yearSpan.textContent = new Date().getFullYear();
    }

    // Animaciones de scroll mejoradas con IntersectionObserver
    const scrollElements = document.querySelectorAll('.reveal-on-scroll');
    
    const observerOptions = {
        threshold: 0.1,
        rootMargin: '0px 0px -50px 0px'
    };

    const observer = new IntersectionObserver((entries) => {
        entries.forEach(entry => {
            if (entry.isIntersecting) {
                entry.target.classList.add('is-visible');
                observer.unobserve(entry.target);
            }
        });
    }, observerOptions);

    scrollElements.forEach(el => {
        observer.observe(el);
    });

    // Smooth scroll para enlaces internos
    document.querySelectorAll('a[href^="#"]').forEach(anchor => {
        anchor.addEventListener('click', function (e) {
            e.preventDefault();
            const target = document.querySelector(this.getAttribute('href'));
            if (target) {
                target.scrollIntoView({
                    behavior: 'smooth',
                    block: 'start'
                });
            }
        });
    });

    // Lazy loading para imágenes (si las hay)
    const images = document.querySelectorAll('img[data-src]');
    const imageObserver = new IntersectionObserver((entries) => {
        entries.forEach(entry => {
            if (entry.isIntersecting) {
                const img = entry.target;
                img.src = img.dataset.src;
                img.removeAttribute('data-src');
                imageObserver.unobserve(img);
            }
        });
    });

    images.forEach(img => imageObserver.observe(img));

    // Mejorar accesibilidad del teclado
    document.addEventListener('keydown', (e) => {
        // Cerrar menú móvil con Escape
        if (e.key === 'Escape') {
            const mobileMenu = document.querySelector('[x-data*="navigation"]');
            if (mobileMenu && Alpine.store) {
                Alpine.evaluate(mobileMenu, 'closeMobileMenu()');
            }
        }
    });

    // Detectar preferencias de movimiento reducido
    const prefersReducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)');
    if (prefersReducedMotion.matches) {
        document.documentElement.style.setProperty('--animation-duration', '0s');
    }

    // Performance: Preload critical resources
    const preloadLink = document.createElement('link');
    preloadLink.rel = 'preload';
    preloadLink.href = '/api/contact';
    preloadLink.as = 'fetch';
    preloadLink.crossOrigin = 'anonymous';
    document.head.appendChild(preloadLink);
});

// Service Worker para cache (opcional)
if ('serviceWorker' in navigator) {
    window.addEventListener('load', () => {
        navigator.serviceWorker.register('/sw.js')
            .then(registration => {
                console.log('SW registrado: ', registration);
            })
            .catch(registrationError => {
                console.log('SW falló: ', registrationError);
            });
    });
}

// Utilidades globales
window.SoftexUtils = {
    // Debounce para optimizar eventos
    debounce(func, wait) {
        let timeout;
        return function executedFunction(...args) {
            const later = () => {
                clearTimeout(timeout);
                func(...args);
            };
            clearTimeout(timeout);
            timeout = setTimeout(later, wait);
        };
    },

    // Formatear números
    formatNumber(num) {
        return new Intl.NumberFormat('es-ES').format(num);
    },

    // Copiar al portapapeles
    async copyToClipboard(text) {
        try {
            await navigator.clipboard.writeText(text);
            return true;
        } catch (err) {
            console.error('Error al copiar: ', err);
            return false;
        }
    }
};