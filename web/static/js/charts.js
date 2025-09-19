/**
 * Charts Module for Trading Exchange Dashboard
 * Handles Chart.js integration and visualization following SOLID principles
 */

// Chart configuration constants
const CHART_COLORS = {
    primary: '#00C851',
    secondary: '#007bff',
    accent: '#6f42c1',
    warning: '#ffc107',
    danger: '#dc3545',
    info: '#17a2b8',
    background: 'rgba(0, 200, 81, 0.1)',
    grid: '#333333',
    text: '#b3b3b3'
};

const CHART_DEFAULTS = {
    responsive: true,
    maintainAspectRatio: false,
    interaction: {
        intersect: false,
        mode: 'index'
    },
    plugins: {
        legend: {
            labels: {
                color: CHART_COLORS.text,
                font: {
                    family: 'Inter, sans-serif',
                    size: 12
                }
            }
        },
        tooltip: {
            backgroundColor: 'rgba(26, 26, 26, 0.95)',
            titleColor: '#ffffff',
            bodyColor: '#b3b3b3',
            borderColor: '#333333',
            borderWidth: 1,
            cornerRadius: 8,
            titleFont: {
                family: 'Inter, sans-serif',
                weight: '600'
            },
            bodyFont: {
                family: 'Inter, sans-serif'
            }
        }
    },
    scales: {
        x: {
            grid: {
                color: CHART_COLORS.grid,
                drawBorder: false
            },
            ticks: {
                color: CHART_COLORS.text,
                font: {
                    family: 'Inter, sans-serif',
                    size: 11
                }
            }
        },
        y: {
            grid: {
                color: CHART_COLORS.grid,
                drawBorder: false
            },
            ticks: {
                color: CHART_COLORS.text,
                font: {
                    family: 'Inter, sans-serif',
                    size: 11
                }
            }
        }
    }
};

/**
 * Base Chart class implementing common chart functionality
 */
class BaseChart {
    constructor(canvasId, type, options = {}) {
        this.canvasId = canvasId;
        this.type = type;
        this.canvas = document.getElementById(canvasId);
        this.ctx = this.canvas?.getContext('2d');
        this.chart = null;
        this.data = [];
        this.options = this.mergeOptions(options);

        if (!this.canvas) {
            throw new Error(`Canvas element with id '${canvasId}' not found`);
        }

        this.init();
    }

    /**
     * Merge default options with custom options
     */
    mergeOptions(customOptions) {
        return {
            ...CHART_DEFAULTS,
            ...customOptions,
            plugins: {
                ...CHART_DEFAULTS.plugins,
                ...customOptions.plugins
            },
            scales: {
                ...CHART_DEFAULTS.scales,
                ...customOptions.scales
            }
        };
    }

    /**
     * Initialize the chart
     */
    init() {
        this.chart = new Chart(this.ctx, {
            type: this.type,
            data: this.getInitialData(),
            options: this.options
        });
    }

    /**
     * Get initial chart data structure
     */
    getInitialData() {
        return {
            labels: [],
            datasets: []
        };
    }

    /**
     * Update chart data
     */
    updateData(newData) {
        if (!this.chart) return;

        this.data = newData;
        this.chart.data = this.processData(newData);
        this.chart.update('none');
    }

    /**
     * Process raw data for chart consumption
     */
    processData(data) {
        // Override in subclasses
        return data;
    }

    /**
     * Destroy the chart
     */
    destroy() {
        if (this.chart) {
            this.chart.destroy();
            this.chart = null;
        }
    }

    /**
     * Resize the chart
     */
    resize() {
        if (this.chart) {
            this.chart.resize();
        }
    }

    /**
     * Show loading state
     */
    showLoading() {
        if (!this.chart) return;

        this.chart.data.labels = ['Loading...'];
        this.chart.data.datasets = [{
            label: 'Loading',
            data: [0],
            backgroundColor: CHART_COLORS.background,
            borderColor: CHART_COLORS.grid
        }];
        this.chart.update();
    }

    /**
     * Show error state
     */
    showError(message = 'Error loading data') {
        if (!this.chart) return;

        this.chart.data.labels = [message];
        this.chart.data.datasets = [{
            label: 'Error',
            data: [0],
            backgroundColor: 'rgba(220, 53, 69, 0.1)',
            borderColor: CHART_COLORS.danger
        }];
        this.chart.update();
    }
}

/**
 * Volume Trends Chart
 */
class VolumeChart extends BaseChart {
    constructor(canvasId) {
        const options = {
            plugins: {
                legend: {
                    display: false
                },
                tooltip: {
                    callbacks: {
                        label: function(context) {
                            return `Volume: $${context.parsed.y.toLocaleString()}`;
                        }
                    }
                }
            },
            scales: {
                y: {
                    ...CHART_DEFAULTS.scales.y,
                    beginAtZero: true,
                    ticks: {
                        ...CHART_DEFAULTS.scales.y.ticks,
                        callback: function(value) {
                            return '$' + value.toLocaleString();
                        }
                    }
                }
            }
        };

        super(canvasId, 'line', options);
        this.maxDataPoints = 20;
    }

