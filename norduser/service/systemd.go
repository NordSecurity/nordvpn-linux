package service

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/NordSecurity/nordvpn-linux/internal"
)

// SystemdNorduser manages norduser service through systemctl
type SystemdNorduser struct{}

// Enable and start norduser service
func (*SystemdNorduser) Enable(uid uint32) error {
	return systemdNorduser(uid, "enable")
}

// Disable and stop norduser service
func (*SystemdNorduser) Disable(uid uint32) error {
	return systemdNorduser(uid, "disable")
}

// Stop without disabling
func (*SystemdNorduser) Stop(uid uint32) error {
	return systemdNorduser(uid, "stop")
}

func systemdNorduser(uid uint32, command string) error {
	if uid == 0 {
		// #nosec G204 -- no input comes from user
		return exec.Command("systemctl", "--now", command, internal.Norduserd).Run()
	}

	// #nosec G204 -- no input comes from user
	cmd := exec.Command(
		"systemctl", "--user", "--now", command, internal.Norduserd,
	)
	cmd.Env = os.Environ()
	dbusAddr, err := internal.DBUSSessionBusAddress(int64(uid))
	if err != nil {
		return fmt.Errorf("failed to find active dbus session: %s", err)
	}
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
