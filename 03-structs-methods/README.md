# 03 — Struct, Method, Pointer vs Value & Embedding

Jalankan:
```bash
go run ./03-structs-methods
```

Go **tidak punya class/inheritance**. Sebagai gantinya: **struct** (data) + **method** (perilaku) + **embedding** (komposisi). Ini cara Go memodelkan objek.

## 1. Struct

```go
type User struct {
	ID    int
	Name  string
	Email string
}

u := User{ID: 1, Name: "Ana"}     // literal by field (disarankan)
u2 := User{2, "Budi", "b@x.id"}    // positional (rapuh, hindari)
```
- **Zero value** struct: semua field jadi zero value-nya (`0`, `""`, dst). Struct siap pakai tanpa konstruktor.
- Struct **dibandingkan dengan `==`** kalau semua field-nya comparable.
- Struct adalah **value type**: di-copy saat di-assign / dikirim ke fungsi.

## 2. Method: value receiver vs pointer receiver (INTI MODUL INI)

```go
func (u User) FullLabel() string { ... }   // value receiver: dapat SALINAN u
func (u *User) SetName(n string) { u.Name = n }  // pointer receiver: bisa MENGUBAH u asli
```

**Aturan praktis memilih receiver:**
- Pakai **pointer receiver** (`*T`) jika: method **mengubah** state, atau struct **besar** (hindari copy mahal), atau butuh konsisten (lihat aturan konsistensi).
- Pakai **value receiver** (`T`) untuk tipe kecil & immutable (mis. `time.Time`).
- **Konsistensi:** kalau ada satu method pakai pointer receiver, buat **semua** method tipe itu pakai pointer receiver.

> Go otomatis mengambil alamat: kalau `u` addressable, `u.SetName("x")` sama dengan `(&u).SetName("x")`.

## 3. Konstruktor: pola `NewXxx`

Go tak punya constructor khusus. Konvensinya fungsi `NewXxx` yang mengembalikan `*T` (atau `T`) yang sudah valid:
```go
func NewUser(name, email string) *User { ... }
```

## 4. Embedding (komposisi, bukan pewarisan)

Sematkan tipe lain **tanpa nama field** → field & method-nya "dipromosikan" ke luar:
```go
type Admin struct {
	User          // embedded
	Level int
}
a.Name          // promoted dari User
a.FullLabel()   // method User ikut terpromosi
```
Ini **komposisi**: `Admin` "punya" `User`, bukan "adalah turunan" `User`. Kalau ada nama bentrok, yang terluar menang (bisa override).

## 5. Struct tag (sekilas)

Metadata string di belakang field, dibaca lewat reflection (dipakai encoding/json, validator, ORM):
```go
type Product struct {
	Name  string `json:"name"`
	Price int    `json:"price"`
}
```
Detailnya di modul JSON & database nanti.

## Latihan
1. Buat struct `Rectangle{Width, Height float64}` dengan method `Area()` dan `Perimeter()` (value receiver).
2. Tambah method `Scale(factor float64)` yang mengubah ukuran rectangle (pointer receiver). Buktikan perubahannya bertahan.
3. Buat konstruktor `NewRectangle(w, h float64) (*Rectangle, error)` yang menolak nilai <= 0.
4. Buat struct `Account{Balance int}` dengan `Deposit` & `Withdraw` (Withdraw menolak saldo kurang, kembalikan error).
5. Buat `type Employee struct { Person; Salary int }` dengan `Person{Name string}` yang punya method `Greet()`. Panggil `emp.Greet()` (promoted) dan tambah `Greet()` khusus Employee (override).

Kerjakan di `03-structs-methods/jawaban-saya/`, lalu minta saya review.

## Kapan & Di Mana Dipakai (Studi Kasus Nyata)

| Konsep | Cara pakai di dunia nyata | Contoh kasus |
|--------|---------------------------|--------------|
| **Struct** | Model domain, DTO, config | `User`, `Order`, `Product`; body request/response API; struct konfigurasi app |
| **Pointer receiver** | Method yang mengubah state / struct besar | `svc.Register(...)`, `cart.AddItem(...)`, method pada service/repository |
| **Value receiver** | Tipe kecil & immutable | `Money`, `Coordinate`, `time.Time` |
| **Konstruktor `NewXxx`** | Dependency injection & objek yang valid sejak lahir | `NewUserService(db, logger)`, `NewServer(cfg)` |
| **Embedding** | Komposisi: perluas/gabung perilaku tanpa pewarisan | Sematkan `sync.Mutex` ke struct agar dapat `Lock()`; base repository; embed `*sql.DB` |

**Contoh nyata — service dengan dependency injection (pola paling umum di backend Go):**
```go
type UserService struct {
    repo   UserRepository // dependency (interface, lihat Modul 4)
    logger *log.Logger
}

func NewUserService(repo UserRepository, l *log.Logger) *UserService {
    return &UserService{repo: repo, logger: l}
}

func (s *UserService) Register(name string) (*User, error) { // pointer receiver
    // ... pakai s.repo, s.logger
}
```

**Contoh nyata — embedding `sync.Mutex`:**
```go
type Counter struct {
    sync.Mutex // embedded -> Counter langsung punya Lock()/Unlock()
    count int
}
func (c *Counter) Inc() { c.Lock(); defer c.Unlock(); c.count++ }
```

**Cocok dipakai saat:** memodelkan "benda" dalam sistemmu (entity/model), dan membangun service/repository. Pola `struct + NewXxx + pointer method` adalah **tulang punggung** hampir semua aplikasi backend Go.

---

## 🚀 Teknik Advanced (Level Up)
> 💻 **Contoh runnable + komentar detail** untuk semua teknik di bawah ada di folder [`advanced/`](advanced). Jalankan: `go run ./03-structs-methods/advanced`


- **Method value vs method expression** — `f := t.Method` (terikat ke `t`) vs `f := T.Method` (butuh receiver sebagai arg pertama: `f(t)`). Berguna untuk higher-order function.
- **Method set & satisfaction interface** — method dengan *pointer receiver* hanya masuk method set `*T`, bukan `T`. Artinya nilai `T` (bukan `&T`) mungkin **tidak** memenuhi interface. Sumber error umum.
- **Field ordering & padding** — urutan field memengaruhi ukuran struct karena alignment. Susun dari lebar terbesar ke terkecil bisa hemat memori — lihat [[42-go-internals]] (`BadStruct` 24B vs `GoodStruct` 16B).
- **Struct comparability** — struct bisa jadi *map key* jika semua field comparable. Struct dengan slice/map/func **tidak** comparable (panic saat `==`).
- **Embedding & override** — method promosi dari tipe tersemat bisa "ditimpa" dengan mendefinisikan method bernama sama di outer struct. Ambiguitas (dua embed punya method sama) → wajib eksplisit.
- **`struct{}` zero-size** — dipakai untuk set & channel sinyal (`chan struct{}`).
