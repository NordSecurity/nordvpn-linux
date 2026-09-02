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
