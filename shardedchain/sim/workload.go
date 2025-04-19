package sim

import (
	"math/rand"
	"time"
)

// RunWorkload generates a stream of random transactions to the scheduler.
// keys: slice of keys to choose from
// valueSize: number of bytes in each value
// rate: approximate number of transactions per second
// duration: total time to run the workload
func RunWorkload(
	s *Scheduler,
	keys []string,
	valueSize int,
	rate int,
	duration time.Duration,
) {
	rand.Seed(time.Now().UnixNano())

	ticker := time.NewTicker(time.Second / time.Duration(rate))
	end := time.After(duration)

	go func() {
		for {
			select {
			case <-ticker.C:
				// pick random key
				k := keys[rand.Intn(len(keys))]
				// generate random value
				v := make([]byte, valueSize)
				rand.Read(v)
				// submit transaction
				s.Submit([]byte(k), v)

			case <-end:
				ticker.Stop()
				return
			}
		}
	}()
}
