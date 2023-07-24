package daemon

import (
	"context"
	"fmt"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core/mesh"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func getEmptyAllowlist(t *testing.T) config.Allowlist {
	t.Helper()

	return config.Allowlist{
		Ports: config.Ports{
			TCP: make(config.PortSet),
			UDP: make(config.PortSet),
		},
		Subnets: make(config.Subnets),
	}
}

func addLANToWhitelist(t *testing.T, whitelist config.Allowlist) config.Allowlist {
	t.Helper()

	whitelist.Subnets["10.0.0.0/8"] = true
	whitelist.Subnets["172.16.0.0/12"] = true
	whitelist.Subnets["192.168.0.0/16"] = true
	whitelist.Subnets["169.254.0.0/16"] = true

	return whitelist
}

func TestSetLANDiscovery_Success(t *testing.T) {
	category.Set(t, category.Unit)

	whitelistLAN := config.Allowlist{
		Subnets: map[string]bool{
			"10.0.0.0/8":     true,
			"172.16.0.0/12":  true,
			"192.168.0.0/16": true,
			"169.254.0.0/16": true,
		},
	}

	whitelist := getEmptyAllowlist(t)
	whitelist.Subnets["207.240.205.230/24"] = true
	whitelist.Subnets["18.198.160.194/12"] = true
	whitelist.Ports.TCP[2000] = true
	whitelist.Ports.UDP[3000] = true

	whitelistWithLAN := whitelist
	whitelistWithLAN.Subnets["10.0.0.0/8"] = true

	tests := []struct {
		name              string
		enabled           bool
		currentEnabled    bool
		expectedStatus    pb.SetLANDiscoveryStatus
		currentWhitelist  config.Allowlist
		expectedWhitelist config.Allowlist
		// LAN subnets should not be included in configuration when added as a part of LAN discovery
		expectedConfigWhitelist config.Allowlist
	}{
		{
			name:                    "enable success",
			enabled:                 true,
			expectedStatus:          pb.SetLANDiscoveryStatus_DISCOVERY_CONFIGURED,
			currentWhitelist:        getEmptyAllowlist(t),
			expectedWhitelist:       whitelistLAN,
			expectedConfigWhitelist: getEmptyAllowlist(t),
		},
		{
			name:                    "disable success",
			enabled:                 false,
			currentEnabled:          true,
			expectedStatus:          pb.SetLANDiscoveryStatus_DISCOVERY_CONFIGURED,
			currentWhitelist:        getEmptyAllowlist(t),
			expectedWhitelist:       config.Allowlist{},
			expectedConfigWhitelist: getEmptyAllowlist(t),
		},
		{
			name:                    "enable, preexisiting whitelist",
			enabled:                 true,
			expectedStatus:          pb.SetLANDiscoveryStatus_DISCOVERY_CONFIGURED,
			currentWhitelist:        whitelist,
			expectedWhitelist:       addLANToWhitelist(t, whitelist),
			expectedConfigWhitelist: addLANToWhitelist(t, whitelist),
		},
		{
			name:                    "disable, preexisiting whitelist contains LAN, LAN is not removed",
			enabled:                 false,
			currentEnabled:          true,
			expectedStatus:          pb.SetLANDiscoveryStatus_DISCOVERY_CONFIGURED,
			currentWhitelist:        whitelistWithLAN,
			expectedWhitelist:       whitelistWithLAN,
			expectedConfigWhitelist: whitelistWithLAN,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			machineID, _ := uuid.NewUUID()
			filesystem := newFilesystemMock(t)
			configManager := config.NewFilesystemConfigManager(
				"/location", "/vault", "",
				&machineIDGetterMock{machineID: machineID},
				&filesystem)

			configManager.SaveWith(func(c config.Config) config.Config {
				c.LanDiscovery = test.currentEnabled
				c.AutoConnectData.Allowlist = test.currentWhitelist
				return c
			})

			networker := mockNetworker{
				allowlist: test.currentWhitelist,
			}

			rpc := RPC{
				cm:           configManager,
				netw:         &networker,
				meshRegistry: &RegistryMock{}}
			resp, err := rpc.SetLANDiscovery(context.Background(), &pb.SetLANDiscoveryRequest{
				Enabled: test.enabled,
			})

			var cfg config.Config
			configManager.Load(&cfg)

			assert.Nil(t, err, "RPC ended in error.")
			assert.IsType(t, &pb.SetLANDiscoveryResponse_SetLanDiscoveryStatus{}, resp.Response,
				"SetLANDiscovery response is of invalid type.")
			assert.Equal(t, test.expectedStatus, resp.GetSetLanDiscoveryStatus(),
				"Invalid status returned in SetLANDiscovery response.")
			assert.Equal(t, test.enabled, cfg.LanDiscovery,
				"LAN discovery was not enabled in config.")
			assert.Equal(t, test.expectedConfigWhitelist, cfg.AutoConnectData.Allowlist,
				"Invalid whitelist saved in the config.")
			assert.Equal(t, test.enabled, networker.lanDiscovery)
		})
	}
}

