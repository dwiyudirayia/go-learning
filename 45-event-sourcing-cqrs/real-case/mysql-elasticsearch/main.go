// REAL-CASE PRODUKSI Modul 45 (Event Sourcing & CQRS): MySQL (write) + Elasticsearch (read).
//
// Ini pola CQRS sesungguhnya:
//   - WRITE MODEL  : event store append-only di MySQL. UNIQUE(aggregate_id, version)
//     menegakkan optimistic concurrency (transaksi ACID).
//   - READ MODEL   : proyeksi di Elasticsearch -> query kaya (full-text, agregasi,
//     sorting relevansi) yang mahal di SQL. Berkomunikasi via REST
//     HTTP (Elasticsearch = HTTP + JSON) sehingga TANPA client berat.
//   - Read-model DERIVATIF: boleh dibuang & di-rebuild dari event.
//
// Lingkungan WSL ini tak menjalankan MySQL/ES; program AUTO-SKIP (cetak panduan)
// bila env kosong, jadi `go build/vet ./...` tetap hijau. Jalankan NYATA:
//
//	docker compose -f 45-event-sourcing-cqrs/real-case/mysql-elasticsearch/docker-compose.yml up -d
//	MYSQL_DSN='root:secret@tcp(127.0.0.1:3306)/esdemo?parseTime=true' \
//	ES_URL='http://127.0.0.1:9200' \
//	go run ./45-event-sourcing-cqrs/real-case/mysql-elasticsearch
package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// ---------------------------------------------------------------------------
// WRITE MODEL — event store di MySQL
// ---------------------------------------------------------------------------

const schema = `
CREATE TABLE IF NOT EXISTS events (
	seq          BIGINT      NOT NULL AUTO_INCREMENT,
	aggregate_id VARCHAR(64) NOT NULL,
	version      INT         NOT NULL,
	type         VARCHAR(32) NOT NULL,
	amount       BIGINT      NOT NULL,
	created_at   DATETIME    NOT NULL,
	PRIMARY KEY (seq),
	UNIQUE KEY uq_agg_ver (aggregate_id, version)
) ENGINE=InnoDB;`

var ErrKonflik = errors.New("konflik versi (optimistic concurrency)")

type WriteStore struct{ db *sql.DB }

func (w *WriteStore) versiTerkini(ctx context.Context, aggID string) (int, error) {
	var v sql.NullInt64
	err := w.db.QueryRowContext(ctx, `SELECT MAX(version) FROM events WHERE aggregate_id=?`, aggID).Scan(&v)
	if err != nil {
		return 0, err
	}
	if v.Valid {
		return int(v.Int64), nil
	}
	return 0, nil
}

// Append menulis event baru; UNIQUE(aggregate_id, version) menolak penulis basi.
func (w *WriteStore) Append(ctx context.Context, aggID string, expectedVersion int, tipe string, amount int) (int, error) {
	newVersion := expectedVersion + 1
	_, err := w.db.ExecContext(ctx,
		`INSERT INTO events(aggregate_id, version, type, amount, created_at) VALUES(?,?,?,?,?)`,
		aggID, newVersion, tipe, amount, time.Now().UTC())
	if err != nil {
		return 0, fmt.Errorf("%w: %v", ErrKonflik, err)
	}
	return newVersion, nil
}

// Load merekonstruksi saldo aggregate dengan replay event dari MySQL.
func (w *WriteStore) Load(ctx context.Context, aggID string) (saldo, versi int, err error) {
	rows, err := w.db.QueryContext(ctx, `SELECT type, amount FROM events WHERE aggregate_id=? ORDER BY version`, aggID)
	if err != nil {
		return 0, 0, err
	}
	defer rows.Close()
	for rows.Next() {
		var tipe string
		var amount int
		if err := rows.Scan(&tipe, &amount); err != nil {
			return 0, 0, err
		}
		if tipe == "Deposited" {
			saldo += amount
		} else {
			saldo -= amount
		}
		versi++
	}
	return saldo, versi, rows.Err()
}

// ---------------------------------------------------------------------------
// READ MODEL — proyeksi ke Elasticsearch via REST (HTTP + JSON)
// ---------------------------------------------------------------------------

type ReadModel struct {
	baseURL string
	http    *http.Client
	index   string
}

// Project meng-index/meng-update dokumen read-model. PUT /{index}/_doc/{id}
// bersifat upsert -> idempoten (aman diproyeksikan ulang dari event).
func (r *ReadModel) Project(ctx context.Context, aggID string, saldo, versi int) error {
	doc, _ := json.Marshal(map[string]any{
		"aggregate_id": aggID,
		"balance":      saldo,
		"version":      versi,
		"updated_at":   time.Now().UTC(),
	})
	url := fmt.Sprintf("%s/%s/_doc/%s", r.baseURL, r.index, aggID)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(doc))
	req.Header.Set("Content-Type", "application/json")
	return r.do(req, nil)
}

