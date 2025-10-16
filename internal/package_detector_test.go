package internal

import (
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

// Integration tests that use real system calls
func TestDetectPackageTypeIntegration(t *testing.T) {
	category.Set(t, category.Integration)

	// This test will run on the actual system
	result := DetectPackageType()

	// We can't predict the result, but it should be one of the valid values
	assert.Contains(t, []PackageType{
		PackageTypeDeb,
		PackageTypeRpm,
		PackageTypeUnknown,
	}, result)

	// The result should be consistent across multiple calls
	result2 := DetectPackageType()
	assert.Equal(t, result, result2)
}

func TestIsDebBasedSystemIntegration(t *testing.T) {
	category.Set(t, category.Integration)

	// We do run tests on debian based system
	result := isDebBasedSystem()

	if result {
		hasDebIndicator := IsCommandAvailable("dpkg") ||
			IsCommandAvailable("apt") ||
			IsCommandAvailable("apt-get") ||
			FileExists("/etc/debian_version") ||
			FileExists("/etc/os-release") ||
			FileExists("/etc/lsb-release")

		assert.True(t, hasDebIndicator, "Deb system detected but no Deb indicators found")
	}
}
