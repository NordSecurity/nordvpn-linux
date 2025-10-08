package daemon

import (
	"fmt"
	"sort"
	"strings"
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

func TestParseRpmVersions_Unit(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name: "single version",
			input: `<?xml version="1.0" encoding="UTF-8"?>
					<filelists xmlns="http://linux.duke.edu/metadata/filelists" packages="1">
					<package arch="x86_64" name="nordvpn" pkgid="test">
					<version epoch="0" rel="1" ver="3.14.0" />
					</package>
					</filelists>`,
			expected: []string{"3.14.0-1"},
		},
		{
			name: "multiple versions",
			input: `<?xml version="1.0" encoding="UTF-8"?>
					<filelists xmlns="http://linux.duke.edu/metadata/filelists" packages="3">
					<package arch="x86_64" name="nordvpn" pkgid="test1">
					<version epoch="0" rel="1" ver="3.14.0" />
					</package>
					<package arch="x86_64" name="nordvpn" pkgid="test2">
					<version epoch="0" rel="2" ver="3.14.1" />
					</package>
					<package arch="x86_64" name="nordvpn" pkgid="test3">
					<version epoch="0" rel="10" ver="3.15.0" />
					</package>
					</filelists>`,
			expected: []string{"3.14.0-1", "3.14.1-2", "3.15.0-10"},
		},
		{
			name: "multiple versions and multiple packages",
			input: `<?xml version="1.0" encoding="UTF-8"?>
					<filelists xmlns="http://linux.duke.edu/metadata/filelists" packages="3">
					<package arch="x86_64" name="nordvpn" pkgid="test1">
					<version epoch="0" rel="1" ver="3.14.0" />
					</package>
					<package arch="x86_64" name="nordvpn" pkgid="test2">
					<version epoch="0" rel="2" ver="3.14.1" />
					</package>
					<package arch="x86_64" name="nordvpn-gui" pkgid="test2">
					<version epoch="0" rel="2" ver="1.0.0" />
					</package>
					<package arch="x86_64" name="nordvpn" pkgid="test3">
					<version epoch="0" rel="10" ver="3.15.0" />
					</package>
					<package arch="x86_64" name="nordvpn-gui" pkgid="test2">
					<version epoch="0" rel="2" ver="2.0.0" />
					</package>
					</filelists>`,
			expected: []string{"3.14.0-1", "3.14.1-2", "3.15.0-10"},
		},
		{
			name: "versions with different formats",
			input: `<package arch="x86_64" name="nordvpn" pkgid="test1"><version epoch="0" rel="1" ver="1.0.0" /></package>
					<package arch="x86_64" name="nordvpn" pkgid="test1"><version epoch="0" rel="99" ver="2.10.5" /></package>
					<package arch="x86_64" name="nordvpn" pkgid="test1"><version epoch="0" rel="123" ver="10.20.30" /></package>`,
			expected: []string{"1.0.0-1", "2.10.5-99", "10.20.30-123"},
		},
		{
			name:     "empty input",
			input:    "",
			expected: []string{},
		},
		{
			name: "no version tags",
			input: `<?xml version="1.0" encoding="UTF-8"?>
					<filelists xmlns="http://linux.duke.edu/metadata/filelists" packages="0">
					</filelists>`,
			expected: []string{},
		},
		{
			name:     "malformed version - missing rel",
			input:    `<package arch="x86_64" name="nordvpn" pkgid="test1"><version epoch="0" ver="3.14.0" /></package>`,
			expected: []string{},
		},
		{
			name:     "malformed version - missing ver",
			input:    `<package arch="x86_64" name="nordvpn" pkgid="test1"><version epoch="0" rel="1" /></package>`,
			expected: []string{},
		},
		{
			name: "mixed valid and invalid versions",
			input: `<package arch="x86_64" name="nordvpn" pkgid="test1"><version epoch="0" rel="1" ver="3.14.0" /></package>
					<package arch="x86_64" name="nordvpn" pkgid="test1"><version epoch="0" rel="2" /></package>
					<package arch="x86_64" name="nordvpn" pkgid="test1"><version epoch="0" ver="3.14.2" /></package>
					<package arch="x86_64" name="nordvpn" pkgid="test1"><version epoch="0" rel="3" ver="3.14.3" /></package>`,
			expected: []string{"3.14.0-1", "3.14.3-3"},
		},
		{
			name: "version with invalid format after parsing",
			input: `<package arch="x86_64" name="nordvpn" pkgid="test1"><version epoch="0" rel="1" ver="3.14.0" /></package>
					<package arch="x86_64" name="nordvpn" pkgid="test1"><version epoch="0" rel="2" ver="3.14" /></package>
					<package arch="x86_64" name="nordvpn" pkgid="test1"><version epoch="0" rel="3" ver="3.14.0.1" /></package>
					<package arch="x86_64" name="nordvpn" pkgid="test1"><version epoch="0" rel="4" ver="not-a-version" /></package>`,
			expected: []string{"3.14.0-1"},
		},
		{
			name: "version attributes in different order",
			input: `<package arch="x86_64" name="nordvpn" pkgid="test1"><version ver="3.14.0" rel="1" epoch="0" /></package>
					<package arch="x86_64" name="nordvpn" pkgid="test1"><version rel="2" epoch="0" ver="3.14.1" /></package>`,
			expected: []string{"3.14.0-1", "3.14.1-2"},
		},
		{
			name:     "version with spaces",
			input:    `<package arch="x86_64" name="nordvpn" pkgid="test1"><version epoch="0" rel="1" ver=" 3.14.0 " /></package>`,
			expected: []string{},
		},
		{
			name:     "large release numbers",
			input:    `<package arch="x86_64" name="nordvpn" pkgid="test1"><version epoch="0" rel="999" ver="1.0.0" /></package>`,
			expected: []string{"1.0.0-999"},
		},
		{
			name:     "version tag with extra attributes",
			input:    `<package arch="x86_64" name="nordvpn" pkgid="test1"><version epoch="0" rel="1" ver="3.14.0" extra="ignored" /></package>`,
			expected: []string{"3.14.0-1"},
		},
		{
			name: "case sensitivity test",
			input: `<package arch="x86_64" name="nordvpn" pkgid="test1"><version epoch="0" REL="1" VER="3.14.0" /></package>
					<package arch="x86_64" name="nordvpn" pkgid="test1"><Version epoch="0" rel="2" ver="3.14.1" /></package>`,
			expected: []string{"3.14.0-1", "3.14.1-2"}, // Both match now (case-insensitive)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseRpmVersions([]byte(tt.input))
			sort.Strings(result)
			sort.Strings(tt.expected)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseRpmVersions_EdgeCases(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name      string
		input     []byte
		assertion func(t *testing.T, result []string)
	}{
		{
			name:  "nil input",
			input: nil,
			assertion: func(t *testing.T, result []string) {
				assert.Empty(t, result)
			},
		},
		{
			name: "very large input",
			input: func() []byte {
				// Create a large input with many versions
				var builder strings.Builder
				builder.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
				for i := 0; i < 100; i++ {
					builder.WriteString(fmt.Sprintf(`<package arch="x86_64" name="nordvpn" pkgid="test1"><version epoch="0" rel="%d" ver="1.0.%d" /></package>`, i%100, i%100))
				}
				return []byte(builder.String())
			}(),
			assertion: func(t *testing.T, result []string) {
				assert.Greater(t, len(result), 0)
				assert.LessOrEqual(t, len(result), 100)
			},
		},
		{
			name: "special characters in version",
			input: []byte(`<package arch="x86_64" name="nordvpn" pkgid="test1"><version epoch="0" rel="1" ver="3.14.0-beta" /></package>
					<package arch="x86_64" name="nordvpn" pkgid="test1"><version epoch="0" rel="2" ver="3.14.0+build" /></package>
					<package arch="x86_64" name="nordvpn-gui" pkgid="test1"><version epoch="0" rel="2" ver="3.14.0+build" /></package>
					<package arch="x86_64" name="nordvpn" pkgid="test1"><version epoch="0" rel="3" ver="3.14.0~rc1" /></package>`),
			assertion: func(t *testing.T, result []string) {
				assert.Empty(t, result) // None should match the validation pattern
			},
		},
		{
			name: "unicode in input",
			input: []byte(`<package arch="x86_64" name="nordvpn" pkgid="test1"><version epoch="0" rel="1" ver="3.14.0" /></package>
					<package arch="x86_64" name="nordvpn" pkgid="test1"><version epoch="0" rel="2" ver="3.14.1" /> 中文</package>
					<package arch="x86_64" name="nordvpn-gui" pkgid="test1"><version epoch="0" rel="2" ver="3.14.1" /> 中文</package>
					<package arch="x86_64" name="nordvpn" pkgid="test1"><version epoch="0" rel="3" ver="3.14.2" /></package>`),
			assertion: func(t *testing.T, result []string) {
				assert.Len(t, result, 3)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseRpmVersions(tt.input)
			tt.assertion(t, result)
		})
	}
}

func TestValidateVersionStrings(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "valid versions",
			input:    []string{"1.0.0-1", "2.10.5-99", "10.20.30-123"},
			expected: []string{"1.0.0-1", "2.10.5-99", "10.20.30-123"},
		},
		{
			name:     "invalid versions filtered out",
			input:    []string{"1.0.0-1", "1.0", "1.0.0", "1.0.0-", "-1", "1.0.0-1-2"},
			expected: []string{"1.0.0-1"},
		},
		{
			name:     "empty input",
			input:    []string{},
			expected: []string{},
		},
		{
			name:     "all invalid",
			input:    []string{"invalid", "1.2", "1.2.3.4", "abc-def"},
			expected: []string{},
		},
		{
			name:     "versions with leading zeros",
			input:    []string{"01.02.03-04", "1.02.3-4", "1.2.03-4"},
			expected: []string{"01.02.03-04", "1.02.3-4", "1.2.03-4"}, // The regex allows leading zeros
		},
		{
			name:     "boundary values",
			input:    []string{"0.0.0-0", "999.999.999-999", "1.1.1-1000"},
			expected: []string{"0.0.0-0", "999.999.999-999"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validateVersionStrings(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
