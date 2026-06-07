package usecase

import (
	"errors"
	"strings"
	"testing"

	"github.com/guferreira1/spec-harbor/internal/core/domain"
)

func TestScanProjectReturnsPopulatedResultForKnownSignals(t *testing.T) {
	fileSystem := newFakeScanFileSystem()
	fileSystem.topLevelEntries = []string{"go.mod", "package.json", "package-lock.json", "Dockerfile", ".github", "openspec"}
	fileSystem.files["go.mod"] = true
	fileSystem.files["package.json"] = true
	fileSystem.files["package-lock.json"] = true
	fileSystem.files["Dockerfile"] = true
	fileSystem.files["openspec/project.md"] = true
	fileSystem.directories[".github/workflows"] = true
	fileSystem.directories["openspec/changes"] = true

	result, err := NewScanProject(fileSystem).Execute(ScanProjectInput{ProjectRoot: "/project"})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if result.ProjectRoot != "/project" {
		t.Fatalf("ProjectRoot = %q, want /project", result.ProjectRoot)
	}
	assertScanSignals(t, "ecosystems", result.Ecosystems, []string{"go.mod", "package.json"})
	assertScanSignals(t, "package managers", result.PackageManagers, []string{"package-lock.json"})
	assertScanSignals(t, "ci", result.CIProviders, []string{".github/workflows/"})
	assertScanSignals(t, "containers", result.ContainerDeployments, []string{"Dockerfile"})
	assertScanSignals(t, "specharbor", result.SpecHarborSignals, []string{"openspec/project.md", "openspec/changes/"})

	if strings.Join(result.TestCommandHints, "|") != "go test ./...|npm test" {
		t.Fatalf("TestCommandHints = %v, want [go test ./... npm test]", result.TestCommandHints)
	}
	if len(result.Notes) != 1 || result.Notes[0] != "No Kubernetes manifests detected." {
		t.Fatalf("Notes = %v, want single Kubernetes note", result.Notes)
	}
}

func TestScanProjectReturnsEmptyResultWithNoSignalsNote(t *testing.T) {
	fileSystem := newFakeScanFileSystem()

	result, err := NewScanProject(fileSystem).Execute(ScanProjectInput{ProjectRoot: "/project"})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if len(result.Ecosystems) != 0 ||
		len(result.PackageManagers) != 0 ||
		len(result.TestCommandHints) != 0 ||
		len(result.CIProviders) != 0 ||
		len(result.ContainerDeployments) != 0 ||
		len(result.SpecHarborSignals) != 0 {
		t.Fatalf("result = %+v, want all detection sections empty", result)
	}
	if len(result.Notes) != 1 || result.Notes[0] != "No known project signals detected." {
		t.Fatalf("Notes = %v, want single no-signals note", result.Notes)
	}
}

func TestScanProjectDetectsDotNetSuffixesThroughTopLevelListing(t *testing.T) {
	fileSystem := newFakeScanFileSystem()
	fileSystem.topLevelEntries = []string{"MyApp.csproj", "MyApp.sln", "Program.cs"}

	result, err := NewScanProject(fileSystem).Execute(ScanProjectInput{ProjectRoot: "/project"})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	assertScanSignals(t, "ecosystems", result.Ecosystems, []string{".csproj", ".sln"})
	for _, detection := range result.Ecosystems {
		if detection.Name != ".NET" {
			t.Fatalf("ecosystem name = %q, want .NET", detection.Name)
		}
	}
	if strings.Join(result.TestCommandHints, "|") != "dotnet test" {
		t.Fatalf("TestCommandHints = %v, want single deduplicated dotnet test", result.TestCommandHints)
	}
}

