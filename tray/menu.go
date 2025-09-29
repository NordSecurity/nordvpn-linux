package tray

import (
	"fmt"
	"log"
	"runtime"
	"slices"
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
	labelServerPrefix          = "Server:"
	labelCityPrefix            = "City:"
	labelCountryPrefix         = "Country:"
	labelDisconnect            = "Disconnect"
	labelQuickConnect          = "Quick Connect"
	labelConnectionSelection   = "Connect to"
	labelRecentConnections     = "Recent Connections:"
	labelReconnectTo           = "Reconnect to"
	labelConnectTo             = "Connect to"
	labelCountries             = "Countries:"
	labelSpecialtyServers      = "Specialty servers:"
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
	tooltipSpecialtyServers    = "Select Specialty server"
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

	// System messages
	msgShutdownNotification = "Shutting down norduserd. To restart the process, run the \"nordvpn set tray on command\"."
)

func handleMenuItemClick(item *systray.MenuItem, action func()) {
	if item == nil || action == nil {
		return
	}
	go func() {
		for {
			_, open := <-item.ClickedCh
			if !open {
				return
			}
			action()
		}
	}()
}

func handleMenuItemClickWithRetry(item *systray.MenuItem, action func() bool) {
	if item == nil || action == nil {
		return
	}
	go func() {
		success := false
		for !success {
			_, open := <-item.ClickedCh
			if !open {
				return
			}
			success = action()
		}
	}()
}

func handleCheckboxOption(ti *Instance, item *systray.MenuItem, setter func(bool) bool) {
	if ti == nil || item == nil || setter == nil {
		return
	}

	ti.checkboxSync.HandleCheckboxOption(item, setter)
}

func addDebugSection(ti *Instance) {
	if ti == nil {
		return
	}
	if !ti.debugMode {
		return
	}

	systray.AddSeparator()
	m := systray.AddMenuItem(labelActiveGoroutines, tooltipActiveGoroutines)
	m.Disable()
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case _, open := <-m.ClickedCh:
				if !open {
					return
				}
			case <-ticker.C:
				m.SetTitle(fmt.Sprintf(labelActiveGoroutinesCount, runtime.NumGoroutine()))
			}
		}
	}()
}

func buildQuitButton(ti *Instance) {
	if ti == nil {
		return
	}
	systray.AddSeparator()
	m := systray.AddMenuItem(labelQuit, tooltipQuit)
	m.Enable()
	go func() {
		_, open := <-m.ClickedCh
		if !open {
			return
		}
		log.Printf("%s %s", internal.InfoPrefix, msgShutdownNotification)
		ti.notify(Force, msgShutdownNotification)
		select {
		case ti.quitChan <- norduser.StopRequest{}:
		default:
		}
	}()
}

func buildDaemonErrorSection(ti *Instance) {
	if ti == nil {
		return
	}
	if ti.state.daemonError == "" {
		return
	}

	if ti.state.daemonAvailable {
		systray.AddSeparator()
	}
	mError := systray.AddMenuItem(ti.state.daemonError, ti.state.daemonError)
	mError.Disable()
}

func buildConnectionSection(ti *Instance) {
	if ti == nil {
		return
	}
	if !ti.state.daemonAvailable || !ti.state.loggedIn {
		return
	}

	buildVPNStatusLabel(ti)
	if ti.state.vpnStatus == pb.ConnectionState_CONNECTED {
		buildConnectedToSection(ti)
		buildDisconnectButton(ti)
	} else {
		buildQuickConnectButton(ti)
	}

	buildConnectToItem(ti)
	systray.AddSeparator()
}

func buildVPNStatusLabel(ti *Instance) {
	if ti == nil {
		return
	}
	status := strings.ToLower(ti.state.vpnStatus.String())
	label := fmt.Sprintf(labelVPNStatus, status)
	mStatus := systray.AddMenuItem(label, label)
	mStatus.Disable()
}

func buildConnectedToSection(ti *Instance) {
	if ti == nil {
		return
	}
	if ti.state.vpnStatus != pb.ConnectionState_CONNECTED {
		return
	}

	if serverName := ti.state.serverName(); serverName != "" {
		label := fmt.Sprintf("%s %s", labelServerPrefix, serverName)
		mHostname := systray.AddMenuItem(label, label)
		mHostname.Disable()
	}

	if ti.state.vpnCity != "" {
		label := fmt.Sprintf("%s %s", labelCityPrefix, ti.state.vpnCity)
		mCity := systray.AddMenuItem(label, label)
		mCity.Disable()
	}

	if ti.state.vpnCountry != "" {
		label := fmt.Sprintf("%s %s", labelCountryPrefix, ti.state.vpnCountry)
		mCountry := systray.AddMenuItem(label, label)
		mCountry.Disable()
	}
}

func buildDisconnectButton(ti *Instance) {
	if ti == nil {
		return
	}
	item := systray.AddMenuItem(labelDisconnect, labelDisconnect)
	go handleDisconnectClick(ti, item)
}

func buildQuickConnectButton(ti *Instance) {
	if ti == nil {
		return
	}
	item := systray.AddMenuItem(labelQuickConnect, labelQuickConnect)
	go handleQuickConnectClick(ti, item)
}

func handleDisconnectClick(ti *Instance, item *systray.MenuItem) {
	if ti == nil {
		return
	}
	handleMenuItemClick(item, func() { ti.disconnect() })
}

func handleQuickConnectClick(ti *Instance, item *systray.MenuItem) {
	if ti == nil {
		return
	}
	handleMenuItemClick(item, func() { ti.connect("", "") })
}

