package remote

import (
	"encoding/json"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

func TestValidateSchemaJson(t *testing.T) {
	category.Set(t, category.Unit)

	// expecting embedded schemas: 1 or more
	assert.Less(t, 0, len(versionSchemaMap))

	var js any
	// embedded schemas should be non-empty and valid json
	for _, sch := range versionSchemaMap {
		assert.Less(t, 1, len(sch), "expecting schema to be non empty")
		err := json.Unmarshal(sch, &js)
		assert.NoError(t, err)
	}
}

func TestJsonValidator(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name        string
		json        string
		expectError bool
	}{
		{
			name:        "valid json",
			json:        nordwhisperJsonConfFile,
			expectError: false,
		},
		{
			name:        "invalid json - invalid version",
			json:        nordvpnInvalidVersionJsonConfFile,
			expectError: true,
		},
		{
			name:        "invalid json - missing version",
			json:        nordvpnMissingVersionJsonConfFile,
			expectError: true,
		},
		{
			name:        "invalid json - invalid field value type 1",
			json:        nordvpnInvalidFieldTypeJsonConfFile,
			expectError: true,
		},
		{
			name:        "invalid json - invalid field value type 2",
			json:        nordvpnInvalidFieldType2JsonConfFile,
			expectError: true,
		},
		{
			name:        "invalid json - invalid fields",
			json:        nordvpnInvalidFieldValuesJsonConfFile,
			expectError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateJsonString([]byte(test.json))
			if test.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestExtractJsonVer(t *testing.T) {
	category.Set(t, category.Unit)

	jsnStr := `{"version": 1,"comment": "no comments"}`

	ver, err := extractVersionJsonStr([]byte(jsnStr))

	assert.Equal(t, 1, ver)
	assert.NoError(t, err)
}
