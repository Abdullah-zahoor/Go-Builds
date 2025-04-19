package block_test

import (
	"testing"

	"github.com/Abdullah-zahoor/dagchain/block"
)

func TestApplyTx_Success(t *testing.T) {
	utxo := make(block.UTXOSet)

	// Mint a new coin
	tx := block.TX{
		ID:      "t1",
		Inputs:  nil,
		Outputs: []block.TXOutput{{Value: 10, Recipient: "Alice"}},
	}
	if err := utxo.ApplyTx(tx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Expect the UTXO to contain that output
	key := block.UTXOKey{TxID: "t1", OutIndex: 0}
	if out, ok := utxo[key]; !ok {
		t.Error("expected utxo to contain new output")
	} else if out.Value != 10 || out.Recipient != "Alice" {
		t.Errorf("got wrong output: %+v", out)
	}
}

func TestApplyTx_DoubleSpend(t *testing.T) {
	utxo := make(block.UTXOSet)
	// Prepare a UTXO to spend
	utxo[block.UTXOKey{"t0", 0}] = block.TXOutput{Value: 5, Recipient: "Bob"}

	// First spend should succeed
	tx1 := block.TX{
		ID:     "t1",
		Inputs: []block.TXInput{{PrevTxID: "t0", OutputIndex: 0}},
	}
	if err := utxo.ApplyTx(tx1); err != nil {
		t.Fatalf("first spend failed: %v", err)
	}

	// Second spend of the same input should error
	tx2 := block.TX{
		ID:     "t2",
		Inputs: []block.TXInput{{PrevTxID: "t0", OutputIndex: 0}},
	}
	if err := utxo.ApplyTx(tx2); err == nil {
		t.Error("expected double‑spend error, got nil")
	}
}
