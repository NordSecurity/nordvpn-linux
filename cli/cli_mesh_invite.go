package cli

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/meshnet/pb"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

const (
	flagAllowIncomingTraffic = "allow-incoming-traffic"
	flagAllowTrafficRouting  = "allow-traffic-routing"
	flagAllowLocalNetwork    = "allow-local-network-access"
	flagAllowFileshare       = "allow-peer-send-files"
)

type meshRespondToInviteReqFn = func(email string) (
	*pb.RespondToInviteResponse,
	error,
)

func (c *cmd) MeshGetInvites(ctx *cli.Context) error {
	resp, err := c.meshClient.GetInvites(
		context.Background(),
		&pb.Empty{},
	)
	if err != nil {
		return formatError(err)
	}
	invites, err := invitesListResponseToInvitesList(resp)
	if err != nil {
		return formatError(err)
	}

	var buf strings.Builder
	boldCol := color.New(color.Bold)
	buf.WriteString(boldCol.Sprintf("Sent Invites:\n"))
	if len(invites.Sent) == 0 {
		buf.WriteString("[no invites]\n")
	}
	for _, invite := range invites.Sent {
		str := fmt.Sprintf("%s: %s",
			color.New(color.FgYellow, color.Bold).Sprintf("Email"),
			color.New(color.FgYellow).Sprintf(invite.Email))
		buf.WriteString(str + "\n")
	}

	buf.WriteString(boldCol.Sprintf("\nReceived Invites:\n"))
	if len(invites.Received) == 0 {
		buf.WriteString("[no invites]\n")
	}
	for _, invite := range invites.Received {
		inviteOs := ""
		if invite.Os != "" {
			inviteOs = "(" + invite.Os + ")"
		}
		str := fmt.Sprintf("%s: %s %s",
			color.New(color.FgYellow, color.Bold).Sprintf("Email"),
			color.New(color.FgYellow).Sprintf(invite.Email),
			color.New(color.FgYellow).Sprintf(inviteOs))
		buf.WriteString(str + "\n")
	}

	fmt.Print(buf.String())
	return nil
}

// MeshInviteSend sends the mesh invitation request to the daemon
func (c *cmd) MeshInviteSend(ctx *cli.Context) error {
	email := ctx.Args().First()
	if email == "" {
		return formatError(argsCountError(ctx))
	}

	{
		resp, err := c.meshClient.GetInvites(context.Background(), &pb.Empty{})
		if err != nil {
			return formatError(err)
		}
		invites, err := invitesListResponseToInvitesList(resp)
		if err != nil {
			return err
		}
		for _, inv := range invites.GetSent() {
			if inv.Email == email {
				return formatError(inviteErrorCodeToError(
					pb.InviteResponseErrorCode_ALREADY_EXISTS, email,
				))
			}
		}
	}

	permissions := c.meshPermissions(ctx)
	resp, err := c.meshClient.Invite(
		context.Background(),
		&pb.InviteRequest{
			Email:                email,
			AllowIncomingTraffic: permissions.allowTraffic,
			AllowTrafficRouting:  permissions.routeTraffic,
			AllowLocalNetwork:    permissions.localNetwork,
			AllowFileshare:       permissions.fileshare,
		},
	)

	if err != nil {
		return formatError(err)
	}

	if err := inviteResponseToError(
		resp,
		email,
	); err != nil {
		return formatError(err)
	}
	color.Green(MsgMeshnetInviteSentSuccess, email)
	return nil
}

type meshPermissions struct {
	allowTraffic bool
	routeTraffic bool
	localNetwork bool
	fileshare    bool
}

// meshPermissions is responsible for prompting the user
// for incoming traffic and traffic routing permissions.
func (c *cmd) meshPermissions(ctx *cli.Context) meshPermissions {
	var permissions meshPermissions

	if ctx.IsSet(flagAllowIncomingTraffic) {
		permissions.allowTraffic = ctx.Bool(flagAllowIncomingTraffic)
	} else {
		permissions.allowTraffic = readForConfirmation(os.Stdin, "Would you like to allow incoming traffic?")
	}

	if ctx.IsSet(flagAllowTrafficRouting) {
		permissions.routeTraffic = ctx.Bool(flagAllowTrafficRouting)
	} else {
		permissions.routeTraffic = readForConfirmation(os.Stdin, "Would you like to allow traffic routing?")
	}

	if ctx.IsSet(flagAllowLocalNetwork) {
		permissions.localNetwork = ctx.Bool(flagAllowLocalNetwork)
	} else {
		permissions.localNetwork = readForConfirmation(os.Stdin, "Would you like to allow access to your local network?")
	}

	if ctx.IsSet(flagAllowFileshare) {
		permissions.fileshare = ctx.Bool(flagAllowFileshare)
	} else {
		permissions.fileshare = readForConfirmation(os.Stdin, "Would you like to allow peer to send you files?")
	}

	return permissions
}

