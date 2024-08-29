package main

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"

	"github.com/bits-and-blooms/bitset"
)

const (
	chunkSize        = 1 << 20 // 1 MB chunks for reading the file
	bitmapSize       = 1 << 32 // Size of the bitmap (2^32 for IPv4)
	workerCount      = 4       // Number of worker goroutines
	hashSetThreshold = 1 << 24 // Use hash set up to ~16 million unique IPs
)

type Counter interface {
	Add(ip uint)
	Count() uint
}

type HashSetCounter struct {
	set map[uint]struct{}
}

func NewHashSetCounter() *HashSetCounter {
	return &HashSetCounter{set: make(map[uint]struct{})}
}

func (h *HashSetCounter) Add(ip uint) {
	h.set[ip] = struct{}{}
}

func (h *HashSetCounter) Count() uint {
	return uint(len(h.set))
}

type BitMapCounter struct {
	bitmap *bitset.BitSet
}

func NewBitMapCounter() *BitMapCounter {
	return &BitMapCounter{bitmap: bitset.New(bitmapSize)}
}

func (b *BitMapCounter) Add(ip uint) {
	b.bitmap.Set(uint(ip))
}

func (b *BitMapCounter) Count() uint {
	return b.bitmap.Count()
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Please provide the input file path")
		return
	}

	filePath := os.Args[1]
	uniqueCount, err := countUniqueIPs(filePath)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Number of unique IP addresses: %d\n", uniqueCount)
}

func countUniqueIPs(filePath string) (uint, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	var counter Counter = NewHashSetCounter()
	var mutex sync.Mutex
	jobs := make(chan []byte, workerCount)
	results := make(chan uint, workerCount)
	var wg sync.WaitGroup

	// Start worker goroutines
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go worker(jobs, results, &wg, &counter, &mutex)
	}

	// Read and process file in chunks
	go func() {
		reader := bufio.NewReaderSize(file, chunkSize)
		for {
			chunk := make([]byte, chunkSize)
			n, err := reader.Read(chunk)
			if err != nil {
				break
			}
			if n > 0 {
				jobs <- chunk[:n]
			}
		}
		close(jobs)
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	var totalCount uint
	for count := range results {
		totalCount += count
	}

	return counter.Count(), nil
}

func worker(jobs <-chan []byte, results chan<- uint, wg *sync.WaitGroup, counter *Counter, mutex *sync.Mutex) {
	defer wg.Done()
	localCounter := NewHashSetCounter()

	for chunk := range jobs {
		lines := strings.Split(string(chunk), "\n")
		for _, line := range lines {
			ip := strings.TrimSpace(line)
			ipInt, err := ipToInt(ip)
			if err != nil {
				continue // Skip invalid IP addresses
			}
			localCounter.Add(ipInt)
		}
	}

	mutex.Lock()
	if hashCounter, ok := (*counter).(*HashSetCounter); ok {
		if uint(len(hashCounter.set)+len(localCounter.set)) > hashSetThreshold {
			newCounter := NewBitMapCounter()
			for ip := range hashCounter.set {
				newCounter.Add(ip)
			}
			for ip := range localCounter.set {
				newCounter.Add(ip)
			}
			*counter = newCounter
		} else {
			for ip := range localCounter.set {
				hashCounter.Add(ip)
			}
		}
	} else if bitmapCounter, ok := (*counter).(*BitMapCounter); ok {
		for ip := range localCounter.set {
			bitmapCounter.Add(ip)
		}
	}
	mutex.Unlock()

	results <- localCounter.Count()
}

func ipToInt(ip string) (uint, error) {
	parts := strings.Split(ip, ".")
	if len(parts) != 4 {
		return 0, fmt.Errorf("invalid IP address: %s", ip)
	}

	var result uint
	for _, part := range parts {
		num := 0
		for i := 0; i < len(part); i++ {
			num = num*10 + int(part[i]-'0')
		}
		if num < 0 || num > 255 {
			return 0, fmt.Errorf("invalid IP address: %s", ip)
		}
		result = result<<8 | uint(num)
	}
	return result, nil
}

func init() {
	// Set the number of OS threads to use all available CPU cores
	runtime.GOMAXPROCS(runtime.NumCPU())
}
