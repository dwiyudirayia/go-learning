// Teknik Advanced Modul 03 (structs & methods) — contoh runnable + penjelasan detail.
//
// Jalankan: go run ./03-structs-methods/advanced
package main

import (
	"fmt"
	"unsafe"
)

// ===========================================================================
// 1. Method value vs method expression
// ===========================================================================
// - Method VALUE : t.Naik  -> sudah terikat ke receiver t (closure).
// - Method EXPR   : Counter.Naik -> fungsi biasa, receiver jadi argumen pertama.
// Keduanya berguna saat mengoper method sebagai nilai (higher-order function).
//
// 🔍 Analogi: method value itu seperti NOMOR SPEEDDIAL — tombolnya sudah
// terikat ke satu orang (receiver), tinggal pencet. Method expression seperti
// TELEPON UMUM: alatnya sama, tapi tiap kali harus sebut dulu mau menelepon
// siapa (receiver dioper sebagai argumen).

type Counter struct{ n int }

func (c *Counter) Naik() { c.n++ }

// ===========================================================================
// 2. Method set & pemenuhan interface (pointer vs value receiver)
// ===========================================================================
// Method dengan POINTER receiver hanya masuk method set *T, BUKAN T.
// Akibatnya: nilai T (bukan &T) mungkin TIDAK memenuhi interface. Ini sumber
// error paling umum bagi pemula Go.
//
// 🔍 Analogi: pointer receiver itu seperti KUNCI RUMAH ASLI — hanya pemegang
// alamat rumah (*T) yang bisa mengubah isinya. Nilai T cuma FOTO rumahnya:
// mau diberi kunci pun percuma, maka Go tak mengakuinya sebagai "pemilik"
// (tidak memenuhi interface).

type Penyapa interface{ Sapa() string }

type Orang struct{ Nama string }

func (o *Orang) Sapa() string { return "Halo, " + o.Nama } // pointer receiver

// ===========================================================================
// 3. Alignment & padding — urutan field memengaruhi ukuran struct
// ===========================================================================
// Field disusun rata (aligned) ke kelipatan ukurannya. Menaruh field kecil di
// antara field besar menyisipkan "padding" kosong. Menyusun dari besar ke
// kecil sering memangkas ukuran total. (Detail lengkap: modul 42.)
//
// 🔍 Analogi: menyusun field itu seperti MENGEPAK KOPER — barang besar ditaruh
// dulu, barang kecil menyelip di celahnya. Kalau barang kecil ditaruh acak di
// antara yang besar, muncul rongga kosong (padding) dan koper "membengkak".

type Boros struct {
	a bool  // 1 byte + 7 padding
	b int64 // 8 byte
	c bool  // 1 byte + 7 padding
} // total 24 byte

type Hemat struct {
	b int64 // 8 byte
	a bool  // 1 byte
	c bool  // 1 byte + 6 padding
} // total 16 byte

// ===========================================================================
// 4. Struct comparability — syarat jadi map key
// ===========================================================================
// Struct bisa dibandingkan dengan == (dan dipakai sebagai kunci map) HANYA
// jika semua field-nya comparable. Struct berisi slice/map/func tidak bisa.
//
// 🔍 Analogi: map key harus seperti KTP — datanya tetap dan bisa dicocokkan
// persis. Struct berisi slice/map itu seperti "alamat yang isinya bisa
// berubah-ubah": tak bisa dijadikan identitas pembanding.

type Titik struct{ X, Y int } // comparable

func main() {
	fmt.Println("== 1. method value vs expression ==")
	c := &Counter{}
	naik := c.Naik // method value: terikat ke c
	naik()
	naik()
	fmt.Println("  method value, n =", c.n) // 2

	naikExpr := (*Counter).Naik // method expression: butuh receiver eksplisit
	naikExpr(c)
	fmt.Println("  method expr , n =", c.n) // 3

	fmt.Println("== 2. method set & interface ==")
	var p Penyapa = &Orang{"Budi"} // OK: *Orang memenuhi Penyapa
	fmt.Println(" ", p.Sapa())
	// var q Penyapa = Orang{"Budi"} // ERROR compile: Orang (value) TIDAK memenuhi Penyapa
	fmt.Println("  catatan: nilai Orang{} (bukan &Orang{}) tak memenuhi interface")

	fmt.Println("== 3. alignment & padding ==")
	fmt.Printf("  Boros = %d byte, Hemat = %d byte (isi field sama!)\n",
		unsafe.Sizeof(Boros{}), unsafe.Sizeof(Hemat{}))

	fmt.Println("== 4. struct sebagai map key ==")
	grid := map[Titik]string{
		{0, 0}: "asal",
		{1, 2}: "target",
	}
	fmt.Println("  grid[{1,2}] =", grid[Titik{1, 2}])
}
