package global

import (
	"crypto/sha256"
)

// BuildGlobalRoot takes each shard’s root hash and combines them
// into a single global root by hashing them in sorted order.
func BuildGlobalRoot(shardRoots [][]byte) []byte {
	// simple approach: concatenate all roots and hash once
	h := sha256.New()
	for _, r := range shardRoots {
		// prefix each with length to avoid ambiguity
		h.Write([]byte{byte(len(r) >> 8), byte(len(r) & 0xff)})
		h.Write(r)
	}
	return h.Sum(nil)
}
