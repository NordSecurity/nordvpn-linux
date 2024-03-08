package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/NordSecurity/nordvpn-linux/fileshare"
	"github.com/NordSecurity/nordvpn-linux/fileshare/pb"
	"golang.org/x/exp/slices"

	"github.com/docker/go-units"
	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

func (c *cmd) getTransfers() ([]*pb.Transfer, error) {
	listClient, err := c.fileshareClient.List(context.Background(), &pb.Empty{})
	if err != nil {
		return nil, formatError(err)
	}

	transfers := []*pb.Transfer{}

	for {
		resp, err := listClient.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, formatError(err)
		}
		if err := getFileshareResponseToError(resp.GetError()); err != nil {
			return nil, formatError(err)
		}

		transfers = append(transfers, resp.GetTransfers()...)
	}

	return transfers, nil
}

// FileshareList rpc
func (c *cmd) FileshareList(ctx *cli.Context) error {
	// fmt.Println("fileshare list")
	transfers, err := c.getTransfers()
	if err != nil {
		return formatError(err)
	}

	if id := ctx.Args().First(); id != "" {
		matchIDFunc := func(t *pb.Transfer) bool { return t.GetId() == id }
		idx := slices.IndexFunc(transfers, matchIDFunc)
		if idx == -1 {
			return errors.New(MsgFileshareTransferNotFound)
		}

		fmt.Println(strings.TrimSpace(transferToOutputString(transfers[idx])))
		return nil
	}

	printIn, printOut := true, true
	if ctx.IsSet(flagFileshareListIn) || ctx.IsSet(flagFileshareListOut) {
		printIn = ctx.IsSet(flagFileshareListIn)
		printOut = ctx.IsSet(flagFileshareListOut)
	}
	fmt.Println(strings.TrimSpace(transfersToOutputString(transfers, printIn, printOut)))
	return nil
}

// Autocompletes first argument as transfer id and following arguments as files from selected transfer
func (c *cmd) fileshareAutoCompleteTransfers(ctx *cli.Context, direction pb.Direction, statusFilter func(pb.Status) bool) {
	// Use default autocomplete for path argument
	// -2 because the last arg is always '--generate-bash-completion'
	if len(os.Args) >= 2 && os.Args[len(os.Args)-2] == "--"+flagFilesharePath {
		c.AutocompleteFilepaths(ctx)
		return
	}

	transfers, err := c.getTransfers()
	if err != nil {
		return
	}

	if ctx.NArg() == 0 {
		// Autocomplete transfer id
		for _, transfer := range transfers {
			if (transfer.GetDirection() == direction || direction == pb.Direction_UNKNOWN_DIRECTION) &&
				statusFilter(transfer.Status) {
				fmt.Println(transfer.GetId())
			}
		}
	} else {
		// Autocomplete transfer files
		for _, transfer := range transfers {
			if transfer.Id == ctx.Args().First() {
				fileshare.ForAllFiles(transfer.Files, func(f *pb.File) {
					fmt.Println(f.Path)
				})
			}
		}
	}
}

// FileshareAutoCompleteTransfersList does transfer id and files autocompletion for `fileshare list`
func (c *cmd) FileshareAutoCompleteTransfersList(ctx *cli.Context) {
	c.fileshareAutoCompleteTransfers(ctx, pb.Direction_UNKNOWN_DIRECTION, func(s pb.Status) bool {
		return true
	})
}

// FileshareAutoCompleteTransfersAccept does transfer id and files autocompletion for `fileshare accept`
func (c *cmd) FileshareAutoCompleteTransfersAccept(ctx *cli.Context) {
	c.fileshareAutoCompleteTransfers(ctx, pb.Direction_INCOMING, func(s pb.Status) bool {
		return s == pb.Status_REQUESTED
	})
}

