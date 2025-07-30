// Package cli provides command line interface to interact with vpn and fileshare daemons.
package cli

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"slices"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/client"
	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/events/logger"
	"github.com/NordSecurity/nordvpn-linux/events/subs"
	filesharepb "github.com/NordSecurity/nordvpn-linux/fileshare/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	meshpb "github.com/NordSecurity/nordvpn-linux/meshnet/pb"
	"github.com/NordSecurity/nordvpn-linux/nstrings"
	"github.com/NordSecurity/nordvpn-linux/snapconf"
	snappb "github.com/NordSecurity/nordvpn-linux/snapconf/pb"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
	"golang.org/x/term"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// AppHelpTemplate is the template we use for forming the cli message on no command
const AppHelpTemplate = `Welcome to NordVPN Linux client app!
Version {{.Version}}
Website: https://nordvpn.com

Usage: {{if .UsageText}}{{.UsageText}}{{else}}{{.HelpName}} {{if .VisibleFlags}}[global options]{{end}}{{if .Commands}} command [command options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}

Commands:{{range .VisibleCategories}}{{if .Name}}
   {{.Name}}:{{end}}{{ $cv := offsetCommands .VisibleCommands 7}}{{range .VisibleCommands}}
     {{$s := join .Names ", "}}{{$s}}{{ $sp := subtract $cv (offset $s 5) }}{{ indent $sp ""}}{{wrap .Usage $cv}}{{end}}{{end}}{{end}}
{{if .VisibleFlags}}
Global options:
   {{range $index, $option := .VisibleFlags}}{{if $index}}
   {{end}}{{$option}}{{end}}{{end}}

For more detailed information, please check manual page.

Our customer support works 24/7 so if you have any questions or issues, drop us a line at https://support.nordvpn.com/
`

// CommandHelpTemplate is the template we use to show help
const CommandHelpTemplate = `Usage: {{if .UsageText}}{{.UsageText}}{{else}}{{.HelpName}}{{if .VisibleFlags}} [command options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}{{end}}

{{if .Description}}{{.Description}}{{else}}{{.Usage}}{{end}}
{{if .VisibleFlags}}
Options:
   {{range .VisibleFlags}}{{.}}
   {{end}}{{end}}
`

// CommandWithoutArgsHelpTemplate is the template we use to show help
const CommandWithoutArgsHelpTemplate = `Usage: {{if .UsageText}}{{.UsageText}}{{else}}{{.HelpName}}{{if .VisibleFlags}} [command options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}{{end}}{{end}}

{{if .Description}}{{.Description}}{{else}}{{.Usage}}{{end}}
{{if .VisibleFlags}}
Options:
   {{range .VisibleFlags}}{{.}}
   {{end}}{{end}}
`

// SubcommandHelpTemplate is the template we use to show subcommand help
const SubcommandHelpTemplate = `Usage: {{if .UsageText}}{{.UsageText}}{{else}}{{.HelpName}} command{{if .VisibleFlags}} [command options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}{{end}}

{{if .Description}}{{.Description}}{{else}}{{.Usage}}{{end}}

Commands:{{range .VisibleCategories}}{{if .Name}}
   {{.Name}}:{{end}}{{ $cv := offsetCommands .VisibleCommands 7}}{{range .VisibleCommands}}
     {{$s := join .Names ", "}}{{$s}}{{ $sp := subtract $cv (offset $s 5) }}{{ indent $sp ""}}{{wrap .Usage $cv}}{{end}}
{{end}}{{if .VisibleFlags}}
Options:
   {{range .VisibleFlags}}{{.}}
   {{end}}{{end}}
`

var ErrConfig = errors.New(client.ConfigMessage)

