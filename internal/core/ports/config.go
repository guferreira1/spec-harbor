package ports

import "github.com/guferreira1/spec-harbor/internal/core/domain"

// ConfigFileSystem provides only the filesystem operations required by
// read-only local SpecHarbor configuration display.
type ConfigFileSystem interface {
	DirectoryExists(root string, relativePath string) (bool, error)
	FileExists(root string, relativePath string) (bool, error)
	ReadFile(root string, relativePath string) (string, error)
}

// ConfigParser decodes local configuration contents into domain values.
type ConfigParser interface {
	ParseLocalConfig(contents string) (domain.LocalConfig, error)
}
