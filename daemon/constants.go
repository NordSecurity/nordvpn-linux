package daemon

import (
	"path/filepath"

	"github.com/NordSecurity/nordvpn-linux/internal"
)

const (
	// app should not use trailing slashes and prefix endpoints
	// with slash instead. e.g. /v1/example

	// BaseURL defines the base uri for the api
	BaseURL = "https://api.nordvpn.com"

	// RepoURL is the url for NordVPN repository
	RepoURL = "https://repo.nordvpn.com"

	// RandomComponentMin defines minimal value of random component
	RandomComponentMin = 0

	// RandomComponentMax defines maximum value of random component
	RandomComponentMax = 0.001

	// Daemon gRPC API current version
	DaemonApiVersion uint32 = 1
)

var (
	// ServersDataFilePath defines path to servers data file
	ServersDataFilePath = filepath.Join(internal.DatFilesPathCommon, "servers.dat")

	// CountryDataFilePath defines path to countries data file
	CountryDataFilePath = filepath.Join(internal.DatFilesPathCommon, "countries.dat")

	// InsightsFilePath defines filename of insights file
	InsightsFilePath = filepath.Join(internal.DatFilesPath, "insights.dat")

	// VersionFilePath defines filename of latest available version file
	VersionFilePath = filepath.Join(internal.DatFilesPathCommon, "version.dat")

	// IconPath defines icon file path
	IconPath = internal.PrefixCommonPath("/usr/share/icons/hicolor/scalable/apps/nordvpn.svg")
)
