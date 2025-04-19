package shard

import "github.com/Abdullah-zahoor/shardedchain/trie"

// Shard holds one Merkle trie + a mutation counter.
type Shard struct {
	Tree      *trie.Node
	Mutations int
}

// NewShard creates an empty shard.
func NewShard() *Shard {
	return &Shard{Tree: trie.NewNode()}
}

// Apply writes value at key and bumps the mutation count.
func (s *Shard) Apply(key, value []byte) {
	s.Tree.Insert(key, value)
	s.Mutations++
}

// Root returns this shard’s current Merkle root.
func (s *Shard) Root() []byte {
	return s.Tree.RootHash()
}
