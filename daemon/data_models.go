package daemon

import (
	"bytes"
	"encoding/gob"
	"errors"
	"time"

	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/internal/caching"
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

	err = internal.FileWrite(data.filePath, buffer.Bytes(), internal.PermUserRW)
	if err != nil {
		return err
	}
	return nil
}

type AccountData struct {
	cache *caching.Cache[*pb.AccountResponse]
}

func NewAccountData() AccountData {
	// create a new cache without a fetch function since we don't have the actual fetch logic here
	return AccountData{
		cache: caching.NewCacheWithTTL[*pb.AccountResponse](accountDataValidityPeriod, nil),
	}
}

func (a *AccountData) set(data *pb.AccountResponse) {
	dataCopy := proto.Clone(data).(*pb.AccountResponse)
	a.cache.Set(dataCopy)
}

func (a *AccountData) unset() {
	a.cache.Invalidate()
}

// get retrieves account data from cache.
// Parameters:
//   - respectDataExpiry: Controls cache retrieval
//     - If true: Tries to return valid cached data, or error if it is invalid
//     - If false: Returns whatever is in the cache regardless of validity, including stale data
//
// Returns:
//   - *pb.AccountResponse: The account data
//   - bool: Whether cached data was used (true) or not (false)
//     - true indicates data came from cache (valid or stale)
//     - false indicates no cache data was available or an error occurred

func (a *AccountData) get(respectDataExpiry bool) (*pb.AccountResponse, bool) {
	data, err := a.cache.Get()
	if data == nil {
		return nil, false
	}

	if respectDataExpiry && err != nil {
		return nil, false
	}

	// when not respecting expiry, we can use stale data
	if !respectDataExpiry && err != nil {
		if isStaleDataError := errors.Is(err, caching.ErrStaleData); !isStaleDataError {
			return nil, false
		}
	}

	return proto.Clone(data).(*pb.AccountResponse), true
}
