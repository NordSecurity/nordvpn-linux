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
	log.Println(internal.DebugPrefix, "addDebugSection(): adding debug UI items")
	systray.AddSeparator()
	m := systray.AddMenuItem("Active goroutines", "Active goroutines")
	m.Disable()
	go func() {
		log.Println(internal.DebugPrefix, "addDebugSection(): starting goroutine counter updater")
		for {
			select {
			case _, open := <-m.ClickedCh:
				if !open {
					log.Println(internal.DebugPrefix, "addDebugSection(): Active goroutines item channel closed; exiting updater")
					return
				}
				log.Println(internal.DebugPrefix, "addDebugSection(): Active goroutines item clicked (no action)")
			case <-time.After(1 * time.Second):
				count := runtime.NumGoroutine()
				m.SetTitle(fmt.Sprintf("Active goroutines: %d", count))
				if ti.debugMode {
					log.Printf("%s addDebugSection(): updated goroutine count -> %d\n", internal.DebugPrefix, count)
				}
			}
		}
	}()
	mRedraw := systray.AddMenuItem("Redraw", "Redraw")
	go func() {
		log.Println(internal.DebugPrefix, "addDebugSection(): starting Redraw click handler")
		for {
			_, open := <-mRedraw.ClickedCh
			if !open {
				log.Println(internal.DebugPrefix, "addDebugSection(): Redraw item channel closed; exiting handler")
				return
			}
			log.Println(internal.DebugPrefix, "addDebugSection(): Redraw clicked — signalling redrawChan")
			ti.redrawChan <- struct{}{}
		}
	}()
	mUpdate := systray.AddMenuItem("Update", "Update")
	go func() {
		log.Println(internal.DebugPrefix, "addDebugSection(): starting Update click handler")
		for {
			_, open := <-mUpdate.ClickedCh
			if !open {
				log.Println(internal.DebugPrefix, "addDebugSection(): Update item channel closed; exiting handler")
				return
			}
			log.Println(internal.DebugPrefix, "addDebugSection(): Update clicked — requesting incremental update")
			ti.updateChan <- false
		}
	}()
	mUpdateFull := systray.AddMenuItem("Full update", "Full update")
	go func() {
		log.Println(internal.DebugPrefix, "addDebugSection(): starting Full update click handler")
		for {
			_, open := <-mUpdateFull.ClickedCh
			if !open {
				log.Println(internal.DebugPrefix, "addDebugSection(): Full update item channel closed; exiting handler")
				return
			}
			log.Println(internal.DebugPrefix, "addDebugSection(): Full update clicked — requesting full update")
			ti.updateChan <- true
		}
	}()
}

func addQuitItem(ti *Instance) {
	log.Println(internal.DebugPrefix, "addQuitItem(): adding Quit menu item")
	systray.AddSeparator()
	m := systray.AddMenuItem("Quit", "Quit")
	m.Enable()
	go func() {
		log.Println(internal.DebugPrefix, "addQuitItem(): starting click handler")
		_, open := <-m.ClickedCh
		if !open {
			log.Println(internal.DebugPrefix, "addQuitItem(): Quit item channel closed; exiting handler")
			return
		}
		log.Printf("%s Shutting down norduserd. To restart the process, run the \"nordvpn set tray on command\".", internal.InfoPrefix)
		ti.notifyForce("Shutting down norduserd. To restart the process, run the \"nordvpn set tray on command\".")
		select {
		case ti.quitChan <- norduser.StopRequest{}:
			log.Println(internal.InfoPrefix, "addQuitItem(): StopRequest sent to quitChan")
		default:
			log.Println(internal.DebugPrefix, "addQuitItem(): quitChan not ready; StopRequest not sent")
		}
	}()
}

func addDaemonErrorSection(ti *Instance) {
	if ti.state.daemonAvailable {
		log.Println(internal.DebugPrefix, "addDaemonErrorSection(): daemonAvailable=true — adding separator before error message")
		systray.AddSeparator()
	}
	log.Printf("%s addDaemonErrorSection(): displaying daemon error: %q\n", internal.ErrorPrefix, ti.state.daemonError)
	mError := systray.AddMenuItem(ti.state.daemonError, ti.state.daemonError)
	mError.Disable()
}

