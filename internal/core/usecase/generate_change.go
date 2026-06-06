package usecase

import (
	"errors"
	"fmt"
	"strings"

	"github.com/guferreira1/spec-harbor/internal/core/domain"
	"github.com/guferreira1/spec-harbor/internal/core/ports"
)

type GenerateChangeInput struct {
	ProjectRoot string
	ChangeID    string
	Mode        domain.GenerationMode
}

type GenerateChange struct {
	fileSystem ports.GenerationFileSystem
	content    ports.BlankChangeContent
}

func NewGenerateChange(fileSystem ports.GenerationFileSystem, content ports.BlankChangeContent) *GenerateChange {
	return &GenerateChange{
		fileSystem: fileSystem,
		content:    content,
	}
}

func (useCase *GenerateChange) Execute(input GenerateChangeInput) (domain.GenerationResult, error) {
	if useCase == nil {
		return domain.GenerationResult{}, errors.New("generate change use case is required")
	}
	if useCase.fileSystem == nil {
		return domain.GenerationResult{}, errors.New("generation filesystem is required")
	}
	if useCase.content == nil {
		return domain.GenerationResult{}, errors.New("blank change content is required")
	}

	projectRoot := strings.TrimSpace(input.ProjectRoot)
	if projectRoot == "" {
		return domain.GenerationResult{}, errors.New("project root is required")
	}

	changeID := strings.TrimSpace(input.ChangeID)
	if changeID == "" {
		return domain.GenerationResult{}, errors.New("change id is required")
	}
	if input.Mode != domain.BlankMode {
		return domain.GenerationResult{}, fmt.Errorf("unsupported generation mode: %s", input.Mode)
	}
	if err := validateGenerationChangeID(changeID); err != nil {
		return domain.GenerationResult{}, err
	}

	changePath := openspecChangesDirectory + "/" + changeID
	if err := useCase.requireOpenSpecProject(projectRoot); err != nil {
		return domain.GenerationResult{}, err
	}

	directoryCreated, err := useCase.ensureChangeDirectory(projectRoot, changePath)
	if err != nil {
		return domain.GenerationResult{}, err
	}

	createdFiles, skippedExistingFiles, err := useCase.writeBlankFiles(projectRoot, changePath)
	if err != nil {
		return domain.GenerationResult{}, err
	}

	return domain.NewGenerationResult(
		changeID,
		domain.BlankMode,
		changePath,
		directoryCreated,
		createdFiles,
		skippedExistingFiles,
	), nil
}

func validateGenerationChangeID(changeID string) error {
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

func (useCase *GenerateChange) requireOpenSpecProject(projectRoot string) error {
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

func (useCase *GenerateChange) ensureChangeDirectory(projectRoot string, changePath string) (bool, error) {
	exists, err := useCase.fileSystem.DirectoryExists(projectRoot, changePath)
	if err != nil {
		return false, fmt.Errorf("check directory %s: %w", changePath, err)
	}
	if exists {
		return false, nil
	}

	if err := useCase.fileSystem.CreateDirectory(projectRoot, changePath); err != nil {
		return false, fmt.Errorf("create directory %s: %w", changePath, err)
	}
	return true, nil
}

func (useCase *GenerateChange) writeBlankFiles(projectRoot string, changePath string) ([]string, []string, error) {
	var createdFiles []string
	var skippedExistingFiles []string

	for _, requiredFile := range domain.RequiredOpenSpecChangeFiles() {
		contents, err := useCase.content.ContentFor(requiredFile)
		if err != nil {
			return nil, nil, fmt.Errorf("load blank content for %s: %w", requiredFile, err)
		}

		created, err := useCase.fileSystem.WriteFileIfAbsent(projectRoot, changePath+"/"+requiredFile, contents)
		if err != nil {
			return nil, nil, fmt.Errorf("write file %s/%s: %w", changePath, requiredFile, err)
		}
		if created {
			createdFiles = append(createdFiles, requiredFile)
			continue
		}

		skippedExistingFiles = append(skippedExistingFiles, requiredFile)
	}

	return createdFiles, skippedExistingFiles, nil
}
