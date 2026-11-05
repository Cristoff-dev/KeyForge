import http from 'k6/http';
import { check, sleep } from 'k6';
import { Counter, Rate, Trend } from 'k6/metrics';

const authFailures = new Counter('auth_failures');
const rateLimitedRequests = new Counter('rate_limited_requests');
const successfulRequests = new Rate('successful_requests');
const authLatency = new Trend('auth_latency_ms');

export const options = {
  stages: [
    { duration: '10s', target: 100 },  // Ramp-up to 100 Virtual Users (VUs)
    { duration: '30s', target: 500 },  // Scale to 500 concurrent VUs
    { duration: '20s', target: 1000 }, // Peak stress load at 1,000 VUs
    { duration: '10s', target: 0 },    // Ramp-down to 0
  ],
  thresholds: {
    http_req_duration: ['p(95)<20', 'p(99)<50'], // 95% of requests under 20ms (Redis cache target)
    http_req_failed: ['rate<0.01'],             // Error rate under 1%
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const API_KEY = __ENV.API_KEY || 'prod_live_4k8zN9xP2mL7vQ5R1tY3wE6bC0jS8uD4_a1b2c3d4';

export default function () {
  const params = {
    headers: {
      'Authorization': `Bearer ${API_KEY}`,
      'Accept': 'application/json',
    },
  };

  const res = http.get(`${BASE_URL}/api/v1/resource`, params);

  authLatency.add(res.timings.duration);

  const isSuccess = check(res, {
    'status is 200': (r) => r.status === 200,
  });

  if (isSuccess) {
    successfulRequests.add(1);
  } else {
    successfulRequests.add(0);
    if (res.status === 401) {
      authFailures.add(1);
    } else if (res.status === 429) {
      rateLimitedRequests.add(1);
    }
  }

  sleep(0.001);
}