package state

import (
	"hash/fnv"
	"sync"

	"github.com/Abdullah-zahoor/shardedchain/shard"
	"github.com/Abdullah-zahoor/shardedchain/trie"
)

// RebalanceProof describes how shards changed during a rebalance.
type RebalanceProof struct {
	PreRoots   [][]byte // shard roots before rebalance
	PostRoots  [][]byte // shard roots after rebalance
	Operation  string   // "split", "merge", or "none"
	ShardIndex []int    // affected shard indices
}

// ShardManager holds all shards and provides routing, rebalance, and trie access.
type ShardManager struct {
	Shards []*shard.Shard
	mu     sync.RWMutex
}

// NewManager creates a manager with numShards empty shards.
func NewManager(numShards int) *ShardManager {
	m := &ShardManager{Shards: make([]*shard.Shard, 0, numShards)}
	for i := 0; i < numShards; i++ {
		m.Shards = append(m.Shards, shard.NewShard())
	}
	return m
}

// ShardCount returns the number of shards.
func (m *ShardManager) ShardCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.Shards)
}

// ShardIndex returns the index for a given key.
func (m *ShardManager) ShardIndex(key []byte) int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.shardIndex(key)
}

// ShardRoots returns each shard's current root hash.
func (m *ShardManager) ShardRoots() [][]byte {
	m.mu.RLock()
	defer m.mu.RUnlock()
	roots := make([][]byte, len(m.Shards))
	for i, s := range m.Shards {
		roots[i] = s.Root()
	}
	return roots
}

// ApplyTx writes a key/value to its shard.
func (m *ShardManager) ApplyTx(key, value []byte) {
	idx := m.shardIndex(key)
	m.mu.Lock()
	m.Shards[idx].Apply(key, value)
	m.mu.Unlock()
}

// CollectStats returns mutation counts.
func (m *ShardManager) CollectStats() []int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	stats := make([]int, len(m.Shards))
	for i, s := range m.Shards {
		stats[i] = s.Mutations
	}
	return stats
}

// GetTrie exposes the underlying trie for a shard.
func (m *ShardManager) GetTrie(shardIdx int) *trie.Node {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.Shards[shardIdx].Tree
}

// Rebalance performs split/merge based on variance thresholds.
func (m *ShardManager) Rebalance(thresholdSplit, thresholdMerge float64) {
	stats := m.CollectStats()
	_, variance := calcStats(stats)
	if variance > thresholdSplit {
		idx := indexOfMax(stats)
		m.mu.Lock()
		m.splitShard(idx)
		m.mu.Unlock()
	} else if variance < thresholdMerge && len(m.Shards) > 1 {
		i1, i2 := twoMinIndices(stats)
		m.mu.Lock()
		m.mergeShards(i1, i2)
		m.mu.Unlock()
	}
}

// RebalanceWithProof rebalances and returns a proof without deadlock.
func (m *ShardManager) RebalanceWithProof(thresholdSplit, thresholdMerge float64) *RebalanceProof {
	m.mu.Lock()
	defer m.mu.Unlock()

	// take pre-roots and stats under write lock
	pre := make([][]byte, len(m.Shards))
	stats := make([]int, len(m.Shards))
	for i, s := range m.Shards {
		pre[i] = s.Root()
		stats[i] = s.Mutations
	}
	_, variance := calcStats(stats)

	var op string
	var affected []int

	if variance > thresholdSplit {
		op = "split"
		i := indexOfMax(stats)
		affected = []int{i}
		m.splitShard(i)
	} else if variance < thresholdMerge && len(m.Shards) > 1 {
		op = "merge"
		i1, i2 := twoMinIndices(stats)
		affected = []int{i1, i2}
		m.mergeShards(i1, i2)
	} else {
		op = "none"
	}

	// snapshot post-roots
	post := make([][]byte, len(m.Shards))
	for i, s := range m.Shards {
		post[i] = s.Root()
	}

	return &RebalanceProof{PreRoots: pre, PostRoots: post, Operation: op, ShardIndex: affected}
}

// shardIndex hashes a key to select a shard.
func (m *ShardManager) shardIndex(key []byte) int {
	h := fnv.New32a()
	h.Write(key)
	return int(h.Sum32()) % len(m.Shards)
}

// calcStats computes mean & variance.
func calcStats(data []int) (mean, variance float64) {
	n := float64(len(data))
	if n == 0 {
		return
	}
	var sum float64
	for _, v := range data {
		sum += float64(v)
	}
	mean = sum / n
	var ss float64
	for _, v := range data {
		d := float64(v) - mean
		ss += d * d
	}
	variance = ss / n
	return
}

// indexOfMax finds largest element index.
func indexOfMax(data []int) int {
	max := 0
	for i, v := range data {
		if v > data[max] {
			max = i
		}
	}
	return max
}

// twoMinIndices finds two smallest.
func twoMinIndices(data []int) (int, int) {
	min1, min2 := 0, 1
	if data[min2] < data[min1] {
		min1, min2 = min2, min1
	}
	for i := 2; i < len(data); i++ {
		if data[i] < data[min1] {
			min2 = min1
			min1 = i
		} else if data[i] < data[min2] {
			min2 = i
		}
	}
	return min1, min2
}

// splitShard redistributes by high bit of first byte.
func (m *ShardManager) splitShard(idx int) {
	old := m.Shards[idx]
	s1, s2 := shard.NewShard(), shard.NewShard()
	for _, kv := range old.Tree.Traverse() {
		if len(kv.Key) > 0 && kv.Key[0]&0x80 == 0 {
			s1.Apply(kv.Key, kv.Value)
		} else {
			s2.Apply(kv.Key, kv.Value)
		}
	}
	new := append([]*shard.Shard{}, m.Shards[:idx]...)
	new = append(new, s1, s2)
	new = append(new, m.Shards[idx+1:]...)
	m.Shards = new
}

// mergeShards combines two shards.
func (m *ShardManager) mergeShards(i1, i2 int) {
	if i2 < i1 {
		i1, i2 = i2, i1
	}
	s1, s2 := m.Shards[i1], m.Shards[i2]
	merged := shard.NewShard()
	for _, kv := range s1.Tree.Traverse() {
		merged.Apply(kv.Key, kv.Value)
	}
	for _, kv := range s2.Tree.Traverse() {
		merged.Apply(kv.Key, kv.Value)
	}
	new := append([]*shard.Shard{}, m.Shards[:i1]...)
	new = append(new, merged)
	new = append(new, m.Shards[i2+1:]...)
	m.Shards = new
}
