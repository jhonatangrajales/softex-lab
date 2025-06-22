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

        validate() {
            // Resetea los errores individualmente para asegurar la reactividad en Alpine.js
            this.errors.name = '';
            this.errors.email = '';
            this.errors.message = '';

            let isValid = true;

            if (!this.formData.name.trim()) {
                this.errors.name = 'El nombre es obligatorio.';
                isValid = false;
            }

            const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
            if (!this.formData.email) {
                this.errors.email = 'El correo electrónico es obligatorio.';
                isValid = false;
            } else if (!emailRegex.test(this.formData.email)) {
                this.errors.email = 'Por favor, introduce un correo electrónico válido.';
                isValid = false;
            }

            if (this.formData.message.trim().length < 10) {
                this.errors.message = 'El mensaje debe tener al menos 10 caracteres.';
                isValid = false;
            }

            return isValid;
        },

        submitData() {
            this.successMessage = '';
            this.errorMessage = '';

            if (!this.validate()) {
                return; // Stop if validation fails
            }

            this.loading = true;
            fetch('/api/contact', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/x-www-form-urlencoded',
                },
                body: new URLSearchParams(this.formData)
            })
            .then(response => {
                if (!response.ok) {
                    // Intenta obtener el error del cuerpo JSON, si falla, usa un mensaje genérico.
                    return response.json()
                        .then(data => { throw new Error(data.error || `Error del servidor: ${response.status}`); })
                        .catch(() => { throw new Error(`Error del servidor: ${response.status}. La API no está disponible o no responde correctamente.`); });
                }
                return response.json();
            })
            .then(data => {
                this.successMessage = data.message || '¡Mensaje enviado con éxito! Puedes enviar otro si lo deseas.';
                this.formData.name = '';
                this.formData.email = '';
                this.formData.message = '';
                // Limpia los errores por si quedaba alguno
                this.validate();
                setTimeout(() => {
                    this.successMessage = '';
                }, 5000);
            })
            .catch((err) => {
                this.errorMessage = err.message;
            })
            .finally(() => {
                this.loading = false;
            });
        }
    }));
});

// All other logic should run after the DOM is ready.
document.addEventListener('DOMContentLoaded', () => {
    // Set current year in footer
    const yearSpan = document.getElementById('year');
    if (yearSpan) {
        yearSpan.textContent = new Date().getFullYear();
    }

    // Scroll reveal animation logic
    const scrollElements = document.querySelectorAll('.reveal-on-scroll');
    
    const elementInView = (el, dividend = 1) => {
        const elementTop = el.getBoundingClientRect().top;
        return (elementTop <= (window.innerHeight || document.documentElement.clientHeight) / dividend);
    };

    const handleScrollAnimation = () => {
        scrollElements.forEach((el) => {
            if (elementInView(el, 1.25)) {
                el.classList.add('is-visible');
            }
        });
    };

    // Initial check on load and add scroll listener
    handleScrollAnimation();
    window.addEventListener('scroll', handleScrollAnimation);
});