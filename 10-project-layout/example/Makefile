# Makefile contoh untuk modul 10.
# Catatan: path relatif di sini ditujukan bila kamu berada di folder example/.
# Dari root repo, jalankan langsung: go run ./10-project-layout/example/cmd/greeter

.PHONY: run test build lint tidy clean

run:              ## jalankan aplikasi
	go run ./cmd/greeter -name Ana

test:             ## jalankan semua test
	go test ./...

build:            ## kompilasi binary ke bin/
	go build -o bin/greeter ./cmd/greeter

lint:             ## analisa statis
	go vet ./...

tidy:             ## rapikan dependency
	go mod tidy

clean:            ## bersihkan artefak build
	rm -rf bin/
