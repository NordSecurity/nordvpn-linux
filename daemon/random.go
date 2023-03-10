package daemon

import "math/rand"

func randFloat(seed int64, min, max float64) float64 {
	// #nosec G404 -- not used for cryptographic purposes
	rng := rand.New(rand.NewSource(seed))
	return min + rng.Float64()*(max-min)
}
