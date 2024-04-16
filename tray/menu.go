package tray

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/NordSecurity/nordvpn-linux/norduser"
	"github.com/NordSecurity/systray"
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
		systray.Quit()
		select {
		case ti.quitChan <- norduser.StopRequest{}:
		default:
		}
	}()
}

func addDaemonSection(ti *Instance) {
	systray.AddSeparator()
	m := systray.AddMenuItem("Daemon not available", "Daemon not available")
	m.Disable()

	mError := systray.AddMenuItem(ti.state.daemonError, ti.state.daemonError)
	mError.Disable()
}

func addVpnSection(ti *Instance) {
	systray.AddSeparator()

	mStatus := systray.AddMenuItem("VPN "+strings.ToLower(ti.state.vpnStatus), "VPN "+strings.ToLower(ti.state.vpnStatus))
	mStatus.Disable()

	if ti.state.vpnStatus == "Connected" {
		mHostname := systray.AddMenuItem("Server: "+ti.state.vpnHostname, "Server: "+ti.state.vpnHostname)
		mHostname.Disable()
		mCity := systray.AddMenuItem("City: "+ti.state.vpnCity, "City: "+ti.state.vpnCity)
		mCity.Disable()
		mCountry := systray.AddMenuItem("Country: "+ti.state.vpnCountry, "Country: "+ti.state.vpnCountry)
		mCountry.Disable()
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
}

func addAccountSection(ti *Instance) {
	systray.AddSeparator()

	if ti.state.loggedIn {
		m := systray.AddMenuItem("Logged in as:", "Logged in as:")
		m.Disable()

		mName := systray.AddMenuItem(ti.state.accountName, ti.state.accountName)
		mName.Disable()

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
