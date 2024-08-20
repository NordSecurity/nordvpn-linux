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
)

const (
	flagGroup         = "group"
	flagToken         = "token"
	flagLoginCallback = "callback"
	stringProtocol    = "protocol"
)
