package fusefs

import (
	"context"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"

	"github.com/NordSecurity/nordvpn-linux/fileshare/utils"
	meshpb "github.com/NordSecurity/nordvpn-linux/meshnet/pb"
)

type receivedDir struct {
	fs.Inode

	peer           *meshpb.Peer
	fuseBackingDir string
}

var (
	_ fs.NodeGetattrer = (*receivedDir)(nil)
	_ fs.NodeReaddirer = (*receivedDir)(nil)
	_ fs.NodeLookuper  = (*receivedDir)(nil)
)

func (d *receivedDir) Getattr(_ context.Context, _ fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	out.Attr.Mode = syscall.S_IFDIR | 0o755
	return 0
}

func (d *receivedDir) Readdir(_ context.Context) (fs.DirStream, syscall.Errno) {
	entries, err := os.ReadDir(d.realDir())
	if err != nil {
		if os.IsNotExist(err) {
			return fs.NewListDirStream(nil), 0
		}
		return nil, syscall.EIO
	}
	result := make([]fuse.DirEntry, 0, len(entries))
	for _, e := range entries {
		mode := uint32(syscall.S_IFREG)
		if e.IsDir() {
			mode = syscall.S_IFDIR
		}
		result = append(result, fuse.DirEntry{Name: e.Name(), Mode: mode})
	}
	return fs.NewListDirStream(result), 0
}

func (d *receivedDir) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
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

func (d *receivedDir) realDir() string {
	return filepath.Join(d.fuseBackingDir, utils.PeerName(d.peer), "received")
}

type ReceivedFile struct {
	fs.Inode
	realPath string
}

var (
	_ fs.NodeGetattrer = (*ReceivedFile)(nil)
	_ fs.NodeOpener    = (*ReceivedFile)(nil)
)

func (f *ReceivedFile) Getattr(_ context.Context, _ fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	info, err := os.Stat(f.realPath)
	if err != nil {
		return syscall.EIO
	}
	out.Attr.Mode = syscall.S_IFREG | 0o400
	out.Attr.Size = uint64(info.Size())
	out.Attr.Mtime = uint64(info.ModTime().Unix())
	out.Attr.Mtimensec = uint32(info.ModTime().Nanosecond())
	out.SetTimeout(time.Second)
	return 0
}

func (f *ReceivedFile) Open(_ context.Context, _ uint32) (fs.FileHandle, uint32, syscall.Errno) {
	file, err := os.Open(f.realPath)
	if err != nil {
		return nil, 0, syscall.EIO
	}
	return &ReceivedFileHandle{file: file}, fuse.FOPEN_KEEP_CACHE, 0
}

type ReceivedFileHandle struct {
	file *os.File
}

var (
	_ fs.FileReader   = (*ReceivedFileHandle)(nil)
	_ fs.FileReleaser = (*ReceivedFileHandle)(nil)
)

func (fh *ReceivedFileHandle) Read(_ context.Context, buf []byte, off int64) (fuse.ReadResult, syscall.Errno) {
	return fuse.ReadResultFd(fh.file.Fd(), off, len(buf)), 0
}

func (fh *ReceivedFileHandle) Release(_ context.Context) syscall.Errno {
	fh.file.Close() // nolint:errcheck
	return 0
}
