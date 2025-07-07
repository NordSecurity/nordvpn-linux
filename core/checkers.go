package core

type ExpirationChecker interface {
	// IsExpired checks if date in '2006-01-02 15:04:05' format has passed
	IsExpired(date string) bool
}
