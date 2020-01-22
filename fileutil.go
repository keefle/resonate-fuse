package resonatefuse

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

func splitPath(path string) []string {
	path = filepath.Clean(path)

	if path == "." {
		return make([]string, 0)
	}

	dir, file := filepath.Split(path)

	return append(splitPath(dir), file)
}

func touch(name string) error {
	file, err := os.OpenFile(name, os.O_RDONLY|os.O_CREATE, 0664)
	if err != nil {
		return err
	}
	return file.Close()
}

func rm(name string) error {
	return os.Remove(name)
}

func mkdir(name string) error {
	return os.Mkdir(name, os.ModeDir|0775)
}

func realify(path string) string {
	return filepath.Join("real", path)
}

func readall(path string) ([]byte, error) {
	return ioutil.ReadFile(path)
}
