// Package daemon provides gRPC interface for management of vpn on the device and various related functionalities,
// such as communication with the backend api and configuration management.
package daemon

import (
	"sync/atomic"
	"time"

	"github.com/NordSecurity/nordvpn-linux/auth"
	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/core/mesh"
	"github.com/NordSecurity/nordvpn-linux/daemon/dns"
	daemonevents "github.com/NordSecurity/nordvpn-linux/daemon/events"
	"github.com/NordSecurity/nordvpn-linux/daemon/firewall"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/daemon/state"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/nc"
	"github.com/NordSecurity/nordvpn-linux/network"
	"github.com/NordSecurity/nordvpn-linux/networker"
	"github.com/NordSecurity/nordvpn-linux/norduser/service"

	"github.com/go-co-op/gocron/v2"
)

// RPC is a gRPC server.
type RPC struct {
	environment     internal.Environment
	ac              auth.Checker
	cm              config.Manager
	dm              *DataManager
	api             core.CombinedAPI
	serversAPI      core.ServersAPI
	credentialsAPI  core.CredentialsAPI
	cdn             core.CDN
	repo            *RepoAPI
	authentication  core.Authentication
	lastServer      core.Server
	version         string
	systemInfoFunc  func(string) string
	networkInfoFunc func() string
	events          *daemonevents.Events
	// factory picks which VPN implementation to use
	factory          FactoryFunc
	endpointResolver network.EndpointResolver
	endpoint         network.Endpoint
	scheduler        gocron.Scheduler
	netw             networker.Networker
	publisher        events.Publisher[string]
	nameservers      dns.Getter
	ncClient         nc.NotificationClient
	analytics        events.Analytics
	norduser         service.Service
	meshRegistry     mesh.Registry
	systemShutdown   atomic.Bool
	statePublisher   *state.StatePublisher
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
	fw firewall.Service,
	events *daemonevents.Events,
	factory FactoryFunc,
	endpointResolver network.EndpointResolver,
	netw networker.Networker,
	publisher events.Publisher[string],
	nameservers dns.Getter,
	ncClient nc.NotificationClient,
	analytics events.Analytics,
	norduser service.Service,
	meshRegistry mesh.Registry,
	statePublisher *state.StatePublisher,
) *RPC {
	scheduler, _ := gocron.NewScheduler(gocron.WithLocation(time.UTC))
	return &RPC{
		environment:      environment,
		ac:               ac,
		cm:               cm,
		dm:               dm,
		api:              api,
		serversAPI:       serversAPI,
		credentialsAPI:   credentialsAPI,
		cdn:              cdn,
		repo:             repo,
		authentication:   authentication,
		version:          version,
		systemInfoFunc:   getSystemInfo,
		networkInfoFunc:  getNetworkInfo,
		factory:          factory,
		events:           events,
		endpointResolver: endpointResolver,
		scheduler:        scheduler,
		netw:             netw,
		publisher:        publisher,
		nameservers:      nameservers,
		ncClient:         ncClient,
		analytics:        analytics,
		norduser:         norduser,
		meshRegistry:     meshRegistry,
		statePublisher:   statePublisher,
	}
}
