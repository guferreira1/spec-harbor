package domain

import (
	"errors"
	"fmt"
	"strings"
)

type HybridSelectedSourceKind string

const (
	HybridSelectedSourceBuiltIn HybridSelectedSourceKind = "built-in"
	HybridSelectedSourceCustom  HybridSelectedSourceKind = "custom"
	HybridSelectedSourceConfig  HybridSelectedSourceKind = "config"
)

type HybridResolvedSourceKind string

const (
	HybridResolvedSourceBuiltin HybridResolvedSourceKind = "builtin"
	HybridResolvedSourceCustom  HybridResolvedSourceKind = "custom"
	HybridResolvedSourceRemote  HybridResolvedSourceKind = "remote"
)

type HybridSourceSelection struct {
	kind       HybridSelectedSourceKind
	name       string
	template   TemplateName
	customName CustomTemplateName
	alias      ConfigTemplateAlias
}

func NewHybridSourceSelection(templateName string, customTemplateName string, configTemplateAlias string) (HybridSourceSelection, error) {
	sources := 0
	if strings.TrimSpace(templateName) != "" {
		sources++
	}
	if strings.TrimSpace(customTemplateName) != "" {
		sources++
	}
	if strings.TrimSpace(configTemplateAlias) != "" {
		sources++
	}

	if sources == 0 {
		return HybridSourceSelection{}, errors.New("hybrid source selector is required")
	}
	if sources > 1 {
		return HybridSourceSelection{}, errors.New("hybrid requires exactly one source selector")
	}

	if strings.TrimSpace(templateName) != "" {
		template, err := ParseTemplateName(templateName)
		if err != nil {
			return HybridSourceSelection{}, err
		}
		return HybridSourceSelection{
			kind:     HybridSelectedSourceBuiltIn,
			name:     string(template),
			template: template,
		}, nil
	}

	if strings.TrimSpace(customTemplateName) != "" {
		customName, err := NewCustomTemplateName(customTemplateName)
		if err != nil {
			return HybridSourceSelection{}, err
		}
		return HybridSourceSelection{
			kind:       HybridSelectedSourceCustom,
			name:       customName.String(),
			customName: customName,
		}, nil
	}

	alias, err := NewConfigTemplateAlias(configTemplateAlias)
	if err != nil {
		return HybridSourceSelection{}, err
	}
	return HybridSourceSelection{
		kind:  HybridSelectedSourceConfig,
		name:  alias.String(),
		alias: alias,
	}, nil
}

func (selection HybridSourceSelection) Kind() HybridSelectedSourceKind {
	return selection.kind
}

func (selection HybridSourceSelection) Name() string {
	return selection.name
}

func (selection HybridSourceSelection) TemplateName() (TemplateName, bool) {
	if selection.kind != HybridSelectedSourceBuiltIn {
		return "", false
	}
	return selection.template, true
}

func (selection HybridSourceSelection) CustomTemplateName() (CustomTemplateName, bool) {
	if selection.kind != HybridSelectedSourceCustom {
		return CustomTemplateName{}, false
	}
	return selection.customName, true
}

func (selection HybridSourceSelection) ConfigTemplateAlias() (ConfigTemplateAlias, bool) {
	if selection.kind != HybridSelectedSourceConfig {
		return ConfigTemplateAlias{}, false
	}
	return selection.alias, true
}

type HybridType string

const (
	HybridTypeFeature  HybridType = "feature"
	HybridTypeBugfix   HybridType = "bugfix"
	HybridTypeDocs     HybridType = "docs"
	HybridTypeRefactor HybridType = "refactor"
)

func ParseHybridType(value string) (HybridType, error) {
	hybridType := HybridType(strings.TrimSpace(value))
	if hybridType == "" {
		return "", errors.New("hybrid type is required")
	}
	if !hybridType.IsSupported() {
		return "", fmt.Errorf("unsupported hybrid type: %s", hybridType)
	}
	return hybridType, nil
}

func (hybridType HybridType) IsSupported() bool {
	switch hybridType {
	case HybridTypeFeature, HybridTypeBugfix, HybridTypeDocs, HybridTypeRefactor:
		return true
	default:
		return false
	}
}

func hybridTypeFromTemplateName(templateName TemplateName) HybridType {
	return HybridType(templateName)
}

type HybridMetadata struct {
	title          string
	summary        string
	providedType   HybridType
	typeProvided   bool
	effectiveType  HybridType
	typeEffective  bool
	typeIsDerived  bool
	effectiveReady bool
}

func NewHybridMetadata(title string, summary string, optionalType string) (HybridMetadata, error) {
	trimmedTitle := strings.TrimSpace(title)
	if trimmedTitle == "" {
		return HybridMetadata{}, errors.New("hybrid title is required")
	}
	trimmedSummary := strings.TrimSpace(summary)
	if trimmedSummary == "" {
		return HybridMetadata{}, errors.New("hybrid summary is required")
	}

	metadata := HybridMetadata{
		title:   trimmedTitle,
		summary: trimmedSummary,
	}
	if strings.TrimSpace(optionalType) == "" {
		return metadata, nil
	}

	hybridType, err := ParseHybridType(optionalType)
	if err != nil {
		return HybridMetadata{}, err
	}
	metadata.providedType = hybridType
	metadata.typeProvided = true
	metadata.effectiveType = hybridType
	metadata.typeEffective = true
	return metadata, nil
}

