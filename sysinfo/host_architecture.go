package sysinfo

import "github.com/NordSecurity/nordvpn-linux/log"

// GetHostArchitecture returns the processor architecture of the host system such as "x86_64",
// "aarch64", etc. Returns an empty string if the architecture cannot be determined.
func GetHostArchitecture() string {
	return getHostArchitecture(defaultCmdRunner)
}

func getHostArchitecture(runner cmdRunner) string {
	out, err := runner("uname", "-m")
	if err != nil {
		log.Errorf("failed to execute 'uname -m': %v", err)
		return ""
	}
	return out
}
