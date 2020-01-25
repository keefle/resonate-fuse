package resonatefuse

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/pkg/errors"
)

// FS implements the hello world file system.
type FS struct {
	Rot *File
}

// Root returns the root directory
func (fs *FS) Root() (fs.Node, error) {
	return fs.Rot, nil
}

func NewFS(name string) *FS {
	return &FS{Rot: NewFile(NewDirectory(name, nil))}
}

// File is the building node of a filesystem
type File struct {
	node *FileTree
}

// NewFile constructs a new disk rapper around a filetree
func NewFile(ft *FileTree) *File {
	return &File{
		node: ft,
	}
}

// Attr returns some attributes about the file
func (f *File) Attr(ctx context.Context, a *fuse.Attr) error {
	log.Println("Attring", f.node.name)

	info, err := os.Lstat(realify(f.node.Path()))
	if err != nil {
		err = errors.Wrapf(err, "could not retrieve file (%v) info", realify(f.node.Path()))
		log.Println(err)
		return fuse.ENOENT
	}

	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		err = errors.New("file system not supported")
		log.Println(err)
		return fuse.ENOENT
	}

	a.Inode = f.node.ID()
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

// Lookup returns info about child
func (f *File) Lookup(ctx context.Context, name string) (fs.Node, error) {
	log.Println("Looking for", name, "in", f.node.name)
	child := f.node.Child(name)
	if child == nil {
		log.Println("lookup faild")
		return nil, fuse.ENOENT
	}

	return NewFile(child), nil
}

// Create creats a new file on disk and filetree
func (f *File) Create(ctx context.Context, req *fuse.CreateRequest, resp *fuse.CreateResponse) (fs.Node, fs.Handle, error) {
	log.Println("Creating", req.Name, "in", f.node.name)
	// First create the file and then add it to the tree (order is important)

	if err := touch(realify(filepath.Join(f.node.Path(), req.Name)), req.Mode); err != nil {
		err = errors.Errorf("could not add file %v to disk", req.Name)
		log.Println(err)
		return nil, nil, fuse.EIO
	}

	if err := f.node.CreateChild(req.Name); err != nil {
		err = errors.Errorf("could not add file %v to filetree", req.Name)
		log.Println(err)
		return nil, nil, fuse.EIO
	}

	return NewFile(f.node.Child(req.Name)), NewFile(f.node.Child(req.Name)), nil
}

// Remove removes file from disk and filetree
func (f *File) Remove(ctx context.Context, req *fuse.RemoveRequest) error {
	log.Println("Removing", req.Name, "in", f.node.name)
	// First remove the file from the tree then remove it from disk (order is important)

	child := f.node.Child(req.Name)
	if child == nil {
		return fuse.ENOENT
	}

	if child.Type() == fuse.DT_Dir && len(child.children) > 0 {
		return fuse.ENOENT

	}

	if err := f.node.RemoveChild(req.Name); err != nil {
		err = errors.Errorf("could not remove file %v from filetree", req.Name)
		log.Println(err)
		return fuse.EIO
	}

	if err := rm(realify(filepath.Join(f.node.Path(), req.Name))); err != nil {
		return fuse.ENOENT
	}

	return nil
}

func (f *File) Write(ctx context.Context, req *fuse.WriteRequest, resp *fuse.WriteResponse) error {
	log.Println("Writing", f.node.name)
	n, err := writeAt(realify(f.node.Path()), req.Data, req.Offset)
	if err != nil {
		resp.Size = n
		log.Println(err)
		return fuse.EIO
	}

	resp.Size = n

	return nil
}

// ReadDirAll returns all children
func (f *File) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	log.Println("ReadDirAlling", f.node.name)
	return f.node.Dirents(), nil
}

// ReadAll returns all bytes in file
func (f *File) ReadAll(ctx context.Context) ([]byte, error) {
	log.Println("ReadAlling", f.node.name)
	return readall(realify(f.node.Path()))
}
func (f *File) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
	log.Println("Reading", f.node.name)
	_, err := readAt(realify(f.node.Path()), resp.Data, req.Offset)
	if err != nil {
		log.Println(err)
		return fuse.EIO
	}

	return nil
}

// Rename moves a file from source to target
func (f *File) Rename(ctx context.Context, req *fuse.RenameRequest, newDir fs.Node) error {
	newParent := newDir.(*File).node
	source := req.OldName
	target := req.NewName
	log.Println("Renaming source", source, "in", f.node.Path(), "to", target, " in ", newParent.Path())

	if err := f.node.Rename(source, target, newParent); err != nil {
		err = errors.Wrapf(err, "could not rename file (%v) from (%v) to (%v)", source, f.node.Path(), newParent.Path())
		log.Println(err)
		return err
	}

	oldn := realify(filepath.Join(f.node.Path(), source))
	newn := realify(filepath.Join(newParent.Path(), target))
	log.Println("source:", oldn)
	log.Println("target:", newn)

	if err := os.Rename(oldn, newn); err != nil {
		err = errors.Wrapf(err, "could not rename file on disk (%v) from (%v) to %v", source, target, f.node.name)
		log.Println(err)
		return err
	}

	return nil
}

