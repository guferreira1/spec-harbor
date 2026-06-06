package filesystem

import (
	"errors"
	"io"
	"os"
	"path/filepath"
)

type LocalFileSystem struct{}

func NewLocalFileSystem() *LocalFileSystem {
	return &LocalFileSystem{}
}

func (fileSystem *LocalFileSystem) DirectoryExists(root string, relativePath string) (bool, error) {
	info, err := os.Stat(fileSystem.fullPath(root, relativePath))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	return info.IsDir(), nil
}

func (fileSystem *LocalFileSystem) FileExists(root string, relativePath string) (bool, error) {
	info, err := os.Stat(fileSystem.fullPath(root, relativePath))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	return !info.IsDir(), nil
}

func (fileSystem *LocalFileSystem) CreateDirectory(root string, relativePath string) error {
	return os.MkdirAll(fileSystem.fullPath(root, relativePath), 0o755)
}

func (fileSystem *LocalFileSystem) WriteFileIfAbsent(root string, relativePath string, contents string) (bool, error) {
	file, err := os.OpenFile(fileSystem.fullPath(root, relativePath), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return false, nil
		}
		return false, err
	}

	if _, err := io.WriteString(file, contents); err != nil {
		closeErr := file.Close()
		if closeErr != nil {
			return false, errors.Join(err, closeErr)
		}
		return false, err
	}

	if err := file.Close(); err != nil {
		return false, err
	}
	return true, nil
}

func (fileSystem *LocalFileSystem) fullPath(root string, relativePath string) string {
	return filepath.Join(root, filepath.FromSlash(relativePath))
}
