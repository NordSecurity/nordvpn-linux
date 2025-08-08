package remote

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

const (
	tmpExt = ".bu"
)

type jsonHash struct {
	Hash string `json:"hash"`
}

// DownloadError holds information about a specific download failure that is then used while reporting analytics about RemoteConfig.
type DownloadError struct {
	Kind  DownloadErrorKind
	Cause error
}

func (e *DownloadError) Error() string {
	return fmt.Sprintf("%s: %v", e.Kind, e.Cause)
}

func (e *DownloadError) Unwrap() error {
	return e.Cause
}

func NewDownloadError(kind DownloadErrorKind, err error) *DownloadError {
	return &DownloadError{
		Kind:  kind,
		Cause: err,
	}
}

func hash(content []byte) string {
	hash := sha256.Sum256(content)
	return hex.EncodeToString(hash[:])
}

func isHashEqual(targetHash string, data ...[]byte) bool {
	return targetHash == hash(bytes.Join(data, nil))
}

func isVersionMatching(appVer, verConstraint string) (bool, error) {
	// ~3.7.1 means: >= 3.7.1 and < 3.8.0
	// ^3.7.1 means: >= 3.7.1 and < 4.0.0
	v, err := semver.NewVersion(appVer)
	if err != nil {
		return false, fmt.Errorf("invalid app version: %w", err)
	}
	constraint, err := semver.NewConstraint(verConstraint)
	if err != nil {
		return false, fmt.Errorf("invalid version constraint: %w", err)
	}
	return constraint.Check(v), nil
}

type fileWriter interface {
	writeFile(string, []byte, os.FileMode) error
}

type fileReader interface {
	readFile(string) ([]byte, error)
}

type validator interface {
	validate([]byte) error
}

// download main json file and check if include files should be downloaded,
// return `true` if remote config was really downloaded.
func (f *Feature) download(cdn fileReader, fw fileWriter, jv validator, cdnBasePath, targetPath string) (success bool, err error) {
	// config file consists of:
	// - main json file e.g. nordvpn.json;
	// - sibling file with hash e.g. nordvpn-hash.json;
	// - 0..n include files e.g. libtelio-3.19.0.json;
	// - include file hash file e.g. libtelio-3.19.0-hash.json;

	defer func() {
		if err != nil {
			internal.CleanupTmpFiles(targetPath, tmpExt)
		}
	}()

	if f.name == "" {
		return false, NewDownloadError(DownloadErrorOther, fmt.Errorf("feature name is not set"))
	}

	if err = internal.EnsureDirFull(targetPath); err != nil {
		return false, NewDownloadError(DownloadErrorLocalFS, fmt.Errorf("setting-up target dir: %w", err))
	}

	mainJsonHashStr, err := cdn.readFile(f.HashFilePath(cdnBasePath))
	if err != nil {
		return false, NewDownloadError(DownloadErrorRemoteHashNotFound, fmt.Errorf("downloading main hash file: %w", err))
	}

	var mainJsonHash jsonHash
	if err = json.Unmarshal(mainJsonHashStr, &mainJsonHash); err != nil {
		return false, NewDownloadError(DownloadErrorHashParsing, fmt.Errorf("parsing main hash file: %w", err))
	}

	if f.hash == mainJsonHash.Hash {
		return false, nil
	}

	// main hash covers the include files as well
	mainJsonStr, err := cdn.readFile(f.FilePath(cdnBasePath))
	if err != nil {
		return false, NewDownloadError(DownloadErrorRemoteFileNotFound, fmt.Errorf("downloading main file: %w", err))
	}

	// validate json against predefined schema
	if err = jv.validate(mainJsonStr); err != nil {
		return false, NewDownloadError(DownloadErrorParsing, fmt.Errorf("validating main: %w", err))
	}

	incFiles, err := walkIncludeFiles(mainJsonStr, cdnBasePath, targetPath, cdn, fw)
	if err != nil {
		return false, NewDownloadError(DownloadErrorIncludeFile, fmt.Errorf("downloading include files: %w", err))
	}

	// verify content integrity
	// if main json has include files - hash should cover whole content
	if !isHashEqual(mainJsonHash.Hash, mainJsonStr, incFiles) {
		return false, NewDownloadError(DownloadErrorHashIntegrity, fmt.Errorf("main file integrity problem"))
	}

	// write main json to file
	localFileName := f.FilePath(targetPath) + tmpExt
	if err = fw.writeFile(localFileName, mainJsonStr, internal.PermUserRW); err != nil {
		return false, NewDownloadError(DownloadErrorWriteJson, fmt.Errorf("writing main file: %w", err))
	}
	// write main hash to file
	localFileName = f.HashFilePath(targetPath) + tmpExt
	if err = fw.writeFile(localFileName, mainJsonHashStr, internal.PermUserRW); err != nil {
		return false, NewDownloadError(DownloadErrorWriteHash, fmt.Errorf("writing main hash file: %w", err))
	}

	// while processing, save files with special extension '*.bu'
	// if download or handling would fail in the middle - previous files are left intact,
	// also need to cleanup tmp files (see above)
	if err = internal.RenameTmpFiles(targetPath, tmpExt); err != nil {
		return false, NewDownloadError(DownloadErrorFileRename, fmt.Errorf("writing/renaming files: %w", err))
	}

	return true, nil
}

