/**
 * k6 Load Test - Health & Public Endpoints (No Auth Required)
 *
 * Quick test for basic API availability and performance.
 * Run: k6 run tests/load/health-test.js
 */

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';
import { BASE_URL, DEFAULT_THRESHOLDS, getScenario } from './config.js';

// Custom metrics
const errorRate = new Rate('errors');
const healthDuration = new Trend('health_duration');
const readyDuration = new Trend('ready_duration');

const scenario = __ENV.SCENARIO || 'smoke';

export const options = {
    scenarios: {
        default: getScenario(scenario),
    },
    thresholds: {
        ...DEFAULT_THRESHOLDS,
        health_duration: ['p(95)<100', 'p(99)<200'],
        ready_duration: ['p(95)<100', 'p(99)<200'],
    },
};

export default function () {
    // Health endpoint
    {
        const start = Date.now();
        const res = http.get(`${BASE_URL}/health`);
        healthDuration.add(Date.now() - start);

        const ok = check(res, {
            'health status is 200': (r) => r.status === 200,
            'health response has status': (r) => {
                if (r.status !== 200) return false;
                const body = JSON.parse(r.body);
                return body.status === 'ok' || body.status === 'healthy';
            },
        });

        errorRate.add(!ok);
    }

    sleep(0.1);

    // Liveness probe
    {
        const res = http.get(`${BASE_URL}/live`);
        check(res, {
            'live status is 200': (r) => r.status === 200,
        });
        errorRate.add(res.status !== 200);
    }

    sleep(0.1);

    // Readiness probe
    {
        const start = Date.now();
        const res = http.get(`${BASE_URL}/ready`);
        readyDuration.add(Date.now() - start);

        check(res, {
            'ready status is 200': (r) => r.status === 200,
        });
        errorRate.add(res.status !== 200);
    }

    sleep(0.1);

    // Swagger docs (if enabled)
    {
        const res = http.get(`${BASE_URL}/swagger/index.html`);
        check(res, {
            'swagger is accessible': (r) => r.status === 200,
        });
        // Don't count swagger errors as critical
    }

    sleep(0.5);
}