func addVpnSection(ti *Instance) {
	log.Printf("%s addVpnSection(): vpnStatus=%s host=%q name=%q city=%q country=%q virtual=%t\n",
		internal.DebugPrefix,
		ti.vpnConnStateToString(ti.state.vpnStatus), ti.state.vpnHostname, ti.state.vpnName, ti.state.vpnCity, ti.state.vpnCountry, ti.state.vpnVirtualLocation)

	mStatus := systray.AddMenuItem(
		"VPN "+strings.ToLower(ti.state.vpnStatus.String()),
		"VPN "+strings.ToLower(ti.state.vpnStatus.String()))
	mStatus.Disable()

	if ti.state.vpnStatus == pb.ConnectionState_CONNECTED {
		vpnServerName := ti.state.serverName()
		if vpnServerName != "" {
			log.Printf("%s addVpnSection(): adding Server line: %q\n", internal.DebugPrefix, vpnServerName)
			mHostname := systray.AddMenuItem("Server: "+vpnServerName, "Server: "+vpnServerName)
			mHostname.Disable()
		}

		if ti.state.vpnCity != "" {
			log.Printf("%s addVpnSection(): adding City line: %q\n", internal.DebugPrefix, ti.state.vpnCity)
			mCity := systray.AddMenuItem("City: "+ti.state.vpnCity, "City: "+ti.state.vpnCity)
			mCity.Disable()
		}

		if ti.state.vpnCountry != "" {
			log.Printf("%s addVpnSection(): adding Country line: %q\n", internal.DebugPrefix, ti.state.vpnCountry)
			mCountry := systray.AddMenuItem("Country: "+ti.state.vpnCountry, "Country: "+ti.state.vpnCountry)
			mCountry.Disable()
		}
		mDisconnect := systray.AddMenuItem("Disconnect", "Disconnect")
		go func() {
			log.Println(internal.DebugPrefix, "addVpnSection(): starting Disconnect click handler")
			success := false
			for !success {
				_, open := <-mDisconnect.ClickedCh
				if !open {
					log.Println(internal.DebugPrefix, "addVpnSection(): Disconnect item channel closed; exiting handler")
					return
				}
				log.Println(internal.InfoPrefix, "Disconnect clicked — attempting to disconnect")
				success = ti.disconnect()
				log.Printf("%s addVpnSection(): disconnect result success=%t\n", internal.DebugPrefix, success)
			}
			log.Println(internal.DebugPrefix, "addVpnSection(): disconnect completed — requesting full update")
			ti.updateChan <- true
		}()
	} else {
		mConnect := systray.AddMenuItem("Quick Connect", "Quick Connect")
		go func() {
			log.Println(internal.DebugPrefix, "addVpnSection(): starting Quick Connect click handler")
			success := false
			for !success {
				_, open := <-mConnect.ClickedCh
				if !open {
					log.Println(internal.DebugPrefix, "addVpnSection(): Quick Connect item channel closed; exiting handler")
					return
				}
				log.Println(internal.InfoPrefix, "Quick Connect clicked — attempting to connect")
				success = ti.connect("", "")
				log.Printf("%s addVpnSection(): quick connect result success=%t\n", internal.DebugPrefix, success)
			}
			log.Println(internal.DebugPrefix, "addVpnSection(): quick connect completed — requesting full update")
			ti.updateChan <- true
		}()
	}
	systray.AddSeparator()
}

