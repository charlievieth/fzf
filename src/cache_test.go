package fzf

import (
	"runtime"
	"strconv"
	"sync"
	"testing"
)

func TestChunkCache(t *testing.T) {
	cache := NewChunkCache()
	chunk1p := &Chunk{}
	chunk2p := &Chunk{count: chunkSize}
	items1 := []Result{{}}
	items2 := []Result{{}, {}}
	cache.Add(chunk1p, "foo", items1)
	cache.Add(chunk2p, "foo", items1)
	cache.Add(chunk2p, "bar", items2)

	{ // chunk1 is not full
		cached := cache.Lookup(chunk1p, "foo")
		if cached != nil {
			t.Error("Cached disabled for non-empty chunks", cached)
		}
	}
	{
		cached := cache.Lookup(chunk2p, "foo")
		if cached == nil || len(cached) != 1 {
			t.Error("Expected 1 item cached", cached)
		}
	}
	{
		cached := cache.Lookup(chunk2p, "bar")
		if cached == nil || len(cached) != 2 {
			t.Error("Expected 2 items cached", cached)
		}
	}
	{
		cached := cache.Lookup(chunk1p, "foobar")
		if cached != nil {
			t.Error("Expected 0 item cached", cached)
		}
	}
	{
		cached := cache.Search(chunk2p, "xbar")
		if cached == nil || len(cached) != 2 {
			t.Error("Expected 2 items cached", cached)
		}
	}
}

func TestChunkCacheParallel(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping: short test")
	}
	cache := NewChunkCache()
	start := make(chan struct{})
	done := make(chan struct{})

	var chunks [1024]*Chunk
	var keys [len(chunks)]string
	for i := 0; i < len(chunks); i++ {
		chunks[i] = &Chunk{count: chunkSize}
		keys[i] = "key_" + strconv.Itoa(i)
	}

	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; ; i++ {
				select {
				case <-done:
					if i >= 1000 {
						return
					}
				default:
					n := i % len(chunks)
					base := keys[n]
					chunk := chunks[n]
					for j := 0; j < 128; j++ {
						key := base + "_" + strconv.Itoa(j)
						if j&1 != 0 {
							cache.Lookup(chunk, key)
						} else {
							cache.Search(chunk, "x"+key)
						}
					}
				}
			}
		}()
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-start
		for i := 0; i < len(chunks); i++ {
			base := keys[i]
			for j := 0; j < 128; j++ {
				key := base + "_" + strconv.Itoa(j)
				cache.Add(chunks[i], key, []Result{{}})
			}

			runtime.Gosched()
		}
		close(done)
	}()
	close(start)
	wg.Wait()
}

func BenchmarkChunkCacheParallel(b *testing.B) {
	cache := NewChunkCache()
	var chunks [8]*Chunk
	var keys [len(chunks)]string
	for i := 0; i < len(chunks); i++ {
		chunks[i] = &Chunk{count: chunkSize}
		keys[i] = "key_" + strconv.Itoa(i)
		cache.Add(chunks[i], keys[i], []Result{{}})
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			n := i % len(chunks)
			cache.Lookup(chunks[n], keys[n])
			i++
		}
	})
}