func TestSetLANDiscovery_Error(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name              string
		enabled           bool
		currentEnabled    bool
		expectedEnabled   bool
		unsetWhitelistErr error
		setWhitelistErr   error
		expectedError     pb.SetErrorCode
	}{
		{
			name:            "already enabled",
			enabled:         true,
			currentEnabled:  true,
			expectedEnabled: true,
			expectedError:   pb.SetErrorCode_ALREADY_SET,
		},
		{
			name:            "already disabled",
			enabled:         false,
			currentEnabled:  false,
			expectedEnabled: false,
			expectedError:   pb.SetErrorCode_ALREADY_SET,
		},
		{
			name:            "set whitelist error",
			enabled:         true,
			currentEnabled:  false,
			expectedEnabled: false,
			setWhitelistErr: fmt.Errorf("failed to set whitelist"),
			expectedError:   pb.SetErrorCode_FAILURE,
		},
		{
			name:              "unset whitelist error",
			enabled:           true,
			currentEnabled:    false,
			expectedEnabled:   false,
			unsetWhitelistErr: fmt.Errorf("failed to unset whitelist"),
			expectedError:     pb.SetErrorCode_FAILURE,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			uuid, _ := uuid.NewUUID()
			filesystem := newFilesystemMock(t)
			configManager := config.NewFilesystemConfigManager(
				"/location", "/vault", "",
				&machineIDGetterMock{machineID: uuid},
				&filesystem)

			configManager.SaveWith(func(c config.Config) config.Config {
				c.LanDiscovery = test.currentEnabled
				c.AutoConnectData.Allowlist = getEmptyAllowlist(t)
				return c
			})

			networker := mockNetworker{
				allowlist:         getEmptyAllowlist(t),
				setWhitelistErr:   test.setWhitelistErr,
				unsetWhitelistErr: test.unsetWhitelistErr,
				vpnActive:         true,
			}

			rpc := RPC{
				cm:   configManager,
				netw: &networker}
			resp, err := rpc.SetLANDiscovery(context.Background(), &pb.SetLANDiscoveryRequest{
				Enabled: test.enabled,
			})

			var cfg config.Config
			configManager.Load(&cfg)

			assert.Nil(t, err, "RPC ended in error.")
			assert.IsType(t, &pb.SetLANDiscoveryResponse_ErrorCode{}, resp.Response,
				"SetLANDiscovery response is of invalid type.")
			assert.Equal(t, test.expectedError, resp.GetErrorCode(),
				"Invalid status returned in SetLANDiscovery response.")
			assert.Equal(t, test.expectedEnabled, cfg.LanDiscovery,
				"LAN discovery was not enabled in config.")
		})
	}
}

func TestSetLANDiscovery_MeshInteraction(t *testing.T) {
	tests := []struct {
		name   string
		enable bool
	}{
		{
			name:   "enable lan discovery",
			enable: true,
		},
		{
			name:   "disable lan discovery",
			enable: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			machineID, _ := uuid.NewUUID()
			meshID := uuid.MustParse("62a6f5c2-2579-11ee-be56-0242ac120002")
			filesystem := newFilesystemMock(t)
			configManager := config.NewFilesystemConfigManager(
				"/location", "/vault", "",
				&machineIDGetterMock{machineID: machineID},
				&filesystem)

			configManager.SaveWith(func(c config.Config) config.Config {
				c.LanDiscovery = !test.enable // set to negation of desired result to avoid AlreadySet error
				c.Mesh = true
				c.MeshDevice = &mesh.Machine{
					ID: meshID,
				}
				return c
			})

			peers := mesh.MachinePeers{
				mesh.MachinePeer{
					Hostname: "test0-pyrenees.nord",
				},
				mesh.MachinePeer{
					Hostname: "test1-himalayas.nord",
				},
			}

			networker := mockNetworker{}

			registry := RegistryMock{
				peers: peers,
			}

			rpc := RPC{
				cm:           configManager,
				netw:         &networker,
				meshRegistry: &registry}

			resp, err := rpc.SetLANDiscovery(context.Background(), &pb.SetLANDiscoveryRequest{
				Enabled: test.enable,
			})
			assert.Nil(t, err, "RPC eneded in error.")
			assert.IsType(t, &pb.SetLANDiscoveryResponse_SetLanDiscoveryStatus{}, resp.Response,
				"Invalid type of response from RPC, means that RPC was not succesfull."+
					"Succesfull responses should be of type pb.SetLANDiscoveryResponse_SetLanDiscoveryStatus")

			assert.Equal(t, peers, networker.meshPeers, "Invalid mesh peers provided to the networker. "+
				"When meshnet is enabled, peers passed to (Networker).SetLanDiscoveryAndResetMesh "+
				"should be the same as peers returned by (Registry).List.")

			assert.Equal(t, test.enable, networker.lanDiscovery, "LAN discovery was not configured in the networker.")
		})
	}
}
