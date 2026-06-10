package usecase

import (
	"errors"
	"fmt"
	"strings"

	"github.com/guferreira1/spec-harbor/internal/core/domain"
	"github.com/guferreira1/spec-harbor/internal/core/ports"
)

type GenerateHybridChangeInput struct {
	ProjectRoot         string
	ChangeID            string
	TemplateName        string
	CustomTemplateName  string
	ConfigTemplateAlias string
	Title               string
	Summary             string
	Type                string
}

type hybridChangeValidator interface {
	Execute(input ValidateChangeInput) (domain.ValidationResult, error)
}

type GenerateHybridChange struct {
	fileSystem          ports.GenerationFileSystem
	templateContent     ports.TemplateChangeContent
	customTemplateFiles ports.CustomTemplateFileSystem
	configFileSystem    ports.ConfigFileSystem
	configParser        ports.ConfigParser
	remoteFetcher       ports.RemoteTemplateFetcher
	remoteBundleReader  ports.RemoteTemplateBundleReader
	validator           hybridChangeValidator
}

func NewGenerateHybridChange(
	fileSystem ports.GenerationFileSystem,
	templateContent ports.TemplateChangeContent,
	customTemplateFiles ports.CustomTemplateFileSystem,
	configFileSystem ports.ConfigFileSystem,
	configParser ports.ConfigParser,
	remoteFetcher ports.RemoteTemplateFetcher,
	remoteBundleReader ports.RemoteTemplateBundleReader,
	validator hybridChangeValidator,
) *GenerateHybridChange {
	return &GenerateHybridChange{
		fileSystem:          fileSystem,
		templateContent:     templateContent,
		customTemplateFiles: customTemplateFiles,
		configFileSystem:    configFileSystem,
		configParser:        configParser,
		remoteFetcher:       remoteFetcher,
		remoteBundleReader:  remoteBundleReader,
		validator:           validator,
	}
}

func (useCase *GenerateHybridChange) Execute(input GenerateHybridChangeInput) (domain.HybridGenerationResult, error) {
	if useCase == nil {
		return domain.HybridGenerationResult{}, errors.New("hybrid generate change use case is required")
	}
	if useCase.fileSystem == nil {
		return domain.HybridGenerationResult{}, errors.New("generation filesystem is required")
	}
	if useCase.validator == nil {
		return domain.HybridGenerationResult{}, errors.New("hybrid validator is required")
	}

	projectRoot := strings.TrimSpace(input.ProjectRoot)
	if projectRoot == "" {
		return domain.HybridGenerationResult{}, errors.New("project root is required")
	}

	changeID := strings.TrimSpace(input.ChangeID)
	if changeID == "" {
		return domain.HybridGenerationResult{}, errors.New("change id is required")
	}
	if err := validateGenerationChangeID(changeID); err != nil {
		return domain.HybridGenerationResult{}, err
	}

	selection, err := domain.NewHybridSourceSelection(input.TemplateName, input.CustomTemplateName, input.ConfigTemplateAlias)
	if err != nil {
		return domain.HybridGenerationResult{}, fmt.Errorf("hybrid source selection error: %w", err)
	}

	metadata, err := domain.NewHybridMetadata(input.Title, input.Summary, input.Type)
	if err != nil {
		return domain.HybridGenerationResult{}, fmt.Errorf("hybrid metadata error: %w", err)
	}

	if err := useCase.requireOpenSpecProject(projectRoot); err != nil {
		return domain.HybridGenerationResult{}, err
	}

	resolved, err := useCase.resolveHybridSource(projectRoot, selection)
	if err != nil {
		return domain.HybridGenerationResult{}, err
	}

	if resolved.builtInTypeSource {
		metadata, err = metadata.WithBuiltInEffectiveType(resolved.builtInTemplate)
		if err != nil {
			return domain.HybridGenerationResult{}, fmt.Errorf("hybrid type mismatch: %w", err)
		}
	} else {
		metadata = metadata.WithoutDerivedType()
	}

	renderedFiles := metadata.RenderFiles(resolved.files, changeID)
	if err := validateHybridRenderedFiles(renderedFiles); err != nil {
		return domain.HybridGenerationResult{}, fmt.Errorf("hybrid source resolution error: %w", err)
	}

	changePath := openspecChangesDirectory + "/" + changeID
	directoryCreated, err := useCase.ensureChangeDirectory(projectRoot, changePath)
	if err != nil {
		return domain.HybridGenerationResult{}, fmt.Errorf("hybrid write error: %w", err)
	}

	createdFiles, skippedExistingFiles, err := useCase.writeRenderedHybridFiles(projectRoot, changePath, renderedFiles)
	if err != nil {
		return domain.HybridGenerationResult{}, fmt.Errorf("hybrid write error: %w", err)
	}

	validationResult, err := useCase.validator.Execute(ValidateChangeInput{
		ProjectRoot: projectRoot,
		ChangeID:    changeID,
	})
	if err != nil {
		return domain.HybridGenerationResult{}, fmt.Errorf("hybrid validation error: %w", err)
	}

	return domain.NewHybridGenerationResult(
		changeID,
		selection,
		resolved.kind,
		resolved.name,
		metadata,
		changePath,
		directoryCreated,
		createdFiles,
		skippedExistingFiles,
		validationResult,
		resolved.remoteFacts,
	), nil
}

