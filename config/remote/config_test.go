package remote

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

func TestVersionMatch(t *testing.T) {
	category.Set(t, category.Unit)

	// ~3.7.1 means: >= 3.7.1 and < 3.8.0
	// ^3.7.1 means: >= 3.7.1 and < 4.0.0

	tests := []struct {
		name         string
		srcVer       string
		trgVer       string
		match        bool
		expectsError bool
	}{
		{
			name:         "exact match",
			srcVer:       "1.1.1",
			trgVer:       "1.1.1",
			match:        true,
			expectsError: false,
		},
		{
			name:         "invalid 1",
			srcVer:       "-",
			trgVer:       "1.1.1",
			match:        false,
			expectsError: true,
		},
		{
			name:         "invalid 2",
			srcVer:       "",
			trgVer:       "1.1.1",
			match:        false,
			expectsError: true,
		},
		{
			name:         "wildcard 1",
			srcVer:       "1.1.1",
			trgVer:       "*",
			match:        true,
			expectsError: false,
		},
		{
			name:         "wildcard 2 invalid",
			srcVer:       "1.1.1",
			trgVer:       "1*",
			match:        false,
			expectsError: true,
		},
		{
			name:         "wildcard 3",
			srcVer:       "1.1.1",
			trgVer:       "1.*",
			match:        true,
			expectsError: false,
		},
		{
			name:         "wildcard 4 invalid",
			srcVer:       "1.1.1",
			trgVer:       "1.1.1.*",
			match:        false,
			expectsError: true,
		},
		{
			name:         "patch 1",
			srcVer:       "1.1.3",
			trgVer:       "~1.1.1",
			match:        true,
			expectsError: false,
		},
		{
			name:         "patch 2",
			srcVer:       "1.2.3",
			trgVer:       "~1.1.1",
			match:        false,
			expectsError: false,
		},
		{
			name:         "fix 1",
			srcVer:       "1.2.3",
			trgVer:       "^1.1.1",
			match:        true,
			expectsError: false,
		},
		{
			name:         "fix 2",
			srcVer:       "2.2.3",
			trgVer:       "^1.1.1",
			match:        false,
			expectsError: false,
		},
		{
			name:         "gt 1",
			srcVer:       "2.2.3",
			trgVer:       ">=1.1.1",
			match:        true,
			expectsError: false,
		},
		{
			name:         "gt 2",
			srcVer:       "2.2.3",
			trgVer:       ">=3.1.1",
			match:        false,
			expectsError: false,
		},
		{
			name:         "lt 1",
			srcVer:       "2.2.3",
			trgVer:       "<=1.1.1",
			match:        false,
			expectsError: false,
		},
		{
			name:         "lt 2",
			srcVer:       "2.2.3",
			trgVer:       "<=3.1.1",
			match:        true,
			expectsError: false,
		},
		{
			name:         "wildcard lt",
			srcVer:       "2.2.3",
			trgVer:       "<=1.*",
			match:        false,
			expectsError: false,
		},
		{
			name:         "wildcard gt",
			srcVer:       "2.2.3",
			trgVer:       ">=1.*",
			match:        true,
			expectsError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rz, err := isVersionMatching(test.srcVer, test.trgVer)
			assert.Equal(t, rz, test.match)
			fmt.Println("match:", rz, ";; err:", err)
			assert.True(t, (!test.expectsError && err == nil) || (test.expectsError && err != nil))
		})
	}
}

