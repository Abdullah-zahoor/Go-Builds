package dag

import (
	"fmt"

	"github.com/Abdullah-zahoor/dagchain/block"
)

// NewDAG initializes an empty DAG.
func NewDAG() *DAG {
	return &DAG{
		Nodes: make(map[string]*Node),
	}
}

// AddGenesis seeds the DAG with a genesis block and its starting UTXO set.
func (d *DAG) AddGenesis(genesis *block.Block, initialUTXO block.UTXOSet) error {
	if len(genesis.Parents) != 0 {
		return fmt.Errorf("genesis block must have no parents")
	}
	node := &Node{
		Block:    genesis,
		Parents:  nil,
		Children: nil,
		Weight:   uint64(len(genesis.TXs)),
		UTXO:     initialUTXO.Clone(),
	}
	d.Nodes[genesis.ID] = node
	return nil
}

// AddBlock inserts blk into the DAG, links it, computes its UTXO snapshot & weight.
func (d *DAG) AddBlock(blk *block.Block) error {
	// 1. Gather parents
	parents := make([]*Node, 0, len(blk.Parents))
	for _, pid := range blk.Parents {
		p, ok := d.Nodes[pid]
		if !ok {
			return fmt.Errorf("parent %s not found", pid)
		}
		parents = append(parents, p)
	}

	// 2. INTERSECTION‑merge parent UTXOs
	var merged block.UTXOSet
	if len(parents) == 0 {
		merged = make(block.UTXOSet)
	} else {
		// start from first parent's snapshot
		merged = parents[0].UTXO.Clone()
		// drop any UTXO not present in *every* parent
		for _, p := range parents[1:] {
			for key := range merged {
				if _, ok := p.UTXO[key]; !ok {
					delete(merged, key)
				}
			}
		}
	}

	// 3. Validate & apply TXs inline
	for _, tx := range blk.TXs {
		if err := merged.ApplyTx(tx); err != nil {
			return fmt.Errorf("block %s has invalid tx %s: %w",
				blk.ID, tx.ID, err)
		}
	}

	// 4. Compute weight = max(parent.Weight) + len(TXs)
	var maxW uint64
	for _, p := range parents {
		if p.Weight > maxW {
			maxW = p.Weight
		}
	}
	weight := maxW + uint64(len(blk.TXs))

	// 5. Create the new node and link it
	newNode := &Node{
		Block:    blk,
		Parents:  parents,
		Children: nil,
		Weight:   weight,
		UTXO:     merged,
	}
	for _, p := range parents {
		p.Children = append(p.Children, newNode)
	}
	d.Nodes[blk.ID] = newNode

	return nil
}
