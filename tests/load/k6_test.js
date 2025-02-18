import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');

// Test configuration
export const options = {
  stages: [
    { duration: '1m', target: 50 },  // Ramp up to 50 users
    { duration: '3m', target: 50 },  // Stay at 50 users
    { duration: '1m', target: 100 }, // Ramp up to 100 users
    { duration: '3m', target: 100 }, // Stay at 100 users
    { duration: '1m', target: 0 },   // Ramp down to 0 users
  ],
  thresholds: {
    'http_req_duration': ['p(95)<500'], // 95% of requests should be below 500ms
    'errors': ['rate<0.1'],             // Error rate should be below 10%
  },
};

const BASE_URL = 'http://localhost:3000';
const TEST_URL = 'https://www.example.com/test';

// Simulated user behavior
export default function() {
  // 1. Create a short URL
  const createPayload = JSON.stringify({
    url: TEST_URL,
  });

  const createRes = http.post(`${BASE_URL}/api/shorten`, createPayload, {
    headers: { 'Content-Type': 'application/json' },
  });

  check(createRes, {
    'create status is 201': (r) => r.status === 201,
    'has shortCode': (r) => JSON.parse(r.body).shortCode !== undefined,
  }) || errorRate.add(1);

  if (createRes.status === 201) {
    const shortCode = JSON.parse(createRes.body).shortCode;

    // 2. Get URL info
    const getRes = http.get(`${BASE_URL}/api/shorten/${shortCode}`);
    check(getRes, {
      'get status is 200': (r) => r.status === 200,
      'original url matches': (r) => JSON.parse(r.body).originalUrl === TEST_URL,
    }) || errorRate.add(1);

    // 3. Access short URL (redirect)
    const redirectRes = http.get(`${BASE_URL}/${shortCode}`, {
      redirects: 0, // Don't follow redirects
    });
    check(redirectRes, {
      'redirect status is 301': (r) => r.status === 301,
      'redirect location is correct': (r) => r.headers['Location'] === TEST_URL,
    }) || errorRate.add(1);

    // 4. Get statistics
    const statsRes = http.get(`${BASE_URL}/api/shorten/${shortCode}/stats`);
    check(statsRes, {
      'stats status is 200': (r) => r.status === 200,
      'has access count': (r) => JSON.parse(r.body).accessCount > 0,
    }) || errorRate.add(1);

    // 5. Update URL
    const updatePayload = JSON.stringify({
      url: TEST_URL + '/updated',
    });
    const updateRes = http.put(`${BASE_URL}/api/shorten/${shortCode}`, updatePayload, {
      headers: { 'Content-Type': 'application/json' },
    });
    check(updateRes, {
      'update status is 200': (r) => r.status === 200,
      'url is updated': (r) => JSON.parse(r.body).originalUrl === TEST_URL + '/updated',
    }) || errorRate.add(1);

    // 6. Delete URL
    const deleteRes = http.del(`${BASE_URL}/api/shorten/${shortCode}`);
    check(deleteRes, {
      'delete status is 204': (r) => r.status === 204,
    }) || errorRate.add(1);
  }

  // Wait between iterations
  sleep(1);
} 