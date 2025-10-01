// Package daemon provides gRPC interface for management of vpn on the device and various related functionalities,
// such as communication with the backend api and configuration management.
package daemon

import (
	"sync/atomic"
	"time"

	"github.com/NordSecurity/nordvpn-linux/auth"
	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/config/remote"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/daemon/dns"
	daemonevents "github.com/NordSecurity/nordvpn-linux/daemon/events"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/daemon/recents"
	"github.com/NordSecurity/nordvpn-linux/daemon/state"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/nc"
	"github.com/NordSecurity/nordvpn-linux/network"
	"github.com/NordSecurity/nordvpn-linux/networker"
	"github.com/NordSecurity/nordvpn-linux/norduser/service"
	"github.com/NordSecurity/nordvpn-linux/sharedctx"

	"github.com/go-co-op/gocron/v2"
)

// RPC is a gRPC server.
type RPC struct {
	environment    internal.Environment
	ac             auth.Checker
	cm             config.Manager
	dm             *DataManager
	api            core.CombinedAPI
	serversAPI     core.ServersAPI
	credentialsAPI core.CredentialsAPI
	cdn            core.CDN
	repo           *RepoAPI
	authentication core.Authentication
	lastServer     core.Server
	version        string
	events         *daemonevents.Events
	// factory picks which VPN implementation to use
	factory             FactoryFunc
	endpointResolver    network.EndpointResolver
	endpoint            network.Endpoint
	scheduler           gocron.Scheduler
	netw                networker.Networker
	publisher           events.Publisher[string]
	nameservers         dns.Getter
	ncClient            nc.NotificationClient
	analytics           events.Analytics
	norduser            service.Service
	systemShutdown      atomic.Bool
	statePublisher      *state.StatePublisher
	RequestedConnParams RequestedConnParamsStorage
	connectContext      *sharedctx.Context
	remoteConfigGetter  remote.ConfigGetter
	connectionInfo      *state.ConnectionInfo
	consentChecker      ConsentChecker
	recentVPNConnStore  *recents.RecentConnectionsStore
	dataUpdateEvents    *daemonevents.DataUpdateEvents
	pb.UnimplementedDaemonServer
}

func NewRPC(
	environment internal.Environment,
	ac auth.Checker,
	cm config.Manager,
	dm *DataManager,
	api core.CombinedAPI,
	serversAPI core.ServersAPI,
	credentialsAPI core.CredentialsAPI,
	cdn core.CDN,
	repo *RepoAPI,
	authentication core.Authentication,
	version string,
	events *daemonevents.Events,
	factory FactoryFunc,
	endpointResolver network.EndpointResolver,
	netw networker.Networker,
	publisher events.Publisher[string],
	nameservers dns.Getter,
	ncClient nc.NotificationClient,
	analytics events.Analytics,
	norduser service.Service,
	statePublisher *state.StatePublisher,
	connectContext *sharedctx.Context,
	remoteConfigGetter remote.ConfigGetter,
	connectionInfo *state.ConnectionInfo,
	consentChecker ConsentChecker,
	recentVPNConnStore *recents.RecentConnectionsStore,
	dataUpdateEvents *daemonevents.DataUpdateEvents,
) *RPC {
	scheduler, _ := gocron.NewScheduler(gocron.WithLocation(time.UTC))
	return &RPC{
		environment:        environment,
		ac:                 ac,
		cm:                 cm,
		dm:                 dm,
		api:                api,
		serversAPI:         serversAPI,
		credentialsAPI:     credentialsAPI,
		cdn:                cdn,
		repo:               repo,
		authentication:     authentication,
		version:            version,
		factory:            factory,
		events:             events,
		endpointResolver:   endpointResolver,
		scheduler:          scheduler,
		netw:               netw,
		publisher:          publisher,
		nameservers:        nameservers,
		ncClient:           ncClient,
		analytics:          analytics,
		norduser:           norduser,
		statePublisher:     statePublisher,
		connectContext:     connectContext,
		remoteConfigGetter: remoteConfigGetter,
		connectionInfo:     connectionInfo,
		consentChecker:     consentChecker,
		recentVPNConnStore: recentVPNConnStore,
		dataUpdateEvents:   dataUpdateEvents,
	}
}
