import { ApiService } from './ApiService.js';

export class GroupsService {
    constructor() {
        this.apiService = new ApiService();
    }

    async getAllGroups() {
        const response = await this.apiService.get('/api/groups', null);
        const allGroups = [{ group_id: -1, group_name: "Не групированные задачи", group_desc: "", created_at: `${Date.now()}`,}];
        const result = await response.json();
        if(response.ok){
            allGroups = allGroups.concat(result);
            return allGroups;
        }

        return await response.json();
    }
    
    async putTaskInGroup(taskId, formData) {
        const response = await this.apiService.put(`/api/tasks/${taskId}/group`, formData);
        return await response;
    }

    async createGroup(formData) {
        const response = await this.apiService.post(`/api/groups`, formData);
        return await response;
        // return new Response;
    }

    // async getAllGroups() {
    //     return [                
    //             {
    //                 group_id: -1,
    //                 group_name: "Не групированные задачи",
    //                 group_desc: "",
    //                 created_at: `${Date.now()}`,
    //             },
    //             {
    //                 group_id: 1,
    //                 group_name: "Первая группа",
    //                 group_desc: "ИИ по ТП",
    //                 created_at: "2026-04-21T06:42:08Z",
    //             },
    //             {
    //                 group_id: 2,
    //                 group_name: "Вторая группа",
    //                 group_desc: "ИИ по ТП",
    //                 created_at: "2026-04-21T06:42:08Z",
    //             }
    //         ];
    // }
}