package internal

import (
	"os"
	"os/exec"
	"strings"
)

// Timezone returns system timezone
func Timezone() string {
	// unfortunately this works only on systemd systems
	out, err := exec.Command("timedatectl", "show").CombinedOutput()
	if err != nil {
		// used as a fallback on non systemd systems
		path, err := os.Readlink("/etc/localtime")
		if err != nil {
			return "N/A"
		}
		zone := strings.TrimPrefix(path, "/usr/share/zoneinfo/")
		zone = strings.TrimPrefix(zone, "posix/") // /usr/share/zoneinfo/posix/
		zone = strings.TrimPrefix(zone, "right/") // /usr/share/zoneinfo/right/
		return zone
	}

	return extractZone(out)
}

func extractZone(input []byte) string {
	for line := range strings.SplitSeq(string(input), "\n") {
		prefix := "Timezone="
		if after, ok := strings.CutPrefix(line, prefix); ok {
			return after
		}
	}
	return ""
}
