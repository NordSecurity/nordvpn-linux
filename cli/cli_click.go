package cli

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

func waitForInput(timeout bool) {
	inputChan := make(chan interface{})
	go func() {
		bufio.NewReader(os.Stdin).ReadBytes('\n')
		close(inputChan)
	}()

	if !timeout {
		<-inputChan
		return
	}

	select {
	case <-inputChan:
	case <-time.After(10 * time.Second):
	}
}

// Click is hidden cmd used to open some info when clicking on desktop icon
func (c *cmd) Click(ctx *cli.Context) (err error) {
	defer func() {
		inputTimeout := false
		if err != nil {
			inputTimeout = false
			color.Red(err.Error())
		}

		fmt.Println()
		fmt.Println("Press 'Enter' to close this window.")
		fmt.Println("To use nordvpn - open new terminal and type command e.g. nordvpn connect")

		waitForInput(inputTimeout)
	}()

	if ctx.NArg() >= 1 {
		url, err := url.Parse(ctx.Args().First())
		if err != nil {
			return formatError(err)
		}

		if url.Scheme == "nordvpn-sl" && url.Host == "claim-online-purchase" {
			resp, err := c.client.ClaimOnlinePurchase(context.Background(), &pb.Empty{})
			if err != nil {
				return formatError(err)
			}

			if !resp.Success {
				return errors.New(ClaimOnlinePurchaseFailure)
			}

			color.Green(ClaimOnlinePurchaseSuccess)
			return nil
		} else if url.Scheme == "nordvpn" {
			// if arg is given
			// run the same as: login --callback %arg
			if err := c.oauth2(ctx); err != nil {
				return formatError(err)
			}
		}
	} else {
		cli.ShowAppHelp(ctx)
	}

	return nil
}
