package concurrent

import (
	"fmt"
	"hash/fnv"
	"math"
)

// BloomFilter represents a Bloom filter.
type BloomFilter struct {
	bitset     []byte
	size       uint
	hashFuncs  uint
	elementCnt uint
}

// NewBloomFilter creates a new Bloom filter with given size and number of hash functions.
func NewBloomFilter(size uint, hashFuncs uint) *BloomFilter {
	byteSize := (size + 7) / 8 // Calculate number of bytes needed
	return &BloomFilter{
		bitset:    make([]byte, byteSize),
		size:      size,
		hashFuncs: hashFuncs,
	}
}

// Add inserts an element into the Bloom filter.
func (bf *BloomFilter) Add(item string) {
	locations := bf.getHashLocations(item)
	for _, loc := range locations {
		byteIndex := loc / 8
		bitIndex := loc % 8
		bf.bitset[byteIndex] |= (1 << bitIndex)
	}
	bf.elementCnt++
}

// Contains checks whether an element is possibly in the Bloom filter.
func (bf *BloomFilter) Contains(item string) bool {
	locations := bf.getHashLocations(item)
	for _, loc := range locations {
		byteIndex := loc / 8
		bitIndex := loc % 8
		if bf.bitset[byteIndex]&(1<<bitIndex) == 0 {
			return false
		}
	}
	return true
}

// getHashLocations generates hash locations for an item.
func (bf *BloomFilter) getHashLocations(item string) []uint {
	locations := make([]uint, bf.hashFuncs)
	hash1, hash2 := hash(item)

	for i := uint(0); i < bf.hashFuncs; i++ {
		combinedHash := hash1 + i*hash2
		locations[i] = combinedHash % bf.size
	}
	return locations
}

// hash generates two hash values using FNV-1a.
func hash(s string) (uint, uint) {
	h1 := fnv.New64a()
	h1.Write([]byte(s))
	sum1 := h1.Sum64()

	h2 := fnv.New64()
	h2.Write([]byte(s))
	sum2 := h2.Sum64()

	return uint(sum1), uint(sum2)
}

// OptimalParameters calculates optimal size and number of hash functions for desired false positive rate.
func OptimalParameters(n uint, p float64) (m uint, k uint) {
	m = uint(math.Ceil(float64(n) * math.Log(p) / math.Log(1.0/math.Pow(2.0, math.Log(2.0)))))
	k = uint(math.Round((float64(m) / float64(n)) * math.Log(2.0)))
	return
}

func main() {
	// Example usage
	// Number of elements expected to be stored
	n := uint(1000)
	// Desired false positive rate
	p := 0.01

	// Calculate optimal size and number of hash functions
	m, k := OptimalParameters(n, p)
	fmt.Printf("Bloom Filter Parameters: size=%d bits, hash functions=%d\n", m, k)

	// Initialize Bloom filter
	bf := NewBloomFilter(m, k)

	// Add elements
	elements := []string{"apple", "banana", "orange", "grape", "watermelon"}
	for _, item := range elements {
		bf.Add(item)
		fmt.Printf("Added: %s\n", item)
	}

	// Check for membership
	testElements := []string{"apple", "banana", "cherry", "date", "fig", "grape"}

	for _, item := range testElements {
		if bf.Contains(item) {
			fmt.Printf("Possibly in set: %s\n", item)
		} else {
			fmt.Printf("Definitely not in set: %s\n", item)
		}
	}
}

