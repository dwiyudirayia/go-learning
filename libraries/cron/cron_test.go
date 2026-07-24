package main

import (
	"testing"
	"time"

	"github.com/robfig/cron/v3"
)

// acuan: Kamis, 23 Juli 2026, 10:30 UTC. Semua perhitungan jadwal diuji dari titik ini.
var acuan = time.Date(2026, 7, 23, 10, 30, 0, 0, time.UTC)

func TestValidasiSpec(t *testing.T) {
	tests := []struct {
		nama    string
		spec    string
		wantErr bool
	}{
		{"lima kotak standar", "0 3 * * *", false},
		{"langkah tiap 15 menit", "*/15 * * * *", false},
		{"rentang hari kerja", "0 9 * * 1-5", false},
		{"daftar bulan", "0 0 1 1,4,7,10 *", false},
		{"pintasan harian", "@daily", false},
		{"interval bebas", "@every 1h30m", false},
		{"kalimat manusia ditolak", "tiap hari pagi", true},
		{"menit di luar jangkauan", "99 * * * *", true},
		{"jam di luar jangkauan", "0 25 * * *", true},
		{"kotak kurang", "0 3 * *", true},
		{"kosong", "", true},
		// ParseStandard hanya mengenal 5 kotak. Spec 6 kotak (dengan detik) HARUS ditolak
		// di sini — itu format khusus yang butuh cron.WithSeconds() saat membuat penjadwal.
		{"enam kotak ditolak ParseStandard", "*/1 * * * * *", true},
	}

	for _, tt := range tests {
		t.Run(tt.nama, func(t *testing.T) {
			err := ValidasiSpec(tt.spec)
			if tt.wantErr && err == nil {
				t.Errorf("ValidasiSpec(%q) = nil, ingin error", tt.spec)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("ValidasiSpec(%q) error tak terduga: %v", tt.spec, err)
			}
		})
	}
}

