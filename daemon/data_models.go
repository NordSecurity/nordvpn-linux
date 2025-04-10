package daemon

import (
	"bytes"
	"encoding/gob"
	"time"

	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"google.golang.org/protobuf/proto"

	"github.com/coreos/go-semver/semver"
)

const accountDataValidityPeriod = time.Minute

type AppData struct {
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

func isAccountCacheValid(added time.Time) bool {
	return time.Now().Before(added.Add(accountDataValidityPeriod))
}

type cacheValidityFunc func(time.Time) bool

type AccountData struct {
	accountData            *pb.AccountResponse
	isSet                  bool
	updatedAt              time.Time
	checkCacheValidityFunc cacheValidityFunc
}

func NewAccountData() AccountData {
	return AccountData{checkCacheValidityFunc: isAccountCacheValid}
}

func (a *AccountData) set(data *pb.AccountResponse) {
	dataCopy := proto.Clone(data).(*pb.AccountResponse)
	a.accountData = dataCopy
	a.isSet = true
	a.updatedAt = time.Now()
}

func (a *AccountData) unset() {
	a.isSet = false
	a.accountData = &pb.AccountResponse{}
}

func (a *AccountData) get(respectValidityPeriod bool) (*pb.AccountResponse, bool) {
	if !a.isSet {
		return nil, false
	}

	if respectValidityPeriod && !a.checkCacheValidityFunc(a.updatedAt) {
		a.isSet = false
		return nil, false
	}

	return proto.Clone(a.accountData).(*pb.AccountResponse), a.isSet
}
