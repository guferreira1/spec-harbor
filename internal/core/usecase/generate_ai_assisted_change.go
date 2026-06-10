package usecase

import (
	"errors"
	"fmt"
	"strings"

	"github.com/guferreira1/spec-harbor/internal/core/domain"
	"github.com/guferreira1/spec-harbor/internal/core/ports"
)

type GenerateAIAssistedChangeInput struct {
	ProjectRoot string
	ChangeID    string
	SourcePath  string
	Overwrite   bool
}

type GenerateAIAssistedChange struct {
	fileSystem ports.AIAssistedGenerationFileSystem
	validator  aiAssistedChangeValidator
}

type aiAssistedChangeValidator interface {
	Execute(input ValidateChangeInput) (domain.ValidationResult, error)
}

func NewGenerateAIAssistedChange(
	fileSystem ports.AIAssistedGenerationFileSystem,
	validator aiAssistedChangeValidator,
) *GenerateAIAssistedChange {
	return &GenerateAIAssistedChange{
		fileSystem: fileSystem,
		validator:  validator,
	}
}

type AIAssistedParseFailure struct {
	ChangeID    string
	SourcePath  string
	ParseResult domain.AIOutputParseResult
}

func (failure *AIAssistedParseFailure) Error() string {
	return "AI-assisted source parse failed"
}

func (useCase *GenerateAIAssistedChange) Execute(input GenerateAIAssistedChangeInput) (domain.AIAssistedGenerationResult, error) {
	if useCase == nil {
		return domain.AIAssistedGenerationResult{}, errors.New("AI-assisted generation use case is required")
	}
	if useCase.fileSystem == nil {
		return domain.AIAssistedGenerationResult{}, errors.New("AI-assisted generation filesystem is required")
	}
	if useCase.validator == nil {
		return domain.AIAssistedGenerationResult{}, errors.New("AI-assisted generation validator is required")
	}

	projectRoot := strings.TrimSpace(input.ProjectRoot)
	if projectRoot == "" {
		return domain.AIAssistedGenerationResult{}, errors.New("project root is required")
	}

	parsedChangeID, err := domain.NewChangeID(input.ChangeID)
	if err != nil {
		return domain.AIAssistedGenerationResult{}, err
	}
	changeID := parsedChangeID.String()

	sourcePath := strings.TrimSpace(input.SourcePath)
	if sourcePath == "" {
		return domain.AIAssistedGenerationResult{}, errors.New("source file is required")
	}

	if err := useCase.requireOpenSpecProject(projectRoot); err != nil {
		return domain.AIAssistedGenerationResult{}, err
	}

	source, err := useCase.fileSystem.ReadSourceFile(sourcePath)
	if err != nil {
		return domain.AIAssistedGenerationResult{}, fmt.Errorf("read AI-assisted source file %s: %w", sourcePath, err)
	}

	parseResult := domain.ParseAIOutputBlocks(source)
	if parseResult.HasErrors() {
		return domain.AIAssistedGenerationResult{}, &AIAssistedParseFailure{
			ChangeID:    changeID,
			SourcePath:  sourcePath,
			ParseResult: parseResult,
		}
	}

	changePath := openspecChangesDirectory + "/" + changeID
	directoryCreated, err := useCase.ensureTargetDirectory(projectRoot, changePath)
	if err != nil {
		return domain.AIAssistedGenerationResult{}, err
	}

	plan, err := useCase.buildWritePlan(projectRoot, changePath, parseResult, input.Overwrite)
	if err != nil {
		return domain.AIAssistedGenerationResult{}, err
	}

	generatedFiles, skippedFiles, overwrittenFiles, err := useCase.executeWritePlan(projectRoot, plan)
	if err != nil {
		return domain.AIAssistedGenerationResult{}, err
	}

	validationResult, err := useCase.validator.Execute(ValidateChangeInput{
		ProjectRoot: projectRoot,
		ChangeID:    changeID,
	})
	if err != nil {
		return domain.AIAssistedGenerationResult{}, fmt.Errorf("validate generated change %s: %w", changeID, err)
	}

	return domain.NewAIAssistedGenerationResult(
		changeID,
		sourcePath,
		changePath,
		directoryCreated,
		input.Overwrite,
		generatedFiles,
		skippedFiles,
		overwrittenFiles,
		validationResult,
	), nil
}

