
import { ApiService } from './ApiService.js';

export class TasksService {
    constructor() {
        this.apiService = new ApiService();
    }

    async getAllTasks(status = '') {
        const response = await this.apiService.get('/api/tasks', null, status);
        return await response.json();
    }

    async getTask(taskId) {
        const response = await this.apiService.get(`/api/tasks/${taskId}`, null);
        return await response.json();
    }

    async createTask(data) {
        const response = await this.apiService.post('/api/tasks', data);
        return await response;
    }

    async completeTask(taskId) {
        const response = await this.apiService.post(`/api/tasks/${taskId}/complete`);;
        return await response;
    }
    
    async claimTask(taskId) {
        const response = await this.apiService.post(`/api/tasks/${taskId}/claim`);
        return await response;
    }

    async updateTask(taskId, formData) {
        const response = await this.apiService.put(`/api/tasks/${taskId}`, formData);
        return await response;
    } 

    async deleteTask(taskId) {
        const response = await this.apiService.delete(`/api/tasks/${taskId}`);;
        return await response;
    }
    
    // async getAllTasks(status = '') {
    //     if(status == ''){
    //         return [
    //             {
    //                 id: 1,
    //                 name: "Пров1",
    //                 author: "Иван",
    //                 status: "open",
    //                 username: "Test",
    //                 created_at: "02.04.2026",
    //                 updated_at: "",
    //                 completed_at: "",
    //             },
    //             {
    //                 id: 2,
    //                 name: "Пров2",
    //                 author: "Иван",
    //                 status: "open",
    //                 username: "Test",
    //                 created_at: "02.04.2026",
    //                 updated_at: "",
    //                 completed_at: "",
    //             },
    //             {
    //                 id: 3,
    //                 name: "Пров3",
    //                 author: "Иван",
    //                 status: "in_progress",
    //                 username: "Test",
    //                 created_at: "02.04.2026",
    //                 updated_at: "",
    //                 completed_at: "",
    //             }
    //         ];
    //     }
    //     else if(status == 'status=open'){
    //                     return [
    //             {
    //                 id: 1,
    //                 name: "Пров1",
    //                 author: "Иван",
    //                 status: "open",
    //                 username: "Test",
    //                 created_at: "02.04.2026",
    //                 updated_at: "",
    //                 completed_at: "",
    //             },
    //             {
    //                 id: 2,
    //                 name: "Пров2",
    //                 author: "Иван",
    //                 status: "open",
    //                 username: "Test",
    //                 created_at: "02.04.2026",
    //                 updated_at: "",
    //                 completed_at: "",
    //             }
    //         ];
    //     }
    //     else if(status == 'status=in_progress'){
    //         return[
    //             {
    //                 id: 3,
    //                 name: "Пров3",
    //                 author: "Иван",
    //                 status: "in_progress",
    //                 username: "Test",
    //                 created_at: "02.04.2026",
    //                 updated_at: "",
    //                 completed_at: "",
    //             }
    //         ];
    //     }
    // }
}