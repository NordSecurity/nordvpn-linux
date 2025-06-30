package sysinfo

// GetHostInfo retrieves the complete system information, including kernel name, version,
// architecture, and additional details about the environment.
func GetHostInfo() string {
	return uname(defaultCmdRunner, "-a")
}
