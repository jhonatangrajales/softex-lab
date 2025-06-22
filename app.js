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

        // Valida un campo específico y actualiza su mensaje de error
        validateField(field) {
            let isValid = true;
            this.errors[field] = ''; // Limpia el error previo para este campo

            switch (field) {
                case 'name':
                    if (!this.formData.name.trim()) {
                        this.errors.name = 'El nombre es obligatorio.';
                        isValid = false;
                    }
                    break;
                case 'email':
                    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
                    if (!this.formData.email) {
                        this.errors.email = 'El correo electrónico es obligatorio.';
                        isValid = false;
                    } else if (!emailRegex.test(this.formData.email)) {
                        this.errors.email = 'Por favor, introduce un correo electrónico válido.';
                        isValid = false;
                    }
                    break;
                case 'message':
                    if (this.formData.message.trim().length < 10) {
                        this.errors.message = 'El mensaje debe tener al menos 10 caracteres.';
                        isValid = false;
                    }
                    break;
            }
            return isValid; // Retorna si el campo es válido
        },

        // Valida todos los campos del formulario para el envío
        validate() {
            const isNameValid = this.validateField('name');
            const isEmailValid = this.validateField('email');
            const isMessageValid = this.validateField('message');
            return isNameValid && isEmailValid && isMessageValid;
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
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(this.formData)
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
                this.formData.message = ''; // Limpia el formulario
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

    // Scroll reveal animation logic with IntersectionObserver for performance
    const scrollElements = document.querySelectorAll('.reveal-on-scroll');
    
    const observer = new IntersectionObserver((entries) => {
        entries.forEach(entry => {
            // When the element is in view, add the 'is-visible' class
            if (entry.isIntersecting) {
                entry.target.classList.add('is-visible');
                // Optional: Stop observing the element once it's visible to save resources
                observer.unobserve(entry.target);
            }
        });
    }, { threshold: 0.1 }); // The callback will run when 10% of the target is visible

    scrollElements.forEach(el => {
        observer.observe(el);
    });
});