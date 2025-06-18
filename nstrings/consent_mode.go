package nstrings

import "github.com/NordSecurity/nordvpn-linux/config/consent"

// UserConsent returns user-friendly representation of `ConsentMode` mode
func UserConsent(mode consent.ConsentMode) string {
	switch mode {
	case consent.ConsentMode_DENIED:
		return disabled
	case consent.ConsentMode_GRANTED:
		return enabled
	case consent.ConsentMode_UNDEFINED:
		return undefined
	default:
		return undefined
	}
}
