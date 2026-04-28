package daemon

import (
	"fmt"
	"testing"

	devicekey "github.com/NordSecurity/nordvpn-linux/device_key"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	testauth "github.com/NordSecurity/nordvpn-linux/test/mock/auth"
	testdevicekey "github.com/NordSecurity/nordvpn-linux/test/mock/devicekey"
	"github.com/stretchr/testify/assert"
)

func TestSyncDevice(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name                             string
		hasDedicatedServersService       bool
		hasDedicatedServerServiceErr     error
		dedicatedServersRegistrationData *devicekey.DedicatedServersRegistrationData
		shouldReturnError                bool
		expectedKeyRegistered            bool
	}{
		{
			name:                             "user has dedicated servers service, key is registered",
			hasDedicatedServersService:       true,
			dedicatedServersRegistrationData: &devicekey.DedicatedServersRegistrationData{},
			expectedKeyRegistered:            true,
			shouldReturnError:                false,
		},
		{
			name:                             "user doesn't have dedicated servers service, key is not registered",
			hasDedicatedServersService:       false,
			dedicatedServersRegistrationData: &devicekey.DedicatedServersRegistrationData{},
			expectedKeyRegistered:            false,
			shouldReturnError:                false,
		},
		{
			name:                             "user has dedicated servers service, key registration fails, error is returned",
			hasDedicatedServersService:       true,
			dedicatedServersRegistrationData: nil,
			expectedKeyRegistered:            false,
			shouldReturnError:                true,
		},
		{
			name:                         "checking dedicated server service status fails, error is returned",
			hasDedicatedServerServiceErr: fmt.Errorf("check failed"),
			shouldReturnError:            true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			authCheckerMock := testauth.AuthCheckerMock{
				DedicatedServerService:       test.hasDedicatedServersService,
				HasDedicatedServerServiceErr: test.hasDedicatedServerServiceErr,
			}
			deviceKeyManagerMock := testdevicekey.MockDeviceKeyManager{
				DedicatedServerRegistrationData: test.dedicatedServersRegistrationData,
			}
			r := RPC{
				ac:                         &authCheckerMock,
				dedicatedServersKeyManager: &deviceKeyManagerMock,
			}

			err := r.SyncDevice()
			if test.shouldReturnError {
				assert.NotNil(t, err, "Error not returned by SyncDevice when expected.")
			} else {
				assert.Nil(t, err, "Error returned by SyncDevice when not expected.")
			}
			assert.Equal(t, test.expectedKeyRegistered, deviceKeyManagerMock.WasKeyRegistered,
				"Unexpected key registration status.")
		})
	}
}
