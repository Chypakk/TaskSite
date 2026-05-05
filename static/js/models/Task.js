export class Task {
    constructor(data = {}) {
        this.id = data.id || 0;
        this.name = data.name || '';
        this.group_id = data.group_id || 0;
        this.group_name = data.group_name || '';
        this.group_desc = data.group_desc || '';
        this.description = data.description || '';
        this.author = data.author || '';
        this.status = data.status || '';
        this.created_at = data.created_at || '';
        this.completed_at = data.completed_at || '';
        this.username = data.username || '';
    }
}