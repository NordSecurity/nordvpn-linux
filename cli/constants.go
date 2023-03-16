package cli

const (
	// MaxLoginAttempts defines maximal login attempts count
	MaxLoginAttempts = 3
	// ConfigDirName defines configuration subdirectory name
	ConfigDirName = "nordvpn/"
	// ConfigFilePath defines config file path
	ConfigFilePath = ConfigDirName + "nordvpn.conf"
	// WhitelistProtocol defines whitelist commands argument
	WhitelistProtocol = "protocol"
	// WhitelistMinPort defines min port which can be whitelisted
	WhitelistMinPort = 1
	// WhitelistMaxPort defines max port which can be whitelisted
	WhitelistMaxPort = 65535
)

const (
	flagGroup         = "group"
	flagUsername      = "username"
	flagPassword      = "password"
	flagLegacy        = "legacy"
	flagToken         = "token"
	flagLoginCallback = "callback"
	stringProtocol    = "protocol"
)
