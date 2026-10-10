// Prueba de rendimiento REST a traves de Kong (transcodificacion REST -> gRPC).
//
// Ejecucion:
//   k6 run -e BASE_URL=https://192.168.0.240 k6/plan_rest_kong.js
//   k6 run --summary-export=k6-rest-summary.json k6/plan_rest_kong.js
//
// Antes de la primera ejecucion, valide con scripts/kong_smoke.sh que Kong
// responde y que los nombres de campos del JSON (snake_case) son correctos.

import http from 'k6/http';
import { check, group, sleep } from 'k6';

const BASE = __ENV.BASE_URL || 'https://192.168.0.240';
const HEADERS = { headers: { 'Content-Type': 'application/json' } };

export const options = {
  insecureSkipTLSVerify: true, // certificado autofirmado del cluster local
  scenarios: {
    rest: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '30s', target: 10 },
        { duration: '1m', target: 10 },
        { duration: '15s', target: 0 },
      ],
    },
  },
  thresholds: {
    http_req_duration: ['p(95)<500'],
    http_req_failed: ['rate<0.01'],
    checks: ['rate>0.99'],
  },
};

export default function () {
  const title = `K6REST-${__VU}-${__ITER}-${Date.now()}`;
  let id = '';

  group('create', () => {
    const res = http.post(
      `${BASE}/api/v1/plan`,
      JSON.stringify({
        data: {
          plan_title: title,
          plan_subtitle: 'Plan de prueba REST',
          plan_type: 'monthly',
          plan_currency_price: 'COP',
          plan_days: 30,
          plan_users: 5,
          plan_offices: 1,
          plan_price: 99900,
          active_plan: true,
        },
      }),
      HEADERS,
    );
    check(res, { 'POST /plan -> 200': (r) => r.status === 200 });
    try {
      id = res.json('data.id') || '';
    } catch (e) {
      id = '';
    }
  });

  if (id) {
    group('read', () => {
      const g = http.get(`${BASE}/api/v1/plan/${id}`);
      check(g, {
        'GET /plan/{id} -> 200': (r) => r.status === 200,
        'GET: mismo titulo': (r) => r.json('data.plan_title') === title,
      });

      const l = http.get(`${BASE}/api/v1/plans?page=1&limit=10`);
      check(l, { 'GET /plans -> 200': (r) => r.status === 200 });
    });

    group('update', () => {
      const u = http.put(
        `${BASE}/api/v1/plan/${id}`,
        JSON.stringify({
          data: {
            plan_title: title,
            plan_subtitle: 'Actualizado REST',
            plan_type: 'monthly',
            plan_currency_price: 'COP',
            plan_days: 30,
            plan_users: 5,
            plan_offices: 1,
            plan_price: 129900,
            active_plan: true,
          },
        }),
        HEADERS,
      );
      check(u, { 'PUT /plan/{id} -> 200': (r) => r.status === 200 });
    });

    group('delete', () => {
      const d = http.del(`${BASE}/api/v1/plan/${id}`);
      check(d, { 'DELETE /plan/{id} -> 200': (r) => r.status === 200 });
    });
  }

  // Negativo: id con formato invalido debe devolver 400 (InvalidArgument), no 500.
  const bad = http.get(`${BASE}/api/v1/plan/no-es-un-objectid`, {
    responseCallback: http.expectedStatuses(400),
  });
  check(bad, { 'id invalido -> 400': (r) => r.status === 400 });

  sleep(0.2);
}
