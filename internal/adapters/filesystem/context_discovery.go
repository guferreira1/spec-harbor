package filesystem

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/guferreira1/spec-harbor/internal/core/ports"
)

type ContextDiscoveryFileSystem struct {
	local *LocalFileSystem
}

func NewContextDiscoveryFileSystem() *ContextDiscoveryFileSystem {
	return &ContextDiscoveryFileSystem{local: NewLocalFileSystem()}
}

func (fileSystem *ContextDiscoveryFileSystem) FileExists(root string, relativePath string) (bool, error) {
	pathInfo, err := fileSystem.safeExistingPath(root, relativePath)
	if err != nil {
		return false, err
	}
	if !pathInfo.exists || pathInfo.hasSymlink {
		return false, nil
	}
	return !pathInfo.info.IsDir(), nil
}

func (fileSystem *ContextDiscoveryFileSystem) DirectoryExists(root string, relativePath string) (bool, error) {
	pathInfo, err := fileSystem.safeExistingPath(root, relativePath)
	if err != nil {
		return false, err
	}
	if !pathInfo.exists || pathInfo.hasSymlink {
		return false, nil
	}
	return pathInfo.info.IsDir(), nil
}

func (fileSystem *ContextDiscoveryFileSystem) ListDirectory(
	root string,
	relativePath string,
) ([]ports.ContextDiscoveryDirectoryEntry, error) {
	pathInfo, err := fileSystem.safeExistingPath(root, relativePath)
	if err != nil {
		return nil, err
	}
	if !pathInfo.exists {
		return nil, fmt.Errorf("path does not exist: %s", relativePath)
	}
	if pathInfo.hasSymlink {
		return nil, unsafeSymlinkPathError(pathInfo.symlinkPath)
	}
	if !pathInfo.info.IsDir() {
		return nil, fmt.Errorf("path is not a directory: %s", relativePath)
	}

	entries, err := os.ReadDir(pathInfo.fullPath)
	if err != nil {
		return nil, err
	}

	result := make([]ports.ContextDiscoveryDirectoryEntry, 0, len(entries))
	for _, entry := range entries {
		isSymlink := entry.Type()&os.ModeSymlink != 0
		if isSymlink {
			result = append(result, ports.ContextDiscoveryDirectoryEntry{
				Name:      entry.Name(),
				IsSymlink: true,
			})
			continue
		}
		entryInfo, err := entry.Info()
		if err != nil {
			return nil, err
		}
		result = append(result, ports.ContextDiscoveryDirectoryEntry{
			Name:        entry.Name(),
			IsDirectory: entryInfo.IsDir(),
			IsRegular:   entryInfo.Mode().IsRegular(),
		})
	}
	return result, nil
}

func (fileSystem *ContextDiscoveryFileSystem) ReadFile(
	root string,
	relativePath string,
	maxBytes int64,
) (string, error) {
	if maxBytes <= 0 {
		return "", errors.New("context discovery read limit is required")
	}

	pathInfo, err := fileSystem.safeExistingPath(root, relativePath)
	if err != nil {
		return "", err
	}
	if !pathInfo.exists {
		return "", fmt.Errorf("path does not exist: %s", relativePath)
	}
	if pathInfo.hasSymlink {
		return "", unsafeSymlinkPathError(pathInfo.symlinkPath)
	}
	if pathInfo.info.IsDir() {
		return "", fmt.Errorf("path is a directory: %s", relativePath)
	}
	if pathInfo.info.Size() > maxBytes {
		return "", fmt.Errorf("context discovery file exceeds %d bytes: %s", maxBytes, relativePath)
	}

	contents, err := readFileWithoutFollowingSymlink(pathInfo.fullPath, pathInfo.info, relativePath)
	if err != nil {
		return "", err
	}
	return string(contents), nil
}

type contextDiscoveryPathInfo struct {
	fullPath    string
	info        os.FileInfo
	exists      bool
	hasSymlink  bool
	symlinkPath string
}

func (fileSystem *ContextDiscoveryFileSystem) safeExistingPath(
	root string,
	relativePath string,
) (contextDiscoveryPathInfo, error) {
	safeRelativePath, fullPath, err := fileSystem.local.safeFullPathParts(root, relativePath)
	if err != nil {
		return contextDiscoveryPathInfo{}, err
	}

	if safeRelativePath == "." {
		info, err := os.Lstat(root)
		return pathInfoFromLstat(fullPath, safeRelativePath, info, err)
	}

	currentPath := ""
	segments := strings.Split(safeRelativePath, "/")
	for index, segment := range segments {
		currentPath = path.Join(currentPath, segment)
		componentPath := filepath.Join(root, filepath.FromSlash(currentPath))
		info, err := os.Lstat(componentPath)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return contextDiscoveryPathInfo{fullPath: fullPath}, nil
			}
			return contextDiscoveryPathInfo{}, err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return contextDiscoveryPathInfo{
				fullPath:    fullPath,
				exists:      true,
				hasSymlink:  true,
				symlinkPath: currentPath,
			}, nil
		}
		if index < len(segments)-1 && !info.IsDir() {
			return contextDiscoveryPathInfo{}, fmt.Errorf("parent path is not a directory: %s", currentPath)
		}
		if index == len(segments)-1 {
			return contextDiscoveryPathInfo{
				fullPath: fullPath,
				info:     info,
				exists:   true,
			}, nil
		}
	}

	return contextDiscoveryPathInfo{fullPath: fullPath}, nil
}

func pathInfoFromLstat(
	fullPath string,
	relativePath string,
	info os.FileInfo,
	err error,
) (contextDiscoveryPathInfo, error) {
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return contextDiscoveryPathInfo{fullPath: fullPath}, nil
		}
		return contextDiscoveryPathInfo{}, err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return contextDiscoveryPathInfo{
			fullPath:    fullPath,
			exists:      true,
			hasSymlink:  true,
			symlinkPath: relativePath,
		}, nil
	}
	return contextDiscoveryPathInfo{
		fullPath: fullPath,
		info:     info,
		exists:   true,
	}, nil
}

func readFileWithoutFollowingSymlink(
	fullPath string,
	expectedInfo os.FileInfo,
	relativePath string,
) ([]byte, error) {
	file, err := os.Open(fullPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	openedInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}
	latestInfo, err := os.Lstat(fullPath)
	if err != nil {
		return nil, err
	}
	if latestInfo.Mode()&os.ModeSymlink != 0 {
		return nil, unsafeSymlinkPathError(relativePath)
	}
	if !os.SameFile(expectedInfo, latestInfo) || !os.SameFile(latestInfo, openedInfo) {
		return nil, fmt.Errorf("target file changed during safety check: %s", relativePath)
	}

	return io.ReadAll(file)
}
