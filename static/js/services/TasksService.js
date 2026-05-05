
import { ApiService } from './ApiService.js';

export class TasksService {
    constructor() {
        this.apiService = new ApiService();
    }

    async getAllTasks(status = '') {
        const response = await this.apiService.get('/api/tasks', null, status);
        
        let allTaskResult = [];
        if(response.ok)
        {
            allTaskResult = await response.json();

            if(document.getElementById('my_task').checked){
                const username = localStorage.getItem('username');
                allTaskResult = allTaskResult.filter(task => task.username == username);
            }
            return allTaskResult;
        }
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

    // async getTask(taskId) {
    //     const task = {
    //                 id: 1,
    //                 group_id: 1,
    //                 group_name: "ИИ по ТП",
    //                 group_desc: "ИИ по ТП",
    //                 name: "text",
    //                 description: "йцуйцуйуцйуйц",
    //                 author: "",
    //                 status: "open",
    //                 created_at: "2026-04-21T06:42:08Z",
    //                 updated_at: "2026-04-21T12:03:26Z",
    //                 completed_at: "0001-01-01T00:00:00Z"
    //             };
    //     return task;
            
    // }

    // async getAllTasks(status = '') {
    //     let allTaskResult = [];
    //     if(status == ''){
    //         allTaskResult = [
    //             {
    //                 id: 1,
    //                 group_id: 1,
    //                 group_name: "ИИ по ТП",
    //                 group_desc: "ИИ по ТП",
    //                 name: "text",
    //                 description: "йцуйцуйуцйуйц",
    //                 author: "",
    //                 status: "open",
    //                 created_at: "2026-04-21T06:42:08Z",
    //                 updated_at: "2026-04-21T12:03:26Z",
    //                 completed_at: "0001-01-01T00:00:00Z"
    //             },
    //             {
    //                 id: 2,
    //                 group_id: 1,
    //                 group_name: "ИИ по ТП",
    //                 group_desc: "ИИ по ТП",
    //                 name: "text",
    //                 description: "йцуйцуйуцйуйц",
    //                 author: "",
    //                 status: "in_progress",
    //                 created_at: "2026-04-21T06:42:08Z",
    //                 updated_at: "2026-04-21T12:03:26Z",
    //                 completed_at: "0001-01-01T00:00:00Z"
    //             },
    //             {
    //                 id: 3,
    //                 name: "text",
    //                 description: "йцуйцуйуцйуйц",
    //                 author: "",
    //                 status: "completed",
    //                 created_at: "2026-04-21T06:42:08Z",
    //                 updated_at: "2026-04-21T12:03:26Z",
    //                 completed_at: "0001-01-01T00:00:00Z"
    //             }
    //         ];
    //     }
    //     else if(status == 'status=open'){
    //             return [
    //             {
    //                 id: 1,
    //                 group_id: 1,
    //                 group_name: "ИИ по ТП",
    //                 group_desc: "ИИ по ТП",
    //                 name: "text",
    //                 description: "йцуйцуйуцйуйц",
    //                 author: "",
    //                 status: "open",
    //                 created_at: "2026-04-21T06:42:08Z",
    //                 updated_at: "2026-04-21T12:03:26Z",
    //                 completed_at: "0001-01-01T00:00:00Z"
    //             },
    //             {
    //                 id: 2,
    //                 group_id: 1,
    //                 group_name: "ИИ по ТП",
    //                 group_desc: "ИИ по ТП",
    //                 name: "text",
    //                 description: "йцуйцуйуцйуйц",
    //                 author: "",
    //                 status: "open",
    //                 created_at: "2026-04-21T06:42:08Z",
    //                 updated_at: "2026-04-21T12:03:26Z",
    //                 completed_at: "0001-01-01T00:00:00Z"
    //             },
    //             {
    //                 id: 3,
    //                 name: "text",
    //                 description: "йцуйцуйуцйуйц",
    //                 author: "",
    //                 status: "open",
    //                 created_at: "2026-04-21T06:42:08Z",
    //                 updated_at: "2026-04-21T12:03:26Z",
    //                 completed_at: "0001-01-01T00:00:00Z"
    //             }
    //         ];
    //     }
    //     else if(status == 'status=in_progress'){
    //         return[
    //             {
    //                 id: 3,
    //                 name: "text",
    //                 description: "йцуйцуйуцйуйц",
    //                 author: "",
    //                 status: "in_progress",
    //                 created_at: "2026-04-21T06:42:08Z",
    //                 updated_at: "2026-04-21T12:03:26Z",
    //                 completed_at: "0001-01-01T00:00:00Z"
    //             }
    //         ];
    //     }

    //     if(document.getElementById('my_task').checked){
    //         const username = localStorage.getItem('username');
    //         allTaskResult = allTaskResult.filter(task => task.username == username);
    //     }

    //     return allTaskResult;
    // }
}