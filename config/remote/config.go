package remote

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

type jsonHash struct {
	Hash string `json:"hash"`
}

func hash(content []byte) string {
	hash := md5.Sum(content)
	return hex.EncodeToString(hash[:])
}

func isVersionMatching(appVer, requiredVer string) (bool, error) {
	// "" "*" "1.*" "1.1.*"
	// ~3.7.1 means: >= 3.7.1 and < 3.8.0
	// ^3.7.1 means: >= 3.7.1 and < 4.0.0
	v, err := semver.NewVersion(appVer)
	if err != nil {
		return false, fmt.Errorf("invalid app version: %w", err)
	}
	// TODO/FIXME: on config loading validate app version constraints!!!
	constraint, err := semver.NewConstraint(requiredVer)
	if err != nil {
		return false, fmt.Errorf("invalid version constraint: %w", err)
	}
	return constraint.Check(v), nil
}

// download main json file and check if include files should be downloaded
func (f *Feature) download(cdn CDN, cdnBasePath, targetPath string) (err error) {
	// config file consists of:
	// - main json file e.g. nordvpn.json;
	// - sibling file with hash e.g. nordvpn-hash.json;
	// - 0..n include files e.g. libtelio-3.19.0.json;
	// - include file hash file e.g. libtelio-3.19.0-hash.json;

	tmpExt := ".bu" //TODO: move to constants

	defer func() {
		if err != nil {
			cleanupFiles(targetPath, tmpExt)
		}
	}()

	if f.Name == "" {
		return fmt.Errorf("feature name is not set")
	}

	if err = internal.EnsureDirFull(targetPath); err != nil {
		return fmt.Errorf("setting-up target dir: %w", err)
	}

	mainJsonHashStr, err := cdn.GetRemoteFile(f.HashFileName(cdnBasePath))
	if err != nil {
		return fmt.Errorf("downloading main hash file: %w", err)
	}

	var mainJsonHash jsonHash
	if err = json.Unmarshal(mainJsonHashStr, &mainJsonHash); err != nil {
		return fmt.Errorf("parsing main hash file: %w", err)
	}

	// main hash covers the include files as well
	if f.Hash != mainJsonHash.Hash {
		mainJsonStr, err := cdn.GetRemoteFile(f.FileName(cdnBasePath))
		if err != nil {
			return fmt.Errorf("downloading main file: %w", err)
		}

		// validate json against predefined schema
		if err = NewJsonValidator().ValidateString(mainJsonStr); err != nil {
			return fmt.Errorf("validating main: %w", err)
		}

		downloadIncludeFilesFunc := func(incFileName string) ([]byte, error) {
			if !strings.Contains(incFileName, ".json") {
				return nil, fmt.Errorf("only json files are allowed to include: %s", incFileName)
			}
			// download include file
			incJsonStr, err := cdn.GetRemoteFile(filepath.Join(cdnBasePath, incFileName))
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
			incHashStr, err := cdn.GetRemoteFile(filepath.Join(cdnBasePath, incHashFileName))
			if err != nil {
				return nil, fmt.Errorf("downloading include file hash: %w", err)
			}
			var incJsonHash jsonHash
			if err = json.Unmarshal(incHashStr, &incJsonHash); err != nil {
				return nil, fmt.Errorf("parsing include file hash: %w", err)
			}
			if !isHashValid(incJsonHash.Hash, incJsonStr) {
				return nil, fmt.Errorf("include file integrity problem, expected hash[%s] got hash[%s]", incJsonHash.Hash, hash(incJsonStr))
			}
			// write an include json to file
			incFileTargetPath := filepath.Join(targetPath, incFileName) + tmpExt
			if err = internal.FileWrite(incFileTargetPath, incJsonStr, internal.PermUserRW); err != nil {
				return nil, fmt.Errorf("wrting include file: %w", err)
			}
			// write an include hash to file
			incFileHashTargetPath := filepath.Join(targetPath, incHashFileName) + tmpExt
			if err = internal.FileWrite(incFileHashTargetPath, incHashStr, internal.PermUserRW); err != nil {
				return nil, fmt.Errorf("wrting include hash file: %w", err)
			}
			return incJsonStr, nil
		} //func()

		incFiles, err := f.walkIncludeFiles(mainJsonStr, downloadIncludeFilesFunc)
		if err != nil {
			return fmt.Errorf("downloading include files: %w", err)
		}

		// verify content integrity
		// if main json has include failes - hash should cover whole content
		if !isHashValid(mainJsonHash.Hash, mainJsonStr, incFiles) {
			return fmt.Errorf("main file integrity problem")
		}

		// write main json to file
		localFileName := filepath.Join(targetPath, f.FileName("")) + tmpExt
		if err = internal.FileWrite(localFileName, mainJsonStr, internal.PermUserRW); err != nil {
			return fmt.Errorf("writing main file: %w", err)
		}
		// write main hash to file
		localFileName = filepath.Join(targetPath, f.HashFileName("")) + tmpExt
		if err = internal.FileWrite(localFileName, mainJsonHashStr, internal.PermUserRW); err != nil {
			return fmt.Errorf("writing main file: %w", err)
		}

		// while processing, save files with special extension '*.bu'
		// if download or handling would fail in the middle - previous files are left intact,
		// also need to cleanup tmp files (see above)
		if err = renameFiles(targetPath, tmpExt); err != nil {
			return fmt.Errorf("writing/renaming files: %w", err)
		}

		// after all is done, set new hash
		f.Hash = mainJsonHash.Hash
	}
	return nil
}

