package daemon

import (
	"io/ioutil"
	"os/exec"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/networker"
)

type ConnectEvent struct {
	Code    int64
	Message string
}

func Connect(
	events chan ConnectEvent,
	creds vpn.Credentials,
	serverData vpn.ServerData,
	whitelist config.Whitelist,
	nameservers []string,
	netw networker.Networker,
) {
	defer close(events)

	events <- ConnectEvent{Code: internal.CodeConnecting}

	err := netw.Start(
		creds,
		serverData,
		whitelist,
		nameservers,
	)
	switch err {
	case vpn.ErrVPNAIsAlreadyStarted:
		events <- ConnectEvent{Code: internal.CodeDisconnected}
	case nil:
		events <- ConnectEvent{Code: internal.CodeConnected}
	default:
		events <- ConnectEvent{
			Code:    internal.CodeFailure,
			Message: err.Error(),
		}
		return
	}
}

func getSystemInfo(version string) string {
	builder := strings.Builder{}
	builder.WriteString("App Version: " + version + "\n")
	out, err := ioutil.ReadFile("/etc/os-release")
	if err == nil {
		builder.WriteString("OS Info:\n" + string(out) + "\n")
	}
	out, err = exec.Command("uname", "-a").CombinedOutput()
	if err == nil {
		builder.WriteString("System Info:" + string(out) + "\n")
	}
	return builder.String()
}

func getNetworkInfo() string {
	builder := strings.Builder{}
	for _, arg := range []string{"4", "6"} {
		// #nosec G204 -- arg values are known before even running the program
		out, err := exec.Command("ip", "-"+arg, "route", "show", "table", "all").CombinedOutput()
		if err != nil {
			continue
		}
		builder.WriteString("Routes for ipv" + arg + ":\n" + string(out) + "\n")

		// #nosec G204 -- arg values are known before even running the program
		out, err = exec.Command("ip", "-"+arg, "rule").CombinedOutput()
		if err != nil {
			continue
		}
		builder.WriteString("IP rules for ipv" + arg + ":\n" + string(out) + "\n")
	}

	for _, iptableVersion := range internal.GetSupportedIPTables() {
		tableRules := ""
		for _, table := range []string{"filter", "nat", "mangle", "raw", "security"} {
			// #nosec G204 -- input is properly sanitized
			out, err := exec.Command(iptableVersion, "-S", "-t", table).CombinedOutput()
			if err == nil {
				tableRules += table + ":\n" + string(out) + "\n"
			}
		}
		version := "4"
		if iptableVersion == "ip6tables" {
			version = "6"
		}
		builder.WriteString("IP tables for ipv" + version + ":\n" + tableRules)
	}

	return builder.String()
}
