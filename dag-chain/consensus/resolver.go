package consensus

import (
	"fmt"

	"github.com/Abdullah-zahoor/dagchain/dag"
)

// Tips returns all tip nodes (those with no children).
func Tips(d *dag.DAG) []*dag.Node {
	var tips []*dag.Node
	for _, n := range d.Nodes {
		if len(n.Children) == 0 {
			tips = append(tips, n)
		}
	}
	return tips
}

// HeaviestTip picks the tip with the highest cumulative weight.
func HeaviestTip(d *dag.DAG) *dag.Node {
	tips := Tips(d)
	if len(tips) == 0 {
		return nil
	}
	heaviest := tips[0]
	for _, t := range tips[1:] {
		if t.Weight > heaviest.Weight {
			heaviest = t
		}
	}
	return heaviest
}

// ancestorSet collects all ancestor IDs of a node (including itself).
func ancestorSet(n *dag.Node) map[string]struct{} {
	set := make(map[string]struct{})
	var dfs func(c *dag.Node)
	dfs = func(c *dag.Node) {
		if _, seen := set[c.Block.ID]; seen {
			return
		}
		set[c.Block.ID] = struct{}{}
		for _, p := range c.Parents {
			dfs(p)
		}
	}
	dfs(n)
	return set
}

// PruneBranches removes from d.Nodes any node not on the heaviest-tip ancestor chain.
func PruneBranches(d *dag.DAG) {
	heaviest := HeaviestTip(d)
	if heaviest == nil {
		fmt.Println("no tips to prune")
		return
	}

	keep := ancestorSet(heaviest)
	for id, node := range d.Nodes {
		if _, ok := keep[id]; !ok {
			// unlink from parents
			for _, p := range node.Parents {
				var children []*dag.Node
				for _, c := range p.Children {
					if c.Block.ID != id {
						children = append(children, c)
					}
				}
				p.Children = children
			}
			// delete the node
			delete(d.Nodes, id)
		}
	}
	fmt.Printf("🔪 Pruned branches; kept path to %q\n", heaviest.Block.ID)
}
