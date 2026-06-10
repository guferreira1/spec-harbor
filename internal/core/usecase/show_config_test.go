package usecase

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/guferreira1/spec-harbor/internal/core/domain"
)

func TestShowConfigReturnsVersionOneConfig(t *testing.T) {
	fileSystem := newFakeConfigFileSystem()
	seedConfigProject(fileSystem, "config contents")
	parser := &fakeConfigParser{config: completeDomainConfig()}

	result, err := NewShowConfig(fileSystem, parser).Execute(ShowConfigInput{ProjectRoot: "/project"})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if result.Path != localConfigPath {
		t.Fatalf("Path = %q, want %q", result.Path, localConfigPath)
	}
	if !reflect.DeepEqual(result.Config, completeDomainConfig()) {
		t.Fatalf("Config = %#v, want complete config", result.Config)
	}
	if fileSystem.readCount != 1 {
		t.Fatalf("ReadFile calls = %d, want 1", fileSystem.readCount)
	}
	if parser.calls != 1 {
		t.Fatalf("ParseLocalConfig calls = %d, want 1", parser.calls)
	}
	if parser.contents != "config contents" {
		t.Fatalf("parser contents = %q, want config contents", parser.contents)
	}
}

func TestShowConfigReturnsRelativeConfigPath(t *testing.T) {
	fileSystem := newFakeConfigFileSystem()
	seedConfigProject(fileSystem, "config contents")
	parser := &fakeConfigParser{config: domain.LocalConfig{Version: domain.SupportedLocalConfigVersion}}

	result, err := NewShowConfig(fileSystem, parser).Execute(ShowConfigInput{ProjectRoot: "/project"})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if result.Path != ".specharbor/config.yml" {
		t.Fatalf("Path = %q, want .specharbor/config.yml", result.Path)
	}
}

func TestShowConfigRejectsMissingDependencies(t *testing.T) {
	_, err := (*ShowConfig)(nil).Execute(ShowConfigInput{})
	if err == nil || !strings.Contains(err.Error(), "show config use case is required") {
		t.Fatalf("nil use case error = %v, want show config use case is required", err)
	}

	_, err = NewShowConfig(nil, &fakeConfigParser{}).Execute(ShowConfigInput{})
	if err == nil || !strings.Contains(err.Error(), "config filesystem is required") {
		t.Fatalf("nil filesystem error = %v, want config filesystem is required", err)
	}

	_, err = NewShowConfig(newFakeConfigFileSystem(), nil).Execute(ShowConfigInput{})
	if err == nil || !strings.Contains(err.Error(), "config parser is required") {
		t.Fatalf("nil parser error = %v, want config parser is required", err)
	}
}

func TestShowConfigRejectsEmptyProjectRootBeforeFilesystemOperations(t *testing.T) {
	fileSystem := newFakeConfigFileSystem()
	parser := &fakeConfigParser{}

	_, err := NewShowConfig(fileSystem, parser).Execute(ShowConfigInput{ProjectRoot: "  "})
	if err == nil || !strings.Contains(err.Error(), "project root is required") {
		t.Fatalf("Execute() error = %v, want project root is required", err)
	}
	if fileSystem.operationCount() != 0 {
		t.Fatalf("filesystem operations = %d, want 0", fileSystem.operationCount())
	}
	if parser.calls != 0 {
		t.Fatalf("parser calls = %d, want 0", parser.calls)
	}
}

func TestShowConfigReturnsExecutionErrorForUnavailableProjectRoot(t *testing.T) {
	tests := []struct {
		name  string
		setup func(fileSystem *fakeConfigFileSystem)
		want  string
	}{
		{
			name:  "root unavailable",
			setup: func(_ *fakeConfigFileSystem) {},
			want:  "project root is unavailable or not a directory",
		},
		{
			name: "root check error",
			setup: func(fileSystem *fakeConfigFileSystem) {
				fileSystem.directoryErrors["."] = errors.New("permission denied")
			},
			want: "check project root",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fileSystem := newFakeConfigFileSystem()
			test.setup(fileSystem)
			parser := &fakeConfigParser{}

			_, err := NewShowConfig(fileSystem, parser).Execute(ShowConfigInput{ProjectRoot: "/project"})
			if err == nil {
				t.Fatalf("Execute() error = nil, want %q", test.want)
			}
			if !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Execute() error = %q, want %q", err.Error(), test.want)
			}
			if fileSystem.readCount != 0 {
				t.Fatalf("ReadFile calls = %d, want 0", fileSystem.readCount)
			}
			if parser.calls != 0 {
				t.Fatalf("parser calls = %d, want 0", parser.calls)
			}
		})
	}
}

