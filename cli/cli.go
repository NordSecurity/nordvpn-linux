// Package cli provides command line interface to interact with vpn and fileshare daemons.
package cli

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/client"
	cconfig "github.com/NordSecurity/nordvpn-linux/client/config"
	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	filesharepb "github.com/NordSecurity/nordvpn-linux/fileshare/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	meshpb "github.com/NordSecurity/nordvpn-linux/meshnet/pb"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// AppHelpTemplate is the template we use for forming the cli message on no command
const AppHelpTemplate = `Welcome to NordVPN Linux client app!
Version {{.Version}}
Website: https://nordvpn.com

Usage: {{if .UsageText}}{{.UsageText}}{{else}}{{.HelpName}} {{if .VisibleFlags}}[global options]{{end}}{{if .Commands}} command [command options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}

Commands:{{range .VisibleCategories}}{{if .Name}}
   {{.Name}}:{{end}}{{range .VisibleCommands}}
     {{join .Names ", "}}{{"\t"}}{{.Usage}}{{end}}{{end}}{{end}}
{{if .VisibleFlags}}
Global options:
   {{range $index, $option := .VisibleFlags}}{{if $index}}
   {{end}}{{$option}}{{end}}{{end}}

For more detailed information, please check manual page.

Our customer support works 24/7 so if you have any questions or issues, drop us a line at https://support.nordvpn.com/
`

// CommandHelpTemplate is the template we use to show help
const CommandHelpTemplate = `{{.HelpName}}
Usage: {{if .UsageText}}{{.UsageText}}{{else}}{{.HelpName}}{{if .VisibleFlags}} [command options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}{{end}}
{{if .VisibleFlags}}
Options:
   {{range .VisibleFlags}}{{.}}
   {{end}}{{end}}
`

// CommandWithoutArgsHelpTemplate is the template we use to show help
const CommandWithoutArgsHelpTemplate = `{{.HelpName}}
Usage: {{if .UsageText}}{{.UsageText}}{{else}}{{.HelpName}}{{if .VisibleFlags}} [command options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}{{end}}{{end}}
{{if .VisibleFlags}}
Options:
   {{range .VisibleFlags}}{{.}}
   {{end}}{{end}}
`

// SubcommandHelpTemplate is the template we use to show subcommand help
const SubcommandHelpTemplate = `
{{.HelpName}} - {{if .Description}}{{.Description}}{{else}}{{.Usage}}{{end}}

Usage: {{if .UsageText}}{{.UsageText}}{{else}}{{.HelpName}} command{{if .VisibleFlags}} [command options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}{{end}}

Commands:{{range .VisibleCategories}}{{if .Name}}
   {{.Name}}:{{end}}{{range .VisibleCommands}}
     {{join .Names ", "}}{{"\t"}}{{.Usage}}{{end}}
{{end}}{{if .VisibleFlags}}
Options:
   {{range .VisibleFlags}}{{.}}
   {{end}}{{end}}
`

var ErrConfig = errors.New(client.ConfigMessage)

