package ports

// ValidationFileSystem provides only the read-only filesystem operations
// required by OpenSpec validation. It exposes no write operations.
type ValidationFileSystem interface {
	DirectoryExists(root string, relativePath string) (bool, error)
	FileExists(root string, relativePath string) (bool, error)
	ReadFile(root string, relativePath string) (string, error)
}
