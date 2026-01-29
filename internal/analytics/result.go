// Package analytics provides shared utilities for analytics events across the application.
package analytics

// Result constants for analytics events
const (
	ResultSuccess = "success"
	ResultFailure = "failure"
)

// BoolToResult converts a boolean success value to a result string.
func BoolToResult(success bool) string {
	if success {
		return ResultSuccess
	}
	return ResultFailure
}