func (useCase *GenerateAIAssistedChange) requireOpenSpecProject(projectRoot string) error {
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

func (useCase *GenerateAIAssistedChange) ensureTargetDirectory(projectRoot string, changePath string) (bool, error) {
	pathExists, err := useCase.fileSystem.PathExists(projectRoot, changePath)
	if err != nil {
		return false, fmt.Errorf("check path %s: %w", changePath, err)
	}
	if pathExists {
		directoryExists, err := useCase.fileSystem.DirectoryExists(projectRoot, changePath)
		if err != nil {
			return false, fmt.Errorf("check directory %s: %w", changePath, err)
		}
		if !directoryExists {
			return false, fmt.Errorf("target path exists and is not a directory: %s", changePath)
		}
		return false, nil
	}

	if err := useCase.fileSystem.CreateDirectory(projectRoot, changePath); err != nil {
		return false, fmt.Errorf("create directory %s: %w", changePath, err)
	}
	return true, nil
}

type aiAssistedWriteAction string

const (
	aiAssistedWriteCreate    aiAssistedWriteAction = "create"
	aiAssistedWriteSkip      aiAssistedWriteAction = "skip"
	aiAssistedWriteOverwrite aiAssistedWriteAction = "overwrite"
)

type aiAssistedWritePlanItem struct {
	fileName     string
	relativePath string
	content      string
	action       aiAssistedWriteAction
}

func (useCase *GenerateAIAssistedChange) buildWritePlan(
	projectRoot string,
	changePath string,
	parseResult domain.AIOutputParseResult,
	overwrite bool,
) ([]aiAssistedWritePlanItem, error) {
	plan := make([]aiAssistedWritePlanItem, 0, len(domain.RequiredAIGeneratedFileNames()))

	for _, fileName := range domain.RequiredAIGeneratedFileNames() {
		content, exists := parseResult.ContentFor(fileName)
		if !exists {
			return nil, fmt.Errorf("parsed AI-assisted output is missing required file %s", fileName)
		}

		relativePath := changePath + "/" + fileName
		if err := useCase.fileSystem.EnsureSafeWriteTarget(projectRoot, relativePath); err != nil {
			return nil, fmt.Errorf("unsafe generated OpenSpec file target %s: %w", relativePath, err)
		}

		pathExists, err := useCase.fileSystem.PathExists(projectRoot, relativePath)
		if err != nil {
			return nil, fmt.Errorf("check path %s: %w", relativePath, err)
		}
		if !pathExists {
			plan = append(plan, aiAssistedWritePlanItem{
				fileName:     fileName,
				relativePath: relativePath,
				content:      content,
				action:       aiAssistedWriteCreate,
			})
			continue
		}

		fileExists, err := useCase.fileSystem.FileExists(projectRoot, relativePath)
		if err != nil {
			return nil, fmt.Errorf("check file %s: %w", relativePath, err)
		}
		if !fileExists {
			return nil, fmt.Errorf("target file path exists and is not a file: %s", relativePath)
		}
		if overwrite {
			plan = append(plan, aiAssistedWritePlanItem{
				fileName:     fileName,
				relativePath: relativePath,
				content:      content,
				action:       aiAssistedWriteOverwrite,
			})
			continue
		}

		plan = append(plan, aiAssistedWritePlanItem{
			fileName:     fileName,
			relativePath: relativePath,
			content:      content,
			action:       aiAssistedWriteSkip,
		})
	}

	return plan, nil
}

func (useCase *GenerateAIAssistedChange) executeWritePlan(
	projectRoot string,
	plan []aiAssistedWritePlanItem,
) ([]string, []string, []string, error) {
	var generatedFiles []string
	var skippedFiles []string
	var overwrittenFiles []string

	for _, item := range plan {
		if err := useCase.fileSystem.EnsureSafeWriteTarget(projectRoot, item.relativePath); err != nil {
			return nil, nil, nil, fmt.Errorf("unsafe generated OpenSpec file target %s: %w", item.relativePath, err)
		}

		switch item.action {
		case aiAssistedWriteSkip:
			skippedFiles = append(skippedFiles, item.fileName)
		case aiAssistedWriteCreate:
			created, err := useCase.fileSystem.WriteFileIfAbsent(projectRoot, item.relativePath, item.content)
			if err != nil {
				return nil, nil, nil, fmt.Errorf("write file %s: %w", item.relativePath, err)
			}
			if created {
				generatedFiles = append(generatedFiles, item.fileName)
				continue
			}
			skippedFiles = append(skippedFiles, item.fileName)
		case aiAssistedWriteOverwrite:
			if err := useCase.fileSystem.WriteFile(projectRoot, item.relativePath, item.content); err != nil {
				return nil, nil, nil, fmt.Errorf("overwrite file %s: %w", item.relativePath, err)
			}
			overwrittenFiles = append(overwrittenFiles, item.fileName)
		default:
			return nil, nil, nil, fmt.Errorf("unsupported AI-assisted write action: %s", item.action)
		}
	}

	return generatedFiles, skippedFiles, overwrittenFiles, nil
}
