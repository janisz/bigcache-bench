package main

import (
	"fmt"
	"github.com/allegro/bigcache/v2"
	"github.com/cespare/xxhash"
	"github.com/coocood/freecache"
	"math/rand"
	"sync"
	"testing"
	"time"
)

const maxEntrySize = 256

var hashes = []bigcache.Hasher{
	nil,
	&xxHasher{},
}

func BenchmarkMapSet(b *testing.B) {
	m := make(map[string][]byte, b.N)
	for i := 0; i < b.N; i++ {
		m[key(i)] = value()
	}
}

func BenchmarkConcurrentMapSet(b *testing.B) {
	var m sync.Map
	for i := 0; i < b.N; i++ {
		m.Store(key(i), value())
	}
}

func BenchmarkFreeCacheSet(b *testing.B) {
	cache := freecache.NewCache(b.N * maxEntrySize)
	for i := 0; i < b.N; i++ {
		cache.Set([]byte(key(i)), value(), 0)
	}
}

func BenchmarkBigCacheSet(b *testing.B) {
	for _, hashFunc := range hashes {
		b.Run(fmt.Sprintf("%T", hashFunc), func(b *testing.B) {
			cache := initBigCache(b.N)
			for i := 0; i < b.N; i++ {
				cache.Set(key(i), value())
			}
		})
	}
}

func BenchmarkMapGet(b *testing.B) {
	b.StopTimer()
	m := make(map[string][]byte)
	for i := 0; i < b.N; i++ {
		m[key(i)] = value()
	}

	b.StartTimer()
	hitCount := 0
	for i := 0; i < b.N; i++ {
		if m[key(i)] != nil {
			hitCount++
		}
	}
}

func BenchmarkConcurrentMapGet(b *testing.B) {
	b.StopTimer()
	var m sync.Map
	for i := 0; i < b.N; i++ {
		m.Store(key(i), value())
	}

	b.StartTimer()
	hitCounter := 0
	for i := 0; i < b.N; i++ {
		_, ok := m.Load(key(i))
		if ok {
			hitCounter++
		}
	}
}

func BenchmarkFreeCacheGet(b *testing.B) {
	b.StopTimer()
	cache := freecache.NewCache(b.N * maxEntrySize)
	for i := 0; i < b.N; i++ {
		cache.Set([]byte(key(i)), value(), 0)
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		cache.Get([]byte(key(i)))
	}
}

func BenchmarkBigCacheGet(b *testing.B) {
	for _, hashFunc := range hashes {
		b.Run(fmt.Sprintf("%T", hashFunc), func(b *testing.B) {
			b.StopTimer()
			cache := initBigCache(b.N)
			for i := 0; i < b.N; i++ {
				cache.Set(key(i), value())
			}

			b.StartTimer()
			for i := 0; i < b.N; i++ {
				cache.Get(key(i))
			}
		})
	}
}

func BenchmarkBigCacheSetParallel(b *testing.B) {
	for _, hashFunc := range hashes {
		b.Run(fmt.Sprintf("%T", hashFunc), func(b *testing.B) {
			cache := initBigCache(b.N)
			rand.Seed(time.Now().Unix())

			b.RunParallel(func(pb *testing.PB) {
				id := rand.Intn(1000)
				counter := 0
				for pb.Next() {
					cache.Set(parallelKey(id, counter), value())
					counter = counter + 1
				}
			})
		})
	}
}

func BenchmarkFreeCacheSetParallel(b *testing.B) {
	cache := freecache.NewCache(b.N * maxEntrySize)
	rand.Seed(time.Now().Unix())

	b.RunParallel(func(pb *testing.PB) {
		id := rand.Intn(1000)
		counter := 0
		for pb.Next() {
			cache.Set([]byte(parallelKey(id, counter)), value(), 0)
			counter = counter + 1
		}
	})
}

func BenchmarkConcurrentMapSetParallel(b *testing.B) {
	var m sync.Map

	b.RunParallel(func(pb *testing.PB) {
		id := rand.Intn(1000)
		for pb.Next() {
			m.Store(key(id), value())
		}
	})
}

func BenchmarkBigCacheGetParallel(b *testing.B) {
	for _, hashFunc := range hashes {
		b.Run(fmt.Sprintf("%T", hashFunc), func(b *testing.B) {
			b.StopTimer()
			cache := initBigCache(b.N)
			for i := 0; i < b.N; i++ {
				cache.Set(key(i), value())
			}

			b.StartTimer()
			b.RunParallel(func(pb *testing.PB) {
				counter := 0
				for pb.Next() {
					cache.Get(key(counter))
					counter = counter + 1
				}
			})
		})
	}
}

func BenchmarkFreeCacheGetParallel(b *testing.B) {
	b.StopTimer()
	cache := freecache.NewCache(b.N * maxEntrySize)
	for i := 0; i < b.N; i++ {
		cache.Set([]byte(key(i)), value(), 0)
	}

	b.StartTimer()
	b.RunParallel(func(pb *testing.PB) {
		counter := 0
		for pb.Next() {
			cache.Get([]byte(key(counter)))
			counter = counter + 1
		}
	})
}

func BenchmarkConcurrentMapGetParallel(b *testing.B) {
	b.StopTimer()
	var m sync.Map
	for i := 0; i < b.N; i++ {
		m.Store(key(i), value())
	}

	b.StartTimer()
	hitCount := 0

	b.RunParallel(func(pb *testing.PB) {
		id := rand.Intn(1000)
		for pb.Next() {
			_, ok := m.Load(key(id))
			if ok {
				hitCount++
			}
		}
	})
}

func key(i int) string {
	return fmt.Sprintf("key-%010d", i)
}

func value() []byte {
	return make([]byte, 100)
}

func parallelKey(threadID int, counter int) string {
	return fmt.Sprintf("key-%04d-%06d", threadID, counter)
}

func initBigCache(entriesInWindow int) *bigcache.BigCache {
	cache, _ := bigcache.NewBigCache(bigcache.Config{
		Shards:             256,
		LifeWindow:         10 * time.Minute,
		MaxEntriesInWindow: entriesInWindow,
		MaxEntrySize:       maxEntrySize,
		Verbose:            true,
	})

	return cache
}

type xxHasher struct{}

func (x *xxHasher) Sum64(b string) uint64 {
	return xxhash.Sum64String(b)
}
