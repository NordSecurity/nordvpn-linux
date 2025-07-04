package sysinfo

// GetHostArchitecture returns the processor architecture of the host system such as "amd64",
// "arm64", etc.
func GetHostArchitecture() string {
	return uname(defaultCmdRunner, "-m")
}