func TestScanProjectDetectsNonGoStackSignals(t *testing.T) {
	fileSystem := newFakeScanFileSystem()
	fileSystem.topLevelEntries = []string{"pom.xml", "pyproject.toml", "Cargo.toml", ".gitlab-ci.yml", "Jenkinsfile", "Dockerfile"}
	fileSystem.files["pom.xml"] = true
	fileSystem.files["pyproject.toml"] = true
	fileSystem.files["Cargo.toml"] = true
	fileSystem.files[".gitlab-ci.yml"] = true
	fileSystem.files["Jenkinsfile"] = true
	fileSystem.files["Dockerfile"] = true

	result, err := NewScanProject(fileSystem).Execute(ScanProjectInput{ProjectRoot: "/project"})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	assertScanSignals(t, "ecosystems", result.Ecosystems, []string{"pom.xml", "pyproject.toml", "Cargo.toml"})
	assertScanSignals(t, "ci", result.CIProviders, []string{".gitlab-ci.yml", "Jenkinsfile"})
	assertScanSignals(t, "containers", result.ContainerDeployments, []string{"Dockerfile"})

	for _, detection := range result.Ecosystems {
		if detection.Name == "Go" || detection.Signal == "go.mod" {
			t.Fatalf("unexpected Go detection %+v in a non-Go project", detection)
		}
	}
	if strings.Join(result.TestCommandHints, "|") != "mvn test|pytest|cargo test" {
		t.Fatalf("TestCommandHints = %v, want [mvn test pytest cargo test]", result.TestCommandHints)
	}
	if len(result.Notes) != 1 || result.Notes[0] != "No Kubernetes manifests detected." {
		t.Fatalf("Notes = %v, want single Kubernetes note", result.Notes)
	}
}

func TestScanProjectSuffixRulesProduceOneDetectionPerRuleNotPerFile(t *testing.T) {
	fileSystem := newFakeScanFileSystem()
	fileSystem.topLevelEntries = []string{"First.csproj", "Second.csproj", "Primary.sln", "Secondary.sln", "Program.cs"}

	result, err := NewScanProject(fileSystem).Execute(ScanProjectInput{ProjectRoot: "/project"})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	assertScanSignals(t, "ecosystems", result.Ecosystems, []string{".csproj", ".sln"})
	if strings.Join(result.TestCommandHints, "|") != "dotnet test" {
		t.Fatalf("TestCommandHints = %v, want single deduplicated dotnet test", result.TestCommandHints)
	}
}

func TestScanProjectDetectsMultipleNodeSignalsAsSeparateDetections(t *testing.T) {
	fileSystem := newFakeScanFileSystem()
	fileSystem.topLevelEntries = []string{"package.json", "tsconfig.json"}
	fileSystem.files["package.json"] = true
	fileSystem.files["tsconfig.json"] = true

	result, err := NewScanProject(fileSystem).Execute(ScanProjectInput{ProjectRoot: "/project"})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	assertScanSignals(t, "ecosystems", result.Ecosystems, []string{"package.json", "tsconfig.json"})
	for _, detection := range result.Ecosystems {
		if detection.Name != "Node" {
			t.Fatalf("ecosystem name = %q, want Node", detection.Name)
		}
	}
	if strings.Join(result.TestCommandHints, "|") != "npm test" {
		t.Fatalf("TestCommandHints = %v, want single npm test (tsconfig has no hint)", result.TestCommandHints)
	}
}

func TestScanProjectRejectsEmptyProjectRoot(t *testing.T) {
	fileSystem := newFakeScanFileSystem()

	_, err := NewScanProject(fileSystem).Execute(ScanProjectInput{ProjectRoot: "  "})
	if err == nil || !strings.Contains(err.Error(), "project root is required") {
		t.Fatalf("Execute() error = %v, want project root is required", err)
	}
	if fileSystem.listCount != 0 {
		t.Fatalf("list calls = %d, want 0", fileSystem.listCount)
	}
}

func TestScanProjectReturnsExecutionErrorForUnlistableRoot(t *testing.T) {
	wantErr := errors.New("permission denied")
	fileSystem := newFakeScanFileSystem()
	fileSystem.listError = wantErr

	_, err := NewScanProject(fileSystem).Execute(ScanProjectInput{ProjectRoot: "/project"})
	if err == nil {
		t.Fatalf("Execute() error = nil, want unlistable root error")
	}
	if !errors.Is(err, wantErr) {
		t.Fatalf("Execute() error = %v, want wrapping %v", err, wantErr)
	}
	if !strings.Contains(err.Error(), "list project root") {
		t.Fatalf("Execute() error = %q, want to mention list project root", err.Error())
	}
}

