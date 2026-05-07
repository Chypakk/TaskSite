
import { AuthModal } from './components/AuthModal.js';
import { TasksModal } from './components/TasksModal.js';
import { TaskViewModal } from './components/TaskViewModal.js';
import { GroupModal } from './components/GroupModal.js';

import { TasksTable } from './components/TaskTable.js';
import { TaskWheel } from './components/TaskWheel.js';

import { AuthService } from './services/AuthService.js';


class Program {

    constructor() {
        this.isInitialized = false;
        this.authModal = null;

        this.authService = new AuthService();
        this.tasksTable= new TasksTable();
        this.tasksModal = new TasksModal();
        this.groupModal = new GroupModal();
        this.taskWheel = new TaskWheel();
        
        this.taskViewModal = null;
    }

    async initialize() {
        try {
            // Пытаемся автоматически войти
            const autoLoginResult =  await this.authService.tryAutoLogin();
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

    async initializeUI() {
        this.authModal = new AuthModal(
            this.authService,
            (user) => this.onAuthSuccess(user)
        );
        this.sidebarToggleInit();
        this.authModal.initialize();
        this.tasksTable.initialize();
        this.tasksModal.initialize();
        this.groupModal.initialize();
        this.taskViewModal = new TaskViewModal(this.tasksModal,this.groupModal );
        this.taskViewModal.initialize();

        this.taskWheel.initialize();

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
            const filtr = document.getElementById('statusFilter');
            const statusText = filtr.value == '' ? '': `status=${filtr.value}`;
            this.tasksTable.fetchData(true, statusText);
        });

        document.getElementById('randomTaskBtn').addEventListener('click', async () => {
            this.taskWheel.open();
        });

        document.getElementById("my_task").addEventListener('change', async () => {
            const filtr = document.getElementById('statusFilter');
            const statusText = filtr.value == '' ? '': `status=${filtr.value}`;
            this.tasksTable.fetchData(true, statusText);
        });
    }

    bindEvents() {
        document.getElementById('statusFilter').addEventListener('change',(e)=>{
            this.tasksTable.fetchData(true, e.target.value != ''? `status=${e.target.value}`: '');
        });
    }

    async fillGroupFiltr(){
        await this.groupModal.uppdateGroupCash();

        const selectGroup = document.getElementById('groupFilter');
        selectGroup.innerHTML = '<option value="" disabled selected>Выберите группу</option>'
            + this.groupModal.groupsCash.map(g => `<option value="${g.group_id}">${g.group_name}</option>`).join('');

    }


    sidebarToggleInit() {
        const sidebar = document.getElementById('sidebar');
        const sidebarToggle = document.getElementById('sidebarToggle');
        
        if (sidebarToggle && sidebar) {
            sidebarToggle.addEventListener('click', function() {
                sidebar.classList.toggle('collapsed');
                
                // Сохраняем состояние в localStorage (опционально)
                const isCollapsed = sidebar.classList.contains('collapsed');
                localStorage.setItem('sidebarCollapsed', isCollapsed);
            });
            
            // Восстанавливаем состояние при загрузке
            const isCollapsed = localStorage.getItem('sidebarCollapsed') === 'true';
            if (isCollapsed) {
                sidebar.classList.add('collapsed');
            }
        }
    }

    async logout() {
        const result = await this.authService.logout();
        if (result.success) {
            document.getElementById('usernameDisplay').value = '';
            document.getElementById('quit').hidden = true;
            document.getElementById('regLog').hidden = false;
            console.log('User logged out');
            await this.tasksTable.fetchData();
        }
    }

    async onAuthSuccess(user) {
        try {
            console.log('User authenticated:', user.username);
            await this.updateUserInterface(user);
            await this.fillGroupFiltr();
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

        await this.tasksTable.fetchData();
    }
}

const app = new Program();

// Запуск приложения
document.addEventListener('DOMContentLoaded', () => {
    app.initialize();
});

