// Modul 23 — Message Queue MULTI-BROKER dengan skema resiliensi.
//
// Kode aplikasi memakai interface Broker, sehingga bisa berganti
// NATS / RabbitMQ / Kafka tanpa mengubah logika bisnis (dependency inversion,
// lihat Modul 4 & 29).
package main

import (
	"context"
	"errors"
)

// 🔍 Analogi besar: message queue itu seperti KANTOR POS antar-aplikasi. Pengirim (producer)
// menaruh surat (pesan) di kotak (topic) lalu PERGI — tak menunggu penerima siap. Nanti penerima
// (consumer) mengambil & memprosesnya. Gunanya: memisahkan pengirim & penerima (decoupling), tahan
// lonjakan (surat mengantre bila penerima sibuk), dan tahan mati (surat tak hilang saat 1 pihak restart).
//
// 🔍 Analogi ack/nack: "ack" = tanda tangan bukti surat SUDAH beres. "nack" = "gagal proses, tolong
// kirim ULANG nanti". Karena ada bukti terima, tak ada pesan yang diam-diam hilang.

// Handler memproses satu pesan.
//   - return nil   -> pesan dianggap SELESAI (ack).
//   - return error -> pesan gagal, broker bisa mengirim ULANG (nack/requeue).
type Handler func(ctx context.Context, data []byte) error

// Broker = abstraksi message broker.
type Broker interface {
	// Publish mengirim pesan ke topic/subject.
	Publish(ctx context.Context, topic string, data []byte) error

	// 🔍 Analogi group: bayangkan siaran vs pembagian tugas.
	//   - group kosong (fan-out) = SIARAN RADIO: semua pendengar dapat lagu yang sama.
	//   - group sama (load balancing) = TIM CS: satu telepon masuk cuma dijawab SATU orang di tim,
	//     supaya beban terbagi. Menambah anggota tim = memproses lebih banyak pesan paralel.
	// Subscribe menjalankan handler untuk tiap pesan (NON-BLOCKING).
	//   - group == "" -> setiap konsumen menerima SEMUA pesan (broadcast/fan-out).
	//   - group != "" -> load balancing antar konsumen dalam grup yang sama.
	//
	// Konsumen berjalan di latar belakang dengan RECONNECT otomatis; berhenti
	// saat ctx dibatalkan atau Close() dipanggil.
	Subscribe(ctx context.Context, topic, group string, handler Handler) error

	// Close menutup koneksi & menghentikan semua konsumen.
	Close() error
}

var (
	// ErrDisconnected: operasi gagal karena koneksi ke broker terputus.
	ErrDisconnected = errors.New("broker: tidak terhubung")
	// ErrClosed: broker sudah ditutup.
	ErrClosed = errors.New("broker: sudah ditutup")
)
