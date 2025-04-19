package proof

import "github.com/Abdullah-zahoor/shardedchain/trie"

// CompressedProof is a compact representation of a single‐shard Merkle proof.
type CompressedProof struct {
	Value     []byte     // leaf value
	SibKeys   [][]byte   // per‐step sorted sibling keys
	SibHashes [][][]byte // per‐step sibling hashes, aligned with SibKeys
}

// CompressProof turns a *trie.Proof into a *CompressedProof.
func CompressProof(p *trie.Proof) *CompressedProof {
	steps := len(p.Steps)
	cp := &CompressedProof{
		Value:     append([]byte(nil), p.Value...),
		SibKeys:   make([][]byte, steps),
		SibHashes: make([][][]byte, steps),
	}

	for i, sibMap := range p.Steps {
		// collect and sort keys
		keys := make([]byte, 0, len(sibMap))
		for k := range sibMap {
			keys = append(keys, k)
		}
		// simple sort
		for i := 0; i < len(keys); i++ {
			for j := i + 1; j < len(keys); j++ {
				if keys[j] < keys[i] {
					keys[i], keys[j] = keys[j], keys[i]
				}
			}
		}
		cp.SibKeys[i] = append([]byte(nil), keys...)
		// copy hashes in same order
		hashes := make([][]byte, len(keys))
		for j, k := range keys {
			hashes[j] = append([]byte(nil), sibMap[k]...)
		}
		cp.SibHashes[i] = hashes
	}

	return cp
}

// DecompressProof rebuilds a *trie.Proof from its compressed form.
func DecompressProof(cp *CompressedProof) *trie.Proof {
	steps := len(cp.SibKeys)
	p := &trie.Proof{
		Value: append([]byte(nil), cp.Value...),
		Steps: make([]map[byte][]byte, steps),
	}

	for i := 0; i < steps; i++ {
		m := make(map[byte][]byte, len(cp.SibKeys[i]))
		for j, k := range cp.SibKeys[i] {
			m[k] = append([]byte(nil), cp.SibHashes[i][j]...)
		}
		p.Steps[i] = m
	}

	return p
}
