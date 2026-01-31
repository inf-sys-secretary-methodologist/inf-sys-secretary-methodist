/**
 * k6 Load Test - Main API Endpoints
 *
 * Run: k6 run tests/load/api-test.js
 * With env: k6 run -e BASE_URL=http://api.example.com tests/load/api-test.js
 * With scenario: k6 run -e SCENARIO=stress tests/load/api-test.js
 */

import http from 'k6/http';
import { check, sleep, group } from 'k6';
import { Rate, Trend } from 'k6/metrics';
import { API_URL, TEST_USER, DEFAULT_THRESHOLDS, getScenario } from './config.js';

// Custom metrics
const errorRate = new Rate('errors');
const authDuration = new Trend('auth_duration');
const documentsDuration = new Trend('documents_duration');
const eventsDuration = new Trend('events_duration');

// Test configuration
const scenario = __ENV.SCENARIO || 'smoke';

export const options = {
    scenarios: {
        default: getScenario(scenario),
    },
    thresholds: {
        ...DEFAULT_THRESHOLDS,
        auth_duration: ['p(95)<300'],
        documents_duration: ['p(95)<500'],
        events_duration: ['p(95)<500'],
    },
};

// Shared state for authenticated requests
let authToken = null;

// Setup - runs once before test
export function setup() {
    console.log(`Running ${scenario} test against ${API_URL}`);
    console.log('Authenticating test user...');

    const loginRes = http.post(`${API_URL}/auth/login`, JSON.stringify({
        email: TEST_USER.email,
        password: TEST_USER.password,
    }), {
        headers: { 'Content-Type': 'application/json' },
    });

    if (loginRes.status === 200) {
        const body = JSON.parse(loginRes.body);
        console.log('Authentication successful');
        return { token: body.access_token || body.token };
    } else {
        console.error(`Authentication failed: ${loginRes.status} - ${loginRes.body}`);
        return { token: null };
    }
}

// Main test function - runs for each VU iteration
export default function (data) {
    const headers = {
        'Content-Type': 'application/json',
    };

    if (data.token) {
        headers['Authorization'] = `Bearer ${data.token}`;
    }

    // Health check (unauthenticated)
    group('Health Endpoints', () => {
        const healthRes = http.get(`${API_URL.replace('/api', '')}/health`);
        check(healthRes, {
            'health status is 200': (r) => r.status === 200,
        });
        errorRate.add(healthRes.status !== 200);
    });

    // Auth endpoints
    group('Authentication', () => {
        const start = Date.now();

        // Login
        const loginRes = http.post(`${API_URL}/auth/login`, JSON.stringify({
            email: TEST_USER.email,
            password: TEST_USER.password,
        }), { headers });

        authDuration.add(Date.now() - start);

        const loginOk = check(loginRes, {
            'login status is 200': (r) => r.status === 200,
            'login returns token': (r) => {
                if (r.status !== 200) return false;
                const body = JSON.parse(r.body);
                return body.access_token || body.token;
            },
        });

        errorRate.add(!loginOk);
    });

    sleep(0.5);

    // Documents endpoints (authenticated)
    if (data.token) {
        group('Documents', () => {
            const start = Date.now();

            // List documents
            const listRes = http.get(`${API_URL}/documents`, { headers });
            documentsDuration.add(Date.now() - start);

            const listOk = check(listRes, {
                'documents list status is 200': (r) => r.status === 200,
                'documents list returns array': (r) => {
                    if (r.status !== 200) return false;
                    const body = JSON.parse(r.body);
                    return Array.isArray(body.documents || body.data || body);
                },
            });

            errorRate.add(!listOk);
        });

        sleep(0.5);

        // Events/Schedule endpoints
        group('Schedule/Events', () => {
            const start = Date.now();

            // List events
            const eventsRes = http.get(`${API_URL}/events`, { headers });
            eventsDuration.add(Date.now() - start);

            const eventsOk = check(eventsRes, {
                'events list status is 200': (r) => r.status === 200,
            });

            errorRate.add(!eventsOk);
        });

        sleep(0.5);

        // Notifications endpoints
        group('Notifications', () => {
            const notifRes = http.get(`${API_URL}/notifications`, { headers });

            check(notifRes, {
                'notifications status is 200': (r) => r.status === 200,
            });

            errorRate.add(notifRes.status !== 200);
        });

        sleep(0.5);

        // Messaging endpoints
        group('Messaging', () => {
            const convRes = http.get(`${API_URL}/conversations`, { headers });

            check(convRes, {
                'conversations status is 200': (r) => r.status === 200,
            });

            errorRate.add(convRes.status !== 200);
        });
    }

    sleep(1);
}

// Teardown - runs once after test
export function teardown(data) {
    console.log('Test completed');
}
