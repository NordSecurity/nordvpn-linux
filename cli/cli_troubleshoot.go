package cli

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

const TroubleshootUsageText = "Collects diagnostic logs and system information for troubleshooting. Share the generated file with NordVPN support to investigate the issue."

func (c *cmd) Troubleshoot(ctx *cli.Context) error {
	stream, err := c.client.CollectDiagnostics(context.Background(), &pb.Empty{})
	if err != nil {
		return formatError(err)
	}

	isTTY := isStdoutTerminal()

	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return errors.New(MsgTroubleshootFailure)
		}

		// Check for error in response
		if resp.ErrorCode != pb.DiagnosticsErrorCode_DIAGNOSTICS_ERROR_CODE_UNSPECIFIED {
			return formatError(errors.New(diagnosticsErrorMessage(resp.ErrorCode)))
		}

		// Final response: daemon signals completion by sending the file
		// path with no error.
		if resp.FilePath != "" {
			color.Green(MsgTroubleshootSuccess, resp.FilePath)
			color.Yellow(MsgTroubleshootDisclaimer)
			return nil
		}

		// Show progress if TTY
		if isTTY {
			fmt.Printf("%s\n", resp.Step)
		}
	}

	return nil
}

func diagnosticsErrorMessage(code pb.DiagnosticsErrorCode) string {
	switch code {
	case pb.DiagnosticsErrorCode_DIAGNOSTICS_ERROR_CODE_ZIP_TOO_LARGE:
		return "Diagnostics file exceeds 40 MB limit. Please contact support for assistance."
	case pb.DiagnosticsErrorCode_DIAGNOSTICS_ERROR_CODE_FAILED_TO_CREATE_ZIP:
		return "Failed to create diagnostics file."
	case pb.DiagnosticsErrorCode_DIAGNOSTICS_ERROR_CODE_CHOWN_FAILED:
		return "Failed to set diagnostics file ownership."
	case pb.DiagnosticsErrorCode_DIAGNOSTICS_ERROR_CODE_FAILED_TO_CLOSE_ZIP:
		return "Failed to finalize diagnostics file."
	case pb.DiagnosticsErrorCode_DIAGNOSTICS_ERROR_CODE_NO_DAEMON_LOG_SOURCE:
		return "We couldn't extract daemon logs automatically because the daemon was not started via systemd or snap. Contact our support team for help collecting logs manually."
	case pb.DiagnosticsErrorCode_DIAGNOSTICS_ERROR_CODE_UNSPECIFIED,
		pb.DiagnosticsErrorCode_DIAGNOSTICS_ERROR_CODE_INTERNAL,
		pb.DiagnosticsErrorCode_DIAGNOSTICS_ERROR_CODE_COLLECTION_FAILED:
		return MsgTroubleshootFailure
	}
	return MsgTroubleshootFailure
}
