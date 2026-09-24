package tray

import (
	"testing"

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

func Test_vpnStateToStatusLabel(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name       string
		state      pb.ConnectionState
		groupLabel string
		want       string
	}{
		{name: "connected without group", state: pb.ConnectionState_CONNECTED, want: "Secured"},
		{name: "connected with group", state: pb.ConnectionState_CONNECTED, groupLabel: "Double VPN", want: "Secured: Double VPN"},
		{name: "connecting ignores group", state: pb.ConnectionState_CONNECTING, groupLabel: "Double VPN", want: "Connecting…"},
		{name: "disconnected ignores group", state: pb.ConnectionState_DISCONNECTED, groupLabel: "Double VPN", want: "Not secured"},
		{name: "paused ignores group", state: pb.ConnectionState_PAUSED, groupLabel: "Double VPN", want: "Not secured"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, vpnStateToStatusLabel(tt.state, tt.groupLabel))
		})
	}
}

func Test_setVpnStatus_GroupLabelChangeRebuildsMenu(t *testing.T) {
	category.Set(t, category.Unit)

	ti := newTrayFixture(t).instance
	connected := func(groupLabel string) bool {
		return ti.setVpnStatus(pb.ConnectionState_CONNECTED, "lt1", "lt1.nordvpn.com", "Vilnius", "Lithuania", groupLabel, false, 0)
	}

	assert.True(t, connected(""), "first connection changes state")
	assert.True(t, connected("Double VPN"), "group label change alone must rebuild the menu")
	assert.Equal(t, "Double VPN", ti.state.vpnGroupLabel)
	assert.False(t, connected("Double VPN"), "unchanged status must not rebuild the menu")
	assert.True(t, connected(""), "clearing the group label must rebuild the menu")
}