func walkFiles(targetPath, tmpExt string, actionFunc func(string)) error {
	err := filepath.WalkDir(targetPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Printf("accessing %s: %v", path, err)
			return nil // continue walking
		}
		// exclude symlinks
		if d.Type()&os.ModeSymlink != 0 {
			return nil
		}
		if !d.IsDir() && strings.HasSuffix(d.Name(), tmpExt) {
			actionFunc(path)
		}
		return nil
	})
	return err
}

// renameFiles on transaction success - rename tmp files to normal
func renameFiles(targetPath, tmpExt string) error {
	return walkFiles(targetPath, tmpExt, func(path string) {
		newPath := strings.TrimSuffix(path, tmpExt)
		if err := os.Rename(path, newPath); err != nil {
			fmt.Printf("REMOVE/DEBUG: rename %s to %s: %v\n", path, newPath, err)
		}
	})
}

// cleanupFiles on transaction failure - remove tmp files
func cleanupFiles(targetPath, tmpExt string) error {
	return walkFiles(targetPath, tmpExt, func(path string) {
		if err := os.Remove(path); err != nil {
			fmt.Printf("REMOVE/DEBUG: remove %s: %v\n", path, err)
		}
	})
}

func isValidExistingDir(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return info.IsDir(), nil
}

func isHashValid(targetHash string, data ...[]byte) bool {
	bytesForHash := []byte{}
	for _, b := range data {
		bytesForHash = append(bytesForHash, b...)
	}
	return targetHash == hash(bytesForHash)
}