func NewApp(version, environment, hash, daemonURL, salt string,
	lastAppError, pingErr error,
	conn *grpc.ClientConn,
	fileshareConn *grpc.ClientConn,
	loaderInterceptor *LoaderInterceptor,
) (*cli.App, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}
	configManager := cconfig.NewEncryptedManager(path.Join(configDir, ConfigFilePath), 0, 0, salt)
	cfg, err := configManager.Load()
	if err != nil {
		cfg = cconfig.NewConfig()
		if err := configManager.Save(cfg); err != nil {
			return nil, ErrConfig
		}
	}

	cmd := newCommander(internal.Environment(environment), configManager, cfg)
	cli.AppHelpTemplate = AppHelpTemplate
	cli.SubcommandHelpTemplate = SubcommandHelpTemplate
	cli.CommandHelpTemplate = CommandHelpTemplate
	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Printf("NordVPN Version %s\n", c.App.Version)
	}
	cli.BashCompletionFlag = &cli.BoolFlag{
		Name:   "complete",
		Hidden: true,
	}

	setCommand := cli.Command{
		Name:    "set",
		Aliases: []string{"s"},
		Usage:   "Sets a configuration option",
		Subcommands: []*cli.Command{
			{
				Name:         "autoconnect",
				Usage:        SetAutoconnectUsageText,
				Action:       cmd.SetAutoConnect,
				BashComplete: cmd.SetAutoConnectAutoComplete,
				ArgsUsage:    SetAutoConnectArgsUsageText,
			},
			{
				Name:         "threatprotectionlite",
				Aliases:      []string{"tplite", "tpl", "cybersec"},
				Usage:        SetThreatProtectionLiteUsageText,
				Action:       cmd.SetThreatProtectionLite,
				BashComplete: cmd.SetBoolAutocomplete,
				ArgsUsage:    SetThreatProtectionLiteArgsUsageText,
			},
			{
				Name:   "defaults",
				Usage:  SetDefaultsUsageText,
				Action: cmd.SetDefaults,
			},
			{
				Name:      "dns",
				Usage:     SetDNSUsageText,
				Action:    cmd.SetDNS,
				ArgsUsage: SetDNSArgsUsageText,
			},
			{
				Name:   "firewall",
				Usage:  SetFirewallUsageText,
				Action: cmd.SetFirewall,
				ArgsUsage: fmt.Sprintf(
					MsgSetBoolArgsUsage,
					SetFirewallUsageText,
					"firewall",
					"firewall",
				),
				BashComplete: cmd.SetBoolAutocomplete,
			},
			{
				Name:   "fwmark",
				Usage:  SetFirewallMarkUsageText,
				Action: cmd.SetFirewallMark,
			},
			{
				Name:   "ipv6",
				Usage:  SetIpv6UsageText,
				Action: cmd.SetIpv6,
				ArgsUsage: fmt.Sprintf(
					MsgSetBoolArgsUsage,
					SetIpv6UsageText,
					"ipv6",
					"ipv6",
				),
				BashComplete: cmd.SetBoolAutocomplete,
			},
			{
				Name:   "routing",
				Usage:  SetRoutingUsageText,
				Action: cmd.SetRouting,
				ArgsUsage: fmt.Sprintf(
					MsgSetBoolArgsUsage,
					SetRoutingUsageText,
					"routing",
					"routing",
				),
				BashComplete: cmd.SetBoolAutocomplete,
			},
			{
				Name:   "analytics",
				Usage:  SetAnalyticsUsageText,
				Action: cmd.SetAnalytics,
				ArgsUsage: fmt.Sprintf(
					MsgSetBoolArgsUsage,
					SetAnalyticsUsageText,
					"analytics",
					"analytics",
				),
				BashComplete: cmd.SetBoolAutocomplete,
			},
			{
				Name:         "killswitch",
				Usage:        SetKillSwitchUsageText,
				Action:       cmd.SetKillSwitch,
				BashComplete: cmd.SetBoolAutocomplete,
				ArgsUsage: fmt.Sprintf(
					MsgSetBoolArgsUsage,
					SetKillSwitchUsageText,
					"killswitch",
					"killswitch",
				),
			},
			{
				Name:         "notify",
				Usage:        SetNotifyUsageText,
				Action:       cmd.SetNotify,
				BashComplete: cmd.SetBoolAutocomplete,
				ArgsUsage: fmt.Sprintf(
					MsgSetBoolArgsUsage,
					SetNotifyUsageText,
					"notify",
					"notify",
				),
			},
			{
				Name:         "obfuscate",
				Usage:        SetObfuscateUsageText,
				Action:       cmd.SetObfuscate,
				BashComplete: cmd.SetBoolAutocomplete,
				ArgsUsage: fmt.Sprintf(
					MsgSetBoolArgsUsage,
					SetObfuscateUsageText,
					"obfuscate",
					"obfuscate",
				),
				Hidden: cmd.Except(config.Technology_OPENVPN),
			},
			{
				Name:         "protocol",
				Usage:        SetProtocolUsageText,
				Action:       cmd.SetProtocol,
				BashComplete: cmd.SetProtocolAutoComplete,
				ArgsUsage:    SetProtocolArgsUsageText,
				Hidden:       cmd.Except(config.Technology_OPENVPN),
			},
			{
				Name:         "technology",
				Usage:        SetTechnologyUsageText,
				Action:       cmd.SetTechnology,
				BashComplete: cmd.SetTechnologyAutoComplete,
				ArgsUsage:    SetTechnologyArgsUsageText,
			},
			{
				Name:         "meshnet",
				Aliases:      []string{"mesh"},
				Usage:        MsgSetMeshnetUsage,
				ArgsUsage:    MsgSetMeshnetArgsUsage,
				Action:       cmd.MeshSet,
				BashComplete: cmd.SetBoolAutocomplete,
			},
			{
				Name:  "lan-discovery",
				Usage: SetLANDiscoveryUsage,
				ArgsUsage: fmt.Sprintf(
					MsgSetBoolArgsUsage,
					SetLANDiscoveryUsage,
					"lan-discovery",
					"lan-discovery",
				),
				Action:       cmd.SetLANDiscovery,
				BashComplete: cmd.SetBoolAutocomplete,
			},
		},
	}

	app := cli.NewApp()
	app.EnableBashCompletion = true
	status.Code(err)
	cmd.loaderInterceptor = loaderInterceptor
	app.After = func(*cli.Context) error {
		if conn != nil {
			return conn.Close()
		}
		return nil
	}

	app.Version = version
	if internal.IsDevEnv(environment) {
		app.Version = fmt.Sprintf("%s - %s (%s)", version, internal.Environment(environment), hash)
	}
	app.Commands = []*cli.Command{
		{
			Name:               "account",
			Usage:              AccountUsageText,
			Action:             cmd.Account,
			CustomHelpTemplate: CommandWithoutArgsHelpTemplate,
		},
		{
			Name:         "cities",
			Usage:        CitiesUsageText,
			Action:       cmd.Cities,
			BashComplete: cmd.CitiesAutoComplete,
			ArgsUsage:    CitiesArgsUsageText,
		},
		{
			Name:         "connect",
			Aliases:      []string{"c"},
			Usage:        ConnectUsageText,
			Action:       cmd.Connect,
			BashComplete: cmd.ConnectAutoComplete,
			ArgsUsage:    ConnectArgsUsageText,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "group, g",
					Usage: ConnectFlagGroupUsageText,
				},
			},
		},
		{
			Name:               "countries",
			Usage:              CountriesUsageText,
			Action:             cmd.Countries,
			CustomHelpTemplate: CommandWithoutArgsHelpTemplate,
		},
		{
			Name:               "disconnect",
			Aliases:            []string{"d"},
			Usage:              DisconnectUsageText,
			Action:             cmd.Disconnect,
			CustomHelpTemplate: CommandWithoutArgsHelpTemplate,
		},
		{
			Name:               "groups",
			Usage:              GroupsUsageText,
			Action:             cmd.Groups,
			CustomHelpTemplate: CommandWithoutArgsHelpTemplate,
		},
		{
			Name:   "login",
			Usage:  LoginUsageText,
			Action: cmd.Login,
			Flags: []cli.Flag{
				&cli.BoolFlag{ // TODO: remove in v4
					Name: "nordaccount",
				},
				&cli.BoolFlag{
					Name:  "callback",
					Usage: LoginCallbackUsageText,
				},
				&cli.BoolFlag{
					Name:  "token",
					Usage: LoginFlagTokenUsageText,
				},
			},
		},
		{
			Name:               "token",
			Usage:              TokenUsageText,
			Hidden:             true,
			Action:             cmd.TokenInfo,
			CustomHelpTemplate: CommandWithoutArgsHelpTemplate,
		},
		{
			Name:               "logout",
			Usage:              LogoutUsageText,
			Action:             cmd.Logout,
			CustomHelpTemplate: CommandWithoutArgsHelpTemplate,
			Flags: []cli.Flag{&cli.BoolFlag{
				Name: flagPersistToken,
			}},
		},
		{
			Name:   "click",
			Action: cmd.Click,
			Hidden: true,
		},
		{
			Name:         "rate",
			Usage:        RateUsageText,
			Action:       cmd.Rate,
			BashComplete: cmd.RateAutoComplete,
			ArgsUsage:    RateArgsUsageText,
		},
		{
			Name:   "register",
			Usage:  RegisterUsageText,
			Action: cmd.Register,
		},
		&setCommand,
		{
			Name:               "settings",
			Usage:              SettingsUsageText,
			Action:             cmd.Settings,
			CustomHelpTemplate: CommandWithoutArgsHelpTemplate,
		},
		{
			Name:               "status",
			Usage:              StatusUsageText,
			Action:             cmd.Status,
			CustomHelpTemplate: CommandWithoutArgsHelpTemplate,
		},
		{
			Name:  "version",
			Usage: "Shows the app version",
			Action: func(c *cli.Context) error {
				cli.VersionPrinter(c)
				return nil
			},
			CustomHelpTemplate: CommandWithoutArgsHelpTemplate,
		},
		{
			Name:    "allowlist",
			Aliases: []string{"whitelist"},
			Usage:   "Adds or removes an option from the allowlist",
			Subcommands: []*cli.Command{
				{
					Name:  "add",
					Usage: "Adds an option to the allowlist",
					Subcommands: []*cli.Command{
						{
							Name:         "port",
							Usage:        AllowlistAddPortUsageText,
							Action:       cmd.AllowlistAddPort,
							BashComplete: cmd.AllowlistAddPortAutoComplete,
							ArgsUsage:    AllowlistAddPortArgsUsageText,
						},
						{
							Name:         "ports",
							Usage:        AllowlistAddPortsUsageText,
							Action:       cmd.AllowlistAddPorts,
							BashComplete: cmd.AllowlistAddPortsAutoComplete,
							ArgsUsage:    AllowlistAddPortsArgsUsageText,
						},
						{
							Name:         "subnet",
							Usage:        AllowlistAddSubnetUsageText,
							Action:       cmd.AllowlistAddSubnet,
							BashComplete: cmd.AllowlistAddSubnetAutoComplete,
							ArgsUsage:    AllowlistAddSubnetArgsUsageText,
						},
					},
				},
				{
					Name:  "remove",
					Usage: "Removes an option from the allowlist",
					Subcommands: []*cli.Command{
						{
							Name:               "all",
							Usage:              AllowlistRemoveAllUsageText,
							Action:             cmd.AllowlistRemoveAll,
							CustomHelpTemplate: CommandWithoutArgsHelpTemplate,
						},
						{
							Name:         "port",
							Usage:        AllowlistRemovePortUsageText,
							Action:       cmd.AllowlistRemovePort,
							BashComplete: cmd.AllowlistRemovePortAutoComplete,
							ArgsUsage:    AllowlistRemovePortArgsUsageText,
						},
						{
							Name:         "ports",
							Usage:        AllowlistRemovePortsUsageText,
							Action:       cmd.AllowlistRemovePorts,
							BashComplete: cmd.AllowlistRemovePortsAutoComplete,
							ArgsUsage:    AllowlistRemovePortsArgsUsageText,
						},
						{
							Name:         "subnet",
							Usage:        AllowlistRemoveSubnetUsageText,
							Action:       cmd.AllowlistRemoveSubnet,
							BashComplete: cmd.AllowlistRemoveSubnetAutoComplete,
							ArgsUsage:    AllowlistRemoveSubnetArgsUsageText,
						},
					},
				},
			},
		},
	}

	app.Commands = append(app.Commands, meshnetCommand(cmd))

	if pingErr == nil {
		cmd.client = pb.NewDaemonClient(conn)
		cmd.meshClient = meshpb.NewMeshnetClient(conn)
		cmd.fileshareClient = filesharepb.NewFileshareClient(fileshareConn)
		app.Commands = append(app.Commands, fileshareCommand(cmd))
	}

	app.Commands = addLoaderToActions(cmd, pingErr, app.Commands, daemonURL, lastAppError)
	// Unknown command handler
	app.CommandNotFound = func(c *cli.Context, command string) {
		color.Red(fmt.Sprintf(NoSuchCommand, command))
		os.Exit(1)
	}

	return app, nil
}

