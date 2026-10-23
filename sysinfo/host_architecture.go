package sysinfo

// GetHostArchitecture returns the processor architecture of the host system such as "x86_64",
// "aarch64", etc.
func GetHostArchitecture() string {
	return uname(defaultCmdRunner, "-m")
}
