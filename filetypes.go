package resonatefuse

import "bazil.org/fuse"

type NodeType uint

const (
	FILE NodeType = iota
	DIR
	LINK
)

func (nt NodeType) ToFUSE() fuse.DirentType {
	switch nt {
	case FILE:
		return fuse.DT_File
	case LINK:
		return fuse.DT_Link
	case DIR:
		return fuse.DT_Dir
	default:
		return fuse.DT_Unknown
	}
}
