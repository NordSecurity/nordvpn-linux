package main

import (
	"fmt"
	"time"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/tray"

	"github.com/NordSecurity/systray"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	DaemonURL = fmt.Sprintf("%s://%s", internal.Proto, internal.DaemonSocket)
)

func onExit() {
	now := time.Now()
	fmt.Println("Exit at", now.String())
}

func main() {
	conn, err := grpc.Dial(
		DaemonURL,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	var client pb.DaemonClient
	if err == nil {
		client = pb.NewDaemonClient(conn)
	} else {
		fmt.Printf("Error connecting to the NordVPN daemon: %s", err)
		return
	}

	ti := tray.NewTrayInstance(client)
	systray.Run(func() { tray.OnReady(ti) }, func() { onExit() })
}
