import { TasksService } from '../services/TasksService.js';
import * as FormatService from '../services/FormatService.js';

export class TasksModal{

    constructor() {
        this.isInitialized = false;
        this.form = null;
        this.bootstrapModal = null;
        this.isEditMode = false;
        this.tasksService = new TasksService();
        this.isFetching = false;
        
    }

    initialize() {
        if (this.isInitialized) return;
        
        this.modalElement = document.getElementById('createTaskModal');
        this.form = document.getElementById('taskForm');


        if (!this.modalElement|| !this.form) {
            console.error('Tasks modal element not found!');
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
        this.isInitialized = true;

        console.log('TasksModal initialized successfully');
    }

    bindEvents() {
        // Обработчики форм
        this.form.addEventListener('submit', (e) => this.handleSubmit(e));
        
        // Очистка при закрытии
        this.modalElement.addEventListener('hidden.bs.modal', () => {
            this.clearForm();
        });

        // Кнопка создания извне
        const createBtn = document.getElementById('createTaskBtn');
        if (createBtn) {
            createBtn.addEventListener('click', () => this.openCreateMode());
        }

        // В классе TaskTable добавь слушатель:
        document.addEventListener('task:saved', () => {
            this.fetchData(); // Перезагружаем таблицу
        });
    }

    // Открытие в режиме создания
    openCreateMode() {
        this.isEditMode = false;
        this.clearForm();
    
        // Заголовок
        document.getElementById('taskModalTitle').innerHTML = 
            '<i class="fas fa-file-alt me-2"></i>Новая заявка';
        
        // Статус по умолчанию
        const statusAttr = document.getElementById('taskStatus');
        statusAttr.value = 'open';
        statusAttr.disabled = true;
        
        this.bootstrapModal.show();
    }

    // Открытие в режиме редактирования
    async openEditMode(taskId) {
        this.isEditMode = true;
        this.clearForm();
        
        // Показываем загрузку
        this.setLoading(true);
        
        try {
            // Загружаем данные задачи
            const task = await this.tasksService.getTask(taskId);
            // const task =             {
            //     id: 1,
            //     name: "1234",
            //     author: "Иван",
            //     status: "closed",
            //     description: "Очень тестовая задача",
            //     username: "Test",
            //     created_at: "2026-04-02T05:05:05Z",
            //     updated_at: "0001-01-01T00:00:00Z",
            //     completed_at: "",
            // };

            // Заполняем форму
            document.getElementById('taskId').value = task.id;
            document.getElementById('taskName').value = task.name;
            document.getElementById('taskDescription').value = task.description;
            document.getElementById('taskAuthor').value = task.author;

            const statusAttr = document.getElementById('taskStatus');
            statusAttr.value = task.status;
            statusAttr.disabled = false;
            
            // Заголовок
            document.getElementById('taskModalTitle').innerHTML = 
                `<i class="fas fa-edit me-2"></i>Редактирование #${task.id}`;
            
            this.bootstrapModal.show();
        } catch (error) {
            this.showError('Не удалось загрузить данные задачи');
        } finally {
            this.setLoading(false);
        }
    }


    async handleSubmit(e) {
        e.preventDefault();
        
        // Валидация Bootstrap
        if (!this.form.checkValidity()) {
            this.form.classList.add('was-validated');
            return;
        }
        
        this.setLoading(true);
        this.clearAlerts();
        
        // Собираем данные
        const formData = {
            name: document.getElementById('taskName').value.trim(),
            description: document.getElementById('taskDescription').value.trim(),
            author: document.getElementById('taskAuthor').value.trim(),
            status: document.getElementById('taskStatus').value
        };
        
        // Валидация длины заголовка
        if (formData.name.length < 3) {
            this.showError('Наименование должно быть не менее 3 символов');
            this.setLoading(false);
            return;
        }
        
        try {
            // Определяем метод и URL
            if (this.isEditMode) {
                const taskId = document.getElementById('taskId').value;
                const response = await this.tasksService.updateTask(taskId,formData);
                if(response.ok){
                    this.showSuccess('Заявка обновлена!');
                    //console.log('Task update successfully');
                }
                else{
                    this.showError(error.message || 'Ошибка при сохранении');
                    this.setLoading(false);
                    return;
                }
            } else {
                await this.tasksService.createTask(formData);
                this.showSuccess('Заявка создана!');
            }
            
            // Закрываем через 1 секунду
            setTimeout(() => {
                this.bootstrapModal.hide();
                // Событие для обновления таблицы
                document.dispatchEvent(new CustomEvent('task:saved'));
            }, 1000);
            
        } catch (error) {
            this.showError(error.message || 'Ошибка при сохранении');
        } finally {
            this.setLoading(false);
        }
    }

    // Утилиты
    clearForm() {
        this.form.reset();
        this.form.classList.remove('was-validated');
        document.getElementById('taskId').value = '';
        this.clearAlerts();
    }

    clearAlerts() {
        document.getElementById('taskErrorAlert').style.display = 'none';
        document.getElementById('taskSuccessAlert').style.display = 'none';
    }

    showError(message) {
        document.getElementById('taskErrorText').textContent = message;
        document.getElementById('taskErrorAlert').style.display = 'block';
        document.getElementById('taskSuccessAlert').style.display = 'none';
        // Автоскрытие через 5 секунд
        setTimeout(() => {
            this.clearAlerts();
        }, 5000);
    }
    
    showSuccess(message) {
        document.getElementById('taskSuccessText').textContent = message;
        document.getElementById('taskSuccessAlert').style.display = 'block';
        document.getElementById('taskErrorAlert').style.display = 'none';
    }
    
    setLoading(isLoading) {
        const btn = document.getElementById('taskSaveBtn');
        if (isLoading) {
            btn.disabled = true;
            btn.innerHTML = '<i class="fas fa-spinner fa-spin me-2"></i>Сохранение...';
        } else {
            btn.disabled = false;
            btn.innerHTML = '<i class="fas fa-save me-2"></i>Сохранить';
        }
    }

    async fetchData(showLoading = true) {
        if (this.isFetching) return;
        this.isFetching = true;
            
        try {
            if (showLoading) {
                document.getElementById('tasksTableBody').innerHTML = `
                    <tr><td colspan="16" style="text-align: center; padding: 20px;">
                        <div class="loading"></div> Обновление данных...
                    </td></tr>
                `;
            }
            
            const data = await this.tasksService.getAllTasks();
            this.renderTable(data);

        } catch (error) {
            console.error('Ошибка загрузки данных:', error);
            document.getElementById('tasksTableBody').innerHTML = `
                <tr><td colspan="16" style="text-align: center; padding: 20px; color: #ef4444;">
                    ❌ Ошибка загрузки данных. Проверьте подключение к сети.
                </td></tr>
            `;
        } finally {
            this.isFetching = false;
        }
    }


    renderTable(data) {
        const tbody = document.getElementById('tasksTableBody');
        tbody.innerHTML = '';
        
        if (data.length === 0) {
            tbody.innerHTML = `<tr><td colspan="16" style="text-align: center; padding: 20px;">Нет данных</td></tr>`;
            return;
        }
        
        data.forEach(item => {
            const row = document.createElement('tr');
            row.setAttribute('data-task-id', item.id);
            // Формирование ячеек с применением классов для стилизации
            row.innerHTML = `
                <td>${item.id}</td>
                <td>${item.name}</td>
                <td>${item.author}</td>
                <td class="highlight">${FormatService.getStatusText(item.status)}</td>
                <td>${item.username}</td>
                <td>${FormatService.formatDate(item.created_at)}</td>
                <td>${FormatService.formatDate(item.updated_at)}</td>
                <td>${FormatService.formatDate(item.completed_at)}</td>
            `;
            // Клик по строке
            row.addEventListener('click', (e) => {
                // Игнорируем клик по кнопкам действий (если они есть)
                if (e.target.closest('.btn')) return;
                
                // Генерируем событие или сразу открываем
                document.dispatchEvent(new CustomEvent('task:view', { 
                    detail: { taskId: item.id } 
                }));
            });
            
            tbody.appendChild(row);
        });
    }

}

