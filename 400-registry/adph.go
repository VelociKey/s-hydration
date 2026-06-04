package registry

import (
	"errors"
	"fmt"
	"sort"
)

// PerfectHash represents a CHD (Compress, Hash, Displace) index structure.
type PerfectHash struct {
	Keys   []string
	Size   int
	G      []int // Displacement values per bucket
	B      []int // Map primary hash to bucket index
	BucketsCount int
}

// NewPerfectHash is the Factory Method to construct a Minimal Perfect Hash.
func NewPerfectHash(keys []string) (*PerfectHash, error) {
	n := len(keys)
	if n == 0 {
		return nil, errors.New("cannot create perfect hash for empty key set")
	}

	// We choose the number of buckets to be roughly n / 2 + 1 to keep displacements small
	bucketsCount := n/2 + 1
	buckets := make([][]string, bucketsCount)

	// Step 1: Place keys into buckets using the primary hash function
	for _, key := range keys {
		bIdx := int(hash1(key) % uint32(bucketsCount))
		buckets[bIdx] = append(buckets[bIdx], key)
	}

	// Step 2: Sort buckets by size in descending order (largest first)
	type bucketInfo struct {
		index int
		keys  []string
	}
	sortedBuckets := make([]bucketInfo, bucketsCount)
	for i := 0; i < bucketsCount; i++ {
		sortedBuckets[i] = bucketInfo{index: i, keys: buckets[i]}
	}
	sort.Slice(sortedBuckets, func(i, j int) bool {
		return len(sortedBuckets[i].keys) > len(sortedBuckets[j].keys)
	})

	// Step 3: Try to find displacements (g) to place keys into the flat visited array
	visited := make([]string, n)
	g := make([]int, bucketsCount)
	
	// Helper to track bucket positions
	bMap := make([]int, bucketsCount)

	for _, bucket := range sortedBuckets {
		if len(bucket.keys) == 0 {
			// Empty buckets get displacement 0
			g[bucket.index] = 0
			continue
		}

		disp := 1
		found := false
		
		// Attempt displacements until we find one without collisions
		for !found && disp < 1000000 {
			slots := make([]int, len(bucket.keys))
			collision := false
			
			for idx, key := range bucket.keys {
				slot := int(hash2(key, disp) % uint32(n))
				if visited[slot] != "" {
					collision = true
					break
				}
				// Verify we don't collision with other keys in the same bucket
				for prevIdx := 0; prevIdx < idx; prevIdx++ {
					if slots[prevIdx] == slot {
						collision = true
						break
					}
				}
				if collision {
					break
				}
				slots[idx] = slot
			}

			if !collision {
				// Mark slots as visited
				for idx, key := range bucket.keys {
					visited[slots[idx]] = key
				}
				g[bucket.index] = disp
				found = true
			} else {
				disp++
			}
		}

		if !found {
			return nil, fmt.Errorf("failed to find perfect hash displacement for bucket %d", bucket.index)
		}
	}

	// Build the mapping array mapping primary hash directly to bucket offset
	for i, bucket := range sortedBuckets {
		bMap[bucket.index] = i
	}

	return &PerfectHash{
		Keys:         visited, // Keys ordered by their slot indices
		Size:         n,
		G:            g,
		BucketsCount: bucketsCount,
	}, nil
}

// LookupIndex returns the slotted index of the key if present, and -1 otherwise.
func (ph *PerfectHash) LookupIndex(key string) int {
	if ph == nil || ph.Size == 0 {
		return -1
	}
	bIdx := int(hash1(key) % uint32(ph.BucketsCount))
	disp := ph.G[bIdx]
	slot := int(hash2(key, disp) % uint32(ph.Size))
	
	if slot >= 0 && slot < ph.Size && ph.Keys[slot] == key {
		return slot
	}
	return -1
}

func hash1(key string) uint32 {
	var hash uint32 = 2166136261
	for i := 0; i < len(key); i++ {
		hash ^= uint32(key[i])
		hash *= 16777619
	}
	return hash
}

func hash2(key string, displacement int) uint32 {
	var hash uint32 = uint32(displacement)
	for i := 0; i < len(key); i++ {
		hash ^= uint32(key[i])
		hash *= 16777619
	}
	return hash
}
