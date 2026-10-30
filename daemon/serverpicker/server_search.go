package serverpicker

import (
	"errors"
	"slices"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/log"
)

const (
	apiServersLimit      = 10
	localServersMinCount = 20
)

type serverSearch struct {
	cfg        config.Config
	input      SearchParams
	group      config.ServerGroup
	tech       core.ServerTechnology
	tag        core.ServerTag
	tagFound   bool
	localSelFn core.Predicate
}

func newServerSearch(
	cfg config.Config,
	input SearchParams,
	countries core.Countries,
	servers core.Servers,
) (*serverSearch, error) {
	proto := cfg.VPNProtocol
	log.ServerSel.Debug("search server", proto, "with input", input)

	serverTech := VPNProtocolToServerTech(proto)
	if serverTech == core.Unknown {
		return nil, ErrUnknownTechnology
	}

	// detect the group from the input params (fix tag arg if needed)
	input, requestedGroup, err := resolveServerGroup(input)
	if err != nil {
		return nil, err
	}

	// the group the servers are actually looked up by can differ from the requested ones
	// on specific protocol (edge case)
	serverGroup := searchGroup(requestedGroup, proto)
	log.ServerSel.Debug("resolved server group", requestedGroup, "search group", serverGroup)

	//exhaustive:ignore - only groups not served by the server search are relevant
	switch serverGroup {
	case config.ServerGroup_DEDICATED_IP:
		// DIP servers are selected from the user subscription services
		return nil, ErrDedicatedIPServer
	case config.ServerGroup_DEDICATED_SERVER:
		// DS servers are taken from another API endpoint
		return nil, ErrDedicatedServer
	}

	// determine how the server search will be made
	serverTag, err := serverTagFromString(input.Tag, serverGroup, countries, servers)
	tagFound := err == nil
	if !tagFound {
		log.ServerSel.Debug("unable to detect server tag:", err)
		if errors.Is(err, internal.ErrTagDoesNotExist) {
			return nil, err
		}
		// for other errors, local servers will be used, so set it to unknown to have a valid value
		serverTag = core.ServerTag{Action: core.ServerByUnknown}
	}

	return &serverSearch{
		cfg:        cfg,
		input:      input,
		group:      serverGroup,
		tech:       serverTech,
		tag:        serverTag,
		tagFound:   tagFound,
		localSelFn: selectFilterForLocalServers(input.Tag, serverGroup),
	}, nil
}

func (s *serverSearch) fetchFromAPI(
	api core.ServersAPI,
	insights core.Insights,
) ([]core.Server, RecommendationUUID, error) {
	if !s.tagFound {
		return nil, emptyUUID, nil
	}

	if s.tag.Action == core.ServerByName {
		// the search is made after a specific server name, e.g. lt1234
		servers, err := getServerByNameFromRemote(api, s.tag, s.matches)
		return servers, emptyUUID, err
	}

	apiFilter := core.ServersFilter{
		Group: s.group,
		Tech:  s.tech,
		Tag:   s.tag,
		Limit: apiServersLimit,
	}
	return getRecommendedServers(api, insights.Longitude, insights.Latitude, apiFilter, s.matches)
}

func (s *serverSearch) findLocally(servers core.Servers) ([]core.Server, error) {
	log.ServerSel.Debug("search locally the server", s.tag)

	selectedServers := internal.Filter(servers, s.matches)
	if len(selectedServers) == 0 {
		return nil, internal.ErrServerIsUnavailable
	}

	log.ServerSel.Debug("found local servers", selectedServers)

	// remove all DIP servers from the list if the search wasn't made for a server name and there is more than 1 server found
	if s.tag.Action != core.ServerByName && len(selectedServers) > 1 {
		log.ServerSel.Debug("removing DIP servers")
		selectedServers = slices.DeleteFunc(selectedServers, core.IsDedicatedIP)
	}

	// prefer physical servers, fill the list up to localServersMinCount with virtual ones only if needed
	var physicalServers, virtualServers []core.Server
	for _, server := range selectedServers {
		if server.IsVirtualLocation() {
			virtualServers = append(virtualServers, server)
		} else {
			physicalServers = append(physicalServers, server)
		}
	}
	missingCount := max(localServersMinCount-len(physicalServers), 0)
	selectedServers = append(physicalServers, virtualServers[:min(missingCount, len(virtualServers))]...)

	return selectedServers, nil
}

func (s *serverSearch) matches(server core.Server) bool {
	return MatchesUserSettings(server, s.cfg) &&
		server.Hostname != s.input.ExcludedServer &&
		// servers without Keys (e.g. received from the API) skip the tag/group keys check
		((len(server.Keys) == 0) || s.localSelFn(server))
}