func fileshareCommand(c *cmd) *cli.Command {
	return &cli.Command{
		Name:   FileshareName,
		Usage:  MsgFileshareUsage,
		Before: c.IsFileshareDaemonReachable,
		Subcommands: []*cli.Command{
			{
				Name:      FileshareSendName,
				Action:    c.FileshareSend,
				Usage:     MsgFileshareSendUsage,
				ArgsUsage: MsgFileshareSendArgsUsage,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  flagFileshareNoWait,
						Usage: MsgFileshareNoWaitUsage,
					},
				},
				BashComplete: c.FileshareAutoCompletePeers,
			},
			{
				Name:      FileshareAcceptName,
				Action:    c.FileshareAccept,
				Usage:     MsgFileshareAcceptUsage,
				ArgsUsage: MsgFileshareAcceptArgsUsage,
				Flags: []cli.Flag{
					&cli.PathFlag{
						Name:  flagFilesharePath,
						Usage: MsgFileshareAcceptPathUsage,
					},
					&cli.BoolFlag{
						Name:  flagFileshareNoWait,
						Usage: MsgFileshareNoWaitUsage,
					},
				},
				BashComplete: c.FileshareAutoCompleteTransfersAccept,
			},
			{
				Name:      FileshareListName,
				Action:    c.FileshareList,
				Usage:     MsgFileshareListUsage,
				ArgsUsage: MsgFileshareListArgsUsage,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  flagFileshareListIn,
						Usage: MsgFileshareListInUsage,
					},
					&cli.BoolFlag{
						Name:  flagFileshareListOut,
						Usage: MsgFileshareListOutUsage,
					},
				},
				BashComplete: c.FileshareAutoCompleteTransfersList,
			},
			{
				Name:         FileshareCancelName,
				Action:       c.FileshareCancel,
				Usage:        MsgFileshareCancelUsage,
				ArgsUsage:    MsgFileshareCancelArgsUsage,
				BashComplete: c.FileshareAutoCompleteTransfersCancel,
			},
		},
	}
}

