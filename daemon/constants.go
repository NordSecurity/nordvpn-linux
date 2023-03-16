package daemon

import "github.com/NordSecurity/nordvpn-linux/internal"

const (
	// app should not use trailing slashes and prefix endpoints
	// with slash instead. e.g. /v1/example

	// BaseURL defines the base uri for the api
	BaseURL = "https://api.nordvpn.com"

	// RepoURL is the url for NordVPN repository
	RepoURL = "https://repo.nordvpn.com"

	// IconPath defines icon file path
	IconPath = internal.AppDataPath + "icon.svg"

	// ServersDataFilePath defines path to servers data file
	ServersDataFilePath = internal.DatFilesPath + "servers.dat"

	// CountryDataFilePath defines path to countries data file
	CountryDataFilePath = internal.DatFilesPath + "countries.dat"

	// InsightsFilePath defines filename of insights file
	InsightsFilePath = internal.DatFilesPath + "insights.dat"

	// VersionFilePath defines filename of latest available version file
	VersionFilePath = internal.DatFilesPath + "version.dat"

	// RandomComponentMin defines minimal value of random component
	RandomComponentMin = 0

	// RandomComponentMin defines maximum value of random component
	RandomComponentMax = 0.001
)
