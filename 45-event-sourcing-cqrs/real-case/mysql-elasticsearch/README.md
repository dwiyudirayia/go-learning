# Real-Case CQRS — MySQL (write) + Elasticsearch (read)

Implementasi CQRS/Event Sourcing dengan **tech stack produksi**, bukan in-memory.

| Sisi | Store | Alasan |
|------|-------|--------|
| **Write model** | **MySQL** (tabel `events`, append-only) | transaksi ACID; `UNIQUE(aggregate_id, version)` menegakkan **optimistic concurrency** |
| **Read model** | **Elasticsearch** (index `accounts`) | query kaya: range, agregasi, full-text, sorting relevansi — mahal di SQL |
| Sinkronisasi | projector (di sini inline; produksi: **Kafka + Debezium CDC** atau outbox modul 31) | read model *eventually consistent* & **derivatif** (bisa di-rebuild) |

Elasticsearch diakses via **REST HTTP** (`net/http`) — ES memang HTTP+JSON, jadi tak perlu client berat.

## Menjalankan

Butuh Docker. Program **auto-skip** (cetak panduan) bila `MYSQL_DSN`/`ES_URL` kosong, jadi aman di CI tanpa infra.

```bash
# 1. nyalakan MySQL + Elasticsearch
docker compose up -d
#    tunggu ES sehat (~20-30 dtk): curl -s localhost:9200/_cluster/health

# 2. jalankan
MYSQL_DSN='root:secret@tcp(127.0.0.1:3306)/esdemo?parseTime=true' \
ES_URL='http://127.0.0.1:9200' \
go run ./45-event-sourcing-cqrs/real-case/mysql-elasticsearch

# 3. bersihkan
docker compose down -v
```

## Alur yang didemokan
1. **Command** (`Deposited`/`Withdrawn`) → tulis event ke **MySQL** (write model).
2. Replay event → hitung saldo → **proyeksikan** ke **Elasticsearch** (read model, upsert idempoten).
3. **Query** saldo via `_search` Elasticsearch (bukan dari MySQL) — memisahkan beban baca dari tulis.

> Versi **runnable tanpa infra** (SQLite, konsep sama) ada di folder induk [`../`](..). Bandingkan keduanya untuk melihat bahwa yang berganti hanya *driver/adapter*, bukan konsep.
