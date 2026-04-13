package tray

import (
	"context"

	"github.com/NordSecurity/nordvpn-linux/fileshare/fileshare_process"
	filesharepb "github.com/NordSecurity/nordvpn-linux/fileshare/pb"
	"github.com/NordSecurity/nordvpn-linux/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type FileshareManager struct {
	fileshareClient filesharepb.FileshareClient
}

// NewFileshare manager builds an empty FileshareManager
func NewFileshareManager() FileshareManager {
	return FileshareManager{fileshareClient: nil}
}

// UpdateFileshareConnection updates the fileshare gRPC connection based on the meshnetEnabled status
func (fs *FileshareManager) UpdateFileshareConnection(meshnetEnabled bool) {
	log.Info("Updating tray's fileshare connection", getFlagText(meshnetEnabled))
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
			log.Error("Error connecting to the NordVPN fileshare daemon:", err)
		}
	}
}

// SetNotifications sets the fileshare notifications on/off
func (fs *FileshareManager) SetNotifications(flag bool) {
	if fs.fileshareClient == nil {
		log.Warn("fileshare client not initialized")
		return
	}
	if _, err := fs.fileshareClient.SetNotifications(context.Background(), &filesharepb.SetNotificationsRequest{Enable: flag}); err != nil {
		log.Errorf("Setting fileshare notifications %s error: %s\n", getFlagText(flag), err)
	}
}
