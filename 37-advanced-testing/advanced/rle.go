// Package advanced menyediakan codec Run-Length Encoding sederhana sebagai
// bahan demonstrasi TEKNIK pengujian lanjutan (fuzzing, property-based,
// integration-skip). Lihat rle_test.go.
//
// Jalankan test   : go test ./37-advanced-testing/advanced
// Jalankan fuzzing : go test -run x -fuzz FuzzRoundTrip -fuzztime 8s ./37-advanced-testing/advanced
package advanced

// Encode memampatkan byte berulang menjadi pasangan (jumlah, byte). Jumlah
// dibatasi 255 (satu byte) sehingga run panjang dipecah -> aman untuk data apa pun.
func Encode(in []byte) []byte {
	var out []byte
	for i := 0; i < len(in); {
		b := in[i]
		n := 1
		for i+n < len(in) && in[i+n] == b && n < 255 {
			n++
		}
		out = append(out, byte(n), b)
		i += n
	}
	return out
}

// Decode membalik Encode.
func Decode(in []byte) []byte {
	var out []byte
	for i := 0; i+1 < len(in); i += 2 {
		n := int(in[i])
		b := in[i+1]
		for j := 0; j < n; j++ {
			out = append(out, b)
		}
	}
	return out
}
