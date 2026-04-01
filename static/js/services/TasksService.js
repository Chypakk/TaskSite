
import { ApiService } from './ApiService.js';

export class TasksService {
    constructor() {
        this.apiService = new ApiService();
    }
    async getAllTasks() {
        //await new Promise(resolve => setTimeout(resolve, 800));

        const response = await this.apiService.get('/api/tasks', null);
        return await response.json();
    }

    async createTask(data) {
        const response = await this.apiService.post('/api/tasks', data);
        return await response.json();
    }
    
    async getTask(taskId) {
        const response = await this.apiService.get('/api/task', taskId);
        return await response.json();
    }


}