type hybridResolvedSource struct {
	kind              domain.HybridResolvedSourceKind
	name              string
	files             map[string]string
	builtInTypeSource bool
	builtInTemplate   domain.TemplateName
	remoteFacts       domain.HybridRemoteFacts
}

func (useCase *GenerateHybridChange) resolveHybridSource(projectRoot string, selection domain.HybridSourceSelection) (hybridResolvedSource, error) {
	switch selection.Kind() {
	case domain.HybridSelectedSourceBuiltIn:
		templateName, _ := selection.TemplateName()
		return useCase.resolveHybridBuiltInSource(templateName)
	case domain.HybridSelectedSourceCustom:
		customTemplateName, _ := selection.CustomTemplateName()
		return useCase.resolveHybridCustomSource(projectRoot, customTemplateName)
	case domain.HybridSelectedSourceConfig:
		alias, _ := selection.ConfigTemplateAlias()
		return useCase.resolveHybridConfigSource(projectRoot, alias)
	default:
		return hybridResolvedSource{}, fmt.Errorf("hybrid source selection error: unsupported hybrid source selector: %s", selection.Kind())
	}
}

func (useCase *GenerateHybridChange) resolveHybridBuiltInSource(templateName domain.TemplateName) (hybridResolvedSource, error) {
	if useCase.templateContent == nil {
		return hybridResolvedSource{}, errors.New("hybrid source resolution error: template change content is required")
	}
	files, err := useCase.loadBuiltInTemplateFiles(templateName)
	if err != nil {
		return hybridResolvedSource{}, fmt.Errorf("hybrid source resolution error: %w", err)
	}
	return hybridResolvedSource{
		kind:              domain.HybridResolvedSourceBuiltin,
		name:              string(templateName),
		files:             files,
		builtInTypeSource: true,
		builtInTemplate:   templateName,
	}, nil
}

func (useCase *GenerateHybridChange) resolveHybridCustomSource(
	projectRoot string,
	customTemplateName domain.CustomTemplateName,
) (hybridResolvedSource, error) {
	if useCase.customTemplateFiles == nil {
		return hybridResolvedSource{}, errors.New("hybrid source resolution error: custom template filesystem is required")
	}
	templatePath := customTemplatesDirectory + "/" + customTemplateName.String()
	customTemplate, err := useCase.loadHybridCustomTemplate(projectRoot, customTemplateName, templatePath)
	if err != nil {
		return hybridResolvedSource{}, fmt.Errorf("hybrid source resolution error: %w", err)
	}
	return hybridResolvedSource{
		kind:  domain.HybridResolvedSourceCustom,
		name:  customTemplateName.String(),
		files: customTemplate.Files(),
	}, nil
}

