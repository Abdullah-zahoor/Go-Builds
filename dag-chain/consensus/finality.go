package consensus

import "github.com/Abdullah-zahoor/dagchain/dag"

// Finalized returns the IDs of all blocks that appear in the ancestor
// sets of a strict majority of current tips.
func Finalized(d *dag.DAG) []string {
	tips := Tips(d)
	if len(tips) == 0 {
		return nil
	}

	// count how many tips “confirm” each block
	counts := make(map[string]int)
	for _, tip := range tips {
		as := ancestorSet(tip)
		for id := range as {
			counts[id]++
		}
	}

	// need > 50% of tips
	minCount := len(tips)/2 + 1
	var finals []string
	for id, cnt := range counts {
		if cnt >= minCount {
			finals = append(finals, id)
		}
	}
	return finals
}

// reuse the same ancestorSet helper from resolver.go
