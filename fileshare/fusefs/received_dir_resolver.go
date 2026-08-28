package fusefs

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/NordSecurity/nordvpn-linux/fileshare/utils"
	meshpb "github.com/NordSecurity/nordvpn-linux/meshnet/pb"
)

type ReceivedDirResolver struct {
	meshClient     meshpb.MeshnetClient
	fuseBackingDir string
	isMounted      func() bool
}

func NewReceivedDirResolver(
	meshClient meshpb.MeshnetClient,
	fuseBackingDir string,
	// XXX: is there a better way to handle this instead of passing func?
	isMounted func() bool,
) *ReceivedDirResolver {
	return &ReceivedDirResolver{
		meshClient:     meshClient,
		fuseBackingDir: fuseBackingDir,
		isMounted:      isMounted,
	}
}

func (r *ReceivedDirResolver) ReceivedDir(peerIP string) (string, error) {
	if !r.isMounted() {
		return "", errors.New("fuse not mounted")
	}
	peer, err := utils.GetPeerByIP(r.meshClient, peerIP)
	if err != nil {
		return "", err
	}
	dir := filepath.Join(r.fuseBackingDir, utils.PeerName(peer), "received")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	return dir, nil
}
