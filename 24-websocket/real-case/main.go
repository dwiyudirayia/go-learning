// REAL-CASE Modul 24 (websocket) — REDIS PUB/SUB sebagai BACKPLANE fan-out.
//
// Masalah nyata: bila WebSocket server jalan >1 instance (di belakang load
// balancer), pesan broadcast dari instance A tak sampai ke client yang terhubung
// ke instance B. Solusi produksi: Redis Pub/Sub sebagai "backplane" — tiap
// instance mem-publish & men-subscribe channel yang sama.
//
// Demo ini mensimulasikan DUA instance yang berbagi Redis. Auto-skip bila
// REDIS_ADDR kosong. Jalankan nyata:
//
//	docker compose -f 24-websocket/real-case/docker-compose.yml up -d
//	REDIS_ADDR=127.0.0.1:6379 go run ./24-websocket/real-case
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

func main() {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		fmt.Println("⏭️  DILEWATI: set REDIS_ADDR untuk versi nyata.")
		fmt.Println("   docker compose -f 24-websocket/real-case/docker-compose.yml up -d")
		fmt.Println("   REDIS_ADDR=127.0.0.1:6379 go run ./24-websocket/real-case")
		return
	}
	ctx := context.Background()

	// "Instance A" dan "Instance B" = dua koneksi Redis berbeda (seolah dua pod).
	instA := redis.NewClient(&redis.Options{Addr: addr})
	instB := redis.NewClient(&redis.Options{Addr: addr})
	defer instA.Close()
	defer instB.Close()
	if err := instA.Ping(ctx).Err(); err != nil {
		panic("gagal konek Redis: " + err.Error())
	}

	const channel = "room:umum"

	// Instance B punya client WebSocket yang subscribe ke room.
	sub := instB.Subscribe(ctx, channel)
	defer sub.Close()
	if _, err := sub.Receive(ctx); err != nil { // tunggu subscription siap
		panic(err)
	}
	pesanMasuk := sub.Channel()

	// Instance A menerima pesan (mis. via HTTP/WS) lalu MEM-PUBLISH ke backplane.
	fmt.Println("== Redis Pub/Sub backplane (fan-out lintas instance) ==")
	for _, teks := range []string{"halo semua", "ada update"} {
		if err := instA.Publish(ctx, channel, teks).Err(); err != nil {
			panic(err)
		}
		select {
		case m := <-pesanMasuk: // sampai ke client di INSTANCE LAIN
			fmt.Printf("  instance A publish %q -> instance B terima %q\n", teks, m.Payload)
		case <-time.After(2 * time.Second):
			fmt.Println("  timeout menunggu pesan")
		}
	}
	fmt.Println("  => broadcast tersebar walau client ada di pod berbeda.")
}
