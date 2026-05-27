package nft

import (
	"net"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

func TestCalculateFirstAndLastV4Prefix(t *testing.T) {
	category.Set(t, category.Unit)
	tests := []struct {
		name          string
		cidr          string
		wantErr       error
		wantedStartIP net.IP
		wantedEndIP   net.IP
	}{
		{
			name:    "Regular cidr",
			cidr:    "192.168.0.0/16",
			wantErr: nil,
			//Inclusive
			wantedStartIP: net.ParseIP("192.168.0.0").To4(),
			//non inclusive upper bound
			wantedEndIP: net.ParseIP("192.169.0.0").To4(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) { // nil != nil, false != false
			startIP, endIP, err := calculateFirstAndLastV4Prefix(tt.cidr)
			if err != tt.wantErr {
				t.Errorf("calculateFirstAndLastV4Prefix() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, tt.wantedStartIP, startIP)
			assert.Equal(t, tt.wantedEndIP, endIP)
		})
	}
}