func addAccountSection(ti *Instance) {
	log.Printf("%s addAccountSection(): loggedIn=%t accountName=%q\n", internal.DebugPrefix, ti.state.loggedIn, ti.state.accountName)
	systray.AddSeparator()

	if ti.state.loggedIn {
		if ti.state.accountName != "" {
			m := systray.AddMenuItem("Logged in as:", "Logged in as:")
			m.Disable()

			log.Printf("%s addAccountSection(): showing account name %q\n", internal.DebugPrefix, ti.state.accountName)
			mName := systray.AddMenuItem(ti.state.accountName, ti.state.accountName)
			mName.Disable()
		}

		mLogout := systray.AddMenuItem("Log out", "Log out")

		go func() {
			log.Println(internal.DebugPrefix, "addAccountSection(): starting Logout click handler")
			success := false
			for !success {
				_, open := <-mLogout.ClickedCh
				if !open {
					log.Println(internal.DebugPrefix, "addAccountSection(): Logout item channel closed; exiting handler")
					return
				}
				log.Println(internal.InfoPrefix, "Logout clicked — attempting to log out")
				success = ti.logout(false)
				log.Printf("%s addAccountSection(): logout result success=%t\n", internal.DebugPrefix, success)
			}
			log.Println(internal.DebugPrefix, "addAccountSection(): logout completed — requesting full update")
			ti.updateChan <- true
		}()
	} else {
		m := systray.AddMenuItem("Not logged in", "Not logged in")
		m.Disable()

		mLogin := systray.AddMenuItem("Log in", "Log in")

		go func() {
			log.Println(internal.DebugPrefix, "addAccountSection(): starting Login click handler")
			for {
				_, open := <-mLogin.ClickedCh
				if !open {
					log.Println(internal.DebugPrefix, "addAccountSection(): Login item channel closed; exiting handler")
					return
				}
				log.Println(internal.InfoPrefix, "Login clicked — invoking login flow")
				ti.login()
			}
		}()
	}
}

func addSettingsSection(ti *Instance) {
	log.Println(internal.DebugPrefix, "addSettingsSection(): adding Settings root item and scheduling submenu population")
	mSettings := systray.AddMenuItem("Settings", "Settings")
	// Workaround over the dbus issue described here: https://github.com/fyne-io/systray/issues/12
	// (It affects not only XFCE, but also other desktop environments.)
	time.AfterFunc(100*time.Millisecond, func() { addSettingsSubitems(ti, mSettings) })
}

func addSettingsSubitems(ti *Instance, mSettings *systray.MenuItem) {
	log.Println(internal.DebugPrefix, "addSettingsSubitems(): populating Settings submenu")
	ti.state.mu.RLock()
	mNotifications := mSettings.AddSubMenuItemCheckbox("Notifications", "Notifications", ti.state.notificationsStatus == Enabled)
	mTray := mSettings.AddSubMenuItemCheckbox("Tray icon", "Tray icon", ti.state.trayStatus == Enabled)
	log.Printf("%s addSettingsSubitems(): initial checkbox states — Notifications=%t Tray=%t\n", internal.DebugPrefix, ti.state.notificationsStatus == Enabled, ti.state.trayStatus == Enabled)
	ti.state.mu.RUnlock()

	go func() {
		log.Println(internal.DebugPrefix, "addSettingsSubitems(): starting Notifications click handler")
		success := false
		for !success {
			_, open := <-mNotifications.ClickedCh
			if !open {
				log.Println(internal.DebugPrefix, "addSettingsSubitems(): Notifications item channel closed; exiting handler")
				return
			}
			action := !mNotifications.Checked()
			log.Printf("%s addSettingsSubitems(): Notifications clicked — target state=%t\n", internal.InfoPrefix, action)
			success = ti.setNotify(action)
			log.Printf("%s addSettingsSubitems(): setNotify result success=%t\n", internal.DebugPrefix, success)
			if success {
				if action {
					mNotifications.Check()
				} else {
					mNotifications.Uncheck()
				}
			}
		}
		log.Println(internal.DebugPrefix, "addSettingsSubitems(): notifications toggle completed — requesting full update")
		ti.updateChan <- true
	}()

	go func() {
		log.Println(internal.DebugPrefix, "addSettingsSubitems(): starting Tray icon click handler")
		success := false
		for !success {
			_, open := <-mTray.ClickedCh
			if !open {
				log.Println(internal.DebugPrefix, "addSettingsSubitems(): Tray icon item channel closed; exiting handler")
				return
			}
			action := !mTray.Checked()
			log.Printf("%s addSettingsSubitems(): Tray icon clicked — target state=%t\n", internal.InfoPrefix, action)
			success = ti.setTray(action)
			log.Printf("%s addSettingsSubitems(): setTray result success=%t\n", internal.DebugPrefix, success)
			if success {
				if action {
					mTray.Check()
				} else {
					mTray.Uncheck()
				}
			}
		}
		log.Println(internal.DebugPrefix, "addSettingsSubitems(): tray toggle completed — requesting full update")
		ti.updateChan <- true
	}()

	systray.Refresh()
	log.Println(internal.DebugPrefix, "addSettingsSubitems(): systray.Refresh() called")
}