// FileshareAutoCompleteTransfersCancel does transfer id and files autocompletion for `fileshare cancel`
func (c *cmd) FileshareAutoCompleteTransfersCancel(ctx *cli.Context) {
	c.fileshareAutoCompleteTransfers(ctx, pb.Direction_UNKNOWN_DIRECTION, func(s pb.Status) bool {
		return s == pb.Status_REQUESTED || s == pb.Status_ONGOING
	})
}

func transferToOutputString(transfer *pb.Transfer) string {
	var builder strings.Builder
	const (
		minwidth = 0
		tabwidth = 1
		padding  = 1
		padchar  = ' '
		flags    = 0
	)
	tableWriter := tabwriter.NewWriter(&builder, minwidth, tabwidth, padding, padchar, flags)
	headingCol := color.New(color.Bold)

	builder.WriteString(headingCol.Sprintf("File list:\n"))
	fmt.Fprintf(tableWriter, "file\tsize\tstatus\t\n")
	for _, file := range transfer.Files {
		progress := ""
		if file.Status == pb.Status_ONGOING && file.Size > 0 {
			progress = " " + fmt.Sprintf("%d%%",
				uint16(float64(file.Transferred)/float64(file.Size)*100))
		}
		fmt.Fprintf(tableWriter, "%s\t%s\t%s%s\t\n",
			file.GetPath(),
			units.HumanSize(float64(file.GetSize())),
			fileshare.GetTransferFileStatus(file, transfer.Direction == pb.Direction_INCOMING),
			progress,
		)
	}

	if err := tableWriter.Flush(); err != nil {
		log.Println(err)
	}

	return builder.String()
}

func transfersToOutputString(transfers []*pb.Transfer, printIn, printOut bool) string {
	var builder strings.Builder
	const (
		minwidth = 0
		tabwidth = 1
		padding  = 1
		padchar  = ' '
		flags    = 0
	)
	tableWriter := tabwriter.NewWriter(&builder, minwidth, tabwidth, padding, padchar, flags)
	headingCol := color.New(color.Bold)

	if printIn {
		builder.WriteString(headingCol.Sprintf("Incoming:\n"))
		buildTransferTable(tableWriter, transfers, pb.Direction_INCOMING)
	}

	if printIn && printOut {
		builder.WriteByte('\n')
	}

	if printOut {
		builder.WriteString(headingCol.Sprintf("Outgoing:\n"))
		buildTransferTable(tableWriter, transfers, pb.Direction_OUTGOING)
	}

	return builder.String()
}

func buildTransferTable(writer *tabwriter.Writer, transfers []*pb.Transfer, direction pb.Direction) {
	fmt.Fprintf(writer, "id\tpeer\tfiles\tsize\tstatus\tpath\t\n")
	for _, transfer := range transfers {
		if transfer.GetDirection() != direction {
			continue
		}

		fileCount := len(transfer.Files)
		fileSize := calcTransferSize(transfer.Files)

		progress := ""
		if transfer.Status == pb.Status_ONGOING {
			progress = " " + calcTransferProgressPercent(transfer)
		}

		fmt.Fprintf(writer, "%s\t%s\t%d\t%s\t%s%s\t%s\t\n",
			transfer.GetId(),
			transfer.GetPeer(),
			fileCount,
			fileSize,
			fileshare.GetTransferStatus(transfer),
			progress,
			transfer.GetPath(),
		)
	}

	if err := writer.Flush(); err != nil {
		log.Println(err)
	}
}

func calcTransferProgressPercent(tr *pb.Transfer) string {
	progress := uint16(0)
	transferred := uint64(0)
	totalSize := uint64(0)
	for _, file := range tr.Files {
		if file.Status != pb.Status_CANCELED {
			transferred += file.Transferred
			totalSize += file.Size
		}
	}
	if totalSize > 0 {
		progress = uint16(float64(transferred) / float64(totalSize) * 100)
	}
	return fmt.Sprintf("%d%%", progress)
}

func calcTransferSize(files []*pb.File) string {
	var size uint64
	fileshare.ForAllFiles(files, func(f *pb.File) {
		size += f.Size
	})
	return units.HumanSize(float64(size))
}
