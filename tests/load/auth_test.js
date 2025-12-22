import http from 'k6/http';
import { check, sleep } from 'k6';
import { Counter, Trend } from 'k6/metrics';
import { SharedArray } from 'k6/data';

const apiVersion = 1;
const apiBaseUrl = `http://localhost:8070/api/v${apiVersion}`;

// Metrics
const logins = new Counter('logins');
const loginDuration = new Trend('login_duration');

// Created users
const testUsers = new SharedArray('users', function() {
  const users = [];
  for (let i = 1; i <= 100; i++) {
    users.push({
      username: `loadtest_user${i}`,
      password: 'password123',
    });
  }
  return users;
});

export const options = {
  scenarios: {
    login_load: {
      executor: 'ramping-vus',
      startVUs: 10,
      stages: [
        { duration: '30s', target: 50 },
        { duration: '1m', target: 50 },
        { duration: '30s', target: 100 },
        { duration: '2m', target: 100 },
        { duration: '30s', target: 0 },
      ],
      gracefulRampDown: '30s',
    },
  },
  thresholds: {
    login_duration: ['p(95)<300'],
    'http_req_duration{type:login}': ['p(95)<500'],
  },
};

// SETUP: registering users before test
export function setup() {
  console.log("Preparing data...");

  const baseUrl = __ENV.BASE_URL || apiBaseUrl;
  const createdUsers = [];
  
  for (let i = 0; i < Math.min(50, testUsers.length); i++) {
    const user = testUsers[i];
    
    const regRes = http.post(`${baseUrl}/auth/register`, JSON.stringify({
      name: user.username,
      password: user.password,
    }), {
      headers: { 'Content-Type': 'application/json' },
      timeout: '10s',
    });
    
    if (regRes.status === 201) {
      createdUsers.push(user);
    } else if (regRes.status === 409) {
      createdUsers.push(user);
    } else {
      console.error(`Registration error: ${user.username}: ${regRes.status}`);
    }
    
    if (i % 10 === 0) {
      sleep(0.1);
    }
  }
  
  console.log(`Registered ${createdUsers.length} users`);
  return { users: createdUsers };
}

// Login test
export default function(data) {
  const baseUrl = __ENV.BASE_URL || apiBaseUrl;
  
  // Getting users from setted up data
  if (!data.users || data.users.length === 0) {
    console.error('No registered users');
    return;
  }
  
  const user = data.users[__VU % data.users.length];
  
  const startTime = Date.now();
  
  const res = http.post(`${baseUrl}/auth/login`, JSON.stringify({
    name: user.username,
    password: user.password,
  }), {
    headers: { 'Content-Type': 'application/json' },
    tags: { type: 'login' },
  });
  
  const duration = Date.now() - startTime;
  loginDuration.add(duration);
  
  const success = check(res, {
    'status is 200': (r) => r.status === 200,
    'response has token': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.token && body.token.length > 10;
      } catch {
        return false;
      }
    },
    'response time < 500ms': (r) => r.timings.duration < 500,
  });
  
  if (success) {
    logins.add(1);
  }
  
  sleep(Math.random() * 2);
}

export function teardown(data) {
  console.log(`Users registered: ${data.users.length}`);
  console.log(`Users logged in: ${logins}`);
}