func handleLogoutClick(ti *Instance, item *systray.MenuItem) {
	if ti == nil {
		return
	}
	handleMenuItemClickWithRetry(item, func() bool { return ti.logout(false) })
}

func handleLoginClick(ti *Instance, item *systray.MenuItem) {
	if ti == nil {
		return
	}
	handleMenuItemClick(item, func() { ti.login() })
}

func buildConnectToItem(ti *Instance) {
	if ti == nil {
		return
	}
	connectionSelector := systray.AddMenuItem(labelConnectionSelection, tooltipConnectionSelection)
	countries := slices.Clone(ti.state.connSelector.countries)
	specialtyServers := slices.Clone(ti.state.connSelector.specialtyServers)
	recentConnections := ti.recentConnections.GetRecentConnections()

	if len(recentConnections) > 0 {
		buildRecentConnectionsSection(ti, connectionSelector, recentConnections)
	}

	buildSpecialtyServersSection(ti, connectionSelector, specialtyServers)
	buildCountriesSection(ti, connectionSelector, countries)
}

func buildRecentConnectionsSection(
	ti *Instance,
	parent *systray.MenuItem,
	connections []RecentConnection,
) {
	if ti == nil || parent == nil {
		return
	}

	parent.AddSubMenuItem(labelRecentConnections, tooltipRecentConnections).Disable()
	for _, conn := range connections {
		displayLabel := makeDisplayLabel(&conn)
		if displayLabel == "" {
			continue
		}

		tooltip := fmt.Sprintf("%s %s", labelReconnectTo, displayLabel)
		item := parent.AddSubMenuItem(displayLabel, tooltip)

		go handleRecentConnectionClick(ti, item, &conn)
	}
}

func buildCountriesSection(ti *Instance, parent *systray.MenuItem, countries []string) {
	if ti == nil || parent == nil {
		return
	}

	parent.AddSubMenuItem(labelCountries, tooltipCountries).Disable()
	for _, country := range countries {
		title := strings.ReplaceAll(country, "_", " ")
		tooltip := fmt.Sprintf("%s %s", labelConnectTo, title)
		item := parent.AddSubMenuItem(title, tooltip)

		go handleCountryClick(ti, item, country)
	}
}

func buildSpecialtyServersSection(ti *Instance, parent *systray.MenuItem, specialtyServers []string) {
	if ti == nil || parent == nil {
		return
	}

	parent.AddSubMenuItem(labelSpecialtyServers, tooltipSpecialtyServers).Disable()
	for _, server := range specialtyServers {
		title := strings.ReplaceAll(server, "_", " ")
		tooltip := fmt.Sprintf("%s%s", labelConnectTo, title)
		item := parent.AddSubMenuItem(title, tooltip)

		go handleSpecialtyServerClick(ti, item, server)
	}
}

func handleRecentConnectionClick(ti *Instance, item *systray.MenuItem, model *RecentConnection) {
	if ti == nil || model == nil {
		return
	}
	handleMenuItemClick(item, func() { connectByConnectionModel(ti, model) })
}

func handleCountryClick(ti *Instance, item *systray.MenuItem, country string) {
	if ti == nil {
		return
	}
	handleMenuItemClick(item, func() { ti.connect(country, "") })
}

func handleSpecialtyServerClick(ti *Instance, item *systray.MenuItem, server string) {
	if ti == nil {
		return
	}

	handleMenuItemClick(item, func() { _ = ti.connect("", server) })
}

func buildAccountSection(ti *Instance) {
	if ti == nil {
		return
	}
	if !ti.state.daemonAvailable {
		return
	}

	systray.AddSeparator()
	if ti.state.loggedIn {
		if ti.state.accountName != "" {
			m := systray.AddMenuItem(labelLoggedInAs, tooltipLoggedInAs)
			m.Disable()

			item := systray.AddMenuItem(ti.state.accountName, ti.state.accountName)
			item.Disable()
		}

		item := systray.AddMenuItem(labelLogOut, tooltipLogOut)
		go handleLogoutClick(ti, item)
		return
	}

	loginTitle := systray.AddMenuItem(labelNotLoggedIn, tooltipNotLoggedIn)
	loginTitle.Disable()

	item := systray.AddMenuItem(labelLogIn, tooltipLogIn)
	go handleLoginClick(ti, item)
}

func buildSettingsSection(ti *Instance) {
	if ti == nil {
		return
	}
	if !ti.state.daemonAvailable {
		return
	}

	item := systray.AddMenuItem(labelSettings, tooltipSettings)
	buildSettingsSubitems(ti, item)
}

func buildSettingsSubitems(ti *Instance, menu *systray.MenuItem) {
	if ti == nil || menu == nil {
		return
	}

	ti.state.mu.RLock()
	notificationsEnabled := ti.state.notificationsStatus == Enabled
	trayEnabled := ti.state.trayStatus == Enabled
	ti.state.mu.RUnlock()

	notificationsCheckbox := menu.AddSubMenuItemCheckbox(
		labelNotifications,
		tooltipNotifications,
		notificationsEnabled,
	)
	trayCheckbox := menu.AddSubMenuItemCheckbox(
		labelTrayIcon,
		tooltipTrayIcon,
		trayEnabled,
	)

	go handleNotificationsOption(ti, notificationsCheckbox)
	go handleTrayOption(ti, trayCheckbox)
}

func handleTrayOption(ti *Instance, item *systray.MenuItem) {
	if ti == nil {
		return
	}
	handleCheckboxOption(ti, item, func(flag bool) bool { return ti.setTray(flag) })
}

func handleNotificationsOption(ti *Instance, item *systray.MenuItem) {
	if ti == nil {
		return
	}
	handleCheckboxOption(ti, item, func(flag bool) bool { return ti.setNotify(flag) })
}