func meshnetCommand(c *cmd) *cli.Command {
	return &cli.Command{
		Name:    "meshnet",
		Aliases: []string{"mesh"},
		Usage:   MsgMeshnetUsage,
		Subcommands: []*cli.Command{
			{
				Name:  "peer",
				Usage: MsgMeshnetPeerUsage,
				Subcommands: []*cli.Command{
					{
						Name:   "list",
						Action: c.MeshPeerList,
						Usage:  MsgMeshnetPeerListUsage,
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:    flagFilter,
								Usage:   MsgMeshnetPeerListFilters,
								Aliases: []string{"f"},
							},
						},
						ArgsUsage:    PeerListArgsUsageText,
						BashComplete: c.FiltersAutoComplete,
					},
					{
						Name:         "remove",
						Action:       c.MeshPeerRemove,
						Usage:        MsgMeshnetPeerRemoveUsage,
						ArgsUsage:    MsgMeshnetPeerArgsUsage,
						BashComplete: c.MeshPeerAutoComplete,
					},
					{
						Name:   "refresh",
						Usage:  MsgMeshnetRefreshUsage,
						Action: c.MeshRefresh,
					},
					{
						Name:  "routing",
						Usage: MsgMeshnetPeerRoutingUsage,
						Subcommands: []*cli.Command{
							{
								Name:         "allow",
								Usage:        MsgMeshnetPeerRoutingAllowUsage,
								ArgsUsage:    MsgMeshnetPeerArgsUsage,
								Action:       c.MeshPeerAllowRouting,
								BashComplete: c.MeshPeerAutoComplete,
							},
							{
								Name:         "deny",
								Usage:        MsgMeshnetPeerRoutingDenyUsage,
								ArgsUsage:    MsgMeshnetPeerArgsUsage,
								Action:       c.MeshPeerDenyRouting,
								BashComplete: c.MeshPeerAutoComplete,
							},
						},
					},
					{
						Name:  "incoming",
						Usage: MsgMeshnetPeerIncomingUsage,
						Subcommands: []*cli.Command{
							{
								Name:         "allow",
								Usage:        MsgMeshnetPeerIncomingAllowUsage,
								ArgsUsage:    MsgMeshnetPeerArgsUsage,
								Action:       c.MeshPeerAllowIncoming,
								BashComplete: c.MeshPeerAutoComplete,
							},
							{
								Name:         "deny",
								Usage:        MsgMeshnetPeerIncomingDenyUsage,
								ArgsUsage:    MsgMeshnetPeerArgsUsage,
								Action:       c.MeshPeerDenyIncoming,
								BashComplete: c.MeshPeerAutoComplete,
							},
						},
					},
					{
						Name:  "local",
						Usage: MsgMeshnetPeerLocalNetworkUsage,
						Subcommands: []*cli.Command{
							{
								Name:         "allow",
								Usage:        MsgMeshnetPeerLocalNetworkAllowUsage,
								ArgsUsage:    MsgMeshnetPeerArgsUsage,
								Action:       c.MeshPeerAllowLocalNetwork,
								BashComplete: c.MeshPeerAutoComplete,
							},
							{
								Name:         "deny",
								Usage:        MsgMeshnetPeerLocalNetworkDenyUsage,
								ArgsUsage:    MsgMeshnetPeerArgsUsage,
								Action:       c.MeshPeerDenyLocalNetwork,
								BashComplete: c.MeshPeerAutoComplete,
							},
						},
					},
					{
						Name:  "fileshare",
						Usage: MsgMeshnetPeerFileshareUsage,
						Subcommands: []*cli.Command{
							{
								Name:         "allow",
								Usage:        MsgMeshnetPeerFileshareAllowUsage,
								ArgsUsage:    MsgMeshnetPeerArgsUsage,
								Action:       c.MeshPeerAllowFileshare,
								BashComplete: c.MeshPeerAutoComplete,
							},
							{
								Name:         "deny",
								Usage:        MsgMeshnetPeerFileshareDenyUsage,
								ArgsUsage:    MsgMeshnetPeerArgsUsage,
								Action:       c.MeshPeerDenyFileshare,
								BashComplete: c.MeshPeerAutoComplete,
							},
						},
					},
					{
						Name:  "auto-accept",
						Usage: MsgMeshnetPeerAutomaticFileshareUsage,
						Subcommands: []*cli.Command{
							{
								Name:         "enable",
								Usage:        MsgMeshnetPeerAutomaticFileshareAllowUsage,
								ArgsUsage:    MsgMeshnetPeerArgsUsage,
								Action:       c.MeshPeerEnableAutomaticFileshare,
								BashComplete: c.MeshPeerAutoComplete,
							},
							{
								Name:         "disable",
								Usage:        MsgMeshnetPeerAutomaticFileshareDenyUsage,
								ArgsUsage:    MsgMeshnetPeerArgsUsage,
								Action:       c.MeshPeerDisableAutomaticFileshare,
								BashComplete: c.MeshPeerAutoComplete,
							},
						},
					},
					{
						Name:         "connect",
						Action:       c.MeshPeerConnect,
						Usage:        MsgMeshnetPeerConnectUsage,
						ArgsUsage:    MsgMeshnetPeerArgsUsage,
						BashComplete: c.MeshPeerAutoComplete,
					},
				},
			},
			{
				Name:      "invite",
				Aliases:   []string{"inv"},
				Usage:     MsgMeshnetInviteUsage,
				ArgsUsage: MsgMeshnetInviteArgsUsage,
				Subcommands: []*cli.Command{
					{
						Name:   "list",
						Action: c.MeshGetInvites,
						Usage:  MsgMeshnetInviteListUsage,
					},
					{
						Name:      "send",
						Action:    c.MeshInviteSend,
						Usage:     MsgMeshnetInviteSendUsage,
						ArgsUsage: MsgMeshnetInviteArgsUsage,
						Flags: []cli.Flag{
							&cli.BoolFlag{
								Name:  flagAllowIncomingTraffic,
								Usage: MsgMeshnetInviteAllowIncomingTrafficUsage,
							},
							&cli.BoolFlag{
								Name:  flagAllowTrafficRouting,
								Usage: MsgMeshnetAllowTrafficRoutingUsage,
							},
							&cli.BoolFlag{
								Name:  flagAllowLocalNetwork,
								Usage: MsgMeshnetAllowLocalNetworkUsage,
							},
							&cli.BoolFlag{
								Name:  flagAllowFileshare,
								Usage: MsgMeshnetAllowFileshare,
							},
						},
					},
					{
						Name:         "accept",
						Action:       c.MeshInviteAccept,
						Usage:        MsgMeshnetInviteAcceptUsage,
						ArgsUsage:    MsgMeshnetInviteArgsUsage,
						BashComplete: c.MeshInviteAutoCompletion,
						Flags: []cli.Flag{
							&cli.BoolFlag{
								Name:  flagAllowIncomingTraffic,
								Usage: MsgMeshnetInviteAllowIncomingTrafficUsage,
							},
							&cli.BoolFlag{
								Name:  flagAllowTrafficRouting,
								Usage: MsgMeshnetAllowTrafficRoutingUsage,
							},
							&cli.BoolFlag{
								Name:  flagAllowLocalNetwork,
								Usage: MsgMeshnetAllowLocalNetworkUsage,
							},
							&cli.BoolFlag{
								Name:  flagAllowFileshare,
								Usage: MsgMeshnetAllowFileshare,
							},
						},
					},
					{
						Name:         "deny",
						Action:       c.MeshInviteDeny,
						Usage:        MsgMeshnetInviteDenyUsage,
						ArgsUsage:    MsgMeshnetInviteArgsUsage,
						BashComplete: c.MeshInviteAutoCompletion,
					},
					{
						Name:         "revoke",
						Action:       c.MeshInviteRevoke,
						Usage:        MsgMeshnetInviteRevokeUsage,
						ArgsUsage:    MsgMeshnetInviteArgsUsage,
						BashComplete: c.MeshInviteAutoCompletion,
					},
				},
			},
		},
	}
}

