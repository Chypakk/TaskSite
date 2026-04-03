export class Task {
    constructor(data = {}) {
        this.id = data.id || 0;
        this.name = data.name || '';
        this.description = data.description || '';
        this.author = data.author || '';
        this.status = data.status || '';
        this.created_at = data.created_at || '';
        this.completed_at = data.completed_at || '';
        this.username = data.username || '';
    }
}
