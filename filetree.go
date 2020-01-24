package resonatefuse

import (
	"log"
	"path/filepath"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/pkg/errors"
)

// FileTree is the building node of a filesystem
type FileTree struct {
	id       uint64
	name     string
	link     string
	parent   *FileTree
	children map[string]*FileTree
}

// NewNode constructs a new file tree node
func NewNode(name string, parent *FileTree) *FileTree {

	ft := &FileTree{
		name:     name,
		parent:   parent,
		id:       0,
		children: nil,
	}

	ft.id = fs.GenerateDynamicInode(ft.parentID(), name)

	return ft
}

func NewLink(name string, parent *FileTree, linked string) *FileTree {

	ft := NewNode(name, parent)
	ft.link = linked

	return ft
}

// NewDirectory constructs a new file tree with given parent
func NewDirectory(name string, parent *FileTree) *FileTree {

	ft := NewNode(name, parent)
	ft.children = make(map[string]*FileTree)

	return ft
}

// parentID gets parent Inode
func (ft *FileTree) parentID() uint64 {
	if ft.parent == nil {
		return 0
	}

	return ft.parent.ID()
}

// ID gets file's Inode
func (ft *FileTree) ID() uint64 {
	return ft.id
}

func (ft *FileTree) Link() string {
	return ft.link
}

// Name gets file's name
func (ft *FileTree) Name() string {
	return ft.name
}

// Type returns the type of filetree (folder, file, symlink, etc)
func (ft *FileTree) Type() fuse.DirentType {
	if ft.link != "" {
		return fuse.DT_Link
	}

	if ft.children == nil {
		return fuse.DT_File
	}

	return fuse.DT_Dir
}

// CreateChild adds a new child to the filetree
func (ft *FileTree) CreateChild(name string) error {
	return ft.AddChild(name, NewNode(filepath.Base(name), ft))
}

// CreateDirChild creates a new folder under the current folder
func (ft *FileTree) CreateDirChild(name string) error {
	return ft.AddChild(name, NewDirectory(filepath.Base(name), ft))
}

// CreateLinkChild creates a new symlink under the current folder
func (ft *FileTree) CreateLinkChild(name string, link string) error {
	return ft.AddChild(name, NewLink(filepath.Base(name), ft, link))
}

// AddChild adds an existing filetree as a child
func (ft *FileTree) AddChild(name string, child *FileTree) error {
	if ft.Type() == fuse.DT_File {
		return errors.New("cannot add child to leaf")
	}

	current := ft.Child(filepath.Dir(name))
	if current == nil {
		return errors.Errorf("path %v does not exist", filepath.Dir(name))
	}

	name = filepath.Base(name)
	// if _, ok := current.children[name]; ok {
	// 	return errors.Errorf("child %v already exists", name)
	// }

	current.children[name] = child
	child.parent = current

	return nil
}

// RemoveChild removes a child from chosen filetree
func (ft *FileTree) RemoveChild(name string) error {

	current := ft.Child(filepath.Dir(name))
	if current == nil {
		return errors.Errorf("path %v does not exist", filepath.Dir(name))
	}

	name = filepath.Base(name)
	if _, ok := current.children[name]; !ok {
		return errors.Errorf("child %v does not exist", name)
	}

	delete(current.children, name)
	return nil
}

// Rename changes the name of the current filetree
func (ft *FileTree) Rename(oldname string, newName string, newParent *FileTree) error {
	child := ft.Child(oldname)
	if child == nil {
		return errors.Errorf("could not rename none existant child (%v)", oldname)
	}

	if err := ft.RemoveChild(oldname); err != nil {
		return errors.Wrapf(err, "could not remove child (%v) while renaming", oldname)
	}

	child.name = newName

	if err := newParent.AddChild(newName, child); err != nil {
		return errors.Wrapf(err, "could not add child (%v) while renaming", newName)
	}

	return nil
}

// Children returns all children from chosen filetree
func (ft *FileTree) Children() []*FileTree {
	childrenList := make([]*FileTree, 0, len(ft.children))

	for _, child := range ft.children {
		childrenList = append(childrenList, child)
	}

	return childrenList
}

// Child returns a specific child from chosen directory
func (ft *FileTree) Child(name string) *FileTree {
	if ft.Type() == fuse.DT_File {
		return nil
	}

	current := ft
	for _, child := range splitPath(name) {
		current = current.children[child]
		if current == nil {
			return nil
		}
	}
	return current
}

// Dirents returns all children from chosen filetree
func (ft *FileTree) Dirents() []fuse.Dirent {
	childrenList := make([]fuse.Dirent, 0, len(ft.children))

	log.Println("traversing children")
	log.Println(ft.children)
	for _, child := range ft.children {
		childrenList = append(childrenList, fuse.Dirent{
			Inode: child.ID(),
			Name:  child.Name(),
			Type:  child.Type(),
		})
	}

	return childrenList
}

// parentID gets parent Inode
func (ft *FileTree) Root() *FileTree {
	current := ft

	for current.parent != nil {
		current = current.parent
	}

	return current
}

// Path returns the path of the filetree with respect to the root parent
func (ft *FileTree) Path() string {
	if ft.parent == nil {
		return "."
	}

	return filepath.Join(ft.parent.Path(), ft.name)
}

func (ft *FileTree) String() string {
	return ft.Path()
}
