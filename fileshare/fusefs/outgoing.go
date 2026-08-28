package fusefs

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"

	"github.com/NordSecurity/nordvpn-linux/fileshare/utils"
	meshpb "github.com/NordSecurity/nordvpn-linux/meshnet/pb"
)

type outgoingNode struct {
	fs.Inode
}

var _ fs.NodeGetattrer = (*outgoingNode)(nil)

func (n *outgoingNode) Getattr(_ context.Context, _ fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	out.Mode = syscall.S_IFREG | 0o600
	return 0
}

type outgoingFileHandle struct {
	peer       *meshpb.Peer
	fileSender fileSender
	tmp        *os.File
	name       string
	sentDir    string
	transfers  transferTracker

	once   sync.Once
	result syscall.Errno
}

var (
	_ fs.FileWriter  = (*outgoingFileHandle)(nil)
	_ fs.FileFlusher = (*outgoingFileHandle)(nil)
)

func (f *outgoingFileHandle) Write(_ context.Context, data []byte, off int64) (uint32, syscall.Errno) {
	n, err := f.tmp.WriteAt(data, off)
	if err != nil {
		return uint32(n), syscall.EIO
	}
	return uint32(n), 0
}

func (f *outgoingFileHandle) Flush(_ context.Context) syscall.Errno {
	f.once.Do(func() {
		f.result = f.send()
	})
	return f.result
}

func (f *outgoingFileHandle) send() syscall.Errno {
	if err := f.tmp.Close(); err != nil {
		os.Remove(f.tmp.Name()) // nolint:errcheck
		return syscall.EIO
	}

	addr, err := utils.PeerAddr(f.peer)
	if err != nil {
		os.Remove(f.tmp.Name()) // nolint:errcheck
		return syscall.EINVAL
	}

	if f.fileSender == nil {
		os.Remove(f.tmp.Name()) // nolint:errcheck
		return syscall.ENODEV
	}

	finalPath := filepath.Join(f.sentDir, f.name)
	if err := os.Rename(f.tmp.Name(), finalPath); err != nil {
		os.Remove(f.tmp.Name()) // nolint:errcheck
		return syscall.EIO
	}

	transferID, err := f.fileSender.Send(addr, []string{finalPath})
	if err != nil {
		os.Remove(finalPath) // nolint:errcheck
		return syscall.EIO
	}

	done := f.transfers.RegisterPending(transferID, finalPath)

	if err := <-done; err != nil {
		os.Remove(finalPath) // nolint:errcheck
		if errno, ok := err.(syscall.Errno); ok {
			return errno
		}
		return syscall.EIO
	}

	return 0
}
