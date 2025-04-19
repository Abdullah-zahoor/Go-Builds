package block

import (
	"errors"
	"fmt"
)

// UTXOKey uniquely identifies a discrete output.
type UTXOKey struct {
	TxID     string
	OutIndex int
}

// UTXOSet maps each UTXOKey to its corresponding output.
type UTXOSet map[UTXOKey]TXOutput

func (u UTXOSet) Clone() UTXOSet {
	dup := make(UTXOSet, len(u))
	for k, v := range u {
		dup[k] = v
	}
	return dup
}

func (u UTXOSet) ApplyTx(tx TX) error {
	// Check & remove inputs
	for _, in := range tx.Inputs {
		key := UTXOKey{TxID: in.PrevTxID, OutIndex: in.OutputIndex}
		if _, exists := u[key]; !exists {
			return fmt.Errorf("input not found or already spent: %v", key)
		}
		delete(u, key)
	}
	for idx, out := range tx.Outputs {
		key := UTXOKey{TxID: tx.ID, OutIndex: idx}
		if _, exists := u[key]; exists {
			// Should never happen: Tx IDs must be unique
			return errors.New("duplicate output key: " + fmt.Sprint(key))
		}
		u[key] = out
	}

	return nil
}
