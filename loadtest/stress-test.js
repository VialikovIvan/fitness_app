import http from 'k6/http';
import { check, fail, sleep } from 'k6';
import { randomString } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';

const baseURL = 'http://app:8080';

const urls = {
    Register: baseURL + '/register',
    Login: baseURL + '/login',
    Logout: baseURL + '/logout',
    GetPlans: baseURL + '/plans',
    Checkout: baseURL + '/checkout',
}

const jsonParams = {
    headers: {
        'Content-Type': 'application/json',
    },
};

export const options = {
    scenarios: {
        looker: {
            executor: 'ramping-arrival-rate',
            startRate: 1,
            timeUnit: '1s',
            preAllocatedVUs: 40,
            maxVUs: 80,
            stages: [
                { duration: '1m', target: 2 },
                { duration: '30s', target: 2 },
                { duration: '2m', target: 5 },
                { duration: '1m', target: 5 },
                { duration: '1m', target: 0 },
            ],
            exec: 'looker',
        },
        hesitatingBuyer: {
            executor: 'ramping-arrival-rate',
            startRate: 1,
            timeUnit: '1s',
            preAllocatedVUs: 75,
            maxVUs: 150,
            stages: [
                { duration: '1m', target: 2 },
                { duration: '30s', target: 2 },
                { duration: '2m', target: 3 },
                { duration: '1m', target: 3 },
                { duration: '1m', target: 0 },
            ],
            exec: 'hesitatingBuyer',
        },
    }
};

export function looker() {
    const registerPayload = JSON.stringify({
        email: `${randomString(10)}@k6test.com`,
        password: 'k6password',
        first_name: `k6user_${randomString(5)}`,
        last_name: `k6_${randomString(5)}`,
    });

    const resRegister = http.post(urls.Register, registerPayload, jsonParams);
    if (!check(resRegister, { 'status is 201': (r) => r.status === 201 })) {
        fail('status code was not 201');
    }

    sleep(2);

    const resPlans = http.get(urls.GetPlans);
    if (!check(resPlans, { 'status is 200': (r) => r.status === 200 })) {
        fail('status code was not 200');
    }

    sleep(5);

    const resLogout = http.post(urls.Logout);
    if (!check(resLogout, { 'status is 200': (r) => r.status === 200 })) {
        fail('status code was not 200');
    }
}

export function hesitatingBuyer() {
    const registerData = {
        email: `${randomString(10)}@k6test.com`,
        password: 'k6password',
        first_name: `k6user_${randomString(5)}`,
        last_name: `k6_${randomString(5)}`,
    };
    const registerPayload = JSON.stringify(registerData);

    const loginPayload = JSON.stringify({
        email: registerData.email,
        password: 'k6password',
    });

    const checkoutPayload = JSON.stringify({
        plan_id: 5,
    });

    const resRegister = http.post(urls.Register, registerPayload, jsonParams);
    if (!check(resRegister, { 'status is 201': (r) => r.status === 201 })) {
        fail('status code was not 201');
    }

    sleep(2);

    const resPlans = http.get(urls.GetPlans);
    if (!check(resPlans, { 'status is 200': (r) => r.status === 200 })) {
        fail('status code was not 200');
    }

    sleep(5);

    const resLogout = http.post(urls.Logout);
    if (!check(resLogout, { 'status is 200': (r) => r.status === 200 })) {
        fail('status code was not 200');
    }

    sleep(7);

    const resLogin = http.post(urls.Login, loginPayload, jsonParams);
    if (!check(resLogin, { 'status is 200': (r) => r.status === 200 })) {
        fail('status code was not 200');
    }

    sleep(2);

    const resPlans2 = http.get(urls.GetPlans);
    if (!check(resPlans2, { 'status is 200': (r) => r.status === 200 })) {
        fail('status code was not 200');
    }

    sleep(4);

    const resCheckout = http.post(urls.Checkout, checkoutPayload, jsonParams);
    if (!check(resCheckout, { 'status is 200': (r) => r.status === 200 })) {
        fail('status code was not 200');
    }

    sleep(2);

    const resLogout2 = http.post(urls.Logout);
    if (!check(resLogout2, { 'status is 200': (r) => r.status === 200 })) {
        fail('status code was not 200');
    }
}