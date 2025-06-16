package sysinfo

import (
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
)

func Test_getDesktopEnvironment(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name     string
		envValue string
		want     string
	}{
		{"Checks valid entry with Unity", "Unity", "unity"},
		{"Extracts Gnome from formatted entry", "ubuntu:GNOME", "gnome"},
		{"Handles trailing spaces in KDE entry", "kde ", "kde"},
		{"Corrects random capitalization in Xfce", "xFce", "xfce"},
		{"Handles preceding spaces in Mate entry", " mate", "mate"},
		{"Returns none for empty input", "", "none"},
		{"Returns none for input with only spaces", "    ", "none"},
	}

	for _, tt := range tests {
		t.Run(tt.name+"using 'XDG_CURRENT_DESKTOP'", func(t *testing.T) {
			mockEnv := func(key string) string {
				if key == "XDG_CURRENT_DESKTOP" {
					return tt.envValue
				}
				return ""
			}
			if got := getDesktopEnvironment(mockEnv); got != tt.want {
				t.Errorf("getDesktopEnvironment() = %v, want %v", got, tt.want)
			}
		})
	}

	for _, tt := range tests {
		t.Run(tt.name+"using 'DESKTOP_SESSION'", func(t *testing.T) {
			mockEnv := func(key string) string {
				if key == "DESKTOP_SESSION" {
					return tt.envValue
				}
				return ""
			}
			if got := getDesktopEnvironment(mockEnv); got != tt.want {
				t.Errorf("getDesktopEnvironment() = %v, want %v", got, tt.want)
			}
		})
	}
}
