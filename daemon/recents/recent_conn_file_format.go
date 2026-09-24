package recents

import (
	"encoding/json"
	"errors"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/log"
)

const (
	unversionedFile   = 0
	currentFileFormat = 1
)

type recentsFile struct {
	Version     int     `json:"version"`
	Connections []Model `json:"connections"`
}

func encode(connections []Model) ([]byte, error) {
	f := recentsFile{
		Version:     currentFileFormat,
		Connections: connections,
	}
	return json.Marshal(f)
}

func decode(data []byte) (recentsFile, error) {
	if len(data) == 0 {
		return recentsFile{
			Version:     currentFileFormat,
			Connections: []Model{},
		}, nil
	}

	var f recentsFile
	err := json.Unmarshal(data, &f)
	if err == nil {
		return f, nil
	}

	var connections []Model
	if e := json.Unmarshal(data, &connections); e != nil {
		return recentsFile{}, errors.Join(e, err)
	}

	return recentsFile{
		Version:     unversionedFile,
		Connections: connections,
	}, nil
}

func migrateToVersion1(connections []Model) []Model {
	noObfuscatedConn := []Model{}
	for _, c := range connections {
		if c.Group == config.ServerGroup_OVPN_OBFUSCATED {
			log.Recents.Debug("deleting obfuscated connection:", c.CountryCode)
			continue
		}
		noObfuscatedConn = append(noObfuscatedConn, c)
	}

	return noObfuscatedConn
}
