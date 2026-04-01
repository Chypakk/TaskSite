import { User } from '../models/User.js';
import { AuthResult } from '../models/AuthResult.js';
import { ApiService } from './ApiService.js';

export class AuthService {
    constructor() {
        this.apiService = new ApiService();;
        this.currentUser = new User();
    }
    
    async login(username, password) {
        try {

            const validationError = this.validateCredentials(username, password);
            if (validationError) {
                return AuthResult.failure(validationError);
            }

            const response = await this.apiService.post('/api/login', {
                username, 
                password 
            });
            
            const userData = await response.json();

            this.currentUser = new User(userData);
            this.saveToStorage();

            return AuthResult.success(this.currentUser);
            
        } catch (error) {
            console.error('Login error:', error);
            this.clearStorage();
            return AuthResult.failure(this.handleApiError(error));
        }
    }
    
    async register(username, password, confirmPassword) {
        try {
            const validationError = this.validateRegistration(username, password, confirmPassword);
            if (validationError) {
                return AuthResult.failure(validationError);
            }

            const response = await this.apiService.post('/api/register', {
                username, 
                password 
            });

            const userData = await response.json();
            this.currentUser = new User(userData);
            this.saveToStorage();
            
            return AuthResult.success(this.currentUser);
            
        } catch (error) {
            console.error('Registration error:', error);
            this.clearStorage();
            return AuthResult.failure(this.handleApiError(error));
        }
    }

    async logout() {
        const token = this.apiService._getSessionToken();
        const result = await this.apiService.post('/api/logout', token);
        if (result.ok){
            this.currentUser = new User();
            this.clearStorage();
            return AuthResult.success(null);
        }
        else{
            this.currentUser = new User();
            this.clearStorage();
            return AuthResult.failure('Logout failed');
        }
    }
    
    async tryAutoLogin() {
        try {
            const token = this.apiService._getSessionToken();
            if (token) {
                const response = await this.apiService.post('/api/me', token);
                if (response.ok){
                    const userData = await response.json();
                    this.currentUser = new User(userData);

                    return AuthResult.success(this.currentUser);
                }
                this.clearStorage();
                return AuthResult.failure('Invalid stored data');
            }
            return AuthResult.failure('No stored session');
        } catch (error) {
            this.clearStorage();
            return AuthResult.failure('Invalid stored data');
        }
    }
    

    handleApiError(error) {
        if (error.name === 'ApiError') {
            // Обрабатываем структурированную ошибку от ApiService
            return error.details.message || error.details.title || error.statusText;
        }
        
        // Сетевые ошибки и пр.
        if (error.message.includes('Failed to fetch')) {
            return 'Сетевая ошибка. Проверьте подключение к интернету.';
        }
        
        return 'Неизвестная ошибка';
    }

    validateCredentials(username, password) {
        if (!username || username.trim().length < 3) {
            return 'Имя пользователя должно быть не менее 3 символов';
        }
        if (!password || password.length < 4) {
            return 'Пароль должен быть не менее 4 символов';
        }
        return null;
    }
    
    validateRegistration(username, password, confirmPassword) {
        const credentialError = this.validateCredentials(username, password);
        if (credentialError) return credentialError;
        
        if (password !== confirmPassword) {
            return 'Пароли не совпадают';
        }
        
        if (username.length > 20) {
            return 'Имя пользователя слишком длинное';
        }
        
        return null;
    }
    
    saveToStorage() {
        localStorage.setItem('token', this.currentUser.token);
        localStorage.setItem('username', this.currentUser.username);
    }
    
    clearStorage() {
        localStorage.removeItem('token');
        localStorage.removeItem('username');
    }
}