/**
 * API Client Module for Trading Exchange Dashboard
 * Handles all REST API communications following SOLID principles
 */

// API Configuration
const API_CONFIG = {
    baseURL: 'http://localhost:8080/api',
    timeout: 10000,
    retryAttempts: 3,
    retryDelay: 1000
};

// API Endpoints
const ENDPOINTS = {
    METRICS: '/metrics',
    HEALTH: '/health',
    ORDERS: '/orders',
    ORDER_BY_ID: (id) => `/orders/${id}`
};

/**
 * HTTP Client class for making API requests
 */
class HTTPClient {
    constructor(config = API_CONFIG) {
        this.baseURL = config.baseURL;
        this.timeout = config.timeout;
        this.retryAttempts = config.retryAttempts;
        this.retryDelay = config.retryDelay;
    }

    /**
     * Make HTTP request with retry logic
     * @param {string} url - Request URL
     * @param {object} options - Fetch options
     * @returns {Promise<object>} Response data
     */
    async request(url, options = {}) {
        const fullUrl = url.startsWith('http') ? url : `${this.baseURL}${url}`;

        const defaultOptions = {
            headers: {
                'Content-Type': 'application/json',
                'Accept': 'application/json',
                ...options.headers
            },
            timeout: this.timeout,
            ...options
        };

        let lastError;

        for (let attempt = 0; attempt <= this.retryAttempts; attempt++) {
            try {
                const controller = new AbortController();
                const timeoutId = setTimeout(() => controller.abort(), this.timeout);

                const response = await fetch(fullUrl, {
                    ...defaultOptions,
                    signal: controller.signal
                });

                clearTimeout(timeoutId);

                if (!response.ok) {
                    throw new APIError(
                        `HTTP ${response.status}: ${response.statusText}`,
                        response.status,
                        fullUrl
                    );
                }

                const data = await response.json();
                return this.handleResponse(data);

            } catch (error) {
                lastError = error;

                if (attempt < this.retryAttempts && this.shouldRetry(error)) {
                    console.warn(`API request failed (attempt ${attempt + 1}/${this.retryAttempts + 1}):`, error.message);
                    await this.delay(this.retryDelay * Math.pow(2, attempt));
                    continue;
                }

                break;
            }
        }

        throw new APIError(
            `Request failed after ${this.retryAttempts + 1} attempts: ${lastError.message}`,
            lastError.status || 0,
            fullUrl,
            lastError
        );
    }

    /**
     * Handle API response format
     * @param {object} data - Response data
     * @returns {object} Processed response
     */
    handleResponse(data) {
        if (data.success === false) {
            throw new APIError(
                data.error?.message || 'API request failed',
                data.error?.code || 'UNKNOWN_ERROR'
            );
        }

        return data.data || data;
    }

    /**
     * Check if error should trigger a retry
     * @param {Error} error - The error that occurred
     * @returns {boolean} Whether to retry
     */
    shouldRetry(error) {
        if (error.name === 'AbortError') return false;
        if (error.status >= 400 && error.status < 500) return false;
        return true;
    }

    /**
     * Delay utility for retry logic
     * @param {number} ms - Milliseconds to wait
     * @returns {Promise} Promise that resolves after delay
     */
    delay(ms) {
        return new Promise(resolve => setTimeout(resolve, ms));
    }

    // HTTP method helpers
    async get(url, options = {}) {
        return this.request(url, { ...options, method: 'GET' });
    }

    async post(url, data, options = {}) {
        return this.request(url, {
            ...options,
            method: 'POST',
            body: JSON.stringify(data)
        });
    }

    async put(url, data, options = {}) {
        return this.request(url, {
            ...options,
            method: 'PUT',
            body: JSON.stringify(data)
        });
    }

    async delete(url, options = {}) {
        return this.request(url, { ...options, method: 'DELETE' });
    }
}

/**
 * Custom API Error class
 */
class APIError extends Error {
    constructor(message, status = 0, url = '', originalError = null) {
        super(message);
        this.name = 'APIError';
        this.status = status;
        this.url = url;
        this.originalError = originalError;
        this.timestamp = new Date().toISOString();
    }

    toJSON() {
        return {
            name: this.name,
            message: this.message,
            status: this.status,
            url: this.url,
            timestamp: this.timestamp
        };
    }
}

/**
 * Trading Exchange API Client
 * Provides high-level methods for interacting with the trading exchange API
 */
class TradingExchangeAPI {
    constructor(httpClient = new HTTPClient()) {
        this.http = httpClient;
        this.eventListeners = new Map();
    }

    /**
     * Get real-time metrics data
     * @returns {Promise<object>} Metrics data
     */
    async getMetrics() {
        try {
            const data = await this.http.get(ENDPOINTS.METRICS);
            this.emit('metricsUpdated', data);
            return data;
        } catch (error) {
            this.emit('error', { type: 'METRICS_FETCH_ERROR', error });
            throw error;
        }
    }

    /**
     * Get system health status
     * @returns {Promise<object>} Health data
     */
    async getHealth() {
        try {
            const data = await this.http.get(ENDPOINTS.HEALTH);
            this.emit('healthUpdated', data);
            return data;
        } catch (error) {
            this.emit('error', { type: 'HEALTH_FETCH_ERROR', error });
            throw error;
        }
    }

    /**
     * Place a new order
     * @param {object} orderData - Order details
     * @returns {Promise<object>} Order response
     */
    async placeOrder(orderData) {
        try {
            const data = await this.http.post(ENDPOINTS.ORDERS, orderData);
            this.emit('orderPlaced', data);
            return data;
        } catch (error) {
            this.emit('error', { type: 'ORDER_PLACE_ERROR', error });
            throw error;
        }
    }

