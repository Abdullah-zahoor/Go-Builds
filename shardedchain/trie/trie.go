package trie

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"sort"
)

// Node is one node in our Merkle trie.
type Node struct {
	children map[byte]*Node
	value    []byte
	hash     []byte
}

// KV is a simple key/value pair for traversal.
type KV struct {
	Key   []byte
	Value []byte
}

// NewNode creates an empty trie node.
func NewNode() *Node {
	return &Node{children: make(map[byte]*Node)}
}

// Insert writes value at the given key path and then updates hashes upward.
func (n *Node) Insert(key []byte, value []byte) {
	if len(key) == 0 {
		n.value = value
	} else {
		b := key[0]
		child, ok := n.children[b]
		if !ok {
			child = NewNode()
			n.children[b] = child
		}
		child.Insert(key[1:], value)
	}
	n.computeHash()
}

// computeHash recomputes this node’s hash from its value and sorted children.
func (n *Node) computeHash() {
	h := sha256.New()
	// value prefix = 0
	if n.value != nil {
		h.Write([]byte{0})
		h.Write(n.value)
	}
	// children prefix = 1, sorted by key byte
	var keys []int
	for b := range n.children {
		keys = append(keys, int(b))
	}
	sort.Ints(keys)
	for _, ki := range keys {
		b := byte(ki)
		child := n.children[b]
		h.Write([]byte{1})
		h.Write([]byte{b})
		h.Write(child.hash)
	}
	n.hash = h.Sum(nil)
}

// RootHash returns the current hash of this node.
func (n *Node) RootHash() []byte {
	return n.hash
}

// GetProof builds a Merkle proof for key (error if not present).
func (n *Node) GetProof(key []byte) (*Proof, error) {
	steps := []map[byte][]byte{}
	node := n
	for _, b := range key {
		if node == nil {
			return nil, errors.New("key not found")
		}
		sibMap := make(map[byte][]byte)
		for sb, sib := range node.children {
			if sb != b {
				sibMap[sb] = sib.hash
			}
		}
		steps = append(steps, sibMap)
		node = node.children[b]
	}
	if node == nil || node.value == nil {
		return nil, errors.New("key not found")
	}
	return &Proof{Value: node.value, Steps: steps}, nil
}

// Proof holds a leaf value plus, for each depth, a map of sibling‑hashes.
type Proof struct {
	Value []byte
	Steps []map[byte][]byte
}

// VerifyProof checks that proof under key yields the given rootHash.
func VerifyProof(rootHash []byte, key []byte, proof *Proof) bool {
	// compute leaf hash
	h := sha256.New()
	h.Write([]byte{0})
	h.Write(proof.Value)
	current := h.Sum(nil)

	// climb back up
	for i := len(key) - 1; i >= 0; i-- {
		sibMap := proof.Steps[i]
		// collect all child‐keys at this level
		set := make(map[int]bool)
		for sb := range sibMap {
			set[int(sb)] = true
		}
		set[int(key[i])] = true
		var ks []int
		for k := range set {
			ks = append(ks, k)
		}
		sort.Ints(ks)

		h2 := sha256.New()
		for _, ki := range ks {
			b := byte(ki)
			h2.Write([]byte{1})
			h2.Write([]byte{b})
			if b == key[i] {
				h2.Write(current)
			} else {
				h2.Write(sibMap[b])
			}
		}
		current = h2.Sum(nil)
	}

	return bytes.Equal(current, rootHash)
}

// Traverse returns all key/value pairs in the trie.
func (n *Node) Traverse() []KV {
	var result []KV
	var walk func(node *Node, prefix []byte)
	walk = func(node *Node, prefix []byte) {
		if node.value != nil {
			// copy slices so mutations later won’t break them
			kCopy := append([]byte(nil), prefix...)
			vCopy := append([]byte(nil), node.value...)
			result = append(result, KV{Key: kCopy, Value: vCopy})
		}
		for b, child := range node.children {
			walk(child, append(prefix, b))
		}
	}
	walk(n, []byte{})
	return result
}
