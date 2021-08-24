package ipfscopy

import (
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestCalculateRateLimitDuration(t *testing.T) {
	tests := []struct {
		name      string
		reqPerSec int
		duration  time.Duration
	}{
		{
			"default",
			10,
			100 * time.Millisecond,
		},
		{
			"slow",
			2,
			500 * time.Millisecond,
		},
		{
			"fast",
			20,
			50 * time.Millisecond,
		},
		{
			"odd",
			3,
			333 * time.Millisecond,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.duration, CalculateRateLimitDuration(tc.reqPerSec))
		})
	}
}
