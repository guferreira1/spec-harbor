package ports

// ArchiveFileSystem provides only the filesystem operations required by
// OpenSpec change archiving.
type ArchiveFileSystem interface {
	DirectoryExists(root string, relativePath string) (bool, error)
	FileExists(root string, relativePath string) (bool, error)
	PathExists(root string, relativePath string) (bool, error)
	CreateDirectory(root string, relativePath string) error
	MoveDirectory(root string, sourceRelativePath string, destinationRelativePath string) error
}
