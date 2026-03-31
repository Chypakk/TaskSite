export class User {
    constructor(data = {}) {
        this.id = data.id || 0;
        this.username = data.username || '';
        this.token = data.token || '';

    }
    
    get isAuthenticated() {
        return this.id > 0;
    }
}