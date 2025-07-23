package remote

import (
	"errors"
	"fmt"
	"os"
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
