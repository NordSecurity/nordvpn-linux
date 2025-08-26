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

const (
	// Menu item labels
	labelVPNStatus             = "VPN %s"
	labelServerPrefix          = "Server: "
	labelCityPrefix            = "City: "
	labelCountryPrefix         = "Country: "
	labelDisconnect            = "Disconnect"
	labelQuickConnect          = "Quick Connect"
	labelConnectionSelection   = "Connect to"
	labelRecentConnections     = "Recent Connections:"
	labelReconnectTo           = "Reconnect to "
	labelConnectTo             = "Connect to "
	labelCountries             = "Countries:"
	labelActiveGoroutines      = "Active goroutines"
	labelActiveGoroutinesCount = "Active goroutines: %d"
	labelRedraw                = "Redraw"
	labelUpdate                = "Update"
	labelFullUpdate            = "Full update"
	labelQuit                  = "Quit"
	labelLoggedInAs            = "Logged in as:"
	labelLogOut                = "Log out"
	labelNotLoggedIn           = "Not logged in"
	labelLogIn                 = "Log in"
	labelSettings              = "Settings"
	labelNotifications         = "Notifications"
	labelTrayIcon              = "Tray icon"

	// Menu item tooltips
	tooltipConnectionSelection = "Choose connection type"
	tooltipRecentConnections   = "Select recent connection"
	tooltipCountries           = "Select Country"
	tooltipActiveGoroutines    = "Shows number of active background processes"
	tooltipRedraw              = "Force refresh the tray menu"
	tooltipUpdate              = "Refresh menu with latest status"
	tooltipFullUpdate          = "Perform complete menu refresh"
	tooltipQuit                = "Close NordVPN tray application"
	tooltipLoggedInAs          = "Your current account"
	tooltipLogOut              = "Sign out of your NordVPN account"
	tooltipNotLoggedIn         = "Sign in required to use VPN"
	tooltipLogIn               = "Sign in to your NordVPN account"
	tooltipSettings            = "Configure application preferences"
	tooltipNotifications       = "Toggle desktop notifications"
	tooltipTrayIcon            = "Show or hide tray icon"

	// Limits
	maxRecentConnections = 3
)

func addDebugSection(ti *Instance) {
	systray.AddSeparator()
	m := systray.AddMenuItem(labelActiveGoroutines, tooltipActiveGoroutines)
	m.Disable()
	go func() {
		for {
			select {
			case _, open := <-m.ClickedCh:
				if !open {
					return
				}
			case <-time.After(1 * time.Second):
				m.SetTitle(fmt.Sprintf(labelActiveGoroutinesCount, runtime.NumGoroutine()))
			}
		}
	}()
	mRedraw := systray.AddMenuItem(labelRedraw, tooltipRedraw)
	go func() {
		for {
			_, open := <-mRedraw.ClickedCh
			if !open {
				return
			}
			ti.redrawChan <- struct{}{}
		}
	}()
	mUpdate := systray.AddMenuItem(labelUpdate, tooltipUpdate)
	go func() {
		for {
			_, open := <-mUpdate.ClickedCh
			if !open {
				return
			}
			ti.updateChan <- false
		}
	}()
	mUpdateFull := systray.AddMenuItem(labelFullUpdate, tooltipFullUpdate)
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
	m := systray.AddMenuItem(labelQuit, tooltipQuit)
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
	addVPNStatusItem(ti)

	if ti.state.vpnStatus == pb.ConnectionState_CONNECTED {
		addConnectedInfo(ti)
		addDisconnectButton(ti)
	} else {
		addQuickConnectButton(ti)
	}

	addConnectionSelector(ti)
	systray.AddSeparator()
}

func addVPNStatusItem(ti *Instance) {
	status := strings.ToLower(ti.state.vpnStatus.String())
	label := fmt.Sprintf(labelVPNStatus, status)
	mStatus := systray.AddMenuItem(label, label)
	mStatus.Disable()
}

func addConnectedInfo(ti *Instance) {
	if serverName := ti.state.serverName(); serverName != "" {
		label := labelServerPrefix + serverName
		mHostname := systray.AddMenuItem(label, label)
		mHostname.Disable()
	}

	if ti.state.vpnCity != "" {
		label := labelCityPrefix + ti.state.vpnCity
		mCity := systray.AddMenuItem(label, label)
		mCity.Disable()
	}

	if ti.state.vpnCountry != "" {
		label := labelCountryPrefix + ti.state.vpnCountry
		mCountry := systray.AddMenuItem(label, label)
		mCountry.Disable()
	}
}

