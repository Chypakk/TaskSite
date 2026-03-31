export class AuthResult {
    constructor(success, data, error = '') {
        this.success = success;
        this.data = data;
        this.error = error;
        this.timestamp = new Date();
    }
    
    static success(data) {
        return new AuthResult(true, data);
    }
    
    static failure(error) {
        return new AuthResult(false, null, error);
    }
}
