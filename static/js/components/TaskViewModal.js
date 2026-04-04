import { TasksService } from '../services/TasksService.js';

import * as FormatService from '../services/FormatService.js';

export class TaskViewModal {
    constructor(taskForm) {
        this.taskForm = taskForm; // Для открытия формы редактирования
        this.tasksService = new TasksService();
        this.modalElement = null;
        this.bootstrapModal = null;
        this.currentTaskId = null;
    }
    
    initialize() {
        this.modalElement = document.getElementById('taskViewModal');
        
        if (!this.modalElement) {
            console.error('TaskViewModal: элемент не найден');
            return;
        }
        
        this.bootstrapModal = new window.bootstrap.Modal(this.modalElement, {
            backdrop: true,
            keyboard: true
        });
        
        this.bindEvents();
        console.log('TaskViewModal initialized successfully');
    }
    
    bindEvents() {
        // Кнопка редактировать
        document.getElementById('taskEditBtn').addEventListener('click', () => {
            this.bootstrapModal.hide();
            const taskId = this.currentTaskId;
            setTimeout(() => {
                this.taskForm.openEditMode(taskId);
            }, 300); // Небольшая задержка для плавности
        });
        
        // Кнопка удалить
        document.getElementById('taskDeleteBtn').addEventListener('click', () => {
            this.handleDelete();
        });
                
        // Кнопка завершить
        document.getElementById('taskEndBtn').addEventListener('click', () => {
            this.handleComplete();
        });

        // Кнопка взять задачу
        document.getElementById('taskClaimBtn').addEventListener('click', () => {
            this.handleClaim();
        });
        // Очистка при закрытии
        this.modalElement.addEventListener('hidden.bs.modal', () => {
            this.currentTaskId = null;
        });
    }
    
    // Открытие модалки с данными задачи
    async showTask(taskId) {
        this.currentTaskId = taskId;
        this.clearError();
        
        // Показываем загрузку
        this.setLoading(true);
        
        try {
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
            
            this.renderTask(task);
            this.bootstrapModal.show();
        } catch (error) {
            this.showError('Не удалось загрузить задачу: ' + error.message);
        } finally {
            this.setLoading(false);
        }
    }
    
    // Отрисовка данных задачи
    renderTask(task) {
        // ID
        document.getElementById('viewTaskId').textContent = task.id;
        
        // Наименование
        document.getElementById('viewTaskTitle').textContent = task.name;
        
        // Статус с бейджем
        const statusBadge = document.getElementById('viewTaskStatus');
        statusBadge.textContent = FormatService.getStatusText(task.status);
        statusBadge.className = `badge bg-${FormatService.getStatusColor(task.status)}`;
        
        // Приоритет с бейджем
        const priorityBadge = document.getElementById('viewTaskPriority');
        priorityBadge.textContent = FormatService.getPriorityText(task.priority);
        priorityBadge.className = `badge bg-${FormatService.getPriorityColor(task.priority)}`;
        
        // Описание (с сохранением переносов строк)
        document.getElementById('viewTaskDescription').textContent = task.description || 'Нет описания';
        
        // Автор
        document.getElementById('viewTaskAuthor').textContent = task.author;

        // Кто взял на исполнение
        document.getElementById('viewUserClaim').textContent = task.username;
        
        // Даты
        document.getElementById('viewTaskCreatedAt').textContent = FormatService.formatDate(task.created_at);
        document.getElementById('viewTaskUpdatedAt').textContent = FormatService.formatDate(task.updated_at);
    }
    
    // Удаление задачи
    async handleDelete() {
        if (!confirm('Вы уверены, что хотите удалить заявку #' + this.currentTaskId + '?')) {
            return;
        }
        
        try {
            await this.tasksService.deleteTask(this.currentTaskId);
            
            this.bootstrapModal.hide();
            
            // Показываем уведомление (можно заменить на toast)
            alert('Заявка успешно удалена');
            
            // Событие для обновления таблицы
            document.dispatchEvent(new CustomEvent('task:deleted'));
            
        } catch (error) {
            this.showError('Не удалось удалить задачу: ' + error.message);
        }
    }
    
    // Завершение задачи
    async handleComplete() {
        if (!confirm('Вы уверены, что хотите завершить заявку #' + this.currentTaskId + '?')) {
            return;
        }
        
        try {
            await this.tasksService.completeTask(this.currentTaskId);
            
            this.bootstrapModal.hide();
            
            alert('Заявка успешно завершена');
            
            // Событие для обновления таблицы
            document.dispatchEvent(new CustomEvent('task:saved'));
            
        } catch (error) {
            this.showError('Не удалось завершить задачу: ' + error.message);
        }
    }

    // Взять задачу
    async handleClaim() {
        try {

            await this.tasksService.claimTask(this.currentTaskId);
            
            this.bootstrapModal.hide();
            
            alert('Заявка успешно взята');
            
            // Событие для обновления таблицы
            document.dispatchEvent(new CustomEvent('task:saved'));
            
        } catch (error) {
            this.showError('Не удалось взять задачу: ' + error.message);
        }
    }

    showError(message) {
        document.getElementById('viewTaskErrorText').textContent = message;
        document.getElementById('viewTaskError').style.display = 'block';
        // Автоскрытие через 5 секунд
        setTimeout(() => {
            this.clearError();
        }, 5000);
    }
    
    clearError() {
        document.getElementById('viewTaskError').style.display = 'none';
    }
    
    setLoading(isLoading) {
        const modalBody = this.modalElement.querySelector('.modal-body');
        if (isLoading) {
            modalBody.style.opacity = '0.6';
            modalBody.style.pointerEvents = 'none';
        } else {
            modalBody.style.opacity = '1';
            modalBody.style.pointerEvents = 'auto';
        }
    }
}