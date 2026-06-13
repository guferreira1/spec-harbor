package filesystem

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type LocalFileSystem struct{}

func NewLocalFileSystem() *LocalFileSystem {
	return &LocalFileSystem{}
}

func (fileSystem *LocalFileSystem) DirectoryExists(root string, relativePath string) (bool, error) {
	fullPath, err := fileSystem.safeFullPath(root, relativePath)
	if err != nil {
		return false, err
	}

	info, err := os.Lstat(fullPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return false, unsafeSymlinkPathError(relativePath)
	}
	return info.IsDir(), nil
}

func (fileSystem *LocalFileSystem) FileExists(root string, relativePath string) (bool, error) {
	fullPath, err := fileSystem.safeFullPath(root, relativePath)
	if err != nil {
		return false, err
	}

	info, err := os.Lstat(fullPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return false, unsafeSymlinkPathError(relativePath)
	}
	return !info.IsDir(), nil
}

func (fileSystem *LocalFileSystem) PathExists(root string, relativePath string) (bool, error) {
	fullPath, err := fileSystem.safeFullPath(root, relativePath)
	if err != nil {
		return false, err
	}

	info, err := os.Lstat(fullPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return false, unsafeSymlinkPathError(relativePath)
	}
	return true, nil
}

func (fileSystem *LocalFileSystem) ListDirectoryNames(root string, relativePath string) ([]string, error) {
	fullPath, err := fileSystem.safeFullPath(root, relativePath)
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	return names, nil
}

func (fileSystem *LocalFileSystem) ReadFile(root string, relativePath string) (string, error) {
	fullPath, err := fileSystem.safeFullPath(root, relativePath)
	if err != nil {
		return "", err
	}

	contents, err := os.ReadFile(fullPath)
	if err != nil {
		return "", err
	}
	return string(contents), nil
}

func (fileSystem *LocalFileSystem) ReadFileSafely(root string, relativePath string) (string, error) {
	fullPath, info, err := fileSystem.safeExistingReadTarget(root, relativePath)
	if err != nil {
		return "", err
	}
	if info.IsDir() {
		return "", fmt.Errorf("path is a directory: %s", relativePath)
	}

	contents, err := readFileWithoutFollowingSymlink(fullPath, info, relativePath)
	if err != nil {
		return "", err
	}
	return string(contents), nil
}

func (fileSystem *LocalFileSystem) ReadSourceFile(sourcePath string) (string, error) {
	if strings.TrimSpace(sourcePath) == "" {
		return "", errors.New("source file path is required")
	}

	info, err := os.Stat(sourcePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("source file not found: %s", sourcePath)
		}
		return "", err
	}
	if info.IsDir() {
		return "", fmt.Errorf("source file is a directory: %s", sourcePath)
	}

	contents, err := os.ReadFile(sourcePath)
	if err != nil {
		return "", err
	}
	return string(contents), nil
}

func (fileSystem *LocalFileSystem) CreateDirectory(root string, relativePath string) error {
	fullPath, err := fileSystem.safeFullPath(root, relativePath)
	if err != nil {
		return err
	}
	if err := fileSystem.ensureSafeDirectoryTarget(root, relativePath); err != nil {
		return err
	}
	return os.MkdirAll(fullPath, 0o755)
}

func (fileSystem *LocalFileSystem) MoveDirectory(root string, sourceRelativePath string, destinationRelativePath string) error {
	sourcePath, err := fileSystem.safeFullPath(root, sourceRelativePath)
	if err != nil {
		return err
	}
	destinationPath, err := fileSystem.safeFullPath(root, destinationRelativePath)
	if err != nil {
		return err
	}

	sourceInfo, err := os.Stat(sourcePath)
	if err != nil {
		return err
	}
	if !sourceInfo.IsDir() {
		return fmt.Errorf("source path is not a directory: %s", sourceRelativePath)
	}

	if _, err := os.Stat(destinationPath); err == nil {
		return fmt.Errorf("destination path already exists: %s", destinationRelativePath)
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	return os.Rename(sourcePath, destinationPath)
}

func (fileSystem *LocalFileSystem) WriteFileIfAbsent(root string, relativePath string, contents string) (bool, error) {
	fullPath, err := fileSystem.safeFullPath(root, relativePath)
	if err != nil {
		return false, err
	}
	if err := fileSystem.EnsureSafeWriteTarget(root, relativePath); err != nil {
		return false, err
	}

	file, err := os.OpenFile(fullPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return false, nil
		}
		return false, err
	}

	if err := writeStringAndClose(file, contents); err != nil {
		return false, err
	}
	return true, nil
}

