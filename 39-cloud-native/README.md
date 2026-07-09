# 39 — Cloud-Native

Lanjutan Modul 30. Tiga topik cloud-native: **Helm** (paketkan aplikasi K8s), **pola controller/reconcile** (jantung Kubernetes & operator), dan **serverless** (FaaS).

Jalankan:
```bash
go run ./39-cloud-native
go test ./39-cloud-native
```

## 1. Pola Controller / Reconcile Loop (`reconciler.go`)

Kubernetes bekerja dengan **control loop**: terus mengamati state aktual, membandingkan dengan state yang diinginkan, dan bertindak untuk menyelaraskannya.
```
observe (actual) → diff (desired vs actual) → act (converge) → ulangi
```
```go
func (p *PodSet) Reconcile() {
    for p.actual < p.desired { p.actual++ } // scale up
    for p.actual > p.desired { p.actual-- } // scale down
}
```
Output: desired=3 → scale up ke 3; desired=1 → scale down ke 1; saat selaras → **tak ada aksi (idempotent)**.

Ini prinsip di balik **Deployment**, **HPA**, dan **Operator** (aplikasi yang mengelola aplikasi lain). Untuk membangun operator sungguhan: [kubebuilder](https://book.kubebuilder.io) / [controller-runtime](https://github.com/kubernetes-sigs/controller-runtime) — API-nya sama: `Reconcile(req) (Result, error)`.

## 2. Helm — package manager Kubernetes (`helm/`)

Manifest K8s mentah (Modul 30) sulit dikonfigurasi per environment. **Helm** membuat **template** + **values**:
```
helm/
├── Chart.yaml              # metadata chart
├── values.yaml            # nilai default (replicaCount, image, resources)
└── templates/
    ├── deployment.yaml    # pakai {{ .Values.replicaCount }}
    └── service.yaml
```
Deploy dengan values berbeda per environment:
```bash
helm install myapp ./helm                          # default
helm install myapp ./helm -f values-prod.yaml      # override untuk produksi
helm install myapp ./helm --set replicaCount=5     # override satu nilai
helm upgrade myapp ./helm                           # update rilis
helm template ./helm                                # render tanpa deploy (cek hasil)
```
Satu chart, banyak environment — tanpa menyalin-tempel manifest.

## 3. Serverless / FaaS (`serverless.go`)

Kamu hanya menulis **fungsi**; platform mengurus server, skala (termasuk ke nol), & lifecycle. Bayar per-eksekusi.
```go
func HandleOrder(ctx context.Context, req OrderRequest) (OrderResponse, error) { ... }
```
Handler = **fungsi murni** → mudah di-test (test langsung, tanpa server). Deploy ke AWS Lambda:
```go
import "github.com/aws/aws-lambda-go/lambda"
func main() { lambda.Start(HandleOrder) }   // bungkus handler
```
```bash
GOOS=linux GOARCH=arm64 go build -o bootstrap ./...   # binari Lambda
# zip & upload, atau pakai AWS SAM / Serverless Framework
```
Go cocok untuk serverless: **cold start cepat** (binari kecil, tanpa runtime berat).

**Kapan serverless?** Beban tak menentu/sporadis (webhook, cron, event handler). **Kapan tidak?** Beban stabil tinggi (kontainer lebih murah), butuh koneksi persisten, atau latensi cold-start kritis.

## Peta cloud-native (lengkap dengan modul lain)
| Kebutuhan | Alat | Modul |
|-----------|------|-------|
| Kontainer | Docker (distroless) | 30 |
| Orkestrasi | Kubernetes (Deployment, Service, probes) | 30 |
| Paketkan | **Helm** | ini |
| Extend K8s | **Controller/Operator** | ini |
| Config/Secret | ConfigMap/Secret + Viper | 19 |
| Observability | Prometheus/OTel | 18, 33 |
| Serverless | Lambda/Cloud Functions | ini |

## Kapan & Di Mana Dipakai
- Helm: setiap deployment K8s serius (multi-environment).
- Operator: mengotomasi operasi stateful (database, message queue) di K8s.
- Serverless: event-driven, sporadis, glue code.

## Latihan
1. Render Helm chart: `helm template ./39-cloud-native/helm` (butuh helm).
2. Tambah `values-prod.yaml` (replicaCount=5) & bandingkan hasil `helm template -f`.
3. Ubah `Reconcile` agar mempertimbangkan pod "unhealthy" (ganti, bukan hanya jumlah).
4. Bungkus `HandleOrder` dengan `aws-lambda-go` & build binari `bootstrap`.
5. Tambah `templates/_helpers.tpl` untuk label bersama (best practice Helm).

## ✅ Solusi Latihan (Pembahasan)

1. **Render chart** — `helm template ./39-cloud-native/helm` mencetak manifest final tanpa memasang. Cara cepat validasi template sebelum `helm install`.
2. **`values-prod.yaml`** — berisi `replicaCount: 5`; `helm template -f values-prod.yaml .` → bandingkan dengan default (replica beda). Satu chart, banyak environment.
3. **`Reconcile` unhealthy** — selain hitung jumlah pod, cek status; ganti pod "unhealthy" (bukan hanya menambah bila kurang). Reconcile = arahkan state aktual → state diinginkan (pola controller K8s).
4. **Lambda** — bungkus `HandleOrder` dengan `lambda.Start(handler)` (aws-lambda-go), build `GOOS=linux go build -o bootstrap`. Satu fungsi, deploy serverless.
5. **`_helpers.tpl`** — definisikan `{{- define "app.labels" }}` untuk label bersama; panggil `{{ include "app.labels" . }}` di tiap template. DRY, best practice Helm.

---

## 🚀 Teknik Advanced (Level Up)
> 💻 **Contoh runnable + komentar detail** untuk teknik di bawah ada di folder [`advanced/`](advanced). Jalankan: `go run ./39-cloud-native/advanced`


- **Reconcile loop (operator)** — bandingkan *desired* vs *actual*, konvergen; **idempoten** & bisa diulang (level-triggered, bukan edge).
- **CRD & operator SDK** — perluas K8s dengan resource kustom + controller.
- **Helm** — templating chart, `values.yaml` per-environment; `helm lint`/`--dry-run`.
- **Serverless** — handler stateless; waspada *cold start* & batas eksekusi; keep-warm bila perlu.
- **Graceful shutdown di K8s** — hormati SIGTERM & grace period [[20-graceful-shutdown]].
- **GitOps** — deklaratif via Git sebagai sumber kebenaran (ArgoCD/Flux).
