// Package serverpicker contains the logic for selecting a VPN server that
// matches a set of user criteria (technology, protocol, obfuscation, location,
// group and tags). It resolves the requested target into a concrete server,
// preferring the recommendations API and falling back to the locally cached
// server list.
package serverpicker

import (
	"errors"
	"math/rand"
	"time"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/log"
	"golang.org/x/exp/slices"
)

const (
	apiServersLimit = 20
	logPrefix       = "[server_sel]"
)

// RecommendationUUID identifies a server recommendation returned by the API.
type RecommendationUUID string

// ServerSelection is the result of picking a server.
type ServerSelection struct {
	Server                *core.Server
	RecommendationUUID    RecommendationUUID
	Remote                bool
	DedicatedServerStatus core.DedicatedServerStatus
}

type SearchParams struct {
	Tag   string
	Group string
}

func NewSearchParams(tag, group string) SearchParams {
	return SearchParams{
		Tag:   tag,
		Group: group,
	}
}

// PickServer by the specified criteria.
func PickServer(
	api core.ServersAPI,
	servers core.Servers,
	countries core.Countries,
	insights core.Insights,
	cfg config.Config,
	input SearchParams,
) (ServerSelection, error) {
	var remote = true
	var recommendationUUID RecommendationUUID
	var selectedServer *core.Server

	selectedServers := []core.Server{}

	tech := cfg.Technology
	protocol := cfg.AutoConnectData.Protocol
	obfuscated := cfg.AutoConnectData.Obfuscate
	log.Debug(logPrefix, "search server", tech, protocol, obfuscated, "with input", input)

	serverTech := TechToServerTech(tech, protocol, obfuscated)
	if serverTech == core.Unknown {
		return ServerSelection{}, errors.New("unknown technology")
	}

	// detect the group from the input params
	serverGroup, err := resolveServerGroup(&input, obfuscated)
	log.Debug(logPrefix, "resolved server group", serverGroup)
	if err != nil {
		return ServerSelection{}, err
	}

	if serverGroup == config.ServerGroup_DEDICATED_IP {
		// DIP servers are selected from the user subscription services
		return ServerSelection{}, ErrDedicatedIPServer
	}

	if serverGroup == config.ServerGroup_DEDICATED_SERVER {
		// DS servers are taken from another API endpoint
		return ServerSelection{}, ErrDedicatedServer
	}

	// construct the servers list filters, for matching the current settings
	localSelFn := selectFilterForLocalServers(input.Tag, serverGroup, obfuscated)
	filterServersFn := func(s core.Server) bool {
		return MatchesUserSettings(s, cfg) &&
			// for local servers only, take into account also the server.Keys
			((len(s.Keys) == 0) || localSelFn(s))
	}

	// determine how the server search will be made
	serverTag, err := serverTagFromString(input.Tag, serverGroup, countries, servers)

	if err != nil {
		log.Debug(logPrefix, "unable to detect server tag", err)
		if errors.Is(err, internal.ErrTagDoesNotExist) {
			return ServerSelection{}, err
		}
		// for other errors, local servers will be used, so set it to unknown to have a valid value
		serverTag = core.ServerTag{Action: core.ServerByUnknown}
	} else {
		// fetch from the API only if serverTag is valid
		apiFilter := core.ServersFilter{
			Group: serverGroup,
			Tech:  serverTech,
			Tag:   serverTag,
			Limit: apiServersLimit,
		}
		selectedServers, recommendationUUID, err = fetchServersFromAPI(api, insights, serverTag, apiFilter, filterServersFn)
	}

	if len(selectedServers) == 0 {
		// if no servers were received from the API, try from locally cached servers
		log.Error(logPrefix, "failed to select server from remote", err)
		remote = false
		selectedServers, err = findServersLocally(servers, serverTag, filterServersFn)
	}

	if err != nil {
		return ServerSelection{}, err
	}

	if len(selectedServers) == 0 {
		log.Debug(logPrefix, "no server found")
		// We were not guarded against this case before
		// So I assume it should not happen, but better be safe
		return ServerSelection{}, internal.ErrServerIsUnavailable
	}

	allowVirtualServer := cfg.VirtualLocation.Get()
	if !allowVirtualServer && len(selectedServers) > 0 {
		selectedServers = slices.DeleteFunc(selectedServers, func(s core.Server) bool { return s.IsVirtualLocation() })
		if len(selectedServers) == 0 {
			// if the selected servers are only virtual, but user has this disabled return an error
			return ServerSelection{}, internal.ErrVirtualServerSelected
		}
	}

	// #nosec G404 -- not used for cryptographic purposes
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	selectedServer = &selectedServers[rng.Int63n(int64(len(selectedServers)))]

	return ServerSelection{
		Server:             selectedServer,
		RecommendationUUID: recommendationUUID,
		Remote:             remote,
	}, nil
}

// fetchServersFromAPI - selects servers from the remote API that match the given
// serverTag and pass the filterServersFn.
func fetchServersFromAPI(
	api core.ServersAPI,
	insights core.Insights,
	serverTag core.ServerTag,
	filter core.ServersFilter,
	filterServersFn func(server core.Server) bool,
) ([]core.Server, RecommendationUUID, error) {
	if serverTag.Action == core.ServerByName {
		// the search is made after a specific server name, e.g. lt1234
		servers, err := getServerByNameFromRemote(api, serverTag, filterServersFn)
		return servers, emptyUUID, err
	}

	return getRecommendedServers(api, insights.Longitude, insights.Latitude, filter, filterServersFn)
}

// findServersLocally selects servers from the locally cached list that pass
// filterServersFn, used as a fallback when the API returns no servers.
func findServersLocally(
	servers core.Servers,
	serverTag core.ServerTag,
	filterServersFn func(server core.Server) bool,
) ([]core.Server, error) {
	log.Debug(logPrefix, "search locally the server", serverTag)

	selectedServers := internal.Filter(servers, filterServersFn)
	if len(selectedServers) == 0 {
		return nil, internal.ErrServerIsUnavailable
	}

	log.Debug(logPrefix, "found local servers", selectedServers)

	// remove all DIP servers from the list if the search wasn't made for a server name and there is more than 1 server found
	if serverTag.Action != core.ServerByName && len(selectedServers) > 1 {
		log.Debug(logPrefix, "removing DIP servers")
		selectedServers = slices.DeleteFunc(selectedServers, func(s core.Server) bool { return core.IsDedicatedIP(s) })
	}

	return selectedServers, nil
}