// readForConfirmation from the reader with a given prompt.
// In case of any invalid input or just enter, return false.
func readForConfirmation(r io.Reader, prompt string) bool {
	fmt.Printf("%s [Y/n] ", prompt)
	answer, _, _ := bufio.NewReader(r).ReadRune()
	switch answer {
	case 'y', 'Y':
		return true
	case 'n', 'N':
		return false
	default:
		return true
	}
}

// MeshInviteRevoke sends a meshnet invite revoke request to a daemon
func (c *cmd) MeshInviteRevoke(ctx *cli.Context) error {
	reqFn := func(email string) (
		*pb.RespondToInviteResponse,
		error,
	) {
		req := &pb.DenyInviteRequest{
			Email: email,
		}

		return c.meshClient.RevokeInvite(
			context.Background(),
			req,
		)
	}

	return c.meshRespondToInvite(
		ctx,
		reqFn,
		MsgMeshnetInviteRevokeSuccess,
	)
}

// MeshInviteDeny sends the meshnet accept invite request to a daemon
func (c *cmd) MeshInviteAccept(ctx *cli.Context) error {
	reqFn := func(email string) (
		*pb.RespondToInviteResponse,
		error,
	) {
		resp, err := c.meshClient.GetInvites(context.Background(), &pb.Empty{})
		if err != nil {
			return nil, err
		}

		invites, err := invitesListResponseToInvitesList(resp)
		if err != nil {
			return nil, err
		}

		inviteFound := false
		for _, inv := range invites.GetReceived() {
			if inv.GetEmail() == email {
				inviteFound = true
				break
			}
		}

		if !inviteFound {
			return &pb.RespondToInviteResponse{
				Response: &pb.RespondToInviteResponse_RespondToInviteErrorCode{
					RespondToInviteErrorCode: pb.RespondToInviteErrorCode_NO_SUCH_INVITATION,
				},
			}, nil
		}

		permissions := c.meshPermissions(ctx)
		return c.meshClient.AcceptInvite(
			context.Background(),
			&pb.InviteRequest{
				Email:                email,
				AllowIncomingTraffic: permissions.allowTraffic,
				AllowTrafficRouting:  permissions.routeTraffic,
				AllowLocalNetwork:    permissions.localNetwork,
				AllowFileshare:       permissions.fileshare,
			},
		)
	}

	return c.meshRespondToInvite(
		ctx,
		reqFn,
		MsgMeshnetInviteAcceptSuccess,
	)
}

// MeshInviteDeny sends the meshnet deny invite request to a daemon
func (c *cmd) MeshInviteDeny(ctx *cli.Context) error {
	reqFn := func(email string) (
		*pb.RespondToInviteResponse,
		error,
	) {
		req := &pb.DenyInviteRequest{
			Email: email,
		}

		return c.meshClient.DenyInvite(
			context.Background(),
			req,
		)
	}
	return c.meshRespondToInvite(
		ctx,
		reqFn,
		MsgMeshnetInviteDenySuccess,
	)
}

// meshRespondToInvite handles the user's input for the invitation
// response, sends the response to the daemon and handles the daemon's
// response
func (c *cmd) meshRespondToInvite(
	ctx *cli.Context,
	reqFn meshRespondToInviteReqFn,
	successMsg string,
) error {
	email := ctx.Args().First()
	if email == "" {
		return formatError(argsCountError(ctx))
	}

	resp, err := reqFn(email)

	if err != nil {
		return formatError(err)
	}
	if err := respondToInviteResponseToError(
		resp,
		email,
	); err != nil {
		return formatError(err)
	}

	color.Green(successMsg, email)
	return nil
}

func (c *cmd) MeshInviteAutoCompletion(ctx *cli.Context) {
	resp, err := c.meshClient.GetInvites(
		context.Background(),
		&pb.Empty{},
	)
	if err != nil {
		return
	}
	invites, err := invitesListResponseToInvitesList(resp)
	if err != nil {
		return
	}

	var invs []*pb.Invite
	if ctx.Command.Name == "revoke" {
		invs = invites.Sent
	} else {
		invs = invites.Received
	}
	for _, invite := range invs {
		fmt.Println(invite.GetEmail())
	}
}

