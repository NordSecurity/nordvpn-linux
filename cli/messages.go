package cli

const (
	ArgumentCountError   = "The command you entered has incorrect number of arguments. Enter '%s --help' to see the options."
	ArgumentParsingError = "The command you entered is not valid. Enter '%s --help' to see the options."

	ClaimOnlinePurchaseSuccess = `You’re all set!
You’ve successfully purchased the NordVPN subscription.
You can use NordVPN on 10 devices at the same time.`
	ClaimOnlinePurchaseFailure = `Payment failed.
We couldn’t process your payment. Please try again.`

	LoginSuccess          = "Welcome to NordVPN! You can now connect to VPN by using '%s connect'."
	LogoutSuccess         = "You are logged out."
	LogoutTokenSuccess    = "You have been logged out. To keep your account secure, we've revoked your current access token. If you want to reuse your next access token despite the potential risks, use the --" + flagPersistToken + " option when logging out."
	LogoutUsageText       = "Logs you out"
	PersistTokenUsageText = "Keep your current access token valid after logging out."

	MsgNordVPNGroup = "By default, all users who are members of the 'nordvpn' group have permission to control the NordVPN application.\nTo limit access exclusively to the root user, remove all users from the 'nordvpn' group."

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
	MsgInUse              = "%s is currently used by %s. Disable it first."
	MsgSetBoolArgsUsage   = `<enabled>|<disabled>`
	MsgSetBoolDescription = `%s

Supported values for <disabled>: 0, false, disable, off, disabled
Example: nordvpn set %s off

Supported values for <enabled>: 1, true, enable, on, enabled
Example: nordvpn set %s on`

	ObfuscateOnServerNotObfuscated              = "We couldn’t turn on obfuscation because the current auto-connect server doesn’t support it. Set a different server for auto-connect to use obfuscation."
	ObfuscateOffServerObfuscated                = "We couldn’t turn off obfuscation because your current auto-connect server is obfuscated by default. Set a different server for auto-connect, then turn off obfuscation."
	AutoConnectOnNonObfuscatedServerObfuscateOn = "Your selected server doesn’t support obfuscation. Choose a different server or turn off obfuscation."
	AutoConnectOnObfuscatedServerObfuscateOff   = "Turn on obfuscation to connect to obfuscated servers."
	SetAutoConnectForceOff                      = "auto-connect was turned off because the setting change is incompatible with current auto-connect options. If you wish to continue using auto-connect, please enable it again."

	SetThreatProtectionLiteDisableDNS = "Disabling DNS."
	SetThreatProtectionLiteAlreadySet = "Threat Protection Lite already set to %s."

	SetDefaultsSuccess = "Settings were successfully restored to defaults."

	FirewallRequired = "Firewall must be enabled to use '%s'."

	SetNotifySuccess      = "Notifications are set to '%s' successfully."
	SetNotifyNothingToSet = "Notifications are already set to '%s'."

	SetTraySuccess      = "Tray set to '%s' successfully."
	SetTrayNothingToSet = "Tray is already set to '%s'."

	SetObfuscateUnavailable = "Obfuscation is not available with the current technology. Change the technology to OpenVPN to use obfuscation."

	SetProtocolUnavailable = "Protocol setting is not available when the set technology is not OpenVPN"
	SetProtocolAlreadySet  = "Protocol is already set to %s"

	SetTechnologyDepsError = "Missing %s kernel module or configuration utility."

	SetDNSDisableThreatProtectionLite = "Disabling Threat Protection Lite."
	SetDNSInvalidAddress              = "The provided IP address is invalid."
	SetDNSTooManyValues               = "More than 3 DNS addresses provided."
	SetDNSAlreadySet                  = "DNS is already set to %s."

	SetLANDiscoveryUsage          = "Access printers, TVs, and other devices on your local network while connected to a VPN."
	SetLANDiscoveryAlreadyEnabled = "LAN discovery is already set to %s."
	SetLANDiscoveryAllowlistReset = "Just a little heads-up: Enabling local network discovery will remove your private subnets from the allowlist."

	AllowlistAddPortExistsError = "Port %d (%s) is already allowlisted."
	AllowlistAddPortSuccess     = "Port %d (%s) is allowlisted successfully."

	AllowlistAddPortsExistsError = "Ports %d - %d (%s) are already allowlisted."
	AllowlistAddPortsSuccess     = "Ports %d - %d (%s) are allowlisted successfully."

	AllowlistAddSubnetExistsError  = "Subnet %s is already allowlisted."
	AllowlistAddSubnetSuccess      = "Subnet %s is allowlisted successfully."
	AllowlistAddSubnetLANDiscovery = "Allowlisting a private subnet is not available while local network discovery is enabled."

	AllowlistRemovePortExistsError = "Port %d (%s) is not allowlisted."
	AllowlistRemovePortSuccess     = "Port %d (%s) is removed from the allowlist successfully."

	AllowlistRemovePortsExistsError = "Ports %d - %d (%s) are not allowlisted."
	AllowlistRemovePortsSuccess     = "Ports %d - %d (%s) are removed from the allowlist successfully."

	AllowlistRemoveSubnetExistsError = "Subnet %s is not allowlisted."
	AllowlistRemoveSubnetSuccess     = "Subnet %s is removed from the allowlist successfully."

	AllowlistRemoveAllError   = "Allowlist elements could not be removed."
	AllowlistRemoveAllSuccess = "All ports and subnets have been removed from the allowlist successfully."

	AllowlistPortRangeError  = "Port %d value is out of range [%d - %d]."
	AllowlistPortsRangeError = "Ports %d - %d value is out of range [%d - %d]."

	AccountCreationSuccess = "Account has been successfully created."
	// AccountInvalidData is displayed when backend returns bad request (400)
	AccountInvalidData = "Invalid email address or password. Please make sure you're entering a valid email address and your password contains at least 8 characters."
	// AccountEmailTaken is displayed when backend returns conflict (409)
	AccountEmailTaken = "User with the specified email address already exists."
	// AccountInternalError is returned when backend returns internal error (500)
	AccountInternalError          = "It's not you, it's us. We're having trouble with our servers. If the issue persists, please contact our customer support."
	AccountTokenUnauthorizedError = "There was a problem with your credentials. Please try to log out and log back in again. If the issue persists, please contact our customer support."
	AccountCantFetchVPNService    = "We were not able to fetch your %s service data. If the issue persists, please contact our customer support."
	UpdateAvailableMessage        = "A new version of NordVPN is available!\nPlease update the application."
	DisconnectNotConnected        = "You are not connected to NordVPN."
	DisconnectConnectionRating    = "How would you rate your connection quality on a scale from 1 (poor) to 5 (excellent)? Type '%s rate [1-5]'."

	CitiesNotFoundError = "Servers by city are not available for this country."

	CheckYourInternetConnMessage           = "Please check your internet connection and try again."
	ExpiredAccountMessage                  = "Your account has expired. Renew your subscription now to continue enjoying the ultimate privacy and security with NordVPN. %s" // TODO: update once we get new error message.
	NoDedicatedIPMessage                   = "You don’t have a dedicated IP subscription. To get a personal IP address that belongs only to you, continue in the browser: \n%s"
	NoDedidcatedIPServerMessage            = "This server isn't currently included in your dedicated IP subscription."
	NoPreferredDedicatedIPLocationSelected = "It looks like you haven’t selected the preferred server location for your dedicated IP. Please head over to Nord Account and set up your dedicated IP server."
	NoSuchCommand                          = "Command '%s' doesn't exist."
	MsgListIsEmpty                         = "We couldn’t load the list of %s. Please try again later."

	// Meshnet
	MsgSetMeshnetUsage       = "Enables or disables Meshnet on this device."
	MsgSetMeshnetArgsUsage   = `<enabled>|<disabled>`
	MsgSetMeshnetDescription = `Use this command to enable or disable Meshnet.

Supported values for <disabled>: 0, false, disable, off, disabled
Example: nordvpn set meshnet off

Supported values for <enabled>: 1, true, enable, on, enabled
Example: nordvpn set meshnet on`

	MsgSetMeshnetSuccess            = "Meshnet is set to '%s' successfully."
	MsgMeshnetAlreadyEnabled        = "Meshnet is already enabled."
	MsgMeshnetAlreadyDisabled       = "Meshnet is already disabled."
	MsgMeshnetNotEnabled            = "Meshnet is not enabled. Use the \"nordvpn set meshnet on\" command to enable it."
	MsgMeshnetNordlynxMustBeEnabled = "NordLynx technology must be set to use this feature."
	MsgMeshnetVersionNotSupported   = "Current application version does not support the Meshnet feature."
	MsgMeshnetUsage                 = "Meshnet is a way to safely access other devices, no matter where in the world they are. Once set up, Meshnet functions just like a secure local area network (LAN) — it connects devices directly. It also allows securely sending files to other devices. Use the \"nordvpn set meshnet on\" command to enable Meshnet. Learn more: https://meshnet.nordvpn.com/"

	MsgMeshnetRefreshUsage = "Refreshes the Meshnet in case it was not updated automatically."
	MsgMeshnetPeerUnknown  = "Peer '%s' is unknown."

	// Invites
	MsgMeshnetInviteUsage                     = "Add other users' devices to your Meshnet."
	MsgMeshnetInviteDescription               = MsgMeshnetInviteUsage + "\n" + "Learn more: https://meshnet.nordvpn.com/features/linking-devices-in-meshnet"
	MsgMeshnetInviteListUsage                 = "Displays the list of all sent and received Meshnet invitations."
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

	// Meshnet set commands group
	MsgMeshnetSetUsage = "Set a Meshnet configuration option."

	MsgMeshnetSetMachineNicknameUsage = "Sets a nickname for this machine within Meshnet."
	MsgMeshnetSetNicknameArgsUsage    = "<new_nickname>"
	MsgMeshnetSetNicknameSuccessful   = "The nickname for this machine is now set to '%s'."

	// Meshnet remove commands group
	MsgMeshnetRemoveUsage = "Remove a Meshnet configuration option."

	MsgMeshnetRemoveMachineNicknameUsage = "Removes the nickname currently set for this machine within Meshnet."
	MsgMeshnetRemoveNicknameSuccessful   = "The nickname for this machine has been removed."

	// Peers
	MsgMeshnetPeerListFilters = "Filters list of available peers in a Meshnet. To apply multiple filters, separate them with a comma. Please note that you will see an empty list if you apply contradictory filters."
	MsgMeshnetPeerUsage       = "Manage Meshnet peers."
	MsgMeshnetPeerDescription = `Manage your Meshnet devices.
Learn more:
	Managing Meshnet devices - https://meshnet.nordvpn.com/getting-started/how-to-start-using-meshnet/using-meshnet-on-linux#manage-devices
	Meshnet permissions explained - https://meshnet.nordvpn.com/features/explaining-permissions
	Routing traffic in Meshnet - https://meshnet.nordvpn.com/features/routing-traffic-in-meshnet`
	MsgMeshnetPeerArgsUsage     = "<peer_hostname>|<peer_nickname>|<peer_ip>|<peer_pubkey>"
	MsgMeshnetPeerListUsage     = "Lists available peers in a Meshnet."
	MsgMeshnetPeerRemoveUsage   = "Removes a peer from a Meshnet."
	MsgMeshnetPeerRemoveSuccess = "Peer '%s' has been removed from the Meshnet."

	MsgMeshnetPeerRoutingUsage          = "Allows/denies a peer device to route all traffic through this device."
	MsgMeshnetPeerRoutingDescription    = MsgMeshnetPeerRoutingUsage + "\n" + "Learn more: https://meshnet.nordvpn.com/features/explaining-permissions/traffic-routing-permissions"
	MsgMeshnetPeerRoutingAllowUsage     = "Allows a Meshnet peer to route its' traffic through this device."
	MsgMeshnetPeerRoutingDenyUsage      = "Denies a Meshnet peer to route its' traffic through this device."
	MsgMeshnetPeerRoutingAlreadyAllowed = "Traffic routing for '%s' is already allowed."
	MsgMeshnetPeerRoutingAlreadyDenied  = "Traffic routing for '%s' is already denied."
	MsgMeshnetPeerRoutingAllowSuccess   = "Traffic routing for '%s' has been allowed."
	MsgMeshnetPeerRoutingDenySuccess    = "Traffic routing for '%s' has been denied."

	MsgMeshnetPeerIncomingUsage          = "Allows/denies a peer device to access this device remotely (incoming connections)."
	MsgMeshnetPeerIncomingDescription    = MsgMeshnetPeerIncomingUsage + "\n" + "Learn more: https://meshnet.nordvpn.com/features/explaining-permissions/remote-access-permissions"
	MsgMeshnetPeerIncomingAllowUsage     = "Allows a Meshnet peer to send traffic to this device."
	MsgMeshnetPeerIncomingDenyUsage      = "Denies a Meshnet peer to send traffic to this device."
	MsgMeshnetPeerIncomingAlreadyAllowed = "Incoming traffic for '%s' is already allowed."
	MsgMeshnetPeerIncomingAlreadyDenied  = "Incoming traffic for '%s' is already denied."
	MsgMeshnetPeerIncomingAllowSuccess   = "Incoming traffic for '%s' has been allowed."
	MsgMeshnetPeerIncomingDenySuccess    = "Incoming traffic for '%s' has been denied."

	MsgMeshnetPeerLocalNetworkUsage          = "Allows/denies access to your local network when a peer device is routing traffic through this device."
	MsgMeshnetPeerLocalNetworkDescription    = MsgMeshnetPeerLocalNetworkUsage + "\n" + "Learn more: https://meshnet.nordvpn.com/features/explaining-permissions/local-network-permissions"
	MsgMeshnetPeerLocalNetworkAllowUsage     = "Allows a Meshnet peer to access local network when routing traffic through this device."
	MsgMeshnetPeerLocalNetworkDenyUsage      = "Denies a Meshnet peer to access local network when routing traffic through this device."
	MsgMeshnetPeerLocalNetworkAlreadyAllowed = "Local network access for '%s' is already allowed."
	MsgMeshnetPeerLocalNetworkAlreadyDenied  = "Local network access for '%s' is already denied."
	MsgMeshnetPeerLocalNetworkAllowSuccess   = "Local network access for '%s' has been allowed."
	MsgMeshnetPeerLocalNetworkDenySuccess    = "Local network access for '%s' has been denied."

	MsgMeshnetPeerFileshareUsage          = "Allows/denies peer to send files to this device."
	MsgMeshnetPeerFileshareDescription    = MsgMeshnetPeerFileshareUsage + "\n" + "Learn more: https://meshnet.nordvpn.com/features/explaining-permissions/file-sharing-permissions"
	MsgMeshnetPeerFileshareAllowUsage     = "Allows a Meshnet peer to send files to this device."
	MsgMeshnetPeerFileshareDenyUsage      = "Denies a Meshnet peer to send files to this device."
	MsgMeshnetPeerFileshareAlreadyAllowed = "Fileshare for '%s' is already allowed."
	MsgMeshnetPeerFileshareAlreadyDenied  = "Fileshare for '%s' is already denied."
	MsgMeshnetPeerFileshareAllowSuccess   = "Fileshare for '%s' has been allowed."
	MsgMeshnetPeerFileshareDenySuccess    = "Fileshare for '%s' has been denied."

	MsgMeshnetPeerAutomaticFileshareUsage              = "Always accept file transfers from a specific peer. We won’t ask you to approve each transfer - files will start downloading automatically."
	MsgMeshnetPeerAutomaticFileshareAllowUsage         = "Enables automatic fileshare from device."
	MsgMeshnetPeerAutomaticFileshareDenyUsage          = "Denies automatic fileshare from device."
	MsgMeshnetPeerAutomaticFileshareAlreadyEnabled     = "Automatic fileshare for '%s' is already allowed."
	MsgMeshnetPeerAutomaticFileshareAlreadyDisabled    = "Automatic fileshare for '%s' is already denied."
	MsgMeshnetPeerAutomaticFileshareEnableSuccess      = "Automatic fileshare for '%s' has been allowed."
	MsgMeshnetPeerAutomaticFileshareDisableSuccess     = "Automatic fileshare for '%s' has been denied."
	MsgMeshnetPeerAutomaticFileshareDefaultDirNotFound = "We couldn't enable auto-accept because the download directory doesn't exist."

	MsgMeshnetPeerConnectUsage        = "Treats a peer as a VPN server and connects to it if the peer has allowed traffic routing."
	MsgMeshnetPeerConnectSuccess      = "You are connected to Meshnet exit node '%s'."
	MsgMeshnetPeerDoesNotAllowRouting = "Meshnet peer '%s' does not allow traffic routing."
	MsgMeshnetPeerAlreadyConnected    = "You are already connected."
	MsgMeshnetPeerConnectFailed       = "Connect to other mesh peer failed - check if peer '%s' is online."

	MsgMeshnetPeerNicknameUsage           = "Sets/removes a peer device nickname within Meshnet."
	MsgMeshnetPeerSetNicknameUsage        = "Sets a nickname for the specified peer device."
	MsgMeshnetPeerSetNicknameArgsUsage    = "<peer_hostname>|<peer_nickname>|<peer_ip>|<peer_pubkey> <new_peer_nickname>"
	MsgMeshnetPeerRemoveNicknameUsage     = "Removes the nickname currently set for the specified peer device."
	MsgMeshnetPeerRemoveNicknameArgsUsage = "<peer_hostname>|<peer_nickname>|<peer_ip>|<peer_pubkey>"
	MsgMeshnetPeerSetNicknameSuccessful   = "The nickname for the peer '%s' is now set to '%s'."
	MsgMeshnetNicknameAlreadyEmpty        = "The nickname is already removed for this device."
	MsgMeshnetPeerResetNicknameSuccessful = "The nickname for the peer '%s' has been removed. The default hostname is '%s'."

	// errors received for meshnet nicknames
	MsgMeshnetSetSameNickname           = "The nickname '%s' is already set for this device."
	MsgMeshnetNicknameIsDomainName      = "The nickname is unavailable: A domain with this name already exists in your system."
	MsgMeshnetRateLimitReach            = "You've reached the weekly limit for nickname changes."
	MsgMeshnetNicknameTooLong           = "This nickname is too long. Nicknames can have up to 25 characters."
	MsgMeshnetDuplicateNickname         = "A device with this nickname already exists."
	MsgMeshnetContainsForbiddenWord     = "This nickname contains a restricted word."
	MsgMeshnetInvalidPrefixOrSuffix     = "This nickname contains a disallowed prefix or suffix."
	MsgMeshnetNicknameWithDoubleHyphens = "Nicknames can't contain double dashes ('--')."
	MsgMeshnetContainsInvalidChars      = "This nickname contains disallowed characters."

	// Fileshare
	FileshareName       = "fileshare"
	FileshareSendName   = "send"
	FileshareAcceptName = "accept"
	FileshareCancelName = "cancel"
	FileshareListName   = "list"
	FileshareClearName  = "clear"

	flagFileshareNoWait  = "background"
	flagFilesharePath    = "path"
	flagFileshareListIn  = "incoming"
	flagFileshareListOut = "outgoing"

	MsgFileshareUsage                     = "Transfer files of any size between Meshnet peers securely and privately"
	MsgFileshareDescription               = MsgFileshareUsage + "\n" + "Learn more: https://meshnet.nordvpn.com/features/sharing-files-in-meshnet\n\nNote: most arguments (peer name, transfer ID, file name) in fileshare commands can be entered faster using auto-completion. Simply press Tab and the app will suggest valid options for you."
	MsgFileshareTransferNotFound          = "Transfer not found."
	MsgFileshareInvalidPath               = "Invalid path provided: %s"
	MsgFilesharePathNotFound              = "Download directory %q does not exist. Make sure the directory exists or provide an alternative via --" + flagFilesharePath
	MsgFileshareAcceptPathIsASymlink      = "A download path can’t be a symbolic link. Please provide a directory as a download path to accept the transfer."
	MsgFileshareAcceptPathIsNotADirectory = "Please provide a directory as a download path to accept the transfer."
	MsgFileshareInvalidPeer               = "Peer name is invalid."
	MsgFileshareDisconnectedPeer          = "Peer is disconnected."
	MsgFileshareFileNotFound              = "File not found."
	MsgFileshareSocketNotFound            = "Enable Meshnet to share files. If Meshnet is already enabled, try disabling and enabling it again. Use \"nordvpn set meshnet on\" to enable it."
	MsgFileshareUserNotLoggedIn           = "You’re not logged in. To share files, please log in to NordVPN and ensure Meshnet is enabled."

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
	MsgNoPermissions                 = "You don’t have write permissions for the download directory %s. To receive the file transfer, choose another download directory using the --" + flagFilesharePath + " parameter."

	MsgFileshareSendUsage       = "Send files or directories to a Meshnet peer."
	MsgFileshareSendArgsUsage   = "<peer_hostname>|<peer_nickname>|<peer_ip>|<peer_pubkey> <path_1> [path_2...]"
	MsgFileshareSendDescription = MsgFileshareSendUsage + "\n\nTo cancel a transfer in progress, press Ctrl+C"
	MsgFileshareNoWaitUsage     = "Send a file transfer in the background instead of seeing its progress. It allows you to continue using the terminal for other commands while a transfer is in progress."
	MsgFileshareSendNoWait      = "File transfer %s has started in the background."
	MsgFileshareAcceptNoWait    = "File transfer has started in the background."
	MsgFileshareWaitAccept      = "Waiting for the peer to accept your transfer..."
	MsgTransferNotCreated       = "Can’t send the files. Please check if you have the \"read\" permission for the files you want to send."

	MsgFileshareListUsage       = "Lists transfers. If transfer ID is provided, lists files in the transfer."
	MsgFileshareListArgsUsage   = `[transfer_id]`
	MsgFileshareListDescription = `Adding no arguments to the command will list transfers.
Provide a [transfer_id] argument to list files in the specified transfer.`
	MsgFileshareListInUsage       = "Show only incoming transfers."
	MsgFileshareListOutUsage      = "Show only outgoing transfers."
	MsgFileshareCancelUsage       = "Cancel a transfer or a single file. To cancel an entire transfer, specify the transfer ID. To cancel a single file, specify the transfer ID and the file ID."
	MsgFileshareCancelArgsUsage   = "<transfer_id> [file_id]"
	MsgFileshareCancelSuccess     = "File transfer canceled."
	MsgFileshareAcceptUsage       = "Accept an incoming file transfer. To download an entire transfer, specify the transfer ID. To download a single file, specify the transfer ID and the file ID."
	MsgFileshareAcceptArgsUsage   = "<transfer_id> [file_id1] [file_id2...]"
	MsgFileshareAcceptDescription = MsgFileshareAcceptUsage + "\n\nTo cancel a transfer in progress, press Ctrl+C"
	MsgFileshareAcceptPathUsage   = "Specify download path (default: $XDG_DOWNLOAD_DIR or $HOME/Downloads)"
	MsgFileshareClearUsage        = "Clear entries older than the specified time period from the file transfer history."
	MsgFileshareClearArgsUsage    = "all|<time_period> [time_period...]"
	MsgFileshareClearDescription  = MsgFileshareClearUsage + "\n\nSpecify the time period using the systemd time span syntax: https://www.freedesktop.org/software/systemd/man/latest/systemd.time.html\n\nFor example, \"nordvpn fileshare clear 1d 12h\" clears entries older than 36 hours. Use \"nordvpn fileshare clear all\" to remove all entries."
	MsgFileshareClearSuccess      = "File transfer history cleared."
	MsgFileshareClearFailure      = "Can't clear file transfer history. See nordfileshared.log for more details."

	MsgFileshareProgressOngoing        = "File transfer [%s] progress [%d%%]"
	MsgFileshareProgressFinished       = "File transfer [%s] completed.      " // Need extra spaces to cover the progress message
	MsgFileshareProgressFinishedErrors = "File transfer [%s] completed. Some of the files have failed to transfer."
	MsgFileshareProgressCanceledByPeer = "File transfer [%s] canceled by peer."
	MsgFileshareProgressCanceled       = "File transfer [%s] canceled by other process."
	MsgFileshareStartedByOtherUser     = "A file sharing session is already in progress under another user account. To use the feature, restart Meshnet and enter your file sharing command again. "

	MsgNoSnapPermissions = "Permission needed. To ensure NordVPN runs smoothly, grant the necessary permissions for the snap using these commands:\n\n%s\n\nTo start using the app, log in to your Nord Account by entering nordvpn login."

	MsgNoSnapPermissionsExt = "Permission needed. To ensure NordVPN runs smoothly, grant the necessary permissions for the snap using these commands:\n\nsudo groupadd nordvpn\nsudo usermod -aG nordvpn $USER\n\n%s\n\nTo start using the app, log in to your Nord Account by entering nordvpn login."

	MsgSnapNoSocketPermissions = "Permission needed. To ensure NordVPN runs smoothly, grant the necessary permissions for the snap using these commands:\n\nsudo groupadd nordvpn\nsudo usermod -aG nordvpn $USER\n\nTo start using the app, log in to your Nord Account by entering nordvpn login."

	MsgNoSocketPermissions = "Permission denied. Please grant necessary permissions before using the application by executing the following commands:\n\nsudo groupadd nordvpn\nsudo usermod -aG nordvpn $USER\n\nAfter doing so, reboot your device afterwards for this to take an effect."

	MsgSnapPermissionsErrorForTray = "Please grant necessary permissions for the snap using this command:\n\n%s"

	MsgSetVirtualLocationUsageText   = "Enables or disables access to virtual locations. Virtual location servers let you access more locations worldwide."
	MsgSetVirtualLocationDescription = "Enables or disables access to virtual locations."
	MsgFooterVirtualLocationNote     = "* Virtual location servers"

	MsgShowListOfServers = "Shows a list of %s where servers are available.\n\nLocations marked with a different color in the list are virtual. Virtual location servers let you connect to more places worldwide. They run on dedicated physical servers, which are placed outside the intended location but configured to use its IP address."

	SetPqUnavailable       = "The post-quantum VPN is not compatible with OpenVPN. Switch to NordLynx to use post-quantum VPN capabilities."
	SetTechnologyDisablePQ = "This setting is not compatible with the post-quantum VPN. To use OpenVPN, disable the post-quantum VPN first."
	SetPqAndMeshnet        = "The post-quantum VPN and Meshnet can't run at the same time. Please disable one feature to use the other."
	SetPqUsageText         = "Enables or disables post-quantum VPN. When enabled, your connection uses cutting-edge cryptography designed to resist quantum computer attacks. Not compatible with Meshnet."
)
