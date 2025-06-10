package sysinfo

import (
	"testing"
)

func Test_getDesktopEnvironment(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		want     string
	}{
		{"check for valid entry", "Unity", "unity"},
		{"check for valid entry", "ubuntu:GNOME", "gnome"},
		{"check for valid entry with trailing spaces", "kde ", "kde"},
		{"check for valid entry with random capital letters", "xFce", "xfce"},
		{"check for valid entry with preceding spaces", " mate", "mate"},
		{"check for empty entry", "", "none"},
		{"check for with spaces", "    ", "none"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
		t.Run(tt.name, func(t *testing.T) {
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