type cmd struct {
	client            pb.DaemonClient
	meshClient        meshpb.MeshnetClient
	fileshareClient   filesharepb.FileshareClient
	environment       internal.Environment
	configManager     cconfig.Manager
	config            cconfig.Config
	loaderInterceptor *LoaderInterceptor
}

func newCommander(environment internal.Environment, configManager cconfig.Manager, config cconfig.Config) *cmd {
	return &cmd{environment: environment, configManager: configManager, config: config}
}

func formatError(e error) error {
	var text string
	if s, ok := status.FromError(e); ok {
		text = s.Message()
	} else {
		text = e.Error()
	}
	capitalized := strings.ToUpper(text[:1]) + text[1:]
	if !strings.HasSuffix(capitalized, ".") {
		capitalized += "."
	}
	return errors.New(capitalized)
}

// LoaderInterceptor is responsible for deciding whether to show loader or not
type LoaderInterceptor struct {
	enabled bool
}

func (i *LoaderInterceptor) UnaryInterceptor(ctx context.Context, method string, req interface{}, reply interface{}, cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	if i.enabled {
		loader := NewLoader()
		loader.Start()
		err := invoker(ctx, method, req, reply, cc, opts...)
		loader.Stop()
		return err
	}
	return invoker(ctx, method, req, reply, cc, opts...)
}

