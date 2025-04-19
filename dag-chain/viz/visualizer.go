package viz

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/Abdullah-zahoor/dagchain/dag"
)

// ASCII returns a simple topological listing of each node and its parents.
func ASCII(d *dag.DAG) string {
	// collect IDs and sort for deterministic output
	ids := make([]string, 0, len(d.Nodes))
	for id := range d.Nodes {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	var buf bytes.Buffer
	for _, id := range ids {
		node := d.Nodes[id]
		parentIDs := make([]string, len(node.Parents))
		for i, p := range node.Parents {
			parentIDs[i] = p.Block.ID
		}
		if len(parentIDs) == 0 {
			buf.WriteString(fmt.Sprintf("%s (genesis)\n", id))
		} else {
			buf.WriteString(fmt.Sprintf("%s -> [%s]\n",
				id, join(parentIDs, ",")))
		}
	}
	return buf.String()
}

// DOT returns a Graphviz DOT description of the DAG.
func DOT(d *dag.DAG) string {
	var buf bytes.Buffer
	buf.WriteString("digraph DAG {\n")
	// optional styling
	buf.WriteString("  node [shape=box fontname=\"Monospace\"];\n")

	// edges
	for _, node := range d.Nodes {
		for _, p := range node.Parents {
			buf.WriteString(fmt.Sprintf("  \"%s\" -> \"%s\";\n",
				p.Block.ID, node.Block.ID))
		}
	}
	buf.WriteString("}\n")
	return buf.String()
}

// helper: join slice of strings with sep
func join(ss []string, sep string) string {
	var buf bytes.Buffer
	for i, s := range ss {
		if i > 0 {
			buf.WriteString(sep)
		}
		buf.WriteString(s)
	}
	return buf.String()
}
