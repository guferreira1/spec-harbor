package ports

import "time"

type RepositoryContextIndexDirectoryEntry struct {
	Name        string
	IsDirectory bool
	IsRegular   bool
	IsSymlink   bool
}

type RepositoryContextIndexFileMetadata struct {
	SizeBytes    int64
	ModifiedTime time.Time
}

// RepositoryContextIndexFileSystem provides only the bounded filesystem
// operations required by deterministic local repository context indexing.
type RepositoryContextIndexFileSystem interface {
	FileExists(root string, relativePath string) (bool, error)
	DirectoryExists(root string, relativePath string) (bool, error)
	ListDirectory(root string, relativePath string) ([]RepositoryContextIndexDirectoryEntry, error)
	FileMetadata(root string, relativePath string) (RepositoryContextIndexFileMetadata, error)
	ReadFileBytes(root string, relativePath string, maxBytes int64) ([]byte, error)
	CreateDirectory(root string, relativePath string) error
	ReadFileSafely(root string, relativePath string) (string, error)
	WriteFileSafely(root string, relativePath string, contents string) error
}
