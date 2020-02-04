package resonatefuse

import (
	"os"
	"time"
)

type CreateHook func(*CreateRequest)
type CreateRequest struct {
	Path string
	Name string
	Mode os.FileMode
}

type WriteHook func(*WriteRequest)
type WriteRequest struct {
	Path   string
	Data   []byte
	Offset int64
}

type RemoveHook func(*RemoveRequest)
type RemoveRequest struct {
	Path string
	Name string
}

type RenameHook func(*RenameRequest)
type RenameRequest struct {
	Path    string
	OldName string
	NewName string
	NewDir  string
}

type MkdirHook func(*MkdirRequest)
type MkdirRequest struct {
	Path string
	Name string
	Mode os.FileMode
}

type LinkHook func(*LinkRequest)
type LinkRequest struct {
	Path    string
	NewName string
	Old     File
}

type SymlinkHook func(*SymlinkRequest)
type SymlinkRequest struct {
	Path    string
	Target  string
	NewName string
}

type SetattrHook func(*SetattrRequest)
type SetattrRequest struct {
	Path  string
	Mode  os.FileMode
	Atime time.Time
	Mtime time.Time
}
