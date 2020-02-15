package resonatefuse

import (
	"context"
	"fmt"
	"log"
	"path/filepath"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

// FS implements the hello world file system.
type FS struct {
	root   *File
	origin string

	createHook  CreateHook
	writeHook   WriteHook
	removeHook  RemoveHook
	mkdirHook   MkdirHook
	renameHook  RenameHook
	linkHook    LinkHook
	symlinkHook SymlinkHook
	setattrHook SetattrHook
}

// Root returns the root directory
func (fs *FS) Root() (fs.Node, error) {
	return fs.root, nil
}

func (fs *FS) Origin() string {
	return fs.origin
}

func (fs *FS) Location() string {
	return fmt.Sprintf("%v-resonate", fs.origin)
}

func NewFS(name string, opts ...Option) *FS {
	if len(opts) != 8 {
		log.Fatal("could not create filesystem")
	}

	fs := &FS{origin: name}
	fs.root = NewFile(NewFFile(NewDirectory(fs.Location(), nil), fs))

	for _, opt := range opts {
		opt(fs)
	}

	return fs
}

func (fs *FS) realify(path string) string {
	return filepath.Join(fs.origin, path)
}

// File is the building node of a filesystem
type File struct {
	FFNode *FFile
}

// NewFile constructs a new disk rapper around a filetree
func NewFile(FFNode *FFile) *File {
	return &File{
		FFNode: FFNode,
	}
}

func (f *File) Child(name string) *File {

	child := f.FFNode.Child(name)
	if child == nil {
		return nil
	}

	return NewFile(child)
}

// Lookup returns info about child
func (f *File) Lookup(ctx context.Context, name string) (fs.Node, error) {
	log.Println("Looking for", name, "in", f.FFNode.Name())
	child, err := f.FFNode.Lookup(name)
	if err != nil {
		log.Println("lookup faild")
		return nil, fuse.ENOENT
	}

	return NewFile(child), nil
}

// Create creats a new file on disk and filetree
func (f *File) Create(ctx context.Context, req *fuse.CreateRequest, resp *fuse.CreateResponse) (fs.Node, fs.Handle, error) {
	log.Println("Creating", req.Name, "in", f.FFNode.Name())
	// First create the file and then add it to the tree (order is important)
	err := f.FFNode.fs.createHook(&CreateRequest{f.FFNode.Path(), req.Name, req.Mode})
	if err != nil {
		return nil, nil, fuse.EIO
	}

	child, err := f.FFNode.Create(req.Name, req.Mode)
	if err != nil {
		return nil, nil, fuse.EIO
	}

	return NewFile(child), NewFile(child), nil
}

// Remove removes file from disk and filetree
func (f *File) Remove(ctx context.Context, req *fuse.RemoveRequest) error {
	log.Println("Removing", req.Name, "in", f.FFNode.Name())
	// First remove the file from the tree then remove it from disk (order is important)

	err := f.FFNode.fs.removeHook(&RemoveRequest{f.FFNode.Path(), req.Name})
	if err != nil {
		return fuse.EIO
	}

	if err := f.FFNode.Remove(req.Name); err != nil {
		return fuse.ENOENT
	}

	return nil
}

func (f *File) Write(ctx context.Context, req *fuse.WriteRequest, resp *fuse.WriteResponse) error {
	log.Println("Writing", f.FFNode.Name())

	err := f.FFNode.fs.writeHook(&WriteRequest{Path: f.FFNode.Path(), Data: req.Data, Offset: req.Offset})
	if err != nil {
		return fuse.EIO
	}

	n, err := f.FFNode.Write(req.Data, req.Offset)

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
	log.Println("ReadDirAlling", f.FFNode.Name())
	return f.FFNode.ReadDirAll()
}

// ReadAll returns all bytes in file
func (f *File) ReadAll(ctx context.Context) ([]byte, error) {
	log.Println("ReadAlling", f.FFNode.Name())
	return f.FFNode.ReadAll()
}
func (f *File) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
	log.Println("Reading", f.FFNode.Name())
	if err := f.FFNode.Read(resp.Data, req.Offset); err != nil {
		log.Println(err)
		return fuse.EIO
	}

	return nil
}

