package cli

const (
	// MaxLoginAttempts defines maximal login attempts count
	MaxLoginAttempts = 3
	// ConfigDirName defines configuration subdirectory name
	ConfigDirName = "nordvpn/"
	// ConfigFilePath defines config file path
	ConfigFilePath = ConfigDirName + "nordvpn.conf"
	// AllowlistProtocol defines allowlist commands argument
	AllowlistProtocol = "protocol"
	// AllowlistMinPort defines min port which can be allowlisted
	AllowlistMinPort = 1
	// AllowlistMaxPort defines max port which can be allowlisted
	AllowlistMaxPort = 65535
)

const (
	flagGroup         = "group"
	flagToken         = "token"
	flagLoginCallback = "callback"
	stringProtocol    = "protocol"
)