// Search menjalankan query Elasticsearch (di sini: saldo >= min). Contoh kekuatan
// read model: range/agregasi/full-text yang mahal di SQL jadi natural di ES.
func (r *ReadModel) Search(ctx context.Context, minSaldo int) (string, error) {
	q, _ := json.Marshal(map[string]any{
		"query": map[string]any{"range": map[string]any{"balance": map[string]any{"gte": minSaldo}}},
	})
	url := fmt.Sprintf("%s/%s/_search", r.baseURL, r.index)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(q))
	req.Header.Set("Content-Type", "application/json")
	var out bytes.Buffer
	if err := r.do(req, &out); err != nil {
		return "", err
	}
	return out.String(), nil
}

func (r *ReadModel) Refresh(ctx context.Context) error { // paksa ES segera bisa di-search
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, r.baseURL+"/"+r.index+"/_refresh", nil)
	return r.do(req, nil)
}

func (r *ReadModel) do(req *http.Request, out io.Writer) error {
	resp, err := r.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ES %s -> %d: %s", req.URL.Path, resp.StatusCode, b)
	}
	if out != nil {
		_, _ = io.Copy(out, resp.Body)
	}
	return nil
}

func main() {
	dsn := os.Getenv("MYSQL_DSN")
	esURL := os.Getenv("ES_URL")
	if dsn == "" || esURL == "" {
		cetakPanduan()
		return // AUTO-SKIP: tanpa infra, keluar rapi (repo tetap hijau)
	}

	ctx := context.Background()
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(10)
	if err := db.PingContext(ctx); err != nil {
		panic("gagal konek MySQL: " + err.Error())
	}
	if _, err := db.ExecContext(ctx, schema); err != nil {
		panic(err)
	}

	write := &WriteStore{db: db}
	read := &ReadModel{baseURL: esURL, http: &http.Client{Timeout: 5 * time.Second}, index: "accounts"}

	const akun = "acc-100"

	// COMMAND -> event ke MySQL -> proyeksikan state terkini ke Elasticsearch.
	fmt.Println("== 1. command -> write MySQL -> project ES ==")
	for _, ev := range []struct {
		tipe   string
		amount int
	}{{"Deposited", 100}, {"Withdrawn", 30}, {"Deposited", 50}} {
		v, _ := write.versiTerkini(ctx, akun)
		if _, err := write.Append(ctx, akun, v, ev.tipe, ev.amount); err != nil {
			panic(err)
		}
		saldo, versi, _ := write.Load(ctx, akun)
		if err := read.Project(ctx, akun, saldo, versi); err != nil {
			panic(err)
		}
		fmt.Printf("  %s %d -> saldo=%d (v%d) terindeks ke ES\n", ev.tipe, ev.amount, saldo, versi)
	}

	// QUERY dari read model Elasticsearch (bukan dari MySQL).
	fmt.Println("== 2. query dari read model (Elasticsearch) ==")
	_ = read.Refresh(ctx)
	hasil, err := read.Search(ctx, 100)
	if err != nil {
		panic(err)
	}
	fmt.Println("  hasil _search balance>=100:", ringkas(hasil))

	fmt.Println("== selesai: write=MySQL (ACID+optimistic concurrency), read=Elasticsearch (query kaya) ==")
}

func ringkas(s string) string {
	if len(s) > 160 {
		return s[:160] + "…"
	}
	return s
}

func cetakPanduan() {
	fmt.Println("⏭️  DILEWATI: set MYSQL_DSN & ES_URL untuk menjalankan versi nyata.")
	fmt.Println()
	fmt.Println("1) Nyalakan infra:")
	fmt.Println("   docker compose -f 45-event-sourcing-cqrs/real-case/mysql-elasticsearch/docker-compose.yml up -d")
	fmt.Println("2) Jalankan:")
	fmt.Println(`   MYSQL_DSN='root:secret@tcp(127.0.0.1:3306)/esdemo?parseTime=true' \`)
	fmt.Println(`   ES_URL='http://127.0.0.1:9200' \`)
	fmt.Println("   go run ./45-event-sourcing-cqrs/real-case/mysql-elasticsearch")
	fmt.Println()
	fmt.Println("Arsitektur: MySQL = write model (event store, ACID, optimistic concurrency).")
	fmt.Println("            Elasticsearch = read model (query kaya). Read model = derivatif.")
}
