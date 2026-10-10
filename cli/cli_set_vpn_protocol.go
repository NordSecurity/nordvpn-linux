package cli

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

// Set protocol help text
const (
	SetVPNProtocolUsageText     = "Sets the VPN protocol"
	SetVPNProtocolArgsUsageText = `<protocol>`
	SetVPNProtocolDescription   = `Use this command to set the protocol for VPN connection.
Supported values for <protocol>: %s

Example: 'nordvpn set protocol %s'`
)

func buildVPNProtocolDescription() string {
	args := vpnProtocolArgs()
	return fmt.Sprintf(SetVPNProtocolDescription, strings.Join(args, ", "), args[0])
}

func (c *cmd) SetVPNProtocol(ctx *cli.Context) error {
	switch ctx.NArg() {
	case 0:
		return formatError(argsCountError(ctx))
	case 1:
	default:
		return formatError(argsParseError(ctx))
	}

	vpnProto, ok := vpnProtocolFromArg(ctx.Args().First())
	if !ok {
		return formatError(argsParseError(ctx))
	}

	req := &pb.SetVPNProtocolRequest{VpnProtocol: vpnProto}
	resp, err := c.client.SetVPNProtocol(context.Background(), req)
	if err != nil {
		return formatError(err)
	}

	switch resp.Type {
	case internal.CodeFeatureHidden:
		return formatError(argsParseError(ctx))
	case internal.CodeNothingToDo:
		color.Yellow(fmt.Sprintf(MsgAlreadySet, "VPN Protocol", resp.Data[0]))
	case internal.CodeConfigError:
		return formatError(ErrConfig)
	case internal.CodeProtocolIncompatibleWithAutoconnect:
		return formatError(fmt.Errorf(MsgIncompatibleTechWithAutoconnect, vpnProto.DisplayName()))
	case internal.CodeSuccessReconnectRequired:
		color.Green(fmt.Sprintf(MsgSetSuccess, "VPN Protocol", resp.Data[0]))
		color.Yellow(SetReconnect)
	case internal.CodeSuccess:
		color.Green(fmt.Sprintf(MsgSetSuccess, "VPN Protocol", resp.Data[0]))
	case internal.CodeDedicatedServersNoNordlynx:
		return errors.New(DedicatedServersAutoconnectNordlynxMessage)
	case internal.CodeObfuscatedNeedsNordwhisper:
		return formatError(fmt.Errorf("To use %s protocol, update or disable auto-connect settings.", resp.Data[0]))
	case internal.CodePqWithoutNordlynx:
		return formatError(fmt.Errorf(SetProtocolDisablePQ, resp.Data[0]))
	}

	return nil
}

func vpnProtocolArgs() []string {
	protocols := config.AvailableVPNProtocols()
	args := make([]string, 0, len(protocols))
	for _, p := range protocols {
		args = append(args, vpnProtocolToArg(p))
	}
	return args
}

const vpnProtocolPrefix = "VPN_PROTOCOL_"

func vpnProtocolToArg(p config.VPNProtocol) string {
	return strings.ToLower(strings.TrimPrefix(p.String(), vpnProtocolPrefix))
}

func vpnProtocolFromArg(arg string) (config.VPNProtocol, bool) {
	completeArg := vpnProtocolPrefix + strings.ToUpper(arg)
	index, ok := config.VPNProtocol_value[completeArg]
	if !ok || index == int32(config.VPNProtocol_VPN_PROTOCOL_UNSPECIFIED) {
		return config.VPNProtocol_VPN_PROTOCOL_UNSPECIFIED, false
	}
	return config.VPNProtocol(index), true
}

func (c *cmd) SetVPNProtocolAutoComplete(ctx *cli.Context) {
	resp, err := c.client.SettingsVPNProtocols(context.Background(), &pb.Empty{})
	if err != nil {
		return
	}

	for _, p := range resp.VpnProtocols {
		fmt.Println(vpnProtocolToArg(p))
	}
}
