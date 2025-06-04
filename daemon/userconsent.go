//go:build moose

package daemon

import (
	"log"
	"strings"
	"time"

	"github.com/NordSecurity/nordvpn-linux/auth"
	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

type countryCode string

var countryCodeToConsentMode = map[countryCode]consentMode{
	countryCode("us"): consentModeStandard,
	countryCode("ca"): consentModeStandard,
	countryCode("jp"): consentModeStandard,
	countryCode("au"): consentModeStandard,
}

type consentMode uint

const (
	// consentModeStandard mode describes countries with loose legal requirements regarding analytics consent
	consentModeStandard consentMode = iota
	// consentModeGDPR mode describes countries with stirct analytics consent requirements
	consentModeGDPR
)

type AnalyticsConsentChecker struct {
	cm          *config.FilesystemConfigManager
	API         *core.DefaultAPI
	authChecker *auth.RenewingChecker
}

func NewConsentChecker(
	cm *config.FilesystemConfigManager,
	API *core.DefaultAPI,
	authChecker *auth.RenewingChecker,
) *AnalyticsConsentChecker {
	return &AnalyticsConsentChecker{cm, API, authChecker}
}

// PrepareDaemonIfConsentNotCompleted sets up the daemon for analytics consent flow.
//
// If consent flow was completed, this is no-op. Otherwise:
//
// - using Insights API find user location
// - based on the location determine if user is in standard consent mode or GDPR mode (more strict)
//
// - for GDPR mode:
//   - do light logout, it forces the user to login to application which triggers consent flow
//
// - for standard mode:
//   - save consent as completed and accepted, no consent flow for standard mode countries
func (acc *AnalyticsConsentChecker) PrepareDaemonIfConsentNotCompleted() {
	if IsConsentFlowCompleted(acc.cm) {
		// nothing to do
		return
	}

	consentMode := acc.consentModeFromUserLocation()

	// logout user if in GDPR consent mode
	if consentMode == consentModeGDPR && acc.authChecker.IsLoggedIn() {
		if err := retryIfFailed(acc.doLightLogout); err != nil {
			log.Println(internal.ErrorPrefix, "failed to perform light logout:", err)
		}
	}

	// standard mode has analytics enabled by default and no required
	// consent flow, so update the config with `AnalyticsConsent := true`
	if consentMode == consentModeStandard {
		if err := retryIfFailed(acc.setConsentTrue); err != nil {
			log.Println(internal.ErrorPrefix, "failed to save analytics consent", err)
		}
	}
}

func retryIfFailed(fn func() error) error {
	return internal.Retry(3, time.Millisecond*200, fn)
}

func (acc *AnalyticsConsentChecker) setConsentTrue() error {
	return acc.cm.SaveWith(func(c config.Config) config.Config {
		enabled := true
		c.AnalyticsConsent = &enabled
		return c
	})
}

// IsConsentFlowCompleted reads configuration file and
// checks if `AnalyticsConsent` field is set.
func IsConsentFlowCompleted(cm config.Manager) bool {
	var cfg config.Config
	if err := cm.Load(&cfg); err != nil {
		log.Println(internal.ErrorPrefix, "failed to load config when checking consent flow", err)
		return false
	}
	return cfg.AnalyticsConsent != nil
}

// consentModeFromUserLocation in a happy path, uses Insights API to get user's
// location and compares it to list of countries in standard mode, if not on the
// list, then user is in GDPR country.
//
// Additionally:
// - in case of issue with reading configuration, fallback to GDPR mode
// - if user has KillSwitch enabled, no traffic is going out, fallback to GDPR mode
// - if there is an issue with making API request, fallback to GDPR mode
func (acc *AnalyticsConsentChecker) consentModeFromUserLocation() consentMode {
	var cfg config.Config
	if err := acc.cm.Load(&cfg); err != nil {
		log.Println(internal.ErrorPrefix, "failed to load config when determining consent mode:", err)
		// fallback to strict mode in case of an issue with config
		return consentModeGDPR
	}

	// can't determine user location with KS on, fallback to strict mode
	if cfg.KillSwitch {
		log.Println(internal.WarningPrefix, "KillSwitch active, falling back to GDPR mode")
		return consentModeGDPR
	}

	// fallback to strict mode in case of an issue with API
	insights, err := acc.API.Insights()
	if err != nil {
		log.Println(internal.WarningPrefix, "insights api error, falling back to GDRP mode:", err)
		return consentModeGDPR
	}

	// fallback to strict mode in case of nil response
	if insights == nil {
		log.Println(internal.WarningPrefix, "insigts data is nil, falling back to GDPR mode")
		return consentModeGDPR
	}

	return modeForCountryCode(countryCode(strings.ToLower(insights.CountryCode)))
}

func (acc *AnalyticsConsentChecker) doLightLogout() error {
	return acc.cm.SaveWith(func(c config.Config) config.Config {
		delete(c.TokensData, c.AutoConnectData.ID)
		c.AutoConnectData.ID = 0
		return c
	})
}

// modeForCountryCode returns analytics consent mode.
//
// It uses country code and list of countries in standard mode to check it.
// Countries not on the standard mode list fall into GDPR mode.
func modeForCountryCode(cc countryCode) consentMode {
	mode, ok := countryCodeToConsentMode[cc]
	if !ok {
		return consentModeGDPR
	}
	return mode
}
