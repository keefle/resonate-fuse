package resonatefuse

import (
	"context"
	"log"
	"os"
	"path/filepath"

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
	return &FS{Rot: NewFile(NewFileTree(name, nil))}
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
	log.Println("Attring ", f.node.name)
	a.Inode = f.node.ID()
	switch f.node.Type() {
	case fuse.DT_File:
		a.Mode = 0664
	case fuse.DT_Dir:
		a.Mode = os.ModeDir | 0775
	}

	info, err := os.Stat(realify(f.node.Path()))
	if err != nil {
		return errors.Wrapf(err, "could not retrieve file (%v) info", realify(f.node.Path()))
	}

	a.Size = uint64(info.Size())
	return nil
}

// Lookup returns info about child
func (f *File) Lookup(ctx context.Context, name string) (fs.Node, error) {
	log.Println("Looking for ", name, " in ", f.node.name)
	if child := f.node.Child(name); child != nil {
		return NewFile(child), nil
	}

	log.Println("Looking failed")

	return nil, fuse.ENOENT
}

// Create creats a new file on disk and filetree
func (f *File) Create(ctx context.Context, req *fuse.CreateRequest, resp *fuse.CreateResponse) (fs.Node, fs.Handle, error) {
	log.Println("Creating ", req.Name, " in ", f.node.name)
	// First create the file and then add it to the tree (order is important)

	if err := touch(realify(filepath.Join(f.node.Path(), req.Name))); err != nil {
		return nil, nil, errors.Errorf("could not add file %v to disk", req.Name)
	}

	if err := f.node.CreateChild(req.Name); err != nil {
		return nil, nil, errors.Errorf("could not add file %v to filetree", req.Name)
	}

	return NewFile(f.node.Child(req.Name)), NewFile(f.node.Child(req.Name)), nil
}

// Remove removes file from disk and filetree

func (f *File) Remove(ctx context.Context, req *fuse.RemoveRequest) error {
	log.Println("Removing ", req.Name, " in ", f.node.name)
	// First remove the file from the tree then remove it from disk (order is important)

	child := f.node.Child(req.Name)
	if child == nil {
		return fuse.ENOENT
	}

	if child.Type() == fuse.DT_Dir && len(child.children) > 0 {
		return fuse.ENOENT

	}

	if err := f.node.RemoveChild(req.Name); err != nil {
		return errors.Errorf("could not remove file %v from filetree", req.Name)
	}

	if err := rm(realify(filepath.Join(f.node.Path(), req.Name))); err != nil {
		return fuse.ENOENT
	}

	return nil
}

// WriteAt writes to a file
func (f *File) WriteAt(data []byte, offset int64) (int, error) {
	log.Println("Writing in ", f.node.name)
	file, err := os.OpenFile(realify(f.node.Path()), os.O_RDWR, 0664)
	if err != nil {
		return 0, errors.Errorf("could not open file %v: %v", f.node.Path(), err)
	}
	defer file.Close()

	n, err := file.WriteAt(data, offset)
	if err != nil {
		return n, errors.Errorf("could not write to file %v: %v", f.node.Path(), err)
	}

	return n, nil
}

func (f *File) Write(ctx context.Context, req *fuse.WriteRequest, resp *fuse.WriteResponse) error {
	n, err := f.WriteAt(req.Data, req.Offset)
	if err != nil {
		resp.Size = n
		return err
	}

	resp.Size = n

	return nil
}

// ReadDirAll returns all children
func (f *File) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	log.Println("ReadDirAlling in ", f.node.name)
	return f.node.Dirents(), nil
}

// ReadAll returns all bytes in file
func (f *File) ReadAll(ctx context.Context) ([]byte, error) {
	log.Println("ReadAlling ", f.node.name)
	return readall(realify(f.node.Path()))
}

// Rename moves a file from source to target
func (f *File) Rename(ctx context.Context, req *fuse.RenameRequest, newDir fs.Node) error {
	newParent := newDir.(*File).node
	source := req.OldName
	target := req.NewName
	log.Println("Renaming source", source, "in", f.node.Path(), "to", target, " in ", newParent.Path())

	if err := f.node.Rename(source, target, newParent); err != nil {
		return errors.Wrapf(err, "could not rename file (%v) from (%v) to (%v)", source, f.node.Path(), newParent.Path())
	}

	oldn := realify(filepath.Join(f.node.Path(), source))
	newn := realify(filepath.Join(newParent.Path(), target))
	log.Println("source:", oldn)
	log.Println("target:", newn)

	if err := os.Rename(oldn, newn); err != nil {
		return errors.Wrapf(err, "could not rename file on disk (%v) from (%v) to %v", source, target, f.node.name)
	}

	return nil
}

// Mkdir creats a directory
func (f *File) Mkdir(ctx context.Context, req *fuse.MkdirRequest) (fs.Node, error) {
	log.Println("Mkdiring ", req.Name, " in ", f.node.name)

	if err := mkdir(realify(filepath.Join(f.node.Path(), req.Name))); err != nil {
		return nil, errors.Errorf("could not create real dir %v", req.Name)
	}

	if err := f.node.CreateDirChild(req.Name); err != nil {
		return nil, errors.Errorf("could not create dir %v to filetree", req.Name)
	}

	child := f.node.Child(req.Name)

	return NewFile(child), nil
}

func (f *File) Open(ctx context.Context, req *fuse.OpenRequest, resp *fuse.OpenResponse) (fs.Handle, error) {
	log.Println("Opening ", f.node.name)
	return f, nil
}

// Fsync to be implemented
func (f *File) Fsync(ctx context.Context, req *fuse.FsyncRequest) error {
	log.Printf("in Fsync: %v", f.node.name)
	return nil
}

// Flush to be implemented
func (f *File) Flush(ctx context.Context, req *fuse.FlushRequest) error {
	log.Printf("in Flush: %v", f.node.name)
	return nil
}

// Release to be implemented
func (f *File) Release(ctx context.Context, req *fuse.ReleaseRequest) error {
	log.Printf("in Release: %v", f.node.name)
	return nil
}

// Setattr to be implemented
func (f *File) Setattr(ctx context.Context, req *fuse.SetattrRequest, resp *fuse.SetattrResponse) error {
	log.Printf("in Setattr: %v", f.node.name)
	return nil
}

var _ fs.HandleFlusher = (*File)(nil)
var _ fs.HandleReadDirAller = (*File)(nil)
var _ fs.HandleReleaser = (*File)(nil)
var _ fs.HandleWriter = (*File)(nil)

var _ fs.Node = (*File)(nil)
var _ fs.NodeCreater = (*File)(nil)
var _ fs.NodeFsyncer = (*File)(nil)
var _ fs.NodeMkdirer = (*File)(nil)
var _ fs.NodeOpener = (*File)(nil)
var _ fs.NodeRemover = (*File)(nil)
var _ fs.NodeRenamer = (*File)(nil)
var _ fs.NodeSetattrer = (*File)(nil)
var _ fs.NodeStringLookuper = (*File)(nil)