func TestValidateField(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name         string
		fieldType    Param
		fieldValue   ParamValue
		fileReadFn   func(string) ([]byte, error)
		expectsError bool
	}{
		{
			name:         "string",
			fieldType:    Param{Type: fieldTypeString},
			fieldValue:   ParamValue{Value: "field-string"},
			expectsError: false,
		},
		{
			name:         "string invalid",
			fieldType:    Param{Type: fieldTypeString},
			fieldValue:   ParamValue{Value: 101},
			expectsError: true,
		},
		{
			name:         "int",
			fieldType:    Param{Type: fieldTypeInt},
			fieldValue:   ParamValue{Value: 1002},
			expectsError: false,
		},
		{
			name:         "int invalid",
			fieldType:    Param{Type: fieldTypeInt},
			fieldValue:   ParamValue{Value: "1002"},
			expectsError: true,
		},
		{
			name:         "bool",
			fieldType:    Param{Type: fieldTypeBool},
			fieldValue:   ParamValue{Value: true},
			expectsError: false,
		},
		{
			name:         "bool invalid",
			fieldType:    Param{Type: fieldTypeBool},
			fieldValue:   ParamValue{Value: "true"},
			expectsError: true,
		},
		{
			name:         "array",
			fieldType:    Param{Type: fieldTypeArray},
			fieldValue:   ParamValue{Value: []any{"one", "two", "three"}},
			expectsError: false,
		},
		{
			name:         "array invalid",
			fieldType:    Param{Type: fieldTypeArray},
			fieldValue:   ParamValue{Value: "not-an-array"},
			expectsError: true,
		},
		{
			name:         "object",
			fieldType:    Param{Type: fieldTypeObject},
			fieldValue:   ParamValue{Value: "{ \"version\": 1}"},
			expectsError: false,
		},
		{
			name:         "object invalid",
			fieldType:    Param{Type: fieldTypeObject},
			fieldValue:   ParamValue{Value: "-/-"},
			expectsError: true,
		},
		{
			name:         "file",
			fieldType:    Param{Type: fieldTypeFile},
			fieldValue:   ParamValue{Value: "include/file1.json"},
			fileReadFn:   func(name string) ([]byte, error) { return nil, nil },
			expectsError: false,
		},
		{
			name:         "file invalid",
			fieldType:    Param{Type: fieldTypeFile},
			fieldValue:   ParamValue{Value: "file1-invalid"},
			fileReadFn:   func(name string) ([]byte, error) { return nil, errors.New("error") },
			expectsError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := validateField(test.fieldType, test.fieldValue, test.fileReadFn)
			if test.expectsError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

type mockReaderWriter struct {
	readCnt, writeCnt  int
	readErr, writeErr  error
	mainJson, hashJson string
}

func (w *mockReaderWriter) writeFile(name string, content []byte, mode os.FileMode) error {
	w.writeCnt++
	return w.writeErr
}
func (w *mockReaderWriter) readFile(name string) ([]byte, error) {
	w.readCnt++
	if strings.Contains(name, "-hash") {
		return []byte(w.hashJson), w.readErr
	}
	return []byte(w.mainJson), w.readErr
}

func TestHandleIncludeFiles(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name        string
		fileName    string
		srcBasePath string
		trgBasePath string
		mainJson    string
		hashJson    string
		expectError bool
		readErr     error
		writeErr    error
	}{
		{
			name:        "sunny day",
			fileName:    "test.json",
			srcBasePath: "s",
			trgBasePath: "t",
			mainJson:    libtelioJsonConfInc1File,
			hashJson:    libtelioJsonConfInc1HashFile,
			expectError: false,
			readErr:     nil,
			writeErr:    nil,
		},
		{
			name:        "invalid file name",
			fileName:    "test.txt",
			srcBasePath: "s",
			trgBasePath: "t",
			mainJson:    "",
			hashJson:    "",
			expectError: true,
			readErr:     nil,
			writeErr:    nil,
		},
		{
			name:        "invalid main json",
			fileName:    "test.json",
			srcBasePath: "s",
			trgBasePath: "t",
			mainJson:    "invalid json",
			hashJson:    "",
			expectError: true,
			readErr:     nil,
			writeErr:    nil,
		},
		{
			name:        "invalid hash json",
			fileName:    "test.json",
			srcBasePath: "s",
			trgBasePath: "t",
			mainJson:    "{}",
			hashJson:    "invalid json",
			expectError: true,
			readErr:     nil,
			writeErr:    nil,
		},
		{
			name:        "hash does not match",
			fileName:    "test.json",
			srcBasePath: "s",
			trgBasePath: "t",
			mainJson:    "{}",
			hashJson:    "{\"hash\":\"aaa\"}",
			expectError: true,
			readErr:     nil,
			writeErr:    nil,
		},
		{
			name:        "file read error",
			fileName:    "test.json",
			srcBasePath: "s",
			trgBasePath: "t",
			mainJson:    "",
			hashJson:    "",
			expectError: true,
			readErr:     errors.New("error"),
			writeErr:    nil,
		},
		{
			name:        "file write error",
			fileName:    "test.json",
			srcBasePath: "s",
			trgBasePath: "t",
			mainJson:    "",
			hashJson:    "",
			expectError: true,
			readErr:     nil,
			writeErr:    errors.New("error"),
		},
		{
			name:        "src same as trg - error",
			fileName:    "test.json",
			srcBasePath: "same",
			trgBasePath: "same",
			mainJson:    "",
			hashJson:    "",
			expectError: true,
			readErr:     nil,
			writeErr:    nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			frw := mockReaderWriter{
				mainJson: test.mainJson,
				hashJson: test.hashJson,
				readErr:  test.readErr,
				writeErr: test.writeErr,
			}
			incf, err := handleIncludeFiles(test.srcBasePath, test.trgBasePath, test.fileName, &frw, &frw)
			if test.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, len(test.mainJson), len(incf))
				assert.Equal(t, 2, frw.readCnt)
				assert.Equal(t, 2, frw.writeCnt)
			}
		})
	}
}

