package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync/atomic"
	"syscall"

	"github.com/NordSecurity/nordvpn-linux/client"
	"github.com/NordSecurity/nordvpn-linux/fileshare/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

type transferStatusClient interface {
	Recv() (*pb.StatusResponse, error)
}

func statusLoop(fileshareClient pb.FileshareClient, client transferStatusClient, transferID string) error {
	transferErrorChan := make(chan error)
	cancelChan := make(chan os.Signal, 1)
	signal.Notify(cancelChan, syscall.SIGINT)
	var canceledBySignal atomic.Bool

	go func() {
		defer close(transferErrorChan)
		for {
			resp, err := client.Recv()

			if err != nil {
				if err != io.EOF {
					transferErrorChan <- err
				}
				return
			}

			if fileshareError := resp.GetError(); fileshareError != nil {
				if err := getFileshareResponseToError(fileshareError); err != nil {
					transferErrorChan <- err
					return
				}
			}

			//exhaustive:ignore
			switch resp.Status {
			case pb.Status_ONGOING:
				fmt.Printf("\r"+MsgFileshareProgressOngoing, resp.TransferId, resp.Progress)
			case pb.Status_SUCCESS:
				fmt.Printf("\r"+MsgFileshareProgressFinished+"\n", resp.TransferId)
				return
			case pb.Status_FINISHED_WITH_ERRORS:
				// The transfer request might not have reached the peer yet, error happens then
				if !canceledBySignal.Load() {
					fmt.Printf("\r"+MsgFileshareProgressFinishedErrors+"\n", resp.TransferId)
				}
				return
			case pb.Status_CANCELED_BY_PEER:
				fmt.Printf("\r"+MsgFileshareProgressCanceledByPeer+"\n", resp.TransferId)
				return
			case pb.Status_CANCELED:
				if !canceledBySignal.Load() {
					fmt.Printf("\r"+MsgFileshareProgressCanceled+"\n", resp.TransferId)
				}
				return
			default:
			}
		}
	}()

	select {
	case <-cancelChan:
		canceledBySignal.Store(true)
		resp, err := fileshareClient.Cancel(context.Background(), &pb.CancelRequest{TransferId: transferID})
		if err != nil {
			return formatError(err)
		}

		if err := getFileshareResponseToError(resp); err != nil {
			return formatError(err)
		}

		color.Green("\n%s", MsgFileshareCancelSuccess)

		return nil
	case err := <-transferErrorChan:
		return err
	}
}

// IsFileshareDaemonReachable returns error if fileshare daemon is not reachable, daemon not running
// being the most likely cause
func (c *cmd) IsFileshareDaemonReachable(ctx *cli.Context) error {
	_, err := c.fileshareClient.Ping(context.Background(), &pb.Empty{})

	if err != nil {
		if strings.Contains(err.Error(), "no such file or directory") {
			return formatError(fmt.Errorf(MsgFileshareSocketNotFound))
		}

		return formatError(fmt.Errorf(internal.UnhandledMessage))
	}

	return nil
}

// FileshareSend rpc
func (c *cmd) FileshareSend(ctx *cli.Context) error {
	args := ctx.Args()

	if args.Len() < 2 {
		return argsParseError(ctx)
	}

	absPaths := []string{}
	for _, path := range args.Slice()[1:] {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return fmt.Errorf(MsgFileshareInvalidPath, formatError(err))
		}
		absPaths = append(absPaths, absPath)
	}

	// disable spinner, we will show message to the user instead
	c.loaderInterceptor.enabled = false
	sendContext, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	client, err := c.fileshareClient.Send(sendContext, &pb.SendRequest{
		Peer:   args.First(),
		Paths:  absPaths,
		Silent: ctx.IsSet(flagFileshareNoWait),
	})
	if err != nil {
		return formatError(err)
	}

	// check first response to determine that transfer was started succesfully
	resp, err := client.Recv()
	if err != nil {
		return formatError(err)
	}

	if resp.GetError() != nil {
		if err := getFileshareResponseToError(resp.GetError()); err != nil {
			return formatError(err)
		}
	}

	if ctx.IsSet(flagFileshareNoWait) {
		color.Green(MsgFileshareSendNoWait, resp.TransferId)
		return nil
	}

	fmt.Printf("\r%s", MsgFileshareWaitAccept)

	return statusLoop(c.fileshareClient, client, resp.TransferId)
}

