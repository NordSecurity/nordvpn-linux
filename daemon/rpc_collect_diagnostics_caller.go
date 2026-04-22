package daemon

import (
	"context"
	"fmt"
	"os/user"
	"path/filepath"
	"strconv"
	"time"

	"github.com/NordSecurity/nordvpn-linux/internal"
	"google.golang.org/grpc/peer"
)

// diagnosticsCaller bundles the identity of the client that invoked
// CollectDiagnostics (resolved from the gRPC peer credentials) together with
// the zip file path that should be written for them.
type diagnosticsCaller struct {
	user    *user.User
	uid     uint32
	gid     uint32
	zipPath string
}

// resolveDiagnosticsCaller extracts the caller's UID/GID from the gRPC
// context, looks up their user record, and picks an output path for the
// diagnostics zip.
func resolveDiagnosticsCaller(ctx context.Context) (*diagnosticsCaller, error) {
	p, ok := peer.FromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("failed to get peer from context")
	}
	cred, ok := p.AuthInfo.(internal.UcredAuth)
	if !ok {
		return nil, fmt.Errorf("failed to get credentials from peer")
	}
	userInfo, err := user.LookupId(strconv.FormatUint(uint64(cred.Uid), 10))
	if err != nil {
		return nil, fmt.Errorf("failed to lookup user: %w", err)
	}
	return &diagnosticsCaller{
		user:    userInfo,
		uid:     cred.Uid,
		gid:     cred.Gid,
		zipPath: resolveZipFilePath(userInfo.HomeDir),
	}, nil
}

// resolveZipFilePath returns the full path of the diagnostics zip file,
// preferring the user's Downloads folder and falling back to their home dir.
// If the resolved directory is a symlink, /tmp is used instead to avoid
// writing through user-controlled symlinks.
func resolveZipFilePath(homeDir string) string {
	timestamp := time.Now().Format("20060102-150405")
	fileName := fmt.Sprintf("nordvpn-diagnostics-%s.zip", timestamp)

	outputDir := homeDir
	downloadsDir := filepath.Join(homeDir, "Downloads")
	if internal.FileExists(downloadsDir) && !internal.IsSymLink(downloadsDir) {
		outputDir = downloadsDir
	}

	if internal.IsSymLink(outputDir) {
		outputDir = "/tmp"
	}

	return filepath.Join(outputDir, fileName)
}