func TestShowConfigReturnsMissingConfigErrorWhenConfigIsMissingOrNotAFile(t *testing.T) {
	tests := []struct {
		name  string
		setup func(fileSystem *fakeConfigFileSystem)
	}{
		{
			name:  "missing config",
			setup: func(_ *fakeConfigFileSystem) {},
		},
		{
			name: "config path is directory",
			setup: func(fileSystem *fakeConfigFileSystem) {
				fileSystem.directories[localConfigPath] = true
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fileSystem := newFakeConfigFileSystem()
			fileSystem.directories["."] = true
			test.setup(fileSystem)
			parser := &fakeConfigParser{}

			_, err := NewShowConfig(fileSystem, parser).Execute(ShowConfigInput{ProjectRoot: "/project"})
			if err == nil {
				t.Fatalf("Execute() error = nil, want missing config error")
			}
			for _, want := range []string{"missing config file", localConfigPath} {
				if !strings.Contains(err.Error(), want) {
					t.Fatalf("Execute() error = %q, want to contain %q", err.Error(), want)
				}
			}
			if fileSystem.readCount != 0 {
				t.Fatalf("ReadFile calls = %d, want 0", fileSystem.readCount)
			}
			if parser.calls != 0 {
				t.Fatalf("parser calls = %d, want 0", parser.calls)
			}
		})
	}
}

func TestShowConfigReturnsUnreadableConfigError(t *testing.T) {
	wantErr := errors.New("permission denied")
	fileSystem := newFakeConfigFileSystem()
	seedConfigProject(fileSystem, "config contents")
	fileSystem.readErrors[localConfigPath] = wantErr
	parser := &fakeConfigParser{}

	_, err := NewShowConfig(fileSystem, parser).Execute(ShowConfigInput{ProjectRoot: "/project"})
	if err == nil {
		t.Fatalf("Execute() error = nil, want unreadable config error")
	}
	if !errors.Is(err, wantErr) {
		t.Fatalf("Execute() error = %v, want wrapping %v", err, wantErr)
	}
	if !strings.Contains(err.Error(), "unreadable config") {
		t.Fatalf("Execute() error = %q, want unreadable config", err.Error())
	}
	if parser.calls != 0 {
		t.Fatalf("parser calls = %d, want 0", parser.calls)
	}
}

func TestShowConfigReturnsInvalidYAMLErrorForParserFailure(t *testing.T) {
	wantErr := errors.New("decode failed")
	fileSystem := newFakeConfigFileSystem()
	seedConfigProject(fileSystem, "invalid yaml")
	parser := &fakeConfigParser{err: wantErr}

	_, err := NewShowConfig(fileSystem, parser).Execute(ShowConfigInput{ProjectRoot: "/project"})
	if err == nil {
		t.Fatalf("Execute() error = nil, want invalid YAML error")
	}
	if !errors.Is(err, wantErr) {
		t.Fatalf("Execute() error = %v, want wrapping %v", err, wantErr)
	}
	if !strings.Contains(err.Error(), "invalid config YAML") {
		t.Fatalf("Execute() error = %q, want invalid config YAML", err.Error())
	}
}

func TestShowConfigReturnsUnsupportedVersionError(t *testing.T) {
	tests := []struct {
		name    string
		version int
	}{
		{name: "missing version", version: 0},
		{name: "zero version", version: 0},
		{name: "negative version", version: -1},
		{name: "future version", version: 2},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fileSystem := newFakeConfigFileSystem()
			seedConfigProject(fileSystem, "config contents")
			parser := &fakeConfigParser{config: domain.LocalConfig{Version: test.version}}

			_, err := NewShowConfig(fileSystem, parser).Execute(ShowConfigInput{ProjectRoot: "/project"})
			if err == nil {
				t.Fatalf("Execute() error = nil, want unsupported version error")
			}
			if !strings.Contains(err.Error(), "unsupported config version") {
				t.Fatalf("Execute() error = %q, want unsupported config version", err.Error())
			}
			if !strings.Contains(err.Error(), "supported version is 1") {
				t.Fatalf("Execute() error = %q, want supported version is 1", err.Error())
			}
		})
	}
}

