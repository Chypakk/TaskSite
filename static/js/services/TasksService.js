
import { ApiService } from './ApiService.js';

export class TasksService {
    constructor() {
        this.apiService = new ApiService();
    }
    async getAllTasks() {
        const response = await this.apiService.get('/api/tasks', null);
        return await response.json();
    }

    async createTask(data) {
        const response = await this.apiService.post('/api/tasks', data);
        return await response.json();
    }
    
    async getTask(taskId) {
        const response = await this.apiService.get(`/api/tasks/${taskId}`, null);
        return await response.json();
    }

    async updateTask(taskId, formData) {
        const response = await this.apiService.put(`/api/tasks/${taskId}`, formData);
        return await response;
    }
}