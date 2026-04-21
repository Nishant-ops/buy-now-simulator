import http from 'k6/http';
import { check, sleep } from 'k6';
import { Counter, Rate, Trend } from 'k6/metrics';

const successCount  = new Counter('purchases_success');
const soldOutCount  = new Counter('purchases_sold_out');
const errorCount    = new Counter('purchases_error');
const successRate   = new Rate('purchase_success_rate');
const buyLatency    = new Trend('buy_latency_ms', true);

// ── Scenario ──────────────────────────────────────────────────────────────────
// Ramp to 5 000 VUs over 30 s, hold for 2 min, ramp down 30 s.
// Each VU fires POST /buy as fast as it can (no sleep between requests).
// With 1 000 000 units this gives enough load to sell everything in ~3-4 min
// at ~5 000–8 000 RPS through 10 app servers backed by one PG node.
export const options = {
    scenarios: {
        flash_sale: {
            executor: 'ramping-vus',
            startVUs: 0,
            stages: [
                { duration: '30s', target: 1000 },
                { duration: '1m',  target: 1000 },
                { duration: '30s', target: 5000 },
                { duration: '2m',  target: 5000 },
                { duration: '30s', target: 0    },
            ],
            gracefulRampDown: '10s',
        },
    },
    thresholds: {
        // P95 buy latency must stay under 500 ms
        buy_latency_ms:        ['p(95)<500'],
        // At least 99 % of responses should be 200 or 409 (not 5xx)
        purchase_success_rate: ['rate>0.01'],
        // Overall HTTP error rate under 1 %
        http_req_failed:       ['rate<0.01'],
    },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

export default function () {
    const start = Date.now();
    const res = http.post(`${BASE_URL}/buy`);
    buyLatency.add(Date.now() - start);

    const ok = check(res, {
        'status is 200 or 409': (r) => r.status === 200 || r.status === 409,
    });

    if (res.status === 200) {
        successCount.add(1);
        successRate.add(1);
    } else if (res.status === 409) {
        soldOutCount.add(1);
        successRate.add(1); // 409 is expected, not an error
    } else {
        errorCount.add(1);
        successRate.add(0);
        console.error(`Unexpected ${res.status}: ${res.body}`);
    }
}

// ── Summary ───────────────────────────────────────────────────────────────────
export function handleSummary(data) {
    const success  = data.metrics.purchases_success?.values?.count  ?? 0;
    const soldOut  = data.metrics.purchases_sold_out?.values?.count ?? 0;
    const errors   = data.metrics.purchases_error?.values?.count    ?? 0;
    const rps      = data.metrics.http_reqs?.values?.rate?.toFixed(0) ?? '?';
    const p95      = data.metrics.buy_latency_ms?.values?.['p(95)']?.toFixed(1) ?? '?';
    const p99      = data.metrics.buy_latency_ms?.values?.['p(99)']?.toFixed(1) ?? '?';

    const summary = `
════════════════════════════════════════
  Buy-Now Simulator — Load Test Results
════════════════════════════════════════
  Purchases (200 OK)  : ${success}
  Sold-out  (409)     : ${soldOut}
  Errors    (5xx/net) : ${errors}
  Total requests      : ${success + soldOut + errors}
  Peak RPS            : ${rps}
  Latency p95         : ${p95} ms
  Latency p99         : ${p99} ms
────────────────────────────────────────
  Oversell check: sold=${success} (must be ≤ 1 000 000)
════════════════════════════════════════
`;
    console.log(summary);
    return { stdout: summary };
}
