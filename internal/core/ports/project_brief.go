package ports

// ProjectBriefFileSystem provides only the filesystem operations required by
// project brief creation.
type ProjectBriefFileSystem interface {
	DirectoryExists(root string, relativePath string) (bool, error)
	FileExists(root string, relativePath string) (bool, error)
	CreateDirectory(root string, relativePath string) error
	WriteFileIfAbsent(root string, relativePath string, contents string) (bool, error)
}
