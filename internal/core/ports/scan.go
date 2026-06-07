package ports

// ScanFileSystem provides only the filesystem operations required by
// deterministic, non-recursive project scanning.
type ScanFileSystem interface {
	FileExists(root string, relativePath string) (bool, error)
	DirectoryExists(root string, relativePath string) (bool, error)
	ListDirectoryNames(root string, relativePath string) ([]string, error)
}
