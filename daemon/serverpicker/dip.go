package serverpicker

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/NordSecurity/nordvpn-linux/auth"
	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/log"
	"golang.org/x/exp/slices"
)

var ErrDedicatedIPServer = fmt.Errorf("selected dedicated IP servers group")

// IsAnyDIPServersAvailable returns true if dedicated IP server is selected for any of the provided services
func IsAnyDIPServersAvailable(dedicatedIPServices []auth.DedicatedIPService) bool {
	index := slices.IndexFunc(dedicatedIPServices, func(dipService auth.DedicatedIPService) bool {
		return len(dipService.ServerIDs) > 0
	})

	return index != -1
}

func getServerByID(servers core.Servers, serverID int64) (*core.Server, error) {
	index := slices.IndexFunc(servers, func(server core.Server) bool {
		return server.ID == serverID
	})

	if index == -1 {
		return nil, fmt.Errorf("server not found")
	}

	return &servers[index], nil
}

// SelectDedicatedIPServer picks a random dedicated IP server available to the
// user from their subscription services.
func SelectDedicatedIPServer(authChecker auth.Checker, servers core.Servers, cfg config.Config) (ServerSelection, error) {
	dedicatedIPServices, err := authChecker.GetDedicatedIPServices()
	if err != nil {
		log.Error(logPrefix, "getting dedicated IP service data:", err)
		if errors.Is(err, core.ErrUnauthorized) {
			return ServerSelection{}, internal.NewErrorWithCode(internal.CodeRevokedAccessToken)
		}
		return ServerSelection{}, internal.ErrUnhandled
	}

	if len(dedicatedIPServices) == 0 {
		return ServerSelection{}, internal.NewErrorWithCode(internal.CodeDedicatedIPRenewError)
	}

	if !IsAnyDIPServersAvailable(dedicatedIPServices) {
		return ServerSelection{}, internal.NewErrorWithCode(internal.CodeDedicatedIPServiceButNoServers)
	}

	serverIDs := []int64{}
	for _, service := range dedicatedIPServices {
		serverIDs = append(serverIDs, service.ServerIDs...)
	}

	if len(serverIDs) >= 2 {
		// #nosec G404 - Using math/rand for serverIDs shuffling is acceptable
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		r.Shuffle(len(serverIDs), func(i, j int) {
			serverIDs[i], serverIDs[j] = serverIDs[j], serverIDs[i]
		})
	}

	for _, serverID := range serverIDs {
		server, err := getServerByID(servers, serverID)
		if err != nil || server == nil {
			log.Error(logPrefix, "DIP server not found:", err)
			continue
		}

		if !MatchesUserSettings(*server, cfg) {
			log.Error(logPrefix, "cannot use server to connect because the server doesn't support user settings")
			continue
		}
		return ServerSelection{Server: server}, nil
	}

	return ServerSelection{}, internal.ErrServerIsUnavailable
}

func CheckDIPServerInSubscription(authChecker auth.Checker, server core.Server, cfg config.Config) error {
	dedicatedIPServices, err := authChecker.GetDedicatedIPServices()
	if err != nil {
		log.Error(logPrefix, "getting dedicated IP service data:", err)
		if errors.Is(err, core.ErrUnauthorized) {
			return err
		}
		return internal.ErrUnhandled
	}

	if len(dedicatedIPServices) == 0 {
		return internal.NewErrorWithCode(internal.CodeDedicatedIPRenewError)
	}

	if !IsAnyDIPServersAvailable(dedicatedIPServices) {
		return internal.NewErrorWithCode(internal.CodeDedicatedIPServiceButNoServers)
	}

	index := slices.IndexFunc(dedicatedIPServices, func(s auth.DedicatedIPService) bool {
		index := slices.Index(s.ServerIDs, server.ID)
		return index != -1
	})
	if index == -1 {
		log.Error(logPrefix, "server is not in the DIP servers list")
		return internal.NewErrorWithCode(internal.CodeDedicatedIPNoServer)
	}

	if !MatchesUserSettings(server, cfg) {
		log.Error(logPrefix, "failed to connect because the server doesn't support user settings")
		return internal.ErrServerIsUnavailable
	}

	return nil
}