// mockValidator implements the validator interface for testing
type mockValidator struct {
	validateErr error
}

func (v *mockValidator) validate(content []byte) error {
	return v.validateErr
}

// mockFileReader implements the fileReader interface for testing
type mockFileReader struct {
	files map[string][]byte
	err   error
}

func (r *mockFileReader) readFile(name string) ([]byte, error) {
	if r.err != nil {
		return nil, r.err
	}
	// Try exact match first
	if content, ok := r.files[name]; ok {
		return content, nil
	}
	// Try with just the base name for backward compatibility
	baseName := filepath.Base(name)
	if content, ok := r.files[baseName]; ok {
		return content, nil
	}
	// Try removing the directory prefix for paths like "testdir/include/file.json"
	// when the mock has "include/file.json"
	parts := strings.Split(name, string(filepath.Separator))
	for i := 1; i < len(parts); i++ {
		subPath := filepath.Join(parts[i:]...)
		if content, ok := r.files[subPath]; ok {
			return content, nil
		}
	}
	return nil, fmt.Errorf("file not found: %s", name)
}

func TestFeatureLoad(t *testing.T) {
	category.Set(t, category.Unit)

	// Create a valid hash for test data
	validMainJson := `{
		"version": 1,
		"configs": [
			{
				"name": "test_param",
				"value_type": "string",
				"settings": [
					{
						"value": "test_value",
						"app_version": "*",
						"weight": 1
					}
				]
			}
		]
	}`
	validHash := hash([]byte(validMainJson))
	validHashJson := fmt.Sprintf(`{"hash": "%s"}`, validHash)

	// JSON with invalid field type that will trigger LoadErrorFieldValidation
	invalidFieldJson := `{
		"version": 1,
		"configs": [
			{
				"name": "test_param",
				"value_type": "string",
				"settings": [
					{
						"value": 123,
						"app_version": "*",
						"weight": 1
					}
				]
			}
		]
	}`
	invalidFieldHash := hash([]byte(invalidFieldJson))
	invalidFieldHashJson := fmt.Sprintf(`{"hash": "%s"}`, invalidFieldHash)

	tests := []struct {
		name          string
		feature       *Feature
		sourcePath    string
		fileReader    fileReader
		validator     validator
		expectedError error
		errorContains string
		errorKind     LoadErrorKind
	}{
		{
			name: "successful load",
			feature: &Feature{
				name: "test",
			},
			sourcePath: "testdir",
			fileReader: &mockFileReader{
				files: map[string][]byte{
					"test.json":      []byte(validMainJson),
					"test-hash.json": []byte(validHashJson),
				},
			},
			validator:     &mockValidator{},
			expectedError: nil,
		},
		{
			name: "empty feature name",
			feature: &Feature{
				name: "",
			},
			sourcePath:    "testdir",
			fileReader:    &mockFileReader{},
			validator:     &mockValidator{},
			errorContains: "feature name is not set",
			errorKind:     LoadErrorOther,
		},
		{
			name: "invalid source path",
			feature: &Feature{
				name: "test",
			},
			sourcePath:    "/nonexistent/path",
			fileReader:    &mockFileReader{},
			validator:     &mockValidator{},
			errorContains: "config source path is not valid",
			errorKind:     LoadErrorOther,
		},
		{
			name: "hash file not found",
			feature: &Feature{
				name: "test",
			},
			sourcePath: "testdir",
			fileReader: &mockFileReader{
				files: map[string][]byte{},
			},
			validator:     &mockValidator{},
			errorContains: "reading hash file",
			errorKind:     LoadErrorFileNotFound,
		},
		{
			name: "invalid hash json",
			feature: &Feature{
				name: "test",
			},
			sourcePath: "testdir",
			fileReader: &mockFileReader{
				files: map[string][]byte{
					"test-hash.json": []byte("invalid json"),
				},
			},
			validator:     &mockValidator{},
			errorContains: "parsing main hash file",
			errorKind:     LoadErrorMainHashJsonParsing,
		},
		{
			name: "main file not found",
			feature: &Feature{
				name: "test",
			},
			sourcePath: "testdir",
			fileReader: &mockFileReader{
				files: map[string][]byte{
					"test-hash.json": []byte(validHashJson),
				},
			},
			validator:     &mockValidator{},
			errorContains: "reading config file",
			errorKind:     LoadErrorFileNotFound,
		},
		{
			name: "json validation failure",
			feature: &Feature{
				name: "test",
			},
			sourcePath: "testdir",
			fileReader: &mockFileReader{
				files: map[string][]byte{
					"test.json":      []byte(validMainJson),
					"test-hash.json": []byte(validHashJson),
				},
			},
			validator: &mockValidator{
				validateErr: fmt.Errorf("validation failed"),
			},
			errorContains: "validating json",
			errorKind:     LoadErrorMainJsonValidationFailure,
		},
		{
			name: "integrity check failure",
			feature: &Feature{
				name: "test",
			},
			sourcePath: "testdir",
			fileReader: &mockFileReader{
				files: map[string][]byte{
					"test.json":      []byte(validMainJson),
					"test-hash.json": []byte(`{"hash": "wrong_hash"}`),
				},
			},
			validator:     &mockValidator{},
			errorContains: "main file integrity problem",
			errorKind:     LoadErrorIntegrity,
		},
		{
			name: "invalid json structure",
			feature: &Feature{
				name: "test",
			},
			sourcePath: "testdir",
			fileReader: &mockFileReader{
				files: map[string][]byte{
					"test.json":      []byte("invalid json"),
					"test-hash.json": []byte(`{"hash": "some_hash"}`),
				},
			},
			validator:     &mockValidator{},
			errorContains: "invalid character",
			errorKind:     LoadErrorIncludeFile,
		},
		{
			name: "field validation error - triggers LoadErrorFieldValidation",
			feature: &Feature{
				name: "test",
			},
			sourcePath: "testdir",
			fileReader: &mockFileReader{
				files: map[string][]byte{
					"test.json":      []byte(invalidFieldJson),
					"test-hash.json": []byte(invalidFieldHashJson),
				},
			},
			validator:     &mockValidator{},
			errorContains: "loading string value",
			errorKind:     LoadErrorFieldValidation,
		},
	}

	// Create a temporary directory just for the path validation tests
	tempDir, err := os.MkdirTemp("", "test_path_validation_*")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// For path validation tests, use the actual temp directory
			if test.name == "invalid source path" {
				test.sourcePath = "/nonexistent/path"
			} else if test.sourcePath == "testdir" {
				test.sourcePath = tempDir
			}

			err := test.feature.load(test.sourcePath, test.fileReader, test.validator)

			if test.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, test.expectedError, err)
			} else if test.errorContains != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), test.errorContains)

				// Check if it's a LoadError and verify the error type
				var loadErr *LoadError
				if errors.As(err, &loadErr) && test.errorKind != 0 {
					assert.Equal(t, test.errorKind, loadErr.Kind)
				}
			} else {
				assert.NoError(t, err)
				// Verify that params were loaded correctly
				assert.NotNil(t, test.feature.params)
				assert.NotEmpty(t, test.feature.hash)
			}
		})
	}
}

