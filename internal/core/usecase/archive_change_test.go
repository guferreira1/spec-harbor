package usecase

import (
	"errors"
	"strings"
	"testing"

	"github.com/guferreira1/spec-harbor/internal/core/domain"
)

func TestArchiveChangeMovesChangeDirectory(t *testing.T) {
	changeID := "implement-archive-foundation"
	archiveDate := "2026-06-06"
	sourcePath := openspecChangesDirectory + "/" + changeID
	archiveDateDirectory := openspecArchiveDirectory + "/" + archiveDate
	archivePath := archiveDateDirectory + "/" + changeID
	fileSystem := newFakeArchiveFileSystem()
	seedArchiveOpenSpecProject(fileSystem)
	seedArchiveChange(fileSystem, changeID)

	result, err := NewArchiveChange(fileSystem).Execute(ArchiveChangeInput{
		ProjectRoot: "/project",
		ChangeID:    changeID,
		ArchiveDate: archiveDate,
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if result.ChangeID != changeID {
		t.Fatalf("ChangeID = %q, want %q", result.ChangeID, changeID)
	}
	if result.SourcePath != sourcePath {
		t.Fatalf("SourcePath = %q, want %q", result.SourcePath, sourcePath)
	}
	if result.ArchivePath != archivePath {
		t.Fatalf("ArchivePath = %q, want %q", result.ArchivePath, archivePath)
	}
	if result.ArchiveDate != archiveDate {
		t.Fatalf("ArchiveDate = %q, want %q", result.ArchiveDate, archiveDate)
	}
	if result.MovedDirectory != (domain.ArchiveMovedDirectory{SourcePath: sourcePath, ArchivePath: archivePath}) {
		t.Fatalf("MovedDirectory = %#v, want source and archive paths", result.MovedDirectory)
	}
	if !result.Moved() {
		t.Fatalf("Moved() = false, want true")
	}
	if !containsString(fileSystem.createdDirectories, openspecArchiveDirectory) {
		t.Fatalf("created directories = %v, want %q", fileSystem.createdDirectories, openspecArchiveDirectory)
	}
	if !containsString(fileSystem.createdDirectories, archiveDateDirectory) {
		t.Fatalf("created directories = %v, want %q", fileSystem.createdDirectories, archiveDateDirectory)
	}
	if !fileSystem.directories[archivePath] {
		t.Fatalf("archive directory %q was not created by move", archivePath)
	}
	if fileSystem.directories[sourcePath] {
		t.Fatalf("source directory %q still exists after move", sourcePath)
	}
	if !fileSystem.files[archivePath+"/nested/proposal.md"] {
		t.Fatalf("nested file was not moved to archive path")
	}
	if fileSystem.files[sourcePath+"/nested/proposal.md"] {
		t.Fatalf("nested file still exists under source path")
	}
	if len(fileSystem.movedDirectories) != 1 {
		t.Fatalf("moved directories = %v, want one move", fileSystem.movedDirectories)
	}
	if fileSystem.movedDirectories[0] != (archiveMove{source: sourcePath, destination: archivePath}) {
		t.Fatalf("move = %#v, want %s -> %s", fileSystem.movedDirectories[0], sourcePath, archivePath)
	}
}

func TestArchiveChangeRejectsInvalidInputBeforeFilesystemOperations(t *testing.T) {
	tests := []struct {
		name  string
		input ArchiveChangeInput
		want  string
	}{
		{
			name:  "empty project root",
			input: ArchiveChangeInput{ProjectRoot: " ", ChangeID: "change", ArchiveDate: "2026-06-06"},
			want:  "project root is required",
		},
		{
			name:  "empty change id",
			input: ArchiveChangeInput{ProjectRoot: "/project", ChangeID: " ", ArchiveDate: "2026-06-06"},
			want:  "change id is required",
		},
		{
			name:  "empty archive date",
			input: ArchiveChangeInput{ProjectRoot: "/project", ChangeID: "change", ArchiveDate: " "},
			want:  "archive date is required",
		},
		{
			name:  "slash archive date",
			input: ArchiveChangeInput{ProjectRoot: "/project", ChangeID: "change", ArchiveDate: "2026/06/06"},
			want:  "archive date must be formatted as YYYY-MM-DD",
		},
		{
			name:  "non padded archive date",
			input: ArchiveChangeInput{ProjectRoot: "/project", ChangeID: "change", ArchiveDate: "2026-6-6"},
			want:  "archive date must be formatted as YYYY-MM-DD",
		},
		{
			name:  "invalid calendar archive date",
			input: ArchiveChangeInput{ProjectRoot: "/project", ChangeID: "change", ArchiveDate: "2026-02-30"},
			want:  "archive date must be formatted as YYYY-MM-DD",
		},
		{
			name:  "dot id",
			input: ArchiveChangeInput{ProjectRoot: "/project", ChangeID: ".", ArchiveDate: "2026-06-06"},
			want:  "change id must be a safe single path segment",
		},
		{
			name:  "dotdot id",
			input: ArchiveChangeInput{ProjectRoot: "/project", ChangeID: "..", ArchiveDate: "2026-06-06"},
			want:  "change id must be a safe single path segment",
		},
		{
			name:  "traversal id",
			input: ArchiveChangeInput{ProjectRoot: "/project", ChangeID: "../unsafe", ArchiveDate: "2026-06-06"},
			want:  "change id must be a safe single path segment",
		},
		{
			name:  "absolute id",
			input: ArchiveChangeInput{ProjectRoot: "/project", ChangeID: "/unsafe", ArchiveDate: "2026-06-06"},
			want:  "change id must be a safe single path segment",
		},
		{
			name:  "backslash id",
			input: ArchiveChangeInput{ProjectRoot: "/project", ChangeID: `bad\id`, ArchiveDate: "2026-06-06"},
			want:  "change id must be a safe single path segment",
		},
		{
			name:  "colon id",
			input: ArchiveChangeInput{ProjectRoot: "/project", ChangeID: "bad:id", ArchiveDate: "2026-06-06"},
			want:  "change id must be a safe single path segment",
		},
		{
			name:  "leading dash id",
			input: ArchiveChangeInput{ProjectRoot: "/project", ChangeID: "-bad", ArchiveDate: "2026-06-06"},
			want:  "change id must be a safe single path segment",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fileSystem := newFakeArchiveFileSystem()

			_, err := NewArchiveChange(fileSystem).Execute(test.input)
			if err == nil {
				t.Fatalf("Execute() error = nil, want %q", test.want)
			}
			if !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Execute() error = %q, want %q", err.Error(), test.want)
			}
			if fileSystem.operationCount() != 0 {
				t.Fatalf("filesystem operations = %d, want 0", fileSystem.operationCount())
			}
		})
	}
}

