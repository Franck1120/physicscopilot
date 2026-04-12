/**
 * PhysicsCopilot — k6 Load Test
 *
 * Scenario: Steady load simulating typical API usage.
 * Thresholds:
 *   - p95 response time < 500ms for health endpoint
 *   - error rate < 1%
 *
 * Run with: k6 run infra/k6/load_test.js
 * Override URL: k6 run --env BASE_URL=http://your-server:8080 infra/k6/load_test.js
 */

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';

const errorRate = new Rate('errors');
const healthLatency = new Trend('health_latency', true);

export const options = {
  stages: [
    { duration: '30s', target: 10 },   // Ramp up to 10 users
    { duration: '1m',  target: 10 },   // Steady state
    { duration: '30s', target: 0 },    // Ramp down
  ],
  thresholds: {
    'http_req_duration{endpoint:health}': ['p(95)<500'],
    'http_req_failed': ['rate<0.01'],
    'errors': ['rate<0.01'],
    'health_latency': ['p(95)<500'],
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

export default function () {
  // Health check — public endpoint, no auth required
  const healthRes = http.get(`${BASE_URL}/health`, {
    tags: { endpoint: 'health' },
  });
  healthLatency.add(healthRes.timings.duration);
  check(healthRes, {
    'health status 200': (r) => r.status === 200,
    'health has status field': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.status === 'ok';
      } catch {
        return false;
      }
    },
  }) || errorRate.add(1);

  // Create session — requires auth; expect 401 without JWT (tests auth guard)
  const sessionRes = http.post(
    `${BASE_URL}/api/sessions`,
    JSON.stringify({ device_brand: 'LoadTest', device_model: 'k6' }),
    { headers: { 'Content-Type': 'application/json' }, tags: { endpoint: 'create_session' } },
  );
  check(sessionRes, {
    'create session returns 201 or 401': (r) => r.status === 201 || r.status === 401,
  }) || errorRate.add(1);

  sleep(1);
}

export function handleSummary(data) {
  return {
    'stdout': JSON.stringify(data, null, 2),
  };
}
