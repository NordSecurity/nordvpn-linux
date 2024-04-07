package openvpn

import (
	"bytes"
	"errors"
	"fmt"
	"net/netip"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/internal"

	"github.com/NordSecurity/gokogiri/xml"
	"github.com/NordSecurity/ratago/xslt"
)

const ovpnConfig = `<?xml version="1.0"?>
<?xml-stylesheet type="xml/xsl"?>
<config>
  <ips>
    <ip address="{{.Address}}" />
  </ips>
  <technology identifier="{{.Identifier}}"/>
</config>
`

type openvpnID string

const (
	techXORUDP openvpnID = "openvpn_xor_udp"
	techUDP    openvpnID = "openvpn_udp"
	techXORTCP openvpnID = "openvpn_xor_tcp"
	techTCP    openvpnID = "openvpn_tcp"

	interfaceType = "tun"
	InterfaceName = "nordtun"
)

var (
	// openVPNConfigFileName is a name of a openvpn config file to be used for connecting to a VPN
	openVPNConfigFileName = filepath.Join(internal.DatFilesPathCommon, ".config.ovpn")

	// openVPNExec defines openvpn executable path
	openVPNExec = filepath.Join(internal.AppDataPathStatic, "openvpn")
)

type ovpnConfigData struct {
	Address    string
	Identifier string
}

// setOpenVPNConfig is used to pass generated config to the OpenVPN process.
// Config has to be passed everytime when new OpenVPN process is started.
func setOpenVPNConfig(protocol config.Protocol, serverIP netip.Addr, obfuscated bool, serverVersion string) error {
	if serverVersion == "" {
		return ErrServerVersion
	}
	return generateConfigFile(protocol, serverIP, obfuscated)
}

func generateConfigFile(protocol config.Protocol, serverIP netip.Addr, obfuscated bool) error {
	templatePath := internal.OvpnTemplatePath
	if obfuscated {
		templatePath = internal.OvpnObfsTemplatePath
	}

	identifier, err := getConfigIdentifier(protocol, obfuscated)
	if err != nil {
		return fmt.Errorf("getting config identifier: %w", err)
	}

	template, err := internal.FileRead(templatePath)
	if err != nil {
		return fmt.Errorf("reading ovpn template file")
	}

	out, err := generateConfig(serverIP, identifier, template)
	if err != nil {
		return fmt.Errorf("generating OpenVPN config: %w", err)
	}

	if err := addExtraParameters(out, serverIP, protocol); err != nil {
		return fmt.Errorf("adding extra parameters to OpenVPN config: %w", err)
	}

	if internal.FileExists(openVPNConfigFileName) {
		if err := internal.FileUnlock(openVPNConfigFileName); err != nil {
			return err
		}
		return internal.FileWrite(openVPNConfigFileName, out, internal.PermUserRW)
	}

	ovpnConfig, err := internal.FileCreate(openVPNConfigFileName, internal.PermUserRW)
	if err != nil {
		return fmt.Errorf("creating OpenVPN config file: %w", err)
	}

	_, err = ovpnConfig.Write(out)
	if err != nil {
		// #nosec G104 -- errors.Join would be useful here
		ovpnConfig.Close()
		return err
	}
	return ovpnConfig.Close()
}

func generateConfig(serverIP netip.Addr, identifier openvpnID, template []byte) ([]byte, error) {
	xmlConfig, err := generateConfigXML(serverIP, identifier)
	if err != nil {
		return nil, fmt.Errorf("generating config XML file: %w", err)
	}

	xmlDoc, err := xml.Parse(xmlConfig, nil, nil, 0, nil)
	if err != nil {
		return nil, fmt.Errorf("parsing XML config: %w", err)
	}

	sheetXMLDoc, err := xml.Parse(template, nil, nil, 0, nil)
	if err != nil {
		return nil, fmt.Errorf("parsing XML template file: %w", err)
	}

	// OpenVPN Templates are single files, therefore fileurl can be empty
	sheet, err := xslt.ParseStylesheet(sheetXMLDoc, "")
	if err != nil {
		return nil, fmt.Errorf("parsing stylesheet: %w", err)
	}
	out, err := sheet.Process(xmlDoc, xslt.StylesheetOptions{})
	if err != nil {
		return nil, fmt.Errorf("Processing XML config: %w", err)
	}
	return []byte(disableEscaping(out)), nil
}

func disableEscaping(out string) string {
	out = strings.ReplaceAll(out, "&lt;", "<")
	out = strings.ReplaceAll(out, "&gt;", ">")
	out = strings.ReplaceAll(out, "&quot;", "\"")
	out = strings.ReplaceAll(out, "&apos;", "'")
	out = strings.ReplaceAll(out, "&amp;", "&")
	return out
}

func generateConfigXML(serverIP netip.Addr, identifier openvpnID) ([]byte, error) {
	var out bytes.Buffer
	configTemplate := template.Must(template.New("ovpnConfig").Parse(ovpnConfig))
	fileConfig := ovpnConfigData{
		Address:    serverIP.String(),
		Identifier: string(identifier),
	}
	err := configTemplate.Execute(&out, fileConfig)
	return out.Bytes(), err
}

func getConfigIdentifier(protocol config.Protocol, obfuscated bool) (openvpnID, error) {
	switch protocol {
	case config.Protocol_UDP:
		if obfuscated {
			return techXORUDP, nil
		}
		return techUDP, nil
	case config.Protocol_TCP:
		if obfuscated {
			return techXORTCP, nil
		}
		return techTCP, nil
	case config.Protocol_UNKNOWN_PROTOCOL:
		fallthrough
	default:
		return "", errors.New("unknown protocol")
	}
}

func addExtraParameters(data []byte, serverIP netip.Addr, protocol config.Protocol) error {
	args := strings.Split(string(data), "\n")
	if !serverIP.Is6() {
		args = addOrReplaceArgument(args, "pull-filter ignore \"ifconfig-ipv6\"", "pull-filter ignore \"ifconfig-ipv6\".*$")
	}
	args = addOrReplaceArgument(args, "pull-filter ignore \"route-ipv6\"", "pull-filter ignore \"route-ipv6\".*$")
	args = addOrReplaceArgument(args, "ping 15", "ping .*$")
	args = addOrReplaceArgument(args, "ping-restart 0", "ping-restart .*$")
	args = addOrReplaceArgument(args, "ping-timer-rem", "ping-timer-rem$")
	// override openvpn proto (obfuscated sets multiple remotes)
	if serverIP.Is6() {
		switch protocol {
		case config.Protocol_UDP:
			args = addOrReplaceArgument(args, "proto udp6", "proto udp6$")
		case config.Protocol_TCP:
			args = addOrReplaceArgument(args, "proto tcp6", "proto tcp6$")
		case config.Protocol_UNKNOWN_PROTOCOL:
			fallthrough
		default:
			return errors.New("unknown protocol")
		}
	}
	data = []byte(strings.Join(args, "\n"))
	return nil
}

func addOrReplaceArgument(args []string, newArg string, regex string) []string {
	index := -1
	reg, _ := regexp.Compile(regex)
	for idx, arg := range args {
		if reg.MatchString(arg) {
			index = idx
			args[idx] = newArg
		}
	}
	if index == -1 {
		args = append(args, newArg)
	}
	return args
}
