package daemon

import "math"

const (
	Alpha  = 0.7
	Beta   = -0.15
	Lambda = 1
	K      = 0.5
	W      = 0.5
	Fi     = 7
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

func obfuscationPenalty(obfuscated bool, timestamp, timestampMin, timestampMax int64) float64 {
	if obfuscated {
		return Beta*math.Pow((float64(timestamp)-float64(timestampMin))/(float64(timestampMax)-float64(timestampMin)), Fi) + Lambda
	}
	return 0
}

func penalty(
	obfuscated bool,
	distance, distanceMin, distanceMax float64,
	timestamp, timestampMin, timestampMax int64,
	load int64,
	userCountryCode, serverCountryCode string,
	hubScore *float64,
	randomComponent float64,
) (float64, float64) {
	distanceP := distancePenalty(distance, distanceMin, distanceMax)
	loadP := loadPenalty(load)
	obfuscationP := obfuscationPenalty(obfuscated, timestamp, timestampMin, timestampMax)
	countryP := countryPenalty(userCountryCode, serverCountryCode)
	hubP := hubPenalty(hubScore)
	partialPenalty := distanceP + randomComponent + obfuscationP - countryP*hubP
	return partialPenalty + loadP, partialPenalty
}
