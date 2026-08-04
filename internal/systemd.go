package internal

// IsSystemShutdown detect if system is being shutdown
func IsSystemShutdown() bool {
	// https://www.freedesktop.org/software/systemd/man/latest/shutdown.html
	return FileExists("/run/nologin")
}

// IsSystemd detect if system is running systemd
func IsSystemd() bool {
	// https://www.freedesktop.org/software/systemd/man/latest/sd_booted.html
	return FileExists("/run/systemd/system")
}
