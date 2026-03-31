
import { AuthModal } from './components/AuthModal.js';


class Program {

    constructor() {
        this.isInitialized = false;
        this.authModal = null;
    }

     async initialize() {
            this.initializeUI();
            this.isInitialized = true;
            console.log('App initialized successfully');
    }

    initializeUI() {
        this.authModal = new AuthModal(
            this.authService
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
        // const result = this.authService.logout();
        // if (result.success) {
            document.getElementById('usernameDisplay').value = '';
            document.getElementById('quit').hidden = true;
            document.getElementById('regLog').hidden = false;
            this.updateBalanceDisplay();
            this.hideAllInteriors();
            console.log('User logged out');
        // }
    }
}

const app = new Program();

// Запуск приложения
document.addEventListener('DOMContentLoaded', () => {
    app.initialize();
});