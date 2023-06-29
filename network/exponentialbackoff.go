package network

import (
	"math/rand"
	"time"
)

func ExponentialBackoff(tries int) time.Duration {
	var minSecs, maxSecs int
	switch {
	case tries < 3:
		minSecs = 5
		maxSecs = 10
	case tries < 10:
		minSecs = 10
		maxSecs = 60
	case tries < 20:
		minSecs = 60
		maxSecs = 300
	default:
		minSecs = 300
		maxSecs = 600
	}

	// #nosec G404 -- not used for cryptographic purposes
	return time.Duration(rand.Intn(maxSecs-minSecs+1)+minSecs) * time.Second
}
