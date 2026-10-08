package daemon

import "math"

const (
	Alpha = 0.7
	W     = 0.5
)

func distancePenalty(distance, distanceMin, distanceMax float64) float64 {
	return Alpha * math.Pow((distance-distanceMin)/(distanceMax-distanceMin), W)
}

func countryPenalty(userCountryCode, serverCountryCode string) float64 {
	if userCountryCode == serverCountryCode {
		return 0
	}
	return 1
}

func loadPenalty(load int64) float64 {
	fLoad := float64(load) / 10
	return math.Pow(fLoad, fLoad)
}

func hubPenalty(hubScore *float64) float64 {
	if hubScore != nil {
		return *hubScore
	}
	return 0
}

func penalty(
	distance, distanceMin, distanceMax float64,
	load int64,
	userCountryCode, serverCountryCode string,
	hubScore *float64,
	randomComponent float64,
) (float64, float64) {
	distanceP := distancePenalty(distance, distanceMin, distanceMax)
	loadP := loadPenalty(load)
	countryP := countryPenalty(userCountryCode, serverCountryCode)
	hubP := hubPenalty(hubScore)
	partialPenalty := distanceP + randomComponent - countryP*hubP
	return partialPenalty + loadP, partialPenalty
}
