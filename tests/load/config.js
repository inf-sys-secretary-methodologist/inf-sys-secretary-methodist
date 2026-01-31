// k6 Load Testing Configuration
// Docs: https://k6.io/docs/

export const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
export const API_URL = `${BASE_URL}/api`;

// Test user credentials (create in test DB before running)
export const TEST_USER = {
    email: __ENV.TEST_USER_EMAIL || 'loadtest@example.com',
    password: __ENV.TEST_USER_PASSWORD || 'LoadTest123!',
};

// Default thresholds for all tests
export const DEFAULT_THRESHOLDS = {
    http_req_duration: ['p(95)<500', 'p(99)<1000'], // 95% < 500ms, 99% < 1000ms
    http_req_failed: ['rate<0.01'],                  // Error rate < 1%
    http_reqs: ['rate>100'],                         // Throughput > 100 req/s
};

// Scenarios for different load patterns
export const SCENARIOS = {
    // Smoke test - minimal load to verify system works
    smoke: {
        executor: 'constant-vus',
        vus: 1,
        duration: '30s',
    },

    // Load test - normal expected load
    load: {
        executor: 'ramping-vus',
        startVUs: 0,
        stages: [
            { duration: '1m', target: 50 },   // Ramp up to 50 users
            { duration: '3m', target: 50 },   // Stay at 50 users
            { duration: '1m', target: 0 },    // Ramp down
        ],
    },

    // Stress test - find breaking point
    stress: {
        executor: 'ramping-vus',
        startVUs: 0,
        stages: [
            { duration: '2m', target: 100 },  // Ramp up
            { duration: '5m', target: 100 },  // Stay at peak
            { duration: '2m', target: 200 },  // Push beyond normal
            { duration: '5m', target: 200 },  // Stay at stress
            { duration: '2m', target: 0 },    // Ramp down
        ],
    },

    // Spike test - sudden traffic burst
    spike: {
        executor: 'ramping-vus',
        startVUs: 0,
        stages: [
            { duration: '10s', target: 100 }, // Sudden spike
            { duration: '1m', target: 100 },  // Stay at peak
            { duration: '10s', target: 0 },   // Sudden drop
        ],
    },

    // Soak test - sustained load over time
    soak: {
        executor: 'constant-vus',
        vus: 50,
        duration: '30m',
    },
};

// Helper function to get scenario by name
export function getScenario(name) {
    return SCENARIOS[name] || SCENARIOS.smoke;
}
