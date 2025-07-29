package remote

import (
	"fmt"
	"path/filepath"
)

const (
	fieldTypeString = "string"
	fieldTypeInt    = "integer"
	fieldTypeNumber = "number"
	fieldTypeBool   = "boolean"
	fieldTypeArray  = "array"
	fieldTypeObject = "object"
	fieldTypeFile   = "file"
)

type ParamValue struct {
	Value      any    `json:"value"`
	incValue   string // include file content, unexported, ingnored by json.unmarshal
	AppVersion string `json:"app_version"`
	Weight     int    `json:"weight"`
	Rollout    int    `json:"rollout"`
}

func (pv ParamValue) AsString() (string, error) {
	v, ok := pv.Value.(string)
	if !ok {
		return "", fmt.Errorf("not a string")
	}
	return v, nil
}
func (pv ParamValue) AsInt() (int, error) {
	v, ok := pv.Value.(int)
	if !ok {
		return 0, fmt.Errorf("not an int")
	}
	return v, nil
}
func (pv ParamValue) AsBool() (bool, error) {
	v, ok := pv.Value.(bool)
	if !ok {
		return false, fmt.Errorf("not a bool")
	}
	return v, nil
}
func (pv ParamValue) AsStringArray() ([]string, error) {
	v, ok := pv.Value.([]any)
	if !ok {
		return nil, fmt.Errorf("not an array, got: %T", pv.Value)
	}
	strSlice := make([]string, len(v))
	for i, elem := range v {
		s, ok := elem.(string)
		if !ok {
			return nil, fmt.Errorf("element %d is not a string, got %T", i, elem)
		}
		strSlice[i] = s
	}
	return strSlice, nil
}

type Param struct {
	Name     string       `json:"name"`
	Type     string       `json:"value_type"`
	Settings []ParamValue `json:"settings"`
}

// Feature is set of configs from one JSON file
type Feature struct {
	Version int               `json:"version"` // JSON schema version
	Configs []Param           `json:"configs"`
	Schema  string            `json:"schema"`
	name    string            // file name (main part)
	hash    string            // JSON file hash (from last download)
	params  map[string]*Param // parsed params
}

func (f Feature) FilePath(basePath string) string {
	return filepath.Join(basePath, f.name) + ".json"
}
func (f Feature) HashFilePath(basePath string) string {
	return filepath.Join(basePath, f.name) + "-hash.json"
}

type FeatureMap struct {
	featureKeys []string
	featureMap  map[string]*Feature
}

func NewFeatureMap() *FeatureMap {
	return &FeatureMap{
		featureKeys: make([]string, 0),
		featureMap:  make(map[string]*Feature),
	}
}

func (m *FeatureMap) keys() []string {
	return m.featureKeys
}

func (m *FeatureMap) features() map[string]*Feature {
	return m.featureMap
}

func (m *FeatureMap) add(name string) {
	m.featureKeys = append(m.featureKeys, name)
	m.featureMap[name] = &Feature{
		name: name,
	}
}

func (m *FeatureMap) get(name string) *Feature {
	f, ok := m.featureMap[name]
	if ok {
		return f
	}
	return nil
}
