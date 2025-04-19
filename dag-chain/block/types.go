package block

import "time"

// TXInput references a previous TX output.
type TXInput struct {
	PrevTxID    string
	OutputIndex int
}

// TXOutput represents a new unspent output.
type TXOutput struct {
	Value     uint64
	Recipient string
}

// TX is a UTXO‐style transaction.
type TX struct {
	ID      string
	Inputs  []TXInput
	Outputs []TXOutput
}

// Block can reference multiple parents.
type Block struct {
	ID        string
	Parents   []string  // parent block IDs
	TXs       []TX      // included transactions
	Timestamp time.Time // creation time
}
