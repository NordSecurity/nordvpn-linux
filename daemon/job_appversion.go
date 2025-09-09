package daemon

import (
	"log"
	"sync"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/daemon/state"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/coreos/go-semver/semver"
)

func JobVersionCheck(dm *DataManager, api *RepoAPI, statePublisher *state.StatePublisher) func() {
	return func() {
		// Always publish version status at the end
		defer publishVersionHealthStatus(dm, statePublisher)

		// if no currentVersion data is available, 0.0.0-0 currentVersion will be used.
		currentVersion := semver.New(api.version)
		vdata := dm.GetVersionData()

		if currentVersion.LessThan(vdata.version) {
			dm.SetVersionData(vdata.version, true)
		}

		// if vdata.version.Major == 0 that means the data is incorrect, possibly the file is missing
		if vdata.newerVersionAvailable && vdata.version.Major != 0 {
			publishVersionHealthStatus(dm, statePublisher)
			return
		}

		// Get info from repo and convert data to strings
		var versionStrings []string
		switch api.packageType {
		case "deb":
			data, err := api.DebianFileList()
			if err != nil {
				dm.SetVersionData(vdata.version, false)
				publishVersionHealthStatus(dm, statePublisher)
				return
			}
			versionStrings = ParseDebianVersions(data)
		default: // use this logic for RPM and unexpected cases
			data, err := api.RpmFileList()
			if err != nil {
				dm.SetVersionData(vdata.version, false)
				publishVersionHealthStatus(dm, statePublisher)
				return
			}
			versionStrings = ParseRpmVersions(data)
		}
		// Convert currentVersion strings to a format that's easier to compare
		versions := StringsToVersions(versionStrings)
		latestVersion := GetLatestVersion(versions)

		newerVersionAvailable := currentVersion.LessThan(latestVersion)
		if newerVersionAvailable || vdata.version.Major == 0 {
			dm.SetVersionData(latestVersion, newerVersionAvailable)
		}
		publishVersionHealthStatus(dm, statePublisher)
	}
}

var (
	lastHealthStatusCode int32 = int32(internal.CodeSuccess)
	versionHealthMutex   sync.Mutex
)

// publishVersionHealthStatus publishes version health status for version updates
func publishVersionHealthStatus(dm *DataManager, statePublisher *state.StatePublisher) {
	versionHealthMutex.Lock()
	defer versionHealthMutex.Unlock()

	versionData := dm.GetVersionData()

	var healthStatusCode int32 = int32(internal.CodeSuccess)
	if versionData.newerVersionAvailable {
		healthStatusCode = int32(internal.CodeOutdated)
	}

	// Only publish if health status changed
	if lastHealthStatusCode != healthStatusCode {
		healthStatus := &pb.VersionHealthStatus{
			StatusCode: healthStatusCode,
		}

		if err := statePublisher.NotifyVersionHealth(healthStatus); err != nil {
			log.Printf("%s Failed to publish version health status: %v\n", internal.ErrorPrefix, err)
		}

		lastHealthStatusCode = healthStatusCode
	}
}

var (
	lastHealthStatusCode int32 = int32(internal.CodeSuccess)
	versionHealthMutex   sync.Mutex
)

// publishVersionHealthStatus publishes version health status for version updates
func publishVersionHealthStatus(dm *DataManager, statePublisher *state.StatePublisher) {
	versionHealthMutex.Lock()
	defer versionHealthMutex.Unlock()

	versionData := dm.GetVersionData()

	var healthStatusCode int32 = int32(internal.CodeSuccess)
	if versionData.newerVersionAvailable {
		healthStatusCode = int32(internal.CodeOutdated)
	}

	// Only publish if health status changed
	if lastHealthStatusCode != healthStatusCode {
		healthStatus := &pb.VersionHealthStatus{
			StatusCode: healthStatusCode,
		}

		if err := statePublisher.NotifyVersionHealth(healthStatus); err != nil {
			log.Printf("%s Failed to publish version health status: %v\n", internal.ErrorPrefix, err)
		}

		lastHealthStatusCode = healthStatusCode
	}
}
