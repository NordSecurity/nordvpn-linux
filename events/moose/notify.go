//go:build moose

package moose

import (
	moose "moose/events"
)

type notifyRequest func(
	eventParams moose.EventParams,
	apiRequestParams moose.ApiRequestParams,
	debugJson *string,
) uint32

func noSuchEndpoint(
	eventParams moose.EventParams,
	apiRequestParams moose.ApiRequestParams,
	debugJson *string,
) uint32 {
	return 0
}

func pickNotifier(endpoint string) notifyRequest {
	switch endpoint {
	case "/v1/servers":
		return func(
			eventParams moose.EventParams,
			apiRequestParams moose.ApiRequestParams,
			debugJson *string,
		) uint32 {
			return moose.NordvpnappSendServiceQualityApiRequestRequestServers(
				eventParams,
				apiRequestParams,
				endpoint,
				debugJson,
			)
		}
	case "/v1/servers/recommendations":
		return func(
			eventParams moose.EventParams,
			apiRequestParams moose.ApiRequestParams,
			debugJson *string,
		) uint32 {
			return moose.NordvpnappSendServiceQualityApiRequestRequestServersRecommendations(
				eventParams,
				apiRequestParams,
				endpoint,
				debugJson,
			)
		}
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
