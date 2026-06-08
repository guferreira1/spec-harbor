package usecase

import (
	"errors"
	"fmt"
	"strings"

	"github.com/guferreira1/spec-harbor/internal/core/domain"
	"github.com/guferreira1/spec-harbor/internal/core/ports"
)

type GenerateChangeInput struct {
	ProjectRoot  string
	ChangeID     string
	Mode         domain.GenerationMode
	TemplateName string
}

type GenerateChange struct {
	fileSystem      ports.GenerationFileSystem
	blankContent    ports.BlankChangeContent
	templateContent ports.TemplateChangeContent
}

func NewGenerateChange(fileSystem ports.GenerationFileSystem, content ports.BlankChangeContent) *GenerateChange {
	return &GenerateChange{
		fileSystem:   fileSystem,
		blankContent: content,
	}
}

func NewGenerateChangeWithTemplateContent(
	fileSystem ports.GenerationFileSystem,
	blankContent ports.BlankChangeContent,
	templateContent ports.TemplateChangeContent,
) *GenerateChange {
	return &GenerateChange{
		fileSystem:      fileSystem,
		blankContent:    blankContent,
		templateContent: templateContent,
	}
}

func (useCase *GenerateChange) Execute(input GenerateChangeInput) (domain.GenerationResult, error) {
	if useCase == nil {
		return domain.GenerationResult{}, errors.New("generate change use case is required")
	}
	if useCase.fileSystem == nil {
		return domain.GenerationResult{}, errors.New("generation filesystem is required")
	}
	if input.Mode == domain.BlankMode && useCase.blankContent == nil {
		return domain.GenerationResult{}, errors.New("blank change content is required")
	}
	if input.Mode == domain.TemplateMode && useCase.templateContent == nil {
		return domain.GenerationResult{}, errors.New("template change content is required")
	}

	projectRoot := strings.TrimSpace(input.ProjectRoot)
	if projectRoot == "" {
		return domain.GenerationResult{}, errors.New("project root is required")
	}

	changeID := strings.TrimSpace(input.ChangeID)
	if changeID == "" {
		return domain.GenerationResult{}, errors.New("change id is required")
	}
	if err := validateGenerationChangeID(changeID); err != nil {
		return domain.GenerationResult{}, err
	}

	var templateName domain.TemplateName
	switch input.Mode {
	case domain.BlankMode:
	case domain.TemplateMode:
		var err error
		templateName, err = domain.ParseTemplateName(input.TemplateName)
		if err != nil {
			return domain.GenerationResult{}, err
		}
	default:
		return domain.GenerationResult{}, fmt.Errorf("unsupported generation mode: %s", input.Mode)
	}

	changePath := openspecChangesDirectory + "/" + changeID
	if err := useCase.requireOpenSpecProject(projectRoot); err != nil {
		return domain.GenerationResult{}, err
	}

	directoryCreated, err := useCase.ensureChangeDirectory(projectRoot, changePath)
	if err != nil {
		return domain.GenerationResult{}, err
	}

	createdFiles, skippedExistingFiles, err := useCase.writeChangeFiles(projectRoot, changePath, input.Mode, templateName)
	if err != nil {
		return domain.GenerationResult{}, err
	}

	if input.Mode == domain.TemplateMode {
		return domain.NewTemplateGenerationResult(
			changeID,
			templateName,
			changePath,
			directoryCreated,
			createdFiles,
			skippedExistingFiles,
		), nil
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

func (useCase *GenerateChange) writeChangeFiles(
	projectRoot string,
	changePath string,
	mode domain.GenerationMode,
	templateName domain.TemplateName,
) ([]string, []string, error) {
	var createdFiles []string
	var skippedExistingFiles []string

	for _, requiredFile := range domain.RequiredOpenSpecChangeFiles() {
		contents, err := useCase.contentFor(mode, templateName, requiredFile)
		if err != nil {
			return nil, nil, err
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

func (useCase *GenerateChange) contentFor(
	mode domain.GenerationMode,
	templateName domain.TemplateName,
	requiredFile string,
) (string, error) {
	if mode == domain.TemplateMode {
		contents, err := useCase.templateContent.ContentFor(templateName, requiredFile)
		if err != nil {
			return "", fmt.Errorf("load %s template content for %s: %w", templateName, requiredFile, err)
		}
		return contents, nil
	}

	contents, err := useCase.blankContent.ContentFor(requiredFile)
	if err != nil {
		return "", fmt.Errorf("load blank content for %s: %w", requiredFile, err)
	}
	return contents, nil
}
