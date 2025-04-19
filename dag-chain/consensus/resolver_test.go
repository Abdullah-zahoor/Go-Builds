package consensus_test

import (
	"testing"

	"github.com/Abdullah-zahoor/dagchain/block"
	"github.com/Abdullah-zahoor/dagchain/consensus"
	"github.com/Abdullah-zahoor/dagchain/dag"
)

func makeSimpleDAG() *dag.DAG {
	d := dag.NewDAG()
	// genesis
	g := &block.Block{ID: "g", Parents: nil}
	d.AddGenesis(g, make(block.UTXOSet))
	// fork1 → g
	f1 := &block.Block{ID: "f1", Parents: []string{"g"}}
	d.AddBlock(f1)
	// fork2 → g
	f2 := &block.Block{ID: "f2", Parents: []string{"g"}}
	d.AddBlock(f2)
	return d
}

func TestHeaviestTip(t *testing.T) {
	d := makeSimpleDAG()
	// both forks have weight 0 (no TXs) ⇒ first returned
	tip := consensus.HeaviestTip(d)
	if tip.Block.ID != "f1" {
		t.Errorf("expected f1 as heaviest tip, got %s", tip.Block.ID)
	}
}

func TestPruneBranches(t *testing.T) {
	d := makeSimpleDAG()
	// artificially add a TX to fork2 to make it heavier
	d.Nodes["f2"].Weight = 1
	consensus.PruneBranches(d)
	// only f2 and g should remain
	if _, ok := d.Nodes["f1"]; ok {
		t.Error("f1 should have been pruned")
	}
	if _, ok := d.Nodes["f2"]; !ok {
		t.Error("f2 should have been kept")
	}
}

func TestFinalized(t *testing.T) {
	d := makeSimpleDAG()
	final := consensus.Finalized(d)
	// 2 tips (f1,f2), majority = 2 → only g is in both ancestor sets
	if len(final) != 1 || final[0] != "g" {
		t.Errorf("expected [g], got %v", final)
	}
}
