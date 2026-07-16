// Teknik Advanced Modul 24 (websocket) — contoh runnable + penjelasan detail.
//
// Round-trip WebSocket IN-PROCESS (httptest) dengan read/write deadline &
// graceful close handshake. Jalankan:
//
//	go run ./24-websocket/advanced
package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/coder/websocket"
)

func main() {
	// =====================================================================
	// SERVER: echo dengan DEADLINE per operasi baca
	// =====================================================================
	// Tiap Read dibungkus context berbatas waktu -> koneksi "zombie" (client
	// diam selamanya) tak menggantung goroutine server tanpa batas.
	//
	// 🔍 Analogi: read deadline itu seperti PENJAGA WARTEL jaman dulu — tiap
	// penelepon diberi jatah waktu; yang diam saja tanpa bicara (koneksi
	// zombie) diputus sambungannya, supaya bilik (goroutine) tak terkunci
	// selamanya oleh satu orang.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}
		defer c.CloseNow()
		for {
			readCtx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
			typ, data, err := c.Read(readCtx)
			cancel()
			if err != nil {
				return // client menutup / timeout -> keluar
			}
			if string(data) == "bye" {
				// Client minta berhenti -> tutup rapi dgn kode status normal.
				_ = c.Close(websocket.StatusNormalClosure, "pamit")
				return
			}
			// echo kembali dengan prefix
			_ = c.Write(r.Context(), typ, append([]byte("echo: "), data...))
		}
	}))
	defer srv.Close()

	ctx := context.Background()

	// =====================================================================
	// CLIENT: dial in-process (skema http -> ws)
	// =====================================================================
	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")
	c, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		panic(err)
	}
	defer c.CloseNow()

	// =====================================================================
	// 1. Kirim/terima dengan write deadline
	// =====================================================================
	fmt.Println("== 1. echo round-trip ==")
	for _, msg := range []string{"halo", "dunia"} {
		writeCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		if err := c.Write(writeCtx, websocket.MessageText, []byte(msg)); err != nil {
			cancel()
			panic(err)
		}
		_, resp, err := c.Read(writeCtx)
		cancel()
		if err != nil {
			panic(err)
		}
		fmt.Printf("  kirim %q -> terima %q\n", msg, resp)
	}

	// =====================================================================
	// 2. GRACEFUL CLOSE HANDSHAKE — membaca STATUS penutupan dari peer
	// =====================================================================
	// Client kirim "bye"; server menutup dgn kode status. Read berikutnya
	// mengembalikan error penutupan yang bisa dibaca kodenya via CloseStatus
	// (bukan sekadar "koneksi putus"). Kode 1000 = normal.
	//
	// 🔍 Analogi: close handshake itu seperti PAMIT SEBELUM PULANG — tamu
	// bilang "saya pamit" (bye) dan tuan rumah mengantar ke pintu sambil
	// menyebut alasannya (kode status 1000 = baik-baik). Tanpa handshake,
	// tamu tiba-tiba hilang dan tuan rumah cuma tahu "koneksi putus" tanpa
	// tahu itu wajar atau musibah.
	fmt.Println("== 2. graceful close handshake ==")
	closeCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	_ = c.Write(closeCtx, websocket.MessageText, []byte("bye"))
	_, _, err = c.Read(closeCtx)
	status := websocket.CloseStatus(err)
	fmt.Printf("  server menutup dgn status %d (normal? %v)\n",
		status, status == websocket.StatusNormalClosure)

	// Catatan HEARTBEAT & HUB: di produksi, kirim ping/pong berkala + read
	// deadline untuk deteksi koneksi mati. Untuk banyak client, satu goroutine
	// HUB kelola register/unregister/broadcast; tiap koneksi punya goroutine
	// baca & tulis TERPISAH (WebSocket tak aman ditulis konkuren). Untuk aliran
	// satu arah server->client, SSE lebih sederhana daripada WebSocket.
}
