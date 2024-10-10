package internal

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
)

const (
	// ListenPID defines process id env key
	ListenPID = "LISTEN_PID"

	// ListenFDS defines systemDFile descriptors env key
	ListenFDS = "LISTEN_FDS"

	// ListenFDNames defines systemDFile descriptors names env key
	ListenFDNames = "LISTEN_FDNAMES"

	// Proto defines protocol to be used
	Proto = "unix"

	// TmpDir defines temporary storage directory
	TmpDir = "/tmp/"

	// NordvpnGroup that can access daemon socket
	NordvpnGroup = "nordvpn"

	// PermUserRWX user permission type to read write and execute
	PermUserRWX = 0700

	// PermUserRW user permission type to read and write
	PermUserRW = 0600

	// PermUserRWGroupRW permission type for user and group to read and write, everyone else - no access.
	PermUserRWGroupRW = 0660

	// PermUserRWGroupROthersR user permission type for user to read and write to it, everyone else can only read it.
	PermUserRWGroupROthersR = 0644

	// PermUserRWGroupROthersR allows user and group to read and write, other only read
	PermUserRWGroupRWOthersR = 0664

	// PermUserRWGroupROthersR user permission type for everyone to read and write to it.
	PermUserRWGroupRWOthersRW = 0666

	// PermUserRWXGroupRXOthersRX forbidding group and others to write to it
	PermUserRWXGroupRXOthersRX = 0755

	// ChattrExec is the chattr command executable name
	ChattrExec = "chattr"

	// LsattrExec is the lsattr command executable name
	LsattrExec = "lsattr"

	// SttyExec is a tool to change or print CLI settings
	SttyExec = "stty"

	// SystemctlExec defines system controller executable
	SystemctlExec = "systemctl"

	// NetworkctlExec defines network controller executable
	NetworkctlExec = "networkctl"

	// ServerDateFormat defines api date format
	ServerDateFormat = "2006-01-02 15:04:05"

	// Fileshared defines filesharing process name
	Fileshare = "nordfileshare"

	Norduserd = "norduserd"

	NorduserdLogFileName = "norduserd" + LogFileExtension

	// FileshareHistoryFile is the storage file used by libdrop
	FileshareHistoryFileName = "fileshare_history.db"

	FileshareSocket = TmpDir + "fileshare.sock"

	FileshareLogFileName = "nordfileshare" + LogFileExtension

	LogFileExtension = ".log"

	SecondsToWaitForIptablesLock = "5" // use string to easier concatenate into the command args

	// MFAEnabledStatusName is a string value used to denote enabled multifactor authentication in a response to the
	// /v1/users/oauth/mfa/status by the core API
	MFAEnabledStatusName = "enabled"
	// AllowlistMinPort defines min port which can be allowlisted
	AllowlistMinPort = 1
	// AllowlistMaxPort defines max port which can be allowlisted
	AllowlistMaxPort = 65535
)

var (
	PlatformSupportsIPv4 = true
	PlatformSupportsIPv6 = true
)

var (
	// RunDir defines default socket directory
	RunDir = PrefixCommonPath("/run/nordvpn")

	// LogPath defines where logs are located if systemd isn't used
	LogPath = PrefixDataPath("/var/log/nordvpn")

	// AppDataPath defines path where app data is stored
	AppDataPath = PrefixDataPath("/var/lib/nordvpn")

	// AppDataPathCommon defines path where common app data files are stored. These files may
	// be removed after every app update
	AppDataPathCommon = PrefixCommonPath("/var/lib/nordvpn")

	// AppDataPathStatic defines path where static app data (such as helper executables) are stored
	AppDataPathStatic = PrefixStaticPath("/usr/lib/nordvpn")

	DatFilesPath = filepath.Join(AppDataPath, "data")

	DatFilesPathCommon = filepath.Join(AppDataPathCommon, "data")

	BakFilesPath = filepath.Join(AppDataPath, "backup")

	// OvpnTemplatePath defines filename of ovpn template file
	OvpnTemplatePath = filepath.Join(DatFilesPathCommon, "ovpn_template.xslt")

	// OvpnObfsTemplatePath defines filename of ovpn obfuscated template file
	OvpnObfsTemplatePath = filepath.Join(DatFilesPathCommon, "ovpn_xor_template.xslt")

	// DaemonSocket defines system daemon socket file location
	DaemonSocket = filepath.Join(RunDir, "/nordvpnd.sock")

	// DaemonPid defines daemon PID file location
	DaemonPid = filepath.Join(RunDir, "/nordvpnd.pid")

	FileshareBinaryPath = filepath.Join(AppDataPathStatic, Fileshare)

	NorduserdBinaryPath = filepath.Join(AppDataPathStatic, Norduserd)
)

func GetSupportedIPTables() []string {
	var iptables []string
	if PlatformSupportsIPv4 {
		iptables = append(iptables, "iptables")
	}
	if PlatformSupportsIPv6 {
		iptables = append(iptables, "ip6tables")
	}
	return iptables
}

func GetNorduserSocketSnap(uid int) string {
	return fmt.Sprintf("%s%d-%s.sock", TmpDir, uid, Norduserd)
}

// GetNorduserdSocket to communicate with norduser daemon
func GetNorduserdSocket(uid int) string {
	if uid == 0 {
		return fmt.Sprintf("/run/%s/%s.sock", Norduserd, Norduserd)
	}
	return fmt.Sprintf("/run/user/%d/%s/%s.sock", uid, Norduserd, Norduserd)
}

func GetNorduserSocketFork(uid int) string {
	return fmt.Sprintf("/tmp/%d-%s.sock", uid, Norduserd)
}

func getHomeDirPath(homeDirectory string) (string, error) {
	snapUserDataDir := os.Getenv("SNAP_USER_COMMON")
	if snapUserDataDir != "" {
		homeDirectory = snapUserDataDir
	}

	_, err := os.Stat(homeDirectory)
	if homeDirectory == "" || os.IsNotExist(err) {
		return "", errors.New("user does not have a home directory")
	}

	return homeDirectory, nil
}

// GetConfigDirPath returns the directory used to store local user config
func GetConfigDirPath(homeDirectory string) (string, error) {
	homeDirectory, err := getHomeDirPath(homeDirectory)

	if err != nil {
		return "", err
	}

	userConfigPath := filepath.Join(homeDirectory, ".config", "nordvpn")

	if err := EnsureDirFull(userConfigPath); err != nil {
		return "", fmt.Errorf("ensuring config dir: %w", err)
	}
	return userConfigPath, nil
}

// GetCacheDirPath returns the directory used to store local user logs
func GetCacheDirPath(homeDirectory string) (string, error) {
	homeDirectory, err := getHomeDirPath(homeDirectory)

	if err != nil {
		return "", err
	}

	userCachePath := filepath.Join(homeDirectory, ".cache", "nordvpn")

	if err := EnsureDirFull(userCachePath); err != nil {
		return "", fmt.Errorf("ensuring cache dir: %w", err)
	}
	return userCachePath, nil
}

// GetNordvpnGid returns id of group defined in NordvpnGroup
func GetNordvpnGid() (int, error) {
	group, err := user.LookupGroup(NordvpnGroup)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(group.Gid)
}