func (i *LoaderInterceptor) StreamInterceptor(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string,
	streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	stream, err := streamer(ctx, desc, cc, method, opts...)
	return cliStream{inner: stream, loaderEnabled: i.enabled}, err
}

type cliStream struct {
	inner         grpc.ClientStream
	loaderEnabled bool
}

func (s cliStream) Header() (metadata.MD, error) {
	return s.inner.Header()
}

func (s cliStream) Trailer() metadata.MD {
	return s.inner.Trailer()
}

func (s cliStream) CloseSend() error {
	return s.inner.CloseSend()
}

func (s cliStream) Context() context.Context {
	return s.inner.Context()
}

func (s cliStream) SendMsg(m interface{}) error {
	return s.inner.SendMsg(m)
}

func (s cliStream) RecvMsg(m interface{}) error {
	if s.loaderEnabled {
		loader := NewLoader()
		loader.Start()
		err := s.inner.RecvMsg(m)
		loader.Stop()
		return err
	}
	return s.inner.RecvMsg(m)
}

func (c *cmd) action(err error, f func(*cli.Context) error, daemonURL string, lastAppError error) func(*cli.Context) error {
	return func(ctx *cli.Context) error {
		c.loaderInterceptor.enabled = true
		if err != nil {
			log.Println(internal.ErrorPrefix, err)
			color.Red(internal.ErrDaemonConnectionRefused.Error())
			os.Exit(1)
		}
		err = c.Ping()
		if err != nil {
			switch err {
			case ErrUpdateAvailable:
				color.Yellow(fmt.Sprintf(UpdateAvailableMessage))
			case ErrInternetConnection:
				color.Red(ErrInternetConnection.Error())
				os.Exit(1)
			case internal.ErrSocketAccessDenied:
				color.Red(formatError(internal.ErrSocketAccessDenied).Error())
				color.Red("Run 'usermod -aG nordvpn $USER' to fix this issue and log out of OS afterwards for this to take an effect.")
				os.Exit(1)
			case internal.ErrDaemonConnectionRefused:
				color.Red(formatError(internal.ErrDaemonConnectionRefused).Error())
				os.Exit(1)
			case internal.ErrSocketNotFound:
				return formatError(internal.ErrSocketNotFound)
			default:
				log.Println(internal.ErrorPrefix, err)
				color.Red(internal.UnhandledMessage)
				os.Exit(1)
			}
		}

		err := f(ctx)
		if err != nil {
			// TODO: Add more error types in the future
			// if more such errors are added
			if err.Error() == "feature not supported" {
				color.Red(MsgMeshnetVersionNotSupported)
				os.Exit(1)
			}
			return err
		}
		return nil
	}
}

