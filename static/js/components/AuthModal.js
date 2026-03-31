


export class AuthModal {
    constructor(authService) {
        this.authService = authService;
        this.modalElement = null;
        this.bootstrapModal = null;
        this.currentTab = 'login';
        this.isInitialized = false;
    }
    
    initialize() {
        if (this.isInitialized) return;
        
        this.modalElement = document.getElementById('authModal');
        if (!this.modalElement) {
            console.error('Auth modal element not found!');
            return;
        }

        // Проверяем, что Bootstrap загрузился
        if (typeof bootstrap === 'undefined') {
            console.error('Bootstrap не загружен! Проверь подключение bootstrap.bundle.min.js');
            return;
        }

        // Инициализируем Bootstrap modal
        this.bootstrapModal = new bootstrap.Modal(this.modalElement, {
            backdrop: true, // Затемнение фона
            keyboard: true, // Закрытие по ESC
            focus: true
        });
        
        this.bindEvents();
        this.setupTabs();
        this.isInitialized = true;
        
        console.log('AuthModal загружен');
    }
    
    bindEvents() {
        // Обработчики форм
        document.getElementById('loginForm').addEventListener('submit', (e) => {
            e.preventDefault();
            this.handleLogin();
        });
        
        document.getElementById('registryForm').addEventListener('submit', (e) => {
            e.preventDefault();
            this.handleRegister();
        });
        
        // Переключение видимости пароля
        document.querySelectorAll('.password-toggle').forEach(toggle => {
            toggle.addEventListener('click', (e) => this.togglePasswordVisibility(e));
        });
        
        // Обработчик закрытия модального окна
        this.modalElement.addEventListener('hidden.bs.modal', () => {
            this.clearForms();
            this.cleanupBackdrops(); // Чистим бэкдропы
        });

        // Обработчик кнопки закрытия
        const closeButton = this.modalElement.querySelector('.btn-close');
        if (closeButton) {
            closeButton.addEventListener('click', () => {
                this.hide();
            });
        }
        // Закрытие по клику на backdrop
        this.modalElement.addEventListener('click', (e) => {
            if (e.target === this.modalElement) {
                this.hide();
            }
        });
    }
    
    setupTabs() {
        // Bootstrap tabs уже настроены в HTML, просто отслеживаем активную вкладку
        const tabElements = this.modalElement.querySelectorAll('[data-bs-toggle="tab"]');
        tabElements.forEach(tab => {
            tab.addEventListener('shown.bs.tab', (event) => {
                this.currentTab = event.target.getAttribute('id') === 'login-tab' ? 'login' : 'register';
                this.clearAlerts();
            });
        });
    }
    
    async handleLogin() {
        const username = document.getElementById('usernameLogin').value;
        const password = document.getElementById('passwordLogin').value;
        
        this.showLoading('login', true);
        this.clearAlerts();
        
        const result = await this.authService.login(username, password);
        
        this.showLoading('login', false);
        
        if (result.success) {
            this.showSuccess('login', 'Успешный вход!');
            setTimeout(() => {
                this.hideModal();
                this.onAuthSuccess(result.data);
            }, 1000);
        } else {
            this.showError('login', result.error);
        }
    }
    
    async handleRegister() {
        const username = document.getElementById('usernameRegistry').value;
        const password = document.getElementById('passwordRegistry').value;
        const confirmPassword = document.getElementById('passwordRegistryConfirm').value;
        
        this.showLoading('register', true);
        this.clearAlerts();
        
        const result = await this.authService.register(username, password, confirmPassword);
        
        this.showLoading('register', false);
        
        if (result.success) {
            this.showSuccess('register', 'Регистрация успешна!');
            setTimeout(() => {
                this.hideModal();
                this.onAuthSuccess(result.data);
            }, 1000);
        } else {
            this.showError('register', result.error);
        }
    }
    
    
    show() {
        if (!this.bootstrapModal) {
            console.error('Bootstrap modal not initialized');
            return;
        }
        
        this.bootstrapModal.show();
    }
    
    hide() {
        if (!this.bootstrapModal) {
            console.error('Bootstrap modal not initialized');
            return;
        }
        
        this.bootstrapModal.hide();
        this.cleanupBackdrops(); // Дополнительная очистка
    }
    
    //Очистка бэкдропов
    cleanupBackdrops() {
        // Удаляем все лишние backdrop элементы
        const backdrops = document.querySelectorAll('.modal-backdrop');
        backdrops.forEach(backdrop => {
            backdrop.remove();
        });
        
        // Убираем классы с body
        document.body.classList.remove('modal-open');
        document.body.style.overflow = '';
        document.body.style.paddingRight = '';
    }
    
    togglePasswordVisibility(event) {
        const targetId = event.currentTarget.getAttribute('data-target');
        const passwordInput = document.getElementById(targetId);
        const icon = event.currentTarget.querySelector('i');
        
        if (passwordInput.type === 'password') {
            passwordInput.type = 'text';
            icon.classList.replace('fa-eye', 'fa-eye-slash');
        } else {
            passwordInput.type = 'password';
            icon.classList.replace('fa-eye-slash', 'fa-eye');
        }
    }
    
    showLoading(formType, isLoading) {
        const button = document.querySelector(`#${formType}Form button[type="submit"]`);
        const originalText = button.innerHTML;
        
        if (isLoading) {
            button.disabled = true;
            button.innerHTML = '<i class="fas fa-spinner fa-spin"></i> Загрузка...';
            button.setAttribute('data-original-text', originalText);
        } else {
            button.disabled = false;
            button.innerHTML = button.getAttribute('data-original-text') || originalText;
        }
    }
    
    showError(formType, message) {
        const alertElement = document.getElementById(`${formType}ErrorAlert`);
        const alertText = document.getElementById(`${formType}ErrorAlertText`);
        
        if (alertElement && alertText) {
            alertText.textContent = message;
            alertElement.style.display = 'block';
            
            // Автоскрытие через 5 секунд
            setTimeout(() => {
                alertElement.style.display = 'none';
            }, 5000);
        }
    }
    
    showSuccess(formType, message) {
        const alertElement = document.getElementById(`${formType}SuccessAlert`);
        const alertText = document.getElementById(`${formType}SuccessAlertText`);
        
        if (alertElement && alertText) {
            alertText.textContent = message;
            alertElement.style.display = 'block';
        }
    }
    
    clearAlerts() {
        // Скрываем все алерты
        const alerts = this.modalElement.querySelectorAll('.alert');
        alerts.forEach(alert => {
            alert.style.display = 'none';
        });
    }
    
    clearForms() {
        // Очищаем все поля форм
        const inputs = this.modalElement.querySelectorAll('input');
        inputs.forEach(input => {
            input.value = '';
        });
        this.clearAlerts();
    }
    
    hideModal() {
        this.hide();
        this.clearForms();
    }
}