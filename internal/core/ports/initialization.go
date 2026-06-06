package ports

// InitializationFileSystem provides only the filesystem operations required by
// project initialization.
type InitializationFileSystem interface {
	DirectoryExists(root string, relativePath string) (bool, error)
	FileExists(root string, relativePath string) (bool, error)
	CreateDirectory(root string, relativePath string) error
	WriteFileIfAbsent(root string, relativePath string, contents string) (bool, error)
}

// InitializationDefaults provides generated content for required initialization
// files.
type InitializationDefaults interface {
	ContentFor(relativePath string) (string, error)
}
