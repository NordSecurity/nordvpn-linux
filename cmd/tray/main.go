package main

import (
	"fmt"
	"time"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	meshpb "github.com/NordSecurity/nordvpn-linux/meshnet/pb"
	"github.com/NordSecurity/nordvpn-linux/tray"

	"github.com/NordSecurity/systray"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	DaemonURL = fmt.Sprintf("%s://%s", internal.Proto, internal.DaemonSocket)
)

func onExit() {
	if tray.DebugMode {
		now := time.Now()
		fmt.Println("Exit at", now.String())
	}
}

func main() {
	conn, err := grpc.Dial(
		DaemonURL,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err == nil {
		tray.Client = pb.NewDaemonClient(conn)
		tray.MeshClient = meshpb.NewMeshnetClient(conn)
		tray.FileshareClient = nil
	}

	tray.NotifyEnabled = true
	tray.DebugMode = true

	systray.Run(tray.OnReady, onExit)
}
