package storage

import (
	"fmt"
	"time"

	"github.com/NordSecurity/nordvpn-linux/fileshare"
	"github.com/NordSecurity/nordvpn-linux/fileshare/pb"
)

type Libdrop struct {
	storage fileshare.Storage
}

func NewLibdrop(storage fileshare.Storage) *Libdrop {
	return &Libdrop{storage: storage}
}

func (l *Libdrop) Load() (map[string]*pb.Transfer, error) {
	transfers, err := l.storage.Load()
	if err != nil {
		return nil, fmt.Errorf("getting transfers from libdrop: %w", err)
	}

	return transfers, nil
}

func (l *Libdrop) PurgeTransfersUntil(until time.Time) error {
	return l.storage.PurgeTransfersUntil(until)
}