func TestScanProjectReturnsExecutionErrorForExistenceCheckFailures(t *testing.T) {
	wantErr := errors.New("filesystem unavailable")
	tests := []struct {
		name  string
		setup func(fileSystem *fakeScanFileSystem)
	}{
		{
			name: "file check",
			setup: func(fileSystem *fakeScanFileSystem) {
				fileSystem.fileErrors["go.mod"] = wantErr
			},
		},
		{
			name: "directory check",
			setup: func(fileSystem *fakeScanFileSystem) {
				fileSystem.directoryErrors[".github/workflows"] = wantErr
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fileSystem := newFakeScanFileSystem()
			test.setup(fileSystem)

			_, err := NewScanProject(fileSystem).Execute(ScanProjectInput{ProjectRoot: "/project"})
			if err == nil {
				t.Fatalf("Execute() error = nil, want existence-check error")
			}
			if !errors.Is(err, wantErr) {
				t.Fatalf("Execute() error = %v, want wrapping %v", err, wantErr)
			}
		})
	}
}

func TestScanProjectListsOnlyProjectRootWithoutRecursion(t *testing.T) {
	fileSystem := newFakeScanFileSystem()
	fileSystem.topLevelEntries = []string{"go.mod", "MyApp.csproj"}
	fileSystem.files["go.mod"] = true
	fileSystem.directories["openspec/changes"] = true

	if _, err := NewScanProject(fileSystem).Execute(ScanProjectInput{ProjectRoot: "/project"}); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if fileSystem.listCount != 1 {
		t.Fatalf("list calls = %d, want exactly 1", fileSystem.listCount)
	}
	if len(fileSystem.listedPaths) != 1 || fileSystem.listedPaths[0] != scanProjectRootListing {
		t.Fatalf("listed paths = %v, want only project root listing %q", fileSystem.listedPaths, scanProjectRootListing)
	}
}

func TestScanProjectRejectsMissingDependencies(t *testing.T) {
	_, err := (*ScanProject)(nil).Execute(ScanProjectInput{})
	if err == nil || !strings.Contains(err.Error(), "scan project use case is required") {
		t.Fatalf("nil use case error = %v, want scan project use case is required", err)
	}

	_, err = NewScanProject(nil).Execute(ScanProjectInput{})
	if err == nil || !strings.Contains(err.Error(), "scan filesystem is required") {
		t.Fatalf("nil filesystem error = %v, want scan filesystem is required", err)
	}
}

func assertScanSignals(t *testing.T, label string, detections []domain.ScanDetection, want []string) {
	t.Helper()

	if len(detections) != len(want) {
		t.Fatalf("%s count = %d, want %d (%v)", label, len(detections), len(want), detections)
	}
	for index, signal := range want {
		if detections[index].Signal != signal {
			t.Fatalf("%s[%d].Signal = %q, want %q", label, index, detections[index].Signal, signal)
		}
	}
}

type fakeScanFileSystem struct {
	topLevelEntries []string
	files           map[string]bool
	directories     map[string]bool
	fileErrors      map[string]error
	directoryErrors map[string]error
	listError       error
	listCount       int
	listedPaths     []string
}

func newFakeScanFileSystem() *fakeScanFileSystem {
	return &fakeScanFileSystem{
		files:           make(map[string]bool),
		directories:     make(map[string]bool),
		fileErrors:      make(map[string]error),
		directoryErrors: make(map[string]error),
	}
}

func (fileSystem *fakeScanFileSystem) FileExists(_ string, relativePath string) (bool, error) {
	if err := fileSystem.fileErrors[relativePath]; err != nil {
		return false, err
	}
	return fileSystem.files[relativePath], nil
}

func (fileSystem *fakeScanFileSystem) DirectoryExists(_ string, relativePath string) (bool, error) {
	if err := fileSystem.directoryErrors[relativePath]; err != nil {
		return false, err
	}
	return fileSystem.directories[relativePath], nil
}

func (fileSystem *fakeScanFileSystem) ListDirectoryNames(_ string, relativePath string) ([]string, error) {
	fileSystem.listCount++
	fileSystem.listedPaths = append(fileSystem.listedPaths, relativePath)
	if fileSystem.listError != nil {
		return nil, fileSystem.listError
	}
	return append([]string(nil), fileSystem.topLevelEntries...), nil
}
