/**
 * PhysicsCopilot — k6 Soak Test
 *
 * Scenario: Low constant load over 30 minutes.
 * Purpose: Detect memory leaks, connection pool exhaustion, and degradation over time.
 *
 * Run with: k6 run infra/k6/soak_test.js
 * (Reduce duration for CI: k6 run --env DURATION=2m infra/k6/soak_test.js)
 */

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Trend } from 'k6/metrics';

const memoryTrend = new Trend('server_memory_mb', true);

export const options = {
  stages: [
    { duration: '2m',   target: 5 },   // Ramp up slowly
    { duration: __ENV.DURATION || '30m', target: 5 },  // Soak at 5 users
    { duration: '2m',   target: 0 },   // Ramp down
  ],
  thresholds: {
    'http_req_failed': ['rate<0.01'],
    'http_req_duration': ['p(95)<1000'],
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

export default function () {
  const res = http.get(`${BASE_URL}/health`);
  check(res, { 'health ok': (r) => r.status === 200 });

  // Track server memory if exposed in health response
  if (res.status === 200) {
    try {
      const body = JSON.parse(res.body);
      if (body.memory_mb) memoryTrend.add(body.memory_mb);
    } catch { /* ignore parse errors */ }
  }

  sleep(2);
}
