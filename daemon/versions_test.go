package daemon

import (
	"sort"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/test/category"

	"github.com/coreos/go-semver/semver"
	"github.com/stretchr/testify/assert"
)

func TestGetLatestVersion(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		input    []semver.Version
		expected semver.Version
	}{
		{[]semver.Version{*semver.New("1.2.3-0"),
			*semver.New("3.0.0-3"),
			*semver.New("0.1.3-1"),
			*semver.New("10.12.19-39")},
			*semver.New("10.12.19-39")},
		{[]semver.Version{*semver.New("3.0.0-0"),
			*semver.New("3.0.0-1")},
			*semver.New("3.0.0-1")},
	}

	for _, item := range tests {
		got := GetLatestVersion(item.input)
		assert.Equal(t, item.expected, got)
	}
}

func TestParseDebianVersions(t *testing.T) {
	category.Set(t, category.File)

	expected := []string{"2.2.0-0", "2.1.0-1", "3.0.0-2", "2.1.0-2", "2.0.0-0", "2.2.0-3", "3.0.0-4",
		"2.1.0-5", "2.1.0-4", "2.1.0-0", "2.2.0-2", "3.0.0-1", "2.2.0-1", "3.0.0-3", "2.1.0-3"}
	data, err := internal.FileRead(TestdataPath + TestVersionDeb)
	assert.NoError(t, err)
	parsed := ParseDebianVersions(data)
	sort.Strings(expected)
	sort.Strings(parsed)
	assert.EqualValues(t, expected, parsed)
}

func TestParseRpmVersions(t *testing.T) {
	category.Set(t, category.File)

	expected := []string{"2.2.0-2", "3.0.0-4", "2.1.0-5", "2.1.0-1", "2.1.0-3", "2.1.0-2", "2.1.0-4",
		"2.2.0-3", "2.1.0-0", "3.0.0-3", "2.2.0-0", "2.2.0-1", "2.0.0-1", "3.0.0-1", "3.0.0-2"}
	data, err := internal.FileRead(TestdataPath + TestVersionRpm)
	if err != nil {
		t.Fatalf("ParseRpmVersions failed. Got error reading test file: %v.\n", err)
	}
	parsed := ParseRpmVersions(data)
	sort.Strings(expected)
	sort.Strings(parsed)
	assert.EqualValues(t, expected, parsed)
}
