package domain

type GenerationResult struct {
	ChangeID               string
	Mode                   GenerationMode
	TemplateName           TemplateName
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
