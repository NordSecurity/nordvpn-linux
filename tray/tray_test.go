package tray

import (
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
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
