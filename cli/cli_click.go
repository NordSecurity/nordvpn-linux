package cli

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
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
		inputTimeout := true
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

		if strings.ToLower(url.Scheme) == "nordvpn" {
			switch strings.ToLower(url.Host) {
			case "claim-online-purchase":
				resp, err := c.client.ClaimOnlinePurchase(context.Background(), &pb.Empty{})
				if err != nil {
					return formatError(err)
				}

				if !resp.Success {
					return errors.New(ClaimOnlinePurchaseFailure)
				}

				color.Green(ClaimOnlinePurchaseSuccess)
				return nil

			case "login":
				// login can be regular, or after new account setup & vpn service purchase (signup)
				regularLogin := true
				if strings.ToLower(url.Query().Get("action")) == "signup" {
					regularLogin = false
				}

				// if arg is given
				// run the same as: login --callback %arg
				if err := c.oauth2(ctx, regularLogin); err != nil {
					return formatError(err)
				}
				return nil

			case "consent":
				// login takes care of the analytics consent flow and continues with login
				return c.Login(ctx)
			}
		}
	}

	// for all unhandled cases
	cli.ShowAppHelp(ctx)

	return nil
}
