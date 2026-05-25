package analytics

// CommonContextPaths defines the standard context paths that should be included
// in most analytics events across the application.
var CommonContextPaths = []string{
	// Device context - captures hardware/OS information
	"device.*",

	// Application identification
	"application.nordvpnapp.version",
	"application.nordvpnapp.platform",
}

// GetCommonContextPaths returns a copy of the common context paths.
func GetCommonContextPaths() []string {
	paths := make([]string, len(CommonContextPaths))
	copy(paths, CommonContextPaths)
	return paths
}

// MergeContextPaths combines common context paths with additional feature-specific paths.
func MergeContextPaths(additionalPaths ...string) []string {
	total := len(CommonContextPaths) + len(additionalPaths)
	merged := make([]string, 0, total)
	merged = append(merged, CommonContextPaths...)
	merged = append(merged, additionalPaths...)
	return merged
}
