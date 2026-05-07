import { ApiService } from './ApiService.js';

export class GroupsService {
    constructor() {
        this.apiService = new ApiService();
    }

    // async getAllGroups() {
    //     const response = await this.apiService.get('/api/groups', null);
    //     //const allGroups = [{ group_id: -1, group_name: "Не групированные задачи", group_desc: "", created_at: `${Date.now()}`,}];
    //     const result = await response.json();
    //     if(response.ok){
    //         // allGroups = allGroups.concat(result);
    //         // return allGroups;
    //         return result;
    //     }
        
    //     return await response.json();
    // }

    async editGroup(groupId, formData) {
        const response = await this.apiService.put(`/api/groups${groupId}`, formData);
        return await response;
    }
    
    async putTaskInGroup(taskId, formData) {
        const response = await this.apiService.put(`/api/tasks/${taskId}/group`, formData);
        return await response;
    }

    async createGroup(formData) {
        const response = await this.apiService.post(`/api/groups`, formData);
        return await response;
    }

    async getAllGroups() {
        return[
            {
                group_id: 1,
                group_name: "ИИ по ТП",
                group_desc: "ИИ по ТП",
            }
        ]
    }
}