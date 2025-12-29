import http from 'k6/http';
import { check, sleep, group } from 'k6';
import { Trend, Rate, Counter } from 'k6/metrics';
import { SharedArray } from 'k6/data';
import { textSummary } from 'https://jslib.k6.io/k6-summary/0.0.1/index.js';

const apiVersion = 1;
const apiBaseUrl = `http://localhost:8070/api/v${apiVersion}`;

export const options = {
    scenarios: {
        typical_user: {
            executor: "ramping-vus",
            stages: [
                { duration: "2m", target: 20 },
                { duration: "5m", target: 10 },
                { duration: "2m", target: 30 },
                { duration: "3m", target: 5 },
                { duration: "1m", target: 0 }
            ],
            gracefulRampDown: "30s",
            startTime: "0s"
        },
        power_user: {
            executor: "constant-vus",
            vus: 5,
            duration: "10m",
            startTime: "1m",
            gracefulStop: "30s"
        }
    },
    
    thresholds: {
        "habit_creation_time": ["p(95)<400"],
        "habit_check_time": ["p(95)<200"],
        "stats_fetch_time": ["p(95)<500"],
        errors: ["rate<0.02"],
        "http_req_duration": ["p(95)<600"]
    }
};

const authenticatedUsers = new SharedArray("auth_users", function() {
    return Array.from({
        length: 50
    }, (_, i) => ({
        username: `loadtest_user${i+1}`,
        password: 'password123',
        token: null
    }));
});

const habitTemplates = [
    { title: "Morning warmup", description: "Get your body stretched" },
    { title: "Books reading", description: "30 minutes a day" },
    { title: "Walk", description: "Get some fresh air" },
    { title: "Gym", description: "" },
    { title: "Drink water", description: "Minimum 2 liters a day" }
];

export function setup() {
    console.log("Logging in test users");
    const updatedUsers = [];
    for (let user of authenticatedUsers) {
        const loginRes = http.post(`${apiBaseUrl}/auth/login`, JSON.stringify({
            name: user.username,
            password: user.password
        }), {
            headers: { "Content-Type": "application/json" },
            timeout: "10s"
        });

        if (loginRes.status === 200) {
            user.token = JSON.parse(loginRes.body).token;
        }
        updatedUsers.push(user);
    }
    sleep(0.1);
    console.log(`Prepared ${updatedUsers.length}`);
    return {
        users: updatedUsers
    };
}

export default function(data) {
    const user = data.users[__VU & data.users.length];

    if (!user.token) {
        console.log("Skipping user: login failed");
        return;
    }

    const headers = {
        "Authorization": `Bearer ${user.token}`,
        "Content-Type": "application/json"
    };

    group("morning", function() {
        const listStart = Date.now();
        const listRes = http.get(`${apiBaseUrl}/habits`, {
            headers: headers,
            tags: { endpoint: "list_habits", group: "habit_crud" }
        });

        check(listRes, {
            "list habits status is 200": (r) => r.status === 200,
            "list habits has data field": (r) => {
                try {
                    return JSON.parse(r.body).habits !== undefined;
                } catch {
                    return false;
                }
            }
        });

        let userHabits = [];
        try {
            const response = JSON.parse(listRes.body);
            userHabits = response.habits || [];
        } catch {
        }


    });
}