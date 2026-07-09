// Teknik Advanced Modul 09 (stdlib) — contoh runnable + penjelasan detail.
//
// Jalankan: go run ./09-stdlib/advanced
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"
)

// ===========================================================================
// 1. Custom MarshalJSON — kendali penuh bentuk JSON sebuah tipe
// ===========================================================================
// Epoch adalah time.Time yang ingin kita serialisasikan sebagai UNIX seconds
// (angka), bukan string RFC3339 default. Implement json.Marshaler.
type Epoch struct{ time.Time }

func (e Epoch) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%d", e.Unix())), nil
}

// Event memakai json.RawMessage untuk MENUNDA decode field "payload":
// bentuknya berbeda-beda tergantung "type", jadi kita simpan mentah dulu.
type Event struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

func main() {
	// =====================================================================
	// 1. Custom marshaler
	// =====================================================================
	b, _ := json.Marshal(Epoch{time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)})
	fmt.Println("== 1. custom MarshalJSON ==")
	fmt.Println("  Epoch -> JSON:", string(b)) // angka unix, bukan string tanggal

	// =====================================================================
	// 2. json.RawMessage — decode dua tahap (payload polimorfik)
	// =====================================================================
	raw := `{"type":"login","payload":{"user":"budi","ip":"10.0.0.1"}}`
	var ev Event
	_ = json.Unmarshal([]byte(raw), &ev)
	fmt.Println("== 2. json.RawMessage ==")
	fmt.Printf("  type=%s, payload mentah=%s\n", ev.Type, ev.Payload)
	// Baru sekarang decode payload sesuai type-nya:
	if ev.Type == "login" {
		var p struct{ User, IP string }
		_ = json.Unmarshal(ev.Payload, &p)
		fmt.Printf("  payload ter-decode -> user=%s ip=%s\n", p.User, p.IP)
	}

	// =====================================================================
	// 3. json.Decoder streaming + DisallowUnknownFields (validasi ketat)
	// =====================================================================
	// Decoder membaca aliran JSON bertubi tanpa memuat semua ke memori.
	// DisallowUnknownFields menolak field asing -> menangkap typo/kontrak salah.
	stream := `{"nama":"A"} {"nama":"B"} {"nama":"C","umur":9}`
	dec := json.NewDecoder(strings.NewReader(stream))
	dec.DisallowUnknownFields()
	fmt.Println("== 3. Decoder streaming + strict ==")
	for {
		var row struct {
			Nama string `json:"nama"`
		}
		err := dec.Decode(&row)
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("  ditolak (field asing):", err) // "umur" tak dikenal
			break
		}
		fmt.Println("  baca:", row.Nama)
	}

	// =====================================================================
	// 4. Komposisi io: TeeReader + MultiWriter + LimitReader
	// =====================================================================
	// TeeReader  : membaca sambil menyalin (mis. hitung hash sambil kirim body).
	// MultiWriter: satu Write menyebar ke banyak tujuan (mis. file + stdout).
	// LimitReader: batasi jumlah byte yang boleh dibaca (cegah abuse memori).
	src := strings.NewReader("HALO-DUNIA-INI-PANJANG")
	var salinan bytes.Buffer
	tee := io.TeeReader(src, &salinan) // apa pun yang dibaca dari tee, tersalin ke salinan

	var a, b2 bytes.Buffer
	multi := io.MultiWriter(&a, &b2)

	// Hanya baca 4 byte pertama lewat LimitReader, tulis ke dua buffer sekaligus.
	_, _ = io.Copy(multi, io.LimitReader(tee, 4))
	fmt.Println("== 4. io composition ==")
	fmt.Printf("  4 byte pertama -> a=%q b=%q | tee menyalin=%q\n",
		a.String(), b2.String(), salinan.String())

	// =====================================================================
	// 5. time.Timer + Stop() — cegah kebocoran timer
	// =====================================================================
	// Timer/Ticker yang tak di-Stop membiarkan resource menggantung. Pola aman:
	// buat timer, dan Stop() saat tak lagi dibutuhkan.
	t := time.NewTimer(50 * time.Millisecond)
	defer t.Stop()
	select {
	case <-t.C:
		fmt.Println("== 5. time.Timer ==")
		fmt.Println("  timer berbunyi")
	case <-time.After(time.Second):
		fmt.Println("  timeout luar (tak terjadi)")
	}
}