// load feature config from JSON file
func (f *Feature) load(sourcePath string, fr fileReader, jv validator) error {
	if f.name == "" {
		return fmt.Errorf("feature name is not set")
	}

	validDir, err := internal.IsValidExistingDir(sourcePath)
	if err != nil {
		return fmt.Errorf("accessing source path: %w", err)
	}
	if !validDir {
		return fmt.Errorf("config source path is not valid")
	}

	mainJsonHashStr, err := fr.readFile(f.HashFilePath(sourcePath))
	if err != nil {
		return fmt.Errorf("reading hash file: %w", err)
	}
	var mainJsonHash jsonHash
	if err = json.Unmarshal(mainJsonHashStr, &mainJsonHash); err != nil {
		return fmt.Errorf("parsing main hash file: %w", err)
	}

	mainJsonFileName := f.FilePath(sourcePath)
	if err := internal.IsFileTooBig(mainJsonFileName); err != nil {
		return fmt.Errorf("reading main file: %w", err)
	}

	mainJsonStr, err := fr.readFile(mainJsonFileName)
	if err != nil {
		return fmt.Errorf("reading config file: %w", err)
	}
	// validate json by predefined schema
	if err := jv.validate(mainJsonStr); err != nil {
		return fmt.Errorf("validating json: %w", err)
	}

	incFiles, err := walkIncludeFiles(mainJsonStr, sourcePath, "", fr, noopWriter{})
	if err != nil {
		return fmt.Errorf("loading include files: %w", err)
	}

	// verify content integrity
	// if main json has include files - hash should cover whole content
	if !isHashEqual(mainJsonHash.Hash, mainJsonStr, incFiles) {
		return fmt.Errorf("main file integrity problem")
	}

	// load json into structures
	var temp Feature
	if err := json.Unmarshal(mainJsonStr, &temp); err != nil {
		return err
	}

	params := make(map[string]*Param)
	for _, cfgItem := range temp.Configs {
		params[cfgItem.Name] = &Param{Name: cfgItem.Name, Type: cfgItem.Type, Settings: []ParamValue{}}
		for _, param := range cfgItem.Settings {
			// function for special field type `file` to read and validate
			fileReadFunc := func(name string) ([]byte, error) {
				jsnIncFile := filepath.Join(sourcePath, name)
				jsn, err := fr.readFile(jsnIncFile)
				if err != nil {
					return nil, fmt.Errorf("loading include file [%s]: %w", jsnIncFile, err)
				}
				var temp any
				// do basic json validation for include file
				if err := json.Unmarshal(jsn, &temp); err != nil {
					return nil, fmt.Errorf("loading json from include file [%s]: %w", jsnIncFile, err)
				}
				return jsn, nil
			}
			// validate field
			incVal, err := validateField(cfgItem, param, fileReadFunc)
			if err != nil {
				return err
			}
			// store valid values in the map
			params[cfgItem.Name].Settings = append(params[cfgItem.Name].Settings,
				ParamValue{Value: param.Value, incValue: string(incVal), AppVersion: param.AppVersion, Weight: param.Weight})
		}
	}
	// set new params and hash
	f.params = params
	f.hash = hash(bytes.Join([][]byte{mainJsonStr, incFiles}, nil))
	return nil
}

