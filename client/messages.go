package client

const (
	ConfigMessage = "It seems there's an issue with the config file. If the issue persists, please contact our customer support."

	LoginFailure       = "Username or password is not correct. Please try again."
	LegacyLoginFailure = "We couldn't log you in. Make sure your credentials are correct. If you have MFA enabled, log in using the 'nordvpn login' command."
	TokenLoginFailure  = "Token parameter value is missing." // #nosec
	TokenInvalid       = "We couldn't log you in - the access token is not valid. Please check if you've entered the token correctly. If the issue persists, contact our customer support."

	AccountTokenRenewError = "We were not able to fetch your account data. Please check your internet connection and try again. If the issue persists, please contact our customer support."
	ConnectStart           = "Connecting to %v (%v)"
	ConnectTimeoutError    = "It's not you, it's us. We're having trouble reaching our servers. If the issue persists, please contact our customer support."
	ConnectCantConnect     = "The VPN connection has failed. Please check your internet connection and try connecting to the VPN again. If the issue persists, contact our customer support."
	ConnectConnected       = "You are already connected to NordVPN."
	RelogRequest           = "For security purposes, please log in again."
	MsgTryAgain            = "We're having trouble reaching our servers. Please try again later. If the issue persists, please contact our customer support."
	UFWDisabledMessage     = "The active UFW firewall on your system prevents us from setting up our firewall properly. We have disabled UFW for the duration of your VPN connection and enabled our firewall to ensure your online security. Your custom UFW rules are imported to our firewall ruleset."

	SubscriptionURL                 = "https://my.nordaccount.com/plans/?product_group=nordvpn&login_target=nordvpn&utm_source=linux&utm_medium=app&utm_campaign=desktop-app&redirect_uri=nordvpn-sl://claim-online-purchase"
	SubscriptionURLLogin            = "https://my.nordaccount.com/plans/?product_group=nordvpn&login_target=nordvpn&utm_source=linux&utm_medium=app&utm_campaign=desktop-app&trusted_pass_token=%s&owner_id=%s&redirect_uri=nordvpn-sl://claim-online-purchase"
	SubscriptionDedicatedIPURL      = "https://my.nordaccount.com/dashboard/nordvpn/dedicatedip/?product_group=nordvpn&utm_source=linux&utm_medium=in-app&utm_campaign=desktop-app"
	SubscriptionDedicatedIPURLLogin = "https://my.nordaccount.com/dashboard/nordvpn/dedicatedip/?product_group=nordvpn&utm_source=linux&utm_medium=in-app&utm_campaign=desktop-app&trusted_pass_token=%s&owner_id=%s"
)
