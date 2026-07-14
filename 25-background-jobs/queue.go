// Modul 25 — Background Jobs: worker queue dengan retry, backoff, & idempotensi.
package main

import (
	"sync"
	"sync/atomic"
	"time"
)

// 🔍 Analogi besar: background job itu seperti "TARUH DI KERANJANG, KERJAKAN NANTI". Saat user
// klik "checkout", kita tak menyuruhnya menunggu email struk terkirim — email dimasukkan antrean,
// user langsung dapat respons, lalu WORKER di belakang layar mengirimnya. Kerja lambat (email,
// resize gambar, laporan) dipindah dari "jalur cepat" (request user) ke "jalur belakang".

// Job = unit kerja yang diproses di latar belakang (mis. kirim email, resize gambar).
type Job struct {
	ID         string       // dipakai untuk idempotensi
	Handler    func() error // pekerjaan sebenarnya
	MaxRetries int          // percobaan ulang bila gagal
}

// Queue: antrean job + worker pool yang memprosesnya konkuren.
type Queue struct {
	jobs      chan Job
	workersWG sync.WaitGroup // menunggu worker berhenti
	jobsWG    sync.WaitGroup // menunggu semua job selesai diproses
	backoff   time.Duration

	mu         sync.Mutex
	processed  map[string]bool // ID job yang sudah sukses (idempotensi)
	handlerHit int64           // total pemanggilan handler (untuk membuktikan retry)
	failed     int64           // job yang gagal permanen (dead letter)
}

func NewQueue(workers, buffer int, backoff time.Duration) *Queue {
	q := &Queue{
		jobs:      make(chan Job, buffer),
		backoff:   backoff,
		processed: make(map[string]bool),
	}
	for i := 0; i < workers; i++ {
		q.workersWG.Add(1)
		go q.worker()
	}
	return q
}

// Enqueue menambahkan job ke antrean (non-blocking selama buffer cukup).
func (q *Queue) Enqueue(j Job) {
	q.jobsWG.Add(1)
	q.jobs <- j
}

func (q *Queue) worker() {
	defer q.workersWG.Done()
	for j := range q.jobs {
		q.process(j)
		q.jobsWG.Done()
	}
}

func (q *Queue) process(j Job) {
	// 🔍 Analogi idempotensi: seperti nomor struk unik. Kalau struk #123 sudah diproses, memproses
	// lagi struk #123 TIDAK menambah efek (mis. tak mengirim email dobel / tak menagih 2x). Penting
	// karena antrean bisa mengirim pesan yang sama dua kali ("at-least-once delivery"). ID job = pengaman.
	// IDEMPOTENSI: kalau job dengan ID ini sudah sukses, jangan proses lagi.
	// (Message queue bisa mengirim pesan ganda — handler harus tahan itu.)
	q.mu.Lock()
	done := q.processed[j.ID]
	q.mu.Unlock()
	if done {
		return
	}

	// 🔍 Analogi backoff: seperti mengetuk pintu — kalau tak dibuka, kamu menunggu makin lama sebelum
	// mengetuk lagi (jeda 1x, 2x, 3x...), bukan menggedor terus-menerus. Memberi waktu layanan yang
	// bermasalah untuk pulih & mencegah membanjiri. Job yang gagal terus akhirnya masuk "dead letter"
	// (kotak surat khusus gagal) untuk diselidiki manusia — bukan dibuang diam-diam.
	// RETRY dengan BACKOFF: coba sampai MaxRetries+1 kali.
	for attempt := 1; attempt <= j.MaxRetries+1; attempt++ {
		atomic.AddInt64(&q.handlerHit, 1)
		if err := j.Handler(); err == nil {
			q.mu.Lock()
			q.processed[j.ID] = true
			q.mu.Unlock()
			return
		}
		if attempt <= j.MaxRetries {
			// Backoff bertambah tiap percobaan (linear di sini; produksi: eksponensial).
			time.Sleep(q.backoff * time.Duration(attempt))
		}
	}
	// Gagal permanen -> "dead letter" (di produksi: simpan untuk investigasi/alert).
	atomic.AddInt64(&q.failed, 1)
}

// Wait menunggu semua job yang sudah di-enqueue selesai diproses.
func (q *Queue) Wait() { q.jobsWG.Wait() }

// Stop menutup antrean & menunggu worker berhenti.
func (q *Queue) Stop() {
	close(q.jobs)
	q.workersWG.Wait()
}

func (q *Queue) HandlerHits() int64 { return atomic.LoadInt64(&q.handlerHit) }
func (q *Queue) Failed() int64      { return atomic.LoadInt64(&q.failed) }
func (q *Queue) ProcessedCount() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.processed)
}
