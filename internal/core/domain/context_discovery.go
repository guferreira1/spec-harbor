package domain

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

type ContextSignalKind string

const (
	ContextSignalKindProjectType            ContextSignalKind = "project_type"
	ContextSignalKindPurposeSummary         ContextSignalKind = "purpose_summary"
	ContextSignalKindTargetUsers            ContextSignalKind = "target_users"
	ContextSignalKindStack                  ContextSignalKind = "stack"
	ContextSignalKindLanguage               ContextSignalKind = "language"
	ContextSignalKindFramework              ContextSignalKind = "framework"
	ContextSignalKindArchitectureHint       ContextSignalKind = "architecture_hint"
	ContextSignalKindPackageManager         ContextSignalKind = "package_manager"
	ContextSignalKindInstallCommand         ContextSignalKind = "install_command"
	ContextSignalKindTestCommand            ContextSignalKind = "test_command"
	ContextSignalKindBuildCommand           ContextSignalKind = "build_command"
	ContextSignalKindRunCommand             ContextSignalKind = "run_command"
	ContextSignalKindDocumentationSource    ContextSignalKind = "documentation_source"
	ContextSignalKindAgentInstructionSource ContextSignalKind = "agent_instruction_source"
	ContextSignalKindOpenSpecSource         ContextSignalKind = "openspec_source"
	ContextSignalKindCLIEntrypoint          ContextSignalKind = "cli_entrypoint"
	ContextSignalKindContainerSignal        ContextSignalKind = "container_signal"
	ContextSignalKindWorkflowSignal         ContextSignalKind = "workflow_signal"
)

type ContextSignalClassification string

const (
	ContextSignalClassificationDetectedFact         ContextSignalClassification = "detected_fact"
	ContextSignalClassificationSuggestedAssumption  ContextSignalClassification = "suggested_assumption"
	ContextSignalClassificationUserConfirmedContext ContextSignalClassification = "user_confirmed_context"
)

type ContextConfidence string

const (
	ContextConfidenceHigh   ContextConfidence = "high"
	ContextConfidenceMedium ContextConfidence = "medium"
	ContextConfidenceLow    ContextConfidence = "low"
)

type ContextSourceCategory string

const (
	ContextSourceCategoryProjectBrief       ContextSourceCategory = "project_brief"
	ContextSourceCategoryReadme             ContextSourceCategory = "readme"
	ContextSourceCategoryContributing       ContextSourceCategory = "contributing"
	ContextSourceCategoryDocumentation      ContextSourceCategory = "documentation"
	ContextSourceCategoryAgentInstruction   ContextSourceCategory = "agent_instruction"
	ContextSourceCategoryOpenSpecProject    ContextSourceCategory = "openspec_project"
	ContextSourceCategoryOpenSpecSpec       ContextSourceCategory = "openspec_spec"
	ContextSourceCategorySpecHarborRules    ContextSourceCategory = "specharbor_rules"
	ContextSourceCategoryPackageManifest    ContextSourceCategory = "package_manifest"
	ContextSourceCategoryBuildManifest      ContextSourceCategory = "build_manifest"
	ContextSourceCategoryDependencyManifest ContextSourceCategory = "dependency_manifest"
	ContextSourceCategoryTaskRunner         ContextSourceCategory = "task_runner"
	ContextSourceCategoryContainerConfig    ContextSourceCategory = "container_config"
	ContextSourceCategoryWorkflowConfig     ContextSourceCategory = "workflow_config"
	ContextSourceCategoryRepositoryLayout   ContextSourceCategory = "repository_layout"
)

type ContextSource struct {
	Path     string
	Category ContextSourceCategory
	Evidence string
}

type ContextSignalInput struct {
	Kind           ContextSignalKind
	Value          string
	Classification ContextSignalClassification
	Confidence     ContextConfidence
	Source         ContextSource
}

type ContextSignal struct {
	Kind           ContextSignalKind
	Value          string
	Classification ContextSignalClassification
	Confidence     ContextConfidence
	Source         ContextSource
}

type ContextDiscoveryNote struct {
	Message string
}

type ContextDiscoveryResult struct {
	signals []ContextSignal
	notes   []ContextDiscoveryNote
}

func NewContextSignal(input ContextSignalInput) (ContextSignal, error) {
	if !input.Kind.IsSupported() {
		return ContextSignal{}, fmt.Errorf("unsupported context signal kind: %s", input.Kind)
	}
	value := strings.TrimSpace(input.Value)
	if value == "" {
		return ContextSignal{}, errors.New("context signal value is required")
	}
	if !input.Classification.IsSupported() {
		return ContextSignal{}, fmt.Errorf("unsupported context signal classification: %s", input.Classification)
	}
	if !input.Confidence.IsSupported() {
		return ContextSignal{}, fmt.Errorf("unsupported context confidence: %s", input.Confidence)
	}
	source, err := NewContextSource(input.Source.Path, input.Source.Category, input.Source.Evidence)
	if err != nil {
		return ContextSignal{}, err
	}

	return ContextSignal{
		Kind:           input.Kind,
		Value:          value,
		Classification: input.Classification,
		Confidence:     input.Confidence,
		Source:         source,
	}, nil
}

