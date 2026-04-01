import { ApiService } from '../services/ApiService.js';

export class TaskViewModal {
    constructor(taskForm) {
        this.apiService = new ApiService();
        this.taskForm = taskForm; // Для открытия формы редактирования
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
            setTimeout(() => {
                this.taskForm.openEditMode(this.currentTaskId);
            }, 300); // Небольшая задержка для плавности
        });
        
        // Кнопка удалить
        document.getElementById('taskDeleteBtn').addEventListener('click', () => {
            this.handleDelete();
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
            const task = await this.apiService.get(`/api/tasks/${taskId}`);
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
        document.getElementById('viewTaskTitle').textContent = task.title;
        
        // Статус с бейджем
        const statusBadge = document.getElementById('viewTaskStatus');
        statusBadge.textContent = this.getStatusText(task.status);
        statusBadge.className = `badge bg-${this.getStatusColor(task.status)}`;
        
        // Приоритет с бейджем
        const priorityBadge = document.getElementById('viewTaskPriority');
        priorityBadge.textContent = this.getPriorityText(task.priority);
        priorityBadge.className = `badge bg-${this.getPriorityColor(task.priority)}`;
        
        // Описание (с сохранением переносов строк)
        document.getElementById('viewTaskDescription').textContent = task.description || 'Нет описания';
        
        // Автор
        document.getElementById('viewTaskAuthor').textContent = task.author;
        
        // Даты
        document.getElementById('viewTaskCreatedAt').textContent = this.formatDate(task.createdAt);
        document.getElementById('viewTaskUpdatedAt').textContent = task.updatedAt ? 
            this.formatDate(task.updatedAt) : '—';
    }
    
    // Удаление задачи
    async handleDelete() {
        if (!confirm('Вы уверены, что хотите удалить заявку #' + this.currentTaskId + '?')) {
            return;
        }
        
        try {
            await this.apiService.delete(`/api/tasks/${this.currentTaskId}`);
            
            this.bootstrapModal.hide();
            
            // Показываем уведомление (можно заменить на toast)
            alert('Заявка успешно удалена');
            
            // Событие для обновления таблицы
            document.dispatchEvent(new CustomEvent('task:deleted'));
            
        } catch (error) {
            this.showError('Не удалось удалить задачу: ' + error.message);
        }
    }
    
    // Вспомогательные методы
    getStatusText(status) {
        const statuses = {
            'open': 'Открыта',
            'in_progress': 'В работе',
            'resolved': 'Завершена',
            'closed': 'Закрыта'
        };
        return statuses[status] || status;
    }
    
    getStatusColor(status) {
        const colors = {
            'open': 'warning',
            'in_progress': 'info',
            'resolved': 'success',
            'closed': 'secondary'
        };
        return colors[status] || 'secondary';
    }
    
    getPriorityText(priority) {
        const priorities = {
            'low': 'Низкий',
            'medium': 'Средний',
            'high': 'Высокий',
            'critical': 'Критический'
        };
        return priorities[priority] || priority;
    }
    
    getPriorityColor(priority) {
        const colors = {
            'low': 'secondary',
            'medium': 'primary',
            'high': 'warning',
            'critical': 'danger'
        };
        return colors[priority] || 'secondary';
    }
    
    formatDate(dateString) {
        if (!dateString) return '—';
        const date = new Date(dateString);
        return date.toLocaleDateString('ru-RU', {
            day: '2-digit',
            month: '2-digit',
            year: 'numeric',
            hour: '2-digit',
            minute: '2-digit'
        });
    }
    
    showError(message) {
        document.getElementById('viewTaskErrorText').textContent = message;
        document.getElementById('viewTaskError').style.display = 'block';
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