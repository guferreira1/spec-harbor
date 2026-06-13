package filesystem

import (
	"errors"
	"fmt"

	"github.com/guferreira1/spec-harbor/internal/core/ports"
)

type RepositoryContextIndexFileSystem struct {
	local   *LocalFileSystem
	context *ContextDiscoveryFileSystem
}

func NewRepositoryContextIndexFileSystem() *RepositoryContextIndexFileSystem {
	local := NewLocalFileSystem()
	return &RepositoryContextIndexFileSystem{
		local:   local,
		context: &ContextDiscoveryFileSystem{local: local},
	}
}

func (fileSystem *RepositoryContextIndexFileSystem) FileExists(root string, relativePath string) (bool, error) {
	return fileSystem.context.FileExists(root, relativePath)
}

func (fileSystem *RepositoryContextIndexFileSystem) DirectoryExists(root string, relativePath string) (bool, error) {
	return fileSystem.context.DirectoryExists(root, relativePath)
}

func (fileSystem *RepositoryContextIndexFileSystem) ListDirectory(
	root string,
	relativePath string,
) ([]ports.RepositoryContextIndexDirectoryEntry, error) {
	entries, err := fileSystem.context.ListDirectory(root, relativePath)
	if err != nil {
		return nil, err
	}

	result := make([]ports.RepositoryContextIndexDirectoryEntry, 0, len(entries))
	for _, entry := range entries {
		result = append(result, ports.RepositoryContextIndexDirectoryEntry{
			Name:        entry.Name,
			IsDirectory: entry.IsDirectory,
			IsRegular:   entry.IsRegular,
			IsSymlink:   entry.IsSymlink,
		})
	}
	return result, nil
}

func (fileSystem *RepositoryContextIndexFileSystem) FileMetadata(
	root string,
	relativePath string,
) (ports.RepositoryContextIndexFileMetadata, error) {
	pathInfo, err := fileSystem.context.safeExistingPath(root, relativePath)
	if err != nil {
		return ports.RepositoryContextIndexFileMetadata{}, err
	}
	if !pathInfo.exists {
		return ports.RepositoryContextIndexFileMetadata{}, fmt.Errorf("path does not exist: %s", relativePath)
	}
	if pathInfo.hasSymlink {
		return ports.RepositoryContextIndexFileMetadata{}, unsafeSymlinkPathError(pathInfo.symlinkPath)
	}
	if pathInfo.info.IsDir() {
		return ports.RepositoryContextIndexFileMetadata{}, fmt.Errorf("path is a directory: %s", relativePath)
	}
	if !pathInfo.info.Mode().IsRegular() {
		return ports.RepositoryContextIndexFileMetadata{}, fmt.Errorf("path is not a regular file: %s", relativePath)
	}

	return ports.RepositoryContextIndexFileMetadata{
		SizeBytes:    pathInfo.info.Size(),
		ModifiedTime: pathInfo.info.ModTime().UTC(),
	}, nil
}

func (fileSystem *RepositoryContextIndexFileSystem) ReadFileBytes(
	root string,
	relativePath string,
	maxBytes int64,
) ([]byte, error) {
	if maxBytes <= 0 {
		return nil, errors.New("repository context index read limit is required")
	}
	pathInfo, err := fileSystem.context.safeExistingPath(root, relativePath)
	if err != nil {
		return nil, err
	}
	if !pathInfo.exists {
		return nil, fmt.Errorf("path does not exist: %s", relativePath)
	}
	if pathInfo.hasSymlink {
		return nil, unsafeSymlinkPathError(pathInfo.symlinkPath)
	}
	if pathInfo.info.IsDir() {
		return nil, fmt.Errorf("path is a directory: %s", relativePath)
	}
	if !pathInfo.info.Mode().IsRegular() {
		return nil, fmt.Errorf("path is not a regular file: %s", relativePath)
	}
	if pathInfo.info.Size() > maxBytes {
		return nil, fmt.Errorf("repository context index file exceeds %d bytes: %s", maxBytes, relativePath)
	}

	return readFileWithoutFollowingSymlink(pathInfo.fullPath, pathInfo.info, relativePath)
}

func (fileSystem *RepositoryContextIndexFileSystem) CreateDirectory(root string, relativePath string) error {
	return fileSystem.local.CreateDirectory(root, relativePath)
}

func (fileSystem *RepositoryContextIndexFileSystem) ReadFileSafely(root string, relativePath string) (string, error) {
	return fileSystem.local.ReadFileSafely(root, relativePath)
}

func (fileSystem *RepositoryContextIndexFileSystem) WriteFileSafely(
	root string,
	relativePath string,
	contents string,
) error {
	return fileSystem.local.WriteFileSafely(root, relativePath, contents)
}
