package client

const (
	ConfigMessage = "It seems there's an issue with the config file. If the issue persists, please contact our customer support."

	LoginFailure       = "Username or password is not correct. Please try again."
	LegacyLoginFailure = "We couldn't log you in. Make sure your credentials are correct. If you have MFA enabled, log in using the 'nordvpn login' command."
	TokenLoginFailure  = "Token parameter value is missing." // #nosec
	TokenInvalid       = "We couldn't log you in - the access token is not valid. Please check if you've entered the token correctly. If the issue persists, contact our customer support."

	AccountTokenRenewError = "We were not able to fetch your account data. Please check your internet connection and try again. If the issue persists, please contact our customer support."
	ExpiredAccountMessage  = "Your account has expired. Renew your subscription now to continue enjoying the ultimate privacy and security with NordVPN."

	ConnectStart         = "Connecting to %v (%v)"
	ConnectTimeoutError  = "It's not you, it's us. We're having trouble reaching our servers. If the issue persists, please contact our customer support."
	ConnectCantConnectTo = "Whoops! We couldn't connect you to '%s'. Please try again. If the problem persists, contact our customer support."
	ConnectCantConnect   = "Whoops! Connection failed. Please try again. If the problem persists, contact our customer support."
	ConnectConnected     = "You are already connected to NordVPN."
	RelogRequest         = "For security purposes, please log in again."
	MsgTryAgain          = "Whoops! We're having trouble reaching our servers. Please try again later. If the issue persists, please contact our customer support."
	UFWDisabledMessage   = "The active UFW firewall on your system prevents us from setting up our firewall properly. We have disabled UFW for the duration of your VPN connection and enabled our firewall to ensure your online security. Your custom UFW rules are imported to our firewall ruleset."

	SubscriptionURL       = "https://join.nordvpn.com/order/?utm_campaign=%s&utm_medium=app&utm_source=linux"
	SubscriptionNoPlanURL = "https://join.nordvpn.com/order/?utm_medium=app&utm_source=linux"
)
