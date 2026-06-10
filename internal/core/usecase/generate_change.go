package usecase

import (
	"errors"
	"fmt"
	"strings"

	"github.com/guferreira1/spec-harbor/internal/core/domain"
	"github.com/guferreira1/spec-harbor/internal/core/ports"
)

const customTemplatesDirectory = ".specharbor/templates"

type GenerateChangeInput struct {
	ProjectRoot        string
	ChangeID           string
	Mode               domain.GenerationMode
	TemplateName       string
	TemplateSource     domain.TemplateSource
	CustomTemplateName string
	GuidedType         string
	Title              string
	Summary            string
}

type GenerateChange struct {
	fileSystem          ports.GenerationFileSystem
	blankContent        ports.BlankChangeContent
	templateContent     ports.TemplateChangeContent
	guidedContent       ports.GuidedChangeContent
	customTemplateFiles ports.CustomTemplateFileSystem
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

func NewGenerateChangeWithContent(
	fileSystem ports.GenerationFileSystem,
	blankContent ports.BlankChangeContent,
	templateContent ports.TemplateChangeContent,
	guidedContent ports.GuidedChangeContent,
) *GenerateChange {
	return &GenerateChange{
		fileSystem:      fileSystem,
		blankContent:    blankContent,
		templateContent: templateContent,
		guidedContent:   guidedContent,
	}
}

func NewGenerateChangeWithCustomTemplates(
	fileSystem ports.GenerationFileSystem,
	blankContent ports.BlankChangeContent,
	templateContent ports.TemplateChangeContent,
	guidedContent ports.GuidedChangeContent,
	customTemplateFiles ports.CustomTemplateFileSystem,
) *GenerateChange {
	return &GenerateChange{
		fileSystem:          fileSystem,
		blankContent:        blankContent,
		templateContent:     templateContent,
		guidedContent:       guidedContent,
		customTemplateFiles: customTemplateFiles,
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
	customTemplateRequested, err := isCustomTemplateRequest(input)
	if err != nil {
		return domain.GenerationResult{}, err
	}
	if input.Mode == domain.TemplateMode && !customTemplateRequested && useCase.templateContent == nil {
		return domain.GenerationResult{}, errors.New("template change content is required")
	}
	if customTemplateRequested && useCase.customTemplateFiles == nil {
		return domain.GenerationResult{}, errors.New("custom template filesystem is required")
	}
	if input.Mode == domain.GuidedMode && useCase.guidedContent == nil {
		return domain.GenerationResult{}, errors.New("guided change content is required")
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
	var customTemplateName domain.CustomTemplateName
	var guidedType domain.GuidedType
	var guidedTitle string
	var guidedSummary string
	switch input.Mode {
	case domain.BlankMode:
	case domain.TemplateMode:
		var err error
		if customTemplateRequested {
			customTemplateName, err = domain.NewCustomTemplateName(input.CustomTemplateName)
		} else {
			templateName, err = domain.ParseTemplateName(input.TemplateName)
		}
		if err != nil {
			return domain.GenerationResult{}, err
		}
	case domain.GuidedMode:
		var err error
		guidedType, err = domain.ParseGuidedType(input.GuidedType)
		if err != nil {
			return domain.GenerationResult{}, err
		}
		guidedTitle = strings.TrimSpace(input.Title)
		if guidedTitle == "" {
			return domain.GenerationResult{}, errors.New("guided title is required")
		}
		guidedSummary = strings.TrimSpace(input.Summary)
		if guidedSummary == "" {
			return domain.GenerationResult{}, errors.New("guided summary is required")
		}
	default:
		return domain.GenerationResult{}, fmt.Errorf("unsupported generation mode: %s", input.Mode)
	}

	changePath := openspecChangesDirectory + "/" + changeID
	if err := useCase.requireOpenSpecProject(projectRoot); err != nil {
		return domain.GenerationResult{}, err
	}

	var templatePath string
	var renderedContents map[string]string
	if customTemplateRequested {
		templatePath = customTemplatesDirectory + "/" + customTemplateName.String()
		customTemplate, err := useCase.loadCustomTemplate(projectRoot, customTemplateName, templatePath)
		if err != nil {
			return domain.GenerationResult{}, err
		}
		renderedContents = customTemplate.Render(changeID, input.Title, input.Summary)
	}

	directoryCreated, err := useCase.ensureChangeDirectory(projectRoot, changePath)
	if err != nil {
		return domain.GenerationResult{}, err
	}

	contentRequest := generationContentRequest{
		mode:             input.Mode,
		templateName:     templateName,
		guidedType:       guidedType,
		guidedTitle:      guidedTitle,
		guidedSummary:    guidedSummary,
		renderedContents: renderedContents,
	}
	createdFiles, skippedExistingFiles, err := useCase.writeChangeFiles(projectRoot, changePath, contentRequest)
	if err != nil {
		return domain.GenerationResult{}, err
	}

	if input.Mode == domain.TemplateMode {
		if customTemplateRequested {
			return domain.NewCustomTemplateGenerationResult(
				changeID,
				customTemplateName,
				templatePath,
				changePath,
				directoryCreated,
				createdFiles,
				skippedExistingFiles,
			), nil
		}
		return domain.NewTemplateGenerationResult(
			changeID,
			templateName,
			changePath,
			directoryCreated,
			createdFiles,
			skippedExistingFiles,
		), nil
	}

	if input.Mode == domain.GuidedMode {
		return domain.NewGuidedGenerationResult(
			changeID,
			guidedType,
			guidedTitle,
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

func isCustomTemplateRequest(input GenerateChangeInput) (bool, error) {
	switch input.TemplateSource {
	case "", domain.BuiltInTemplateSource:
		return false, nil
	case domain.CustomTemplateSource:
		if input.Mode != domain.TemplateMode {
			return false, fmt.Errorf("custom templates require template generation mode, got: %s", input.Mode)
		}
		return true, nil
	default:
		return false, fmt.Errorf("unsupported template source: %s", input.TemplateSource)
	}
}

type generationContentRequest struct {
	mode             domain.GenerationMode
	templateName     domain.TemplateName
	guidedType       domain.GuidedType
	guidedTitle      string
	guidedSummary    string
	renderedContents map[string]string
}

func (useCase *GenerateChange) loadCustomTemplate(
	projectRoot string,
	customTemplateName domain.CustomTemplateName,
	templatePath string,
) (domain.CustomTemplate, error) {
	templateDirectoryExists, err := useCase.customTemplateFiles.DirectoryExists(projectRoot, templatePath)
	if err != nil {
		return domain.CustomTemplate{}, fmt.Errorf("check directory %s: %w", templatePath, err)
	}
	if !templateDirectoryExists {
		return domain.CustomTemplate{}, fmt.Errorf(
			"unknown custom template: %s. Expected directory: %s",
			customTemplateName,
			templatePath,
		)
	}

	var missingFiles []string
	templateFiles := make(map[string]string)
	for _, requiredFile := range domain.AllowedCustomTemplateFiles() {
		filePath := templatePath + "/" + requiredFile
		fileExists, err := useCase.customTemplateFiles.FileExists(projectRoot, filePath)
		if err != nil {
			return domain.CustomTemplate{}, fmt.Errorf("check file %s: %w", filePath, err)
		}
		if !fileExists {
			missingFiles = append(missingFiles, requiredFile)
			continue
		}

		contents, err := useCase.customTemplateFiles.ReadFile(projectRoot, filePath)
		if err != nil {
			return domain.CustomTemplate{}, fmt.Errorf("read file %s: %w", filePath, err)
		}
		templateFiles[requiredFile] = contents
	}
	if len(missingFiles) > 0 {
		return domain.CustomTemplate{}, fmt.Errorf(
			"custom template %s is missing required files: %s",
			customTemplateName,
			strings.Join(missingFiles, ", "),
		)
	}

	return domain.NewCustomTemplate(customTemplateName, templateFiles)
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
	contentRequest generationContentRequest,
) ([]string, []string, error) {
	var createdFiles []string
	var skippedExistingFiles []string

	for _, requiredFile := range domain.RequiredOpenSpecChangeFiles() {
		contents, err := useCase.contentFor(contentRequest, requiredFile)
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
	contentRequest generationContentRequest,
	requiredFile string,
) (string, error) {
	if contentRequest.renderedContents != nil {
		contents, exists := contentRequest.renderedContents[requiredFile]
		if !exists {
			return "", fmt.Errorf("missing rendered custom template content for %s", requiredFile)
		}
		return contents, nil
	}

	if contentRequest.mode == domain.TemplateMode {
		contents, err := useCase.templateContent.ContentFor(contentRequest.templateName, requiredFile)
		if err != nil {
			return "", fmt.Errorf("load %s template content for %s: %w", contentRequest.templateName, requiredFile, err)
		}
		return contents, nil
	}

	if contentRequest.mode == domain.GuidedMode {
		contents, err := useCase.guidedContent.ContentFor(
			contentRequest.guidedType,
			contentRequest.guidedTitle,
			contentRequest.guidedSummary,
			requiredFile,
		)
		if err != nil {
			return "", fmt.Errorf("load %s guided content for %s: %w", contentRequest.guidedType, requiredFile, err)
		}
		return contents, nil
	}

	contents, err := useCase.blankContent.ContentFor(requiredFile)
	if err != nil {
		return "", fmt.Errorf("load blank content for %s: %w", requiredFile, err)
	}
	return contents, nil
}
