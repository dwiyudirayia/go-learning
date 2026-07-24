package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/rs/zerolog"
)

// baris menguraikan satu baris JSON log jadi map agar mudah diperiksa.
//
// Inilah keuntungan nyata log terstruktur: log-nya bisa DIUJI. Dengan fmt.Println
// kamu cuma bisa mencocokkan potongan teks; dengan JSON kamu memeriksa kolom.
func baris(t *testing.T, s string) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(s)), &m); err != nil {
		t.Fatalf("keluaran bukan JSON valid (%q): %v", s, err)
	}
	return m
}

func TestLogPesanan(t *testing.T) {
	var buf bytes.Buffer
	l := NewLogger(&buf, zerolog.InfoLevel)

	LogPesanan(l, "ORD-001", 250_000, 42*time.Millisecond)

	m := baris(t, buf.String())
	if got := m["order_id"]; got != "ORD-001" {
		t.Errorf("order_id = %v, ingin ORD-001", got)
	}
	// Angka JSON selalu diurai jadi float64 di Go.
	if got := m["total"]; got != float64(250_000) {
		t.Errorf("total = %v, ingin 250000", got)
	}
	if got := m["level"]; got != "info" {
		t.Errorf("level = %v, ingin info", got)
	}
	if got := m["message"]; got != "pesanan dibuat" {
		t.Errorf("message = %v, ingin 'pesanan dibuat'", got)
	}
	// Dur() menulis durasi dalam milidetik (satuan baku zerolog).
	if got := m["durasi"]; got != float64(42) {
		t.Errorf("durasi = %v, ingin 42 (ms)", got)
	}
}

func TestLogGagalMenyimpanError(t *testing.T) {
	var buf bytes.Buffer
	l := NewLogger(&buf, zerolog.InfoLevel)

	LogGagal(l, "ORD-002", errors.New("stok habis"))

	m := baris(t, buf.String())
	if got := m["level"]; got != "error" {
		t.Errorf("level = %v, ingin error", got)
	}
	if got := m["error"]; got != "stok habis" {
		t.Errorf("error = %v, ingin 'stok habis'", got)
	}
}

// Sub-logger harus menempelkan kolom konteks ke SETIAP baris turunannya.
func TestSubLoggerMembawaKonteksKeSemuaBaris(t *testing.T) {
	var buf bytes.Buffer
	induk := NewLogger(&buf, zerolog.InfoLevel)
	req := SubLogger(induk, "checkout", "req-abc-123")

	req.Info().Msg("mulai")
	req.Info().Msg("selesai")

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("ingin 2 baris log, dapat %d", len(lines))
	}
	for i, ln := range lines {
		m := baris(t, ln)
		if m["layanan"] != "checkout" || m["request_id"] != "req-abc-123" {
			t.Errorf("baris %d kehilangan konteks: %v", i, m)
		}
	}

	// Logger induk TIDAK ikut tercemar konteks anaknya.
	buf.Reset()
	induk.Info().Msg("dari induk")
	if m := baris(t, buf.String()); m["request_id"] != nil {
		t.Error("logger induk seharusnya tak membawa request_id")
	}
}

func TestPenyaringanLevel(t *testing.T) {
	tests := []struct {
		nama       string
		level      zerolog.Level
		wantBaris  int
		wantLevels []string
	}{
		{"debug meloloskan semua", zerolog.DebugLevel, 4, []string{"debug", "info", "warn", "error"}},
		{"info menyaring debug", zerolog.InfoLevel, 3, []string{"info", "warn", "error"}},
		{"warn hanya warn ke atas", zerolog.WarnLevel, 2, []string{"warn", "error"}},
		{"error hanya error", zerolog.ErrorLevel, 1, []string{"error"}},
		{"disabled membungkam semua", zerolog.Disabled, 0, nil},
	}

	for _, tt := range tests {
		t.Run(tt.nama, func(t *testing.T) {
			var buf bytes.Buffer
			TulisSemuaLevel(NewLogger(&buf, tt.level))

			out := strings.TrimSpace(buf.String())
			if out == "" {
				if tt.wantBaris != 0 {
					t.Fatalf("tidak ada keluaran, ingin %d baris", tt.wantBaris)
				}
				return
			}
			lines := strings.Split(out, "\n")
			if len(lines) != tt.wantBaris {
				t.Fatalf("dapat %d baris, ingin %d", len(lines), tt.wantBaris)
			}
			for i, ln := range lines {
				if got := baris(t, ln)["level"]; got != tt.wantLevels[i] {
					t.Errorf("baris %d level = %v, ingin %s", i, got, tt.wantLevels[i])
				}
			}
		})
	}
}

func TestLevelDariString(t *testing.T) {
	tests := []struct {
		input string
		want  zerolog.Level
	}{
		{"debug", zerolog.DebugLevel},
		{"info", zerolog.InfoLevel},
		{"warn", zerolog.WarnLevel},
		{"error", zerolog.ErrorLevel},
		{"", zerolog.InfoLevel},          // kosong -> jatuh ke default aman
		{"ngawur", zerolog.InfoLevel},    // tak dikenal -> default aman
		{"TRACE", zerolog.TraceLevel},    // ParseLevel TIDAK peka huruf besar/kecil
		{"  info", zerolog.InfoLevel},    // spasi di depan tak dikenali -> default aman
		{"informasi", zerolog.InfoLevel}, // mirip tapi salah -> default aman
	}

	for _, tt := range tests {
		t.Run("input="+tt.input, func(t *testing.T) {
			if got := LevelDariString(tt.input); got != tt.want {
				t.Errorf("LevelDariString(%q) = %v, ingin %v", tt.input, got, tt.want)
			}
		})
	}
}

// Jebakan nomor satu: rantai tanpa .Msg() tidak menghasilkan apa pun,
// dan kompiler sama sekali tidak menegur.
func TestLupaMsgTidakMenulisApaPun(t *testing.T) {
	var buf bytes.Buffer
	l := NewLogger(&buf, zerolog.InfoLevel)

	LogTanpaMsg(l)
	if buf.Len() != 0 {
		t.Errorf("ingin keluaran kosong (log hilang), dapat: %q", buf.String())
	}

	LogDenganSend(l)
	if buf.Len() == 0 {
		t.Fatal(".Send() seharusnya mengirim log")
	}
	if got := baris(t, buf.String())["penting"]; got != "data ini terkirim" {
		t.Errorf("kolom penting = %v", got)
	}
}

func TestLogPesananSlogMenghasilkanJSONSetara(t *testing.T) {
	var buf bytes.Buffer
	LogPesananSlog(&buf, "ORD-004", 10_000)

	m := baris(t, buf.String())
	if m["order_id"] != "ORD-004" || m["total"] != float64(10_000) {
		t.Errorf("slog kehilangan kolom: %v", m)
	}
	// slog memakai nama kolom "msg", zerolog memakai "message" — beda konvensi,
	// isi informasinya sama.
	if m["msg"] != "pesanan dibuat" {
		t.Errorf("msg = %v", m["msg"])
	}
}
