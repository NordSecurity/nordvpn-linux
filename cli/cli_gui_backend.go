package cli

import (
	"context"
	"fmt"

	"github.com/NordSecurity/nordvpn-linux/state/pb"
	"github.com/urfave/cli/v2"
)

// FileshareSend rpc
func (c *cmd) SubscribeToStatus(ctx *cli.Context) error {
	fmt.Println("Subscribing to bakckend state.")
	srv, err := c.stateClient.Subscribe(context.Background(), &pb.Empty{})
	if err != nil {
		fmt.Println("Failed to subscribe to state changes: ", err)
	}

	fmt.Println("Subscribed to bakckend state.")
	for {
		stateUpdate, err := srv.Recv()
		if err != nil {
			fmt.Println("Failed to receive state update: ", err)
		} else {
			fmt.Printf("State update: %+v\n", stateUpdate)
		}
	}
}
