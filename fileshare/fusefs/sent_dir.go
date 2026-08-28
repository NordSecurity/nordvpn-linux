package fusefs

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"

	"github.com/NordSecurity/nordvpn-linux/fileshare/utils"
	meshpb "github.com/NordSecurity/nordvpn-linux/meshnet/pb"
)

type sentDir struct {
	fs.Inode

	peer           *meshpb.Peer
	fileSender     fileSender
	fuseBackingDir string
	transfers      transferTracker
}

var (
	_ fs.NodeGetattrer = (*sentDir)(nil)
	_ fs.NodeReaddirer = (*sentDir)(nil)
	_ fs.NodeLookuper  = (*sentDir)(nil)
	_ fs.NodeCreater   = (*sentDir)(nil)
)

func (d *sentDir) Getattr(_ context.Context, _ fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	out.Attr.Mode = syscall.S_IFDIR | 0o755
	return 0
}

func (d *sentDir) Readdir(_ context.Context) (fs.DirStream, syscall.Errno) {
	entries, err := os.ReadDir(d.realDir())
	if err != nil {
		if os.IsNotExist(err) {
			return fs.NewListDirStream(nil), 0
		}
		return nil, syscall.EIO
	}
	result := make([]fuse.DirEntry, 0, len(entries))
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), sendingPrefix) {
			continue
		}
		mode := uint32(syscall.S_IFREG)
		if e.IsDir() {
			mode = syscall.S_IFDIR
		}
		result = append(result, fuse.DirEntry{Name: e.Name(), Mode: mode})
	}
	return fs.NewListDirStream(result), 0
}

func (d *sentDir) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	path := filepath.Join(d.realDir(), name)
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil, syscall.ENOENT
	}
	if err != nil {
		return nil, syscall.EIO
	}
	mode := uint32(syscall.S_IFREG)
	if info.IsDir() {
		mode = syscall.S_IFDIR
	}
	out.Attr.Mode = mode | 0o400
	out.Attr.Size = uint64(info.Size())
	out.Attr.Mtime = uint64(info.ModTime().Unix())
	out.Attr.Mtimensec = uint32(info.ModTime().Nanosecond())
	out.SetAttrTimeout(time.Second)
	out.SetEntryTimeout(time.Second)
	node := &ReceivedFile{realPath: path}
	attr := fs.StableAttr{Mode: mode, Ino: utils.StableIno("file", path)}
	return d.NewInode(ctx, node, attr), 0
}

func (d *sentDir) Create(
	ctx context.Context,
	name string,
	_ uint32,
	_ uint32,
	out *fuse.EntryOut,
) (*fs.Inode, fs.FileHandle, uint32, syscall.Errno) {
	sentDir := d.realDir()
	if err := os.MkdirAll(sentDir, 0o700); err != nil {
		return nil, nil, 0, syscall.EIO
	}
	// tmp file lives in sentDir so the post-send rename stays on one filesystem.
	tmp, err := os.CreateTemp(sentDir, sendingPrefix+"*")
	if err != nil {
		return nil, nil, 0, syscall.EIO
	}
	out.Attr.Mode = syscall.S_IFREG | 0o600
	out.SetAttrTimeout(0)
	out.SetEntryTimeout(0)
	node := &outgoingNode{}
	inode := d.NewInode(ctx, node, fs.StableAttr{Mode: syscall.S_IFREG})
	fh := &outgoingFileHandle{
		peer:       d.peer,
		fileSender: d.fileSender,
		tmp:        tmp,
		name:       name,
		sentDir:    sentDir,
		transfers:  d.transfers,
	}
	return inode, fh, fuse.FOPEN_DIRECT_IO, 0
}

func (d *sentDir) realDir() string {
	return filepath.Join(d.fuseBackingDir, utils.PeerName(d.peer), "sent")
}

const sendingPrefix = ".nordvpn-meshfs-sending-"
