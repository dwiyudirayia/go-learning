// Teknik Advanced Modul 20 (graceful shutdown) — contoh runnable + penjelasan detail.
//
// Mendemokan pola LENGKAP: signal.NotifyContext + srv.Shutdown yang MENUNGGU
// request in-flight selesai. Sinyal SIGTERM dikirim ke diri sendiri agar demo
// jalan otomatis. Jalankan:
//
//	go run ./20-graceful-shutdown/advanced
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// =====================================================================
	// 1. signal.NotifyContext — context yang batal saat SIGINT/SIGTERM
	// =====================================================================
	// Cara idiomatik (Go 1.16+): ctx.Done() menyala saat Ctrl+C atau SIGTERM
	// dari orchestrator (K8s). stop() melepas pendaftaran sinyal.
	//
	// 🔍 Analogi: NotifyContext itu seperti BEL "TOKO AKAN TUTUP" — begitu
	// berbunyi (SIGTERM), semua bagian toko yang mendengarkan bel yang sama
	// (ctx.Done) tahu harus mulai beres-beres, tanpa disuruh satu per satu.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Handler lambat (300ms) untuk membuktikan request in-flight tak terputus.
	mux := http.NewServeMux()
	mux.HandleFunc("/kerja", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(300 * time.Millisecond)
		fmt.Fprintln(w, "pekerjaan selesai penuh")
	})

	// Listen di port acak (127.0.0.1:0) agar tak bentrok.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	srv := &http.Server{Handler: mux}
	go func() {
		if err := srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()
	url := "http://" + ln.Addr().String() + "/kerja"

	// Mulai satu request "in-flight" di latar; ia butuh 300ms.
	hasil := make(chan string, 1)
	go func() {
		resp, err := http.Get(url)
		if err != nil {
			hasil <- "request GAGAL: " + err.Error()
			return
		}
		defer resp.Body.Close()
		b, _ := io.ReadAll(resp.Body)
		hasil <- fmt.Sprintf("request OK (%d): %s", resp.StatusCode, b)
	}()

	// Simulasikan operator/K8s mengirim SIGTERM 100ms kemudian (saat request
	// masih berjalan). Di produksi ini datang dari luar, bukan dikirim sendiri.
	// (os.Process.Signal dipakai alih-alih syscall.Kill agar kode ini tetap
	// ter-compile lintas platform — syscall.Kill tak ada di Windows.)
	time.AfterFunc(100*time.Millisecond, func() {
		fmt.Println("== SIGTERM diterima (disimulasikan) ==")
		p, _ := os.FindProcess(os.Getpid())
		_ = p.Signal(syscall.SIGTERM)
	})

	<-ctx.Done() // tunggu sinyal
	fmt.Println("  mulai graceful shutdown...")

	// =====================================================================
	// 2. Shutdown BERBATAS WAKTU — berhenti terima koneksi baru, TUNGGU in-flight
	// =====================================================================
	// Server berhenti menerima koneksi baru lalu menunggu request yang sedang
	// jalan selesai, sampai batas timeout. Jika lewat, dipaksa berhenti.
	//
	// 🔍 Analogi: graceful shutdown itu seperti RESTORAN JAM TUTUP — pintu
	// dikunci untuk tamu baru, tapi tamu yang sedang makan (request in-flight)
	// dibiarkan menghabiskan piringnya. Timeout = batas kesabaran: lewat jam
	// itu, lampu dimatikan walau masih ada yang mengunyah.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		fmt.Println("  shutdown dipaksa (timeout):", err)
	} else {
		fmt.Println("  server berhenti rapi (in-flight tuntas)")
	}

	// Request yang tadi in-flight tetap mendapat respons penuh -> bukti graceful.
	fmt.Println("  ->", <-hasil)

	// Urutan produksi: (1) readiness=false, (2) Shutdown drain, (3) tutup DB/Redis.
}