// FileshareAccept rpc
func (c *cmd) FileshareAccept(ctx *cli.Context) error {
	args := ctx.Args()

	if args.Len() < 1 {
		return argsParseError(ctx)
	}

	var path string
	if ctx.IsSet(flagFilesharePath) {
		var err error
		path, err = filepath.Abs(ctx.String(flagFilesharePath))
		if err != nil {
			return fmt.Errorf(MsgFileshareInvalidPath, formatError(err))
		}
	} else {
		downloads, ok := os.LookupEnv("XDG_DOWNLOAD_DIR")
		if !ok {
			home, err := os.UserHomeDir()
			if err != nil {
				log.Print("determining user home directory: " + err.Error())
				return fmt.Errorf(MsgFileshareAcceptHomeError)
			}
			path = filepath.Join(home, "Downloads")
		} else {
			path = downloads
		}
	}

	if !internal.FileExists(path) {
		return fmt.Errorf(MsgFilesharePathNotFound, path)
	}

	// disable spinner, we will show message to the user instead
	c.loaderInterceptor.enabled = false
	transferID := args.First()
	acceptContext, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	client, err := c.fileshareClient.Accept(acceptContext, &pb.AcceptRequest{
		TransferId: transferID,
		DstPath:    path,
		Silent:     ctx.IsSet(flagFileshareNoWait),
		Files:      args.Tail(),
	})
	if err != nil {
		return formatError(err)
	}

	resp, err := client.Recv()
	if err != nil {
		return formatError(err)
	}

	if resp.GetError() != nil {
		if err := getFileshareResponseToError(resp.GetError()); err != nil {
			return formatError(err)
		}
	}

	if ctx.IsSet(flagFileshareNoWait) {
		color.Green(MsgFileshareAcceptNoWait)
		return nil
	}

	return statusLoop(c.fileshareClient, client, transferID)
}

// FileshareCancel rpc
func (c *cmd) FileshareCancel(ctx *cli.Context) error {
	if ctx.NArg() != 1 && ctx.NArg() != 2 {
		return formatError(argsCountError(ctx))
	}

	var resp *pb.Error
	var err error

	args := ctx.Args()
	switch args.Len() {
	case 1:
		resp, err = c.fileshareClient.Cancel(context.Background(), &pb.CancelRequest{TransferId: args.Get(0)})
	case 2:
		resp, err = c.fileshareClient.CancelFile(context.Background(), &pb.CancelFileRequest{TransferId: args.Get(0), FileId: args.Get(1)})
	default:
		return formatError(argsCountError(ctx))
	}

	if err != nil {
		return formatError(err)
	}

	if err := getFileshareResponseToError(resp); err != nil {
		return formatError(err)
	}

	color.Green(MsgFileshareCancelSuccess)

	return nil
}

func getFileshareResponseToError(resp *pb.Error) error {
	if resp == nil {
		return errors.New(AccountInternalError)
	}

	switch resp := resp.Response.(type) {
	case *pb.Error_Empty:
		return nil
	case *pb.Error_ServiceError:
		return fileshareServiceErrorCodeToError(resp.ServiceError)
	case *pb.Error_FileshareError:
		return fileshareErrorCodeToError(resp.FileshareError)
	default:
		return errors.New(AccountInternalError)
	}
}

// fileshareServiceErrorCodeToError determines the human readable error by
// the error code provided
func fileshareServiceErrorCodeToError(code pb.ServiceErrorCode) error {
	switch code {
	case pb.ServiceErrorCode_NOT_LOGGED_IN:
		return internal.ErrNotLoggedIn
	case pb.ServiceErrorCode_MESH_NOT_ENABLED:
		return errors.New(MsgMeshnetNotEnabled)
	case pb.ServiceErrorCode_INTERNAL_FAILURE:
		fallthrough
	default:
		return errors.New(AccountInternalError)
	}
}

// fileshareErrorCodeToError determines the human readable from the given
// error code
func fileshareErrorCodeToError(code pb.FileshareErrorCode) error {
	switch code {
	case pb.FileshareErrorCode_LIB_FAILURE:
		return errors.New(client.ConnectCantConnect)
	case pb.FileshareErrorCode_TRANSFER_NOT_FOUND:
		return errors.New(MsgFileshareTransferNotFound)
	case pb.FileshareErrorCode_INVALID_PEER:
		return errors.New(MsgFileshareInvalidPeer)
	case pb.FileshareErrorCode_FILE_NOT_FOUND:
		return errors.New(MsgFileshareFileNotFound)
	case pb.FileshareErrorCode_ACCEPT_ALL_FILES_FAILED:
		return errors.New(MsgFileshareAcceptAllError)
	case pb.FileshareErrorCode_ACCEPT_OUTGOING:
		return errors.New(MsgFileshareAcceptOutgoingError)
	case pb.FileshareErrorCode_ALREADY_ACCEPTED:
		return errors.New(MsgFileshareAlreadyAcceptedError)
	case pb.FileshareErrorCode_FILE_INVALIDATED:
		return errors.New(MsgFileshareFileInvalidated)
	case pb.FileshareErrorCode_TRANSFER_INVALIDATED:
		return errors.New(MsgFileshareTransferInvalidated)
	case pb.FileshareErrorCode_TOO_MANY_FILES:
		return errors.New(MsgTooManyFiles)
	case pb.FileshareErrorCode_DIRECTORY_TOO_DEEP:
		return errors.New(MsgDirectoryToDeep)
	case pb.FileshareErrorCode_SENDING_NOT_ALLOWED:
		return errors.New(MsgSendingNotAllowed)
	case pb.FileshareErrorCode_PEER_DISCONNECTED:
		return errors.New(MsgFileshareDisconnectedPeer)
	case pb.FileshareErrorCode_FILE_NOT_IN_PROGRESS:
		return errors.New(MsgFileNotInProgress)
	case pb.FileshareErrorCode_TRANSFER_NOT_CREATED:
		return errors.New(MsgTransferNotCreated)
	default:
		return errors.New(AccountInternalError)
	}
}
