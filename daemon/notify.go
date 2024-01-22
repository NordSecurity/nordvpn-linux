package daemon

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

const (
	summary = "NordVPN"
)

type NotificationType int64

func Notify(cm config.Manager, notificationType NotificationType, args []string) error {
	var cfg config.Config
	err := cm.Load(&cfg)
	if err != nil {
		log.Println(internal.ErrorPrefix, err)
	}

	// early return if no user enabled his notifications
	if cfg.UsersData == nil {
		return nil
	}

	for id := range cfg.UsersData.Notify {
		err = notify(id, handleNotificationType(notificationType, args))
		if err != nil {
			log.Println(internal.ErrorPrefix, err)
			continue
		}
	}

	return nil
}

func notify(id int64, body string) error {
	var cmd *exec.Cmd
	commandContext, cancelFunc := context.WithDeadline(context.Background(), time.Now().Add(time.Second*5))
	defer cancelFunc()
	if internal.IsCommandAvailable("notify-send") {
		cmd = exec.CommandContext(commandContext, "notify-send", "-t", "3000", "-i", IconPath, summary, body)
	} else if internal.IsCommandAvailable("kdialog") {
		cmd = exec.CommandContext(commandContext,
			"kdialog",
			"--title",
			summary,
			"--passivepopup",
			body,
			"--icon",
			IconPath,
			"3")
	} else {
		return nil
	}
	dbusAddr := internal.DBUSSessionBusAddress(id)
	if dbusAddr == "" {
		// user does not have active dbus session - cannot send notification
		return nil
	}
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "DISPLAY=:0.0")
	cmd.Env = append(cmd.Env, dbusAddr)
	cmd.SysProcAttr = &syscall.SysProcAttr{Credential: &syscall.Credential{Uid: uint32(id)}}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return errors.New(strings.Trim(string(out), "\n"))
	}
	return nil
}

func handleNotificationType(notificationType NotificationType, args []string) string {
	switch notificationType {
	case internal.NotificationConnected:
		return fmt.Sprintf(internal.ConnectSuccess, internal.StringsToInterfaces(args)...)
	case internal.NotificationReconnected:
		return fmt.Sprintf(internal.ReconnectSuccess, internal.StringsToInterfaces(args)...)
	case internal.NotificationDisconnected:
		return internal.DisconnectSuccess
	default:
		return fmt.Sprintf("Unknown type (%v)", notificationType)
	}
}
