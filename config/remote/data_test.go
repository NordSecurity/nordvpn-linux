package remote

import (
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"gotest.tools/v3/assert"
)

func TestFeaturePaths(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name             string
		featureName      string
		basePath         string
		expectedPath     string
		expectedHashPath string
	}{
		{
			name:             "valid name",
			featureName:      "valid",
			basePath:         "",
			expectedPath:     "valid.json",
			expectedHashPath: "valid-hash.json",
		},
		{
			name:             "valid name 2",
			featureName:      "valid",
			basePath:         "base/path",
			expectedPath:     "base/path/valid.json",
			expectedHashPath: "base/path/valid-hash.json",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			feature := Feature{name: test.featureName}
			assert.Equal(t, test.expectedPath, feature.FilePath(test.basePath))
			assert.Equal(t, test.expectedHashPath, feature.HashFilePath(test.basePath))
		})
	}
}
