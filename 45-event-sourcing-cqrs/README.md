# 45 — Event Sourcing & CQRS

Lanjutan Modul 31 (saga/outbox). Dua pola arsitektur data yang kuat untuk sistem yang butuh **audit lengkap** & **skala baca-tulis terpisah**. Domain contoh: rekening bank.

Jalankan:
```bash
go run ./45-event-sourcing-cqrs
go test ./45-event-sourcing-cqrs
```

## Event Sourcing — simpan PERISTIWA, bukan STATE

Cara biasa (CRUD): simpan state terkini (`Balance = 120`) dan menimpanya tiap update — **sejarah hilang**.

Event Sourcing: simpan **urutan peristiwa**; state = hasil **memutar ulang** event.
```
Peristiwa: [AccountOpened, MoneyDeposited(150), MoneyWithdrawn(50), MoneyDeposited(20)]
                     └────────── replay ──────────┘ = Balance 120
```
```go
func NewAccountFromEvents(events []Event) *Account {
    a := &Account{}
    for _, e := range events { a.apply(e) } // putar ulang
    return a
}
```

### Command vs Event
- **Command** (imperative, bisa gagal): "Tarik 50". Divalidasi (saldo cukup?).
- **Event** (past tense, fakta): "MoneyWithdrawn(50)". Sudah terjadi, tak bisa dibatalkan.
```go
func (a *Account) Withdraw(amount int) ([]Event, error) {
    if amount > a.Balance { return nil, ErrInsufficient } // validasi
    return []Event{MoneyWithdrawn{a.ID, amount}}, nil       // hasilkan event
}
```
Command handler **tidak** mengubah state langsung — ia menghasilkan event; event yang mengubah state (via `apply`). Test membuktikan: withdraw > saldo ditolak & **tak ada event ditulis**.

### Keuntungan
- **Audit trail lengkap** — output menampilkan seluruh riwayat (untuk compliance, forensik).
- **Time-travel & debug** — rekonstruksi state di titik waktu mana pun.
- **Sumber kebenaran tunggal** — event store append-only (tak pernah UPDATE/DELETE).

## CQRS — Command Query Responsibility Segregation

Pisahkan **model tulis** (command → event) dari **model baca** (query → projeksi).
```
COMMAND  ──► Aggregate ──► Event ──► Event Store
                                        │ (subscribe)
QUERY    ◄── Read Model (Projection) ◄──┘
```
- **Sisi write** (`app.go` BankService): validasi & hasilkan event.
- **Sisi read** (`BalanceProjection`): mendengarkan event, memelihara tampilan cepat (saldo). Bisa dioptimasi khusus baca (denormalisasi, cache, DB terpisah).

Output membuktikan read model & write model **konsisten** (saldo 120 sama).

### Kenapa dipisah?
- Baca & tulis sering punya kebutuhan berbeda (baca banyak & butuh cepat; tulis butuh konsistensi).
- Bisa **skala independen** (banyak read replica, satu write).
- Read model bisa banyak bentuk (satu untuk dashboard, satu untuk pencarian).

## ⚠️ Trade-off
- **Kompleksitas tinggi** — jangan pakai untuk CRUD sederhana.
- **Eventual consistency** — read model sedikit tertinggal dari write.
- **Skema event harus di-versioning** (event lama harus tetap bisa dibaca selamanya).
- **Snapshot** diperlukan bila event per-aggregate sangat banyak (agar replay tak lambat).

## Kapan & Di Mana Dipakai
- Domain dengan audit ketat (keuangan, kesehatan, e-commerce order).
- Sistem kolaboratif, analitik atas riwayat, "undo/redo".
- Kombinasi dengan message queue (Modul 23) & saga (Modul 31) di microservices.

## Latihan
1. Tambah event `AccountClosed` + command `Close` (tolak bila saldo > 0).
2. Tambah read model kedua: daftar transaksi (mutasi) per rekening.
3. Simpan event store ke SQLite (tabel `events`) alih-alih memori.
4. Tambah **snapshot** tiap 100 event agar replay cepat.
5. Publikasikan event ke NATS (Modul 23) agar service lain bereaksi (integrasi).
