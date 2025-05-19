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
		if strings.Contains(err.Error(), "permission denied") || strings.Contains(err.Error(), "connection reset by peer") {
			return internal.ErrSocketAccessDenied
		}
		if snapErr := RetrieveSnapConnsError(err); snapErr != nil {
			return err
		}
		if strings.Contains(err.Error(), internal.MissingConsentMsg) {
			// NOTE: CLI ping doesn't need to fail with consent error.
			// For CLI, specific calls need to fail to trigger consent flow.
			return nil
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
