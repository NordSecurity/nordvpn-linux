package tray

import (
	"fmt"
	"runtime"
	"slices"
	"strings"
	"time"

	"github.com/NordSecurity/systray"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/log"
	"github.com/NordSecurity/nordvpn-linux/norduser"
)

const (
	// Menu item labels
	labelServerPrefix          = "Server:"
	labelCityPrefix            = "City:"
	labelCountryPrefix         = "Country:"
	labelDisconnect            = "Disconnect"
	labelPause                 = "Pause"
	labelSecureMyConnection    = "Secure my connection"
	labelConnectionSelection   = "All connections"
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
	labelLoggedInAs            = "Account"
	labelLogOut                = "Log out"
	labelNotLoggedIn           = "Not logged in"
	labelLogIn                 = "Log in"
	labelSettings              = "Settings"
	labelOpenGui               = "Open NordVPN app"
	labelDownloadGui           = "Download NordVPN app"
	labelNotifications         = "Notifications"
	labelTrayIcon              = "Tray icon"
	labelPause5Min             = "Pause for 5 minutes"
	labelPause15Min            = "Pause for 15 minutes"
	labelPause30Min            = "Pause for 30 minutes"
	labelPause1H               = "Pause for 1 hour"
	labelPause24H              = "Pause for 24 hours"

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
	tooltipOpenGui             = "Open the NordVPN app"
	tooltipDownloadGui         = "Download the NordVPN app"
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

	ti.checkboxSync.HandleCheckboxOption(newSystrayMenuItemAdapter(item), setter)
}

// systrayMenuItemAdapter wraps a *systray.MenuItem to satisfy the CheckableMenuItem interface.
type systrayMenuItemAdapter struct {
	item *systray.MenuItem
}

func newSystrayMenuItemAdapter(item *systray.MenuItem) *systrayMenuItemAdapter {
	return &systrayMenuItemAdapter{item: item}
}

func (a *systrayMenuItemAdapter) ClickedCh() <-chan struct{} {
	return a.item.ClickedCh
}

func (a *systrayMenuItemAdapter) Checked() bool {
	return a.item.Checked()
}

func (a *systrayMenuItemAdapter) Check() {
	a.item.Check()
}

func (a *systrayMenuItemAdapter) Uncheck() {
	a.item.Uncheck()
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
		log.Info(msgShutdownNotification)
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

	if ti.state.vpnStatus == pb.ConnectionState_PAUSED {
		buildPauseTimer(ti)
	}

	if ti.state.vpnStatus == pb.ConnectionState_CONNECTED {
		buildConnectedToSection(ti)
		if ti.state.vpnIsMeshPeer {
			buildDisconnectButton(ti)
		} else {
			buildPauseMenu(ti)
		}
	} else {
		buildQuickConnectButton(ti)
	}

	buildConnectToItem(ti)
	systray.AddSeparator()
}

func vpnStateToStatusLabel(state pb.ConnectionState) string {
	switch state {
	case pb.ConnectionState_CONNECTED:
		return "Secured"
	case pb.ConnectionState_CONNECTING:
		return "Connecting…"
	case pb.ConnectionState_UNKNOWN_STATE:
		fallthrough
	case pb.ConnectionState_PAUSED:
		fallthrough
	case pb.ConnectionState_DISCONNECTED:
		fallthrough
	default:
		return "Not secured"
	}
}

func buildVPNStatusLabel(ti *Instance) {
	if ti == nil {
		return
	}
	label := vpnStateToStatusLabel(ti.state.vpnStatus)
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
	go handleDisconnectClick(ti, item, pb.UIEvent_DISCONNECT, pb.UIEvent_ITEM_VALUE_UNSPECIFIED)
}

func buildQuickConnectButton(ti *Instance) {
	if ti == nil {
		return
	}
	item := systray.AddMenuItem(labelSecureMyConnection, labelSecureMyConnection)
	go handleQuickConnectClick(ti, item)
}

type pauseLength struct {
	Name            string
	Tooltip         string
	DurationSeconds uint32
	EventValue      pb.UIEvent_ItemValue
}

var pauseLengths = []pauseLength{
	{
		Name:            labelPause5Min,
		Tooltip:         labelPause5Min,
		DurationSeconds: 5 * 60,
		EventValue:      pb.UIEvent_PAUSE_5_MIN,
	},
	{
		Name:            labelPause15Min,
		Tooltip:         labelPause15Min,
		DurationSeconds: 15 * 60,
		EventValue:      pb.UIEvent_PAUSE_15_MIN,
	},
	{
		Name:            labelPause30Min,
		Tooltip:         labelPause30Min,
		DurationSeconds: 30 * 60,
		EventValue:      pb.UIEvent_PAUSE_30_MIN,
	},
	{
		Name:            labelPause1H,
		Tooltip:         labelPause1H,
		DurationSeconds: 60 * 60,
		EventValue:      pb.UIEvent_PAUSE_1_HOUR,
	},
	{
		Name:            labelPause24H,
		Tooltip:         labelPause24H,
		DurationSeconds: 24 * 60 * 60,
		EventValue:      pb.UIEvent_PAUSE_24_HOURS,
	},
}

