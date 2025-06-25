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

	err := NewJsonValidator().ValidateString([]byte(nordwhisperJsonConfFile))
	assert.NoError(t, err)
}

func TestExtractJsonVer(t *testing.T) {
	category.Set(t, category.Unit)

	jsnStr := `{"version": 1,"comment": "no comments"}`

	ver, err := extractVersionJsonStr([]byte(jsnStr))

	assert.Equal(t, 1, ver)
	assert.NoError(t, err)
}
