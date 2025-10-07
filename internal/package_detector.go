package internal

import "strings"

// PackageType represents the type of package manager
// Deb-based systems are checked first and take priority over Rpm-based systems
type PackageType string

const (
	// PackageTypeUnknown represents an unknown or unsupported package manager
	PackageTypeUnknown PackageType = "unknown"
	// PackageTypeDeb represents Deb-based package managers (APT/DPKG)
	PackageTypeDeb PackageType = "deb"
	// PackageTypeRpm represents Rpm-based package managers (YUM/DNF/Zypper)
	PackageTypeRpm PackageType = "rpm"
)

// DetectPackageType detects whether the system uses Deb or Rpm package management
func DetectPackageType() PackageType {
	// Check for Deb-based systems
	if isDebBasedSystem() {
		return PackageTypeDeb
	}

	// Check for Rpm-based systems
	if isRpmBasedSystem() {
		return PackageTypeRpm
	}

	return PackageTypeUnknown
}

// isDebBasedSystem checks if the system uses Deb packages
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

	// Check multiple files
	releaseFiles := []string{"/etc/os-release", "/etc/lsb-release"}

	for _, file := range releaseFiles {
		if FileExists(file) {
			data, err := FileRead(file)
			if err == nil && len(data) > 0 {
				content := string(data)
				if strings.Contains(content, "Ubuntu") || strings.Contains(content, "Debian") {
					return true
				}
			}
		}
	}

	return false
}

// isRpmBasedSystem checks if the system uses Rpm packages
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

// String returns the string representation of the PackageType
func (pm PackageType) String() string {
	return string(pm)
}

// IsDeb returns true if the package manager is Deb-based
func (pm PackageType) IsDeb() bool {
	return pm == PackageTypeDeb
}

// IsRpm returns true if the package manager is Rpm-based
func (pm PackageType) IsRpm() bool {
	return pm == PackageTypeRpm
}
