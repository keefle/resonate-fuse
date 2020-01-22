package resonatefuse

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateChild(t *testing.T) {
	childName := "joe"
	root := NewFileTree("root", nil)

	// Add child that does not prevously exist
	assert.Nil(t, root.children[childName])
	assert.Nil(t, root.CreateChild(childName))
	assert.NotNil(t, root.children[childName])
	assert.Equal(t, root, root.children[childName].parent)

	// Add child that does already exist
	assert.NotNil(t, root.children[childName])
	assert.Nil(t, root.CreateChild(childName))
	assert.NotNil(t, root.children[childName])
	assert.Equal(t, root, root.children[childName].parent)
}

func TestCreateDirChild(t *testing.T) {
	childName := "joe"
	root := NewFileTree("root", nil)

	// Add child that does not prevously exist
	assert.Nil(t, root.children[childName])
	assert.Nil(t, root.CreateDirChild(childName))
	assert.NotNil(t, root.children[childName])
	assert.Equal(t, root, root.children[childName].parent)

	// Add child that does already exist
	assert.NotNil(t, root.children[childName])
	assert.Nil(t, root.CreateDirChild(childName))
	assert.NotNil(t, root.children[childName])
	assert.Equal(t, root, root.children[childName].parent)

	// Add child via path syntax (unix: joe/ali)
	childMultiName := []string{childName, "ali"}

	assert.Nil(t, root.CreateDirChild(filepath.Join(childMultiName...)))
	assert.NotNil(t, root.children[childMultiName[0]].children[childMultiName[1]])
	assert.Equal(t, root.children[childMultiName[0]], root.children[childMultiName[0]].children[childMultiName[1]].parent)
}

func TestRemoveChild(t *testing.T) {

	tt := []struct {
		state  []string
		target string
	}{
		{state: []string{"leo"}, target: "leo"},
		{state: []string{"muhammad", "muhammad/ali"}, target: "muhammad/ali"},
	}

	for _, tc := range tt {
		root := NewFileTree("root", nil)

		for _, file := range tc.state {
			assert.Nil(t, root.CreateDirChild(file))
		}

		// remove a child that exists
		assert.Nil(t, root.RemoveChild(tc.target))
		assert.Nil(t, root.Child(tc.target))
	}
}

func TestRename(t *testing.T) {

	tt := []struct {
		state  []string
		source string
		target string
	}{
		{state: []string{"joe"}, source: "joe", target: "leo"},
		{state: []string{"muhammad", "muhammad/ali"}, source: "muhammad/ali", target: "ali"},
	}

	for _, tc := range tt {
		root := NewFileTree("root", nil)

		for _, file := range tc.state {
			assert.Nil(t, root.CreateDirChild(file))
		}

		assert.Nil(t, root.Rename(tc.source, tc.target, root))
		assert.Nil(t, root.Child(tc.source))
		assert.NotNil(t, root.Child(tc.target))
	}
}

func TestChild(t *testing.T) {
	childrenNames := []string{"jeo", "ali", "leo"}

	root := NewFileTree("root", nil)

	for _, childName := range childrenNames {
		root.children[childName] = NewFileTree(childName, root)
	}

	for _, name := range childrenNames {
		assert.Equal(t, root.Child(name), root.children[name])
	}
}

func TestChildren(t *testing.T) {
	root := NewFileTree("root", nil)

	for _, childName := range []string{"jeo", "ali", "leo"} {
		root.children[childName] = NewFileTree(childName, root)
	}

	childrenList := make([]*FileTree, 0, len(root.children))

	for _, child := range root.children {
		childrenList = append(childrenList, child)
	}

	assert.ElementsMatch(t, root.Children(), childrenList)
}

func TestPath(t *testing.T) {
	root := NewFileTree("root", nil)

	childName := "joe"
	child := NewFileTree(childName, root)

	assert.Equal(t, root.Path(), filepath.Join("."))
	assert.Equal(t, child.Path(), filepath.Join(".", childName))
}
