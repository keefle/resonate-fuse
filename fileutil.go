package resonatefuse

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

func splitPath(path string) []string {
	path = filepath.Clean(path)

	if path == "." {
		return make([]string, 0)
	}

	dir, file := filepath.Split(path)

	return append(splitPath(dir), file)
}

func writeAt(name string, data []byte, offset int64) (int, error) {
	file, err := os.OpenFile(name, os.O_RDWR, 0664)
	if err != nil {
		return 0, errors.Errorf("could not open file %v: %v", name, err)
	}
	defer file.Close()

	n, err := file.WriteAt(data, offset)
	if err != nil {
		return n, errors.Errorf("could not write to file %v: %v", name, err)
	}

	return n, nil
}

func readAt(name string, data []byte, offset int64) (int, error) {
	file, err := os.OpenFile(name, os.O_RDONLY, 0664)
	if err != nil {
		return 0, errors.Errorf("could not open file %v: %v", name, err)
	}
	defer file.Close()

	n, err := file.ReadAt(data, offset)
	if err != nil {
		return n, errors.Errorf("could not write to file %v: %v", name, err)
	}

	return n, nil
}

func touch(name string, mode os.FileMode) error {
	file, err := os.OpenFile(name, os.O_RDONLY|os.O_CREATE, mode)
	if err != nil {
		return err
	}
	return file.Close()
}

func rm(name string) error {
	return os.Remove(name)
}

func mkdir(name string, mode os.FileMode) error {
	return os.Mkdir(name, mode)
}

func realify(path string) string {
	return filepath.Join("real", path)
}

func readall(path string) ([]byte, error) {
	return ioutil.ReadFile(path)
}
