package internal

import "strings"

// PackageManager represents the type of package manager
type PackageManager string

const (
	// PackageManagerUnknown represents an unknown or unsupported package manager
	PackageManagerUnknown PackageManager = "unknown"
	// PackageManagerDEB represents DEB-based package managers (APT/DPKG)
	PackageManagerDEB PackageManager = "deb"
	// PackageManagerRPM represents RPM-based package managers (YUM/DNF/Zypper)
	PackageManagerRPM PackageManager = "rpm"
)

// DetectPackageManager detects whether the system uses DEB or RPM package management
func DetectPackageManager() PackageManager {
	// Check for DEB-based systems
	if isDebBasedSystem() {
		return PackageManagerDEB
	}

	// Check for RPM-based systems
	if isRpmBasedSystem() {
		return PackageManagerRPM
	}

	return PackageManagerUnknown
}

// isDebBasedSystem checks if the system uses DEB packages
func isDebBasedSystem() bool {
	// Check for dpkg command
	if IsCommandAvailable("dpkg") {
		return true
	}

	// Check for apt or apt-get commands
	if IsCommandAvailable("apt") || IsCommandAvailable("apt-get") {
		return true
	}

	// Check for Debian/Ubuntu specific files
	if FileExists("/etc/debian_version") {
		return true
	}

	// Check for Ubuntu specific files
	if FileExists("/etc/lsb-release") {
		data, err := FileRead("/etc/lsb-release")
		if err == nil && len(data) > 0 {
			content := string(data)
			if strings.Contains(content, "Ubuntu") || strings.Contains(content, "Debian") {
				return true
			}
		}
	}

	return false
}

// isRpmBasedSystem checks if the system uses RPM packages
func isRpmBasedSystem() bool {
	// Check for rpm command
	if IsCommandAvailable("rpm") {
		return true
	}

	// Check for YUM/DNF/Zypper commands
	if IsCommandAvailable("yum") || IsCommandAvailable("dnf") || IsCommandAvailable("zypper") {
		return true
	}

	// Check for Red Hat/CentOS/Fedora specific files
	if FileExists("/etc/redhat-release") {
		return true
	}

	// Check for SUSE specific files
	if FileExists("/etc/SuSE-release") || FileExists("/etc/SUSE-brand") {
		return true
	}

	// Check for openSUSE specific files
	if FileExists("/etc/os-release") {
		data, err := FileRead("/etc/os-release")
		if err == nil && len(data) > 0 {
			content := string(data)
			if strings.Contains(content, "openSUSE") || strings.Contains(content, "SUSE") {
				return true
			}
		}
	}

	return false
}

// String returns the string representation of the PackageManager
func (pm PackageManager) String() string {
	return string(pm)
}

// IsDEB returns true if the package manager is DEB-based
func (pm PackageManager) IsDEB() bool {
	return pm == PackageManagerDEB
}

// IsRPM returns true if the package manager is RPM-based
func (pm PackageManager) IsRPM() bool {
	return pm == PackageManagerRPM
}
