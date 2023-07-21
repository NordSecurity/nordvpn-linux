package cli

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/client"
	cconfig "github.com/NordSecurity/nordvpn-linux/client/config"
	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/nstrings"

	"github.com/urfave/cli/v2"
	"golang.org/x/exp/slices"
)

// SettingsUsageText is show next to settings command by nordvpn --help
const SettingsUsageText = "Shows current settings"

type PortRange struct {
	start     int64
	end       int64
	protocols []string
}

func (c *cmd) Settings(ctx *cli.Context) error {
	resp, err := c.client.Settings(context.Background(), &pb.SettingsRequest{
		Uid: int64(os.Getuid()),
	})
	if err != nil {
		return formatError(err)
	}

	switch resp.Type {
	case internal.CodeConfigError:
		return formatError(ErrConfig)
	case internal.CodeSuccess:
		break
	default:
		return formatError(internal.ErrUnhandled)
	}

	fmt.Printf("Technology: %s\n", resp.Data.GetTechnology())
	if resp.Data.Technology == config.Technology_OPENVPN {
		fmt.Printf("Protocol: %s\n", resp.Data.GetProtocol())
	}
	fmt.Printf("Firewall: %+v\n", nstrings.GetBoolLabel(resp.Data.GetFirewall()))
	fmt.Printf("Firewall Mark: 0x%x\n", resp.Data.GetFwmark())
	fmt.Printf("Routing: %+v\n", nstrings.GetBoolLabel(resp.Data.GetRouting()))
	fmt.Printf("Analytics: %+v\n", nstrings.GetBoolLabel(resp.Data.GetAnalytics()))
	fmt.Printf("Kill Switch: %+v\n", nstrings.GetBoolLabel(resp.Data.GetKillSwitch()))
	fmt.Printf("Threat Protection Lite: %+v\n", nstrings.GetBoolLabel(resp.Data.ThreatProtectionLite))
	if resp.Data.Technology == config.Technology_OPENVPN {
		fmt.Printf("Obfuscate: %+v\n", nstrings.GetBoolLabel(c.config.Obfuscate))
	}
	fmt.Printf("Notify: %+v\n", nstrings.GetBoolLabel(resp.Data.Notify))
	fmt.Printf("Auto-connect: %+v\n", nstrings.GetBoolLabel(resp.Data.AutoConnect))
	fmt.Printf("IPv6: %+v\n", nstrings.GetBoolLabel(resp.Data.Ipv6))
	fmt.Printf("Meshnet: %+v\n", nstrings.GetBoolLabel(resp.Data.Meshnet))
	if len(resp.Data.Dns) == 0 {
		fmt.Printf("DNS: %+v\n", nstrings.GetBoolLabel(false))
	} else {
		fmt.Printf("DNS: %+v\n", strings.Join(resp.Data.Dns, ", "))
	}
	fmt.Printf("LAN Discovery: %+v\n", nstrings.GetBoolLabel(resp.Data.LanDiscovery))

	displayAllowlist(&c.config.Allowlist)
	return nil
}

func displayAllowlist(allowlist *cconfig.Allowlist) {
	if allowlist != nil {
		udpPorts := allowlist.Ports.UDP.ToSlice()
		tcpPorts := allowlist.Ports.TCP.ToSlice()
		if len(udpPorts)+len(tcpPorts) > 0 {
			allPorts := allowlist.Ports.UDP.Union(allowlist.Ports.TCP).ToSlice()
			sort.Slice(allPorts, func(i, j int) bool {
				return client.InterfaceToInt64(allPorts[i]) < client.InterfaceToInt64(allPorts[j])
			})
			allowlistedRanges := make([]PortRange, 0)
			for _, port := range allPorts {
				//find current iteration's protocols
				var protos []string
				if allowlist.Ports.UDP.Contains(port) {
					protos = append(protos, "UDP")
				}
				if allowlist.Ports.TCP.Contains(port) {
					protos = append(protos, "TCP")
				}

				var lastProtos []string
				var lastEndPort int64
				if len(allowlistedRanges) > 0 {
					last := allowlistedRanges[len(allowlistedRanges)-1]
					lastProtos = last.protocols
					lastEndPort = last.end
				}
				//check if the range allowlist range continues or should we be starting a new one
				if !slices.Equal(protos, lastProtos) || client.InterfaceToInt64(port)-lastEndPort > 1 {
					allowlistedRanges = append(allowlistedRanges, PortRange{start: client.InterfaceToInt64(port), protocols: protos})
				}
				//populate the range
				allowlistedRanges[len(allowlistedRanges)-1].end = client.InterfaceToInt64(port)
			}
			fmt.Printf("Allowlisted ports:\n")
			maxLength := len(strconv.FormatInt(client.InterfaceToInt64(allPorts[len(allPorts)-1]), 10))
			for _, wlRange := range allowlistedRanges {
				protoString := strings.Join(wlRange.protocols, "|")
				if wlRange.start == wlRange.end {
					fmt.Printf("  %*d (%s)\n", maxLength*2+3, wlRange.start, protoString)
				} else {
					fmt.Printf("  %*d - %*d (%s)\n", maxLength, wlRange.start, maxLength, wlRange.end, protoString)
				}
			}
		}
		subnets := allowlist.Subnets.ToSlice()
		if len(subnets) > 0 {
			fmt.Printf("Allowlisted subnets:\n")
			for _, subnet := range subnets {
				fmt.Printf("\t%s\n", subnet)
			}
		}
	}
}