func TestArchiveChangeRejectsMissingOpenSpecProjectBeforeArchiveWrites(t *testing.T) {
	tests := []struct {
		name  string
		setup func(fileSystem *fakeArchiveFileSystem)
	}{
		{
			name: "missing project file",
			setup: func(fileSystem *fakeArchiveFileSystem) {
				fileSystem.directories[openspecChangesDirectory] = true
			},
		},
		{
			name: "missing changes directory",
			setup: func(fileSystem *fakeArchiveFileSystem) {
				fileSystem.files[openspecProjectFile] = true
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fileSystem := newFakeArchiveFileSystem()
			test.setup(fileSystem)

			_, err := NewArchiveChange(fileSystem).Execute(ArchiveChangeInput{
				ProjectRoot: "/project",
				ChangeID:    "change",
				ArchiveDate: "2026-06-06",
			})
			if err == nil {
				t.Fatalf("Execute() error = nil, want missing project structure error")
			}
			for _, want := range []string{"OpenSpec project structure is missing", "specharbor init"} {
				if !strings.Contains(err.Error(), want) {
					t.Fatalf("Execute() error = %q, want to contain %q", err.Error(), want)
				}
			}
			assertNoArchiveWritesOrMoves(t, fileSystem)
			if fileSystem.directories[openspecArchiveDirectory] {
				t.Fatalf("archive created openspec/archive")
			}
		})
	}
}

