export class ApiService {
    constructor() {
    }
    
    async request(endpoint, options = {}, query = '') {
        const url = query == ''?  `${endpoint}`: `${endpoint}?${query}`;
        const config = {
            headers: {
                'Content-Type': 'application/json',
                ...options.headers
            },
            ...options

        };
        const token = this._getSessionToken();

        if (token) {
            config.headers['X-Session-Token'] = token;
        }

        if (config.body && typeof config.body === 'object') {
            config.body = JSON.stringify(config.body);
        }
        
        const response = await fetch(url, config);

        if (!response.ok) {
            throw new ApiError(response.status, response.statusText);
        }
        
        return await response;
    }
    
    async get(endpoint, data, query) {
        return await this.request(endpoint, {
            method: 'GET',
            body: data
        }, query);
    }
    
    async post(endpoint, data) {
        return await this.request(endpoint, {
            method: 'POST',
            body: data
        });
    }
    
    async put(endpoint, data) {
        return await this.request(endpoint, {
            method: 'PUT',
            body: data
        });
    }    
    
    async patch(endpoint, data) {
        return await this.request(endpoint, {
            method: 'PATCH',
            body: data
        });
    }
    
    async delete(endpoint) {
        return await this.request(endpoint, {
            method: 'DELETE'
        });
    }

    _getSessionToken() {
        return localStorage.getItem('token');
    }

}

class ApiError extends Error {
    constructor(status, statusText, responseText) {
        super(`API Error: ${status} ${statusText}`);
        this.name = 'ApiError';
        this.status = status;
        this.statusText = statusText;
        this.responseText = responseText;
        
        // Пытаемся распарсить JSON ошибку
        try {
            this.details = JSON.parse(responseText);
        } catch {
            this.details = { message: responseText };
        }
    }
}