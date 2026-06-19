package serverpicker

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/auth"
	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	devicekey "github.com/NordSecurity/nordvpn-linux/device_key"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/log"
)

var ErrDedicatedServer = fmt.Errorf("selected dedicated servers group")

type DSConnectionData struct {
	Ip               string
	Port             int64
	ServerPublicKey  string
	DevicePrivateKey string
}

func SelectDedicatedServer(
	authChecker auth.Checker,
	api core.DedicatedServersAPI,
	keyManager devicekey.DedicatedServersKeyManager,
) (ServerSelection, error) {
	service, err := authChecker.GetDedicatedServerService()
	if err != nil {
		log.Println(internal.ErrorPrefix, "checking dedicated servers service status:", err)
		if errors.Is(err, core.ErrUnauthorized) {
			return ServerSelection{}, internal.NewErrorWithCode(internal.CodeRevokedAccessToken)
		}
		return ServerSelection{}, internal.ErrUnhandled
	}

	if !service.Active {
		return ServerSelection{}, internal.NewErrorWithCode(internal.CodeDedicatedServersRenewError)
	}

	dedicatedServers, err := api.DedicatedServers()
	if err != nil {
		log.Println(internal.ErrorPrefix, "getting dedicated servers list:", err)
		return ServerSelection{}, internal.ErrUnhandled
	}

	if len(dedicatedServers) == 0 {
		return ServerSelection{}, internal.NewErrorWithCode(internal.CodeDedicatedServersServiceButNoServers)
	}

	// Currently there can be only one dedicated server per user.
	dedicatedServer := dedicatedServers[0]

	normalizedStatusValue := strings.ToLower(string(dedicatedServer.Status))
	if normalizedStatusValue == string(core.DedicatedServerStatusStopped) ||
		normalizedStatusValue == string(core.DedicatedServerStatusStopping) {
		return ServerSelection{}, internal.NewErrorWithCode(internal.CodeDedicatedServersCanNotConnect)
	}
	if normalizedStatusValue == string(core.DedicatedServerStatusNew) {
		return ServerSelection{}, internal.NewErrorWithCode(internal.CodeDedicatedServersServerNotSetUp)
	}
	if normalizedStatusValue != string(core.DedicatedServerStatusRunning) {
		return ServerSelection{}, internal.NewErrorWithCode(internal.CodeDedicatedServersNotReady)
	}

	dedicatedServerRegistrationData := keyManager.CheckAndRegisterDedicatedServers()
	if dedicatedServerRegistrationData == nil {
		return ServerSelection{}, fmt.Errorf("failed to register the dedicated server")
	}

	server := &core.Server{
		Name:   dedicatedServer.Name,
		Status: core.Online,
		Groups: core.Groups{core.Group{
			ID:    config.ServerGroup_DEDICATED_SERVER,
			Title: config.ServerGroup_DEDICATED_SERVER.String(),
		}},
		Locations:           core.Locations{dedicatedServer.Location},
		DedicatedServerUUID: dedicatedServer.UUID,
		// Invariant: dedicated servers are always of Nordlynx, hence Wireguard tech.
		// This is needed to correctly display dedicated servers in the recent connections list.
		Technologies: core.Technologies{core.Technology{ID: core.WireguardTech}},
	}

	result := ServerSelection{
		DedicatedServerStatus: dedicatedServer.Status,
		Server:                server,
	}
	return result, nil
}

func FetchDedicatedServerData(
	keyMan devicekey.DedicatedServersKeyManager,
	dedicatedServersAPI core.DedicatedServersAPI,
	serverUUID string,
) (DSConnectionData, error) {
	dedicatedServersDeviceData := keyMan.CheckAndRegisterDedicatedServers()
	if dedicatedServersDeviceData == nil {
		log.Error("failed to fetch the device key for dedicated server connection")
		return DSConnectionData{}, internal.ErrUnhandled
	}

	dedicatedServerConnectionData, err := getDedicatedServerConnectionData(
		dedicatedServersAPI,
		serverUUID,
		*dedicatedServersDeviceData,
	)
	if errors.Is(err, core.ErrDedicatedServersDeviceNotFound) {
		dedicatedServersDeviceData = keyMan.ForceRegisterDedicatedServers()
		if dedicatedServersDeviceData == nil {
			log.Error("failed to force dedicated server device registration")
			return DSConnectionData{}, core.ErrDedicatedServersDeviceNotRegistered
		}
		dedicatedServerConnectionData, err = getDedicatedServerConnectionData(
			dedicatedServersAPI,
			serverUUID,
			*dedicatedServersDeviceData)
	}

	if err == nil {
		dedicatedServerConnectionData.DevicePrivateKey = dedicatedServersDeviceData.DevicePrivateKey
	}

	return dedicatedServerConnectionData, err
}

func getDedicatedServerConnectionData(
	api core.DedicatedServersAPI,
	serverUUID string,
	deviceConnectionData devicekey.DedicatedServersConnectionData,
) (DSConnectionData, error) {
	connectResponse, err := api.DedicatedServerConnectCheck(serverUUID, core.DedicatedServerConnectRequest{
		DeviceUUID:      deviceConnectionData.DeviceUUID.String(),
		DevicePublicKey: deviceConnectionData.DevicePublicKey,
	})
	if err != nil {
		return DSConnectionData{}, fmt.Errorf("getting dedicated server connection data: %w", err)
	}

	addrPort := strings.Split(connectResponse.ServerEndpoint, ":")

	ip := addrPort[0]

	var dedicatedServerPort int64
	if len(addrPort) > 1 {
		port, err := strconv.Atoi(addrPort[1])
		if err != nil {
			log.Println(internal.ErrorPrefix, "parsing dedicated server port:", err)
			return DSConnectionData{}, fmt.Errorf("parsing dedicated server port: %w", err)
		}
		dedicatedServerPort = int64(port)
	}

	return DSConnectionData{
		Ip:              ip,
		Port:            dedicatedServerPort,
		ServerPublicKey: connectResponse.ServerPublicKey,
	}, nil
}

// IsDedicatedServer returns true if either serverTag or serverGroup represents the dedicated server group
func IsDedicatedServer(serverTag string, serverGroup string) bool {
	return groupConvert(serverTag) == config.ServerGroup_DEDICATED_SERVER ||
		groupConvert(serverGroup) == config.ServerGroup_DEDICATED_SERVER
}
