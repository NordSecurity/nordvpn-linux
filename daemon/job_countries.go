package daemon

import (
	"errors"
	"time"

	"github.com/NordSecurity/nordvpn-linux/core"
)

func JobCountries(dm *DataManager, api core.ServersAPI) func() error {
	return func() error {
		if dm.CountryDataExists() {
			// if db is still valid, make sure it's locked and do nothing
			if dm.IsCountryDataValid() {
				return nil
			}
		}

		// save execution start time
		currentTime := time.Now()
		countries, headers, err := api.ServersCountries()
		if err != nil {
			return err
		}

		if len(countries) == 0 {
			return errors.New("empty country list")
		}

		err = dm.SetCountryData(currentTime, countries, headers.Get(core.HeaderDigest))
		if err != nil {
			return err
		}
		return nil
	}
}
