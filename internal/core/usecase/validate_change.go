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

	parsedChangeID, err := domain.NewChangeID(input.ChangeID)
	if err != nil {
		return domain.ValidationResult{}, err
	}
	changeID := parsedChangeID.String()

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

	findings, loadedFiles, err := useCase.loadRequiredFiles(projectRoot, checkedPath, requiredFiles)
	if err != nil {
		return domain.ValidationResult{}, err
	}

	findings = append(findings, domain.ValidateChangeFileContents(loadedFiles)...)

	return domain.NewValidationResult(changeID, checkedPath, requiredFiles, findings), nil
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

func (useCase *ValidateChange) loadRequiredFiles(projectRoot string, checkedPath string, requiredFiles []string) ([]domain.ValidationFinding, []domain.ChangeFileContent, error) {
	var findings []domain.ValidationFinding
	var loadedFiles []domain.ChangeFileContent

	for _, requiredFile := range requiredFiles {
		relativePath := checkedPath + "/" + requiredFile
		exists, err := useCase.fileSystem.FileExists(projectRoot, relativePath)
		if err != nil {
			return nil, nil, fmt.Errorf("check file %s: %w", relativePath, err)
		}
		if !exists {
			findings = append(findings, requiredFileMissingFinding(relativePath, requiredFile))
			continue
		}

		content, err := useCase.fileSystem.ReadFile(projectRoot, relativePath)
		if err != nil {
			return nil, nil, fmt.Errorf("read file %s: %w", relativePath, err)
		}
		loadedFiles = append(loadedFiles, domain.ChangeFileContent{
			FileName:     requiredFile,
			RelativePath: relativePath,
			Content:      content,
		})
	}

	return findings, loadedFiles, nil
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
