package internal

type Environment string

const (
	// Development defines development environment
	Development Environment = "dev"
	// QA defines qa environment
	QA Environment = "qa"
	// Production defines production environment
	Production Environment = "prod"
	// Downloader modifies configs and servers jobs
	Downloader Environment = "downloader"
)

// IsProdEnv short hand of condition check, for clear reading
func IsProdEnv(env string) bool {
	return Environment(env) == Production
}

// IsDevEnv short hand of condition check, for clear reading
func IsDevEnv(env string) bool {
	return !IsProdEnv(env)
}
