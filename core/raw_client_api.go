package core

import "github.com/NordSecurity/nordvpn-linux/core/mesh"

type RawClientAPI interface {
	RawCredentialsAPI
	RawInsightsAPI
	RawServersAPI
	RawCombinedAPI
	RawSubscriptionAPI
	mesh.Mapper
	mesh.Registry
	mesh.Inviter
}
