package resonatefuse

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"bazil.org/fuse"
	"github.com/pkg/errors"
)

// FFile is the building node of a filesystem

type FFile struct {
	fs   *FS
	node *FileTree
}

// NewFFile constructs a new disk rapper around a filetree
func NewFFile(ft *FileTree, fs *FS) *FFile {
	return &FFile{
		node: ft,
		fs:   fs,
	}
}

func (f *FFile) Child(name string) *FFile {

	child := f.node.Child(name)
	if child == nil {
		return nil
	}

	return NewFFile(child, f.fs)
}

func (f *FFile) Name() string {
	return f.node.Name()
}

func (f *FFile) Path() string {
	return f.node.Path()
}

// ReadDirAll returns all children
func (f *FFile) ReadDirAll() ([]fuse.Dirent, error) {
	log.Println("ReadDirAlling", f.Name())
	return f.node.Dirents(), nil
}

// Lookup returns info about child
func (f *FFile) Lookup(name string) (*FFile, error) {
	log.Println("Looking for", name, "in", f.node.name)
	child := f.node.Child(name)
	if child == nil {
		log.Println("lookup faild")
		return nil, errors.Errorf("could not find child file (%v) during lookup", name)
	}

	return NewFFile(child, f.fs), nil
}

// Create creats a new file on disk and filetree
func (f *FFile) Create(name string, mode os.FileMode) (*FFile, error) {
	log.Println("Creating", name, "in", f.node.name)

	if err := Touch(f.fs.realify(filepath.Join(f.node.Path(), name)), mode); err != nil {
		return nil, errors.Wrapf(err, "could not add file %v to disk", name)
	}

	if err := f.node.CreateChild(name); err != nil {
		return nil, errors.Wrapf(err, "could not add file %v to filetree", name)
	}

	return NewFFile(f.node.Child(name), f.fs), nil
}

// Remove removes file from disk and filetree
func (f *FFile) Remove(name string) error {
	log.Println("Removing", name, "in", f.node.name)
	// First remove the file from the tree then remove it from disk (order is important)

	child := f.node.Child(name)
	if child == nil {
		return errors.Errorf("could not remove file (%v) as it does was not found", name)
	}

	if child.Type() == DIR && len(child.children) > 0 {
		return errors.Errorf("could not remove directory (%v) as it is not empty", name)
	}

	if err := f.node.RemoveChild(name); err != nil {
		return errors.Wrapf(err, "could not remove file %v from filetree", name)
	}

	if err := rm(f.fs.realify(filepath.Join(f.node.Path(), name))); err != nil {
		return errors.Wrapf(err, "could not remove file %v from disk", name)
	}

	return nil
}

func (f *FFile) Write(data []byte, offset int64) (int, error) {
	log.Println("Writing", f.node.name)

	n, err := writeAt(f.fs.realify(f.node.Path()), data, offset)
	if err != nil {
		log.Println(err)
		return n, errors.Wrapf(err, "could not write data to file (%v)", f.node.name)
	}

	return n, nil
}

// ReadAll returns all bytes in file
func (f *FFile) ReadAll() ([]byte, error) {
	log.Println("ReadAlling", f.node.name)
	return readall(f.fs.realify(f.node.Path()))
}

func (f *FFile) Read(data []byte, offset int64) error {
	log.Println("Reading", f.node.name)
	_, err := readAt(f.fs.realify(f.node.Path()), data, offset)
	if err != nil {
		log.Println(err)
		return errors.Wrapf(err, "could not read data from file (%v)", f.node.name)
	}

	return nil
}

// Rename moves a file from source to target
func (f *FFile) Rename(oldName, newName string, newDir *FFile) error {
	newParent := newDir.node
	source := oldName
	target := newName
	log.Println("Renaming source", source, "in", f.node.Path(), "to", target, " in ", newParent.Path())

	if err := f.node.Rename(source, target, newParent); err != nil {

		log.Println(err)
		return errors.Wrapf(err, "could not rename file (%v) from (%v) to (%v)", source, f.node.Path(), newParent.Path())
	}

	oldn := f.fs.realify(filepath.Join(f.node.Path(), source))
	newn := f.fs.realify(filepath.Join(newParent.Path(), target))
	log.Println("source:", oldn)
	log.Println("target:", newn)

	if err := rename(oldn, newn); err != nil {
		log.Println(err)
		return errors.Wrapf(err, "could not rename file on disk (%v) from (%v) to %v", source, target, f.node.name)
	}

	return nil
}

// Mkdir creats a directory
func (f *FFile) Mkdir(name string, mode os.FileMode) (*FFile, error) {
	log.Println("Mkdiring", name, "in", f.node.name)

	if err := mkdir(f.fs.realify(filepath.Join(f.node.Path(), name)), mode); err != nil {
		return nil, errors.Errorf("could not create real dir %v", name)
	}

	if err := f.node.CreateDirChild(name); err != nil {
		return nil, errors.Errorf("could not create dir %v to filetree", name)
	}

	child := f.node.Child(name)

	return NewFFile(child, f.fs), nil
}

func (f *FFile) Link(newName string, old *FFile) (*FFile, error) {
	oldnode := old.node
	log.Println("Linking", f.node.Name())

	if err := f.node.CreateChild(newName); err != nil {
		return nil, errors.Wrapf(err, "could not create link to file (%v)", newName)
	}

	child := f.node.Child(newName)
	if child == nil {
		return nil, errors.Errorf("could not find created link in file (%v)", newName)
	}

	if err := os.Link(f.fs.realify(oldnode.Path()), f.fs.realify(filepath.Join(f.node.Path(), newName))); err != nil {
		return nil, errors.Wrapf(err, "could not link file (%v) on disk", newName)
	}

	return NewFFile(child, f.fs), nil

}

func (f *FFile) Symlink(target, newName string) (*FFile, error) {
	log.Println("Symlinkig", f.node.Name())

	if err := os.Symlink(target, f.fs.realify(filepath.Join(f.node.Path(), newName))); err != nil {
		log.Println("symlink", err)
		return nil, errors.Wrapf(err, "could not symlink file (%v) with target (%v) on disk", newName, target)
	}

	if err := f.node.CreateLinkChild(newName, target); err != nil {
		log.Println("symlink create link child", err)
		return nil, errors.Wrapf(err, "could not symlink file (%v) with target (%v) on filetree", newName, target)
	}

	child := f.node.Child(newName)

	return NewFFile(child, f.fs), nil
}

func (f *FFile) Readlink() (string, error) {
	if f.node.Type() != LINK {
		return "", errors.New("file is not a symlink")
	}

	return f.node.Link(), nil
}

// Setattr to be implemented
func (f *FFile) Setattr(mode os.FileMode, atime, mtime time.Time) error {
	log.Println("Setattring", f.node.name)

	if err := os.Chmod(f.fs.realify(f.node.Path()), mode); err != nil {
		err = errors.Wrapf(err, "could not setattr chmod file")
		log.Println(err)
		return err
	}

	if err := os.Chtimes(f.fs.realify(f.node.Path()), atime, mtime); err != nil {
		err = errors.Wrapf(err, "could not setattr chtimes file")
		log.Println(err)
		return err
	}

	// NOTE: Changing owner not supported yet
	// TODO: Invistigate previlage escliation to apply chown

	return nil
}
