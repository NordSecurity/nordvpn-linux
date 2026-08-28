// Package fusefs implements FUSE mounting functionality.
package fusefs

import (
	"context"
	"fmt"
	"net/netip"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/NordSecurity/nordvpn-linux/fileshare/utils"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"

	meshpb "github.com/NordSecurity/nordvpn-linux/meshnet/pb"
)

type transferTracker interface {
	RegisterPending(transferID, tmpPath string) <-chan error
}

type fileSender interface {
	Send(peer netip.Addr, paths []string) (string, error)
}

type FUSEMounter struct {
	fs.Inode

	mountpoint     string
	fuseBackingDir string
	meshClient     meshpb.MeshnetClient
	transfers      transferTracker
	fileSender     fileSender

	mu     sync.Mutex
	server *fuse.Server
}

func NewFUSEMounter(
	moutpoint string,
	fuseBackingDir string,
	meshClient meshpb.MeshnetClient,
	transfers transferTracker,
	fileSender fileSender,
) *FUSEMounter {
	return &FUSEMounter{
		mountpoint:     moutpoint,
		fuseBackingDir: fuseBackingDir,
		meshClient:     meshClient,
		transfers:      transfers,
		fileSender:     fileSender,
	}
}

func (fm *FUSEMounter) Mount() error {
	if err := os.MkdirAll(fm.mountpoint, 0o755); err != nil {
		return fmt.Errorf("creating mount point %s: %w", fm.mountpoint, err)
	}

	server, err := fs.Mount(fm.mountpoint, fm, &fs.Options{
		// Nodes report Uid/Gid 0 unless overridden per-attr - without these,
		// go-fuse defaults every inode's owner to root, and file managers
		// that stat-check writability client-side then see a root-owned dir
		// and deny paste based on the "other" bits.
		UID: uint32(os.Getuid()),
		GID: uint32(os.Getgid()),
		MountOptions: fuse.MountOptions{
			FsName: "nordvpn-meshnet",
			Name:   "nordvpn-meshnet",
		},
	})
	if err != nil {
		return fmt.Errorf("mounting meshnet fs at %s: %w", fm.mountpoint, err)
	}
	fm.mu.Lock()
	fm.server = server
	fm.mu.Unlock()
	return nil
}

func (fm *FUSEMounter) Unmount() error {
	fm.mu.Lock()
	server := fm.server
	fm.mu.Unlock()
	if server == nil {
		return nil
	}
	if err := server.Unmount(); err != nil {
		return err
	}
	fm.mu.Lock()
	fm.server = nil
	fm.mountpoint = ""
	fm.mu.Unlock()
	return nil
}

func (fm *FUSEMounter) IsMounted() bool {
	fm.mu.Lock()
	defer fm.mu.Unlock()
	return fm.server != nil
}

func (fm *FUSEMounter) Wait() {
	fm.mu.Lock()
	server := fm.server
	fm.mu.Unlock()
	if server != nil {
		server.Wait()
	}
}

func (fm *FUSEMounter) Readdir(ctx context.Context) (fs.DirStream, syscall.Errno) {
	peers, err := utils.GetPeers(fm.meshClient)
	if err != nil {
		return nil, syscall.EIO
	}
	entries := make([]fuse.DirEntry, 0, len(peers))
	for _, p := range peers {
		entries = append(entries, fuse.DirEntry{
			Name: utils.PeerName(p),
			Mode: syscall.S_IFDIR,
		})
	}
	return fs.NewListDirStream(entries), 0
}

func (fm *FUSEMounter) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	peers, err := utils.GetPeers(fm.meshClient)
	if err != nil {
		return nil, syscall.EIO
	}
	for _, p := range peers {
		if utils.PeerName(p) == name {
			// XXX: Do I need to push arguments through all those layers? Isn't there anything better?
			// XXX: Create constructors for all those structs like `peerDir`, `outgoing` etc.
			node := &peerDir{
				peer:           p,
				fileSender:     fm.fileSender,
				fuseBackingDir: fm.fuseBackingDir,
				transfers:      fm.transfers,
			}
			out.Mode = syscall.S_IFDIR | 0o755
			out.SetAttrTimeout(time.Second)
			out.SetEntryTimeout(time.Second)
			attr := fs.StableAttr{Mode: syscall.S_IFDIR, Ino: utils.StableIno("peer", p.Identifier)}
			return fm.NewInode(ctx, node, attr), 0
		}
	}
	return nil, syscall.ENOENT
}

func (fm *FUSEMounter) MountedPath(realPath string) (mountedPath string, ok bool) {
	fm.mu.Lock()
	mountpoint := fm.mountpoint
	fm.mu.Unlock()
	if mountpoint == "" {
		return "", false
	}

	rel, err := filepath.Rel(fm.fuseBackingDir, realPath)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", false
	}

	return filepath.Join(mountpoint, rel), true
}
