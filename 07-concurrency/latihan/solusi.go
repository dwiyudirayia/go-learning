// Solusi latihan Modul 07 — Concurrency.
// Jalankan:        go run ./07-concurrency/latihan
// Dengan race det: go run -race ./07-concurrency/latihan
package main

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"
)

func main() {
	fmt.Println("=== Solusi Latihan Modul 07 ===")
	latihan1()
	latihan2()
	latihan3()
	latihan4()
	latihan5()
}

// ------------------------------------------------------------------
// Latihan 1: 5 goroutine cetak ID + WaitGroup
// ------------------------------------------------------------------
func latihan1() {
	fmt.Println("\n-- Latihan 1: 5 goroutine + WaitGroup --")

	var wg sync.WaitGroup
	var mu sync.Mutex
	var ids []int

	for i := 1; i <= 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			// Kumpulkan ID dengan aman (mutex) supaya bebas data race.
			mu.Lock()
			ids = append(ids, id)
			mu.Unlock()
		}(i)
	}
	wg.Wait()

	sort.Ints(ids) // urutan eksekusi goroutine acak; urutkan agar output stabil
	fmt.Printf("ID goroutine yang selesai: %v\n", ids)
}

// ------------------------------------------------------------------
// Latihan 2: generator <-chan int 1..n
// ------------------------------------------------------------------
func generator(n int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out) // pengirim menutup channel
		for i := 1; i <= n; i++ {
			out <- i
		}
	}()
	return out
}

func latihan2() {
	fmt.Println("\n-- Latihan 2: generator + range --")
	fmt.Print("nilai: ")
	for v := range generator(6) {
		fmt.Printf("%d ", v)
	}
	fmt.Println()
}

// ------------------------------------------------------------------
// Latihan 3: Counter aman-konkuren dengan Mutex
// ------------------------------------------------------------------
type SafeCounter struct {
	mu    sync.Mutex
	value int
}

func (c *SafeCounter) Inc() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.value++
}

func (c *SafeCounter) Value() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.value
}

func latihan3() {
	fmt.Println("\n-- Latihan 3: Counter + Mutex (uji -race) --")

	c := &SafeCounter{}
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); c.Inc() }()
	}
	wg.Wait()
	fmt.Printf("nilai akhir = %d (harus tepat 100)\n", c.Value())
}

// ------------------------------------------------------------------
// Latihan 4: worker pool 3 worker, 9 job (kuadrat)
// ------------------------------------------------------------------
func latihan4() {
	fmt.Println("\n-- Latihan 4: worker pool --")

	const workers, jobsN = 3, 9
	jobs := make(chan int, jobsN)
	results := make(chan int, jobsN)

	var wg sync.WaitGroup
	for w := 1; w <= workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				results <- j * j
			}
		}()
	}

	for j := 1; j <= jobsN; j++ {
		jobs <- j
	}
	close(jobs)

	go func() { wg.Wait(); close(results) }()

	var out []int
	for r := range results {
		out = append(out, r)
	}
	sort.Ints(out)
	fmt.Printf("hasil kuadrat = %v\n", out)
}

// ------------------------------------------------------------------
// Latihan 5: fetchWithTimeout memakai context + select
// ------------------------------------------------------------------
var ErrTimeout = errors.New("operasi melebihi batas waktu")

// fetchWithTimeout mensimulasikan fetch yang butuh 'kerja' waktu, dibatasi 'batas'.
func fetchWithTimeout(kerja, batas time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), batas)
	defer cancel()

	hasil := make(chan string, 1)
	go func() {
		time.Sleep(kerja) // simulasi I/O
		hasil <- "data berhasil diambil"
	}()

	select {
	case res := <-hasil:
		return res, nil
	case <-ctx.Done():
		return "", fmt.Errorf("%w: %v", ErrTimeout, ctx.Err())
	}
}

func latihan5() {
	fmt.Println("\n-- Latihan 5: fetchWithTimeout --")

	// Cepat: kerja 20ms, batas 100ms -> sukses.
	if res, err := fetchWithTimeout(20*time.Millisecond, 100*time.Millisecond); err == nil {
		fmt.Println("cepat  ->", res)
	}
	// Lambat: kerja 200ms, batas 50ms -> timeout.
	if _, err := fetchWithTimeout(200*time.Millisecond, 50*time.Millisecond); err != nil {
		fmt.Println("lambat ->", err)
		fmt.Printf("  errors.Is(err, ErrTimeout) = %t\n", errors.Is(err, ErrTimeout))
	}
}
