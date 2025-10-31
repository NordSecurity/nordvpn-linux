package remote

import (
	_ "embed"
	"encoding/json"
	"fmt"

	"github.com/xeipuuv/gojsonschema"
)

func extractVersionJsonStr(jsonStr []byte) (int, error) {
	var data struct {
		Version int `json:"version"`
	}
	if err := json.Unmarshal(jsonStr, &data); err != nil {
		return 0, fmt.Errorf("parsing json: %w", err)
	}

	return data.Version, nil
}

func validateJsonString(jsn []byte) error {
	documentLoader := gojsonschema.NewBytesLoader(jsn)

	schemaVersion, err := extractVersionJsonStr(jsn)
	if err != nil {
		return fmt.Errorf("extract json schema version: %w", err)
	}
	schema, found := versionSchemaMap[schemaVersion]
	if !found {
		return fmt.Errorf("json schema not found by version: %d", schemaVersion)
	}
	schemaLoader := gojsonschema.NewStringLoader(string(schema))

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return err
	}
	if !result.Valid() {
		var errStr string
		for idx, desc := range result.Errors() {
			if errStr == "" {
				errStr = fmt.Sprintf("[%d] %s", idx, desc)
			} else {
				errStr = fmt.Sprintf("%s;; [%d] %s", errStr, idx, desc)
			}
		}
		return fmt.Errorf("%s", errStr)
	}
	return nil
}

// embedded json schemas for validation,
// schemas are mapped by version.
type jsonSchemaEmbeded []byte

var versionSchemaMap = map[int]jsonSchemaEmbeded{1: jsonSchemaV1}

//go:embed json/schema_v1.json
var jsonSchemaV1 jsonSchemaEmbeded
