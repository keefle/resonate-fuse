package resonatefuse

import (
	"io/ioutil"
	"log"
	"os"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/pkg/errors"
)

type Volume struct {
	fs   *FS
	conn *fuse.Conn
	serv *fs.Server
}

func (v *Volume) Server() *fs.Server {
	return v.serv
}

func NewVolume(name string, ch CreateHook, wh WriteHook, rh RemoveHook, mh MkdirHook, rnh RenameHook) *Volume {
	if ch == nil {
		log.Fatal("createhook cannot be nil")
	}

	if wh == nil {
		log.Fatal("writeHook cannot be nil")
	}

	if rh == nil {
		log.Fatal("removeHook cannot be nil")
	}

	if mh == nil {
		log.Fatal("mkdirHook cannot be nil")
	}

	if rnh == nil {
		log.Fatal("renameHook cannot be nil")
	}

	v := &Volume{fs: NewFS(name, ch, wh, rh, mh, rnh)}
	return v
}

func (v *Volume) Fuse() *FS {
	return v.fs
}

func (v *Volume) mount() error {
	if err := mkdir(v.fs.Location(), os.ModeDir|0774); err != nil {
		return errors.Wrapf(err, "could not create mount point for volume (%v)", v.fs.Location())
	}

	c, err := fuse.Mount(
		v.fs.Location(),
		fuse.FSName("resonatefuse"),
		fuse.Subtype("resonatefuse"),
		fuse.LocalVolume(),
		fuse.VolumeName(v.fs.Location()),
	)

	if err != nil {
		return errors.Wrapf(err, "could not initalize mount volume (%v)", v.fs.Location())
	}

	// check if the mount process has an error to report
	<-c.Ready

	if c.MountError != nil {
		return errors.Wrapf(err, "could not mount volume (%v)", v.fs.Location())
	}

	v.conn = c
	v.serv = fs.New(v.conn, nil)

	return nil
}

func (v *Volume) Serve() error {

	if v.conn == nil {
		if err := v.mount(); err != nil {
			return errors.Wrapf(err, "could serve volume (%v)", v.fs.origin)
		}
	}

	if err := v.serv.Serve(v.fs); err != nil {
		return errors.Wrapf(err, "faced error when serving volume (%v)", v.fs.Location())
	}

	return nil
}

func (v *Volume) Stop() error {
	if v.conn == nil {
		return nil
	}

	// hacky line to force umount
	_, _ = ioutil.ReadDir(v.fs.Location())

	if err := v.conn.Close(); err != nil {
		return errors.Wrapf(err, "could not stop volume (%v)", v.fs.Location())
	}
	if err := fuse.Unmount(v.fs.Location()); err != nil {
		return errors.Wrapf(err, "could not unmount volume (%v)", v.fs.Location())
	}

	if err := rm(v.fs.Location()); err != nil {
		return errors.Wrapf(err, "could not remove mountpoint of volume (%v)", v.fs.Location())
	}

	v.conn = nil
	v.serv = nil

	return nil
}
