package usecase

import (
	"errors"
	"fmt"
	"strings"

	"github.com/guferreira1/spec-harbor/internal/core/domain"
	"github.com/guferreira1/spec-harbor/internal/core/ports"
)

const openspecProjectFile = "openspec/project.md"
const openspecChangesDirectory = "openspec/changes"

type ValidateChangeInput struct {
	ProjectRoot string
	ChangeID    string
}

type ValidateChange struct {
	fileSystem ports.ValidationFileSystem
}

func NewValidateChange(fileSystem ports.ValidationFileSystem) *ValidateChange {
	return &ValidateChange{fileSystem: fileSystem}
}

func (useCase *ValidateChange) Execute(input ValidateChangeInput) (domain.ValidationResult, error) {
	if useCase == nil {
		return domain.ValidationResult{}, errors.New("validate change use case is required")
	}
	if useCase.fileSystem == nil {
		return domain.ValidationResult{}, errors.New("validation filesystem is required")
	}

	projectRoot := strings.TrimSpace(input.ProjectRoot)
	if projectRoot == "" {
		return domain.ValidationResult{}, errors.New("project root is required")
	}

	changeID := strings.TrimSpace(input.ChangeID)
	if changeID == "" {
		return domain.ValidationResult{}, errors.New("change id is required")
	}
	if err := validateChangeID(changeID); err != nil {
		return domain.ValidationResult{}, err
	}

	requiredFiles := domain.RequiredOpenSpecChangeFiles()
	checkedPath := openspecChangesDirectory + "/" + changeID

	if result, available, err := useCase.validateProjectStructure(projectRoot, changeID, checkedPath, requiredFiles); err != nil {
		return domain.ValidationResult{}, err
	} else if !available {
		return result, nil
	}

	changeDirectoryExists, err := useCase.fileSystem.DirectoryExists(projectRoot, checkedPath)
	if err != nil {
		return domain.ValidationResult{}, fmt.Errorf("check directory %s: %w", checkedPath, err)
	}
	if !changeDirectoryExists {
		return domain.NewValidationResult(changeID, checkedPath, requiredFiles, []domain.ValidationFinding{
			changeDirectoryMissingFinding(checkedPath),
		}), nil
	}

	findings, err := useCase.validateRequiredFiles(projectRoot, checkedPath, requiredFiles)
	if err != nil {
		return domain.ValidationResult{}, err
	}

	return domain.NewValidationResult(changeID, checkedPath, requiredFiles, findings), nil
}

func validateChangeID(changeID string) error {
	if changeID == "." || changeID == ".." || strings.Contains(changeID, "/") || strings.Contains(changeID, "\\") {
		return errors.New("change id must be a single path segment")
	}
	return nil
}

func (useCase *ValidateChange) validateProjectStructure(projectRoot string, changeID string, checkedPath string, requiredFiles []string) (domain.ValidationResult, bool, error) {
	projectFileExists, err := useCase.fileSystem.FileExists(projectRoot, openspecProjectFile)
	if err != nil {
		return domain.ValidationResult{}, false, fmt.Errorf("check file %s: %w", openspecProjectFile, err)
	}

	changesDirectoryExists, err := useCase.fileSystem.DirectoryExists(projectRoot, openspecChangesDirectory)
	if err != nil {
		return domain.ValidationResult{}, false, fmt.Errorf("check directory %s: %w", openspecChangesDirectory, err)
	}

	if projectFileExists && changesDirectoryExists {
		return domain.ValidationResult{}, true, nil
	}

	return domain.NewValidationResult(changeID, checkedPath, requiredFiles, []domain.ValidationFinding{
		projectRootUnavailableFinding(),
	}), false, nil
}

func (useCase *ValidateChange) validateRequiredFiles(projectRoot string, checkedPath string, requiredFiles []string) ([]domain.ValidationFinding, error) {
	var findings []domain.ValidationFinding

	for _, requiredFile := range requiredFiles {
		relativePath := checkedPath + "/" + requiredFile
		exists, err := useCase.fileSystem.FileExists(projectRoot, relativePath)
		if err != nil {
			return nil, fmt.Errorf("check file %s: %w", relativePath, err)
		}
		if exists {
			continue
		}

		findings = append(findings, requiredFileMissingFinding(relativePath, requiredFile))
	}

	return findings, nil
}

func projectRootUnavailableFinding() domain.ValidationFinding {
	return domain.ValidationFinding{
		Severity:     domain.ValidationFindingSeverityError,
		Code:         domain.ValidationFindingCodeProjectRootUnavailable,
		Message:      "OpenSpec project structure is unavailable.",
		RelativePath: "openspec",
	}
}

func changeDirectoryMissingFinding(checkedPath string) domain.ValidationFinding {
	return domain.ValidationFinding{
		Severity:     domain.ValidationFindingSeverityError,
		Code:         domain.ValidationFindingCodeChangeDirectoryMissing,
		Message:      "Missing change directory: " + checkedPath,
		RelativePath: checkedPath,
		Subject:      checkedPath,
	}
}

func requiredFileMissingFinding(relativePath string, fileName string) domain.ValidationFinding {
	return domain.ValidationFinding{
		Severity:     domain.ValidationFindingSeverityError,
		Code:         domain.ValidationFindingCodeRequiredFileMissing,
		Message:      "Missing required file: " + fileName,
		RelativePath: relativePath,
		Subject:      fileName,
	}
}
