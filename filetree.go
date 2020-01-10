package resonatefs

import (
	"path/filepath"

	"github.com/pkg/errors"
)

// FileTree is the building node of a filesystem
type FileTree struct {
	name     string
	parent   *FileTree
	children map[string]*FileTree
}

// NewFileTree constructs a new file tree with root name
func NewFileTree(name string, parent *FileTree) *FileTree {
	return &FileTree{
		name:     name,
		parent:   parent,
		children: make(map[string]*FileTree),
	}
}

// AddChild adds a new child to chosen filetree
func (ft *FileTree) AddChild(name string) error {

	current := ft.Child(filepath.Dir(name))
	if current == nil {
		return errors.Errorf("path %v does not exist", filepath.Dir(name))
	}

	name = filepath.Base(name)
	if _, ok := current.children[name]; ok {
		return errors.Errorf("child %v already exists", name)
	}

	current.children[name] = NewFileTree(name, ft)

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

	delete(ft.children, name)
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

// Child returns a specific child from chosen filetree
func (ft *FileTree) Child(name string) *FileTree {
	current := ft
	for _, child := range splitPath(name) {
		current = current.children[child]
		if current == nil {
			return nil
		}
	}
	return current
}

// Path returns the path of the filetree with respect to the root parent
func (ft *FileTree) Path() string {
	if ft.parent == nil {
		return ft.name
	}

	return filepath.Join(ft.parent.Path(), ft.name)
}

func splitPath(path string) []string {
	path = filepath.Clean(path)

	if path == "." {
		return make([]string, 0)
	}

	dir, file := filepath.Split(path)

	return append(splitPath(dir), file)
}
