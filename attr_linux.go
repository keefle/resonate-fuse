// +build linux

package resonatefuse

import (
	"context"
	"log"
	"os"
	"syscall"
	"time"

	"bazil.org/fuse"
	"github.com/pkg/errors"
)

// Attr returns some attributes about the file
func (f *File) Attr(ctx context.Context, a *fuse.Attr) error {
	log.Println("Attring", f.FFNode.Name())

	info, err := os.Lstat(f.FFNode.fs.realify(f.FFNode.Path()))
	if err != nil {
		err = errors.Wrapf(err, "could not retrieve file (%v) info", f.FFNode.fs.realify(f.FFNode.Path()))
		log.Println(err)
		return fuse.ENOENT
	}

	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		err = errors.New("file system not supported")
		log.Println(err)
		return fuse.ENOENT
	}

	a.Inode = stat.Ino
	a.Nlink = uint32(stat.Nlink)
	a.Uid = stat.Uid
	a.Gid = stat.Gid
	a.Rdev = uint32(stat.Rdev)
	a.Mode = info.Mode()
	a.Size = uint64(info.Size())
	a.Atime = time.Unix(stat.Atim.Unix())
	a.Mtime = time.Unix(stat.Mtim.Unix())
	a.Ctime = time.Unix(stat.Ctim.Unix())
	a.Blocks = uint64(stat.Blocks)
	a.BlockSize = uint32(stat.Blksize)

	return nil
}
