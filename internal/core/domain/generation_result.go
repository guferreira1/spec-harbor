package domain

type GenerationResult struct {
	ChangeID               string
	Mode                   GenerationMode
	TemplateName           TemplateName
	TemplateSource         TemplateSource
	CustomTemplateName     string
	TemplatePath           string
	ConfigTemplateAlias    string
	ConfigTemplateSource   ConfigTemplateSourceKind
	ConfigTemplateName     string
	GuidedType             GuidedType
	GuidedTitle            string
	ChangePath             string
	ChangeDirectoryCreated bool
	createdFiles           []string
	skippedExistingFiles   []string
}

func NewGenerationResult(
	changeID string,
	mode GenerationMode,
	changePath string,
	changeDirectoryCreated bool,
	createdFiles []string,
	skippedExistingFiles []string,
) GenerationResult {
	return GenerationResult{
		ChangeID:               changeID,
		Mode:                   mode,
		ChangePath:             changePath,
		ChangeDirectoryCreated: changeDirectoryCreated,
		createdFiles:           append([]string(nil), createdFiles...),
		skippedExistingFiles:   append([]string(nil), skippedExistingFiles...),
	}
}

func NewTemplateGenerationResult(
	changeID string,
	templateName TemplateName,
	changePath string,
	changeDirectoryCreated bool,
	createdFiles []string,
	skippedExistingFiles []string,
) GenerationResult {
	return GenerationResult{
		ChangeID:               changeID,
		Mode:                   TemplateMode,
		TemplateName:           templateName,
		TemplateSource:         BuiltInTemplateSource,
		ChangePath:             changePath,
		ChangeDirectoryCreated: changeDirectoryCreated,
		createdFiles:           append([]string(nil), createdFiles...),
		skippedExistingFiles:   append([]string(nil), skippedExistingFiles...),
	}
}

func NewCustomTemplateGenerationResult(
	changeID string,
	customTemplateName CustomTemplateName,
	templatePath string,
	changePath string,
	changeDirectoryCreated bool,
	createdFiles []string,
	skippedExistingFiles []string,
) GenerationResult {
	return GenerationResult{
		ChangeID:               changeID,
		Mode:                   TemplateMode,
		TemplateSource:         CustomTemplateSource,
		CustomTemplateName:     customTemplateName.String(),
		TemplatePath:           templatePath,
		ChangePath:             changePath,
		ChangeDirectoryCreated: changeDirectoryCreated,
		createdFiles:           append([]string(nil), createdFiles...),
		skippedExistingFiles:   append([]string(nil), skippedExistingFiles...),
	}
}

func NewConfigTemplateBuiltInGenerationResult(
	changeID string,
	configTemplateAlias ConfigTemplateAlias,
	templateName TemplateName,
	changePath string,
	changeDirectoryCreated bool,
	createdFiles []string,
	skippedExistingFiles []string,
) GenerationResult {
	return GenerationResult{
		ChangeID:               changeID,
		Mode:                   TemplateMode,
		TemplateName:           templateName,
		TemplateSource:         BuiltInTemplateSource,
		ConfigTemplateAlias:    configTemplateAlias.String(),
		ConfigTemplateSource:   ConfigTemplateSourceBuiltin,
		ConfigTemplateName:     string(templateName),
		ChangePath:             changePath,
		ChangeDirectoryCreated: changeDirectoryCreated,
		createdFiles:           append([]string(nil), createdFiles...),
		skippedExistingFiles:   append([]string(nil), skippedExistingFiles...),
	}
}

func NewConfigTemplateCustomGenerationResult(
	changeID string,
	configTemplateAlias ConfigTemplateAlias,
	customTemplateName CustomTemplateName,
	templatePath string,
	changePath string,
	changeDirectoryCreated bool,
	createdFiles []string,
	skippedExistingFiles []string,
) GenerationResult {
	return GenerationResult{
		ChangeID:               changeID,
		Mode:                   TemplateMode,
		TemplateSource:         CustomTemplateSource,
		CustomTemplateName:     customTemplateName.String(),
		TemplatePath:           templatePath,
		ConfigTemplateAlias:    configTemplateAlias.String(),
		ConfigTemplateSource:   ConfigTemplateSourceCustom,
		ConfigTemplateName:     customTemplateName.String(),
		ChangePath:             changePath,
		ChangeDirectoryCreated: changeDirectoryCreated,
		createdFiles:           append([]string(nil), createdFiles...),
		skippedExistingFiles:   append([]string(nil), skippedExistingFiles...),
	}
}

func NewGuidedGenerationResult(
	changeID string,
	guidedType GuidedType,
	guidedTitle string,
	changePath string,
	changeDirectoryCreated bool,
	createdFiles []string,
	skippedExistingFiles []string,
) GenerationResult {
	return GenerationResult{
		ChangeID:               changeID,
		Mode:                   GuidedMode,
		GuidedType:             guidedType,
		GuidedTitle:            guidedTitle,
		ChangePath:             changePath,
		ChangeDirectoryCreated: changeDirectoryCreated,
		createdFiles:           append([]string(nil), createdFiles...),
		skippedExistingFiles:   append([]string(nil), skippedExistingFiles...),
	}
}

func (result GenerationResult) CreatedFiles() []string {
	return append([]string(nil), result.createdFiles...)
}

func (result GenerationResult) SkippedExistingFiles() []string {
	return append([]string(nil), result.skippedExistingFiles...)
}
