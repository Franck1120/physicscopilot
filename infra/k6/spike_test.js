/**
 * PhysicsCopilot — k6 Spike Test
 *
 * Scenario: Sudden spike to 100 users, then back to baseline.
 * Purpose: Verify server handles sudden traffic burst without crashing.
 *
 * Run with: k6 run infra/k6/spike_test.js
 */

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

const errorRate = new Rate('errors');

export const options = {
  stages: [
    { duration: '10s', target: 5 },    // Baseline
    { duration: '10s', target: 100 },  // Spike!
    { duration: '30s', target: 100 },  // Hold spike
    { duration: '10s', target: 5 },    // Recovery
    { duration: '10s', target: 0 },    // Ramp down
  ],
  thresholds: {
    'http_req_failed': ['rate<0.05'],  // Allow up to 5% errors during spike
    'errors': ['rate<0.05'],
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

export default function () {
  const res = http.get(`${BASE_URL}/health`);
  check(res, { 'status 200': (r) => r.status === 200 }) || errorRate.add(1);
  sleep(0.5);
}