func buildPauseMenu(ti *Instance) {
	if ti == nil {
		return
	}

	pauseMenu := systray.AddMenuItem(labelPause, labelPause)
	for _, pauseLength := range pauseLengths {
		pause := pauseMenu.AddSubMenuItem(pauseLength.Name, pauseLength.Tooltip)
		go handlePauseClick(ti, pause, pauseLength)
	}

	disconnect := pauseMenu.AddSubMenuItem(labelDisconnect, labelDisconnect)
	go handleDisconnectClick(ti, disconnect, pb.UIEvent_PAUSE, pb.UIEvent_PAUSE_DISCONNECT)
}

func buildPauseTimer(ti *Instance) {
	initialValue := ti.state.pauseRemainingSec
	if initialValue <= 0 {
		return
	}

	timer := systray.AddMenuItem(
		buildTimerString(initialValue), "",
	)
	timer.Disable()

	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		currentValue := initialValue
		for currentValue > 0 {
			select {
			case _, open := <-timer.ClickedCh:
				if !open {
					return
				}
			case <-ticker.C:
				ti.state.mu.Lock()

				currentValue = ti.state.pauseRemainingSec
				if currentValue > 0 {
					currentValue--
					ti.state.pauseRemainingSec = currentValue
				}

				ti.state.mu.Unlock()

				if ti.isVisible.Load() {
					timer.SetTitleQuiet(
						buildTimerString(currentValue),
					)
				}
			}
		}
	}()
}

func buildTimerString(remaining int) string {
	hours := remaining / 3600
	minutes := (remaining % 3600) / 60
	seconds := remaining % 60

	if hours > 0 {
		return fmt.Sprintf("VPN connection resumes in %02d:%02d:%02d", hours, minutes, seconds)
	} else {
		return fmt.Sprintf("VPN connection resumes in %02d:%02d", minutes, seconds)
	}
}

func handlePauseClick(ti *Instance, item *systray.MenuItem, pauseLength pauseLength) {
	if ti == nil {
		return
	}
	handleMenuItemClick(item, func() { ti.pause(pauseLength) })
}

func handleDisconnectClick(ti *Instance, item *systray.MenuItem, itemName pb.UIEvent_ItemName, itemValue pb.UIEvent_ItemValue) {
	if ti == nil {
		return
	}
	handleMenuItemClick(item, func() { ti.disconnect(itemName, itemValue) })
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

		displayLabel = tryApplyVirtualLocationSuffix(displayLabel, conn.VirtualLocation)
		tooltip := fmt.Sprintf("%s %s", labelReconnectTo, displayLabel)
		item := parent.AddSubMenuItem(displayLabel, tooltip)

		go handleRecentConnectionClick(ti, item, &conn)
	}
}

func buildCountriesSection(ti *Instance, parent *systray.MenuItem, countries []Server) {
	if ti == nil || parent == nil {
		return
	}

	parent.AddSubMenuItem(labelCountries, tooltipCountries).Disable()
	for _, country := range countries {
		title := country.displayLabel
		tooltip := fmt.Sprintf("%s %s", labelConnectTo, title)
		item := parent.AddSubMenuItem(title, tooltip)

		go handleCountryClick(ti, item, country.name)
	}
}

func buildSpecialtyServersSection(ti *Instance, parent *systray.MenuItem, specialtyServers []Server) {
	if ti == nil || parent == nil {
		return
	}

	parent.AddSubMenuItem(labelSpecialtyServers, tooltipSpecialtyServers).Disable()
	for _, server := range specialtyServers {
		title := server.displayLabel
		tooltip := fmt.Sprintf("%s%s", labelConnectTo, title)
		item := parent.AddSubMenuItem(title, tooltip)

		go handleSpecialtyServerClick(ti, item, server.name)
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
	itemValue := pb.UIEvent_ITEM_VALUE_UNSPECIFIED
	serverSearchStr := strings.ToLower(server)
	if group, ok := config.GroupMap[serverSearchStr]; ok {
		itemValue = ItemValueFromServerGroup(group)
	}
	handleMenuItemClick(item, func() {
		ti.connectWithUIEvent("", server, pb.UIEvent_CONNECT, itemValue)
	})
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

	notificationsEnabled := ti.state.notificationsStatus == Enabled
	trayEnabled := ti.state.trayStatus == Enabled

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

func buildGuiSection(ti *Instance) {
	if ti == nil {
		return
	}
	if !ti.state.daemonAvailable {
		return
	}

	systray.AddSeparator()
	if isGuiAvailable() {
		item := systray.AddMenuItem(labelOpenGui, tooltipOpenGui)
		handleMenuItemClick(item, ti.openGui)
	} else {
		item := systray.AddMenuItem(labelDownloadGui, tooltipDownloadGui)
		handleMenuItemClick(item, ti.openDownloadPage)
	}
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
