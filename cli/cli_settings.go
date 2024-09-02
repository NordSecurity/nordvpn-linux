package cli

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/client"
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
	settings, err := c.getSettings()
	if err != nil {
		return formatError(err)
	}

	fmt.Printf("Technology: %s\n", settings.GetTechnology())
	if settings.Technology == config.Technology_OPENVPN {
		fmt.Printf("Protocol: %s\n", settings.GetProtocol())
	}
	fmt.Printf("Firewall: %+v\n", nstrings.GetBoolLabel(settings.GetFirewall()))
	fmt.Printf("Firewall Mark: 0x%x\n", settings.GetFwmark())
	fmt.Printf("Routing: %+v\n", nstrings.GetBoolLabel(settings.GetRouting()))
	fmt.Printf("Analytics: %+v\n", nstrings.GetBoolLabel(settings.GetAnalytics()))
	fmt.Printf("Kill Switch: %+v\n", nstrings.GetBoolLabel(settings.GetKillSwitch()))
	fmt.Printf("Threat Protection Lite: %+v\n", nstrings.GetBoolLabel(settings.ThreatProtectionLite))
	if settings.Technology == config.Technology_OPENVPN {
		fmt.Printf("Obfuscate: %+v\n", nstrings.GetBoolLabel(settings.GetObfuscate()))
	}
	fmt.Printf("Notify: %+v\n", nstrings.GetBoolLabel(settings.UserSettings.Notify))
	fmt.Printf("Tray: %+v\n", nstrings.GetBoolLabel(settings.UserSettings.Tray))
	fmt.Printf("Auto-connect: %+v\n", nstrings.GetBoolLabel(settings.AutoConnectData.Enabled))
	if settings.AutoConnectData.Enabled && internal.IsDevEnv(string(c.environment)) {
		fmt.Printf("Auto-connect country: %s\n", settings.AutoConnectData.Country)
		fmt.Printf("Auto-connect city: %s\n", settings.AutoConnectData.City)
		fmt.Printf("Auto-connect group: %s\n", settings.AutoConnectData.ServerGroup)
	}

	fmt.Printf("IPv6: %+v\n", nstrings.GetBoolLabel(settings.Ipv6))
	fmt.Printf("Meshnet: %+v\n", nstrings.GetBoolLabel(settings.Meshnet))
	if len(settings.Dns) == 0 {
		fmt.Printf("DNS: %+v\n", nstrings.GetBoolLabel(false))
	} else {
		fmt.Printf("DNS: %+v\n", strings.Join(settings.Dns, ", "))
	}
	fmt.Printf("LAN Discovery: %+v\n", nstrings.GetBoolLabel(settings.LanDiscovery))
	fmt.Printf("Virtual Location: %+v\n", nstrings.GetBoolLabel(settings.VirtualLocation))
	if settings.Technology == config.Technology_NORDLYNX {
		fmt.Printf("Post-quantum VPN: %+v\n", nstrings.GetBoolLabel(settings.PostquantumVpn))
	}

	displayAllowlist(settings.Allowlist)
	return nil
}

func (c *cmd) getSettings() (*pb.Settings, error) {
	resp, err := c.client.Settings(context.Background(), &pb.Empty{})
	if err != nil {
		return nil, err
	}
	switch resp.Type {
	case internal.CodeConfigError:
		return nil, ErrConfig
	case internal.CodeSuccess:
		return resp.GetData(), nil
	default:
		return nil, internal.ErrUnhandled
	}
}

func displayAllowlist(allowlist *pb.Allowlist) {
	if allowlist != nil {
		udpPorts := allowlist.GetPorts().GetUdp()
		tcpPorts := allowlist.GetPorts().GetTcp()
		if len(udpPorts)+len(tcpPorts) > 0 {
			allPorts := append(udpPorts, tcpPorts...)
			sort.Slice(allPorts, func(i, j int) bool {
				return client.InterfaceToInt64(allPorts[i]) < client.InterfaceToInt64(allPorts[j])
			})
			allowlistedRanges := make([]PortRange, 0)
			for _, port := range allPorts {
				//find current iteration's protocols
				var protos []string
				if index := slices.Index(udpPorts, port); index != -1 {
					protos = append(protos, "UDP")
				}
				if index := slices.Index(tcpPorts, port); index != -1 {
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
		subnets := allowlist.GetSubnets()
		if len(subnets) > 0 {
			fmt.Printf("Allowlisted subnets:\n")
			for _, subnet := range subnets {
				fmt.Printf("\t%s\n", subnet)
			}
		}
	}
}