func TestShowConfigDoesNotReadConfigWhenProjectRootUnavailable(t *testing.T) {
	fileSystem := newFakeConfigFileSystem()
	fileSystem.files[localConfigPath] = "config contents"
	parser := &fakeConfigParser{config: domain.LocalConfig{Version: 1}}

	_, err := NewShowConfig(fileSystem, parser).Execute(ShowConfigInput{ProjectRoot: "/project"})
	if err == nil {
		t.Fatalf("Execute() error = nil, want unavailable project root error")
	}
	if fileSystem.readCount != 0 {
		t.Fatalf("ReadFile calls = %d, want 0", fileSystem.readCount)
	}
}

func TestShowConfigDoesNotCallParserWhenConfigIsMissingOrUnreadable(t *testing.T) {
	tests := []struct {
		name  string
		setup func(fileSystem *fakeConfigFileSystem)
	}{
		{
			name: "missing",
			setup: func(fileSystem *fakeConfigFileSystem) {
				fileSystem.directories["."] = true
			},
		},
		{
			name: "unreadable",
			setup: func(fileSystem *fakeConfigFileSystem) {
				seedConfigProject(fileSystem, "config contents")
				fileSystem.readErrors[localConfigPath] = errors.New("permission denied")
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fileSystem := newFakeConfigFileSystem()
			test.setup(fileSystem)
			parser := &fakeConfigParser{config: domain.LocalConfig{Version: 1}}

			_, err := NewShowConfig(fileSystem, parser).Execute(ShowConfigInput{ProjectRoot: "/project"})
			if err == nil {
				t.Fatalf("Execute() error = nil, want config access error")
			}
			if parser.calls != 0 {
				t.Fatalf("parser calls = %d, want 0", parser.calls)
			}
		})
	}
}

type fakeConfigFileSystem struct {
	directories     map[string]bool
	files           map[string]string
	directoryErrors map[string]error
	fileErrors      map[string]error
	readErrors      map[string]error
	directoryChecks int
	fileChecks      int
	readCount       int
}

func newFakeConfigFileSystem() *fakeConfigFileSystem {
	return &fakeConfigFileSystem{
		directories:     make(map[string]bool),
		files:           make(map[string]string),
		directoryErrors: make(map[string]error),
		fileErrors:      make(map[string]error),
		readErrors:      make(map[string]error),
	}
}

func (fileSystem *fakeConfigFileSystem) DirectoryExists(_ string, relativePath string) (bool, error) {
	fileSystem.directoryChecks++
	if err := fileSystem.directoryErrors[relativePath]; err != nil {
		return false, err
	}
	return fileSystem.directories[relativePath], nil
}

func (fileSystem *fakeConfigFileSystem) FileExists(_ string, relativePath string) (bool, error) {
	fileSystem.fileChecks++
	if err := fileSystem.fileErrors[relativePath]; err != nil {
		return false, err
	}
	_, exists := fileSystem.files[relativePath]
	return exists, nil
}

func (fileSystem *fakeConfigFileSystem) ReadFile(_ string, relativePath string) (string, error) {
	fileSystem.readCount++
	if err := fileSystem.readErrors[relativePath]; err != nil {
		return "", err
	}
	return fileSystem.files[relativePath], nil
}

func (fileSystem *fakeConfigFileSystem) operationCount() int {
	return fileSystem.directoryChecks + fileSystem.fileChecks + fileSystem.readCount
}

type fakeConfigParser struct {
	config   domain.LocalConfig
	err      error
	calls    int
	contents string
}

func (parser *fakeConfigParser) ParseLocalConfig(contents string) (domain.LocalConfig, error) {
	parser.calls++
	parser.contents = contents
	if parser.err != nil {
		return domain.LocalConfig{}, parser.err
	}
	return parser.config, nil
}

func seedConfigProject(fileSystem *fakeConfigFileSystem, contents string) {
	fileSystem.directories["."] = true
	fileSystem.files[localConfigPath] = contents
}

func completeDomainConfig() domain.LocalConfig {
	return domain.LocalConfig{
		Version: 1,
		Defaults: domain.ConfigDefaults{
			AgentRole:      "implementer",
			GenerationMode: "blank",
		},
		Validation: domain.ConfigValidation{
			RequireAllChangeFiles: true,
		},
		Review: domain.ConfigReview{
			RequireCompletedTasks: true,
		},
		Archive: domain.ConfigArchive{
			DateLayout: "2006-01-02",
		},
		Scan: domain.ConfigScan{
			IncludeCommonProjectFiles: true,
		},
		Output: domain.ConfigOutput{
			Format: "text",
		},
		Templates: domain.NewConfigTemplates(domain.EmptyConfigTemplateAliases()),
	}
}