func NewApp(version, environment, hash, salt string,
	pingErr error,
	conn *grpc.ClientConn,
	fileshareConn grpc.ClientConnInterface,
	loaderInterceptor *LoaderInterceptor,
) (*cli.App, error) {
	cmd := newCommander(internal.Environment(environment))
	if pingErr == nil {
		cmd.client = pb.NewDaemonClient(conn)
		cmd.meshClient = meshpb.NewMeshnetClient(conn)
		cmd.fileshareClient = filesharepb.NewFileshareClient(fileshareConn)
	}

	cli.AppHelpTemplate = AppHelpTemplate
	cli.SubcommandHelpTemplate = SubcommandHelpTemplate
	cli.CommandHelpTemplate = CommandHelpTemplate
	// Configure line wrapping for command descriptions
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	cli.HelpPrinter = func(w io.Writer, templ string, data interface{}) {
		funcMap := map[string]interface{}{"wrapAt": func() int { return width }}
		cli.HelpPrinterCustom(w, templ, data, funcMap)
	}

	cli.VersionPrinter = func(c *cli.Context) {
		cmd.Version(c)
	}
	cli.BashCompletionFlag = &cli.BoolFlag{
		Name:   "complete",
		Hidden: true,
	}
	// Capitalize to be uniform with usage text of all other commands
	cli.HelpFlag.(*cli.BoolFlag).Usage = "Show help"
	cli.VersionFlag.(*cli.BoolFlag).Usage = "Print the version"

	isMeshnetEnabled := isMeshnetEnabled(cmd)

	setCommand := cli.Command{
		Name:        "set",
		Aliases:     []string{"s"},
		Usage:       "Sets a configuration option",
		Subcommands: getSetSubcommands(cmd, isMeshnetEnabled),
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

	app.Version = composeAppVersion(version, environment, snapconf.IsUnderSnap())

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
			Description:  CitiesDescription,
		},
		{
			Name:         "connect",
			Aliases:      []string{"c"},
			Usage:        ConnectUsageText,
			Action:       cmd.Connect,
			BashComplete: cmd.ConnectAutoComplete,
			ArgsUsage:    ConnectArgsUsageText,
			Description:  ConnectDescription,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "group",
					Aliases: []string{"g"},
					Usage:   ConnectFlagGroupUsageText,
				},
			},
		},
		{
			Name:               "countries",
			Usage:              fmt.Sprintf(MsgShowListOfServers, "countries"),
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
			Name:        "login",
			Usage:       LoginUsageText,
			Description: LoginDescription,
			Action:      cmd.Login,
			Flags: []cli.Flag{
				&cli.BoolFlag{ // TODO: remove in v4
					Name:  "nordaccount",
					Usage: LoginNordAccountUsageText,
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
				Name:  flagPersistToken,
				Usage: PersistTokenUsageText,
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
			Description:  RateDescription,
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
			Name:               "version",
			Usage:              "Shows daemon version",
			Action:             cmd.Version,
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
							Description:  AllowlistAddPortDescription,
						},
						{
							Name:         "ports",
							Usage:        AllowlistAddPortsUsageText,
							Action:       cmd.AllowlistAddPorts,
							BashComplete: cmd.AllowlistAddPortsAutoComplete,
							ArgsUsage:    AllowlistAddPortsArgsUsageText,
							Description:  AllowlistAddPortsDescription,
						},
						{
							Name:         "subnet",
							Usage:        AllowlistAddSubnetUsageText,
							Action:       cmd.AllowlistAddSubnet,
							BashComplete: cmd.AllowlistAddSubnetAutoComplete,
							ArgsUsage:    AllowlistAddSubnetArgsUsageText,
							Description:  AllowlistAddSubnetDescription,
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
							Description:  AllowlistRemovePortArgsDescription,
						},
						{
							Name:         "ports",
							Usage:        AllowlistRemovePortsUsageText,
							Action:       cmd.AllowlistRemovePorts,
							BashComplete: cmd.AllowlistRemovePortsAutoComplete,
							ArgsUsage:    AllowlistRemovePortsArgsUsageText,
							Description:  AllowlistRemovePortsArgsDescription,
						},
						{
							Name:         "subnet",
							Usage:        AllowlistRemoveSubnetUsageText,
							Action:       cmd.AllowlistRemoveSubnet,
							BashComplete: cmd.AllowlistRemoveSubnetAutoComplete,
							ArgsUsage:    AllowlistRemoveSubnetArgsUsageText,
							Description:  AllowlistRemoveSubnetArgsDescription,
						},
					},
				},
			},
		},
		{
			Name:   "user",
			Action: cmd.User,
			Hidden: true,
		},
	}

	if isMeshnetEnabled {
		app.Commands = append(app.Commands, meshnetCommand(cmd))

		if pingErr == nil {
			fsCommand := fileshareCommand(cmd)
			// TODO: This will currently result in Ping executed twice for every fileshare
			// command but it helps to properly display errors.
			fsCommand.Before = cmd.action(pingErr, fsCommand.Before)
			app.Commands = append(app.Commands, fsCommand)
		}
	}

	app.Commands = addLoaderToActions(cmd, pingErr, app.Commands)
	// Unknown command handler
	app.CommandNotFound = func(c *cli.Context, command string) {
		color.Red(fmt.Sprintf(NoSuchCommand, command))
		os.Exit(1)
	}

	return app, nil
}

