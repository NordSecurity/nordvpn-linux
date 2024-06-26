package daemon

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/test/category"

	"github.com/stretchr/testify/assert"
)

const TestdataC2DatPath = TestdataPath + "c2.dat"
const TestdataS2DatPath = TestdataPath + "s2.dat"
const TestdataVersionDatPath = TestdataPath + "version.dat"

type mockCountriesAPI struct{}

func (mockCountriesAPI) Servers() (core.Servers, http.Header, error) {
	return nil, nil, nil
}

func (mockCountriesAPI) RecommendedServers(core.ServersFilter, float64, float64) (core.Servers, http.Header, error) {
	return nil, nil, nil
}

func (mockCountriesAPI) Server(int64) (*core.Server, error) {
	return nil, nil
}

func (m mockCountriesAPI) ServersCountries() (core.Countries, http.Header, error) {
	return countriesList(), nil, nil
}

func (mockCountriesAPI) ServersTechnologiesConfigurations(string, int64, core.ServerTechnology) ([]byte, error) {
	return nil, nil
}

type mockFailingCountriesAPI struct{}

func (mockFailingCountriesAPI) Servers() (core.Servers, http.Header, error) {
	return nil, nil, nil
}

func (mockFailingCountriesAPI) RecommendedServers(core.ServersFilter, float64, float64) (core.Servers, http.Header, error) {
	return nil, nil, nil
}

func (mockFailingCountriesAPI) Server(int64) (*core.Server, error) {
	return nil, nil
}

func (mockFailingCountriesAPI) ServersCountries() (core.Countries, http.Header, error) {
	return nil, nil, fmt.Errorf("500")
}

func (mockFailingCountriesAPI) ServersTechnologiesConfigurations(string, int64, core.ServerTechnology) ([]byte, error) {
	return nil, nil
}

// TestJobCountries and its sub-tests check if the country list gets populated properly
func TestJobCountries(t *testing.T) {
	category.Set(t, category.File)
	defer testsCleanup()
	dm := testNewDataManager()
	err := JobCountries(dm, mockCountriesAPI{})()
	assert.NoError(t, err)

	// check if Latvia exist
	checkLatvia := func(t *testing.T) {
		cntrExist := false
		for _, c := range dm.GetCountryData().Countries {
			if c.Name == "Latvia" {
				cntrExist = true
				break
			}
		}
		assert.True(t, cntrExist)
	}

	checkLondon := func(t *testing.T) {
		cityExist := false
		for _, cntr := range dm.GetCountryData().Countries {
			for _, city := range cntr.Cities {
				if city.Name == "London" {
					cityExist = true
					break
				}
			}
		}
		assert.True(t, cityExist)
	}
	t.Run("check country", checkLatvia)
	t.Run("check city", checkLondon)
}

// TestJobCountries_InvalidData checks if unparsable document returns an error
func TestJobCountries_InvalidData(t *testing.T) {
	category.Set(t, category.File)

	dm := testNewDataManager()
	err := JobCountries(dm, &mockFailingCountriesAPI{})()
	assert.Error(t, err)
}

// TestJobCountries_Valid checks if IsValid() condition is executed correctly
func TestJobCountries_Valid(t *testing.T) {
	category.Set(t, category.File)

	defer testsCleanup()

	internal.FileCopy(TestdataS2DatPath, TestdataPath+TestServersFile)
	internal.FileCopy(TestdataC2DatPath, TestdataPath+TestCountryFile)
	internal.FileCopy(TestdataPath+"i2.dat", TestdataPath+TestInsightsFile)
	internal.FileCopy(TestdataVersionDatPath, TestdataPath+TestVersionFile)

	dm := testNewDataManager()
	assert.NoError(t, dm.LoadData())
	original := dm.GetCountryData().Countries
	dm.SetCountryData(time.Now(), original, "")

	err := JobCountries(dm, &mockCountriesAPI{})()
	assert.NoError(t, err)
	assert.ElementsMatch(t, dm.GetCountryData().Countries, original)
}

// TestJobCountries_Expired checks if IsValid() condition is executed correctly
func TestJobCountries_Expired(t *testing.T) {
	category.Set(t, category.File)

	defer testsCleanup()

	internal.FileCopy(TestdataC2DatPath, TestdataPath+TestCountryFile)

	dm := testNewDataManager()
	original, _, _ := mockCountriesAPI{}.ServersCountries() // do not use filesystem
	dm.SetCountryData(time.Now().Add(time.Duration(-7)*time.Hour), original, "")

	err := JobCountries(dm, &mockCountriesAPI{})()
	assert.NoError(t, err)
	assert.ElementsMatch(t, dm.GetCountryData().Countries, original)
}
