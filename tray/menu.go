package tray

import (
	"os/exec"
	"strings"

	"github.com/NordSecurity/systray"
)

func addDebugSection() {
	systray.AddSeparator()
	mRedraw := systray.AddMenuItem("Redraw", "Redraw")
	go func() {
		for {
			<-mRedraw.ClickedCh
			redrawChan <- struct{}{}
		}
	}()
	mUpdate := systray.AddMenuItem("Update", "Update")
	go func() {
		for {
			<-mUpdate.ClickedCh
			updateChan <- false
		}
	}()
	mUpdateFull := systray.AddMenuItem("Full update", "Full update")
	go func() {
		for {
			<-mUpdateFull.ClickedCh
			updateChan <- true
		}
	}()
}

func addQuitItem() {
	systray.AddSeparator()
	m := systray.AddMenuItem("Quit", "Quit")
	m.Enable()
	go func() {
		<-m.ClickedCh
		systray.Quit()
	}()
}

func addAppSection() {
	m := systray.AddMenuItem("NordVPN", "NordVPN")
	go func() {
		for {
			<-m.ClickedCh
			cmd := exec.Command("zenity", "--info", "--text", "This will be the NordVPN GUI", "--no-wrap")
			err := cmd.Start()
			if err != nil {
				notification("error", "Failed to start NordVPN GUI: %s", err)
			} else {
				err = cmd.Wait()
			}
		}
	}()
}

func addDaemonSection() {
	systray.AddSeparator()
	m := systray.AddMenuItem("Daemon not available", "Daemon not available")
	m.Disable()

	mError := systray.AddMenuItem(state.daemonError, state.daemonError)
	mError.Disable()
}

func addVpnSection() {
	systray.AddSeparator()

	mStatus := systray.AddMenuItem("VPN "+strings.ToLower(state.vpnStatus), "VPN "+strings.ToLower(state.vpnStatus))
	mStatus.Disable()

	if state.vpnStatus == "Connected" {
		mHostname := systray.AddMenuItem("Server: "+state.vpnHostname, "Server: "+state.vpnHostname)
		mHostname.Disable()
		mCity := systray.AddMenuItem("City: "+state.vpnCity, "City: "+state.vpnCity)
		mCity.Disable()
		mCountry := systray.AddMenuItem("Country: "+state.vpnCountry, "Country: "+state.vpnCountry)
		mCountry.Disable()
		mDisconnect := systray.AddMenuItem("Disconnect", "Disconnect")
		go func() {
			success := false
			for !success {
				<-mDisconnect.ClickedCh
				success = disconnect(Client)
			}
			updateChan <- true
		}()
	} else {
		mConnect := systray.AddMenuItem("Quick Connect", "Quick Connect")
		go func() {
			success := false
			for !success {
				<-mConnect.ClickedCh
				success = connect(Client, "", "")
			}
			updateChan <- true
		}()
	}
}

// nolint:unused
func addMeshnetSection() {
	systray.AddSeparator()

	status := ""
	if state.meshnetEnabled {
		status = "enabled"
	} else {
		status = "disabled"
	}
	mStatus := systray.AddMenuItem("Meshnet "+status, "Meshnet "+status)
	mStatus.Disable()

	if state.meshnetEnabled {
		mDisconnect := systray.AddMenuItem("Disable", "Disable")
		go func() {
			success := false
			for !success {
				<-mDisconnect.ClickedCh
				success = disableMeshnet(MeshClient)
			}
			updateChan <- true
		}()
	} else {
		mConnect := systray.AddMenuItem("Enable", "Enable")
		go func() {
			success := false
			for !success {
				<-mConnect.ClickedCh
				success = enableMeshnet(MeshClient)
			}
			updateChan <- true
		}()
	}
}

func addAccountSection() {
	systray.AddSeparator()

	if state.loggedIn {
		m := systray.AddMenuItem("Logged in as:", "Logged in as:")
		m.Disable()

		mName := systray.AddMenuItem(state.accountName, state.accountName)
		mName.Disable()

		mLogout := systray.AddMenuItem("Log out", "Log out")

		go func() {
			success := false
			for !success {
				<-mLogout.ClickedCh
				success = logout(Client, false)
			}
			updateChan <- true
		}()
	} else {
		m := systray.AddMenuItem("Not logged in", "Not logged in")
		m.Disable()

		mLogin := systray.AddMenuItem("Log in", "Log in")

		go func() {
			for {
				<-mLogin.ClickedCh
				login(Client)
			}
		}()
	}
}
