package fusefs

import (
	"context"
	"syscall"
	"time"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"

	"github.com/NordSecurity/nordvpn-linux/fileshare/utils"
	meshpb "github.com/NordSecurity/nordvpn-linux/meshnet/pb"
)

type peerDir struct {
	fs.Inode

	peer           *meshpb.Peer
	fileSender     fileSender
	fuseBackingDir string
	transfers      transferTracker
}

var (
	_ fs.NodeGetattrer = (*peerDir)(nil)
	_ fs.NodeReaddirer = (*peerDir)(nil)
	_ fs.NodeLookuper  = (*peerDir)(nil)
)

func (d *peerDir) Getattr(_ context.Context, _ fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	out.Mode = syscall.S_IFDIR | 0o755
	return 0
}

func (d *peerDir) Readdir(_ context.Context) (fs.DirStream, syscall.Errno) {
	return fs.NewListDirStream([]fuse.DirEntry{
		{Name: "sent", Mode: syscall.S_IFDIR},
		{Name: "received", Mode: syscall.S_IFDIR},
	}), 0
}

func (d *peerDir) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	out.Mode = syscall.S_IFDIR | 0o755
	out.SetAttrTimeout(time.Second)
	out.SetEntryTimeout(time.Second)
	switch name {
	case "sent":
		// XXX: maybe it should be an object that takes fileSender, fuseBackingDir and transfers and produces new
		// instances based on d.peer - same for receivedDir
		node := &sentDir{
			peer:           d.peer,
			fileSender:     d.fileSender,
			fuseBackingDir: d.fuseBackingDir,
			transfers:      d.transfers,
		}
		attr := fs.StableAttr{Mode: syscall.S_IFDIR, Ino: utils.StableIno("sent", d.peer.Identifier)}
		return d.NewInode(ctx, node, attr), 0
	case "received":
		node := &receivedDir{peer: d.peer, fuseBackingDir: d.fuseBackingDir}
		attr := fs.StableAttr{Mode: syscall.S_IFDIR, Ino: utils.StableIno("received", d.peer.Identifier)}
		return d.NewInode(ctx, node, attr), 0
	default:
		return nil, syscall.ENOENT
	}
}
