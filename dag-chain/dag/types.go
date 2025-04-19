package dag

import "github.com/Abdullah-zahoor/dagchain/block"

// Node wraps a block and links to parents/children.
type Node struct {
	Block    *block.Block
	Parents  []*Node
	Children []*Node
	Weight   uint64 // cumulative work or tx count
	UTXO     block.UTXOSet
}

// DAG holds all nodes by their Block.ID.
type DAG struct {
	Nodes map[string]*Node
}
