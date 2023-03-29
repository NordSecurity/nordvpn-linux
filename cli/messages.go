package cli

const (
	ArgumentCountError   = "The command you entered has incorrect number of arguments. Enter '%s %s --help' to see the options."
	ArgumentParsingError = "The command you entered is not valid. Enter '%s %s --help' to see the options."

	LoginStart           = "Please enter your login details."
	LoginSuccess         = "Welcome to NordVPN! You can now connect to VPN by using '%s connect'."
	LoginAttempt         = "Attempt %d/%d"
	LoginTooManyAttempts = "Too many login attempts. Type '%s login' to start over."
	LogoutSuccess        = "You are logged out."

	RateNoArgsMessage    = "Type [1–5] to rate your previous connection (1 – poor, 5 – great): "
	RateNoConnectionMade = "It seems you haven’t connected to VPN yet. Please rate your experience after your first session."
	RateAlreadyRated     = "You have already provided a rating for your active/previous connection."
	RateSuccess          = "Thank you for your feedback!"

	SetReconnect = "You are connected to NordVPN. Please reconnect to enable the setting."

	MsgNothingToRate = "There was no connection - nothing to rate."
	// MsgSetSuccess is a generic success message template.
	MsgSetSuccess = "%s is set to '%s' successfully."
	// MsgAlreadySet is a generic noop message template.
	MsgAlreadySet = "%s is already set to '%s'."
	// MsgInUse is a generic dependency error message template.
	MsgInUse            = "%s is currently used by %s. Disable it first."
	MsgSetBoolArgsUsage = `[enabled]/[disabled]

%s

Supported values for [disabled]: 0, false, disable, off, disabled
Example: nordvpn set %s off

Supported values for [enabled]: 1, true, enable, on, enabled
Example: nordvpn set %s on`

	ObfuscateOnServerNotObfuscated              = "We couldn’t turn on obfuscation because the current auto-connect server doesn’t support it. Set a different server for auto-connect to use obfuscation."
	ObfuscateOffServerObfuscated                = "We couldn’t turn off obfuscation because your current auto-connect server is obfuscated by default. Set a different server for auto-connect, then turn off obfuscation."
	AutoConnectOnNonObfuscatedServerObfuscateOn = "Your selected server doesn’t support obfuscation. Choose a different server or turn off obfuscation."
	AutoConnectOnObfuscatedServerObfuscateOff   = "Turn on obfuscation to connect to obfuscated servers."
	SetAutoConnectForceOff                      = "auto-connect was turned off because the setting change is incompatible with current auto-connect options. If you wish to continue using auto-connect, please enable it again."

	SetThreatProtectionLiteDisableDNS = "Disabling DNS."

	SetDefaultsSuccess = "Settings were successfully restored to defaults."

	FirewallRequired = "Firewall must be enabled to use '%s'."

	SetNotifySuccess      = "Notifications are set to '%s' successfully."
	SetNotifyNothingToSet = "Notifications are already set to '%s'."

	SetObfuscateUnavailable = "Obfuscation is not available with the current technology. Change the technology to OpenVPN to use obfuscation."
	SetProtocolUnavailable  = "Protocol setting is not available when the set technology is not OpenVPN"

	SetTechnologyDepsError = "Missing %s kernel module or configuration utility."

	WhitelistAddPortExistsError = "Port %s (%s) is already whitelisted."
	WhitelistAddPortSuccess     = "Port %s (%s) is whitelisted successfully."

	WhitelistAddPortsExistsError = "Ports %s - %s (%s) are already whitelisted."
	WhitelistAddPortsSuccess     = "Ports %s - %s (%s) are whitelisted successfully."

	WhitelistAddSubnetExistsError = "Subnet %s is already whitelisted."
	WhitelistAddSubnetSuccess     = "Subnet %s is whitelisted successfully."

	WhitelistRemovePortExistsError = "Port %s (%s) is not whitelisted."
	WhitelistRemovePortSuccess     = "Port %s (%s) is removed from the whitelist successfully."

	WhitelistRemovePortsExistsError = "Ports %s - %s (%s) are not whitelisted."
	WhitelistRemovePortsSuccess     = "Ports %s - %s (%s) are removed from the whitelist successfully."

	WhitelistRemoveSubnetExistsError = "Subnet %s is not whitelisted."
	WhitelistRemoveSubnetSuccess     = "Subnet %s is removed from a whitelist successfully."

	WhitelistRemoveAllError   = "Whitelist elements could not be removed."
	WhitelistRemoveAllSuccess = "All ports and subnets have been removed from the whitelist successfully."

	WhitelistPortRangeError  = "Port %s value is out of range [%s - %s]."
	WhitelistPortsRangeError = "Ports %s - %s value is out of range [%s - %s]."

	AccountCreationSuccess = "Account has been successfully created."
	// AccountInvalidData is displayed when backend returns bad request (400)
	AccountInvalidData = "Invalid email address or password. Please make sure you're entering a valid email address and your password contains at least 8 characters."
	// AccountEmailTaken is displayed when backend returns conflict (409)
	AccountEmailTaken = "User with the specified email address already exists."
	// AccountInternalError is returned when backend returns internal error (500)
	AccountInternalError          = "It's not you, it's us. We're having trouble with our servers. If the issue persists, please contact our customer support."
	AccountTokenUnauthorizedError = "There was a problem with your credentials. Please try to log out and log back in again. If the issue persists, please contact our customer support."
	AccountCantFetchVPNService    = "We were not able to fetch your VPN service data. If the issue persists, please contact our customer support."
	UpdateAvailableMessage        = "A new version of NordVPN is available! Please update the application."
	DisconnectNotConnected        = "You are not connected to NordVPN."
	DisconnectConnectionRating    = "How would you rate your connection quality on a scale from 1 (poor) to 5 (excellent)? Type '%s rate [1-5]'."

	CitiesNotFoundError = "Servers by city are not available for this country."

	CheckYourInternetConnMessage = "Please check your internet connection and try again."
	ExpiredAccountMessage        = "Your account has expired. Renew your subscription now to continue enjoying the ultimate privacy and security with NordVPN."
	NoSuchCommand                = "Command '%s' doesn't exist."

	// Meshnet
	MsgSetMeshnetUsage     = "Enables or disables meshnet on this device."
	MsgSetMeshnetArgsUsage = `[enabled]/[disabled]

Use this command to enable or disable meshnet.

Supported values for [disabled]: 0, false, disable, off, disabled
Example: nordvpn set meshnet off

Supported values for [enabled]: 1, true, enable, on, enabled
Example: nordvpn set meshnet on`

	MsgSetMeshnetSuccess            = "Meshnet is set to '%s' successfully."
	MsgMeshnetAlreadyEnabled        = "Meshnet is already enabled."
	MsgMeshnetAlreadyDisabled       = "Meshnet is already disabled."
	MsgMeshnetNotEnabled            = "Meshnet is not enabled."
	MsgMeshnetNordlynxMustBeEnabled = "NordLynx technology must be set to use this feature."
	MsgMeshnetVersionNotSupported   = "Current application version does not support the meshnet feature."
	MsgMeshnetUsage                 = "Manages mesh network and access to it. In order to enable the feature, execute `nordvpn set meshnet on`"
	MsgMeshnetRefreshUsage          = "Refreshes the meshnet in case it was not updated automatically."
	MsgMeshnetPeerUnknown           = "Peer '%s' is unknown.\n"

	// Invites
	MsgMeshnetInviteUsage = "Displays the list of all sent and received meshnet invitations. " +
		"If [email] argument is passed, sends an invitation to join the mesh network to a specified email."
	MsgMeshnetInviteListUsage                 = "Displays the list of all sent and received meshnet invitations."
	MsgMeshnetInviteAcceptUsage               = "Accepts an invitation to join inviter's mesh network."
	MsgMeshnetInviteDenyUsage                 = "Denies an invitation to join inviter's mesh network."
	MsgMeshnetInviteRevokeUsage               = "Revokes a sent invitation."
	MsgMeshnetInviteNoInvitationFound         = "no invitation from '%s' was found"
	MsgMeshnetInviteArgsUsage                 = "[email]"
	MsgMeshnetInviteAcceptSuccess             = "Meshnet invitation from '%s' was accepted."
	MsgMeshnetInviteAcceptDeviceCount         = "Maximum device count reached. Consider removing one or more of your devices."
	MsgMeshnetInviteSentSuccess               = "Meshnet invitation to '%s' was sent."
	MsgMeshnetInviteDenySuccess               = "Meshnet invitation from '%s' was denied."
	MsgMeshnetInviteRevokeSuccess             = "Meshnet invitation to '%s' was revoked."
	MsgMeshnetInviteSendUsage                 = "Sends an invitation to join the mesh network."
	MsgMeshnetInviteSendAlreadyExists         = "Meshnet invitation for '%s' already exists."
	MsgMeshnetInviteSendInvalidEmail          = "Invalid email '%s'."
	MsgMeshnetInviteSendSameAccountEmail      = "Email should belong to a different user."
	MsgMeshnetInviteSendDeviceCount           = "The peer you're trying to invite has maximum device count reached."
	MsgMeshnetInviteWeeklyLimit               = "Weekly invitation limit reached."
	MsgMeshnetInviteAllowIncomingTrafficUsage = "Allow incoming traffic from a peer."
	MsgMeshnetAllowTrafficRoutingUsage        = "Allow the peer to route traffic through this device."
	MsgMeshnetAllowLocalNetworkUsage          = "Allow the peer to access local network when routing traffic through this device."
	MsgMeshnetAllowFileshare                  = "Allow the peer to send you files."

	// Peers
	MsgMeshnetPeerListFilters   = "Filters list of available peers in a meshnet. To apply multiple filters, separate them with a comma. Please note that you will see an empty list if you apply contradictory filters."
	MsgMeshnetPeerUsage         = "Handles meshnet peer list."
	MsgMeshnetPeerArgsUsage     = "[public_key|hostname|ip]"
	MsgMeshnetPeerListUsage     = "Lists available peers in a meshnet."
	MsgMeshnetPeerRemoveUsage   = "Removes a peer from a meshnet."
	MsgMeshnetPeerRemoveSuccess = "Peer '%s' has been removed from the meshnet."

	MsgMeshnetPeerRoutingUsage          = "Allows/denies a peer device to route all traffic through this device."
	MsgMeshnetPeerRoutingAllowUsage     = "Allows a meshnet peer to route its' traffic through this device."
	MsgMeshnetPeerRoutingDenyUsage      = "Denies a meshnet peer to route its' traffic through this device."
	MsgMeshnetPeerRoutingAlreadyAllowed = "Traffic routing for '%s' is already allowed."
	MsgMeshnetPeerRoutingAlreadyDenied  = "Traffic routing for '%s' is already denied."
	MsgMeshnetPeerRoutingAllowSuccess   = "Traffic routing for '%s' has been allowed."
	MsgMeshnetPeerRoutingDenySuccess    = "Traffic routing for '%s' has been denied."

	MsgMeshnetPeerIncomingUsage          = "Allows/denies a peer device to access this device remotely (incoming connections)."
	MsgMeshnetPeerIncomingAllowUsage     = "Allows a meshnet peer to send traffic to this device."
	MsgMeshnetPeerIncomingDenyUsage      = "Denies a meshnet peer to send traffic to this device."
	MsgMeshnetPeerIncomingAlreadyAllowed = "Incoming traffic for '%s' is already allowed."
	MsgMeshnetPeerIncomingAlreadyDenied  = "Incoming traffic for '%s' is already denied."
	MsgMeshnetPeerIncomingAllowSuccess   = "Incoming traffic for '%s' has been allowed."
	MsgMeshnetPeerIncomingDenySuccess    = "Incoming traffic for '%s' has been denied."

	MsgMeshnetPeerLocalNetworkUsage          = "Allows/denies access to your local network when a peer device is routing traffic through this device."
	MsgMeshnetPeerLocalNetworkAllowUsage     = "Allows a meshnet peer to access local network when routing traffic through this device."
	MsgMeshnetPeerLocalNetworkDenyUsage      = "Denies a meshnet peer to access local network when routing traffic through this device."
	MsgMeshnetPeerLocalNetworkAlreadyAllowed = "Local network access for '%s' is already allowed."
	MsgMeshnetPeerLocalNetworkAlreadyDenied  = "Local network access for '%s' is already denied."
	MsgMeshnetPeerLocalNetworkAllowSuccess   = "Local network access for '%s' has been allowed."
	MsgMeshnetPeerLocalNetworkDenySuccess    = "Local network access for '%s' has been denied."

	MsgMeshnetPeerFileshareUsage          = "Allows/denies peer to send files to this device."
	MsgMeshnetPeerFileshareAllowUsage     = "Allows a meshnet peer to send files to this device."
	MsgMeshnetPeerFileshareDenyUsage      = "Denies a meshnet peer to send files to this device."
	MsgMeshnetPeerFileshareAlreadyAllowed = "Fileshare for '%s' is already allowed."
	MsgMeshnetPeerFileshareAlreadyDenied  = "Fileshare for '%s' is already denied."
	MsgMeshnetPeerFileshareAllowSuccess   = "Fileshare for '%s' has been allowed."
	MsgMeshnetPeerFileshareDenySuccess    = "Fileshare for '%s' has been denied."

	MsgMeshnetPeerConnectUsage        = "Treats a peer as a VPN server and connects to it if the peer has allowed traffic routing."
	MsgMeshnetPeerConnectSuccess      = "You are connected to meshnet exit node '%s'."
	MsgMeshnetPeerDoesNotAllowRouting = "Meshnet peer '%s' does not allow traffic routing."
	MsgMeshnetPeerAlreadyConnected    = "You are already connected."
	MsgMeshnetPeerConnectFailed       = "Connect to other mesh peer failed - check if peer '%s' is online."

	// Fileshare
	FileshareName       = "fileshare"
	FileshareSendName   = "send"
	FileshareAcceptName = "accept"
	FileshareCancelName = "cancel"
	FileshareListName   = "list"

	flagFileshareNoWait  = "background"
	flagFilesharePath    = "path"
	flagFileshareListIn  = "incoming"
	flagFileshareListOut = "outgoing"

	MsgFileshareUsage                     = "Transfer files of any size between Meshnet peers securely and privately."
	MsgFileshareTransferNotFound          = "Transfer not found."
	MsgFileshareInvalidPath               = "Invalid path provided: %s"
	MsgFilesharePathNotFound              = "Download directory %q does not exist. Make sure the directory exists or provide an alternative via --" + flagFilesharePath
	MsgFileshareAcceptPathIsASymlink      = "A download path can’t be a symbolic link. Please provide a directory as a download path to accept the transfer."
	MsgFileshareAcceptPathIsNotADirectory = "Please provide a directory as a download path to accept the transfer."
	MsgFileshareInvalidPeer               = "Peer name is invalid."
	MsgFileshareDisconnectedPeer          = "Peer is disconnected."
	MsgFileshareFileNotFound              = "File not found."
	MsgFileshareSocketNotFound            = "Enable Meshnet to share files. If Meshnet is already enabled, try disabling and enabling it again."

	MsgFileshareAcceptHomeError      = "Cannot determine default download path. Please provide download path explicitly via --" + flagFilesharePath
	MsgFileshareAcceptAllError       = "Download couldn't start."
	MsgFileshareAcceptOutgoingError  = "Can't accept outgoing transfer."
	MsgFileshareAlreadyAcceptedError = "This transfer is already completed."
	MsgFileshareFileInvalidated      = "The transfer of this file is already completed or canceled."
	MsgFileshareTransferInvalidated  = "This transfer is already completed or canceled."
	MsgTooManyFiles                  = "Number of files in a transfer cannot exceed 1000. Try archiving the directory."
	MsgNoFiles                       = "The directory you’re trying to send is empty. Please choose another one."
	MsgDirectoryToDeep               = "File depth cannot exceed 5 directories. Try archiving the directory."
	MsgSendingNotAllowed             = "This peer does not allow file transfers from you."
	MsgFileNotInProgress             = "This file is not in progress"
	MsgNotEnoughSpace                = "The transfer can't be accepted because there's not enough storage on your device."

	MsgFileshareSendUsage     = "Send files or directories to a Meshnet peer."
	MsgFileshareSendArgsUsage = "[peer ip|peer hostname|peer pubkey] [path_1] [path_2]...\n\nTo cancel a transfer in progress, press Ctrl+C"
	MsgFileshareNoWaitUsage   = "Send a file transfer in the background instead of seeing its progress. It allows you to continue using the terminal for other commands while a transfer is in progress."
	MsgFileshareSendNoWait    = "File transfer %s has started in the background."
	MsgFileshareAcceptNoWait  = "File transfer has started in the background."
	MsgFileshareWaitAccept    = "Waiting for the peer to accept your transfer..."
	MsgTransferNotCreated     = "Can’t send the files. Please check if you have the \"read\" permission for the files you want to send."

	MsgFileshareListUsage     = "Lists transfers. If transfer ID is provided, lists files in the transfer."
	MsgFileshareListArgsUsage = `[transfer_id]

Adding no arguments to the command will list transfers.
Provide a [transfer_id] argument to list files in the specified transfer.`
	MsgFileshareListInUsage     = "Show only incoming transfers."
	MsgFileshareListOutUsage    = "Show only outgoing transfers."
	MsgFileshareCancelUsage     = "Cancel a transfer or a single file. To cancel an entire transfer, specify the transfer ID. To cancel a single file, specify the transfer ID and the file ID."
	MsgFileshareCancelArgsUsage = "[transfer_id] [file_id]"
	MsgFileshareCancelSuccess   = "File transfer canceled"
	MsgFileshareAcceptUsage     = "Accept an incoming file transfer. To download an entire transfer, specify the transfer ID. To download a single file, specify the transfer ID and the file ID."
	MsgFileshareAcceptArgsUsage = "[transfer_id] [file_id1] [file_id2]...\n\nTo cancel a transfer in progress, press Ctrl+C"
	MsgFileshareAcceptPathUsage = "Specify download path (default: $XDG_DOWNLOAD_DIR or $HOME/Downloads)"

	MsgFileshareProgressOngoing        = "File transfer [%s] progress [%d%%]"
	MsgFileshareProgressFinished       = "File transfer [%s] completed.      " // Need extra spaces to cover the progress message
	MsgFileshareProgressFinishedErrors = "File transfer [%s] completed. Some of the files have failed to transfer."
	MsgFileshareProgressCanceledByPeer = "File transfer [%s] canceled by peer."
	MsgFileshareProgressCanceled       = "File transfer [%s] canceled by other process."
)
