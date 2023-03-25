package internal

import (
	"crypto/sha256"
	"fmt"
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
		zone := strings.TrimLeft(path, "/usr/share/zoneinfo/")
		zone = strings.TrimLeft(zone, "posix/") // /usr/share/zoneinfo/posix/
		zone = strings.TrimLeft(zone, "right/") // /usr/share/zoneinfo/right/
		return zone
	}

	return extractZone(out)
}

func extractZone(input []byte) string {
	for _, line := range strings.Split(string(input), "\n") {
		prefix := "Timezone="
		if strings.HasPrefix(line, prefix) {
			return strings.TrimPrefix(line, prefix)
		}
	}
	return ""
}

// Obfuscate given string
func Obfuscate(text string) string {
	// removing excess symbols which might interfere with hashing
	text = strings.TrimSpace(text)

	sum := sha256.Sum256([]byte(text))
	return fmt.Sprintf("%x", sum[:7])
}
