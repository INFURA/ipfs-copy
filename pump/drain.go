package pump

import (
	"time"

	ipfsPump "github.com/INFURA/ipfs-pump/pump"
)

type RateLimitedDrain struct {
	drain           ipfsPump.CountedDrain
	sleepAfterDrain time.Duration
}

var _ ipfsPump.CountedDrain = (*ipfsPump.CounterDrain)(nil)

func NewRateLimitedDrain(drain ipfsPump.Drain, sleepAfterDrain time.Duration) ipfsPump.CountedDrain {
	countedDrain := ipfsPump.NewCountedDrain(drain)

	return &RateLimitedDrain{drain: countedDrain, sleepAfterDrain: sleepAfterDrain}
}

func (d *RateLimitedDrain) Drain(block ipfsPump.Block) error {
	err := d.drain.Drain(block)
	if err != nil {
		return err
	}

	// Avoid getting rate limited
	time.Sleep(d.sleepAfterDrain)

	return nil
}

func (d *RateLimitedDrain) SuccessfulBlocksCount() uint64 {
	return d.drain.SuccessfulBlocksCount()
}
