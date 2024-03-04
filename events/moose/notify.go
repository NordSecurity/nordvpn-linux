//go:build moose

package moose

import (
	"fmt"

	moose "moose/events"
)

type notifyRequest func(
	eventDuration int32,
	eventStatus moose.NordvpnappEventStatus,
	eventTrigger moose.NordvpnappEventTrigger,
	apiHostName string,
	responseCode int32,
	transferProtocol string,
	dnsResolutionTime int32,
	requestFilters string,
	requestFields string,
	limits string,
	offset string,
	responseSummary string,
) uint32

func noSuchEndpoint(
	eventDuration int32,
	eventStatus moose.NordvpnappEventStatus,
	eventTrigger moose.NordvpnappEventTrigger,
	apiHostName string,
	responseCode int32,
	transferProtocol string,
	dnsResolutionTime int32,
	requestFilters string,
	requestFields string,
	limits string,
	offset string,
	responseSummary string,
) uint32 {
	return 0
}

func pickNotifier(endpoint string) (notifyRequest, error) {
	switch endpoint {
	case "/v1/servers":
		return moose.NordvpnappSendServiceQualityApiRequestRequestServers, nil
	case "/v1/servers/recommendations":
		return moose.NordvpnappSendServiceQualityApiRequestRequestServersRecommendations, nil
	case "/v1/users/current":
		return moose.NordvpnappSendServiceQualityApiRequestRequestCurrentUser, nil
	case "/v1/users/services":
		return moose.NordvpnappSendServiceQualityApiRequestRequestUserServices, nil
	case "/v1/users/services/credentials":
		return moose.NordvpnappSendServiceQualityApiRequestRequestServiceCredentials, nil
	case "/v1/users/tokens":
		return moose.NordvpnappSendServiceQualityApiRequestRequestTokenCreation, nil
	case "/v1/users/tokens/renew":
		return moose.NordvpnappSendServiceQualityApiRequestRequestTokenRenew, nil
	default:
		return noSuchEndpoint, fmt.Errorf("%s is not important to moose", endpoint)
	}
}
