package remote

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"

	"github.com/xeipuuv/gojsonschema"
)

// Validator generic interface for config validator
type Validator interface {
	ValidateFile(string)
	ValidateString(string)
}

// JsonValidator config from json file validator
type JsonValidator struct {
}

func NewJsonValidator() *JsonValidator {
	return &JsonValidator{}
}

// extract top level version field to determine which schema to use for validation
func extractVersionJsonFile(fileName string) (int, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return 0, fmt.Errorf("open json file: %w", err)
	}
	defer file.Close()

	var data struct {
		Version int `json:"version"`
	}

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&data); err != nil {
		return 0, fmt.Errorf("decode json file: %w", err)
	}

	return data.Version, nil
}
func extractVersionJsonStr(jsonStr []byte) (int, error) {
	var data struct {
		Version int `json:"version"`
	}
	if err := json.Unmarshal(jsonStr, &data); err != nil {
		return 0, fmt.Errorf("parsing json file: %w", err)
	}

	return data.Version, nil
}

func (v *JsonValidator) ValidateFile(name string) error {
	documentLoader := gojsonschema.NewReferenceLoader(fmt.Sprintf("file://%s", name))

	schemaVersion, err := extractVersionJsonFile(name)
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
			return fmt.Errorf(errStr)
		}
	}
	return nil
}

func (v *JsonValidator) ValidateString(jsn []byte) error {
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
			return fmt.Errorf(errStr)
		}
	}
	return nil
}

// embedded json schemas for validation,
// schemas are mapped by version.
type jsonSchemaEmbeded []byte

var versionSchemaMap = map[int]jsonSchemaEmbeded{1: jsonSchemaV1}

//go:embed json/schema_v1.json
var jsonSchemaV1 jsonSchemaEmbeded