func validateField(param Param, paramVal ParamValue, fileReader func(name string) ([]byte, error)) (fileVal []byte, e error) {
	switch param.Type {
	case fieldTypeString:
		if _, err := paramVal.AsString(); err != nil {
			return nil, fmt.Errorf("loading string value [%s]: %w", paramVal.Value, err)
		}
	case fieldTypeInt, fieldTypeNumber:
		if _, err := paramVal.AsInt(); err != nil {
			return nil, fmt.Errorf("loading int value [%s]: %w", paramVal.Value, err)
		}
	case fieldTypeBool:
		if _, err := paramVal.AsBool(); err != nil {
			return nil, fmt.Errorf("loading bool value [%s]: %w", paramVal.Value, err)
		}
	case fieldTypeArray:
		if _, err := paramVal.AsStringArray(); err != nil {
			return nil, fmt.Errorf("loading string array value [%s]: %w", paramVal.Value, err)
		}
	case fieldTypeObject:
		// load as string and validate json
		val, err := paramVal.AsString()
		if err != nil {
			return nil, fmt.Errorf("loading json object as string value [%s]: %w", paramVal.Value, err)
		}
		var temp any
		if err := json.Unmarshal([]byte(val), &temp); err != nil {
			return nil, fmt.Errorf("loading json object [%s]: %w", val, err)
		}
	case fieldTypeFile:
		// primary field value is an include file name
		val, err := paramVal.AsString()
		if err != nil {
			return nil, fmt.Errorf("loading include file name as string value [%s]: %w", paramVal.Value, err)
		}
		fileVal, err = fileReader(val)
		if err != nil {
			return nil, fmt.Errorf("loading include file name as string value [%s]: %w", paramVal.Value, err)
		}
	}
	return fileVal, nil
}

// walkIncludeFiles iterate through include files and handle each of them
func walkIncludeFiles(mainJason []byte, srcBasePath, trgBasePath string, fr fileReader, fw fileWriter) ([]byte, error) {
	var temp Feature
	if err := json.Unmarshal(mainJason, &temp); err != nil {
		return nil, err
	}
	incFilesJson := []byte{}
	for _, cfgItem := range temp.Configs {
		for _, param := range cfgItem.Settings {
			switch cfgItem.Type {
			case fieldTypeFile:
				// primary field value is an include file name
				incFileName, err := param.AsString()
				if err != nil {
					return nil, fmt.Errorf("loading include file name as string value [%s]: %w", param.Value, err)
				}
				incFile, err := handleIncludeFiles(srcBasePath, trgBasePath, incFileName, fr, fw)
				if err != nil {
					return nil, fmt.Errorf("downloading include file [%s]: %w", incFileName, err)
				}
				incFilesJson = append(incFilesJson, incFile...)
			}
		}
	}
	return incFilesJson, nil
}

func handleIncludeFiles(srcBasePath, trgBasePath, incFileName string, fr fileReader, fw fileWriter) ([]byte, error) {
	if !strings.Contains(incFileName, ".json") {
		return nil, fmt.Errorf("only json files are allowed to include: %s", incFileName)
	}
	if srcBasePath == trgBasePath {
		return nil, fmt.Errorf("source and target base paths cannot be equal")
	}
	// get include file
	incJsonStr, err := fr.readFile(filepath.Join(srcBasePath, incFileName))
	if err != nil {
		return nil, fmt.Errorf("downloading include file: %w", err)
	}
	// do basic json validation
	var tmpJson any
	if err = json.Unmarshal(incJsonStr, &tmpJson); err != nil {
		return nil, fmt.Errorf("parsing include file: %w", err)
	}
	// verify include file content integrity
	incHashFileName := strings.ReplaceAll(incFileName, ".json", "-hash.json")
	incHashStr, err := fr.readFile(filepath.Join(srcBasePath, incHashFileName))
	if err != nil {
		return nil, fmt.Errorf("downloading include file hash: %w", err)
	}
	var incJsonHash jsonHash
	if err = json.Unmarshal(incHashStr, &incJsonHash); err != nil {
		return nil, fmt.Errorf("parsing include file hash: %w", err)
	}
	if !isHashEqual(incJsonHash.Hash, incJsonStr) {
		return nil, fmt.Errorf("include file integrity problem, expected hash[%s] got hash[%s]", incJsonHash.Hash, hash(incJsonStr))
	}
	// perform write only if target path is specified
	if trgBasePath != "" {
		// write an include json to file
		incFileTargetPath := filepath.Join(trgBasePath, incFileName) + tmpExt
		if err = fw.writeFile(incFileTargetPath, incJsonStr, internal.PermUserRW); err != nil {
			return nil, fmt.Errorf("writing include file: %w", err)
		}
		// write an include hash to file
		incFileHashTargetPath := filepath.Join(trgBasePath, incHashFileName) + tmpExt
		if err = fw.writeFile(incFileHashTargetPath, incHashStr, internal.PermUserRW); err != nil {
			return nil, fmt.Errorf("writing include hash file: %w", err)
		}
	}
	return incJsonStr, nil
} //func()
