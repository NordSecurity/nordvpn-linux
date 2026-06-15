package internal

import "time"

// SystemTimeExpirationChecker checks whether a date string has passed
// using the system clock.
type SystemTimeExpirationChecker struct{}

// IsExpired reports whether the date in ServerDateFormat has passed.
func (SystemTimeExpirationChecker) IsExpired(expiryTime string) bool {
	if expiryTime == "" {
		return true
	}

	expiry, err := time.Parse(ServerDateFormat, expiryTime)
	if err != nil {
		return true
	}

	return time.Now().After(expiry)
}