func NewContextSource(sourcePath string, category ContextSourceCategory, evidence string) (ContextSource, error) {
	trimmedPath := strings.TrimSpace(sourcePath)
	if trimmedPath == "" {
		return ContextSource{}, errors.New("context source path is required")
	}
	if strings.HasPrefix(trimmedPath, "/") || strings.Contains(trimmedPath, "\\") || pathContainsTraversal(trimmedPath) {
		return ContextSource{}, fmt.Errorf("context source path must be a safe relative path: %s", sourcePath)
	}
	if !category.IsSupported() {
		return ContextSource{}, fmt.Errorf("unsupported context source category: %s", category)
	}
	return ContextSource{
		Path:     normalizeContextPath(trimmedPath),
		Category: category,
		Evidence: strings.TrimSpace(evidence),
	}, nil
}

func NewContextDiscoveryNote(message string) (ContextDiscoveryNote, error) {
	trimmed := strings.TrimSpace(message)
	if trimmed == "" {
		return ContextDiscoveryNote{}, errors.New("context discovery note message is required")
	}
	return ContextDiscoveryNote{Message: trimmed}, nil
}

func NewContextDiscoveryResult(signals []ContextSignal, notes []ContextDiscoveryNote) ContextDiscoveryResult {
	copiedSignals := append([]ContextSignal(nil), signals...)
	SortContextSignals(copiedSignals)

	return ContextDiscoveryResult{
		signals: deduplicateContextSignals(copiedSignals),
		notes:   append([]ContextDiscoveryNote(nil), notes...),
	}
}

func (result ContextDiscoveryResult) Signals() []ContextSignal {
	return append([]ContextSignal(nil), result.signals...)
}

func (result ContextDiscoveryResult) Notes() []ContextDiscoveryNote {
	return append([]ContextDiscoveryNote(nil), result.notes...)
}

func (result ContextDiscoveryResult) SignalsByClassification(
	classification ContextSignalClassification,
) []ContextSignal {
	var signals []ContextSignal
	for _, signal := range result.signals {
		if signal.Classification == classification {
			signals = append(signals, signal)
		}
	}
	return signals
}

func SortContextSignals(signals []ContextSignal) {
	sort.SliceStable(signals, func(i, j int) bool {
		left := signals[i]
		right := signals[j]
		if left.Kind != right.Kind {
			return left.Kind < right.Kind
		}
		if left.Source.Path != right.Source.Path {
			return left.Source.Path < right.Source.Path
		}
		if left.Value != right.Value {
			return left.Value < right.Value
		}
		if left.Classification != right.Classification {
			return left.Classification < right.Classification
		}
		if left.Confidence != right.Confidence {
			return left.Confidence < right.Confidence
		}
		return left.Source.Evidence < right.Source.Evidence
	})
}

func deduplicateContextSignals(signals []ContextSignal) []ContextSignal {
	seen := make(map[string]bool)
	deduplicated := make([]ContextSignal, 0, len(signals))
	for _, signal := range signals {
		key := strings.Join([]string{
			string(signal.Kind),
			signal.Value,
			string(signal.Classification),
			string(signal.Confidence),
			signal.Source.Path,
			string(signal.Source.Category),
			signal.Source.Evidence,
		}, "\x00")
		if seen[key] {
			continue
		}
		seen[key] = true
		deduplicated = append(deduplicated, signal)
	}
	return deduplicated
}

func (kind ContextSignalKind) IsSupported() bool {
	switch kind {
	case ContextSignalKindProjectType,
		ContextSignalKindPurposeSummary,
		ContextSignalKindTargetUsers,
		ContextSignalKindStack,
		ContextSignalKindLanguage,
		ContextSignalKindFramework,
		ContextSignalKindArchitectureHint,
		ContextSignalKindPackageManager,
		ContextSignalKindInstallCommand,
		ContextSignalKindTestCommand,
		ContextSignalKindBuildCommand,
		ContextSignalKindRunCommand,
		ContextSignalKindDocumentationSource,
		ContextSignalKindAgentInstructionSource,
		ContextSignalKindOpenSpecSource,
		ContextSignalKindCLIEntrypoint,
		ContextSignalKindContainerSignal,
		ContextSignalKindWorkflowSignal:
		return true
	default:
		return false
	}
}

