package resonatefuse

import (
	"os"
	"time"
)

type HookType uint16

const (
	CreateType HookType = iota + 1
	WriteType
	RemoveType
	RenameType
	MkdirType
	LinkType
	SymlinkType
	SetattrType
)

type GeneralHook func(*GeneralRequest) error
type GeneralRequest struct {
	Atime   time.Time
	Data    []byte
	Mode    os.FileMode
	Mtime   time.Time
	Name    string
	NewDir  string
	NewName string
	Offset  int64
	OldName string
	Old     string
	Path    string
	Target  string
}
