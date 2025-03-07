package tray

import (
	"fmt"
	"log"
	"runtime"
	"strings"
	"time"

	"github.com/NordSecurity/systray"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/norduser"
)

func addDebugSection(ti *Instance) {
	systray.AddSeparator()
	m := systray.AddMenuItem("Active goroutines", "Active goroutines")
	m.Disable()
	go func() {
		for {
			select {
			case _, open := <-m.ClickedCh:
				if !open {
					return
				}
			case <-time.After(1 * time.Second):
				m.SetTitle(fmt.Sprintf("Active goroutines: %d", runtime.NumGoroutine()))
			}
		}
	}()
	mRedraw := systray.AddMenuItem("Redraw", "Redraw")
	go func() {
		for {
			_, open := <-mRedraw.ClickedCh
			if !open {
				return
			}
			ti.redrawChan <- struct{}{}
		}
	}()
	mUpdate := systray.AddMenuItem("Update", "Update")
	go func() {
		for {
			_, open := <-mUpdate.ClickedCh
			if !open {
				return
			}
			ti.updateChan <- false
		}
	}()
	mUpdateFull := systray.AddMenuItem("Full update", "Full update")
	go func() {
		for {
			_, open := <-mUpdateFull.ClickedCh
			if !open {
				return
			}
			ti.updateChan <- true
		}
	}()
}

func addQuitItem(ti *Instance) {
	systray.AddSeparator()
	m := systray.AddMenuItem("Quit", "Quit")
	m.Enable()
	go func() {
		_, open := <-m.ClickedCh
		if !open {
			return
		}
		log.Printf("%s Shutting down norduserd. To restart the process, run the \"nordvpn set tray on command\".", internal.InfoPrefix)
		ti.notifyForce("Shutting down norduserd. To restart the process, run the \"nordvpn set tray on command\".")
		select {
		case ti.quitChan <- norduser.StopRequest{}:
		default:
		}
	}()
}

func addDaemonErrorSection(ti *Instance) {
	if ti.state.daemonAvailable {
		systray.AddSeparator()
	}
	mError := systray.AddMenuItem(ti.state.daemonError, ti.state.daemonError)
	mError.Disable()
}

func addVpnSection(ti *Instance) {
	mStatus := systray.AddMenuItem(
		"VPN "+strings.ToLower(ti.state.vpnStatus.String()),
		"VPN "+strings.ToLower(ti.state.vpnStatus.String()))
	mStatus.Disable()

	if ti.state.vpnStatus == pb.ConnectionState_CONNECTED {
		vpnServerName := ti.state.serverName()
		if vpnServerName != "" {
			mHostname := systray.AddMenuItem("Server: "+vpnServerName, "Server: "+vpnServerName)
			mHostname.Disable()
		}

		if ti.state.vpnCity != "" {
			mCity := systray.AddMenuItem("City: "+ti.state.vpnCity, "City: "+ti.state.vpnCity)
			mCity.Disable()
		}

		if ti.state.vpnCountry != "" {
			mCountry := systray.AddMenuItem("Country: "+ti.state.vpnCountry, "Country: "+ti.state.vpnCountry)
			mCountry.Disable()
		}
		mDisconnect := systray.AddMenuItem("Disconnect", "Disconnect")
		go func() {
			success := false
			for !success {
				_, open := <-mDisconnect.ClickedCh
				if !open {
					return
				}
				success = ti.disconnect()
			}
			ti.updateChan <- true
		}()
	} else {
		mConnect := systray.AddMenuItem("Quick Connect", "Quick Connect")
		go func() {
			success := false
			for !success {
				_, open := <-mConnect.ClickedCh
				if !open {
					return
				}
				success = ti.connect("", "")
			}
			ti.updateChan <- true
		}()
	}
	systray.AddSeparator()
}

func addAccountSection(ti *Instance) {
	systray.AddSeparator()

	if ti.state.loggedIn {
		if ti.state.accountName != "" {
			m := systray.AddMenuItem("Logged in as:", "Logged in as:")
			m.Disable()

			mName := systray.AddMenuItem(ti.state.accountName, ti.state.accountName)
			mName.Disable()
		}

		mLogout := systray.AddMenuItem("Log out", "Log out")

		go func() {
			success := false
			for !success {
				_, open := <-mLogout.ClickedCh
				if !open {
					return
				}
				success = ti.logout(false)
			}
			ti.updateChan <- true
		}()
	} else {
		m := systray.AddMenuItem("Not logged in", "Not logged in")
		m.Disable()

		mLogin := systray.AddMenuItem("Log in", "Log in")

		go func() {
			for {
				_, open := <-mLogin.ClickedCh
				if !open {
					return
				}
				ti.login()
			}
		}()
	}
}

func addSettingsSection(ti *Instance) {
	mSettings := systray.AddMenuItem("Settings", "Settings")
	// Workaround over the dbus issue described here: https://github.com/fyne-io/systray/issues/12
	// (It affects not only XFCE, but also other desktop environments.)
	time.AfterFunc(100*time.Millisecond, func() { addSettingsSubitems(ti, mSettings) })
}

func addSettingsSubitems(ti *Instance, mSettings *systray.MenuItem) {
	ti.state.mu.RLock()
	mNotifications := mSettings.AddSubMenuItemCheckbox("Notifications", "Notifications", ti.state.notificationsStatus == Enabled)
	mTray := mSettings.AddSubMenuItemCheckbox("Tray icon", "Tray icon", ti.state.trayStatus == Enabled)
	ti.state.mu.RUnlock()

	go func() {
		success := false
		for !success {
			_, open := <-mNotifications.ClickedCh
			if !open {
				return
			}
			action := !mNotifications.Checked()
			success = ti.setNotify(action)
			if success {
				if action {
					mNotifications.Check()
				} else {
					mNotifications.Uncheck()
				}
			}
		}
		ti.updateChan <- true
	}()

	go func() {
		success := false
		for !success {
			_, open := <-mTray.ClickedCh
			if !open {
				return
			}
			action := !mTray.Checked()
			success = ti.setTray(action)
			if success {
				if action {
					mTray.Check()
				} else {
					mTray.Uncheck()
				}
			}
		}
		ti.updateChan <- true
	}()

	systray.Refresh()
}