// Mkdir creats a directory
func (f *File) Mkdir(ctx context.Context, req *fuse.MkdirRequest) (fs.Node, error) {
	log.Println("Mkdiring", req.Name, "in", f.node.name)

	if err := mkdir(realify(filepath.Join(f.node.Path(), req.Name)), req.Mode); err != nil {
		return nil, errors.Errorf("could not create real dir %v", req.Name)
	}

	if err := f.node.CreateDirChild(req.Name); err != nil {
		return nil, errors.Errorf("could not create dir %v to filetree", req.Name)
	}

	child := f.node.Child(req.Name)

	return NewFile(child), nil
}

func (f *File) Link(ctx context.Context, req *fuse.LinkRequest, old fs.Node) (fs.Node, error) {
	oldnode := old.(*File).node
	log.Println("Linking", f.node.Name())

	if err := f.node.CreateChild(req.NewName); err != nil {
		return nil, fuse.ENOENT
	}

	child := f.node.Child(req.NewName)
	if child == nil {
		return nil, fuse.ENOENT
	}

	if err := os.Link(realify(oldnode.Path()), realify(filepath.Join(f.node.Path(), req.NewName))); err != nil {
		return nil, fuse.ENOENT
	}

	return NewFile(child), nil

}
func (f *File) Symlink(ctx context.Context, req *fuse.SymlinkRequest) (fs.Node, error) {
	log.Println("Symlinkig", f.node.Name())

	if err := os.Symlink(req.Target, realify(filepath.Join(f.node.Path(), req.NewName))); err != nil {
		log.Println("symlinl", err)
		return nil, fuse.ENOENT
	}

	if err := f.node.CreateLinkChild(req.NewName, req.Target); err != nil {
		log.Println("symlinl create link child", err)
		return nil, fuse.ENOENT
	}

	child := f.node.Child(req.NewName)

	return NewFile(child), nil
}

func (f *File) Readlink(ctx context.Context, req *fuse.ReadlinkRequest) (string, error) {
	if f.node.Type() != fuse.DT_Link {
		return "", fuse.ENOENT
	}

	return f.node.Link(), nil
}

func (f *File) Open(ctx context.Context, req *fuse.OpenRequest, resp *fuse.OpenResponse) (fs.Handle, error) {
	log.Println("Opening", f.node.name)
	return f, nil
}

// Fsync to be implemented
func (f *File) Fsync(ctx context.Context, req *fuse.FsyncRequest) error {
	log.Println("Fsyncing", f.node.name)
	return nil
}

// Flush to be implemented
func (f *File) Flush(ctx context.Context, req *fuse.FlushRequest) error {
	log.Println("Flushing", f.node.name)
	return nil
}

// Release to be implemented
func (f *File) Release(ctx context.Context, req *fuse.ReleaseRequest) error {
	log.Println("Releasing", f.node.name)
	return nil
}

// Setattr to be implemented
func (f *File) Setattr(ctx context.Context, req *fuse.SetattrRequest, resp *fuse.SetattrResponse) error {
	log.Println("Setattring", f.node.name)
	if !req.Valid.Mode() {
		return nil
	}

	if err := os.Chmod(realify(f.node.Path()), req.Mode); err != nil {
		err = errors.Wrapf(err, "could not setattr chmod file")
		log.Println(err)
		return fuse.EIO

	}

	if err := os.Chtimes(realify(f.node.Path()), req.Atime, req.Mtime); err != nil {
		err = errors.Wrapf(err, "could not setattr chtimes file")
		log.Println(err)
		return fuse.EIO
	}

	// NOTE: Changing owner not supported yet
	// TODO: Invistigate previlage escliation
	// if err := os.Chown(realify(f.node.Path()), int(req.Uid), int(req.Gid)); err != nil {
	//  err = return errors.Wrapf(err, "could not setattr chown file to new id (%v) gid (%v)", req.Uid, req.Gid)
	//  log.Println(err)
	// 	return err
	// }

	return nil
}

var _ fs.HandleFlusher = (*File)(nil)
var _ fs.HandleReadDirAller = (*File)(nil)
var _ fs.HandleReader = (*File)(nil)
var _ fs.HandleReadAller = (*File)(nil)
var _ fs.HandleReleaser = (*File)(nil)
var _ fs.HandleWriter = (*File)(nil)

var _ fs.Node = (*File)(nil)
var _ fs.NodeCreater = (*File)(nil)
var _ fs.NodeFsyncer = (*File)(nil)
var _ fs.NodeLinker = (*File)(nil)
var _ fs.NodeMkdirer = (*File)(nil)
var _ fs.NodeOpener = (*File)(nil)
var _ fs.NodeReadlinker = (*File)(nil)
var _ fs.NodeRemover = (*File)(nil)
var _ fs.NodeRenamer = (*File)(nil)
var _ fs.NodeSetattrer = (*File)(nil)
var _ fs.NodeStringLookuper = (*File)(nil)
var _ fs.NodeSymlinker = (*File)(nil)
