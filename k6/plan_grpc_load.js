// Prueba de rendimiento gRPC directa contra ms-plan (sin pasar por Kong).
//
// Preparacion:
//   kubectl port-forward -n sdata svc/ms-plan 50051:50051
// Ejecucion:
//   k6 run k6/plan_grpc_load.js
//   k6 run -e GRPC_ADDR=localhost:50051 -e PROFILE=stress k6/plan_grpc_load.js
//   k6 run --summary-export=k6-grpc-summary.json k6/plan_grpc_load.js
//
// Cada iteracion hace un ciclo CRUD completo y borra lo que crea.

import grpc from 'k6/net/grpc';
import { check, group, sleep } from 'k6';
import { Trend, Counter } from 'k6/metrics';

const client = new grpc.Client();
client.load(['.'], 'plan_k6.proto');

const ADDR = __ENV.GRPC_ADDR || 'localhost:50051';
const SVC = 'plan.v1.PlanService';

const tCreate = new Trend('rpc_create_ms', true);
const tGet = new Trend('rpc_get_ms', true);
const tList = new Trend('rpc_list_ms', true);
const tUpdate = new Trend('rpc_update_ms', true);
const tDelete = new Trend('rpc_delete_ms', true);
const cleanupFailures = new Counter('cleanup_failures');

const profiles = {
  // Humo: verifica que el flujo funciona con 1 usuario.
  smoke: [
    { duration: '1s', target: 1 },
    { duration: '20s', target: 1 },
    { duration: '1s', target: 0 },
  ],
  // Carga normal.
  load: [
    { duration: '30s', target: 10 },
    { duration: '1m', target: 10 },
    { duration: '15s', target: 0 },
  ],
  // Estres: busca el punto de degradacion.
  stress: [
    { duration: '30s', target: 20 },
    { duration: '1m', target: 50 },
    { duration: '1m', target: 100 },
    { duration: '30s', target: 0 },
  ],
};

export const options = {
  scenarios: {
    crud: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: profiles[__ENV.PROFILE || 'load'],
      gracefulRampDown: '10s',
    },
  },
  thresholds: {
    // Umbrales de aceptacion (ajustar a los criterios del plan de pruebas).
    grpc_req_duration: ['p(95)<300', 'p(99)<800'],
    checks: ['rate>0.99'],
    cleanup_failures: ['count==0'],
  },
};

function ok(res) {
  return res && res.status === grpc.StatusOK;
}

function invokeMeasured(trend, method, payload) {
  const started = Date.now();
  const res = client.invoke(method, payload);
  trend.add(Date.now() - started);
  return res;
}

export default function () {
  if (__ITER === 0) {
    client.connect(ADDR, { plaintext: true });
  }

  const title = `K6-${__VU}-${__ITER}-${Date.now()}`;
  let id = '';

  group('create', () => {
    const res = invokeMeasured(tCreate, `${SVC}/CreatePlan`, {
      data: {
        planTitle: title,
        planSubtitle: 'Plan de prueba de carga',
        planType: 'monthly',
        planCurrencyPrice: 'COP',
        planDays: 30,
        planUsers: 5,
        planOffices: 1,
        planPrice: 99900,
        activePlan: true,
      },
    });
    check(res, {
      'create: status OK': (r) => ok(r),
      'create: devuelve id': (r) => ok(r) && !!r.message.data.id,
    });
    if (ok(res)) id = res.message.data.id;
  });

  if (id) {
    group('read', () => {
      const g = invokeMeasured(tGet, `${SVC}/GetPlan`, { id });
        check(g, {
        'get: status OK': (r) => ok(r),
        'get: mismo titulo': (r) => ok(r) && r.message.data.planTitle === title,
      });

      const l = invokeMeasured(tList, `${SVC}/ListPlan`, { page: 1, limit: 10 });
        check(l, { 'list: status OK': (r) => ok(r) });
    });

    group('update', () => {
      const u = invokeMeasured(tUpdate, `${SVC}/UpdatePlan`, {
        id,
        data: {
          planTitle: title,
          planSubtitle: 'Actualizado en carga',
          planType: 'monthly',
          planCurrencyPrice: 'COP',
          planDays: 30,
          planUsers: 5,
          planOffices: 1,
          planPrice: 129900,
          activePlan: true,
        },
      });
        check(u, {
        'update: status OK': (r) => ok(r),
        'update: precio nuevo': (r) => ok(r) && r.message.data.planPrice === 129900,
      });
    });

    group('delete', () => {
      const d = invokeMeasured(tDelete, `${SVC}/DeletePlan`, { id });
        const good = check(d, { 'delete: status OK': (r) => ok(r) });
      if (!good) cleanupFailures.add(1);
    });
  }

  // Caso negativo: id inexistente debe ser NotFound (no un error 5xx).
  const nf = client.invoke(`${SVC}/GetPlan`, { id: '000000000000000000000001' });
  check(nf, { 'get inexistente: NotFound': (r) => r && r.status === grpc.StatusNotFound });

  sleep(0.2);
}

export function teardown() {
  client.close();
}
