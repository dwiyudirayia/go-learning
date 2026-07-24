// golang.org/x/time/rate — membatasi laju permintaan dengan algoritma token bucket.
//
// Jalankan: go run ./libraries/ratelimit
// Test:     go test ./libraries/ratelimit
//
// 🔍 Analogi besar: bayangkan EMBER yang ditetesi air dengan laju tetap, misalnya
// 5 tetes per detik. Setiap permintaan yang masuk harus MENGAMBIL SATU TETES. Kalau
// embernya kosong, permintaan itu ditolak (atau disuruh menunggu sampai ada tetesan baru).
//
// Kenapa pakai ember, bukan sekadar menghitung "maksimal 5 per detik"? Karena ember
// punya KAPASITAS (burst). Kalau layananmu menganggur 10 detik, ember terisi penuh —
// lalu pengguna boleh mengirim beberapa permintaan sekaligus tanpa ditolak. Ini jauh
// lebih ramah daripada aturan kaku "satu per 200 milidetik", yang menolak dua klik
// beruntun sekalipun sistem sedang santai.
package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

func main() {
	fmt.Println("=== golang.org/x/time/rate ===")
	demoTokenBucket()
	demoBurst()
	demoMenunggu()
	demoPerIP()
	demoMiddleware()
}

// ------------------------------------------------------------------
// 1. Token bucket dasar
// ------------------------------------------------------------------

// NewPembatas membuat pembatas: 'perDetik' permintaan/detik dengan kapasitas 'kapasitas'.
//
// 🔍 Analogi dua angka ini:
//
//	perDetik  = KECEPATAN KERAN mengisi ember (laju jangka panjang).
//	kapasitas = UKURAN EMBER (burst: berapa banyak yang boleh menumpuk saat senggang).
//
// Kesalahan yang sering terjadi: menyetel kapasitas = 1. Itu memaksa jarak antar
// permintaan benar-benar rata, sehingga dua klik yang wajar pun ditolak.
// Aturan praktis: kapasitas kira-kira sebesar lonjakan wajar yang ingin kamu maafkan.
func NewPembatas(perDetik float64, kapasitas int) *rate.Limiter {
	return rate.NewLimiter(rate.Limit(perDetik), kapasitas)
}

// NewPembatasInterval membuat pembatas dengan cara baca yang lebih manusiawi:
// "satu permintaan tiap X waktu".
func NewPembatasInterval(interval time.Duration, kapasitas int) *rate.Limiter {
	return rate.NewLimiter(rate.Every(interval), kapasitas)
}

// HitungLolos menghitung berapa permintaan yang diizinkan bila 'jumlah' permintaan
// datang serentak pada saat 'saat'.
//
// 🔍 Analogi: memakai AllowN(saat, 1) alih-alih Allow() itu seperti memakai JAM PALSU
// yang bisa kita putar sesuka hati. Test jadi pasti hasilnya dan berjalan seketika —
// tak perlu benar-benar menunggu satu detik demi satu detik.
func HitungLolos(l *rate.Limiter, saat time.Time, jumlah int) int {
	lolos := 0
	for range jumlah {
		if l.AllowN(saat, 1) {
			lolos++
		}
	}
	return lolos
}

func demoTokenBucket() {
	fmt.Println("\n-- Token bucket dasar --")

	l := NewPembatas(5, 5) // 5 per detik, ember muat 5
	kini := time.Now()

	fmt.Printf("   10 permintaan serentak -> lolos %d\n", HitungLolos(l, kini, 10))
	fmt.Printf("   1 detik kemudian, 10 lagi -> lolos %d\n",
		HitungLolos(l, kini.Add(time.Second), 10))
	fmt.Printf("   2 detik kemudian, 10 lagi -> lolos %d\n",
		HitungLolos(l, kini.Add(3*time.Second), 10))
}

func demoBurst() {
	fmt.Println("\n-- Pengaruh ukuran ember (burst) --")

	kini := time.Now()
	for _, kapasitas := range []int{1, 3, 10} {
		l := NewPembatas(5, kapasitas)
		fmt.Printf("   kapasitas %2d -> dari 10 permintaan serentak, lolos %d\n",
			kapasitas, HitungLolos(l, kini, 10))
	}
}

// ------------------------------------------------------------------
// 2. Menunggu vs menolak
// ------------------------------------------------------------------

// 🔍 Analogi dua sikap terhadap antrean:
//
//	Allow() = SATPAM DI PINTU KELAB: penuh? "Maaf, tidak bisa masuk." Pemanggil ditolak
//	          seketika (HTTP 429). Cocok untuk API publik — lebih baik menolak cepat
//	          daripada membuat ribuan koneksi menggantung.
//	Wait()  = NOMOR ANTREAN DI BANK: kamu tetap dilayani, tapi harus menunggu giliran.
//	          Cocok untuk pekerjaan latar (worker) yang memang tak buru-buru, mis.
//	          menghormati batas laju API pihak ketiga.
//
// Jebakan Wait(): SELALU pakai context dengan timeout. Tanpa itu, worker-mu bisa
// menunggu berjam-jam tanpa ada yang tahu.

// KerjaDenganAntrean menjalankan 'jumlah' pekerjaan sambil menghormati pembatas.
// Mengembalikan berapa yang berhasil sebelum context habis.
func KerjaDenganAntrean(ctx context.Context, l *rate.Limiter, jumlah int) (int, error) {
	selesai := 0
	for range jumlah {
		if err := l.Wait(ctx); err != nil {
			return selesai, fmt.Errorf("berhenti setelah %d pekerjaan: %w", selesai, err)
		}
		selesai++
	}
	return selesai, nil
}

