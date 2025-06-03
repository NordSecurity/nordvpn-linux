//go:build moose

package daemon

import (
	"log"

	"github.com/NordSecurity/nordvpn-linux/auth"
	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

type countryCode string

var countryCodeToConsentMode = map[countryCode]consentMode{
	"US": standard,
	"CA": standard,
	"JP": standard,
	"AU": standard,
}

type consentMode uint

const (
	// standard mode describes countries with loose legal requirements regarding analytics consent
	standard consentMode = iota
	// GDPR mode describes countries with stirct analytics consent requirements
	GDPR
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

// PrepareDaemonIfConsentNotCompleted sets up the deamon for analytics consent flow.
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
	if consentMode == GDPR && acc.authChecker.IsLoggedIn() {
		if err := acc.doLightLogout(); err != nil {
			// XXX: improve this
			log.Println(internal.ErrorPrefix, "failed to logout the user:", err)
		}
	}

	// standard mode has analytics enabled by default and no required
	// consent flow, so update the config with `AnalyticsConsent := true`
	if consentMode == standard {
		if err := acc.cm.SaveWith(func(c config.Config) config.Config {
			enabled := true
			c.AnalyticsConsent = &enabled
			return c
		}); err != nil {
			// XXX: improve and what's next?
			log.Println(internal.ErrorPrefix, "failed to save analytics consent", err)
		}
	}
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
		// fallback to strict mode in case of an issue
		return GDPR
	}

	// can't determine user location with KS on, fallback to strict mode
	if cfg.KillSwitch {
		return GDPR
	}

	insights, err := acc.API.Insights()
	if insights == nil || err != nil {
		log.Println(internal.ErrorPrefix, "failed to get insights: (insights, error) =", insights, err)
		// fallback to strict mode in case of an issue
		return GDPR
	}

	return modeForCountryCode(countryCode(insights.CountryCode))
}

func (acc *AnalyticsConsentChecker) doLightLogout() error {
	if err := acc.cm.SaveWith(func(c config.Config) config.Config {
		delete(c.TokensData, c.AutoConnectData.ID)
		c.AutoConnectData.ID = 0
		c.Mesh = false
		c.MeshPrivateKey = ""
		return c
	}); err != nil {
		return err
	}
	return nil
}

// modeForCountryCode returns analytics consent mode.
//
// It uses country code and list of countries in standard mode to check it.
// Countries not on the standard mode list fall into GDPR mode.
func modeForCountryCode(cc countryCode) consentMode {
	mode, ok := countryCodeToConsentMode[cc]
	if !ok {
		return GDPR
	}
	return mode
}
