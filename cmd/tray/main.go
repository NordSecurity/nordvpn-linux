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

func onExit(ti *tray.Instance) {
	if ti.DebugMode {
		now := time.Now()
		fmt.Println("Exit at", now.String())
	}
}

func main() {
	var ti = tray.Instance{}

	conn, err := grpc.Dial(
		DaemonURL,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err == nil {
		ti.Client = pb.NewDaemonClient(conn)
		ti.MeshClient = meshpb.NewMeshnetClient(conn)
		ti.FileshareClient = nil
	}

	ti.NotifyEnabled = true

	systray.Run(func() { tray.OnReady(&ti) }, func() { onExit(&ti) })
}