func TestArchiveChangeRejectsMissingOrFileSourceBeforeArchiveWrites(t *testing.T) {
	tests := []struct {
		name  string
		setup func(fileSystem *fakeArchiveFileSystem, sourcePath string)
		want  string
	}{
		{
			name:  "missing source",
			setup: func(_ *fakeArchiveFileSystem, _ string) {},
			want:  "missing change directory",
		},
		{
			name: "source is file",
			setup: func(fileSystem *fakeArchiveFileSystem, sourcePath string) {
				fileSystem.files[sourcePath] = true
			},
			want: "source change path must be a directory",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			changeID := "change"
			sourcePath := openspecChangesDirectory + "/" + changeID
			fileSystem := newFakeArchiveFileSystem()
			seedArchiveOpenSpecProject(fileSystem)
			test.setup(fileSystem, sourcePath)

			_, err := NewArchiveChange(fileSystem).Execute(ArchiveChangeInput{
				ProjectRoot: "/project",
				ChangeID:    changeID,
				ArchiveDate: "2026-06-06",
			})
			if err == nil {
				t.Fatalf("Execute() error = nil, want %q", test.want)
			}
			if !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Execute() error = %q, want %q", err.Error(), test.want)
			}
			assertNoArchiveWritesOrMoves(t, fileSystem)
		})
	}
}

func TestArchiveChangeRejectsArchiveParentsThatAreFiles(t *testing.T) {
	tests := []struct {
		name  string
		setup func(fileSystem *fakeArchiveFileSystem)
		want  string
	}{
		{
			name: "archive root is file",
			setup: func(fileSystem *fakeArchiveFileSystem) {
				fileSystem.files[openspecArchiveDirectory] = true
			},
			want: "archive path must be a directory: " + openspecArchiveDirectory,
		},
		{
			name: "archive date directory is file",
			setup: func(fileSystem *fakeArchiveFileSystem) {
				fileSystem.directories[openspecArchiveDirectory] = true
				fileSystem.files[openspecArchiveDirectory+"/2026-06-06"] = true
			},
			want: "archive path must be a directory: " + openspecArchiveDirectory + "/2026-06-06",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fileSystem := newFakeArchiveFileSystem()
			seedArchiveOpenSpecProject(fileSystem)
			seedArchiveChange(fileSystem, "change")
			test.setup(fileSystem)

			_, err := NewArchiveChange(fileSystem).Execute(ArchiveChangeInput{
				ProjectRoot: "/project",
				ChangeID:    "change",
				ArchiveDate: "2026-06-06",
			})
			if err == nil {
				t.Fatalf("Execute() error = nil, want %q", test.want)
			}
			if !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Execute() error = %q, want %q", err.Error(), test.want)
			}
			assertNoArchiveWritesOrMoves(t, fileSystem)
		})
	}
}