    getInitialData() {
        return {
            labels: [],
            datasets: [{
                label: 'Volume',
                data: [],
                borderColor: CHART_COLORS.primary,
                backgroundColor: CHART_COLORS.background,
                fill: true,
                tension: 0.4,
                borderWidth: 2,
                pointBackgroundColor: CHART_COLORS.primary,
                pointBorderColor: '#ffffff',
                pointBorderWidth: 2,
                pointRadius: 4,
                pointHoverRadius: 6
            }]
        };
    }

    processData(metricsData) {
        if (!metricsData || typeof metricsData.total_volume === 'undefined') {
            return this.getInitialData();
        }

        // Add new data point
        const timestamp = new Date().toLocaleTimeString([], {
            hour: '2-digit',
            minute: '2-digit'
        });

        // Maintain rolling window of data
        if (this.chart.data.labels.length >= this.maxDataPoints) {
            this.chart.data.labels.shift();
            this.chart.data.datasets[0].data.shift();
        }

        this.chart.data.labels.push(timestamp);
        this.chart.data.datasets[0].data.push(metricsData.total_volume || 0);

        return this.chart.data;
    }

    updatePeriod(period) {
        // In a real application, this would fetch historical data for the period
        console.log(`Updating volume chart for period: ${period}`);

        // For demo purposes, adjust the number of data points shown
        switch (period) {
            case '1h':
                this.maxDataPoints = 20;
                break;
            case '1d':
                this.maxDataPoints = 48;
                break;
            case '1w':
                this.maxDataPoints = 168;
                break;
        }
    }
}

/**
 * Performance Chart (Orders and Trades per second)
 */
class PerformanceChart extends BaseChart {
    constructor(canvasId) {
        const options = {
            plugins: {
                legend: {
                    display: true,
                    position: 'top'
                },
                tooltip: {
                    callbacks: {
                        label: function(context) {
                            return `${context.dataset.label}: ${context.parsed.y.toFixed(2)}/sec`;
                        }
                    }
                }
            },
            scales: {
                y: {
                    ...CHART_DEFAULTS.scales.y,
                    beginAtZero: true,
                    ticks: {
                        ...CHART_DEFAULTS.scales.y.ticks,
                        callback: function(value) {
                            return value.toFixed(1) + '/s';
                        }
                    }
                }
            }
        };

        super(canvasId, 'line', options);
        this.maxDataPoints = 20;
    }

    getInitialData() {
        return {
            labels: [],
            datasets: [
                {
                    label: 'Orders/sec',
                    data: [],
                    borderColor: CHART_COLORS.primary,
                    backgroundColor: 'rgba(0, 200, 81, 0.1)',
                    tension: 0.4,
                    borderWidth: 2,
                    pointRadius: 3,
                    pointHoverRadius: 5
                },
                {
                    label: 'Trades/sec',
                    data: [],
                    borderColor: CHART_COLORS.secondary,
                    backgroundColor: 'rgba(0, 123, 255, 0.1)',
                    tension: 0.4,
                    borderWidth: 2,
                    pointRadius: 3,
                    pointHoverRadius: 5
                }
            ]
        };
    }

    processData(metricsData) {
        if (!metricsData) {
            return this.getInitialData();
        }

        // Add new data point
        const timestamp = new Date().toLocaleTimeString([], {
            hour: '2-digit',
            minute: '2-digit'
        });

        // Maintain rolling window of data
        if (this.chart.data.labels.length >= this.maxDataPoints) {
            this.chart.data.labels.shift();
            this.chart.data.datasets[0].data.shift();
            this.chart.data.datasets[1].data.shift();
        }

        this.chart.data.labels.push(timestamp);
        this.chart.data.datasets[0].data.push(metricsData.orders_per_sec || 0);
        this.chart.data.datasets[1].data.push(metricsData.trades_per_sec || 0);

        return this.chart.data;
    }
}

/**
 * Symbol Distribution Chart (could be used for pie/doughnut charts)
 */
class SymbolDistributionChart extends BaseChart {
    constructor(canvasId) {
        const options = {
            plugins: {
                legend: {
                    display: true,
                    position: 'bottom',
                    labels: {
                        padding: 20,
                        usePointStyle: true
                    }
                },
                tooltip: {
                    callbacks: {
                        label: function(context) {
                            const percentage = ((context.parsed / context.dataset.data.reduce((a, b) => a + b, 0)) * 100).toFixed(1);
                            return `${context.label}: ${percentage}% ($${context.parsed.toLocaleString()})`;
                        }
                    }
                }
            }
        };

        super(canvasId, 'doughnut', options);
    }

