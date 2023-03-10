package cli

import (
	"context"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

func (c *cmd) Ping() error {
	resp, err := c.client.Ping(context.Background(), &pb.Empty{})
	if err != nil {
		if strings.Contains(err.Error(), "no such file or directory") {
			return internal.ErrSocketNotFound
		}
		if strings.Contains(err.Error(), "permission denied") {
			return internal.ErrSocketAccessDenied
		}
		return internal.ErrDaemonConnectionRefused
	}

	switch resp.Type {
	case internal.CodeOffline:
		return ErrInternetConnection
	case internal.CodeDaemonOffline:
		return internal.ErrDaemonConnectionRefused
	case internal.CodeOutdated:
		return ErrUpdateAvailable
	}

	return nil
}