func demoMenunggu() {
	fmt.Println("\n-- Wait: mengantre, bukan ditolak --")

	l := NewPembatasInterval(20*time.Millisecond, 1) // 50 per detik
	ctx, batal := context.WithTimeout(context.Background(), 120*time.Millisecond)
	defer batal()

	mulai := time.Now()
	selesai, err := KerjaDenganAntrean(ctx, l, 100)
	fmt.Printf("   %d pekerjaan selesai dalam %v\n", selesai,
		time.Since(mulai).Round(10*time.Millisecond))
	if err != nil {
		fmt.Println("   berhenti karena:", err)
	}
}

// ------------------------------------------------------------------
// 3. Pembatas per pengguna
// ------------------------------------------------------------------

// 🔍 Analogi: satu pembatas untuk SELURUH server itu seperti satu gelas air untuk seluruh
// kantor — satu orang haus bisa menghabiskan jatah semua orang. Yang biasanya kita mau
// adalah SATU EMBER PER PENGGUNA (per IP, per kunci API), sehingga penyalahguna hanya
// merugikan dirinya sendiri.
//
// Jebakan besar: map yang isinya bertambah terus tanpa pernah dibersihkan adalah
// KEBOCORAN MEMORI. Setiap IP yang pernah singgah akan tersimpan selamanya. Karena itu
// ada Bersihkan() — di produksi dipanggil berkala oleh goroutine latar.

// PembatasPerKunci menyimpan satu pembatas untuk tiap kunci (IP/pengguna).
type PembatasPerKunci struct {
	mu        sync.Mutex
	pembatas  map[string]*entriPembatas
	perDetik  float64
	kapasitas int
	sekarang  func() time.Time
}

type entriPembatas struct {
	limiter  *rate.Limiter
	terakhir time.Time
}

func NewPembatasPerKunci(perDetik float64, kapasitas int) *PembatasPerKunci {
	return &PembatasPerKunci{
		pembatas:  make(map[string]*entriPembatas),
		perDetik:  perDetik,
		kapasitas: kapasitas,
		sekarang:  time.Now,
	}
}

// DenganJam mengganti sumber waktu — dipakai test agar hasilnya pasti.
func (p *PembatasPerKunci) DenganJam(jam func() time.Time) *PembatasPerKunci {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.sekarang = jam
	return p
}

// Izinkan memeriksa apakah kunci ini boleh mengirim satu permintaan lagi.
func (p *PembatasPerKunci) Izinkan(kunci string) bool {
	p.mu.Lock()
	e, ada := p.pembatas[kunci]
	if !ada {
		e = &entriPembatas{limiter: NewPembatas(p.perDetik, p.kapasitas)}
		p.pembatas[kunci] = e
	}
	kini := p.sekarang()
	e.terakhir = kini
	p.mu.Unlock()

	return e.limiter.AllowN(kini, 1)
}

// Bersihkan membuang entri yang tak dipakai lebih lama dari 'usia'.
func (p *PembatasPerKunci) Bersihkan(usia time.Duration) int {
	p.mu.Lock()
	defer p.mu.Unlock()

	batas := p.sekarang().Add(-usia)
	dibuang := 0
	for k, e := range p.pembatas {
		if e.terakhir.Before(batas) {
			delete(p.pembatas, k)
			dibuang++
		}
	}
	return dibuang
}

// Jumlah mengembalikan banyaknya kunci yang sedang dilacak.
func (p *PembatasPerKunci) Jumlah() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.pembatas)
}

func demoPerIP() {
	fmt.Println("\n-- Pembatas per IP --")

	p := NewPembatasPerKunci(2, 2)
	for _, ip := range []string{"1.1.1.1", "1.1.1.1", "1.1.1.1", "2.2.2.2"} {
		fmt.Printf("   %s -> %t\n", ip, p.Izinkan(ip))
	}
	fmt.Printf("   kunci dilacak: %d\n", p.Jumlah())
}

// ------------------------------------------------------------------
// 4. Middleware HTTP
// ------------------------------------------------------------------

// MiddlewarePembatas menolak permintaan berlebih dengan status 429.
//
// 🔍 Analogi header Retry-After: menolak tanpa memberi tahu "coba lagi kapan" itu seperti
// menutup pintu tanpa penjelasan — klien akan langsung mencoba lagi dan lagi, memperparah
// keadaan. Retry-After adalah papan "buka lagi 1 menit lagi" yang membuat klien yang
// sopan menunggu dengan benar.
func MiddlewarePembatas(p *PembatasPerKunci, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !p.Izinkan(KunciDariRequest(r)) {
			w.Header().Set("Retry-After", "1")
			http.Error(w, "terlalu banyak permintaan", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// KunciDariRequest mengambil identitas pemanggil.
//
// PERINGATAN produksi: di belakang proxy/load balancer, r.RemoteAddr adalah alamat
// PROXY-nya, bukan pengguna — semua pengguna akan berbagi satu ember. Di sana kamu perlu
// membaca X-Forwarded-For, TAPI hanya bila proxy-nya tepercaya (header itu mudah dipalsukan
// bila server terbuka langsung ke internet).
func KunciDariRequest(r *http.Request) string {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

func demoMiddleware() {
	fmt.Println("\n-- Middleware HTTP --")

	p := NewPembatasPerKunci(1, 3)
	h := MiddlewarePembatas(p, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintln(w, "ok")
	}))
	_ = h // handler siap dipasang ke mux; peragaan lengkapnya ada di test

	for i := 1; i <= 5; i++ {
		fmt.Printf("   permintaan %d -> diizinkan? %t\n", i, p.Izinkan("10.0.0.1"))
	}
}
