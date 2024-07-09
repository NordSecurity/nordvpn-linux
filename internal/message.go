package internal

const (
	ConnectSuccess    = "You are connected to %s (%s)%s!"
	ReconnectSuccess  = "You have been reconnected to %s (%s)"
	DisconnectSuccess = "You are disconnected from NordVPN."

	ProtocolErrorMessage   = "protocol: failed to parse %s"
	TechnologyErrorMessage = "technology: failed to parse %s"

	DaemonConnRefusedErrorMessage = "Cannot reach System Daemon."

	ServerUnavailableErrorMessage = "The specified server is not available at the moment or does not support your connection settings."
	TagNonexistentErrorMessage    = "The specified server does not exist."
	GroupNonexistentErrorMessage  = "The specified group does not exist."
	FilterNonExistentErrorMessage = "The specified filter does not exist."
	DoubleGroupErrorMessage       = "You cannot connect to a group and set the group option at the same time."

	DebugPrefix = "[Debug]"
	// DeferPrefix is used when logging errors in deferred or cleanup code.
	DeferPrefix = "[Defer]"
	// ErrorPrefix is used when logging errors, which impact control flow.
	ErrorPrefix = "[Error]"
	// WarningPrefix is used when logging errors, which don't impact control flow.
	WarningPrefix = "[Warning]"
	InfoPrefix    = "[Info]"

	// UnhandledMessage represents the default message for unhandled errors
	UnhandledMessage = "Something went wrong. Please try again. If the problem persists, contact our customer support."

	// Error message when the server is a virtual location, but user has virtual-location off
	SpecifiedServerIsVirtualLocation = "Please enable virtual location access to connect to this server."
)
