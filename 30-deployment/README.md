# 30 — Deployment (Docker + Kubernetes + CI/CD)

Modul penutup: membawa aplikasi Go ke produksi. Menggabungkan graceful shutdown (20), config (19), observability (18), dan security (27).

Jalankan lokal:
```bash
go run ./30-deployment
curl localhost:8080/healthz   # liveness
curl localhost:8080/readyz    # readiness
curl localhost:8080/version
```
Verifikasi otomatis: `go test ./30-deployment`

## 1. Binari & Dockerfile produksi

Go menghasilkan **binari statis tunggal** → image kontainer sangat kecil.
```dockerfile
# multi-stage: build di image besar, jalankan di image minimal
FROM golang:1.26 AS build
RUN CGO_ENABLED=0 go build -ldflags "-s -w -X main.version=${VERSION}" -o /out/app ./30-deployment

FROM gcr.io/distroless/static-debian12:nonroot   # tanpa shell/OS, non-root
COPY --from=build /out/app /app
USER nonroot:nonroot
ENTRYPOINT ["/app"]
```
- **`CGO_ENABLED=0`** → binari statis, jalan di image kosong.
- **`-ldflags "-s -w"`** → buang debug info, kecilkan ukuran.
- **`-X main.version=...`** → suntik versi saat build (lihat `/version`).
- **distroless + nonroot** → permukaan serangan minimal (tak ada shell untuk exploit).
- **`.dockerignore`** → konteks build kecil, file rahasia/test tak masuk image.

Build:
```bash
docker build -f 30-deployment/Dockerfile --build-arg VERSION=1.0.0 -t goapp:1.0.0 .
```

## 2. Health Probes (wajib di Kubernetes)

| Probe | Endpoint | Gagal → |
|-------|----------|---------|
| **Liveness** | `/healthz` | K8s **restart** pod (proses macet) |
| **Readiness** | `/readyz` | K8s **hentikan traffic** ke pod (tak restart) |

Saat menerima SIGTERM (Modul 20), set `readyz` → 503 dulu → load balancer berhenti kirim traffic baru → selesaikan request in-flight → mati. Zero-downtime deploy.

## 3. Kubernetes manifests (`k8s/`)

- **`deployment.yaml`** — 3 replika, probes, resource limits, env dari Secret, `terminationGracePeriodSeconds: 30`.
- **`service.yaml`** — expose pod via satu alamat internal (ClusterIP).

Deploy:
```bash
kubectl apply -f 30-deployment/k8s/
kubectl rollout status deployment/goapp
```
**Resource requests/limits** penting: mencegah satu pod menghabiskan node & membantu penjadwalan.

## 4. CI/CD

Repo ini sudah punya **CI** (`.github/workflows/ci.yml`): gofmt + vet + build + test-race tiap push. Alur CD lengkap:
```
push -> CI (test) -> build image -> push ke registry -> deploy (kubectl/Helm/ArgoCD)
```
Praktik: tag image dengan versi/commit (bukan `latest`), promosikan artefak yang sama dev→staging→prod (12-factor).

## Checklist kesiapan produksi (rangkuman kurikulum)
- [x] Structured logging + metrics + health (Modul 18, 30)
- [x] Config & secret via env (Modul 19)
- [x] Graceful shutdown (Modul 20)
- [x] Migrasi database (Modul 21)
- [x] Rate limit, security headers, TLS (Modul 27)
- [x] Test + CI (Modul 8, CI)
- [x] Image kecil, non-root, resource limits (modul ini)

## Latihan
1. Build & jalankan image Docker secara lokal (`docker run -p 8080:8080 goapp:1.0.0`).
2. Tambah `k8s/configmap.yaml` & `secret.yaml`, referensikan dari deployment.
3. Tambah `HorizontalPodAutoscaler` (autoscale berdasar CPU).
4. Buat GitHub Actions workflow yang build & push image ke GHCR.
5. Bungkus manifest jadi **Helm chart** agar mudah dikonfigurasi per environment.

---
🎓 **SELESAI SEMUA!** 30 modul dari sintaks dasar hingga deployment produksi. Kamu kini punya fondasi lengkap sebagai Go backend engineer. Lihat `README.md` root untuk peta penuh.

## ✅ Solusi Latihan (Pembahasan)

1. **Jalankan image lokal** — `docker build -t goapp:1.0.0 . && docker run -p 8080:8080 goapp:1.0.0`. Uji `curl localhost:8080/healthz`. Image distroless → kecil & minim permukaan serangan.
2. **ConfigMap & Secret** — `configmap.yaml` untuk config non-rahasia, `secret.yaml` (base64) untuk rahasia; referensikan di Deployment via `envFrom` / `valueFrom.secretKeyRef`. Jangan hardcode di image.
3. **HPA** — `HorizontalPodAutoscaler` dengan `metrics: cpu, averageUtilization: 70`. Pod bertambah saat CPU naik. Butuh `resources.requests.cpu` terisi agar HPA bisa menghitung.
4. **GitHub Actions → GHCR** — workflow: `docker/login-action` ke `ghcr.io`, `docker/build-push-action` dengan tag `ghcr.io/<user>/goapp:${{ github.sha }}`. CI/CD otomatis tiap push.
5. **Helm chart** — `helm create` lalu parametrikan image tag, replica, resources di `values.yaml`. Satu chart, beda `values-{dev,prod}.yaml` per environment (lihat Modul 39).