func fileshareCommand(c *cmd) *cli.Command {
	return &cli.Command{
		Name:        FileshareName,
		Usage:       MsgFileshareUsage,
		Description: MsgFileshareDescription,
		Before:      c.IsFileshareDaemonReachable,
		Subcommands: []*cli.Command{
			{
				Name:        FileshareSendName,
				Action:      c.FileshareSend,
				Usage:       MsgFileshareSendUsage,
				ArgsUsage:   MsgFileshareSendArgsUsage,
				Description: MsgFileshareSendDescription,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  flagFileshareNoWait,
						Usage: MsgFileshareNoWaitUsage,
					},
				},
				BashComplete: c.FileshareAutoCompletePeers,
			},
			{
				Name:        FileshareAcceptName,
				Action:      c.FileshareAccept,
				Usage:       MsgFileshareAcceptUsage,
				ArgsUsage:   MsgFileshareAcceptArgsUsage,
				Description: MsgFileshareAcceptDescription,
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
				Name:        FileshareListName,
				Action:      c.FileshareList,
				Usage:       MsgFileshareListUsage,
				ArgsUsage:   MsgFileshareListArgsUsage,
				Description: MsgFileshareListDescription,
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
			{
				Name:         FileshareClearName,
				Action:       c.FileshareClear,
				Usage:        MsgFileshareClearUsage,
				ArgsUsage:    MsgFileshareClearArgsUsage,
				Description:  MsgFileshareClearDescription,
				BashComplete: c.FileshareAutoCompleteClear,
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
				Name:        "peer",
				Usage:       MsgMeshnetPeerUsage,
				Description: MsgMeshnetPeerDescription,
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
						Description:  PeerListDescription,
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
						Name:        "incoming",
						Usage:       MsgMeshnetPeerIncomingUsage,
						Description: MsgMeshnetPeerIncomingDescription,
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
						Name:        "routing",
						Usage:       MsgMeshnetPeerRoutingUsage,
						Description: MsgMeshnetPeerRoutingDescription,
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
						Name:        "local",
						Usage:       MsgMeshnetPeerLocalNetworkUsage,
						Description: MsgMeshnetPeerLocalNetworkDescription,
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
						Name:        "fileshare",
						Usage:       MsgMeshnetPeerFileshareUsage,
						Description: MsgMeshnetPeerFileshareDescription,
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
					{
						Name:    "nickname",
						Aliases: []string{"nick"},
						Usage:   MsgMeshnetPeerNicknameUsage,
						Subcommands: []*cli.Command{
							{
								Name:         "set",
								Aliases:      []string{"s"},
								Usage:        MsgMeshnetPeerSetNicknameUsage,
								ArgsUsage:    MsgMeshnetPeerSetNicknameArgsUsage,
								Action:       c.MeshPeerSetNickname,
								BashComplete: c.MeshPeerNicknameAutoComplete,
							},
							{
								Name:         "remove",
								Aliases:      []string{"r"},
								Usage:        MsgMeshnetPeerRemoveNicknameUsage,
								ArgsUsage:    MsgMeshnetPeerRemoveNicknameArgsUsage,
								Action:       c.MeshPeerRemoveNickname,
								BashComplete: c.MeshPeerNicknameAutoComplete,
							},
						},
					},
				},
			},
			{
				Name:        "invite",
				Aliases:     []string{"inv"},
				Usage:       MsgMeshnetInviteUsage,
				Description: MsgMeshnetInviteDescription,
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
			{
				Name:    "set",
				Aliases: []string{"s"},
				Usage:   MsgMeshnetSetUsage,
				Subcommands: []*cli.Command{
					{
						Name:      "nickname",
						Aliases:   []string{"nick"},
						Usage:     MsgMeshnetSetMachineNicknameUsage,
						ArgsUsage: MsgMeshnetSetNicknameArgsUsage,
						Action:    c.MeshSetMachineNickname,
					},
				},
			},
			{
				Name:    "remove",
				Aliases: []string{"r"},
				Usage:   MsgMeshnetRemoveUsage,
				Subcommands: []*cli.Command{
					{
						Name:               "nickname",
						Aliases:            []string{"nick"},
						Usage:              MsgMeshnetRemoveMachineNicknameUsage,
						Action:             c.MeshRemoveMachineNickname,
						CustomHelpTemplate: CommandWithoutArgsHelpTemplate,
					},
				},
			},
		},
	}
}

