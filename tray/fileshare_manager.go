package tray

import (
	"context"
	"log"

	"github.com/NordSecurity/nordvpn-linux/fileshare/fileshare_process"
	filesharepb "github.com/NordSecurity/nordvpn-linux/fileshare/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type FileshareManager struct {
	fileshareClient filesharepb.FileshareClient
}

func (fs *FileshareManager) UpdateFileshareConnection(meshnetEnabled bool) {
	if !meshnetEnabled {
		fs.fileshareClient = nil
		return
	}

	if fs.fileshareClient == nil {
		// Meshnet is enabled, we must connect to the fileshare daemon
		fileShareConn, err := grpc.NewClient(
			fileshare_process.FileshareURL,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err == nil {
			fs.fileshareClient = filesharepb.NewFileshareClient(fileShareConn)
		} else {
			log.Println(internal.ErrorPrefix, "Error connecting to the NordVPN fileshare daemon:", err)
		}
	}
}

func (fs *FileshareManager) SetNotifications(flag bool) {
	if fs.fileshareClient != nil {
		_, err := fs.fileshareClient.SetNotifications(context.Background(), &filesharepb.SetNotificationsRequest{Enable: flag})
		if err != nil {
			log.Printf("%s Setting fileshare notifications %s error: %s", internal.ErrorPrefix, getFlagText(flag), err)
		}
	}
}
