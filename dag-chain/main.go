package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/Abdullah-zahoor/dagchain/block"
	"github.com/Abdullah-zahoor/dagchain/consensus"
	"github.com/Abdullah-zahoor/dagchain/dag"
	"github.com/Abdullah-zahoor/dagchain/sim"
	"github.com/Abdullah-zahoor/dagchain/viz"
)

func main() {
	// --- Bootstrap & Simulation (unchanged) ---
	d := dag.NewDAG()
	genesis := &block.Block{ID: "genesis", Parents: nil, TXs: nil, Timestamp: time.Now()}
	initialUTXO := make(block.UTXOSet)
	if err := d.AddGenesis(genesis, initialUTXO); err != nil {
		panic(err)
	}
	fmt.Println("✅ Genesis added")

	simulator := sim.NewSimulator(d)
	fmt.Println("▶️ Starting simulation of 3 validators for 5s…")
	simulator.Run(3, 5*time.Second)
	fmt.Println("⏹ Simulation complete")

	// --- Consensus & Visualization (unchanged) ---
	if tip := consensus.HeaviestTip(d); tip != nil {
		fmt.Printf("🏆 Heaviest tip: %s (weight=%d)\n", tip.Block.ID, tip.Weight)
	}
	consensus.PruneBranches(d)
	fmt.Print("Remaining nodes:")
	for id := range d.Nodes {
		fmt.Printf(" %s", id)
	}
	fmt.Println()
	fmt.Printf("🔒 Finalized: %v\n", consensus.Finalized(d))

	// dump dot
	if err := os.WriteFile("dag.dot", []byte(viz.DOT(d)), 0o644); err != nil {
		panic(err)
	}
	fmt.Println("· Wrote dag.dot (use `dot -Tpng dag.dot -o dag.png`)")

	// --- HTTP API ---
	http.HandleFunc("/tips", func(w http.ResponseWriter, r *http.Request) {
		tips := consensus.HeaviestTip(d)
		json.NewEncoder(w).Encode(tips)
	})
	http.HandleFunc("/finalized", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(consensus.Finalized(d))
	})
	http.HandleFunc("/ascii", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, viz.ASCII(d))
	})
	http.HandleFunc("/dot", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/vnd.graphviz")
		fmt.Fprint(w, viz.DOT(d))
	})

	fmt.Println("🚀 HTTP API listening on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
