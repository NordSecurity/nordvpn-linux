package daemon

import (
	"bytes"
	"encoding/gob"
	"time"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/internal"

	"github.com/coreos/go-semver/semver"
	mapset "github.com/deckarep/golang-set"
)

type AppData struct {
	CountryNames map[bool]map[config.Protocol]mapset.Set
	CityNames    map[bool]map[config.Protocol]map[string]mapset.Set
	GroupNames   map[bool]map[config.Protocol]mapset.Set
}

type VersionData struct {
	filePath              string
	version               semver.Version
	newerVersionAvailable bool
}

type InsightsData struct {
	filePath string
	Insights core.Insights
}

func (data *InsightsData) load() error {
	content, err := internal.FileRead(data.filePath)
	if err != nil {
		return err
	}
	decoder := gob.NewDecoder(bytes.NewReader(content))
	return decoder.Decode(data)
}

func (data *InsightsData) save() error {
	buffer := &bytes.Buffer{}
	encoder := gob.NewEncoder(buffer)
	err := encoder.Encode(data)
	if err != nil {
		return err
	}

	err = internal.FileWrite(data.filePath, buffer.Bytes(), internal.PermUserRWGroupROthersR)
	if err != nil {
		return err
	}
	return nil
}

type CountryData struct {
	filePath  string
	UpdatedAt time.Time
	Countries core.Countries
	Hash      string
}

func (data *CountryData) load() error {
	content, err := internal.FileRead(data.filePath)
	if err != nil {
		return err
	}
	decoder := gob.NewDecoder(bytes.NewReader(content))
	return decoder.Decode(data)
}

func (data *CountryData) save() error {
	buffer := &bytes.Buffer{}
	encoder := gob.NewEncoder(buffer)
	err := encoder.Encode(data)
	if err != nil {
		return err
	}

	err = internal.FileWrite(data.filePath, buffer.Bytes(), internal.PermUserRWGroupROthersR)
	if err != nil {
		return err
	}
	return nil
}

func (data *CountryData) exists() bool {
	return internal.FileExists(data.filePath)
}

func (data *CountryData) isValid() bool {
	return data.UpdatedAt.Add(6 * time.Hour).After(time.Now())
}

type ServersData struct {
	filePath  string
	UpdatedAt time.Time
	Servers   core.Servers
	Hash      string
}

func (data *ServersData) load() error {
	content, err := internal.FileRead(data.filePath)
	if err != nil {
		return err
	}

	decoder := gob.NewDecoder(bytes.NewReader(content))
	return decoder.Decode(data)
}

func (data *ServersData) save() error {
	buffer := &bytes.Buffer{}
	encoder := gob.NewEncoder(buffer)
	err := encoder.Encode(data)
	if err != nil {
		return err
	}

	err = internal.FileWrite(data.filePath, buffer.Bytes(), internal.PermUserRWGroupROthersR)
	if err != nil {
		return err
	}
	return nil
}

func (data *ServersData) exists() bool {
	return internal.FileExists(data.filePath)
}

func (data *ServersData) isValid() bool {
	// in order not to override servers.dat - uncomment
	// return true
	return data.UpdatedAt.Add(1 * time.Hour).After(time.Now())
}

func (data *VersionData) load() error {
	content, err := internal.FileRead(data.filePath)
	if err != nil {
		return err
	}
	decoder := gob.NewDecoder(bytes.NewReader(content))
	return decoder.Decode(&data.version)
}

func (data *VersionData) save() error {
	buffer := &bytes.Buffer{}
	encoder := gob.NewEncoder(buffer)
	err := encoder.Encode(data.version)
	if err != nil {
		return err
	}

	err = internal.FileWrite(data.filePath, buffer.Bytes(), internal.PermUserRWGroupROthersR)
	if err != nil {
		return err
	}
	return nil
}
