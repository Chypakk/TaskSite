
import { AuthModal } from './components/AuthModal.js';
import { AuthService } from './services/AuthService.js';

class Program {

    constructor() {
        this.isInitialized = false;
        this.authModal = null;
        this.authService = new AuthService();
    }

     async initialize() {
            this.initializeUI();
            this.isInitialized = true;
            console.log('App initialized successfully');
    }

    initializeUI() {
        this.authModal = new AuthModal(
            this.authService,
            (user) => this.onAuthSuccess(user)
        );
        this.authModal.initialize();
        this.bindAuthButtons();
    }

    //Привязываем 
    bindAuthButtons() {
        document.getElementById('regLog').addEventListener('click', () => {
            this.authModal.show();
        });
        
        document.getElementById('quit').addEventListener('click', () => {
            this.logout();
        });
    }

    logout() {
        const result = this.authService.logout();
        if (result.success) {
            document.getElementById('usernameDisplay').value = '';
            document.getElementById('quit').hidden = true;
            document.getElementById('regLog').hidden = false;
            console.log('User logged out');
        }
    }

    async onAuthSuccess(user) {
        try {
            console.log('User authenticated:', user.username);
            this.updateUserInterface(user);
        } catch (error) {
            console.error('Error handling auth success:', error);
        }
    }

    updateUserInterface(user) {
        const usernameDisplay = document.getElementById('usernameDisplay');
        if (usernameDisplay) {
            usernameDisplay.value = user.username;
        }
        
        document.getElementById('quit').hidden = false;
        document.getElementById('regLog').hidden = true;
    }
}

const app = new Program();

// Запуск приложения
document.addEventListener('DOMContentLoaded', () => {
    app.initialize();
});