package ports

// ReviewFileSystem provides only the filesystem operations required by
// deterministic OpenSpec implementation review.
type ReviewFileSystem interface {
	DirectoryExists(root string, relativePath string) (bool, error)
	FileExists(root string, relativePath string) (bool, error)
	ReadFile(root string, relativePath string) (string, error)
}