func TestJadwalBerikutnya(t *testing.T) {
	tests := []struct {
		nama string
		spec string
		want []string // format "2006-01-02 15:04"
	}{
		{
			nama: "tiap hari jam 3 pagi",
			spec: "0 3 * * *",
			want: []string{"2026-07-24 03:00", "2026-07-25 03:00", "2026-07-26 03:00"},
		},
		{
			nama: "tiap 15 menit",
			spec: "*/15 * * * *",
			want: []string{"2026-07-23 10:45", "2026-07-23 11:00", "2026-07-23 11:15"},
		},
		{
			// Acuan hari Kamis. Setelah Jumat, Sabtu & Minggu DILEWATI — langsung Senin.
			nama: "jam 9 hari kerja saja",
			spec: "0 9 * * 1-5",
			want: []string{"2026-07-24 09:00", "2026-07-27 09:00", "2026-07-28 09:00"},
		},
		{
			nama: "pintasan harian jatuh tengah malam",
			spec: "@daily",
			want: []string{"2026-07-24 00:00", "2026-07-25 00:00", "2026-07-26 00:00"},
		},
		{
			nama: "interval bebas dihitung dari waktu acuan",
			spec: "@every 1h30m",
			want: []string{"2026-07-23 12:00", "2026-07-23 13:30", "2026-07-23 15:00"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.nama, func(t *testing.T) {
			got, err := JadwalBerikutnya(tt.spec, acuan, len(tt.want))
			if err != nil {
				t.Fatalf("JadwalBerikutnya(%q) error: %v", tt.spec, err)
			}
			for i, w := range tt.want {
				if g := got[i].Format("2006-01-02 15:04"); g != w {
					t.Errorf("jadwal ke-%d = %s, ingin %s", i+1, g, w)
				}
			}
		})
	}
}

func TestJadwalBerikutnyaSpecSalah(t *testing.T) {
	if _, err := JadwalBerikutnya("ngawur", acuan, 3); err == nil {
		t.Error("spec tidak valid seharusnya menghasilkan error, bukan jadwal")
	}
}

func TestJadwalBerikutnyaNolPermintaan(t *testing.T) {
	got, err := JadwalBerikutnya("@daily", acuan, 0)
	if err != nil {
		t.Fatalf("error tak terduga: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("meminta 0 jadwal, dapat %d", len(got))
	}
}

// Jadwal selalu MAJU: hasil berikutnya tak pernah sama atau lebih awal dari sebelumnya.
func TestJadwalSelaluMaju(t *testing.T) {
	got, err := JadwalBerikutnya("*/15 * * * *", acuan, 20)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	sebelumnya := acuan
	for i, w := range got {
		if !w.After(sebelumnya) {
			t.Fatalf("jadwal ke-%d (%s) tidak maju dari %s", i+1, w, sebelumnya)
		}
		sebelumnya = w
	}
}

// Zona waktu benar-benar mengubah kapan job berjalan — bukan sekadar tampilan.
func TestZonaWaktuMengubahJadwal(t *testing.T) {
	jakarta, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		t.Skip("basis data zona waktu tidak tersedia di sistem ini")
	}

	jadwal, err := cron.ParseStandard("0 3 * * *")
	if err != nil {
		t.Fatalf("parse gagal: %v", err)
	}

	// Waktu acuan yang sama, dibaca dari dua zona berbeda.
	berikutUTC := jadwal.Next(acuan)
	berikutWIB := jadwal.Next(acuan.In(jakarta))

	if berikutUTC.Equal(berikutWIB) {
		t.Error("jam 3 pagi UTC dan jam 3 pagi WIB seharusnya jatuh pada saat berbeda")
	}
	// WIB = UTC+7, jadi "jam 3 pagi WIB" berikutnya datang 7 jam lebih awal dalam UTC.
	if selisih := berikutUTC.Sub(berikutWIB); selisih != 7*time.Hour {
		t.Errorf("selisih = %v, ingin 7 jam", selisih)
	}
}

func TestPenjadwalAmanTidakRobohSaatJobPanic(t *testing.T) {
	if testing.Short() {
		t.Skip("melewati test yang menunggu waktu nyata")
	}

	c := NewPenjadwalAman()
	// Job ini SELALU panic. Tanpa cron.Recover, panic di goroutine job akan
	// merobohkan seluruh proses test.
	if _, err := c.AddFunc("@every 100ms", func() {
		panic("job sengaja rusak")
	}); err != nil {
		t.Fatalf("AddFunc gagal: %v", err)
	}

	c.Start()
	time.Sleep(350 * time.Millisecond)
	<-c.Stop().Done()

	// Sampai di baris ini artinya proses selamat — itulah yang diuji.
}

func TestPenghitungJobBerjalanBerulang(t *testing.T) {
	if testing.Short() {
		t.Skip("melewati test yang menunggu waktu nyata")
	}

	job := &PenghitungJob{}
	c := cron.New(cron.WithSeconds(), cron.WithLogger(cron.DiscardLogger))
	if _, err := c.AddFunc("*/1 * * * * *", job.Jalankan); err != nil {
		t.Fatalf("AddFunc gagal: %v", err)
	}

	c.Start()
	time.Sleep(2200 * time.Millisecond)
	<-c.Stop().Done()

	// Longgar dengan sengaja: penjadwalan waktu nyata tak pernah presisi sempurna,
	// dan test yang menuntut angka pasti akan jadi rapuh di mesin CI yang sibuk.
	if got := job.Jumlah(); got < 1 {
		t.Errorf("job berjalan %d kali dalam ~2 detik, ingin minimal 1", got)
	}
}

func TestBolehJalanHanyaSatuReplika(t *testing.T) {
	diambil := map[string]bool{}
	kunci := func(k string) bool {
		if diambil[k] {
			return false
		}
		diambil[k] = true
		return true
	}

	if !BolehJalan(kunci, "laporan-harian") {
		t.Error("replika pertama seharusnya menang merebut kunci")
	}
	for i := 2; i <= 5; i++ {
		if BolehJalan(kunci, "laporan-harian") {
			t.Errorf("replika %d seharusnya kalah, tapi ikut menjalankan job", i)
		}
	}
	// Job dengan nama berbeda punya kunci sendiri — tidak saling menghalangi.
	if !BolehJalan(kunci, "bersih-bersih") {
		t.Error("job berbeda seharusnya punya kunci terpisah")
	}
}