// addLoaderToActions wraps all actions with ping error handling and enabling loader
func addLoaderToActions(c *cmd, err error, commands []*cli.Command, daemonURL string, lastAppError error) []*cli.Command {
	var actionCommands []*cli.Command
	for _, command := range commands {
		actionCommands = append(actionCommands, addLoaderToCommandRecursively(c, err, command, daemonURL, lastAppError))
	}
	return actionCommands
}

func addLoaderToCommandRecursively(c *cmd, err error, command *cli.Command, daemonURL string, lastAppError error) *cli.Command {
	if command.Action != nil {
		command.Action = c.action(err, command.Action, daemonURL, lastAppError)
	}
	for _, subc := range command.Subcommands {
		addLoaderToCommandRecursively(c, err, subc, daemonURL, lastAppError)
	}
	return command
}

// meshnetResponseToError returns a human readable error
func meshnetResponseToError(resp *meshpb.MeshnetResponse) error {
	if resp == nil {
		return errors.New(AccountInternalError)
	}
	switch resp := resp.Response.(type) {
	case *meshpb.MeshnetResponse_Empty:
		return nil
	case *meshpb.MeshnetResponse_ServiceError:
		return serviceErrorCodeToError(resp.ServiceError)
	case *meshpb.MeshnetResponse_MeshnetError:
		return meshnetErrorToError(resp.MeshnetError)
	}
	return nil
}

