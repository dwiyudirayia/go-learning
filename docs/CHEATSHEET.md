# ⚡ Go Cheatsheet (Sintaks Padat)

Rujukan cepat sintaks Go. Semua bisa disalin & dicoba.

## Variabel & konstanta

```go
var x int = 10
var y = 10          // tipe di-infer
z := 10             // short decl (hanya di dalam fungsi)
var a, b = 1, "hi"  // banyak sekaligus
const Pi = 3.14
const ( A = iota; B; C ) // 0,1,2
```

## Tipe dasar

```go
bool
string
int int8 int16 int32 int64      // int = 64-bit di platform modern
uint uint8(byte) ... uint64
float32 float64
complex64 complex128
rune  // = int32, satu titik-kode Unicode
```

Zero value: `0`, `""`, `false`, `nil` (pointer/slice/map/chan/func/interface).

## Kontrol alur

```go
if x > 0 { ... } else if x < 0 { ... } else { ... }
if v, err := f(); err != nil { ... }   // init + kondisi

for i := 0; i < n; i++ { ... }          // klasik
for i := range n { ... }                // Go 1.22+: 0..n-1
for i, v := range s { ... }             // slice/array/string/map/chan
for cond { ... }                        // while
for { ... }                             // loop tak henti

switch x {
case 1, 2: ...        // banyak nilai
case 3: ...; fallthrough
default: ...
}
switch { case x>0: ... }                // tanpa ekspresi = if-else
switch v := x.(type) { case int: ... }  // type switch
```

## Fungsi

```go
func add(a, b int) int { return a + b }
func divmod(a, b int) (int, int) { return a / b, a % b } // multi-return
func sum(nums ...int) int { ... }                        // variadic
func named() (hasil int, err error) { return }           // named return
f := func(x int) int { return x * 2 }                    // closure/anonim
defer cleanup()                                          // jalan saat keluar (LIFO)
```

## Struct & method

```go
type Point struct{ X, Y int }
p := Point{X: 1, Y: 2}
p2 := Point{1, 2}
pp := &Point{1, 2}

func (p Point) Jarak() float64 { ... }  // value receiver
func (p *Point) Geser(dx int)  { p.X += dx } // pointer receiver

type Base struct{ ID int }
type User struct{ Base; Name string } // embedding -> u.ID promoted
```

## Slice

```go
s := []int{1, 2, 3}
s = append(s, 4)
s = append(s, other...)
sub := s[1:3]          // [low:high], berbagi backing array
safe := s[1:3:3]       // three-index: patok cap
n := len(s); c := cap(s)
m := make([]int, 0, 10) // len 0, cap 10
copy(dst, src)
clear(s)               // Go 1.21+
```

## Map

```go
m := map[string]int{"a": 1}
m["b"] = 2
v, ok := m["c"]        // comma-ok
delete(m, "a")
for k, v := range m { ... } // urutan ACAK
set := map[string]struct{}{} // set idiomatik (nilai 0-byte)
```

## String, byte, rune

```go
s := "héllo"
len(s)                          // jumlah BYTE
for i, r := range s { ... }     // r = rune (titik-kode)
b := []byte(s); r := []rune(s)
utf8.RuneCountInString(s)       // jumlah rune
strings.Fields, Split, Join, Contains, Builder
strconv.Itoa, Atoi, ParseFloat
```

## Pointer

```go
x := 10
p := &x     // alamat
*p = 20     // dereference
var q *int  // nil pointer
// Tak ada aritmetika pointer (kecuali unsafe).
```

## Interface & type assertion

```go
type Stringer interface{ String() string }
var i any = "hi"
s, ok := i.(string)   // aman
n := i.(int)          // panic jika salah tipe
```

## Error

```go
if err != nil { return fmt.Errorf("konteks: %w", err) }
var ErrX = errors.New("x")
errors.Is(err, ErrX)
errors.As(err, &target)
errors.Join(e1, e2)
```

## Goroutine, channel, select

```go
go func() { ... }()
ch := make(chan int)      // unbuffered
ch := make(chan int, 5)   // buffered
ch <- v; v := <-ch; v, ok := <-ch
close(ch); for v := range ch { ... }

select {
case v := <-ch1: ...
case ch2 <- x: ...
case <-time.After(time.Second): ...
default: ...              // non-blocking
}

var wg sync.WaitGroup
wg.Add(1); go func(){ defer wg.Done(); ... }(); wg.Wait()
var mu sync.Mutex; mu.Lock(); defer mu.Unlock()
var n atomic.Int64; n.Add(1); n.Load()
```

## Generics

```go
func Max[T cmp.Ordered](a, b T) T { if a > b { return a }; return b }
type Set[T comparable] struct{ m map[T]struct{} }
type Number interface{ ~int | ~float64 }        // type set + tilde
func Sum[T Number](s []T) (t T) { for _, v := range s { t += v }; return }
```

## Paket & impor

```go
package main
import (
    "fmt"
    _ "modernc.org/sqlite" // blank: hanya untuk efek init()
    m "math/rand"          // alias
)
```
