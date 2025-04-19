package sim

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/Abdullah-zahoor/dagchain/block"
	"github.com/Abdullah-zahoor/dagchain/consensus"
	"github.com/Abdullah-zahoor/dagchain/dag"
)

// Simulator holds the shared DAG and a mutex for safe concurrent access.
type Simulator struct {
	DAG *dag.DAG
	mu  sync.Mutex
}

// NewSimulator returns a new Simulator instance.
func NewSimulator(d *dag.DAG) *Simulator {
	return &Simulator{DAG: d}
}

// Run starts `numValidators` goroutines that each propose blocks for `duration`.
func (s *Simulator) Run(numValidators int, duration time.Duration) {
	stop := make(chan struct{})
	var wg sync.WaitGroup

	// Launch validators
	for i := 0; i < numValidators; i++ {
		wg.Add(1)
		go s.validator(i, stop, &wg)
	}

	// Let them run
	time.Sleep(duration)
	close(stop)
	wg.Wait()
}

// validator is a loop that proposes blocks until stop is closed.
func (s *Simulator) validator(id int, stop <-chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()
	randSrc := rand.New(rand.NewSource(time.Now().UnixNano() + int64(id)))

	for {
		select {
		case <-stop:
			return
		default:
			s.mu.Lock()
			// Pick the heaviest tip as parent (or genesis if none)
			parent := consensus.HeaviestTip(s.DAG)
			if parent == nil {
				parent = s.DAG.Nodes["genesis"]
			}

			// Create a new “mint” transaction
			tx := block.TX{
				ID:     fmt.Sprintf("tx-v%d-%d", id, time.Now().UnixNano()),
				Inputs: nil,
				Outputs: []block.TXOutput{
					{
						Value:     uint64(randSrc.Intn(100) + 1),
						Recipient: fmt.Sprintf("V%d", id),
					},
				},
			}

			// Create and add the block
			blk := &block.Block{
				ID:        fmt.Sprintf("block-v%d-%d", id, time.Now().UnixNano()),
				Parents:   []string{parent.Block.ID},
				TXs:       []block.TX{tx},
				Timestamp: time.Now(),
			}
			_ = s.DAG.AddBlock(blk)
			s.mu.Unlock()

			// Sleep a bit before proposing the next block
			time.Sleep(time.Duration(randSrc.Intn(500)+100) * time.Millisecond)
		}
	}
}