func addDisconnectButton(ti *Instance) {
	mDisconnect := systray.AddMenuItem(labelDisconnect, labelDisconnect)
	go handleDisconnectClick(ti, mDisconnect)
}

func addQuickConnectButton(ti *Instance) {
	mConnect := systray.AddMenuItem(labelQuickConnect, labelQuickConnect)
	go handleQuickConnectClick(ti, mConnect)
}

func handleDisconnectClick(ti *Instance, item *systray.MenuItem) {
	for {
		_, open := <-item.ClickedCh
		if !open {
			return
		}
		if ti.disconnect() {
			ti.updateChan <- true
		}
	}
}

func handleQuickConnectClick(ti *Instance, item *systray.MenuItem) {
	for {
		_, open := <-item.ClickedCh
		if !open {
			return
		}
		if ti.connect("", "") {
			ti.updateChan <- true
		}
	}
}

func addConnectionSelector(ti *Instance) {
	connectionSelector := systray.AddMenuItem(labelConnectionSelection, tooltipConnectionSelection)

	ti.state.mu.RLock()
	countries := append([]string(nil), ti.state.connSelector.countries...)
	ti.state.mu.RUnlock()

	recentConnections := fetchRecentConnections(ti)
	if len(recentConnections) > 0 {
		addRecentConnectionsSection(ti, connectionSelector, recentConnections)
		connectionSelector.AddSeparator()
	}

	addCountriesSection(ti, connectionSelector, countries)
}

func addRecentConnectionsSection(
	ti *Instance,
	parent *systray.MenuItem,
	connections []*pb.RecentConnectionModel,
) {
	parent.AddSubMenuItem(labelRecentConnections, tooltipRecentConnections).Disable()
	for _, conn := range connections {
		if conn == nil {
			continue
		}

		displayLabel := makeDisplayLabel(conn)
		if displayLabel == "" {
			continue
		}

		tooltip := labelReconnectTo + displayLabel
		item := parent.AddSubMenuItem(displayLabel, tooltip)

		go handleRecentConnectionClick(ti, item, conn)
	}
}

func addCountriesSection(
	ti *Instance,
	parent *systray.MenuItem,
	countries []string,
) {
	parent.AddSubMenuItem(labelCountries, tooltipCountries).Disable()
	for _, country := range countries {
		title := strings.ReplaceAll(country, "_", " ")
		tooltip := labelConnectTo + country
		item := parent.AddSubMenuItem(title, tooltip)

		go handleCountryClick(ti, item, country)
	}
}

func handleRecentConnectionClick(
	ti *Instance,
	item *systray.MenuItem,
	model *pb.RecentConnectionModel,
) {
	for {
		_, open := <-item.ClickedCh
		if !open {
			return
		}

		success := connectByConnectionModel(ti, model)
		if success {
			ti.updateChan <- true
		}
	}
}

func handleCountryClick(ti *Instance, item *systray.MenuItem, country string) {
	for {
		_, open := <-item.ClickedCh
		if !open {
			return
		}

		success := ti.connect(country, "")
		if success {
			ti.updateChan <- true
		}
	}
}

func addAccountSection(ti *Instance) {
	systray.AddSeparator()

	if ti.state.loggedIn {
		if ti.state.accountName != "" {
			m := systray.AddMenuItem(labelLoggedInAs, tooltipLoggedInAs)
			m.Disable()

			mName := systray.AddMenuItem(ti.state.accountName, ti.state.accountName)
			mName.Disable()
		}

		mLogout := systray.AddMenuItem(labelLogOut, tooltipLogOut)

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
		m := systray.AddMenuItem(labelNotLoggedIn, tooltipNotLoggedIn)
		m.Disable()

		mLogin := systray.AddMenuItem(labelLogIn, tooltipLogIn)

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
	mSettings := systray.AddMenuItem(labelSettings, tooltipSettings)
	// Workaround over the dbus issue described here: https://github.com/fyne-io/systray/issues/12
	// (It affects not only XFCE, but also other desktop environments.)
	time.AfterFunc(100*time.Millisecond, func() { addSettingsSubitems(ti, mSettings) })
}

func addSettingsSubitems(ti *Instance, mSettings *systray.MenuItem) {
	ti.state.mu.RLock()
	mNotifications := mSettings.AddSubMenuItemCheckbox(
		labelNotifications,
		tooltipNotifications,
		ti.state.notificationsStatus == Enabled,
	)
	mTray := mSettings.AddSubMenuItemCheckbox(
		labelTrayIcon,
		tooltipTrayIcon,
		ti.state.trayStatus == Enabled,
	)
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
