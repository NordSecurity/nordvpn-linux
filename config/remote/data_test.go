package remote

import (
	"fmt"
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

func TestFeatureOrderFixed(t *testing.T) {
	category.Set(t, category.Unit)

	const featureCount = 5
	const featureNamePattern = "feature-%d"

	featureMap := NewFeatureMap()
	for i := 0; i < featureCount; i++ {
		featureMap.add(fmt.Sprintf(featureNamePattern, i))
	}
	assert.Equal(t, featureCount, len(featureMap.keys()))
	assert.Equal(t, featureCount, len(featureMap.features()))

	for i, f := range featureMap.keys() {
		assert.Equal(t, f, fmt.Sprintf(featureNamePattern, i))
	}
}
