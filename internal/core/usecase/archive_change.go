package usecase

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/guferreira1/spec-harbor/internal/core/domain"
	"github.com/guferreira1/spec-harbor/internal/core/ports"
)

const openspecArchiveDirectory = "openspec/archive"

type ArchiveChangeInput struct {
	ProjectRoot string
	ChangeID    string
	ArchiveDate string
}

type ArchiveChange struct {
	fileSystem ports.ArchiveFileSystem
}

func NewArchiveChange(fileSystem ports.ArchiveFileSystem) *ArchiveChange {
	return &ArchiveChange{fileSystem: fileSystem}
}

func (useCase *ArchiveChange) Execute(input ArchiveChangeInput) (domain.ArchiveResult, error) {
	if useCase == nil {
		return domain.ArchiveResult{}, errors.New("archive change use case is required")
	}
	if useCase.fileSystem == nil {
		return domain.ArchiveResult{}, errors.New("archive filesystem is required")
	}

	projectRoot, changeID, archiveDate, err := validateArchiveInput(input)
	if err != nil {
		return domain.ArchiveResult{}, err
	}

	sourcePath := openspecChangesDirectory + "/" + changeID
	archiveDateDirectory := openspecArchiveDirectory + "/" + archiveDate
	archivePath := archiveDateDirectory + "/" + changeID

	if err := useCase.requireOpenSpecProject(projectRoot); err != nil {
		return domain.ArchiveResult{}, err
	}
	if err := useCase.requireSourceChangeDirectory(projectRoot, sourcePath); err != nil {
		return domain.ArchiveResult{}, err
	}
	if err := useCase.ensureArchiveDirectory(projectRoot, openspecArchiveDirectory); err != nil {
		return domain.ArchiveResult{}, err
	}
	if err := useCase.ensureArchiveDirectory(projectRoot, archiveDateDirectory); err != nil {
		return domain.ArchiveResult{}, err
	}
	if err := useCase.requireArchiveTargetAvailable(projectRoot, archivePath); err != nil {
		return domain.ArchiveResult{}, err
	}
	if err := useCase.fileSystem.MoveDirectory(projectRoot, sourcePath, archivePath); err != nil {
		return domain.ArchiveResult{}, fmt.Errorf("move directory %s to %s: %w", sourcePath, archivePath, err)
	}

	return domain.NewArchiveResult(
		changeID,
		sourcePath,
		archivePath,
		archiveDate,
		domain.ArchiveMovedDirectory{SourcePath: sourcePath, ArchivePath: archivePath},
	), nil
}

func validateArchiveInput(input ArchiveChangeInput) (string, string, string, error) {
	projectRoot := strings.TrimSpace(input.ProjectRoot)
	if projectRoot == "" {
		return "", "", "", errors.New("project root is required")
	}

	changeID := strings.TrimSpace(input.ChangeID)
	if changeID == "" {
		return "", "", "", errors.New("change id is required")
	}
	if err := validateArchiveChangeID(changeID); err != nil {
		return "", "", "", err
	}

	archiveDate := strings.TrimSpace(input.ArchiveDate)
	if archiveDate == "" {
		return "", "", "", errors.New("archive date is required")
	}
	if err := validateArchiveDate(archiveDate); err != nil {
		return "", "", "", err
	}

	return projectRoot, changeID, archiveDate, nil
}

func validateArchiveChangeID(changeID string) error {
	if changeID == "." || changeID == ".." {
		return errors.New("change id must be a safe single path segment")
	}
	if strings.HasPrefix(changeID, "-") ||
		strings.Contains(changeID, "/") ||
		strings.Contains(changeID, "\\") ||
		strings.Contains(changeID, ":") {
		return errors.New("change id must be a safe single path segment")
	}
	return nil
}

func validateArchiveDate(archiveDate string) error {
	parsed, err := time.Parse("2006-01-02", archiveDate)
	if err != nil || parsed.Format("2006-01-02") != archiveDate {
		return errors.New("archive date must be formatted as YYYY-MM-DD")
	}
	return nil
}

func (useCase *ArchiveChange) requireOpenSpecProject(projectRoot string) error {
	projectFileExists, err := useCase.fileSystem.FileExists(projectRoot, openspecProjectFile)
	if err != nil {
		return fmt.Errorf("check file %s: %w", openspecProjectFile, err)
	}

	changesDirectoryExists, err := useCase.fileSystem.DirectoryExists(projectRoot, openspecChangesDirectory)
	if err != nil {
		return fmt.Errorf("check directory %s: %w", openspecChangesDirectory, err)
	}

	if projectFileExists && changesDirectoryExists {
		return nil
	}

	return errors.New("OpenSpec project structure is missing. Run specharbor init first.")
}

func (useCase *ArchiveChange) requireSourceChangeDirectory(projectRoot string, sourcePath string) error {
	sourceDirectoryExists, err := useCase.fileSystem.DirectoryExists(projectRoot, sourcePath)
	if err != nil {
		return fmt.Errorf("check directory %s: %w", sourcePath, err)
	}
	if sourceDirectoryExists {
		return nil
	}

	sourcePathExists, err := useCase.fileSystem.PathExists(projectRoot, sourcePath)
	if err != nil {
		return fmt.Errorf("check path %s: %w", sourcePath, err)
	}
	if sourcePathExists {
		return fmt.Errorf("source change path must be a directory: %s", sourcePath)
	}

	return fmt.Errorf("missing change directory: %s", sourcePath)
}

func (useCase *ArchiveChange) ensureArchiveDirectory(projectRoot string, relativePath string) error {
	pathExists, err := useCase.fileSystem.PathExists(projectRoot, relativePath)
	if err != nil {
		return fmt.Errorf("check path %s: %w", relativePath, err)
	}
	if !pathExists {
		if err := useCase.fileSystem.CreateDirectory(projectRoot, relativePath); err != nil {
			return fmt.Errorf("create directory %s: %w", relativePath, err)
		}
		return nil
	}

	directoryExists, err := useCase.fileSystem.DirectoryExists(projectRoot, relativePath)
	if err != nil {
		return fmt.Errorf("check directory %s: %w", relativePath, err)
	}
	if !directoryExists {
		return fmt.Errorf("archive path must be a directory: %s", relativePath)
	}
	return nil
}

func (useCase *ArchiveChange) requireArchiveTargetAvailable(projectRoot string, archivePath string) error {
	targetExists, err := useCase.fileSystem.PathExists(projectRoot, archivePath)
	if err != nil {
		return fmt.Errorf("check path %s: %w", archivePath, err)
	}
	if targetExists {
		return fmt.Errorf("archive target already exists: %s", archivePath)
	}
	return nil
}
