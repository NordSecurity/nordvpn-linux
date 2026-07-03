package daemon

import (
	"context"
	"errors"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/NordSecurity/nordvpn-linux/test/mock"

	"github.com/stretchr/testify/assert"
)

type injectorVPN struct {
	*mock.WorkingVPN
	err error
}

func (i *injectorVPN) InjectVPNConnectionError(int32, string) error {
	return i.err
}

func TestInjectVpnConnectionError(t *testing.T) {
	category.Set(t, category.Unit)

	const maintenanceCode = int32(3)

	tests := []struct {
		name        string
		environment internal.Environment
		factory     FactoryFunc
		wantErr     bool
		wantType    int64
	}{
		{
			name:        "rejected in production",
			environment: internal.Production,
			factory: func(config.Technology) (vpn.VPN, error) {
				return &injectorVPN{WorkingVPN: &mock.WorkingVPN{}}, nil
			},
			wantErr: true,
		},
		{
			name:        "factory error is propagated",
			environment: internal.Development,
			factory: func(config.Technology) (vpn.VPN, error) {
				return nil, errors.New("no factory")
			},
			wantErr: true,
		},
		{
			name:        "backend does not support injection",
			environment: internal.Development,
			factory: func(config.Technology) (vpn.VPN, error) {
				return &mock.WorkingVPN{}, nil
			},
			wantErr: true,
		},
		{
			name:        "injector error (no active connection) is propagated",
			environment: internal.Development,
			factory: func(config.Technology) (vpn.VPN, error) {
				return &injectorVPN{WorkingVPN: &mock.WorkingVPN{}, err: errors.New("no monitor")}, nil
			},
			wantErr: true,
		},
		{
			name:        "success in dev returns CodeSuccess",
			environment: internal.Development,
			factory: func(config.Technology) (vpn.VPN, error) {
				return &injectorVPN{WorkingVPN: &mock.WorkingVPN{}}, nil
			},
			wantErr:  false,
			wantType: internal.CodeSuccess,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &RPC{environment: tt.environment, factory: tt.factory}

			resp, err := r.InjectVpnConnectionError(context.Background(), &pb.InjectVpnConnectionErrorRequest{
				TelioCode: maintenanceCode,
				Pubkey:    "test-key",
			})

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, resp)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantType, resp.GetType())
		})
	}
}
