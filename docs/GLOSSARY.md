# 📖 Glosarium Istilah Go

Definisi singkat istilah yang sering muncul. Diurutkan agar mudah dipindai.

### Bahasa & tipe
- **Zero value** — nilai default tipe tanpa inisialisasi (`0`, `""`, `false`, `nil`). Idiomnya: zero value harus langsung berguna.
- **Rune** — alias `int32`, satu titik-kode Unicode. `len(string)` menghitung **byte**, bukan rune.
- **Slice** — pandangan (view) atas **backing array**: `{pointer, len, cap}`. Beberapa slice bisa berbagi array yang sama.
- **Backing array** — array sebenarnya di balik slice. `append` bisa mengalokasikan yang baru bila `cap` habis.
- **iota** — penghitung konstanta dalam blok `const`, mulai 0, naik 1 tiap baris. Untuk enum & bitmask.
- **Struct tag** — metadata string di field struct (mis. `` `json:"name"` ``) dibaca via refleksi.
- **Embedding** — menyematkan tipe dalam struct/interface; method/field "dipromosikan". Reuse ala Go (bukan pewarisan).

### Fungsi & method
- **Receiver** — "penerima" method: `func (r T) M()`. **Value receiver** dapat salinan; **pointer receiver** (`*T`) bisa mengubah & menghindari copy.
- **Method set** — kumpulan method milik sebuah tipe. Method pointer-receiver hanya di method set `*T`, memengaruhi pemenuhan interface.
- **Closure** — fungsi anonim yang "menangkap" variabel dari lingkup luarnya.
- **Named return** — nilai kembalian bernama; bisa dimodifikasi di `defer`.
- **Variadic** — parameter `...T` menerima 0..n argumen sebagai slice.
- **defer** — menjadwalkan pemanggilan saat fungsi keluar (LIFO); argumen dievaluasi saat baris ditulis.

### Interface & error
- **Interface** — kumpulan method; dipenuhi **implisit**. Nilainya menyimpan `(tipe, nilai)`.
- **Typed nil** — pointer nil bertipe yang disimpan di interface → interface `!= nil` (jebakan).
- **Type assertion / type switch** — mengekstrak tipe konkret dari interface (`x.(T)`, `switch x.(type)`).
- **Stringer** — interface `String() string`; otomatis dipakai `fmt`/`%v`.
- **Sentinel error** — nilai error tetap (`var ErrX = errors.New(...)`) dicek dengan `errors.Is`.
- **Error wrapping** — `fmt.Errorf("...: %w", err)` mempertahankan rantai untuk `errors.Is/As`.
- **panic / recover** — panic menghentikan alur normal; `recover` (dalam `defer`) menangkapnya. Bukan untuk kontrol alur biasa.

### Konkurensi
- **Goroutine** — unit eksekusi konkuren yang sangat ringan, dijadwal oleh runtime Go.
- **Channel** — pipa berjenis untuk komunikasi antar-goroutine (unbuffered = bersinkron; buffered = ada penyangga).
- **select** — memilih di antara beberapa operasi channel yang siap; `default` membuatnya non-blocking.
- **WaitGroup / Mutex / atomic** — primitif `sync` untuk menunggu, mengunci, dan operasi atomik.
- **context.Context** — membawa pembatalan, deadline, & nilai lintas panggilan; jadi parameter pertama.
- **Data race** — dua goroutine mengakses memori sama tanpa sinkronisasi (min. satu menulis) → bug. Dideteksi `-race`.
- **Happens-before** — jaminan urutan memori; hanya diberikan oleh channel/`sync`/`atomic`.
- **Goroutine leak** — goroutine yang tak pernah selesai (mis. blok di channel selamanya).

### Generics
- **Type parameter** — parameter tipe pada fungsi/tipe: `func F[T any](...)`.
- **Constraint** — interface yang membatasi tipe yang boleh (mis. `comparable`, `cmp.Ordered`, union `~int | ~float64`).
- **`~` (tilde)** — mencakup semua tipe dengan *underlying type* tsb (mis. `~int` termasuk `type MyInt int`).
- **`iter.Seq[T]`** — iterator lazy (range-over-func, Go 1.23).

### Runtime & tooling
- **GMP** — model scheduler Go: **G**oroutine dijadwal ke **M** (thread OS) lewat **P** (processor logis). `GOMAXPROCS` = jumlah P.
- **GC** — garbage collector concurrent (tri-color mark-sweep); disetel `GOGC` & `GOMEMLIMIT`.
- **Escape analysis** — analisa compiler apakah nilai "lolos" ke heap (butuh GC) vs cukup di stack.
- **Build tag** — `//go:build ...` menentukan file ikut dikompilasi untuk platform/fitur tertentu.
- **go:embed** — menyematkan file ke dalam biner saat compile.
- **Module** — unit versi dependensi (`go.mod`); `go.sum` mengunci hash integritas.
- **Vendoring** — menyalin dependensi ke folder `vendor/` (opsional).

### Arsitektur (dipakai di modul backend)
- **Layered / hexagonal** — pemisahan handler → service → repository; domain di pusat, infra di luar (Dependency Inversion).
- **Port & adapter** — *port* = interface kebutuhan domain; *adapter* = implementasi konkret (Postgres, Redis, dll).
- **Sentinel/DTO/idempoten/at-least-once** — lihat modul terkait (05, 15, 23, 45) untuk konteks penuh.
