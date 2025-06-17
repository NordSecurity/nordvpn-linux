package sysinfo

import (
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
)

func Test_DisplayProtocol(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name   string
		envVar string
		want   string
	}{
		{"X11 lowercase", "x11", "x11"},
		{"X11 uppercase", "X11", "x11"},
		{"Wayland lowercase", "wayland", "wayland"},
		{"Wayland uppercase", "WAYLAND", "wayland"},
		{"Empty value", "", EnvValueUnset},
		{"Space for value", " ", EnvValueUnset},
		{"Leading spaces", "   x11", "x11"},
		{"Trailing spaces", "wayland   ", "wayland"},
		{"Mixed case and spaces", "  WaYlAnD  ", "wayland"},
		{"Mir value", "mir", "mir"},
	}

	for _, tt := range tests {
		mockReader := func(in string) string {
			if in == "XDG_SESSION_TYPE" {
				return tt.envVar
			}
			return ""
		}
		t.Run(tt.name, func(t *testing.T) {
			if got := getDisplayProtocol(mockReader); got != tt.want {
				t.Errorf("getDisplayProtocol() = %v, want %v", got, tt.want)
			}
		})
	}
}
