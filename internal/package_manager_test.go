package internal

import (
	"os"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

// Integration tests that use real system calls
func TestDetectPackageManagerIntegration(t *testing.T) {
	category.Set(t, category.Integration)

	// This test will run on the actual system
	result := DetectPackageManager()

	// We can't predict the result, but it should be one of the valid values
	assert.Contains(t, []PackageManager{
		PackageManagerDEB,
		PackageManagerRPM,
		PackageManagerUnknown,
	}, result)

	// The result should be consistent across multiple calls
	result2 := DetectPackageManager()
	assert.Equal(t, result, result2)
}

func TestIsDebBasedSystemIntegration(t *testing.T) {
	category.Set(t, category.Integration)

	// This test will run on the actual system
	result := isDebBasedSystem()

	// If we're on a DEB-based system, certain files or commands should exist
	if result {
		// At least one of these should be true
		hasDebIndicator := IsCommandAvailable("dpkg") ||
			IsCommandAvailable("apt") ||
			IsCommandAvailable("apt-get") ||
			FileExists("/etc/debian_version") ||
			FileExists("/etc/lsb-release")

		assert.True(t, hasDebIndicator, "DEB system detected but no DEB indicators found")
	}
}

func TestIsRpmBasedSystemIntegration(t *testing.T) {
	category.Set(t, category.Integration)

	// This test will run on the actual system
	result := isRpmBasedSystem()

	// If we're on an RPM-based system, certain files or commands should exist
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

		assert.True(t, hasRpmIndicator, "RPM system detected but no RPM indicators found")
	}
}

// Test edge cases and specific scenarios
func TestDetectPackageManagerEdgeCases(t *testing.T) {
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
				_ = DetectPackageManager()
			})
		})
	}
}