// Test for include files handling
func TestFeatureLoadWithIncludeFiles(t *testing.T) {
	category.Set(t, category.Unit)

	// Create include file content
	includeFileContent := `{
		"setting1": "value1",
		"setting2": 123,
		"setting3": true
	}`
	includeFileHash := hash([]byte(includeFileContent))
	includeFileHashJson := fmt.Sprintf(`{"hash": "%s"}`, includeFileHash)

	// Create main JSON with file type field
	mainJsonWithInclude := `{
		"version": 1,
		"configs": [
			{
				"name": "config_with_file",
				"value_type": "file",
				"settings": [
					{
						"value": "include/settings.json",
						"app_version": "*",
						"weight": 1
					}
				]
			}
		]
	}`
	mainHash := hash([]byte(mainJsonWithInclude + includeFileContent))
	mainHashJson := fmt.Sprintf(`{"hash": "%s"}`, mainHash)

	// Create main JSON with multiple include files
	mainJsonMultipleIncludes := `{
		"version": 1,
		"configs": [
			{
				"name": "config1",
				"value_type": "file",
				"settings": [
					{
						"value": "include/settings1.json",
						"app_version": "*",
						"weight": 1
					}
				]
			},
			{
				"name": "config2",
				"value_type": "file",
				"settings": [
					{
						"value": "include/settings2.json",
						"app_version": "*",
						"weight": 1
					}
				]
			}
		]
	}`
	multiHash := hash([]byte(mainJsonMultipleIncludes + includeFileContent + includeFileContent))
	multiHashJson := fmt.Sprintf(`{"hash": "%s"}`, multiHash)

	// Create temporary directory for path validation
	tempDir, err := os.MkdirTemp("", "test_include_files_*")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name          string
		feature       *Feature
		fileReader    fileReader
		validator     validator
		errorContains string
		errorKind     LoadErrorKind
		expectSuccess bool
	}{
		{
			name: "successful load with include file",
			feature: &Feature{
				name: "test",
			},
			fileReader: &mockFileReader{
				files: map[string][]byte{
					"test.json":                        []byte(mainJsonWithInclude),
					"test-hash.json":                   []byte(mainHashJson),
					"include/settings.json":            []byte(includeFileContent),
					"include/settings-hash.json":       []byte(includeFileHashJson),
				},
			},
			validator:     &mockValidator{},
			expectSuccess: true,
		},
		{
			name: "include file not found",
			feature: &Feature{
				name: "test",
			},
			fileReader: &mockFileReader{
				files: map[string][]byte{
					"test.json":      []byte(mainJsonWithInclude),
					"test-hash.json": []byte(mainHashJson),
					// Missing include file
				},
			},
			validator:     &mockValidator{},
			errorContains: "loading include file",
			errorKind:     LoadErrorIncludeFile,
		},
		{
			name: "include file invalid json",
			feature: &Feature{
				name: "test",
			},
			fileReader: &mockFileReader{
				files: map[string][]byte{
					"test.json":                  []byte(mainJsonWithInclude),
					"test-hash.json":             []byte(mainHashJson),
					"include/settings.json":      []byte("invalid json"),
					"include/settings-hash.json": []byte(includeFileHashJson),
				},
			},
			validator:     &mockValidator{},
			errorContains: "parsing include file",
			errorKind:     LoadErrorIncludeFile,
		},
		{
			name: "multiple include files success",
			feature: &Feature{
				name: "test",
			},
			fileReader: &mockFileReader{
				files: map[string][]byte{
					"test.json":                   []byte(mainJsonMultipleIncludes),
					"test-hash.json":              []byte(multiHashJson),
					"include/settings1.json":      []byte(includeFileContent),
					"include/settings1-hash.json": []byte(includeFileHashJson),
					"include/settings2.json":      []byte(includeFileContent),
					"include/settings2-hash.json": []byte(includeFileHashJson),
				},
			},
			validator:     &mockValidator{},
			expectSuccess: true,
		},
		{
			name: "include file with wrong value type",
			feature: &Feature{
				name: "test",
			},
			fileReader: &mockFileReader{
				files: map[string][]byte{
					"test.json": []byte(`{
						"version": 1,
						"configs": [{
							"name": "config",
							"value_type": "file",
							"settings": [{
								"value": 123,
								"app_version": "*",
								"weight": 1
							}]
						}]
					}`),
					"test-hash.json": []byte(`{"hash": "somehash"}`),
				},
			},
			validator:     &mockValidator{},
			errorContains: "loading include file name as string value",
			errorKind:     LoadErrorIncludeFile,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.feature.load(tempDir, test.fileReader, test.validator)

			if test.expectSuccess {
				assert.NoError(t, err)
				assert.NotNil(t, test.feature.params)
				assert.NotEmpty(t, test.feature.hash)
				
				// Verify file type parameters were loaded
				for _, param := range test.feature.params {
					if param.Type == fieldTypeFile {
						assert.NotEmpty(t, param.Settings)
						for _, setting := range param.Settings {
							// Check that include file content was loaded
							assert.NotEmpty(t, setting.incValue)
						}
					}
				}
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), test.errorContains)

				var loadErr *LoadError
				if errors.As(err, &loadErr) && test.errorKind != 0 {
					assert.Equal(t, test.errorKind, loadErr.Kind)
				}
			}
		})
	}
}

