package serverpicker

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/log"
	"github.com/google/uuid"
)

const (
	recommendationUUIDHeader = "X-Recommendation-Uuid"
	emptyUUID                = ""
)

func getServerByNameFromRemote(
	api core.ServersAPI,
	serverTag core.ServerTag,
	filterFn func(server core.Server) bool,
) ([]core.Server, error) {
	if serverTag.Action != core.ServerByName {
		return nil, errors.New("search must be made after a specific server")
	}
	server, err := api.Server(serverTag.ID)
	if err != nil {
		return nil, err
	}

	log.Debug(logPrefix, "server received from the API", server.Hostname)

	filteredServers := internal.Filter(core.Servers{*server}, filterFn)
	if len(filteredServers) == 0 {
		log.Warn(logPrefix, "server is not matching the user settings")
		return nil, internal.ErrServerIsUnavailable
	}
	return filteredServers, nil
}

func getRecommendedServers(
	api core.ServersAPI,
	longitude float64,
	latitude float64,
	apiFilter core.ServersFilter,
	filterFn func(server core.Server) bool,
) ([]core.Server, RecommendationUUID, error) {
	servers, header, err := api.RecommendedServers(apiFilter, longitude, latitude)
	if err != nil {
		return nil, emptyUUID, err
	}

	servers = internal.Filter(servers, filterFn)

	if len(servers) == 0 {
		return nil, emptyUUID, internal.ErrServerIsUnavailable
	}

	recommendationUUID, err := extractRecommendationUUID(header)
	if err != nil {
		log.Warn(logPrefix, "failed to extract recommendation UUID from the HTTP header", err)
		return servers, emptyUUID, nil
	}
	return servers, recommendationUUID, nil
}

func extractRecommendationUUID(h http.Header) (RecommendationUUID, error) {
	raw := h.Get(recommendationUUIDHeader)
	if raw == "" {
		return emptyUUID, fmt.Errorf("missing the recommendation UUID header")
	}

	u, err := uuid.Parse(raw)
	if err != nil {
		return emptyUUID, err
	}

	return RecommendationUUID(u.String()), nil
}