func (metadata HybridMetadata) Title() string {
	return metadata.title
}

func (metadata HybridMetadata) Summary() string {
	return metadata.summary
}

func (metadata HybridMetadata) ProvidedType() (HybridType, bool) {
	if !metadata.typeProvided {
		return "", false
	}
	return metadata.providedType, true
}

func (metadata HybridMetadata) EffectiveType() (HybridType, bool) {
	if !metadata.typeEffective {
		return "", false
	}
	return metadata.effectiveType, true
}

func (metadata HybridMetadata) TypeIsDerived() bool {
	return metadata.typeIsDerived
}

func (metadata HybridMetadata) WithBuiltInEffectiveType(templateName TemplateName) (HybridMetadata, error) {
	expectedType := hybridTypeFromTemplateName(templateName)
	if metadata.typeProvided && metadata.providedType != expectedType {
		return HybridMetadata{}, fmt.Errorf("hybrid type %s does not match built-in template %s", metadata.providedType, templateName)
	}
	metadata.effectiveType = expectedType
	metadata.typeEffective = true
	metadata.typeIsDerived = !metadata.typeProvided
	metadata.effectiveReady = true
	return metadata, nil
}

func (metadata HybridMetadata) WithoutDerivedType() HybridMetadata {
	metadata.effectiveReady = true
	return metadata
}

func (metadata HybridMetadata) RenderContent(source string, changeID string) string {
	replacements := []string{
		"{{change_id}}", changeID,
		"{{title}}", metadata.title,
		"{{summary}}", metadata.summary,
	}
	if metadata.typeEffective {
		replacements = append(replacements, "{{type}}", string(metadata.effectiveType))
	}
	return strings.NewReplacer(replacements...).Replace(source)
}

func (metadata HybridMetadata) RenderFiles(files map[string]string, changeID string) map[string]string {
	rendered := make(map[string]string, len(files))
	for file, contents := range files {
		rendered[file] = metadata.RenderContent(contents, changeID)
	}
	return rendered
}

type HybridRemoteFacts struct {
	Host              string
	Format            RemoteTemplateFormat
	ChecksumAlgorithm ChecksumAlgorithm
}

type HybridGenerationResult struct {
	ChangeID               string
	Mode                   GenerationMode
	SelectedSourceKind     HybridSelectedSourceKind
	SelectedSourceName     string
	ResolvedSourceKind     HybridResolvedSourceKind
	ResolvedSourceName     string
	Title                  string
	Summary                string
	EffectiveType          HybridType
	TypeAvailable          bool
	ChangePath             string
	ChangeDirectoryCreated bool
	RemoteFacts            HybridRemoteFacts

	ProviderAPIsCalled     bool
	LLMAPIsCalled          bool
	AgentCommandsExecuted  bool
	AIOutputFileImported   bool
	ProductionCodeModified bool
	VCSCommandsRun         bool
	AutomationPerformed    bool

	createdFiles         []string
	skippedExistingFiles []string
	validationResult     ValidationResult
	validationRan        bool
}

func NewHybridGenerationResult(
	changeID string,
	selection HybridSourceSelection,
	resolvedKind HybridResolvedSourceKind,
	resolvedSourceName string,
	metadata HybridMetadata,
	changePath string,
	changeDirectoryCreated bool,
	createdFiles []string,
	skippedExistingFiles []string,
	validationResult ValidationResult,
	remoteFacts HybridRemoteFacts,
) HybridGenerationResult {
	effectiveType, typeAvailable := metadata.EffectiveType()
	return HybridGenerationResult{
		ChangeID:               changeID,
		Mode:                   HybridMode,
		SelectedSourceKind:     selection.Kind(),
		SelectedSourceName:     selection.Name(),
		ResolvedSourceKind:     resolvedKind,
		ResolvedSourceName:     resolvedSourceName,
		Title:                  metadata.Title(),
		Summary:                metadata.Summary(),
		EffectiveType:          effectiveType,
		TypeAvailable:          typeAvailable,
		ChangePath:             changePath,
		ChangeDirectoryCreated: changeDirectoryCreated,
		RemoteFacts:            remoteFacts,
		createdFiles:           append([]string(nil), createdFiles...),
		skippedExistingFiles:   append([]string(nil), skippedExistingFiles...),
		validationResult:       copyValidationResult(validationResult),
		validationRan:          true,
	}
}

func (result HybridGenerationResult) CreatedFiles() []string {
	return append([]string(nil), result.createdFiles...)
}

func (result HybridGenerationResult) SkippedExistingFiles() []string {
	return append([]string(nil), result.skippedExistingFiles...)
}

func (result HybridGenerationResult) ValidationResult() (ValidationResult, bool) {
	if !result.validationRan {
		return ValidationResult{}, false
	}
	return copyValidationResult(result.validationResult), true
}