func (fileSystem *LocalFileSystem) WriteFile(root string, relativePath string, contents string) error {
	fullPath, err := fileSystem.safeFullPath(root, relativePath)
	if err != nil {
		return err
	}
	if err := fileSystem.EnsureSafeWriteTarget(root, relativePath); err != nil {
		return err
	}
	return replaceFileWithoutFollowingSymlink(fullPath, relativePath, contents)
}

func (fileSystem *LocalFileSystem) WriteFileSafely(root string, relativePath string, contents string) error {
	fullPath, err := fileSystem.safeFullPath(root, relativePath)
	if err != nil {
		return err
	}
	if err := fileSystem.EnsureSafeWriteTarget(root, relativePath); err != nil {
		return err
	}
	return replaceFileAtomicallyWithoutFollowingSymlink(fullPath, relativePath, contents)
}

func (fileSystem *LocalFileSystem) EnsureSafeWriteTarget(root string, relativePath string) error {
	safeRelativePath, fullPath, err := fileSystem.safeFullPathParts(root, relativePath)
	if err != nil {
		return err
	}
	if err := fileSystem.ensureSafeExistingParents(root, safeRelativePath); err != nil {
		return err
	}

	info, err := os.Lstat(fullPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return unsafeGeneratedSymlinkTargetError(relativePath)
	}
	return nil
}

func (fileSystem *LocalFileSystem) ensureSafeDirectoryTarget(root string, relativePath string) error {
	safeRelativePath, fullPath, err := fileSystem.safeFullPathParts(root, relativePath)
	if err != nil {
		return err
	}
	if err := fileSystem.ensureSafeExistingParents(root, safeRelativePath); err != nil {
		return err
	}

	info, err := os.Lstat(fullPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return unsafeSymlinkPathError(relativePath)
	}
	return nil
}

func (fileSystem *LocalFileSystem) ensureSafeExistingParents(root string, safeRelativePath string) error {
	parentPath := path.Dir(safeRelativePath)
	if parentPath == "." {
		return nil
	}

	currentPath := ""
	for _, segment := range strings.Split(parentPath, "/") {
		if currentPath == "" {
			currentPath = segment
		} else {
			currentPath += "/" + segment
		}

		fullPath := filepath.Join(root, filepath.FromSlash(currentPath))
		info, err := os.Lstat(fullPath)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return nil
			}
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return unsafeGeneratedSymlinkParentError(currentPath)
		}
		if !info.IsDir() {
			return fmt.Errorf("parent path is not a directory: %s", currentPath)
		}
	}

	return nil
}

func (fileSystem *LocalFileSystem) safeExistingReadTarget(
	root string,
	relativePath string,
) (string, os.FileInfo, error) {
	safeRelativePath, fullPath, err := fileSystem.safeFullPathParts(root, relativePath)
	if err != nil {
		return "", nil, err
	}

	if safeRelativePath == "." {
		info, err := os.Lstat(fullPath)
		if err != nil {
			return "", nil, err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return "", nil, unsafeSymlinkPathError(safeRelativePath)
		}
		return fullPath, info, nil
	}

	currentPath := ""
	segments := strings.Split(safeRelativePath, "/")
	for index, segment := range segments {
		currentPath = path.Join(currentPath, segment)
		componentPath := filepath.Join(root, filepath.FromSlash(currentPath))
		info, err := os.Lstat(componentPath)
		if err != nil {
			return "", nil, err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return "", nil, unsafeSymlinkPathError(currentPath)
		}
		if index < len(segments)-1 && !info.IsDir() {
			return "", nil, fmt.Errorf("parent path is not a directory: %s", currentPath)
		}
		if index == len(segments)-1 {
			return fullPath, info, nil
		}
	}

	return "", nil, fmt.Errorf("path does not exist: %s", relativePath)
}

func (fileSystem *LocalFileSystem) safeFullPath(root string, relativePath string) (string, error) {
	_, fullPath, err := fileSystem.safeFullPathParts(root, relativePath)
	return fullPath, err
}

func (fileSystem *LocalFileSystem) safeFullPathParts(root string, relativePath string) (string, string, error) {
	safeRelativePath, err := safeRelativePath(relativePath)
	if err != nil {
		return "", "", err
	}
	if safeRelativePath == "." {
		return safeRelativePath, root, nil
	}
	return safeRelativePath, filepath.Join(root, filepath.FromSlash(safeRelativePath)), nil
}

func (fileSystem *LocalFileSystem) fullPath(root string, relativePath string) string {
	return filepath.Join(root, filepath.FromSlash(relativePath))
}

