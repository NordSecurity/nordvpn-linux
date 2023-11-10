package storage

import (
	"fmt"
	"time"

	"github.com/NordSecurity/nordvpn-linux/fileshare"
	"github.com/NordSecurity/nordvpn-linux/fileshare/pb"
)

type Libdrop struct {
	fileshare fileshare.Fileshare
}

func NewLibdrop(fileshare fileshare.Fileshare) *Libdrop {
	return &Libdrop{fileshare: fileshare}
}

func (l *Libdrop) Load() (map[string]*pb.Transfer, error) {
	libdropTransfers, err := l.fileshare.GetTransfersSince(time.Time{})
	if err != nil {
		return nil, fmt.Errorf("getting transfers from libdrop: %w", err)
	}

	transfers := map[string]*pb.Transfer{}
	for _, libdropTransfer := range libdropTransfers {
		transfers[libdropTransfer.ID] = fileshare.LibdropTransferToInternalTransfer(libdropTransfer)
	}

	return transfers, nil
}
