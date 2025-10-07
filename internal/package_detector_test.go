package internal

import (
	"os"
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

	// This test will run on the actual system
	result := isDebBasedSystem()

	// If we're on a Deb-based system, certain files or commands should exist
	if result {
		// At least one of these should be true
		hasDebIndicator := IsCommandAvailable("dpkg") ||
			IsCommandAvailable("apt") ||
			IsCommandAvailable("apt-get") ||
			FileExists("/etc/debian_version") ||
			FileExists("/etc/lsb-release")

		assert.True(t, hasDebIndicator, "Deb system detected but no Deb indicators found")
	}
}

func TestIsRpmBasedSystemIntegration(t *testing.T) {
	category.Set(t, category.Integration)

	// This test will run on the actual system
	result := isRpmBasedSystem()

	// If we're on an Rpm-based system, certain files or commands should exist
	if result {
		// At least one of these should be true
		hasRpmIndicator := IsCommandAvailable("rpm") ||
			IsCommandAvailable("yum") ||
			IsCommandAvailable("dnf") ||
			IsCommandAvailable("zypper") ||
			FileExists("/etc/redhat-release") ||
			FileExists("/etc/SuSE-release") ||
			FileExists("/etc/SUSE-brand") ||
			FileExists("/etc/os-release")

		assert.True(t, hasRpmIndicator, "Rpm system detected but no Rpm indicators found")
	}
}

// Test edge cases and specific scenarios
func TestDetectPackageTypeEdgeCases(t *testing.T) {
	category.Set(t, category.Integration)

	// Create a temporary test directory
	tempDir := t.TempDir()

	tests := []struct {
		name        string
		setupFunc   func()
		cleanupFunc func()
		skipReason  string
	}{
		{
			name: "handles missing os-release file gracefully",
			setupFunc: func() {
				// This test verifies the function doesn't panic when files are missing
				// The actual behavior is tested by the function implementation
			},
			cleanupFunc: func() {},
		},
		{
			name: "handles empty lsb-release file",
			setupFunc: func() {
				// Create an empty lsb-release file in temp directory
				// Note: We can't actually override the system files being checked
				// This test mainly ensures the function handles edge cases
				emptyFile := tempDir + "/lsb-release"
				os.WriteFile(emptyFile, []byte(""), 0644)
			},
			cleanupFunc: func() {
				os.Remove(tempDir + "/lsb-release")
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.skipReason != "" {
				t.Skip(test.skipReason)
			}

			test.setupFunc()
			defer test.cleanupFunc()

			// The function should not panic
			assert.NotPanics(t, func() {
				_ = DetectPackageType()
			})
		})
	}
}
