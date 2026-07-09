# Makefile untuk repo belajar Go. Ketik `make` atau `make help` untuk daftar perintah.
.DEFAULT_GOAL := help
MOD ?= 01-basics

.PHONY: help run advanced realcase-up realcase-down test test-race cover fmt vet tidy proto clean tools

help: ## Tampilkan daftar perintah
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN{FS=":.*?## "}{printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2}'

run: ## Jalankan satu modul. Contoh: make run MOD=01-basics
	go run ./$(MOD)

advanced: ## Jalankan demo teknik advanced. Contoh: make advanced MOD=07-concurrency (modul 08 & 37: pakai `go test`)
	go run ./$(MOD)/advanced

realcase-up: ## Nyalakan infra docker-compose real-case. Contoh: make realcase-up MOD=22-caching
	docker compose -f $(MOD)/real-case/docker-compose.yml up -d

realcase-down: ## Matikan & bersihkan infra real-case. Contoh: make realcase-down MOD=22-caching
	docker compose -f $(MOD)/real-case/docker-compose.yml down -v

test: ## Jalankan semua test
	go test ./...

test-race: ## Jalankan test + race detector (untuk modul concurrency)
	go test -race ./...

cover: ## Jalankan test + laporan coverage
	go test -cover ./...

fmt: ## Format seluruh kode (gofmt)
	gofmt -w .

vet: ## Analisa statis (go vet)
	go vet ./...

tidy: ## Rapikan go.mod & go.sum
	go mod tidy

tools: ## Pasang plugin protoc untuk Go (perlu untuk `make proto`)
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

proto: ## Regenerate kode dari file .proto (butuh protoc + `make tools`)
	protoc --go_out=. --go_opt=module=go-learning --go-grpc_out=. --go-grpc_opt=module=go-learning 16-grpc/proto/calculator.proto
	protoc --go_out=. --go_opt=module=go-learning --go-grpc_out=. --go-grpc_opt=module=go-learning 17-studi-kasus-microservices/proto/inventory.proto

clean: ## Hapus artefak (file .db & folder bin)
	rm -f *.db 15-studi-kasus-rest/*.db
	rm -rf bin/
