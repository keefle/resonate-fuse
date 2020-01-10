package resonatefs

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddChild(t *testing.T) {
	childName := "joe"
	root := NewFileTree("root", nil)

	// Add child that does not prevously exist
	assert.Nil(t, root.children[childName])
	assert.Nil(t, root.AddChild(childName))
	assert.NotNil(t, root.children[childName])

	// Add child that does already exist
	assert.NotNil(t, root.children[childName])
	assert.Error(t, root.AddChild(childName))
	assert.NotNil(t, root.children[childName])

	// Add child via path syntax (unix: joe/ali)
	childMultiName := []string{childName, "ali"}

	assert.Nil(t, root.AddChild(filepath.Join(childMultiName...)))
	assert.NotNil(t, root.children[childMultiName[0]])
	assert.NotNil(t, root.children[childMultiName[0]].children[childMultiName[1]])

}

func TestRemoveChild(t *testing.T) {
	childName := "jeo"
	root := NewFileTree("root", nil)
	root.children[childName] = NewFileTree(childName, root)

	// remove a child that exists
	assert.Nil(t, root.RemoveChild(childName))
	assert.Nil(t, root.children[childName])

	noneExistantChildName := "leo"
	// remove a child that does not exist (should return error)
	assert.NotNil(t, root.RemoveChild(noneExistantChildName))
	assert.Nil(t, root.children[noneExistantChildName])
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

	assert.Equal(t, root.Path(), filepath.Join("root"))
	assert.Equal(t, child.Path(), filepath.Join("root", childName))
}

func TestSplitPath(t *testing.T) {
	path := "to/joe/man"
	assert.Equal(t, splitPath(path), []string{"to", "joe", "man"})
}
