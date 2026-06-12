package ports

type ContextDiscoveryDirectoryEntry struct {
	Name        string
	IsDirectory bool
	IsRegular   bool
	IsSymlink   bool
}

// ContextDiscoveryFileSystem provides only the bounded read and listing
// operations required by deterministic local context discovery.
type ContextDiscoveryFileSystem interface {
	FileExists(root string, relativePath string) (bool, error)
	DirectoryExists(root string, relativePath string) (bool, error)
	ListDirectory(root string, relativePath string) ([]ContextDiscoveryDirectoryEntry, error)
	ReadFile(root string, relativePath string, maxBytes int64) (string, error)
}
