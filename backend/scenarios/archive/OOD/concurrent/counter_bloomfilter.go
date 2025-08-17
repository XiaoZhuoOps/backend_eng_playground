package concurrent

import "fmt"

// CountingBloomFilter represents a counting Bloom filter that supports deletion.
type CountingBloomFilter struct {
	counts     []uint16 // Array of counters instead of bits
	size       uint     // Number of bits (same conceptual "size", but each slot has a counter)
	hashFuncs  uint     // Number of hash functions
	elementCnt uint     // Number of elements currently added
}

// NewCountingBloomFilter creates a new Counting Bloom Filter
// with the specified size and number of hash functions.
func NewCountingBloomFilter(size uint, hashFuncs uint) *CountingBloomFilter {
	return &CountingBloomFilter{
		counts:    make([]uint16, size), // Each index starts at 0
		size:      size,
		hashFuncs: hashFuncs,
	}
}

// Add increments the counters at each hash location for the given item.
func (cbf *CountingBloomFilter) Add(item string) {
	locations := cbf.getHashLocations(item)
	for _, loc := range locations {
		cbf.counts[loc]++ // Increment counter
	}
	cbf.elementCnt++
}

// Remove decrements the counters at each hash location for the given item.
// If a counter is already 0, it will not be decremented further (prevent underflow).
func (cbf *CountingBloomFilter) Remove(item string) {
	locations := cbf.getHashLocations(item)
	for _, loc := range locations {
		if cbf.counts[loc] > 0 {
			cbf.counts[loc]--
		}
	}
	if cbf.elementCnt > 0 {
		cbf.elementCnt--
	}
}

// Contains checks whether an element is possibly in the set.
// Returns false if any of the corresponding counters are zero.
func (cbf *CountingBloomFilter) Contains(item string) bool {
	locations := cbf.getHashLocations(item)
	for _, loc := range locations {
		if cbf.counts[loc] == 0 {
			return false
		}
	}
	return true
}

// getHashLocations generates the positions in the filter
// where we increment/decrement counters for the given item.
func (cbf *CountingBloomFilter) getHashLocations(item string) []uint {
	locations := make([]uint, cbf.hashFuncs)
	hash1, hash2 := hash(item)

	for i := uint(0); i < cbf.hashFuncs; i++ {
		combinedHash := hash1 + i*hash2
		locations[i] = combinedHash % cbf.size
	}
	return locations
}

func main() {
	// Example usage of Counting Bloom Filter
	n := uint(1000) // Estimated number of elements
	p := 0.01       // Desired false positive rate

	// Calculate optimal size and number of hash functions (optional usage)
	m, k := OptimalParameters(n, p)
	fmt.Printf("Counting Bloom Filter Parameters: size=%d bits, hash functions=%d\n", m, k)

	// Initialize Counting Bloom Filter
	cbf := NewCountingBloomFilter(m, k)

	// Add elements
	elements := []string{"apple", "banana", "orange"}
	for _, item := range elements {
		cbf.Add(item)
		fmt.Printf("Added: %s\n", item)
	}

	// Check membership before removal
	for _, item := range elements {
		if cbf.Contains(item) {
			fmt.Printf("[Before Removal] Possibly in set: %s\n", item)
		} else {
			fmt.Printf("[Before Removal] Definitely not in set: %s\n", item)
		}
	}

	// Remove an element
	cbf.Remove("banana")
	fmt.Println("Removed: banana")

	// Check membership after removal
	testItems := []string{"apple", "banana", "grape"}
	for _, item := range testItems {
		if cbf.Contains(item) {
			fmt.Printf("[After Removal] Possibly in set: %s\n", item)
		} else {
			fmt.Printf("[After Removal] Definitely not in set: %s\n", item)
		}
	}
}
