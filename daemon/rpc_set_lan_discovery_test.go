package daemon

import (
	"context"
	"fmt"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/config"
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

func addLANToAllowlist(t *testing.T, allowlist config.Allowlist) config.Allowlist {
	t.Helper()

	allowlist.Subnets["10.0.0.0/8"] = true
	allowlist.Subnets["172.16.0.0/12"] = true
	allowlist.Subnets["192.168.0.0/16"] = true
	allowlist.Subnets["169.254.0.0/16"] = true

	return allowlist
}

func TestSetLANDiscovery_Success(t *testing.T) {
	category.Set(t, category.Unit)

	allowlistLAN := config.Allowlist{
		Subnets: map[string]bool{
			"10.0.0.0/8":     true,
			"172.16.0.0/12":  true,
			"192.168.0.0/16": true,
			"169.254.0.0/16": true,
		},
	}

	getAllowlist := func() config.Allowlist {
		allowlist := getEmptyAllowlist(t)
		allowlist.Subnets["207.240.205.230/24"] = true
		allowlist.Subnets["18.198.160.194/12"] = true
		allowlist.Ports.TCP[2000] = true
		allowlist.Ports.UDP[3000] = true
		return allowlist
	}

	tests := []struct {
		name              string
		enabled           bool
		currentEnabled    bool
		expectedStatus    pb.SetLANDiscoveryStatus
		currentAllowlist  config.Allowlist
		expectedAllowlist config.Allowlist
		// LAN subnets should not be included in configuration when added as a part of LAN discovery
		expectedConfigAllowlist config.Allowlist
	}{
		{
			name:                    "enable success",
			enabled:                 true,
			expectedStatus:          pb.SetLANDiscoveryStatus_DISCOVERY_CONFIGURED,
			currentAllowlist:        getEmptyAllowlist(t),
			expectedAllowlist:       allowlistLAN,
			expectedConfigAllowlist: getEmptyAllowlist(t),
		},
		{
			name:                    "disable success",
			enabled:                 false,
			currentEnabled:          true,
			expectedStatus:          pb.SetLANDiscoveryStatus_DISCOVERY_CONFIGURED,
			currentAllowlist:        getEmptyAllowlist(t),
			expectedAllowlist:       config.Allowlist{},
			expectedConfigAllowlist: getEmptyAllowlist(t),
		},
		{
			name:                    "enable preexisiting allowlist",
			enabled:                 true,
			expectedStatus:          pb.SetLANDiscoveryStatus_DISCOVERY_CONFIGURED,
			currentAllowlist:        getAllowlist(),
			expectedAllowlist:       addLANToAllowlist(t, getAllowlist()),
			expectedConfigAllowlist: getAllowlist(),
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
				c.AutoConnectData.Allowlist = test.currentAllowlist
				return c
			})

			networker := mockNetworker{
				allowlist: test.currentAllowlist,
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
			assert.Equal(t, test.expectedConfigAllowlist, cfg.AutoConnectData.Allowlist,
				"Invalid allowlist saved in the config.")
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
		unsetAllowlistErr error
		setAllowlistErr   error
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
			name:            "set allowlist error",
			enabled:         true,
			currentEnabled:  false,
			expectedEnabled: false,
			setAllowlistErr: fmt.Errorf("failed to set allowlist"),
			expectedError:   pb.SetErrorCode_FAILURE,
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
				setAllowlistErr:   test.setAllowlistErr,
				unsetAllowlistErr: test.unsetAllowlistErr,
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
