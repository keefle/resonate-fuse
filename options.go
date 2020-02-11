package resonatefuse

type Option func(*FS)

func CreateOption(h CreateHook) Option {
	return func(rfs *FS) {
		rfs.createHook = h
	}
}

func WriteOption(h WriteHook) Option {
	return func(rfs *FS) {
		rfs.writeHook = h
	}
}

func RemoveOption(h RemoveHook) Option {
	return func(rfs *FS) {
		rfs.removeHook = h
	}
}

func RenameOption(h RenameHook) Option {
	return func(rfs *FS) {
		rfs.renameHook = h
	}
}

func MkdirOption(h MkdirHook) Option {
	return func(rfs *FS) {
		rfs.mkdirHook = h
	}
}

func LinkOption(h LinkHook) Option {
	return func(rfs *FS) {
		rfs.linkHook = h
	}
}

func SymlinkOption(h SymlinkHook) Option {
	return func(rfs *FS) {
		rfs.symlinkHook = h
	}
}

func SetattrOption(h SetattrHook) Option {
	return func(rfs *FS) {
		rfs.setattrHook = h
	}
}
