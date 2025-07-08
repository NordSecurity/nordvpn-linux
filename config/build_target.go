package config

// BuildTarget mirrors values passed to a compiler when building
type BuildTarget struct {
	Version      string
	Environment  string
	Architecture string
}