    /**
     * Get order by ID
     * @param {string} orderId - Order ID
     * @returns {Promise<object>} Order data
     */
    async getOrder(orderId) {
        try {
            const data = await this.http.get(ENDPOINTS.ORDER_BY_ID(orderId));
            this.emit('orderFetched', data);
            return data;
        } catch (error) {
            this.emit('error', { type: 'ORDER_FETCH_ERROR', error });
            throw error;
        }
    }

    /**
     * Cancel an order
     * @param {string} orderId - Order ID
     * @returns {Promise<object>} Cancellation response
     */
    async cancelOrder(orderId) {
        try {
            const data = await this.http.delete(ENDPOINTS.ORDER_BY_ID(orderId));
            this.emit('orderCancelled', data);
            return data;
        } catch (error) {
            this.emit('error', { type: 'ORDER_CANCEL_ERROR', error });
            throw error;
        }
    }

    /**
     * Batch fetch multiple data sources
     * @returns {Promise<object>} Combined data
     */
    async fetchDashboardData() {
        try {
            const [metrics, health] = await Promise.allSettled([
                this.getMetrics(),
                this.getHealth()
            ]);

            const result = {
                timestamp: new Date().toISOString(),
                metrics: metrics.status === 'fulfilled' ? metrics.value : null,
                health: health.status === 'fulfilled' ? health.value : null,
                errors: []
            };

            if (metrics.status === 'rejected') {
                result.errors.push({ type: 'METRICS_ERROR', error: metrics.reason });
            }

            if (health.status === 'rejected') {
                result.errors.push({ type: 'HEALTH_ERROR', error: health.reason });
            }

            this.emit('dashboardDataFetched', result);
            return result;

        } catch (error) {
            this.emit('error', { type: 'DASHBOARD_FETCH_ERROR', error });
            throw error;
        }
    }

    /**
     * Event system for API client
     */
    on(event, callback) {
        if (!this.eventListeners.has(event)) {
            this.eventListeners.set(event, []);
        }
        this.eventListeners.get(event).push(callback);
    }

    off(event, callback) {
        if (!this.eventListeners.has(event)) return;
        const listeners = this.eventListeners.get(event);
        const index = listeners.indexOf(callback);
        if (index > -1) {
            listeners.splice(index, 1);
        }
    }

    emit(event, data) {
        if (!this.eventListeners.has(event)) return;
        const listeners = this.eventListeners.get(event);
        listeners.forEach(callback => {
            try {
                callback(data);
            } catch (error) {
                console.error(`Error in event listener for ${event}:`, error);
            }
        });
    }

    /**
     * Connection status checker
     */
    async checkConnection() {
        try {
            await this.getHealth();
            this.emit('connectionStatusChanged', { connected: true });
            return true;
        } catch (error) {
            this.emit('connectionStatusChanged', { connected: false, error });
            return false;
        }
    }
}

/**
 * Real-time data manager using polling
 * (In a real application, you might use WebSockets)
 */
class RealTimeDataManager {
    constructor(api, pollInterval = 5000) {
        this.api = api;
        this.pollInterval = pollInterval;
        this.isPolling = false;
        this.pollTimeoutId = null;
        this.connectionCheckInterval = 30000; // Check connection every 30 seconds
        this.connectionCheckTimeoutId = null;
    }

    /**
     * Start real-time polling
     */
    start() {
        if (this.isPolling) return;

        this.isPolling = true;
        console.log('Starting real-time data polling...');

        // Initial fetch
        this.fetchData();

        // Start polling
        this.scheduleNextPoll();

        // Start connection checking
        this.startConnectionCheck();
    }

    /**
     * Stop real-time polling
     */
    stop() {
        if (!this.isPolling) return;

        this.isPolling = false;
        console.log('Stopping real-time data polling...');

        if (this.pollTimeoutId) {
            clearTimeout(this.pollTimeoutId);
            this.pollTimeoutId = null;
        }

        if (this.connectionCheckTimeoutId) {
            clearTimeout(this.connectionCheckTimeoutId);
            this.connectionCheckTimeoutId = null;
        }
    }

    /**
     * Fetch data and schedule next poll
     */
    async fetchData() {
        if (!this.isPolling) return;

        try {
            await this.api.fetchDashboardData();
        } catch (error) {
            console.error('Error fetching dashboard data:', error);
        }

        this.scheduleNextPoll();
    }

    /**
     * Schedule next poll
     */
    scheduleNextPoll() {
        if (!this.isPolling) return;

        this.pollTimeoutId = setTimeout(() => {
            this.fetchData();
        }, this.pollInterval);
    }

    /**
     * Start connection checking
     */
    startConnectionCheck() {
        const checkConnection = async () => {
            if (!this.isPolling) return;

            await this.api.checkConnection();

            this.connectionCheckTimeoutId = setTimeout(checkConnection, this.connectionCheckInterval);
        };

        checkConnection();
    }

    /**
     * Update poll interval
     * @param {number} interval - New interval in milliseconds
     */
    updateInterval(interval) {
        this.pollInterval = interval;

        if (this.isPolling) {
            this.stop();
            this.start();
        }
    }
}

// Export classes for use in other modules
export {
    HTTPClient,
    APIError,
    TradingExchangeAPI,
    RealTimeDataManager,
    API_CONFIG,
    ENDPOINTS
};