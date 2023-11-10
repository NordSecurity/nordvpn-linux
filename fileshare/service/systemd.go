package service

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/NordSecurity/nordvpn-linux/internal"
)

// SystemdFileshare manages fileshare service through systemctl
type SystemdFileshare struct{}

// Enable and start fileshare service
func (*SystemdFileshare) Enable(uid, _ uint32) error {
	return systemdFileshare(uid, "enable")
}

// Disable and stop fileshare service
func (*SystemdFileshare) Disable(uid, _ uint32) error {
	return systemdFileshare(uid, "disable")
}

// Stop without disabling
func (*SystemdFileshare) Stop(uid, _ uint32) error {
	return systemdFileshare(uid, "stop")
}

func systemdFileshare(uid uint32, command string) error {
	if uid == 0 {
		// #nosec G204 -- no input comes from user
		return exec.Command("systemctl", "--now", command, internal.Fileshared).Run()
	}

	// #nosec G204 -- no input comes from user
	cmd := exec.Command(
		"systemctl", "--user", "--now", command, internal.Fileshared,
	)
	cmd.Env = os.Environ()
	dbusAddr := internal.DBUSSessionBusAddress(int64(uid))
	if dbusAddr == "" {
		return fmt.Errorf("active dbus session not found")
	}
	cmd.Env = append(cmd.Env, dbusAddr)
	cmd.SysProcAttr = &syscall.SysProcAttr{Credential: &syscall.Credential{Uid: uid}}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("systemctl execution error: %w; output - %s", err, output)
	}

	return nil
}
