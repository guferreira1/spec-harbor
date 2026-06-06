package ports

// ValidationFileSystem provides only the filesystem operations required by
// structural OpenSpec validation.
type ValidationFileSystem interface {
	DirectoryExists(root string, relativePath string) (bool, error)
	FileExists(root string, relativePath string) (bool, error)
}