func (c *cmd) MeshInviteRevokeAutoCompletion(ctx *cli.Context) {
	resp, err := c.meshClient.GetInvites(
		context.Background(),
		&pb.Empty{},
	)
	if err != nil {
		return
	}
	invites, err := invitesListResponseToInvitesList(resp)
	if err != nil {
		return
	}

	for _, invite := range invites.Sent {
		fmt.Println(invite.GetEmail())
	}
}

// invitesListResponseToInvitesList determines whether the invites
// response is an error and returns a human readable form of it. If
// this is a valid invite list, it returns that.
func invitesListResponseToInvitesList(
	resp *pb.GetInvitesResponse,
) (*pb.InvitesList, error) {
	if resp == nil {
		return nil, errors.New(AccountInternalError)
	}
	switch resp := resp.Response.(type) {
	case *pb.GetInvitesResponse_Invites:
		return resp.Invites, nil
	case *pb.GetInvitesResponse_ServiceErrorCode:
		return nil, serviceErrorCodeToError(resp.ServiceErrorCode)
	case *pb.GetInvitesResponse_MeshnetErrorCode:
		return nil, meshnetErrorToError(resp.MeshnetErrorCode)
	default:
		return nil, errors.New(AccountInternalError)
	}
}

func inviteResponseToError(
	resp *pb.InviteResponse,
	email string,
) error {
	if resp == nil {
		return errors.New(AccountInternalError)
	}
	switch resp := resp.Response.(type) {
	case *pb.InviteResponse_Empty:
		return nil
	case *pb.InviteResponse_InviteResponseErrorCode:
		return inviteErrorCodeToError(
			resp.InviteResponseErrorCode,
			email,
		)
	case *pb.InviteResponse_ServiceErrorCode:
		return serviceErrorCodeToError(resp.ServiceErrorCode)
	case *pb.InviteResponse_MeshnetErrorCode:
		return meshnetErrorToError(resp.MeshnetErrorCode)
	}
	return nil
}

func inviteErrorCodeToError(
	code pb.InviteResponseErrorCode,
	email string,
) error {
	switch code {
	case pb.InviteResponseErrorCode_ALREADY_EXISTS:
		return fmt.Errorf(
			MsgMeshnetInviteSendAlreadyExists,
			email,
		)
	case pb.InviteResponseErrorCode_INVALID_EMAIL:
		return fmt.Errorf(
			MsgMeshnetInviteSendInvalidEmail,
			email,
		)
	case pb.InviteResponseErrorCode_SAME_ACCOUNT_EMAIL:
		return fmt.Errorf(MsgMeshnetInviteSendSameAccountEmail)
	case pb.InviteResponseErrorCode_PEER_COUNT:
		return errors.New(MsgMeshnetInviteSendDeviceCount)
	case pb.InviteResponseErrorCode_LIMIT_REACHED:
		return errors.New(MsgMeshnetInviteWeeklyLimit)
	default:
		return errors.New(AccountInternalError)
	}
}

// respondToInviteResponseToError determines whether the response
// contains a generic service resonse or meshnet invitation respond
// response and returns the according error if any
func respondToInviteResponseToError(
	resp *pb.RespondToInviteResponse,
	email string,
) error {
	if resp == nil {
		return errors.New(AccountInternalError)
	}
	switch resp := resp.Response.(type) {
	case *pb.RespondToInviteResponse_Empty:
		return nil
	case *pb.RespondToInviteResponse_ServiceErrorCode:
		return serviceErrorCodeToError(
			resp.ServiceErrorCode,
		)
	case *pb.RespondToInviteResponse_RespondToInviteErrorCode:
		return respondToInviteErrorCodeToError(
			resp.RespondToInviteErrorCode,
			email,
		)
	case *pb.RespondToInviteResponse_MeshnetErrorCode:
		return meshnetErrorToError(resp.MeshnetErrorCode)
	default:
		return errors.New(AccountInternalError)
	}
}

// respondToInviteErrorCodeToError determines the human readable error
// by the error code provided
func respondToInviteErrorCodeToError(
	code pb.RespondToInviteErrorCode,
	email string,
) error {
	switch code {
	case pb.RespondToInviteErrorCode_NO_SUCH_INVITATION:
		return fmt.Errorf(
			MsgMeshnetInviteNoInvitationFound,
			email,
		)
	case pb.RespondToInviteErrorCode_DEVICE_COUNT:
		return errors.New(MsgMeshnetInviteAcceptDeviceCount)
	case pb.RespondToInviteErrorCode_UNKNOWN:
		fallthrough
	default:
		return errors.New(AccountInternalError)
	}
}