// load feature config from JSON file
func (f *Feature) load(sourcePath string) error {

	if f.Name == "" {
		return fmt.Errorf("feature name is not set")
	}

	validDir, err := isValidExistingDir(sourcePath)
	if err != nil {
		return fmt.Errorf("accessing source path: %w", err)
	}
	if !validDir {
		return fmt.Errorf("config source path is not valid")
	}

	mainJsonHashStr, err := internal.FileRead(f.HashFileName(sourcePath))
	if err != nil {
		return fmt.Errorf("reading hash file: %w", err)
	}
	var mainJsonHash jsonHash
	if err = json.Unmarshal(mainJsonHashStr, &mainJsonHash); err != nil {
		return fmt.Errorf("parsing main hash file: %w", err)
	}

	mainJsonFileName := f.FileName(sourcePath)
	mainJsonStr, err := internal.FileRead(mainJsonFileName)
	if err != nil {
		return fmt.Errorf("reading config file: %w", err)
	}

	// validate json by predefined schema
	if err := NewJsonValidator().ValidateString(mainJsonStr); err != nil {
		return fmt.Errorf("validating json: %w", err)
	}

	// load json into structures
	var temp Feature
	if err := json.Unmarshal(mainJsonStr, &temp); err != nil {
		return err
	}

	validateIncludeFilesFunc := func(incFileName string) ([]byte, error) {
		if !strings.Contains(incFileName, ".json") {
			return nil, fmt.Errorf("only json files are allowed to include: %s", incFileName)
		}
		// read include file
		incJsonStr, err := internal.FileRead(filepath.Join(sourcePath, incFileName))
		if err != nil {
			return nil, fmt.Errorf("loading include file: %w", err)
		}
		// do basic json validation
		var tmpJson any
		if err = json.Unmarshal(incJsonStr, &tmpJson); err != nil {
			return nil, fmt.Errorf("parsing include file: %w", err)
		}
		// verify include file content integrity
		incHashFileName := strings.ReplaceAll(incFileName, ".json", "-hash.json")
		incHashStr, err := internal.FileRead(filepath.Join(sourcePath, incHashFileName))
		if err != nil {
			return nil, fmt.Errorf("loading include file hash: %w", err)
		}
		var incJsonHash jsonHash
		if err = json.Unmarshal(incHashStr, &incJsonHash); err != nil {
			return nil, fmt.Errorf("parsing include file hash: %w", err)
		}
		if !isHashValid(incJsonHash.Hash, incJsonStr) {
			return nil, fmt.Errorf("include file integrity problem")
		}
		return incJsonStr, nil
	} //func()

	incFiles, err := f.walkIncludeFiles(mainJsonStr, validateIncludeFilesFunc)
	if err != nil {
		return fmt.Errorf("loading include files: %w", err)
	}

	// verify content integrity
	// if main json has include failes - hash should cover whole content
	if !isHashValid(mainJsonHash.Hash, mainJsonStr, incFiles) {
		return fmt.Errorf("main file integrity problem")
	}

	f.Params = make(map[string]*Param)
	for _, cfgItem := range temp.Configs {
		f.Params[cfgItem.Name] = &Param{Name: cfgItem.Name, Type: cfgItem.Type, Settings: []ParamValue{}}
		for _, prm := range cfgItem.Settings {
			var incVal string
			switch cfgItem.Type {
			case "string":
				if _, err := prm.AsString(); err != nil {
					return fmt.Errorf("loading string value [%s]: %w", prm.Value, err)
				}
			case "integer", "int", "number":
				if _, err := prm.AsInt(); err != nil {
					return fmt.Errorf("loading int value [%s]: %w", prm.Value, err)
				}
			case "boolean", "bool":
				if _, err := prm.AsBool(); err != nil {
					return fmt.Errorf("loading bool value [%s]: %w", prm.Value, err)
				}
			case "array":
				if _, err := prm.AsStringArray(); err != nil {
					return fmt.Errorf("loading string array value [%s]: %w", prm.Value, err)
				}
			case "object":
				// load as string and validate json
				val, err := prm.AsString()
				if err != nil {
					return fmt.Errorf("loading json object as string value [%s]: %w", prm.Value, err)
				}
				var temp any
				if err := json.Unmarshal([]byte(val), &temp); err != nil {
					return fmt.Errorf("loading json object [%s]: %w", val, err)
				}
			case "file":
				// primary field value is an include file name
				val, err := prm.AsString()
				if err != nil {
					return fmt.Errorf("loading include file name as string value [%s]: %w", prm.Value, err)
				}
				// include file expect to be located in the same directory as main json file
				dir, _ := filepath.Split(mainJsonFileName)
				jsnIncFile := dir + val
				jsn, err := internal.FileRead(jsnIncFile)
				if err != nil {
					return fmt.Errorf("loading include file [%s]: %w", jsnIncFile, err)
				}
				var temp any
				// do basic json validation for include file
				if err := json.Unmarshal(jsn, &temp); err != nil {
					return fmt.Errorf("loading json from include file [%s]: %w", jsnIncFile, err)
				}
				// primary field value is still file name, store loaded file content in additional field
				incVal = string(jsn)
			}
			f.Params[cfgItem.Name].Settings = append(f.Params[cfgItem.Name].Settings,
				ParamValue{Value: prm.Value, IncValue: incVal, AppVersion: prm.AppVersion, Weight: prm.Weight})
		}
	}
	return nil
}

// walkIncludeFiles
func (f *Feature) walkIncludeFiles(mainJason []byte, fileActionFunc func(string) ([]byte, error)) ([]byte, error) {
	var temp Feature
	if err := json.Unmarshal(mainJason, &temp); err != nil {
		return nil, err
	}
	incFilesJson := []byte{}
	for _, cfgItem := range temp.Configs {
		for _, prm := range cfgItem.Settings {
			switch cfgItem.Type {
			case "file":
				// primary field value is an include file name
				incFileName, err := prm.AsString()
				if err != nil {
					return nil, fmt.Errorf("loading include file name as string value [%s]: %w", prm.Value, err)
				}
				incFile, err := fileActionFunc(incFileName)
				if err != nil {
					return nil, fmt.Errorf("downloading include file [%s]: %w", incFileName, err)
				}
				incFilesJson = append(incFilesJson, incFile...)
			}
		}
	}
	return incFilesJson, nil
}

// TODO/FIME: improve output
func (f *Feature) Print() error {
	for _, cfgItem := range f.Params {
		fmt.Println("~~~config, name:", cfgItem.Name, "type:", cfgItem.Type)
		for _, prm := range cfgItem.Settings {
			fmt.Println("~~~param, weight:", prm.Weight, ", appVersion:", prm.AppVersion)
			switch cfgItem.Type {
			case "string":
				val, _ := prm.AsString()
				fmt.Println("~~~~strVal:", val)
			case "integer", "int", "number":
				val, _ := prm.AsInt()
				fmt.Println("~~~~intVal:", val)
			case "boolean", "bool":
				val, _ := prm.AsBool()
				fmt.Println("~~~~boolVal:", val)
			case "array":
				val, _ := prm.AsStringArray()
				fmt.Println("~~~~strArrVal:", val)
			case "object":
				val, _ := prm.AsString()
				fmt.Println("~~~~jsonObjVal:", val)
			case "file":
				val, _ := prm.AsString()
				fmt.Println("~~~~file:", val)
				fmt.Println("~~~~fileStral:", prm.IncValue)
			}
		}
	}
	return nil
}
