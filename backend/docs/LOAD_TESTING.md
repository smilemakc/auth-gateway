# Load Testing Guide

## Overview

This guide explains how to perform load testing on Auth Gateway to ensure it can handle production workloads.

## Tools

### Recommended Tools

1. **k6** - Modern load testing tool (recommended)
2. **Apache Bench (ab)** - Simple HTTP benchmarking
3. **wrk** - High-performance HTTP benchmarking
4. **JMeter** - Full-featured load testing tool
5. **Gatling** - Scala-based load testing framework

## k6 Setup

### Installation

```bash
# macOS
brew install k6

# Linux
sudo gpg -k
sudo gpg --no-default-keyring --keyring /usr/share/keyrings/k6-archive-keyring.gpg --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
echo "deb [signed-by=/usr/share/keyrings/k6-archive-keyring.gpg] https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
sudo apt-get update
sudo apt-get install k6

# Docker
docker pull grafana/k6
```

## Test Scenarios

### 1. Authentication Load Test

Test login endpoint under load:

```javascript
// load-test-auth.js
import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

const errorRate = new Rate('errors');

export const options = {
  stages: [
    { duration: '30s', target: 50 },   // Ramp up to 50 users
    { duration: '1m', target: 50 },    // Stay at 50 users
    { duration: '30s', target: 100 }, // Ramp up to 100 users
    { duration: '1m', target: 100 },  // Stay at 100 users
    { duration: '30s', target: 0 },   // Ramp down
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'], // 95% of requests under 500ms
    errors: ['rate<0.1'],              // Error rate under 10%
  },
};

export default function () {
  const baseURL = __ENV.BASE_URL || 'http://localhost:8181';
  
  // Test login
  const loginPayload = JSON.stringify({
    email: 'test@example.com',
    password: 'testpassword',
  });
  
  const loginParams = {
    headers: { 'Content-Type': 'application/json' },
  };
  
  const loginRes = http.post(`${baseURL}/api/auth/login`, loginPayload, loginParams);
  const loginSuccess = check(loginRes, {
    'login status is 200': (r) => r.status === 200,
    'login has token': (r) => JSON.parse(r.body).access_token !== undefined,
  });
  
  errorRate.add(!loginSuccess);
  
  sleep(1);
}
```

Run:
```bash
k6 run load-test-auth.js
```

### 2. Token Validation Load Test

Test token validation endpoint:

```javascript
// load-test-token-validation.js
import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '1m', target: 200 },
    { duration: '2m', target: 200 },
    { duration: '1m', target: 0 },
  ],
};

const tokens = []; // Pre-populate with valid tokens

export default function () {
  const baseURL = __ENV.BASE_URL || 'http://localhost:8181';
  const token = tokens[Math.floor(Math.random() * tokens.length)];
  
  const params = {
    headers: {
      'Authorization': `Bearer ${token}`,
    },
  };
  
  const res = http.get(`${baseURL}/api/auth/validate`, params);
  
  check(res, {
    'status is 200': (r) => r.status === 200,
    'response time < 100ms': (r) => r.timings.duration < 100,
  });
  
  sleep(0.1);
}
```

### 3. API Endpoint Load Test

Test general API endpoints:

```javascript
// load-test-api.js
import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '30s', target: 100 },
    { duration: '1m', target: 100 },
    { duration: '30s', target: 200 },
    { duration: '1m', target: 200 },
    { duration: '30s', target: 0 },
  ],
};

export default function () {
  const baseURL = __ENV.BASE_URL || 'http://localhost:8181';
  const token = __ENV.ACCESS_TOKEN; // Pre-obtained token
  
  const endpoints = [
    '/api/auth/validate',
    '/api/admin/users',
    '/api/admin/roles',
  ];
  
  const endpoint = endpoints[Math.floor(Math.random() * endpoints.length)];
  
  const params = {
    headers: {
      'Authorization': `Bearer ${token}`,
    },
  };
  
  const res = http.get(`${baseURL}${endpoint}`, params);
  
  check(res, {
    'status is 200 or 401': (r) => r.status === 200 || r.status === 401,
  });
  
  sleep(0.5);
}
```

## Apache Bench Examples

### Simple Load Test

```bash
# 1000 requests, 10 concurrent
ab -n 1000 -c 10 http://localhost:8181/api/auth/health

# With authentication
ab -n 1000 -c 10 \
  -H "Authorization: Bearer YOUR_TOKEN" \
  http://localhost:8181/api/auth/validate
```

