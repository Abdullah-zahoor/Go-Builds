package proof

import (
	"bytes"
	"fmt"

	"github.com/Abdullah-zahoor/shardedchain/trie"
)

// CrossProof bundles two single‑shard proofs into one composite proof.
type CrossProof struct {
	// shard indexes
	SrcShard int
	DstShard int

	// keys involved
	SrcKey []byte
	DstKey []byte

	// pre‑transaction roots
	PreSrcRoot []byte
	PreDstRoot []byte

	// the Merkle proofs before applying the tx
	SrcProof *trie.Proof
	DstProof *trie.Proof

	// post‑transaction roots
	PostSrcRoot []byte
	PostDstRoot []byte
}

// GenerateCrossProof builds a CrossProof for moving `amount` from src→dst.
// You’ll need to:
// 1. Lookup both shard tries.
// 2. Get pre‑state proofs.
// 3. Apply the deduction/addition.
// 4. Get post‑state roots.
func GenerateCrossProof(
	srcShard int,
	dstShard int,
	keyFrom, keyTo []byte,
	amount []byte,
	getTrie func(shardIdx int) *trie.Node,
) (*CrossProof, error) {

	// 1. Pre‑state
	srcRoot := getTrie(srcShard).RootHash()
	dstRoot := getTrie(dstShard).RootHash()

	srcProof, err := getTrie(srcShard).GetProof(keyFrom)
	if err != nil {
		return nil, fmt.Errorf("src proof: %w", err)
	}
	dstProof, err := getTrie(dstShard).GetProof(keyTo)
	if err != nil {
		return nil, fmt.Errorf("dst proof: %w", err)
	}

	// 2. Apply state changes
	getTrie(srcShard).Insert(keyFrom, amount) // assume new value = old - amount
	getTrie(dstShard).Insert(keyTo, amount)   // assume new value = old + amount

	// 3. Post‑state roots
	newSrcRoot := getTrie(srcShard).RootHash()
	newDstRoot := getTrie(dstShard).RootHash()

	return &CrossProof{
		SrcShard:    srcShard,
		DstShard:    dstShard,
		SrcKey:      append([]byte(nil), keyFrom...),
		DstKey:      append([]byte(nil), keyTo...),
		PreSrcRoot:  append([]byte(nil), srcRoot...),
		PreDstRoot:  append([]byte(nil), dstRoot...),
		SrcProof:    srcProof,
		DstProof:    dstProof,
		PostSrcRoot: append([]byte(nil), newSrcRoot...),
		PostDstRoot: append([]byte(nil), newDstRoot...),
	}, nil
}

// VerifyCrossProof checks that both the pre‑state proofs are valid,
// and that the post roots differ from pre roots in the expected way.
// (Your actual verification may involve checking "new = old ± amount" on the client.)
func (cp *CrossProof) VerifyCrossProof(
	verifySingle func(root, key []byte, p *trie.Proof) bool,
) error {
	if !verifySingle(cp.PreSrcRoot, cp.SrcKey, cp.SrcProof) {
		return fmt.Errorf("invalid source pre‑proof")
	}
	if !verifySingle(cp.PreDstRoot, cp.DstKey, cp.DstProof) {
		return fmt.Errorf("invalid dest pre‑proof")
	}
	if bytes.Equal(cp.PreSrcRoot, cp.PostSrcRoot) {
		return fmt.Errorf("source root didn’t change")
	}
	if bytes.Equal(cp.PreDstRoot, cp.PostDstRoot) {
		return fmt.Errorf("dest root didn’t change")
	}
	return nil
}
