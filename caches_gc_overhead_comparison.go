package main

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"time"

	"github.com/allegro/bigcache"
)

var previousPause time.Duration

func gcPause() time.Duration {
	runtime.GC()
	var stats debug.GCStats
	debug.ReadGCStats(&stats)
	pause := stats.PauseTotal - previousPause
	previousPause = stats.PauseTotal
	return pause
}

const (
	entries   = 20000000
	valueSize = 100
	repeat    = 50
)

func main() {
	debug.SetGCPercent(10)
	fmt.Println("Number of entries: ", entries)
	fmt.Println("Number of repeats: ", repeat)

	fmt.Println("GC pause for startup: ", gcPause())

	stdMap()
	bigCache()

	fmt.Println("GC pause for warmup: ", gcPause())

	for i := 0; i < repeat; i++ {
		bigCache()
	}
	fmt.Println("GC pause for bigcache: ", gcPause())
	for i := 0; i < repeat; i++ {
		stdMap()
	}
	fmt.Println("GC pause for map: ", gcPause())
}

func stdMap() {
	mapCache := make(map[string][]byte)
	for i := 0; i < entries; i++ {
		key, val := generateKeyValue(i, valueSize)
		mapCache[key] = val
	}
}

func bigCache() {
	config := bigcache.Config{
		Shards:             256,
		LifeWindow:         100 * time.Minute,
		MaxEntriesInWindow: entries,
		MaxEntrySize:       200,
		Verbose:            true,
	}

	bigcache, _ := bigcache.NewBigCache(config)
	for i := 0; i < entries; i++ {
		key, val := generateKeyValue(i, valueSize)
		bigcache.Set(key, val)
	}

	firstKey, _ := generateKeyValue(1, valueSize)
	checkFirstElement(bigcache.Get(firstKey))
}

func checkFirstElement(val []byte, err error) {
	_, expectedVal := generateKeyValue(1, valueSize)
	if err != nil {
		fmt.Println("Error in get: ", err.Error())
	} else if string(val) != string(expectedVal) {
		fmt.Println("Wrong first element: ", string(val))
	}
}

func generateKeyValue(index int, valSize int) (string, []byte) {
	key := fmt.Sprintf("key-%010d", index)
	fixedNumber := []byte(fmt.Sprintf("%010d", index))
	val := append(make([]byte, valSize-10), fixedNumber...)

	return key, val
}
