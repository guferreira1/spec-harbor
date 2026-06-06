package usecase

import "testing"

func TestInitializeProjectCreatesMissingItems(t *testing.T) {
	fileSystem := newFakeInitializationFileSystem()
	defaults := fakeInitializationDefaults{}
	useCase := NewInitializeProject(fileSystem, defaults)

	result, err := useCase.Execute(InitializeProjectInput{Root: "/project"})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if result.Status != InitializationStatusInitialized {
		t.Fatalf("Status = %q, want %q", result.Status, InitializationStatusInitialized)
	}
	if len(result.Created) != totalRequiredInitializationItems() {
		t.Fatalf("Created count = %d, want %d", len(result.Created), totalRequiredInitializationItems())
	}
	if len(result.Skipped) != 0 {
		t.Fatalf("Skipped count = %d, want 0", len(result.Skipped))
	}

	for _, directory := range requiredInitializationDirectories() {
		if !fileSystem.directories[directory] {
			t.Fatalf("directory %q was not created", directory)
		}
	}
	for _, file := range requiredInitializationFiles() {
		if fileSystem.files[file] != defaultContent(file) {
			t.Fatalf("file %q content = %q, want %q", file, fileSystem.files[file], defaultContent(file))
		}
	}
}

func TestInitializeProjectCreatesOnlyMissingPartialItems(t *testing.T) {
	fileSystem := newFakeInitializationFileSystem()
	fileSystem.directories["openspec"] = true
	fileSystem.directories["openspec/specs"] = true
	fileSystem.directories[".specharbor"] = true
	fileSystem.directories[".specharbor/rules"] = true
	fileSystem.files["openspec/project.md"] = "custom project"
	fileSystem.files[".specharbor/rules/global.md"] = "custom global"

	useCase := NewInitializeProject(fileSystem, fakeInitializationDefaults{})

	result, err := useCase.Execute(InitializeProjectInput{Root: "/project"})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if result.Status != InitializationStatusInitialized {
		t.Fatalf("Status = %q, want %q", result.Status, InitializationStatusInitialized)
	}
	if len(result.Created) != totalRequiredInitializationItems()-6 {
		t.Fatalf("Created count = %d, want %d", len(result.Created), totalRequiredInitializationItems()-6)
	}
	if len(result.Skipped) != 6 {
		t.Fatalf("Skipped count = %d, want 6", len(result.Skipped))
	}
	if fileSystem.files["openspec/project.md"] != "custom project" {
		t.Fatalf("existing project.md was overwritten")
	}
	if fileSystem.files[".specharbor/rules/global.md"] != "custom global" {
		t.Fatalf("existing global rule was overwritten")
	}
	if !containsInitializationItem(result.Created, InitializationItemKindDirectory, "openspec/changes") {
		t.Fatalf("missing directory was not reported as created")
	}
	if !containsInitializationItem(result.Skipped, InitializationItemKindFile, "openspec/project.md") {
		t.Fatalf("existing file was not reported as skipped")
	}
}

func TestInitializeProjectReportsAlreadyInitialized(t *testing.T) {
	fileSystem := newFakeInitializationFileSystem()
	for _, directory := range requiredInitializationDirectories() {
		fileSystem.directories[directory] = true
	}
	for _, file := range requiredInitializationFiles() {
		fileSystem.files[file] = "existing"
	}

	useCase := NewInitializeProject(fileSystem, fakeInitializationDefaults{})

	result, err := useCase.Execute(InitializeProjectInput{Root: "/project"})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if result.Status != InitializationStatusAlreadyInitialized {
		t.Fatalf("Status = %q, want %q", result.Status, InitializationStatusAlreadyInitialized)
	}
	if len(result.Created) != 0 {
		t.Fatalf("Created count = %d, want 0", len(result.Created))
	}
	if len(result.Skipped) != totalRequiredInitializationItems() {
		t.Fatalf("Skipped count = %d, want %d", len(result.Skipped), totalRequiredInitializationItems())
	}
}

func totalRequiredInitializationItems() int {
	return len(requiredInitializationDirectories()) + len(requiredInitializationFiles())
}

func containsInitializationItem(items []InitializationItem, kind InitializationItemKind, path string) bool {
	for _, item := range items {
		if item.Kind == kind && item.Path == path {
			return true
		}
	}
	return false
}

type fakeInitializationFileSystem struct {
	directories map[string]bool
	files       map[string]string
}

func newFakeInitializationFileSystem() *fakeInitializationFileSystem {
	return &fakeInitializationFileSystem{
		directories: make(map[string]bool),
		files:       make(map[string]string),
	}
}

func (fileSystem *fakeInitializationFileSystem) DirectoryExists(_ string, relativePath string) (bool, error) {
	return fileSystem.directories[relativePath], nil
}

func (fileSystem *fakeInitializationFileSystem) FileExists(_ string, relativePath string) (bool, error) {
	_, exists := fileSystem.files[relativePath]
	return exists, nil
}

func (fileSystem *fakeInitializationFileSystem) CreateDirectory(_ string, relativePath string) error {
	fileSystem.directories[relativePath] = true
	return nil
}

func (fileSystem *fakeInitializationFileSystem) WriteFileIfAbsent(_ string, relativePath string, contents string) (bool, error) {
	if _, exists := fileSystem.files[relativePath]; exists {
		return false, nil
	}
	fileSystem.files[relativePath] = contents
	return true, nil
}

type fakeInitializationDefaults struct{}

func (defaults fakeInitializationDefaults) ContentFor(relativePath string) (string, error) {
	return defaultContent(relativePath), nil
}

func defaultContent(relativePath string) string {
	return "default:" + relativePath
}
