//go:build moose

package moose

import (
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
	debugJson *string,
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
	debugJson *string,
) uint32 {
	return 0
}

func pickNotifier(endpoint string) notifyRequest {
	switch endpoint {
	case "/v1/servers":
		return moose.NordvpnappSendServiceQualityApiRequestRequestServers
	case "/v1/servers/recommendations":
		return moose.NordvpnappSendServiceQualityApiRequestRequestServersRecommendations
	case "/v1/users/current":
		return moose.NordvpnappSendServiceQualityApiRequestRequestCurrentUser
	case "/v1/users/services":
		return moose.NordvpnappSendServiceQualityApiRequestRequestUserServices
	case "/v1/users/services/credentials":
		return moose.NordvpnappSendServiceQualityApiRequestRequestServiceCredentials
	case "/v1/users/tokens":
		return moose.NordvpnappSendServiceQualityApiRequestRequestTokenCreation
	case "/v1/users/tokens/renew":
		return moose.NordvpnappSendServiceQualityApiRequestRequestTokenRenew
	default:
		return noSuchEndpoint
	}
}