func TestArchiveChangeRejectsExistingArchiveTargetWithoutOverwrite(t *testing.T) {
	tests := []struct {
		name  string
		setup func(fileSystem *fakeArchiveFileSystem, archivePath string)
		check func(t *testing.T, fileSystem *fakeArchiveFileSystem, archivePath string)
	}{
		{
			name: "target directory exists",
			setup: func(fileSystem *fakeArchiveFileSystem, archivePath string) {
				fileSystem.directories[archivePath] = true
				fileSystem.files[archivePath+"/existing.md"] = true
			},
			check: func(t *testing.T, fileSystem *fakeArchiveFileSystem, archivePath string) {
				t.Helper()
				if !fileSystem.directories[archivePath] || !fileSystem.files[archivePath+"/existing.md"] {
					t.Fatalf("existing archive directory content was overwritten")
				}
			},
		},
		{
			name: "target file exists",
			setup: func(fileSystem *fakeArchiveFileSystem, archivePath string) {
				fileSystem.files[archivePath] = true
			},
			check: func(t *testing.T, fileSystem *fakeArchiveFileSystem, archivePath string) {
				t.Helper()
				if !fileSystem.files[archivePath] {
					t.Fatalf("existing archive file was overwritten")
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			changeID := "change"
			sourcePath := openspecChangesDirectory + "/" + changeID
			archivePath := openspecArchiveDirectory + "/2026-06-06/" + changeID
			fileSystem := newFakeArchiveFileSystem()
			seedArchiveOpenSpecProject(fileSystem)
			seedArchiveChange(fileSystem, changeID)
			fileSystem.directories[openspecArchiveDirectory] = true
			fileSystem.directories[openspecArchiveDirectory+"/2026-06-06"] = true
			test.setup(fileSystem, archivePath)

			_, err := NewArchiveChange(fileSystem).Execute(ArchiveChangeInput{
				ProjectRoot: "/project",
				ChangeID:    changeID,
				ArchiveDate: "2026-06-06",
			})
			if err == nil {
				t.Fatalf("Execute() error = nil, want existing archive target error")
			}
			if !strings.Contains(err.Error(), "archive target already exists") {
				t.Fatalf("Execute() error = %q, want existing archive target", err.Error())
			}
			if len(fileSystem.movedDirectories) != 0 {
				t.Fatalf("moved directories = %v, want none", fileSystem.movedDirectories)
			}
			if !fileSystem.directories[sourcePath] {
				t.Fatalf("source directory was removed")
			}
			test.check(t, fileSystem, archivePath)
		})
	}
}

func TestArchiveChangeReturnsFilesystemErrors(t *testing.T) {
	wantErr := errors.New("filesystem unavailable")
	changeID := "change"
	sourcePath := openspecChangesDirectory + "/" + changeID
	archiveDateDirectory := openspecArchiveDirectory + "/2026-06-06"
	archivePath := archiveDateDirectory + "/" + changeID

	tests := []struct {
		name  string
		setup func(fileSystem *fakeArchiveFileSystem)
	}{
		{
			name: "project file check",
			setup: func(fileSystem *fakeArchiveFileSystem) {
				fileSystem.fileErrors[openspecProjectFile] = wantErr
			},
		},
		{
			name: "changes directory check",
			setup: func(fileSystem *fakeArchiveFileSystem) {
				fileSystem.files[openspecProjectFile] = true
				fileSystem.directoryErrors[openspecChangesDirectory] = wantErr
			},
		},
		{
			name: "source directory check",
			setup: func(fileSystem *fakeArchiveFileSystem) {
				seedArchiveOpenSpecProject(fileSystem)
				fileSystem.directoryErrors[sourcePath] = wantErr
			},
		},
		{
			name: "source path check",
			setup: func(fileSystem *fakeArchiveFileSystem) {
				seedArchiveOpenSpecProject(fileSystem)
				fileSystem.pathErrors[sourcePath] = wantErr
			},
		},
		{
			name: "archive root path check",
			setup: func(fileSystem *fakeArchiveFileSystem) {
				seedArchiveOpenSpecProject(fileSystem)
				seedArchiveChange(fileSystem, changeID)
				fileSystem.pathErrors[openspecArchiveDirectory] = wantErr
			},
		},
		{
			name: "archive root directory check",
			setup: func(fileSystem *fakeArchiveFileSystem) {
				seedArchiveOpenSpecProject(fileSystem)
				seedArchiveChange(fileSystem, changeID)
				fileSystem.directories[openspecArchiveDirectory] = true
				fileSystem.directoryErrors[openspecArchiveDirectory] = wantErr
			},
		},
		{
			name: "archive root create",
			setup: func(fileSystem *fakeArchiveFileSystem) {
				seedArchiveOpenSpecProject(fileSystem)
				seedArchiveChange(fileSystem, changeID)
				fileSystem.createDirectoryErrors[openspecArchiveDirectory] = wantErr
			},
		},
		{
			name: "archive date path check",
			setup: func(fileSystem *fakeArchiveFileSystem) {
				seedArchiveOpenSpecProject(fileSystem)
				seedArchiveChange(fileSystem, changeID)
				fileSystem.directories[openspecArchiveDirectory] = true
				fileSystem.pathErrors[archiveDateDirectory] = wantErr
			},
		},
		{
			name: "archive date directory check",
			setup: func(fileSystem *fakeArchiveFileSystem) {
				seedArchiveOpenSpecProject(fileSystem)
				seedArchiveChange(fileSystem, changeID)
				fileSystem.directories[openspecArchiveDirectory] = true
				fileSystem.directories[archiveDateDirectory] = true
				fileSystem.directoryErrors[archiveDateDirectory] = wantErr
			},
		},
		{
			name: "archive date create",
			setup: func(fileSystem *fakeArchiveFileSystem) {
				seedArchiveOpenSpecProject(fileSystem)
				seedArchiveChange(fileSystem, changeID)
				fileSystem.directories[openspecArchiveDirectory] = true
				fileSystem.createDirectoryErrors[archiveDateDirectory] = wantErr
			},
		},
		{
			name: "archive target path check",
			setup: func(fileSystem *fakeArchiveFileSystem) {
				seedArchiveOpenSpecProject(fileSystem)
				seedArchiveChange(fileSystem, changeID)
				fileSystem.directories[openspecArchiveDirectory] = true
				fileSystem.directories[archiveDateDirectory] = true
				fileSystem.pathErrors[archivePath] = wantErr
			},
		},
		{
			name: "move directory",
			setup: func(fileSystem *fakeArchiveFileSystem) {
				seedArchiveOpenSpecProject(fileSystem)
				seedArchiveChange(fileSystem, changeID)
				fileSystem.directories[openspecArchiveDirectory] = true
				fileSystem.directories[archiveDateDirectory] = true
				fileSystem.moveErrors[moveKey(sourcePath, archivePath)] = wantErr
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fileSystem := newFakeArchiveFileSystem()
			test.setup(fileSystem)

			_, err := NewArchiveChange(fileSystem).Execute(ArchiveChangeInput{
				ProjectRoot: "/project",
				ChangeID:    changeID,
				ArchiveDate: "2026-06-06",
			})
			if err == nil {
				t.Fatalf("Execute() error = nil, want filesystem error")
			}
			if !errors.Is(err, wantErr) {
				t.Fatalf("Execute() error = %v, want wrapping %v", err, wantErr)
			}
		})
	}
}

func TestArchiveChangeRejectsMissingDependencies(t *testing.T) {
	_, err := (*ArchiveChange)(nil).Execute(ArchiveChangeInput{})
	if err == nil || !strings.Contains(err.Error(), "archive change use case is required") {
		t.Fatalf("nil use case error = %v, want required use case", err)
	}

	_, err = NewArchiveChange(nil).Execute(ArchiveChangeInput{})
	if err == nil || !strings.Contains(err.Error(), "archive filesystem is required") {
		t.Fatalf("nil filesystem error = %v, want required filesystem", err)
	}
}

func seedArchiveOpenSpecProject(fileSystem *fakeArchiveFileSystem) {
	fileSystem.files[openspecProjectFile] = true
	fileSystem.directories[openspecChangesDirectory] = true
}

func seedArchiveChange(fileSystem *fakeArchiveFileSystem, changeID string) {
	sourcePath := openspecChangesDirectory + "/" + changeID
	fileSystem.directories[sourcePath] = true
	fileSystem.directories[sourcePath+"/nested"] = true
	fileSystem.files[sourcePath+"/nested/proposal.md"] = true
}

func assertNoArchiveWritesOrMoves(t *testing.T, fileSystem *fakeArchiveFileSystem) {
	t.Helper()

	if len(fileSystem.createdDirectories) != 0 {
		t.Fatalf("created directories = %v, want none", fileSystem.createdDirectories)
	}
	if len(fileSystem.movedDirectories) != 0 {
		t.Fatalf("moved directories = %v, want none", fileSystem.movedDirectories)
	}
}

type archiveMove struct {
	source      string
	destination string
}

type fakeArchiveFileSystem struct {
	directories           map[string]bool
	files                 map[string]bool
	directoryErrors       map[string]error
	fileErrors            map[string]error
	pathErrors            map[string]error
	createDirectoryErrors map[string]error
	moveErrors            map[string]error
	checkedDirectories    []string
	checkedFiles          []string
	checkedPaths          []string
	createdDirectories    []string
	movedDirectories      []archiveMove
}

func newFakeArchiveFileSystem() *fakeArchiveFileSystem {
	return &fakeArchiveFileSystem{
		directories:           make(map[string]bool),
		files:                 make(map[string]bool),
		directoryErrors:       make(map[string]error),
		fileErrors:            make(map[string]error),
		pathErrors:            make(map[string]error),
		createDirectoryErrors: make(map[string]error),
		moveErrors:            make(map[string]error),
	}
}

func (fileSystem *fakeArchiveFileSystem) DirectoryExists(_ string, relativePath string) (bool, error) {
	fileSystem.checkedDirectories = append(fileSystem.checkedDirectories, relativePath)
	if err := fileSystem.directoryErrors[relativePath]; err != nil {
		return false, err
	}
	return fileSystem.directories[relativePath], nil
}

func (fileSystem *fakeArchiveFileSystem) FileExists(_ string, relativePath string) (bool, error) {
	fileSystem.checkedFiles = append(fileSystem.checkedFiles, relativePath)
	if err := fileSystem.fileErrors[relativePath]; err != nil {
		return false, err
	}
	return fileSystem.files[relativePath], nil
}

func (fileSystem *fakeArchiveFileSystem) PathExists(_ string, relativePath string) (bool, error) {
	fileSystem.checkedPaths = append(fileSystem.checkedPaths, relativePath)
	if err := fileSystem.pathErrors[relativePath]; err != nil {
		return false, err
	}
	return fileSystem.directories[relativePath] || fileSystem.files[relativePath], nil
}

func (fileSystem *fakeArchiveFileSystem) CreateDirectory(_ string, relativePath string) error {
	fileSystem.createdDirectories = append(fileSystem.createdDirectories, relativePath)
	if err := fileSystem.createDirectoryErrors[relativePath]; err != nil {
		return err
	}
	fileSystem.directories[relativePath] = true
	return nil
}

func (fileSystem *fakeArchiveFileSystem) MoveDirectory(_ string, sourceRelativePath string, destinationRelativePath string) error {
	fileSystem.movedDirectories = append(fileSystem.movedDirectories, archiveMove{
		source:      sourceRelativePath,
		destination: destinationRelativePath,
	})
	if err := fileSystem.moveErrors[moveKey(sourceRelativePath, destinationRelativePath)]; err != nil {
		return err
	}
	if !fileSystem.directories[sourceRelativePath] {
		return errors.New("source directory is missing")
	}
	if fileSystem.directories[destinationRelativePath] || fileSystem.files[destinationRelativePath] {
		return errors.New("destination already exists")
	}

	fileSystem.moveDirectoryEntries(sourceRelativePath, destinationRelativePath)
	return nil
}

func (fileSystem *fakeArchiveFileSystem) moveDirectoryEntries(sourceRelativePath string, destinationRelativePath string) {
	directoriesToMove := matchingPaths(fileSystem.directories, sourceRelativePath)
	filesToMove := matchingPaths(fileSystem.files, sourceRelativePath)

	for _, directory := range directoriesToMove {
		delete(fileSystem.directories, directory)
		fileSystem.directories[destinationRelativePath+strings.TrimPrefix(directory, sourceRelativePath)] = true
	}
	for _, file := range filesToMove {
		delete(fileSystem.files, file)
		fileSystem.files[destinationRelativePath+strings.TrimPrefix(file, sourceRelativePath)] = true
	}
}

func (fileSystem *fakeArchiveFileSystem) operationCount() int {
	return len(fileSystem.checkedDirectories) +
		len(fileSystem.checkedFiles) +
		len(fileSystem.checkedPaths) +
		len(fileSystem.createdDirectories) +
		len(fileSystem.movedDirectories)
}

func matchingPaths(values map[string]bool, rootPath string) []string {
	var paths []string
	for path, exists := range values {
		if exists && (path == rootPath || strings.HasPrefix(path, rootPath+"/")) {
			paths = append(paths, path)
		}
	}
	return paths
}

func moveKey(source string, destination string) string {
	return source + " -> " + destination
}