// Test handleIncludeFiles edge cases
func TestHandleIncludeFilesEdgeCases(t *testing.T) {
	category.Set(t, category.Unit)

	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "test_handle_include_*")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name          string
		feature       *Feature
		mainJson      string
		fileReader    fileReader
		validator     validator
		errorContains string
		errorKind     LoadErrorKind
	}{
		{
			name: "include file without .json extension",
			feature: &Feature{
				name: "test",
			},
			mainJson: `{
				"version": 1,
				"configs": [{
					"name": "config",
					"value_type": "file",
					"settings": [{
						"value": "include/settings.txt",
						"app_version": "*",
						"weight": 1
					}]
				}]
			}`,
			fileReader: &mockFileReader{
				files: map[string][]byte{
					"test.json":      []byte("dummy"),
					"test-hash.json": []byte(`{"hash": "somehash"}`),
				},
			},
			validator:     &mockValidator{},
			errorContains: "only json files are allowed to include",
			errorKind:     LoadErrorIncludeFile,
		},
		{
			name: "include file hash not found",
			feature: &Feature{
				name: "test",
			},
			mainJson: `{
				"version": 1,
				"configs": [{
					"name": "config",
					"value_type": "file",
					"settings": [{
						"value": "include/settings.json",
						"app_version": "*",
						"weight": 1
					}]
				}]
			}`,
			fileReader: &mockFileReader{
				files: map[string][]byte{
					"test.json":             []byte("dummy"),
					"test-hash.json":        []byte(`{"hash": "somehash"}`),
					"include/settings.json": []byte(`{"key": "value"}`),
					// Missing include/settings-hash.json
				},
			},
			validator:     &mockValidator{},
			errorContains: "handling include file hash",
			errorKind:     LoadErrorIncludeFile,
		},
		{
			name: "include file hash invalid json",
			feature: &Feature{
				name: "test",
			},
			mainJson: `{
				"version": 1,
				"configs": [{
					"name": "config",
					"value_type": "file",
					"settings": [{
						"value": "include/settings.json",
						"app_version": "*",
						"weight": 1
					}]
				}]
			}`,
			fileReader: &mockFileReader{
				files: map[string][]byte{
					"test.json":                  []byte("dummy"),
					"test-hash.json":             []byte(`{"hash": "somehash"}`),
					"include/settings.json":      []byte(`{"key": "value"}`),
					"include/settings-hash.json": []byte("invalid json"),
				},
			},
			validator:     &mockValidator{},
			errorContains: "parsing include file hash",
			errorKind:     LoadErrorIncludeFile,
		},
		{
			name: "include file integrity check failure",
			feature: &Feature{
				name: "test",
			},
			mainJson: `{
				"version": 1,
				"configs": [{
					"name": "config",
					"value_type": "file",
					"settings": [{
						"value": "include/settings.json",
						"app_version": "*",
						"weight": 1
					}]
				}]
			}`,
			fileReader: &mockFileReader{
				files: map[string][]byte{
					"test.json":                  []byte("dummy"),
					"test-hash.json":             []byte(`{"hash": "somehash"}`),
					"include/settings.json":      []byte(`{"key": "value"}`),
					"include/settings-hash.json": []byte(`{"hash": "wronghash"}`),
				},
			},
			validator:     &mockValidator{},
			errorContains: "include file integrity problem",
			errorKind:     LoadErrorIncludeFile,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Update file reader with the test's main JSON
			fr := test.fileReader.(*mockFileReader)
			fr.files["test.json"] = []byte(test.mainJson)
			
			// Calculate proper hash for the main file
			mainHash := hash([]byte(test.mainJson))
			fr.files["test-hash.json"] = []byte(fmt.Sprintf(`{"hash": "%s"}`, mainHash))

			err := test.feature.load(tempDir, test.fileReader, test.validator)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), test.errorContains)

			var loadErr *LoadError
			if errors.As(err, &loadErr) {
				assert.Equal(t, test.errorKind, loadErr.Kind)
			}
		})
	}
}

