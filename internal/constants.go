package internal

import (
	"errors"
	"fmt"
	"log"
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

	// TempDir defines temporary storage directory
	TempDir = "/tmp/"

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

	// Column is a tool to format data into columns for neater display in CLI
	ColumnExec = "column"

	// SttyExec is a tool to change or print CLI settings
	SttyExec = "stty"

	// SystemctlExec defines system controller executable
	SystemctlExec = "systemctl"

	// NetworkctlExec defines network controller executable
	NetworkctlExec = "networkctl"

	// ServerDateFormat defines api date format
	ServerDateFormat = "2006-01-02 15:04:05"

	// Fileshared defines filesharing daemon name
	Fileshared = "nordfileshared"

	Norduserd = "norduserd"

	// FileshareHistoryFile is the storage file used by libdrop
	FileshareHistoryFile = "fileshare_history.db"

	LogFileExtension = ".log"
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

	// AppDataPathStatic defines path where static app data (such as helper executables) are
	// stored. Normally it is the same as AppDataPath
	AppDataPathStatic = PrefixStaticPath("/var/lib/nordvpn")

	DatFilesPath = filepath.Join(AppDataPath, "data")

	DatFilesPathCommon = filepath.Join(AppDataPathCommon, "data")

	BakFilesPath = filepath.Join(AppDataPath, "backup")

	// OvpnTemplatePath defines filename of ovpn template file
	OvpnTemplatePath = filepath.Join(DatFilesPathCommon, "ovpn_template.xslt")

	// OvpnObfsTemplatePath defines filename of ovpn obfuscated template file
	OvpnObfsTemplatePath = filepath.Join(DatFilesPathCommon, "ovpn_xor_template.xslt")

	// ConfigDirectory is used for configuration files storage. Hardcoded only for nordfileshared, in
	// other cases consider using os.UserConfigDir instead.
	ConfigDirectory = filepath.Join(".config", "nordvpn")

	// LogFilePath defines CLI log path
	LogFilePath = filepath.Join("nordvpn", "cli.log")

	// DaemonSocket defines system daemon socket file location
	DaemonSocket = filepath.Join(RunDir, "/nordvpnd.sock")
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

// GetFilesharedSocket to communicate with fileshare daemon
func GetFilesharedSocket(uid int) string {
	_, err := os.Stat(fmt.Sprintf("/run/user/%d", uid))
	if uid == 0 || os.IsNotExist(err) {
		return fmt.Sprintf("/run/%s/%s.sock", Fileshared, Fileshared)
	}
	return fmt.Sprintf("/run/user/%d/%s/%s.sock", uid, Fileshared, Fileshared)
}

// GetNorduserdSocket to communicate with norduser daemon
func GetNorduserdSocket(uid int) string {
	_, err := os.Stat(fmt.Sprintf("/run/user/%d", uid))
	if uid == 0 || os.IsNotExist(err) {
		return fmt.Sprintf("/run/%s/%s.sock", Norduserd, Norduserd)
	}
	return fmt.Sprintf("/run/user/%d/%s/%s.sock", uid, Norduserd, Norduserd)
}

// GetConfigDirPath returns the directory used to store local user config and logs
func GetConfigDirPath(homeDirectory string) (string, error) {
	if homeDirectory == "" {
		return "", errors.New("user does not have a home directory")
	}
	// We are running as root, so we cannot retrieve user config directory path dynamically. We
	// hardcode it to /home/<username>/.config, and if it doesn't exist on the expected path
	// (i.e XDG_CONFIG_HOME is set), we default to /var/log/nordvpn/nordfileshared-<username>-<uid>.log
	userConfigPath := filepath.Join(homeDirectory, ConfigDirectory)
	_, err := os.Stat(userConfigPath)
	if err == nil {
		return userConfigPath, nil
	}
	return "", fmt.Errorf("%s directory not found in users home directory", ConfigDirectory)
}

// GetFilesharedLogPath when logs aren't handled by systemd
func GetFilesharedLogPath(uid string) string {
	filesharedLogFilename := Fileshared + LogFileExtension
	if uid == "0" {
		return filepath.Join(LogPath, filesharedLogFilename)
	}

	usr, err := user.LookupId(uid)
	if err != nil {
		log.Printf("failed to lookup user, users fileshared logs will be stored in %s: %s", LogPath, err.Error())
	}

	configDir, err := GetConfigDirPath(usr.HomeDir)

	if err != nil {
		log.Printf("users fileshared logs will be stored in %s: %s", LogPath, err.Error())
		return filepath.Join(LogPath, Fileshared+"-"+uid+LogFileExtension)
	}

	return filepath.Join(configDir, filesharedLogFilename)
}

// GetNorduserdLogPath when logs aren't handled by systemd
func GetNorduserdLogPath(uid string) string {
	norduserdLogFilename := Norduserd + LogFileExtension
	if uid == "0" {
		return filepath.Join(LogPath, norduserdLogFilename)
	}

	usr, err := user.LookupId(uid)
	if err != nil {
		log.Printf("failed to lookup user, users norduser logs will be stored in %s: %s", LogPath, err.Error())
	}

	configDir, err := GetConfigDirPath(usr.HomeDir)

	if err != nil {
		log.Printf("users norduserd logs will be stored in %s: %s", LogPath, err.Error())
		return filepath.Join(LogPath, Norduserd+"-"+uid+LogFileExtension)
	}

	return filepath.Join(configDir, norduserdLogFilename)
}

// GetNordvpnGid returns id of group defined in NordvpnGroup
func GetNordvpnGid() (int, error) {
	group, err := user.LookupGroup(NordvpnGroup)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(group.Gid)
}
