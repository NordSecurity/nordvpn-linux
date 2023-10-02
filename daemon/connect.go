package daemon

import (
	"io/ioutil"
	"log"
	"net/netip"
	"os/exec"
	"regexp"
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
	allowlist config.Allowlist,
	nameservers []string,
	netw networker.Networker,
) {
	defer close(events)

	events <- ConnectEvent{Code: internal.CodeConnecting}

	err := netw.Start(
		creds,
		serverData,
		allowlist,
		nameservers,
		false, // here vpn connect, no need route to remote peer's LAN
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

// maskIPRouteOutput changes any non-local ip address in the output to ***
func maskIPRouteOutput(output string) string {
	expIPv4 := regexp.MustCompile(`\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}`)
	expIPv6 := regexp.MustCompile(`(([0-9a-fA-F]{1,4}:){7,7}[0-9a-fA-F]{1,4}` +
		`|([0-9a-fA-F]{1,4}:){1,7}:` +
		`|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}` +
		`|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}` +
		`|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}` +
		`|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}` +
		`|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}` +
		`|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})` +
		`|:((:[0-9a-fA-F]{1,4}){1,7}|:)` +
		`|fe80:(:[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]{1,}` +
		`|::(ffff(:0{1,4}){0,1}:){0,1}((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])` +
		`|([0-9a-fA-F]{1,4}:){1,4}:((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9]))`)

	ips := expIPv4.FindAllString(output, -1)
	ips = append(ips, expIPv6.FindAllString(output, -1)...)
	for _, ip := range ips {
		parsed, err := netip.ParseAddr(ip)
		if err != nil {
			log.Printf("Failed to parse ip address %s for masking: %v", ip, err)
			continue
		}

		if !parsed.IsLinkLocalMulticast() && !parsed.IsLinkLocalUnicast() && !parsed.IsLoopback() && !parsed.IsPrivate() {
			output = strings.Replace(output, ip, "***", -1)
		}
	}

	return output
}

func getNetworkInfo() string {
	builder := strings.Builder{}
	for _, arg := range []string{"4", "6"} {
		// #nosec G204 -- arg values are known before even running the program
		out, err := exec.Command("ip", "-"+arg, "route", "show", "table", "all").CombinedOutput()
		if err != nil {
			continue
		}
		maskedOutput := maskIPRouteOutput(string(out))
		builder.WriteString("Routes for ipv" + arg + ":\n")
		builder.WriteString(maskedOutput)

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