### POST Request

```bash
ab -n 1000 -c 10 \
  -p login.json \
  -T application/json \
  http://localhost:8181/api/auth/login
```

Where `login.json`:
```json
{
  "email": "test@example.com",
  "password": "testpassword"
}
```

## Performance Targets

### Recommended Targets

- **Response Time (p95)**: < 200ms for authentication
- **Response Time (p95)**: < 100ms for token validation
- **Response Time (p95)**: < 500ms for admin operations
- **Error Rate**: < 0.1% (1 error per 1000 requests)
- **Throughput**: > 1000 requests/second for token validation
- **Concurrent Users**: Support 500+ concurrent users

### Database Performance

- **Query Time**: < 50ms for simple queries
- **Connection Pool**: Monitor connection pool usage
- **Transaction Time**: < 100ms for user creation

## Monitoring During Load Tests

### Metrics to Monitor

1. **Application Metrics**:
   - Request rate
   - Response time (p50, p95, p99)
   - Error rate
   - Active connections

2. **Database Metrics**:
   - Connection pool usage
   - Query duration
   - Transaction rate
   - Lock wait time

3. **System Metrics**:
   - CPU usage
   - Memory usage
   - Network I/O
   - Disk I/O

### Prometheus Queries

```promql
# Request rate
rate(auth_gateway_http_requests_total[1m])

# Response time (p95)
histogram_quantile(0.95, auth_gateway_http_request_duration_seconds_bucket)

# Error rate
rate(auth_gateway_http_requests_total{status=~"5.."}[1m])

# Database connections
auth_gateway_database_connections{state="in_use"}
```

## Load Test Checklist

### Before Testing

- [ ] Set up test environment (separate from production)
- [ ] Prepare test data (users, tokens, etc.)
- [ ] Configure monitoring (Prometheus, Grafana)
- [ ] Set up database backups
- [ ] Document baseline metrics

### During Testing

- [ ] Monitor application logs
- [ ] Watch database performance
- [ ] Monitor system resources
- [ ] Track error rates
- [ ] Document any issues

### After Testing

- [ ] Analyze results
- [ ] Identify bottlenecks
- [ ] Document findings
- [ ] Create improvement plan
- [ ] Share results with team

## Common Issues and Solutions

### High Response Times

**Symptoms**: p95 response time > 500ms

**Possible Causes**:
- Database connection pool exhausted
- Slow database queries
- High CPU usage
- Network latency

**Solutions**:
- Increase database connection pool
- Optimize queries
- Add database indexes
- Scale horizontally

### High Error Rates

**Symptoms**: Error rate > 1%

**Possible Causes**:
- Database connection failures
- Rate limiting too aggressive
- Memory exhaustion
- Deadlocks

**Solutions**:
- Increase connection pool
- Adjust rate limits
- Increase memory
- Optimize transactions

### Connection Pool Exhaustion

**Symptoms**: "too many connections" errors

**Solutions**:
- Increase `DB_MAX_OPEN_CONNS`
- Reduce connection lifetime
- Use connection pooling
- Scale database

## Continuous Load Testing

### CI/CD Integration

Add load tests to CI/CD pipeline:

```yaml
# .github/workflows/load-test.yml
name: Load Test
on:
  schedule:
    - cron: '0 2 * * *' # Daily at 2 AM
  workflow_dispatch:

jobs:
  load-test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Run k6 tests
        uses: grafana/k6-action@v0.2.0
        with:
          filename: load-test-auth.js
          cloud: true
          token: ${{ secrets.K6_CLOUD_TOKEN }}
```

## Best Practices

1. **Start Small**: Begin with low load and gradually increase
2. **Test Realistic Scenarios**: Use realistic user behavior patterns
3. **Monitor Everything**: Monitor all layers (app, DB, system)
4. **Test Regularly**: Run load tests regularly, not just before releases
5. **Document Results**: Keep records of test results for comparison
6. **Test Failure Scenarios**: Test how system handles failures
7. **Use Production-like Data**: Use realistic data volumes and patterns

## Resources

- [k6 Documentation](https://k6.io/docs/)
- [Apache Bench Guide](https://httpd.apache.org/docs/2.4/programs/ab.html)
- [Load Testing Best Practices](https://k6.io/docs/test-types/introduction/)

