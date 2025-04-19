package main

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/Abdullah-zahoor/shardedchain/global"
	"github.com/Abdullah-zahoor/shardedchain/proof"
	"github.com/Abdullah-zahoor/shardedchain/sim"
	"github.com/Abdullah-zahoor/shardedchain/state"
	"github.com/Abdullah-zahoor/shardedchain/trie"
)

func main() {
	fmt.Println("ShardedChain starting…")

	// Phase 1: Merkle-trie smoke test
	root := trie.NewNode()
	key := []byte("account42")
	val := []byte("1000")
	root.Insert(key, val)
	fmt.Println("Trie root hash:", hex.EncodeToString(root.RootHash()))

	proof1, err := root.GetProof(key)
	if err != nil {
		panic(err)
	}
	fmt.Println("Trie proof valid?", trie.VerifyProof(root.RootHash(), key, proof1))

	// Phase 2: ShardManager smoke test
	mgr := state.NewManager(4)
	mgr.ApplyTx([]byte("alice"), []byte("500"))
	mgr.ApplyTx([]byte("bob"), []byte("250"))
	fmt.Println("Shard stats:", mgr.CollectStats())

	// Phase 5: Cross-shard proof demonstration
	srcKey := []byte("alice")
	dstKey := []byte("bob")
	srcIdx := mgr.ShardIndex(srcKey)
	dstIdx := mgr.ShardIndex(dstKey)
	cp, err := proof.GenerateCrossProof(
		srcIdx, dstIdx,
		srcKey, dstKey,
		[]byte("50"), // transfer amount
		mgr.GetTrie,  // function to retrieve each shard's trie
	)
	if err != nil {
		panic(err)
	}
	if err := cp.VerifyCrossProof(trie.VerifyProof); err != nil {
		fmt.Println("Cross-shard proof invalid:", err)
	} else {
		fmt.Println("Cross-shard proof valid? true")
	}

	// Phase 6: Global root assembly
	roots := mgr.ShardRoots()
	globalRoot := global.BuildGlobalRoot(roots)
	fmt.Println("Global root hash:", hex.EncodeToString(globalRoot))

	// Phase 8: Scheduler with rebalance
	scheduler := sim.NewScheduler(
		mgr,
		2*time.Second, // tick interval
		2,             // two active shards per tick
		2.0,           // split variance threshold
		0.5,           // merge variance threshold
	)
	scheduler.Start()

	// Phase 9: Workload generator
	keys := []string{"alice", "bob", "charlie", "dave", "eve", "frank", "grace", "heidi", "ivan", "judy"}
	sim.RunWorkload(scheduler, keys, 16, 50, 30*time.Second)

	// Wait for workload to complete
	time.Sleep(35 * time.Second)

	// Final stats and global root
	fmt.Println("Final shard stats:", mgr.CollectStats())
	finalGlobal := global.BuildGlobalRoot(mgr.ShardRoots())
	fmt.Println("Final global root:", hex.EncodeToString(finalGlobal))
}