// Test specifically for LoadErrorFieldValidation with different field types
func TestFeatureLoadFieldValidationErrors(t *testing.T) {
	category.Set(t, category.Unit)

	// Create temporary test directory for path validation only
	tempDir, err := os.MkdirTemp("", "test_field_validation_*")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name       string
		configJson string
	}{
		{
			name: "invalid string field",
			configJson: `{
				"version": 1,
				"configs": [{
					"name": "test",
					"value_type": "string",
					"settings": [{
						"value": 123,
						"app_version": "*",
						"weight": 1
					}]
				}]
			}`,
		},
		{
			name: "invalid int field",
			configJson: `{
				"version": 1,
				"configs": [{
					"name": "test",
					"value_type": "integer",
					"settings": [{
						"value": "not_an_int",
						"app_version": "*",
						"weight": 1
					}]
				}]
			}`,
		},
		{
			name: "invalid bool field",
			configJson: `{
				"version": 1,
				"configs": [{
					"name": "test",
					"value_type": "boolean",
					"settings": [{
						"value": "not_a_bool",
						"app_version": "*",
						"weight": 1
					}]
				}]
			}`,
		},
		{
			name: "invalid array field",
			configJson: `{
				"version": 1,
				"configs": [{
					"name": "test",
					"value_type": "array",
					"settings": [{
						"value": "not_an_array",
						"app_version": "*",
						"weight": 1
					}]
				}]
			}`,
		},
		{
			name: "invalid object field",
			configJson: `{
				"version": 1,
				"configs": [{
					"name": "test",
					"value_type": "object",
					"settings": [{
						"value": "invalid json object",
						"app_version": "*",
						"weight": 1
					}]
				}]
			}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			feature := &Feature{name: "test"}
			testHash := hash([]byte(test.configJson))
			hashJson := fmt.Sprintf(`{"hash": "%s"}`, testHash)

			fileReader := &mockFileReader{
				files: map[string][]byte{
					"test.json":      []byte(test.configJson),
					"test-hash.json": []byte(hashJson),
				},
			}

			err := feature.load(tempDir, fileReader, &mockValidator{})

			assert.Error(t, err)

			// Verify it's a LoadError with LoadErrorFieldValidation kind
			var loadErr *LoadError
			assert.True(t, errors.As(err, &loadErr))
			assert.Equal(t, LoadErrorFieldValidation, loadErr.Kind)
		})
	}
}
