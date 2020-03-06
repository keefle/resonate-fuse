package resonatefuse

type Option func(*FS)

func GeneralOption(operation HookType, h GeneralHook) Option {
	return func(rfs *FS) {
		rfs.hooks[operation] = h
	}
}
