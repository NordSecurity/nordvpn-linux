package tray

import (
	"testing"

	"github.com/NordSecurity/nordvpn-linux/alert"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

func Test_selectIcon(t *testing.T) {
	category.Set(t, category.Unit)
	type args struct {
		desktopEnv string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "KDE environment", args: args{desktopEnv: "kde"}, want: IconBlack},
		{name: "MATE environment", args: args{desktopEnv: "mate"}, want: IconGray},
		{name: "Unknown environment", args: args{desktopEnv: "gnome"}, want: IconWhite},
		{name: "Empty environment", args: args{desktopEnv: ""}, want: IconWhite},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := selectIcon(tt.args.desktopEnv); got != tt.want {
				t.Errorf("selectIcon() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sortedConnections(t *testing.T) {
	category.Set(t, category.Unit)

	groups := func(names ...string) []*pb.ServerGroup {
		result := make([]*pb.ServerGroup, 0, len(names))
		for _, name := range names {
			result = append(result, &pb.ServerGroup{Name: name})
		}
		return result
	}

	tests := []struct {
		name string
		args []*pb.ServerGroup
		want []Server
	}{
		{
			name: "no groups",
			args: nil,
			want: []Server{},
		},
		{
			name: "empty and blank names are dropped",
			args: groups("", "   ", "\t"),
			want: []Server{},
		},
		{
			name: "surrounding whitespace is trimmed off the name",
			args: groups("  P2P  "),
			want: []Server{{name: "P2P", displayLabel: "P2P"}},
		},
		{
			name: "underscores become spaces in the label only",
			args: groups("Onion_Over_VPN"),
			want: []Server{{name: "Onion_Over_VPN", displayLabel: "Onion Over VPN"}},
		},
		{
			name: "groups are sorted by name",
			args: groups("P2P", "Double_VPN", "Obfuscated_Servers"),
			want: []Server{
				{name: "Double_VPN", displayLabel: "Double VPN"},
				{name: "Obfuscated_Servers", displayLabel: "Obfuscated Servers"},
				{name: "P2P", displayLabel: "P2P"},
			},
		},
		{
			name: "sorting is byte wise, so upper case comes before lower case",
			args: groups("dedicated", "Dedicated"),
			want: []Server{
				{name: "Dedicated", displayLabel: "Dedicated"},
				{name: "dedicated", displayLabel: "dedicated"},
			},
		},
		{
			name: "repeated names collapse into one entry",
			args: groups("Dedicated server", "P2P", "Dedicated server"),
			want: []Server{
				{name: "Dedicated server", displayLabel: "Dedicated server"},
				{name: "P2P", displayLabel: "P2P"},
			},
		},
		{
			name: "names that differ only by whitespace collapse together",
			args: groups(" P2P", "P2P ", "P2P"),
			want: []Server{{name: "P2P", displayLabel: "P2P"}},
		},
		{
			name: "blanks are dropped while the rest is kept",
			args: groups("P2P", "", "Double_VPN", "   "),
			want: []Server{
				{name: "Double_VPN", displayLabel: "Double VPN"},
				{name: "P2P", displayLabel: "P2P"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, sortedConnections(tt.args))
		})
	}
}

func Test_buildTimerString(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name         string
		remainingMin int
		expected     string
	}{
		{name: "zero minutes", remainingMin: 0, expected: "VPN connection resumes in 0min"},
		{name: "single digit minutes", remainingMin: 5, expected: "VPN connection resumes in 5min"},
		{name: "ten minutes", remainingMin: 10, expected: "VPN connection resumes in 10min"},
		{name: "59 minutes", remainingMin: 59, expected: "VPN connection resumes in 59min"},
		{name: "1 hour", remainingMin: 60, expected: "VPN connection resumes in 1h"},
		{name: "1 hour 5 minutes", remainingMin: 65, expected: "VPN connection resumes in 1h 5min"},
		{name: "2 hours 30 minutes", remainingMin: 150, expected: "VPN connection resumes in 2h 30min"},
		{name: "24 hours", remainingMin: 1440, expected: "VPN connection resumes in 24h"},
		{name: "99 hours 59 minutes", remainingMin: 5999, expected: "VPN connection resumes in 99h 59min"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildTimerString(tt.remainingMin)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func Test_setVpnStatus_pauseRemaining(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name                      string
		previousStatus            pb.ConnectionState
		newStatus                 pb.ConnectionState
		pauseRemainingDurationSec uint32
		expectedPauseRemainingMin int
		expectedStatusChanged     bool
		serverNameChange          bool
	}{
		{
			name:                      "CONNECTED state clears pause remaining",
			previousStatus:            pb.ConnectionState_PAUSED,
			newStatus:                 pb.ConnectionState_CONNECTED,
			pauseRemainingDurationSec: 0,
			expectedPauseRemainingMin: 0,
			expectedStatusChanged:     true,
			serverNameChange:          false,
		},
		{
			name:                      "PAUSED with 300 sec converts to 5 minutes",
			previousStatus:            pb.ConnectionState_CONNECTED,
			newStatus:                 pb.ConnectionState_PAUSED,
			pauseRemainingDurationSec: 300,
			expectedPauseRemainingMin: 5,
			expectedStatusChanged:     true,
			serverNameChange:          false,
		},
		{
			name:                      "PAUSED with 60 sec converts to 1 minute",
			previousStatus:            pb.ConnectionState_DISCONNECTED,
			newStatus:                 pb.ConnectionState_PAUSED,
			pauseRemainingDurationSec: 60,
			expectedPauseRemainingMin: 1,
			expectedStatusChanged:     true,
			serverNameChange:          false,
		},
		{
			name:                      "PAUSED with 0 sec stores as 0 minutes",
			previousStatus:            pb.ConnectionState_DISCONNECTED,
			newStatus:                 pb.ConnectionState_PAUSED,
			pauseRemainingDurationSec: 0,
			expectedPauseRemainingMin: 0,
			expectedStatusChanged:     true,
			serverNameChange:          false,
		},
		{
			name:                      "UNKNOWN_STATE clears pause remaining",
			previousStatus:            pb.ConnectionState_CONNECTED,
			newStatus:                 pb.ConnectionState_UNKNOWN_STATE,
			pauseRemainingDurationSec: 0,
			expectedPauseRemainingMin: 0,
			expectedStatusChanged:     true,
			serverNameChange:          false,
		},
		{
			name:                      "CONNECTING clears pause remaining",
			previousStatus:            pb.ConnectionState_PAUSED,
			newStatus:                 pb.ConnectionState_CONNECTING,
			pauseRemainingDurationSec: 0,
			expectedPauseRemainingMin: 0,
			expectedStatusChanged:     true,
			serverNameChange:          false,
		},
		{
			name:                      "24 hour pause converts correctly",
			previousStatus:            pb.ConnectionState_CONNECTED,
			newStatus:                 pb.ConnectionState_PAUSED,
			pauseRemainingDurationSec: 86400,
			expectedPauseRemainingMin: 1440,
			expectedStatusChanged:     true,
			serverNameChange:          false,
		},
		{
			name:                      "rounding: 45 seconds rounds up to 1 minute",
			previousStatus:            pb.ConnectionState_CONNECTED,
			newStatus:                 pb.ConnectionState_PAUSED,
			pauseRemainingDurationSec: 45,
			expectedPauseRemainingMin: 1,
			expectedStatusChanged:     true,
			serverNameChange:          false,
		},
		{
			name:                      "rounding: 1 second rounds up to 1 minute",
			previousStatus:            pb.ConnectionState_CONNECTED,
			newStatus:                 pb.ConnectionState_PAUSED,
			pauseRemainingDurationSec: 1,
			expectedPauseRemainingMin: 1,
			expectedStatusChanged:     true,
			serverNameChange:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ti := &Instance{
				client: nil,
				n:      &mockNotifier{},
				state: trayState{
					vpnStatus: tt.previousStatus,
					vpnName:   "OldServer",
				},
			}

			// Function signature: setVpnStatus(vpnStatus, vpnName, vpnHostname, vpnCity, vpnCountry, isMeshPeer, pauseRemainingDurationSec)
			changed := ti.setVpnStatus(
				tt.newStatus,
				"TestServer",
				"test.example.com",
				"Test City",
				"Test Country",
				false,
				tt.pauseRemainingDurationSec,
			)

			assert.Equal(t, tt.expectedStatusChanged, changed,
				"status change should match expected")
			assert.Equal(t, tt.expectedPauseRemainingMin, ti.state.pauseRemainingMin,
				"pauseRemainingMin should match expected value")
		})
	}
}

type mockNotifier struct{}

func (m *mockNotifier) Alert(body string) *alert.AlertBuilder {
	return alert.NewAlertBuilder(func(a alert.Alert) bool { return true }, body)
}

func (m *mockNotifier) Mute() {}

func (m *mockNotifier) Unmute() {}

func (m *mockNotifier) Close() error {
	return nil
}
