
import { AuthModal } from './components/AuthModal.js';
import { TasksModal } from './components/TasksModal.js';
import { TaskViewModal } from './components/TaskViewModal.js';

import { AuthService } from './services/AuthService.js';


class Program {

    constructor() {
        this.isInitialized = false;
        this.authModal = null;

        this.authService = new AuthService();
        this.tasksModal = new TasksModal();
        this.taskViewModal = null;
    }

     async initialize() {
        try {
            
            // Пытаемся автоматически войти
            const autoLoginResult = this.authService.tryAutoLogin();
            if (autoLoginResult.success) {
                await this.onAuthSuccess(autoLoginResult.data);
            } else {
                console.log('Auto-login failed:', autoLoginResult.error);
            }

            this.initializeUI();
            this.isInitialized = true;
            console.log('App initialized successfully');
        } catch (error) {
            console.error('App initialization failed:', error);
        }
    }

    initializeUI() {
        this.authModal = new AuthModal(
            this.authService,
            (user) => this.onAuthSuccess(user)
        );
        this.authModal.initialize();
        this.tasksModal.initialize();
        this.taskViewModal = new TaskViewModal(this.tasksModal);
        this.taskViewModal.initialize();
        this.bindButtons();
        this.bindEvents();
    }

    //Привязываем действия к кнопкам 
    bindButtons() {
        document.getElementById('regLog').addEventListener('click', () => {
            this.authModal.show();
        });
        
        document.getElementById('quit').addEventListener('click', () => {
            this.logout();
        });

        document.getElementById('uppdateTaskBtn').addEventListener('click', () => {
            this.tasksModal.fetchData();
        });


    }
      
    bindEvents() {
        // Глобальный слушатель для открытия задачи
        document.addEventListener('task:view', (e) => {
            this.taskViewModal.showTask(e.detail.taskId);
        });

        // Обновление таблицы после удаления
        document.addEventListener('task:deleted', () => {
            this.taskTable.loadTasks();
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

    async updateUserInterface(user) {
        const usernameDisplay = document.getElementById('usernameDisplay');
        if (usernameDisplay) {
            usernameDisplay.value = user.username;
        }
        document.getElementById('quit').hidden = false;
        document.getElementById('regLog').hidden = true;

        await this.tasksModal.fetchData();
    }
}

const app = new Program();

// Запуск приложения
document.addEventListener('DOMContentLoaded', () => {
    app.initialize();
});