// Rename moves a file from source to target
func (f *File) Rename(ctx context.Context, req *fuse.RenameRequest, newDir fs.Node) error {
	log.Println("Renaming source", req.OldName, "in", f.FFNode.Path(), "to", req.NewName)

	err := f.FFNode.fs.renameHook(&RenameRequest{f.FFNode.Path(), req.OldName, req.NewName, newDir.(*File).FFNode.Path()})
	if err != nil {
		return fuse.EIO
	}
	err = f.FFNode.Rename(req.OldName, req.NewName, newDir.(*File).FFNode)
	if err != nil {
		return fuse.EIO
	}

	return nil
}

// Mkdir creats a directory
func (f *File) Mkdir(ctx context.Context, req *fuse.MkdirRequest) (fs.Node, error) {
	log.Println("Mkdiring", req.Name, "in", f.FFNode.Name())

	err := f.FFNode.fs.mkdirHook(&MkdirRequest{f.FFNode.Path(), req.Name, req.Mode})
	if err != nil {
		return nil, fuse.EIO
	}
	dir, err := f.FFNode.Mkdir(req.Name, req.Mode)
	if err != nil {
		return nil, fuse.EIO
	}

	return NewFile(dir), nil
}

func (f *File) Link(ctx context.Context, req *fuse.LinkRequest, old fs.Node) (fs.Node, error) {
	oldnode := old.(*File)
	log.Println("Linking", f.FFNode.Name())

	err := f.FFNode.fs.linkHook(&LinkRequest{Path: f.FFNode.Path(), NewName: req.NewName, Old: oldnode.FFNode.Path()})
	if err != nil {
		return nil, fuse.EIO
	}

	link, err := f.FFNode.Link(req.NewName, oldnode.FFNode)
	if err != nil {
		return nil, fuse.EIO
	}

	return NewFile(link), nil

}
func (f *File) Symlink(ctx context.Context, req *fuse.SymlinkRequest) (fs.Node, error) {
	log.Println("Symlinkig", f.FFNode.Name())

	err := f.FFNode.fs.symlinkHook(&SymlinkRequest{Path: f.FFNode.Path(), Target: req.Target, NewName: req.NewName})
	if err != nil {
		return nil, fuse.EIO
	}

	link, err := f.FFNode.Symlink(req.Target, req.NewName)
	if err != nil {
		return nil, fuse.EIO
	}

	return NewFile(link), nil
}

// Setattr to be implemented
func (f *File) Setattr(ctx context.Context, req *fuse.SetattrRequest, resp *fuse.SetattrResponse) error {
	log.Println("Setattring", f.FFNode.Name())
	if !req.Valid.Mode() {
		return nil
	}

	err := f.FFNode.fs.setattrHook(&SetattrRequest{Path: f.FFNode.Path(), Mode: req.Mode, Atime: req.Atime, Mtime: req.Mtime})
	if err != nil {
		return fuse.EIO
	}

	if err := f.FFNode.Setattr(req.Mode, req.Atime, req.Mtime); err != nil {
		return fuse.EPERM
	}

	return nil
}

func (f *File) Readlink(ctx context.Context, req *fuse.ReadlinkRequest) (string, error) {
	return f.FFNode.Readlink()
}

func (f *File) Open(ctx context.Context, req *fuse.OpenRequest, resp *fuse.OpenResponse) (fs.Handle, error) {
	log.Println("Opening", f.FFNode.Name())
	resp.Flags |= fuse.OpenDirectIO
	return f, nil
}

// Fsync to be implemented
func (f *File) Fsync(ctx context.Context, req *fuse.FsyncRequest) error {
	log.Println("Fsyncing", f.FFNode.Name())
	return nil
}

// Flush to be implemented
func (f *File) Flush(ctx context.Context, req *fuse.FlushRequest) error {
	log.Println("Flushing", f.FFNode.Name())
	return nil
}

// Release to be implemented
func (f *File) Release(ctx context.Context, req *fuse.ReleaseRequest) error {
	log.Println("Releasing", f.FFNode.Name())
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