func (useCase *GenerateHybridChange) resolveHybridConfigSource(
	projectRoot string,
	alias domain.ConfigTemplateAlias,
) (hybridResolvedSource, error) {
	if useCase.configFileSystem == nil {
		return hybridResolvedSource{}, errors.New("hybrid source resolution error: config filesystem is required")
	}
	if useCase.configParser == nil {
		return hybridResolvedSource{}, errors.New("hybrid source resolution error: config parser is required")
	}

	config, err := useCase.loadHybridConfig(projectRoot)
	if err != nil {
		return hybridResolvedSource{}, fmt.Errorf("hybrid source resolution error: %w", err)
	}
	reference, err := config.Templates.Aliases().Lookup(alias)
	if err != nil {
		return hybridResolvedSource{}, fmt.Errorf("hybrid source resolution error: %w", err)
	}

	switch reference.SourceKind() {
	case domain.ConfigTemplateSourceBuiltin:
		templateName, ok := reference.BuiltInTemplateName()
		if !ok {
			return hybridResolvedSource{}, fmt.Errorf("hybrid source resolution error: invalid config template builtin reference: %s", reference.Template())
		}
		return useCase.resolveHybridBuiltInSource(templateName)
	case domain.ConfigTemplateSourceCustom:
		customTemplateName, ok := reference.CustomTemplateName()
		if !ok {
			return hybridResolvedSource{}, fmt.Errorf("hybrid source resolution error: invalid config template custom reference: %s", reference.Template())
		}
		return useCase.resolveHybridCustomSource(projectRoot, customTemplateName)
	case domain.ConfigTemplateSourceRemote:
		remoteTemplateReference, ok := reference.RemoteTemplateReference()
		if !ok {
			return hybridResolvedSource{}, errors.New("hybrid source resolution error: invalid config template remote reference")
		}
		return useCase.resolveHybridRemoteSource(alias, remoteTemplateReference)
	default:
		return hybridResolvedSource{}, fmt.Errorf("hybrid source resolution error: unsupported config template source: %s", reference.SourceKind())
	}
}

func (useCase *GenerateHybridChange) resolveHybridRemoteSource(
	alias domain.ConfigTemplateAlias,
	reference domain.RemoteTemplateReference,
) (hybridResolvedSource, error) {
	if useCase.remoteFetcher == nil {
		return hybridResolvedSource{}, errors.New("hybrid remote error: remote template fetcher is required")
	}
	if useCase.remoteBundleReader == nil {
		return hybridResolvedSource{}, errors.New("hybrid remote error: remote template bundle reader is required")
	}

	remoteBundle, err := useCase.loadHybridRemoteTemplate(alias, reference)
	if err != nil {
		return hybridResolvedSource{}, fmt.Errorf("hybrid remote error: %w", err)
	}
	return hybridResolvedSource{
		kind:  domain.HybridResolvedSourceRemote,
		name:  reference.URL().Host(),
		files: remoteBundle.Files(),
		remoteFacts: domain.HybridRemoteFacts{
			Host:              reference.URL().Host(),
			Format:            reference.Format(),
			ChecksumAlgorithm: reference.Checksum().Algorithm(),
		},
	}, nil
}

func (useCase *GenerateHybridChange) loadBuiltInTemplateFiles(templateName domain.TemplateName) (map[string]string, error) {
	files := make(map[string]string, len(domain.RequiredOpenSpecChangeFiles()))
	for _, requiredFile := range domain.RequiredOpenSpecChangeFiles() {
		contents, err := useCase.templateContent.ContentFor(templateName, requiredFile)
		if err != nil {
			return nil, fmt.Errorf("load %s template content for %s: %w", templateName, requiredFile, err)
		}
		files[requiredFile] = contents
	}
	return files, nil
}

