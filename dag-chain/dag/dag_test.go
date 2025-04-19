package dag_test

import (
	"testing"
	"time"

	"github.com/Abdullah-zahoor/dagchain/block"
	"github.com/Abdullah-zahoor/dagchain/dag"
)

func TestAddGenesisAndBlock(t *testing.T) {
	d := dag.NewDAG()

	// Add genesis
	gen := &block.Block{ID: "g", Parents: nil, TXs: nil, Timestamp: time.Now()}
	if err := d.AddGenesis(gen, make(block.UTXOSet)); err != nil {
		t.Fatalf("AddGenesis failed: %v", err)
	}

	// Add a child block with one TX
	tx := block.TX{
		ID:      "tx",
		Inputs:  nil,
		Outputs: []block.TXOutput{{Value: 1, Recipient: "X"}},
	}
	b := &block.Block{ID: "b", Parents: []string{"g"}, TXs: []block.TX{tx}, Timestamp: time.Now()}
	if err := d.AddBlock(b); err != nil {
		t.Fatalf("AddBlock failed: %v", err)
	}

	// Check parent‑child links
	parent := d.Nodes["g"]
	child := d.Nodes["b"]
	if len(parent.Children) != 1 || parent.Children[0] != child {
		t.Error("child link missing")
	}
	if len(child.Parents) != 1 || child.Parents[0] != parent {
		t.Error("parent link missing")
	}

	// Weight should equal number of TXs = 1
	if child.Weight != 1 {
		t.Errorf("expected weight=1, got %d", child.Weight)
	}
}
