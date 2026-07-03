// Modul 38 — Concurrency Advanced: errgroup, semaphore, singleflight, sync.Pool.
package main

import (
	"bytes"
	"context"
	"fmt"
	"sync"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
	"golang.org/x/sync/singleflight"
)

// ------------------------------------------------------------------
// 1. errgroup — jalankan banyak tugas paralel; error PERTAMA membatalkan sisanya.
// ------------------------------------------------------------------
// Lebih ringkas & aman dari WaitGroup+channel manual (Modul 7).
func FetchAll(ctx context.Context, ids []int, fetch func(ctx context.Context, id int) (string, error)) ([]string, error) {
	g, ctx := errgroup.WithContext(ctx)
	results := make([]string, len(ids)) // tiap goroutine tulis slot berbeda -> aman

	for i, id := range ids {
		i, id := i, id
		g.Go(func() error {
			r, err := fetch(ctx, id)
			if err != nil {
				return err // -> ctx dibatalkan -> goroutine lain berhenti
			}
			results[i] = r
			return nil
		})
	}
	if err := g.Wait(); err != nil { // tunggu semua; kembalikan error pertama
		return nil, err
	}
	return results, nil
}

// ------------------------------------------------------------------
//  2. semaphore berbobot — batasi jumlah operasi bersamaan (mirip bulkhead Modul 32,
//     tapi mendukung "bobot": satu operasi bisa mengambil >1 slot).
//
// ------------------------------------------------------------------
func RunLimited(ctx context.Context, tasks []func(), maxConcurrent int64) error {
	sem := semaphore.NewWeighted(maxConcurrent)
	var wg sync.WaitGroup
	for _, task := range tasks {
		if err := sem.Acquire(ctx, 1); err != nil { // blok bila slot penuh
			return err
		}
		wg.Add(1)
		go func(t func()) {
			defer wg.Done()
			defer sem.Release(1)
			t()
		}(task)
	}
	wg.Wait()
	return nil
}

// ------------------------------------------------------------------
//  3. singleflight — gabungkan panggilan IDENTIK yang bersamaan menjadi SATU.
//     Mengatasi "cache stampede": 1000 request miss serempak -> hanya 1 ke DB.
//
// ------------------------------------------------------------------
type Loader struct {
	group singleflight.Group
	calls int64 // berapa kali loader asli dipanggil
	mu    sync.Mutex
}

func (l *Loader) Get(key string, load func() (string, error)) (string, error) {
	// Do menjamin: untuk key sama, load hanya dijalankan SEKALI meski dipanggil
	// dari banyak goroutine bersamaan; semua menerima hasil yang sama.
	v, err, _ := l.group.Do(key, func() (any, error) {
		l.mu.Lock()
		l.calls++
		l.mu.Unlock()
		return load()
	})
	if err != nil {
		return "", err
	}
	return v.(string), nil
}

func (l *Loader) Calls() int64 {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.calls
}

// ------------------------------------------------------------------
// 4. sync.Pool — daur ulang objek untuk mengurangi alokasi (tekanan GC).
// ------------------------------------------------------------------
var bufPool = sync.Pool{
	New: func() any { return new(bytes.Buffer) },
}

// FormatGreeting memakai buffer dari pool alih-alih mengalokasi baru tiap panggil.
func FormatGreeting(names []string) string {
	buf := bufPool.Get().(*bytes.Buffer)
	buf.Reset()            // WAJIB reset — objek dari pool bisa bekas pakai
	defer bufPool.Put(buf) // kembalikan untuk dipakai ulang

	for i, n := range names {
		if i > 0 {
			buf.WriteString(", ")
		}
		fmt.Fprintf(buf, "Halo %s", n)
	}
	return buf.String()
}