func (useCase *GenerateHybridChange) loadHybridCustomTemplate(
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

func (useCase *GenerateHybridChange) loadHybridConfig(projectRoot string) (domain.LocalConfig, error) {
	configExists, err := useCase.configFileSystem.FileExists(projectRoot, localConfigPath)
	if err != nil {
		return domain.LocalConfig{}, fmt.Errorf("check config file %s: %w", localConfigPath, err)
	}
	if !configExists {
		return domain.LocalConfig{}, fmt.Errorf("missing config file: %s", localConfigPath)
	}

	contents, err := useCase.configFileSystem.ReadFile(projectRoot, localConfigPath)
	if err != nil {
		return domain.LocalConfig{}, fmt.Errorf("unreadable config %s: %w", localConfigPath, err)
	}

	config, err := useCase.configParser.ParseLocalConfig(contents)
	if err != nil {
		return domain.LocalConfig{}, fmt.Errorf("invalid config YAML in %s: %w", localConfigPath, err)
	}
	if config.Version == 0 {
		return domain.LocalConfig{}, fmt.Errorf(
			"missing config version in %s: supported version is %d",
			localConfigPath,
			domain.SupportedLocalConfigVersion,
		)
	}
	if !domain.IsSupportedLocalConfigVersion(config.Version) {
		return domain.LocalConfig{}, fmt.Errorf(
			"unsupported config version %d in %s: supported version is %d",
			config.Version,
			localConfigPath,
			domain.SupportedLocalConfigVersion,
		)
	}

	return config, nil
}

func (useCase *GenerateHybridChange) loadHybridRemoteTemplate(
	alias domain.ConfigTemplateAlias,
	reference domain.RemoteTemplateReference,
) (domain.RemoteTemplateBundle, error) {
	fetchResult, err := useCase.remoteFetcher.FetchRemoteTemplate(domain.NewRemoteTemplateFetchRequest(reference.URL()))
	if err != nil {
		return domain.RemoteTemplateBundle{}, fmt.Errorf("fetch remote template for alias %s: %w", alias, err)
	}

	downloadedBytes := fetchResult.Body()
	actualChecksum, matches := reference.Checksum().MatchesBytes(downloadedBytes)
	if !matches {
		return domain.RemoteTemplateBundle{}, fmt.Errorf(
			"remote template checksum mismatch for alias %s: expected %s, got %s",
			alias,
			reference.Checksum(),
			actualChecksum,
		)
	}

	bundle, err := useCase.remoteBundleReader.ReadRemoteTemplateBundle(
		downloadedBytes,
		domain.DefaultRemoteTemplateArchivePolicy(),
	)
	if err != nil {
		return domain.RemoteTemplateBundle{}, fmt.Errorf("read remote template archive for alias %s: %w", alias, err)
	}
	return domain.NewRemoteTemplateBundle(bundle.Files())
}

func (useCase *GenerateHybridChange) requireOpenSpecProject(projectRoot string) error {
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

func (useCase *GenerateHybridChange) ensureChangeDirectory(projectRoot string, changePath string) (bool, error) {
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

func validateHybridRenderedFiles(files map[string]string) error {
	requiredSet := make(map[string]struct{}, len(domain.RequiredOpenSpecChangeFiles()))
	for _, requiredFile := range domain.RequiredOpenSpecChangeFiles() {
		requiredSet[requiredFile] = struct{}{}
		contents, exists := files[requiredFile]
		if !exists {
			return fmt.Errorf("missing rendered hybrid template content for %s", requiredFile)
		}
		if strings.TrimSpace(contents) == "" {
			return fmt.Errorf("rendered hybrid template content for %s is empty", requiredFile)
		}
	}
	for file := range files {
		if _, exists := requiredSet[file]; !exists {
			return fmt.Errorf("unsupported rendered hybrid template file: %s", file)
		}
	}
	return nil
}

func (useCase *GenerateHybridChange) writeRenderedHybridFiles(
	projectRoot string,
	changePath string,
	files map[string]string,
) ([]string, []string, error) {
	var createdFiles []string
	var skippedExistingFiles []string

	for _, requiredFile := range domain.RequiredOpenSpecChangeFiles() {
		created, err := useCase.fileSystem.WriteFileIfAbsent(projectRoot, changePath+"/"+requiredFile, files[requiredFile])
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
