// Load test dengan k6 (https://k6.io). Bukan Go, tapi tool standar menguji beban.
// Jalankan:  k6 run 37-advanced-testing/loadtest.js
//
// Menargetkan endpoint (mis. server Modul 13/15). Ganti URL sesuai punyamu.
import http from "k6/http";
import { check, sleep } from "k6";

export const options = {
  // Naikkan beban bertahap: 0 -> 50 user selama 30 detik, tahan, lalu turun.
  stages: [
    { duration: "10s", target: 50 }, // ramp-up
    { duration: "20s", target: 50 }, // steady
    { duration: "10s", target: 0 },  // ramp-down
  ],
  thresholds: {
    http_req_duration: ["p(95)<500"], // 95% request harus < 500ms
    http_req_failed: ["rate<0.01"],   // error rate < 1%
  },
};

export default function () {
  const res = http.get("http://localhost:3000/api/books");
  check(res, {
    "status 200": (r) => r.status === 200,
    "cepat (<300ms)": (r) => r.timings.duration < 300,
  });
  sleep(1);
}
