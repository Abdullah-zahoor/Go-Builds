package sim

import (
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/Abdullah-zahoor/shardedchain/state"
)

// Transaction represents a simple key/value write.
type Transaction struct {
	Key   []byte
	Value []byte
}

// Scheduler processes transactions in time-sliced fashion and performs dynamic rebalance.
type Scheduler struct {
	mgr            *state.ShardManager
	txQueue        chan Transaction
	tickInterval   time.Duration
	activeCount    int
	cursor         int
	splitThreshold float64
	mergeThreshold float64
}

// NewScheduler creates a scheduler with rebalance thresholds.
// mgr: your ShardManager
// tickInterval: how often to rotate/execute
// activeCount: number of shards active per tick
// splitThreshold: variance above which to split shards
// mergeThreshold: variance below which to merge shards
func NewScheduler(
	mgr *state.ShardManager,
	tickInterval time.Duration,
	activeCount int,
	splitThreshold, mergeThreshold float64,
) *Scheduler {
	return &Scheduler{
		mgr:            mgr,
		txQueue:        make(chan Transaction, 1000),
		tickInterval:   tickInterval,
		activeCount:    activeCount,
		splitThreshold: splitThreshold,
		mergeThreshold: mergeThreshold,
	}
}

// Submit enqueues a transaction.
func (s *Scheduler) Submit(key, value []byte) {
	s.txQueue <- Transaction{Key: key, Value: value}
}

// Start begins the scheduler loop.
func (s *Scheduler) Start() {
	ticker := time.NewTicker(s.tickInterval)
	go func() {
		for tick := range ticker.C {
			total := s.mgr.ShardCount()
			if total == 0 {
				continue
			}

			// Determine active shards round-robin
			active := make(map[int]bool, s.activeCount)
			for i := 0; i < s.activeCount; i++ {
				idx := (s.cursor + i) % total
				active[idx] = true
			}

			fmt.Printf("┌─ Tick %s | Active Shards: %v\n",
				tick.Format("15:04:05"), keys(active))

			// Process all queued transactions
			n := len(s.txQueue)
			for i := 0; i < n; i++ {
				tx := <-s.txQueue
				shardIdx := s.mgr.ShardIndex(tx.Key)
				if active[shardIdx] {
					s.mgr.ApplyTx(tx.Key, tx.Value)
					fmt.Printf("✅ Applied to shard %d: %q → %q\n",
						shardIdx, tx.Key, tx.Value)
				} else {
					fmt.Printf("⏭ Queued (inactive shard %d)\n", shardIdx)
					s.txQueue <- tx
				}
			}

			// Perform rebalance with proof
			rp := s.mgr.RebalanceWithProof(s.splitThreshold, s.mergeThreshold)
			if rp.Operation != "none" {
				fmt.Printf("⇄ Rebalance: %s on shards %v\n", rp.Operation, rp.ShardIndex)
				fmt.Printf("    Pre-roots:  %s\n", joinHex(rp.PreRoots))
				fmt.Printf("    Post-roots: %s\n", joinHex(rp.PostRoots))
			}

			fmt.Println("└───────────────────────────────")
			s.cursor = (s.cursor + s.activeCount) % total
		}
	}()
}

// keys returns a sorted slice of map keys.
func keys(m map[int]bool) []int {
	ks := make([]int, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	sort.Ints(ks)
	return ks
}

// joinHex formats a slice of byte-slices as a string of truncated hex values.
func joinHex(blist [][]byte) string {
	parts := make([]string, len(blist))
	for i, b := range blist {
		hexStr := hex.EncodeToString(b)
		if len(hexStr) > 8 {
			hexStr = hexStr[:8]
		}
		parts[i] = hexStr
	}
	return fmt.Sprintf("[%s]", strings.Join(parts, " "))
}
