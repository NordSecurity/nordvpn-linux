package recents

import (
	"encoding/json"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/log"
)

const (
	unversionedFile   = 0
	currentFileFormat = 1
)

type FileFormat struct {
	Version     int     `json:"version"`
	Connections []Model `json:"connections"`
}

func encode(connections []Model) ([]byte, error) {
	f := FileFormat{
		Version:     currentFileFormat,
		Connections: connections,
	}
	return json.Marshal(f)
}

func decode(data []byte) (FileFormat, error) {
	if len(data) == 0 {
		return FileFormat{
			Version:     currentFileFormat,
			Connections: []Model{},
		}, nil
	}

	var f FileFormat
	if err := json.Unmarshal(data, &f); err == nil {
		return f, nil
	}

	var connections []Model
	if err := json.Unmarshal(data, &connections); err != nil {
		return FileFormat{}, err
	}

	return FileFormat{
		Version:     unversionedFile,
		Connections: connections,
	}, nil
}

func migrateToVersion1(connections []Model) FileFormat {
	noObfuscatedConn := []Model{}
	for _, c := range connections {
		if c.Group == config.ServerGroup_OBFUSCATED {
			log.Recents.Debug("deleting obfuscated connection:", c.CountryCode)
			continue
		}
		if c.ConnectionTech == config.Technology_UNKNOWN_TECHNOLOGY {
			c.ConnectionTech = config.Technology_NORDLYNX
		}
		noObfuscatedConn = append(noObfuscatedConn, c)
	}

	return FileFormat{
		Version:     1,
		Connections: noObfuscatedConn,
	}
}
