package pump

import (
	ipfsPump "github.com/INFURA/ipfs-pump/pump"
)

type RateLimitedDrain struct {
	drain ipfsPump.CountedDrain
}

var _ ipfsPump.CountedDrain = (*ipfsPump.CounterDrain)(nil)

func NewRateLimitedDrain(drain ipfsPump.Drain) ipfsPump.CountedDrain {
	countedDrain := ipfsPump.NewCountedDrain(drain)

	return &RateLimitedDrain{drain: countedDrain}
}

func (d *RateLimitedDrain) Drain(block ipfsPump.Block) error {
	err := d.drain.Drain(block)
	if err != nil {
		return err
	}

	return nil
}

func (d *RateLimitedDrain) SuccessfulBlocksCount() uint64 {
	return d.drain.SuccessfulBlocksCount()
}
