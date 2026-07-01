package tray

import (
	"time"

	"github.com/NordSecurity/nordvpn-linux/log"
	"github.com/NordSecurity/nordvpn-linux/notify"
	"github.com/NordSecurity/nordvpn-linux/sysinfo"
)

const logTag = "[systray]"

func Start() {
	for !sniIsAvailable() {
		log.Errorf("%s system tray not available, retrying in 10s", logTag)
		<-time.After(10 * time.Second)
	}

	baseIcon := notify.GetIconPath(selectIcon(sysinfo.GetDisplayDesktopEnvironment()))
	connectedIcon := notify.GetIconPath("nordvpn-tray-blue")

	go listenForActivate()
	sniRun(baseIcon, connectedIcon)
}

func Stop() {
	sniStop()
}

func selectIcon(desktopEnv string) string {
	switch desktopEnv {
	case "kde":
		return "nordvpn-tray-black"
	case "mate":
		return "nordvpn-tray-gray"
	default:
		return "nordvpn-tray-white"
	}
}

func listenForActivate() {
	for range ActivateCh {
		log.Infof("%s tray activated", logTag)
		openGUI()
	}
}