func safeRelativePath(relativePath string) (string, error) {
	if strings.TrimSpace(relativePath) == "" {
		return "", errors.New("relative path is required")
	}
	if strings.ContainsRune(relativePath, 0) {
		return "", errors.New("relative path contains a null byte")
	}

	normalized := strings.ReplaceAll(relativePath, "\\", "/")
	if filepath.IsAbs(relativePath) || strings.HasPrefix(normalized, "/") || isWindowsDrivePath(normalized) {
		return "", fmt.Errorf("relative path must not be absolute: %s", relativePath)
	}
	for _, segment := range strings.Split(normalized, "/") {
		if segment == ".." {
			return "", fmt.Errorf("relative path must not contain path traversal: %s", relativePath)
		}
	}

	cleaned := path.Clean(normalized)
	if cleaned == ".." || strings.HasPrefix(cleaned, "../") {
		return "", fmt.Errorf("relative path must not contain path traversal: %s", relativePath)
	}
	return cleaned, nil
}

func isWindowsDrivePath(value string) bool {
	return len(value) >= 2 && isASCIIAlpha(value[0]) && value[1] == ':'
}

func isASCIIAlpha(value byte) bool {
	return (value >= 'a' && value <= 'z') || (value >= 'A' && value <= 'Z')
}

func unsafeGeneratedSymlinkTargetError(relativePath string) error {
	return fmt.Errorf("symlink target paths are not allowed for generated OpenSpec files: %s", relativePath)
}

func unsafeGeneratedSymlinkParentError(relativePath string) error {
	return fmt.Errorf("symlink parent directories are not allowed for generated OpenSpec files: %s", relativePath)
}

func unsafeSymlinkPathError(relativePath string) error {
	return fmt.Errorf("symlink paths are not allowed: %s", relativePath)
}

func replaceFileWithoutFollowingSymlink(fullPath string, relativePath string, contents string) error {
	initialInfo, err := os.Lstat(fullPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			file, err := os.OpenFile(fullPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
			if err != nil {
				return err
			}
			return writeStringAndClose(file, contents)
		}
		return err
	}
	if initialInfo.Mode()&os.ModeSymlink != 0 {
		return unsafeGeneratedSymlinkTargetError(relativePath)
	}

	file, err := os.OpenFile(fullPath, os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}

	openedInfo, err := file.Stat()
	if err != nil {
		return closeWithError(file, err)
	}
	latestInfo, err := os.Lstat(fullPath)
	if err != nil {
		return closeWithError(file, err)
	}
	if latestInfo.Mode()&os.ModeSymlink != 0 {
		return closeWithError(file, unsafeGeneratedSymlinkTargetError(relativePath))
	}
	if !os.SameFile(latestInfo, openedInfo) {
		return closeWithError(file, fmt.Errorf("target file changed during safety check: %s", relativePath))
	}
	if err := file.Truncate(0); err != nil {
		return closeWithError(file, err)
	}
	if _, err := file.Seek(0, 0); err != nil {
		return closeWithError(file, err)
	}

	return writeStringAndClose(file, contents)
}

func writeStringAndClose(file *os.File, contents string) error {
	if _, err := io.WriteString(file, contents); err != nil {
		return closeWithError(file, err)
	}
	return file.Close()
}

func closeWithError(file *os.File, err error) error {
	if closeErr := file.Close(); closeErr != nil {
		return errors.Join(err, closeErr)
	}
	return err
}

func replaceFileAtomicallyWithoutFollowingSymlink(fullPath string, relativePath string, contents string) error {
	initialInfo, err := os.Lstat(fullPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err == nil && initialInfo.Mode()&os.ModeSymlink != 0 {
		return unsafeGeneratedSymlinkTargetError(relativePath)
	}

	parent := filepath.Dir(fullPath)
	temp, err := os.CreateTemp(parent, "."+filepath.Base(fullPath)+".tmp-*")
	if err != nil {
		return err
	}
	tempPath := temp.Name()
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.Remove(tempPath)
		}
	}()

	if err := temp.Chmod(0o644); err != nil {
		return closeWithError(temp, err)
	}
	if _, err := io.WriteString(temp, contents); err != nil {
		return closeWithError(temp, err)
	}
	if err := temp.Close(); err != nil {
		return err
	}

	latestInfo, err := os.Lstat(fullPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err == nil && latestInfo.Mode()&os.ModeSymlink != 0 {
		return unsafeGeneratedSymlinkTargetError(relativePath)
	}
	if initialInfo != nil && latestInfo != nil && !os.SameFile(initialInfo, latestInfo) {
		return fmt.Errorf("target file changed during safety check: %s", relativePath)
	}

	if err := os.Rename(tempPath, fullPath); err != nil {
		return err
	}
	cleanup = false
	return nil
}
