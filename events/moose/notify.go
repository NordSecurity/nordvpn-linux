//go:build moose

package moose

import (
	"fmt"

	events "moose/events"
)

type notifyRequest func(
	hostname string,
	dnsResolutionTime int,
	eventDuration int,
	eventStatus events.Enum_SS_NordvpnappEventStatus,
	eventTrigger events.Enum_SS_NordvpnappEventTrigger,
	requestLimits string,
	requestOffset string,
	requestFields string,
	requestFilters string,
	responseCode int,
	responseSummary string,
	transferProtocol string,
) uint

func noSuchEndpoint(
	string,
	int,
	int,
	events.Enum_SS_NordvpnappEventStatus,
	events.Enum_SS_NordvpnappEventTrigger,
	string,
	string,
	string,
	string,
	int,
	string,
	string,
) uint {
	return 0
}

func pickNotifier(endpoint string) (notifyRequest, error) {
	switch endpoint {
	case "/v1/servers":
		return events.Send_serviceQuality_apiRequest_requestServers, nil
	case "/v1/servers/recommendations":
		return events.Send_serviceQuality_apiRequest_requestServersRecommendations, nil
	case "/v1/users/current":
		return events.Send_serviceQuality_apiRequest_requestCurrentUser, nil
	case "/v1/users/services":
		return events.Send_serviceQuality_apiRequest_requestUserServices, nil
	case "/v1/users/services/credentials":
		return events.Send_serviceQuality_apiRequest_requestServiceCredentials, nil
	case "/v1/users/tokens":
		return events.Send_serviceQuality_apiRequest_requestTokenCreation, nil
	case "/v1/users/tokens/renew":
		return events.Send_serviceQuality_apiRequest_requestTokenRenew, nil
	default:
		return noSuchEndpoint, fmt.Errorf("%s is not important to moose", endpoint)
	}
}