func getSetSubcommands(cmd *cmd, isMeshnetEnabled bool) []*cli.Command {
	setSubcommands := []*cli.Command{
		{
			Name:         "autoconnect",
			Usage:        SetAutoconnectUsageText,
			Action:       cmd.SetAutoConnect,
			BashComplete: cmd.SetAutoConnectAutoComplete,
			ArgsUsage:    SetAutoConnectArgsUsageText,
			Description:  SetAutoConnectDescription,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "group",
					Aliases: []string{"g"},
					Usage:   ConnectFlagGroupUsageText,
				},
			},
		},
		{
			Name:         "threatprotectionlite",
			Aliases:      []string{"tplite", "tpl", "cybersec"},
			Usage:        SetThreatProtectionLiteUsageText,
			Action:       cmd.SetThreatProtectionLite,
			BashComplete: cmd.SetBoolAutocomplete,
			ArgsUsage:    SetThreatProtectionLiteArgsUsageText,
			Description:  SetThreatProtectionLiteDescription,
		},
		{
			Name:  "defaults",
			Usage: SetDefaultsUsageText,
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:  flagLogout,
					Usage: SetDefaultsLogoutFlagText,
				},
				&cli.BoolFlag{
					Name:  flagOffKillswitch,
					Usage: SetDefaultsOffKillswitchFlagText,
				},
			},
			Action: cmd.SetDefaults,
		},
		{
			Name:        "dns",
			Usage:       SetDNSUsageText,
			Action:      cmd.SetDNS,
			ArgsUsage:   SetDNSArgsUsageText,
			Description: SetDNSDescription,
		},
		{
			Name:      "firewall",
			Usage:     SetFirewallUsageText,
			Action:    cmd.SetFirewall,
			ArgsUsage: MsgSetBoolArgsUsage,
			Description: fmt.Sprintf(
				MsgSetBoolDescription,
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
			Name:      "routing",
			Usage:     SetRoutingUsageText,
			Action:    cmd.SetRouting,
			ArgsUsage: MsgSetBoolArgsUsage,
			Description: fmt.Sprintf(
				MsgSetBoolDescription,
				SetRoutingUsageText,
				"routing",
				"routing",
			),
			BashComplete: cmd.SetBoolAutocomplete,
		},
		{
			Name:      "analytics",
			Usage:     SetAnalyticsUsageText,
			Action:    cmd.SetAnalytics,
			ArgsUsage: MsgSetBoolArgsUsage,
			Description: fmt.Sprintf(
				MsgSetBoolDescription,
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
			ArgsUsage:    MsgSetBoolArgsUsage,
			Description: fmt.Sprintf(
				MsgSetBoolDescription,
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
			ArgsUsage:    MsgSetBoolArgsUsage,
			Description: fmt.Sprintf(
				MsgSetBoolDescription,
				SetNotifyUsageText,
				"notify",
				"notify",
			),
		},
		{
			Name:         "tray",
			Usage:        SetTrayUsageText,
			Action:       cmd.SetTray,
			BashComplete: cmd.SetBoolAutocomplete,
			ArgsUsage:    MsgSetBoolArgsUsage,
			Description: fmt.Sprintf(
				MsgSetBoolDescription,
				SetTrayUsageText,
				"tray",
				"tray",
			),
		},
		{
			Name:         "obfuscate",
			Usage:        SetObfuscateUsageText,
			Action:       cmd.SetObfuscate,
			BashComplete: cmd.SetBoolAutocomplete,
			ArgsUsage:    MsgSetBoolArgsUsage,
			Description: fmt.Sprintf(
				MsgSetBoolDescription,
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
			Description:  SetProtocolDescription,
			Hidden:       cmd.Except(config.Technology_OPENVPN),
		},
		{
			Name:         "technology",
			Usage:        SetTechnologyUsageText,
			Action:       cmd.SetTechnology,
			BashComplete: cmd.SetTechnologyAutoComplete,
			ArgsUsage:    SetTechnologyArgsUsageText,
			Description:  fmt.Sprintf(SetTechnologyDescription),
		},
		{
			Name:      "lan-discovery",
			Usage:     SetLANDiscoveryUsage,
			ArgsUsage: MsgSetBoolArgsUsage,
			Description: fmt.Sprintf(
				MsgSetBoolDescription,
				SetLANDiscoveryUsage,
				"lan-discovery",
				"lan-discovery",
			),
			Action:       cmd.SetLANDiscovery,
			BashComplete: cmd.SetBoolAutocomplete,
		},
		{
			Name:      "virtual-location",
			Usage:     MsgSetVirtualLocationUsageText,
			ArgsUsage: MsgSetBoolArgsUsage,
			Description: fmt.Sprintf(
				MsgSetBoolDescription,
				MsgSetVirtualLocationDescription,
				"virtual-location",
				"virtual-location",
			),
			Action:       cmd.SetVirtualLocation,
			BashComplete: cmd.SetBoolAutocomplete,
		},
		{
			Name:         "post-quantum",
			Aliases:      []string{"pq"},
			Usage:        SetPqUsageText,
			Action:       cmd.SetPostquantumVpn,
			BashComplete: cmd.SetBoolAutocomplete,
			ArgsUsage:    MsgSetBoolArgsUsage,
			Description: fmt.Sprintf(
				MsgSetBoolDescription,
				SetPqUsageText,
				"post-quantum",
				"post-quantum",
			),
			Hidden: cmd.Except(config.Technology_NORDLYNX),
		},
	}

	setMeshCommand := cli.Command{
		Name:         "meshnet",
		Aliases:      []string{"mesh"},
		Usage:        MsgSetMeshnetUsage,
		ArgsUsage:    MsgSetMeshnetArgsUsage,
		Description:  MsgSetMeshnetDescription,
		Action:       cmd.MeshSet,
		BashComplete: cmd.SetBoolAutocomplete,
	}

	if isMeshnetEnabled {
		setSubcommands = append(setSubcommands, &setMeshCommand)
	}

	return setSubcommands
}

type cmd struct {
	client            pb.DaemonClient
	meshClient        meshpb.MeshnetClient
	fileshareClient   filesharepb.FileshareClient
	environment       internal.Environment
	loaderInterceptor *LoaderInterceptor
}

func newCommander(environment internal.Environment) *cmd {
	return &cmd{
		environment: environment,
	}
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
	return loaderStream{ClientStream: stream, loaderEnabled: i.enabled}, err
}

// RetrieveSnapConnsError checks whether ny of the details inside gRPC error is a
// `SnapPermissionError` and returns pointer to it. Otherwise, returns nil
func RetrieveSnapConnsError(err error) *snappb.ErrMissingConnections {
	s := status.Convert(err)
	for _, d := range s.Details() {
		permError, ok := d.(*snappb.ErrMissingConnections)
		if ok {
			return permError
		}
	}
	return nil
}

// Concatenate snap missing connections and returns a string
func JoinSnapMissingPermissions(err *snappb.ErrMissingConnections) string {
	if len(err.MissingConnections) == 0 {
		return ""
	}

	return "sudo snap connect nordvpn:" + strings.Join(err.MissingConnections, "\nsudo snap connect nordvpn:")
}

func FormatSnapMissingConnsErr(err *snappb.ErrMissingConnections) string {
	return fmt.Sprintf(MsgNoSnapPermissions, JoinSnapMissingPermissions(err))
}

func FormatSnapMissingConnsExtErr(err *snappb.ErrMissingConnections) string {
	return fmt.Sprintf(MsgNoSnapPermissionsExt, JoinSnapMissingPermissions(err))
}

type loaderStream struct {
	grpc.ClientStream
	loaderEnabled bool
}

func (s loaderStream) RecvMsg(m interface{}) error {
	if s.loaderEnabled {
		loader := NewLoader()
		loader.Start()
		err := s.ClientStream.RecvMsg(m)
		loader.Stop()
		return err
	}
	return s.ClientStream.RecvMsg(m)
}

func isLoaderEnabled() bool {
	value, exists := os.LookupEnv("DISABLE_TUI_LOADER")
	return !(exists && value == "1")
}

func (c *cmd) action(err error, f func(*cli.Context) error) func(*cli.Context) error {
	return func(ctx *cli.Context) error {
		c.loaderInterceptor.enabled = isLoaderEnabled()
		if err != nil {
			log.Println(internal.ErrorPrefix, err)
			color.Red(internal.ErrDaemonConnectionRefused.Error())
			os.Exit(1)
		}
		err = c.Ping()
		if err != nil {
			// this is snap-check is performed on daemon side
			if snapErr := RetrieveSnapConnsError(err); snapErr != nil {
				color.Red(FormatSnapMissingConnsErr(snapErr))
				os.Exit(1)
			}
			switch {
			case errors.Is(err, ErrUpdateAvailable):
				color.Yellow(fmt.Sprintf(UpdateAvailableMessage))
			case errors.Is(err, ErrInternetConnection):
				color.Red(ErrInternetConnection.Error())
				os.Exit(1)
			case errors.Is(err, internal.ErrSocketAccessDenied):
				if snapconf.IsUnderSnap() {
					// this is additional snap-check on client side to minimize user actions
					errSubject := &subs.Subject[error]{}
					errSubject.Subscribe(logger.Subscriber{}.NotifyError)
					err := snapconf.NewSnapChecker(errSubject).PermissionCheck()
					if snapErr := RetrieveSnapConnsError(err); snapErr != nil {
						color.Red(FormatSnapMissingConnsExtErr(snapErr))
					} else {
						color.Red(MsgSnapNoSocketPermissions)
					}
				} else {
					color.Red(MsgNoSocketPermissions)
				}
				os.Exit(1)
			case errors.Is(err, internal.ErrDaemonConnectionRefused):
				color.Red(formatError(internal.ErrDaemonConnectionRefused).Error())
				os.Exit(1)
			case errors.Is(err, internal.ErrSocketNotFound):
				color.Red(formatError(internal.ErrSocketNotFound).Error())
				color.Red("The NordVPN background service isn't running. Execute the \"systemctl enable --now nordvpnd\" command with root privileges to start the background service. If you're using NordVPN in an environment without systemd (a container, for example), use the \"/etc/init.d/nordvpn start\" command.")
				os.Exit(1)
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
func addLoaderToActions(c *cmd, err error, commands []*cli.Command) []*cli.Command {
	var actionCommands []*cli.Command
	for _, command := range commands {
		actionCommands = append(actionCommands, addLoaderToCommandRecursively(c, err, command))
	}
	return actionCommands
}

func addLoaderToCommandRecursively(c *cmd, err error, command *cli.Command) *cli.Command {
	if command.Action != nil {
		command.Action = c.action(err, command.Action)
	}
	for _, subc := range command.Subcommands {
		addLoaderToCommandRecursively(c, err, subc)
	}
	return command
}

// MeshnetResponseToError returns a human readable error
func MeshnetResponseToError(resp *meshpb.MeshnetResponse) error {
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
	case meshpb.MeshnetErrorCode_CONFLICT_WITH_PQ:
		return errors.New(SetPqAndMeshnet)
	case meshpb.MeshnetErrorCode_CONFLICT_WITH_PQ_SERVER:
		return errors.New(SetPqAndMeshnetServer)
	default:
		return errors.New(AccountInternalError)
	}
}

func argsCountError(ctx *cli.Context) error {
	return fmt.Errorf(
		ArgumentCountError,
		commandFullName(ctx, os.Args),
	)
}

func argsParseError(ctx *cli.Context) error {
	return fmt.Errorf(
		ArgumentParsingError,
		commandFullName(ctx, os.Args),
	)
}

// because ctx.Command.FullName() doesn't work: https://github.com/urfave/cli/issues/1859
func commandFullName(ctx *cli.Context, args []string) string {
	fullCommand := []string{ctx.App.Name}
	if len(args) < 2 {
		if ctx.Command.Name != "" {
			fullCommand = append(fullCommand, ctx.Command.Name)
		}
	} else {
		var cmd *cli.Command
		for _, arg := range args[1:] {
			if cmd == nil {
				cmd = ctx.App.Command(arg)
			} else {
				cmd = cmd.Command(arg)
			}
			if cmd == nil {
				break
			}

			fullCommand = append(fullCommand, cmd.Name)
		}
	}

	return strings.Join(fullCommand, " ")
}

// Get the value for a flag from the command line arguments
// if the flag exists but it doesn't have value will return empty string and found = true
func getFlagValue(name string, ctx *cli.Context) (value string, found bool) {
	if ctx.IsSet(name) {
		// value exists and has value
		return ctx.String(name), true
	}

	names := []string{name}
	for _, flag := range ctx.Command.Flags {
		if slices.Index(flag.Names(), name) == -1 {
			continue
		}

		names = flag.Names()
		break
	}

	searchFn := func(args []string, argName string) (string, bool) {
		for index, v := range args {
			if v == "-"+argName || v == "--"+argName {
				if index < len(os.Args)-1 && os.Args[index+1][0] != '-' {
					// if there is another argument after
					return os.Args[index+1], true
				}
				return "", true
			}
		}
		return "", false
	}

	for _, name := range names {
		value, found := searchFn(os.Args, name)
		if found {
			return value, found
		}

		// this is normally used for tests where ctx has the test arguments instead of os.Args
		value, found = searchFn(ctx.Args().Slice(), name)
		if found {
			return value, found
		}
	}

	return "", false
}

func (c *cmd) printServersForAutoComplete(country string, hasGroupFlag bool, groupName string) {
	// if no country name or --group flag exists don't show cities
	if hasGroupFlag || country == "" {
		resp, err := c.client.Groups(context.Background(), &pb.Empty{})
		if err != nil {
			log.Println(internal.ErrorPrefix, "failed to get the groups", err)
			return
		}

		output := ""
		for _, server := range resp.Servers {
			if hasGroupFlag && groupName == server.Name {
				// if the group is equal to one of the group names then exists don't return anything
				return
			}
			output += server.Name + "\n"
		}

		fmt.Print(output)

		if hasGroupFlag {
			// if --group flag exists don't show the countries
			return
		}

		resp, err = c.client.Countries(context.Background(), &pb.Empty{})
		if err != nil {
			log.Println(internal.ErrorPrefix, "failed to get the countries", err)
			return
		}
		for _, server := range resp.Servers {
			fmt.Println(server.Name)
		}
	} else {
		// get the cities from the given country
		resp, err := c.client.Cities(context.Background(), &pb.CitiesRequest{
			Country: country,
		})
		if err != nil {
			log.Println(internal.ErrorPrefix, "failed to get the cities", err)
			return
		}

		for _, server := range resp.Servers {
			fmt.Println(server.Name)
		}
	}
}

// removeGroupFlagFromArgs removes flag and its value from args
// The function assumes that flag value occurs after flag name, i.e --<flag> <value>
func removeFlagFromArgs(args []string, flag string) []string {
	for index, arg := range args {
		if arg == "--"+flag {
			if index+1 >= len(args) {
				return slices.Delete(args, index, index+1)
			}
			// return the args slice sans the flag and its argument
			return slices.Delete(args, index, index+2)
		}
	}

	return args
}

// parseConnectArgs extracts server tag and server group from the arguments provided to the connect and set autoconnect
// commands. It also accommodates for the issue in github.com/urfave/cli/v2 where a flag is only interpreted as a flag if
// it's the first agument to the command.
func parseConnectArgs(ctx *cli.Context) (string, string, error) {
	groupName, hasGroupFlag := getFlagValue(flagGroup, ctx)
	args := ctx.Args()
	if !args.Present() && !hasGroupFlag {
		return "", "", nil
	}

	var serverTag string
	var serverGroup string
	argsSlice := args.Slice()
	if hasGroupFlag {
		if groupName == "" {
			return "", "", argsCountError(ctx)
		}

		// remove group flags if present, as they were already processed
		if slices.Contains(argsSlice, "--"+flagGroup) {
			argsSlice = removeFlagFromArgs(argsSlice, flagGroup)
		}

		serverGroup = groupName
	}

	// remove any arguments that successfully parse as an on/off switch
	argsSlice = slices.DeleteFunc(argsSlice, func(arg string) bool {
		_, boolFromStringErr := nstrings.BoolFromString(arg)
		return boolFromStringErr == nil
	})
	serverTag = strings.Join(argsSlice, " ")
	serverTag = strings.ToLower(serverTag)
	serverGroup = strings.ToLower(serverGroup)

	return serverTag, serverGroup, nil
}

// readForConfirmation from the reader with a given prompt.
// Returns the given answer and a status. If invalid response was given, status will be set to false.
func readForConfirmation(r io.Reader, prompt string) (bool, bool) {
	fmt.Print(prompt)
	answer, _, _ := bufio.NewReader(r).ReadRune()
	switch answer {
	case 'y', 'Y':
		return true, true
	case 'n', 'N':
		return false, true
	default:
		return false, false
	}
}

func readForConfirmationDefaultValue(r io.Reader, prompt string, defaultValue bool) bool {
	if defaultValue {
		prompt = fmt.Sprintf("%s [Y/n]", prompt)
	} else {
		prompt = fmt.Sprintf("%s [y/N] ", prompt)
	}

	answer, ok := readForConfirmation(r, prompt)
	if !ok {
		return defaultValue
	}
	return answer
}

func readForConfirmationBlockUntilValid(r io.Reader, prompt string) bool {
	for {
		answer, ok := readForConfirmation(r, prompt)
		if ok {
			return answer
		}
		fmt.Println(InputParsingError)
	}
}

// composeAppVersion concatenates the provided information to produce a string
// representing the application version in the format: 1.2.3 [snap] - dev
// Later it is used by the version command to display to the user
func composeAppVersion(buildVersion string, environment string, isSnap bool) string {
	env := ""
	if internal.IsDevEnv(environment) {
		env = fmt.Sprintf(" - %s", environment)
	}

	snap := ""
	if isSnap {
		snap = " [snap]"
	}

	return fmt.Sprintf("%s%s%s", buildVersion, snap, env)
}

func isMeshnetEnabled(cmd *cmd) bool {
	featureToggles := cmd.GetFeatureToggles()
	return featureToggles.meshnetEnabled
}