// serviceErrorCodeToError determines the human readable error by
// the error code provided
func serviceErrorCodeToError(code meshpb.ServiceErrorCode) error {
	switch code {
	case meshpb.ServiceErrorCode_NOT_LOGGED_IN:
		return internal.ErrNotLoggedIn
	case meshpb.ServiceErrorCode_API_FAILURE, meshpb.ServiceErrorCode_CONFIG_FAILURE:
		fallthrough
	default:
		return errors.New(AccountInternalError)
	}
}

// meshnetErrorToError determines the human readable from the given
// error code
func meshnetErrorToError(code meshpb.MeshnetErrorCode) error {
	switch code {
	case meshpb.MeshnetErrorCode_NOT_REGISTERED:
		return errors.New(client.MsgTryAgain) // or reset config
	case meshpb.MeshnetErrorCode_LIB_FAILURE:
		return errors.New(client.ConnectCantConnect)
	case meshpb.MeshnetErrorCode_ALREADY_DISABLED:
		return errors.New(MsgMeshnetAlreadyDisabled)
	case meshpb.MeshnetErrorCode_ALREADY_ENABLED:
		return errors.New(MsgMeshnetAlreadyEnabled)
	case meshpb.MeshnetErrorCode_NOT_ENABLED:
		return errors.New(MsgMeshnetNotEnabled)
	case meshpb.MeshnetErrorCode_TECH_FAILURE:
		return errors.New(MsgMeshnetNordlynxMustBeEnabled)
	case meshpb.MeshnetErrorCode_TUNNEL_CLOSED:
		return errors.New(DisconnectNotConnected)
	default:
		return errors.New(AccountInternalError)
	}
}

func argsCountError(ctx *cli.Context) error {
	return fmt.Errorf(
		ArgumentCountError,
		ctx.App.Name,
		ctx.Command.Name,
	)
}

func argsParseError(ctx *cli.Context) error {
	return fmt.Errorf(
		ArgumentParsingError,
		ctx.App.Name,
		ctx.Command.Name,
	)
}