    getInitialData() {
        return {
            labels: ['No Data'],
            datasets: [{
                data: [1],
                backgroundColor: [CHART_COLORS.grid],
                borderColor: [CHART_COLORS.text],
                borderWidth: 1
            }]
        };
    }

    processData(symbolMetrics) {
        if (!symbolMetrics || Object.keys(symbolMetrics).length === 0) {
            return this.getInitialData();
        }

        const symbols = Object.keys(symbolMetrics);
        const volumes = symbols.map(symbol => symbolMetrics[symbol].volume || 0);

        // Generate colors for each symbol
        const colors = this.generateColors(symbols.length);

        return {
            labels: symbols,
            datasets: [{
                data: volumes,
                backgroundColor: colors.background,
                borderColor: colors.border,
                borderWidth: 2,
                hoverOffset: 4
            }]
        };
    }

    generateColors(count) {
        const baseColors = [
            CHART_COLORS.primary,
            CHART_COLORS.secondary,
            CHART_COLORS.accent,
            CHART_COLORS.warning,
            CHART_COLORS.info
        ];

        const background = [];
        const border = [];

        for (let i = 0; i < count; i++) {
            const color = baseColors[i % baseColors.length];
            background.push(color + '40'); // Add alpha
            border.push(color);
        }

        return { background, border };
    }
}

/**
 * Chart Manager - Coordinates all charts
 */
class ChartManager {
    constructor() {
        this.charts = new Map();
        this.isInitialized = false;
    }

    /**
     * Initialize all charts
     */
    init() {
        try {
            // Initialize volume chart
            this.charts.set('volume', new VolumeChart('volumeChart'));

            // Initialize performance chart
            this.charts.set('performance', new PerformanceChart('performanceChart'));

            this.isInitialized = true;
            console.log('Chart manager initialized successfully');

            // Set up chart period controls
            this.setupPeriodControls();

        } catch (error) {
            console.error('Error initializing charts:', error);
        }
    }

    /**
     * Set up period control buttons
     */
    setupPeriodControls() {
        const periodButtons = document.querySelectorAll('.chart-btn[data-period]');

        periodButtons.forEach(button => {
            button.addEventListener('click', (e) => {
                // Remove active class from all buttons
                periodButtons.forEach(btn => btn.classList.remove('active'));

                // Add active class to clicked button
                e.target.classList.add('active');

                // Update chart period
                const period = e.target.dataset.period;
                this.updatePeriod(period);
            });
        });
    }

    /**
     * Update all charts with new data
     */
    updateCharts(data) {
        if (!this.isInitialized || !data) return;

        try {
            // Update volume chart
            const volumeChart = this.charts.get('volume');
            if (volumeChart && data.metrics) {
                volumeChart.updateData(data.metrics);
            }

            // Update performance chart
            const performanceChart = this.charts.get('performance');
            if (performanceChart && data.metrics) {
                performanceChart.updateData(data.metrics);
            }

        } catch (error) {
            console.error('Error updating charts:', error);
        }
    }

    /**
     * Show loading state on all charts
     */
    showLoading() {
        this.charts.forEach(chart => {
            try {
                chart.showLoading();
            } catch (error) {
                console.error('Error showing loading state:', error);
            }
        });
    }

    /**
     * Show error state on all charts
     */
    showError(message) {
        this.charts.forEach(chart => {
            try {
                chart.showError(message);
            } catch (error) {
                console.error('Error showing error state:', error);
            }
        });
    }

    /**
     * Update chart period
     */
    updatePeriod(period) {
        const volumeChart = this.charts.get('volume');
        if (volumeChart && typeof volumeChart.updatePeriod === 'function') {
            volumeChart.updatePeriod(period);
        }
    }

    /**
     * Resize all charts
     */
    resize() {
        this.charts.forEach(chart => {
            try {
                chart.resize();
            } catch (error) {
                console.error('Error resizing chart:', error);
            }
        });
    }

    /**
     * Destroy all charts
     */
    destroy() {
        this.charts.forEach(chart => {
            try {
                chart.destroy();
            } catch (error) {
                console.error('Error destroying chart:', error);
            }
        });

        this.charts.clear();
        this.isInitialized = false;
    }

    /**
     * Get chart instance
     */
    getChart(name) {
        return this.charts.get(name);
    }
}

// Handle window resize for responsive charts
let resizeTimeout;
window.addEventListener('resize', () => {
    clearTimeout(resizeTimeout);
    resizeTimeout = setTimeout(() => {
        if (window.chartManager) {
            window.chartManager.resize();
        }
    }, 250);
});

// Export classes for use in other modules
export {
    BaseChart,
    VolumeChart,
    PerformanceChart,
    SymbolDistributionChart,
    ChartManager,
    CHART_COLORS,
    CHART_DEFAULTS
};