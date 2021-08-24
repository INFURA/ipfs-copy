package ipfscopy

import "time"

func CalculateRateLimitDuration(maxReqsPerSec int) time.Duration {
	sleepPerReqInMs := int64(time.Second/time.Millisecond) / int64(maxReqsPerSec)

	return time.Duration(sleepPerReqInMs) * time.Millisecond
}
