// Package serverpicker contains the logic for selecting a VPN server that
// matches a set of user criteria (technology, protocol, location, group and
// tags). It resolves the requested target into a concrete server,
// preferring the recommendations API and falling back to the locally cached
// server list.
package serverpicker

import (
	"errors"
	"math/rand"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/log"
)

var ErrUnknownTechnology = errors.New("unknown technology")

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
	Tag            string // argument
	Group          string // flag
	ExcludedServer string
}

func NewSearchParams(tag, group string, excludedServer string) SearchParams {
	return SearchParams{
		Tag:            strings.TrimSpace(tag),
		Group:          strings.TrimSpace(group),
		ExcludedServer: strings.TrimSpace(excludedServer),
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
	search, err := newServerSearch(cfg, input, countries, servers)
	if err != nil {
		return ServerSelection{}, err
	}

	selectedServers, recommendationUUID, err := search.fetchFromAPI(api, insights)
	remote := len(selectedServers) > 0
	if !remote {
		log.ServerSel.Debug("falling back to locally cached servers, API error", err)
		selectedServers, err = search.findLocally(servers)
	}

	if err != nil {
		log.ServerSel.Error("failed to select server (from remote or locally cached)", err)
		return ServerSelection{}, err
	}

	if len(selectedServers) == 0 {
		log.ServerSel.Debug("no server found")
		return ServerSelection{}, internal.ErrServerIsUnavailable
	}

	return ServerSelection{
		Server:             chooseServer(selectedServers, remote),
		RecommendationUUID: recommendationUUID,
		Remote:             remote,
	}, nil
}

func chooseServer(servers []core.Server, remote bool) *core.Server {
	if remote {
		// servers from the API are ordered by recommendation, take the best one
		return &servers[0]
	}
	// #nosec G404 -- not used for cryptographic purposes
	return &servers[rand.Intn(len(servers))]
}
