package sysinfo

// HostInfo retrieves the complete system information, including kernel name, version,
// architecture, and additional details about the environment.
func HostInfo() string {
	return uname(defaultCmdRunner, "-a")
}