func (kind ContextSignalKind) Label() string {
	switch kind {
	case ContextSignalKindProjectType:
		return "Project type"
	case ContextSignalKindPurposeSummary:
		return "Purpose summary"
	case ContextSignalKindTargetUsers:
		return "Target users"
	case ContextSignalKindStack:
		return "Stack"
	case ContextSignalKindLanguage:
		return "Language"
	case ContextSignalKindFramework:
		return "Framework"
	case ContextSignalKindArchitectureHint:
		return "Architecture"
	case ContextSignalKindPackageManager:
		return "Package manager"
	case ContextSignalKindInstallCommand:
		return "Install command"
	case ContextSignalKindTestCommand:
		return "Test command"
	case ContextSignalKindBuildCommand:
		return "Build command"
	case ContextSignalKindRunCommand:
		return "Run command"
	case ContextSignalKindDocumentationSource:
		return "Documentation source"
	case ContextSignalKindAgentInstructionSource:
		return "Agent rules"
	case ContextSignalKindOpenSpecSource:
		return "OpenSpec source"
	case ContextSignalKindCLIEntrypoint:
		return "CLI entrypoint"
	case ContextSignalKindContainerSignal:
		return "Container"
	case ContextSignalKindWorkflowSignal:
		return "Workflow"
	default:
		return string(kind)
	}
}

func (classification ContextSignalClassification) IsSupported() bool {
	switch classification {
	case ContextSignalClassificationDetectedFact,
		ContextSignalClassificationSuggestedAssumption,
		ContextSignalClassificationUserConfirmedContext:
		return true
	default:
		return false
	}
}

func (confidence ContextConfidence) IsSupported() bool {
	switch confidence {
	case ContextConfidenceHigh, ContextConfidenceMedium, ContextConfidenceLow:
		return true
	default:
		return false
	}
}

func (category ContextSourceCategory) IsSupported() bool {
	switch category {
	case ContextSourceCategoryProjectBrief,
		ContextSourceCategoryReadme,
		ContextSourceCategoryContributing,
		ContextSourceCategoryDocumentation,
		ContextSourceCategoryAgentInstruction,
		ContextSourceCategoryOpenSpecProject,
		ContextSourceCategoryOpenSpecSpec,
		ContextSourceCategorySpecHarborRules,
		ContextSourceCategoryPackageManifest,
		ContextSourceCategoryBuildManifest,
		ContextSourceCategoryDependencyManifest,
		ContextSourceCategoryTaskRunner,
		ContextSourceCategoryContainerConfig,
		ContextSourceCategoryWorkflowConfig,
		ContextSourceCategoryRepositoryLayout:
		return true
	default:
		return false
	}
}

func ContextDiscoverySensitiveFilePatterns() []string {
	return []string{
		".env",
		".env.*",
		"*.pem",
		"*.key",
		"id_rsa",
		"id_ed25519",
		"secrets.*",
		"credentials.*",
	}
}

func ContextDiscoveryGeneratedDirectoryNames() []string {
	return []string{
		".git",
		"node_modules",
		"dist",
		"build",
		"target",
		"vendor",
		"coverage",
		".tmp",
		".cache",
		".next",
		".nuxt",
		"out",
		"bin",
		"obj",
	}
}

func ShouldSkipContextDiscoveryPath(relativePath string) bool {
	normalized := normalizeContextPath(relativePath)
	if normalized == "" || normalized == "." {
		return false
	}

	segments := strings.Split(normalized, "/")
	for _, segment := range segments[:len(segments)-1] {
		if isGeneratedContextDirectoryName(segment) {
			return true
		}
	}

	base := segments[len(segments)-1]
	return isSensitiveContextFileName(base) || isGeneratedContextDirectoryName(base)
}

func isSensitiveContextFileName(name string) bool {
	switch name {
	case ".env", "id_rsa", "id_ed25519":
		return true
	}
	return strings.HasPrefix(name, ".env.") ||
		strings.HasSuffix(name, ".pem") ||
		strings.HasSuffix(name, ".key") ||
		strings.HasPrefix(name, "secrets.") ||
		strings.HasPrefix(name, "credentials.")
}

func isGeneratedContextDirectoryName(name string) bool {
	for _, generated := range ContextDiscoveryGeneratedDirectoryNames() {
		if name == generated {
			return true
		}
	}
	return false
}

func normalizeContextPath(value string) string {
	normalized := strings.ReplaceAll(strings.TrimSpace(value), "\\", "/")
	for strings.Contains(normalized, "//") {
		normalized = strings.ReplaceAll(normalized, "//", "/")
	}
	return strings.TrimPrefix(normalized, "./")
}

func pathContainsTraversal(value string) bool {
	normalized := normalizeContextPath(value)
	for _, segment := range strings.Split(normalized, "/") {
		if segment == ".." {
			return true
		}
	}
	return false
}
