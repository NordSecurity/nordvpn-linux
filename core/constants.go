package core

import (
	"time"

	"github.com/NordSecurity/nordvpn-linux/internal"
)

var (
	// ManualLoginTokenExpiryDate represents the expiration date for manually created tokens
	// Setting a very big expiration date here as real expiration date is unknown just from the
	// token, and there is no way to check for it. In case token is used but expired, automatic
	// logout will happen. See: core/login_token_manager.go
	// Note: bigger year cannot be used as time.Parse cannot parse year longer than 4 digits as
	// of Go 1.21
	ManualLoginTokenExpiryDate = time.Date(9999, time.December, 31, 0, 0, 0, 0, time.UTC).Format(internal.ServerDateFormat)
)
