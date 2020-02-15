package resonatefuse

import (
	"os"
	"time"
)

type CreateHook func(*CreateRequest) error
type CreateRequest struct {
	Path string
	Name string
	Mode os.FileMode
}

type WriteHook func(*WriteRequest) error
type WriteRequest struct {
	Path   string
	Data   []byte
	Offset int64
}

type RemoveHook func(*RemoveRequest) error
type RemoveRequest struct {
	Path string
	Name string
}

type RenameHook func(*RenameRequest) error
type RenameRequest struct {
	Path    string
	OldName string
	NewName string
	NewDir  string
}

type MkdirHook func(*MkdirRequest) error
type MkdirRequest struct {
	Path string
	Name string
	Mode os.FileMode
}

type LinkHook func(*LinkRequest) error
type LinkRequest struct {
	Path    string
	NewName string
	Old     string
}

type SymlinkHook func(*SymlinkRequest) error
type SymlinkRequest struct {
	Path    string
	Target  string
	NewName string
}

type SetattrHook func(*SetattrRequest) error
type SetattrRequest struct {
	Path  string
	Mode  os.FileMode
	Atime time.Time
	Mtime time.Time
}
