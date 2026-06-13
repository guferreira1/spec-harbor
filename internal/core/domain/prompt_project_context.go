package domain

import (
	"fmt"
	"sort"
	"strings"
)

const (
	promptContextMissingInstructions = "No confirmed project context, detected facts, or suggested assumptions are available from supported local context sources."
	promptContextOmittedNotice       = "Additional context signals omitted by prompt size limit."
)

type PromptContextRenderPolicy struct {
	MaxUserConfirmedItems       int
	MaxDetectedFactItems        int
	MaxSuggestedAssumptionItems int
	MaxValueLength              int
	MaxSourceLength             int
	MaxTotalCharacters          int
}

type PromptContextItem struct {
	Kind       ContextSignalKind
	Label      string
	Value      string
	Source     ContextSource
	Confidence ContextConfidence
}

type PromptContextConflict struct {
	Message string
}

type PromptProjectContext struct {
	policy               PromptContextRenderPolicy
	userConfirmed        []PromptContextItem
	detectedFacts        []PromptContextItem
	suggestedAssumptions []PromptContextItem
	conflicts            []PromptContextConflict
	omitted              bool
}

func DefaultPromptContextRenderPolicy() PromptContextRenderPolicy {
	return PromptContextRenderPolicy{
		MaxUserConfirmedItems:       12,
		MaxDetectedFactItems:        22,
		MaxSuggestedAssumptionItems: 6,
		MaxValueLength:              160,
		MaxSourceLength:             120,
		MaxTotalCharacters:          6000,
	}
}

func NewPromptProjectContext(
	result ContextDiscoveryResult,
	policy PromptContextRenderPolicy,
) PromptProjectContext {
	normalizedPolicy := normalizePromptContextRenderPolicy(policy)
	assembler := newPromptContextAssembler(normalizedPolicy)
	return assembler.assemble(result)
}

func (context PromptProjectContext) HasUserConfirmedContext() bool {
	return len(context.userConfirmed) > 0
}

func (context PromptProjectContext) HasContextSignals() bool {
	return len(context.userConfirmed) > 0 ||
		len(context.detectedFacts) > 0 ||
		len(context.suggestedAssumptions) > 0
}

func (context PromptProjectContext) UserConfirmedContext() []PromptContextItem {
	return append([]PromptContextItem(nil), context.userConfirmed...)
}

func (context PromptProjectContext) DetectedFacts() []PromptContextItem {
	return append([]PromptContextItem(nil), context.detectedFacts...)
}

func (context PromptProjectContext) SuggestedAssumptions() []PromptContextItem {
	return append([]PromptContextItem(nil), context.suggestedAssumptions...)
}

func (context PromptProjectContext) Conflicts() []PromptContextConflict {
	return append([]PromptContextConflict(nil), context.conflicts...)
}

func (context PromptProjectContext) RenderMarkdown() string {
	rendered := renderPromptProjectContext(context)
	policy := normalizePromptContextRenderPolicy(context.policy)
	if len(rendered) <= policy.MaxTotalCharacters {
		return rendered
	}

	bounded := context
	for len(rendered) > policy.MaxTotalCharacters && bounded.removeLastLowestPriorityItem() {
		bounded.omitted = true
		rendered = renderPromptProjectContext(bounded)
	}
	return rendered
}

func (context *PromptProjectContext) removeLastLowestPriorityItem() bool {
	if len(context.suggestedAssumptions) > 0 {
		context.suggestedAssumptions = context.suggestedAssumptions[:len(context.suggestedAssumptions)-1]
		return true
	}
	if len(context.detectedFacts) > 0 {
		context.detectedFacts = context.detectedFacts[:len(context.detectedFacts)-1]
		return true
	}
	if len(context.userConfirmed) > 0 {
		context.userConfirmed = context.userConfirmed[:len(context.userConfirmed)-1]
		return true
	}
	return false
}

type promptContextAssembler struct {
	policy PromptContextRenderPolicy
}

func newPromptContextAssembler(policy PromptContextRenderPolicy) promptContextAssembler {
	return promptContextAssembler{policy: policy}
}

func (assembler promptContextAssembler) assemble(result ContextDiscoveryResult) PromptProjectContext {
	signals := result.Signals()
	sort.SliceStable(signals, func(i, j int) bool {
		left := signals[i]
		right := signals[j]
		leftPriority := promptContextKindPriority(left.Kind)
		rightPriority := promptContextKindPriority(right.Kind)
		if leftPriority != rightPriority {
			return leftPriority < rightPriority
		}
		if left.Kind != right.Kind {
			return left.Kind < right.Kind
		}
		leftConfidence := promptContextConfidencePriority(left.Confidence)
		rightConfidence := promptContextConfidencePriority(right.Confidence)
		if leftConfidence != rightConfidence {
			return leftConfidence < rightConfidence
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
		return left.Source.Evidence < right.Source.Evidence
	})

	confirmedByKind := make(map[ContextSignalKind][]ContextSignal)
	detectedByKind := make(map[ContextSignalKind][]ContextSignal)
	for _, signal := range signals {
		if !isPromptRelevantContextSignal(signal) {
			continue
		}
		switch signal.Classification {
		case ContextSignalClassificationUserConfirmedContext:
			confirmedByKind[signal.Kind] = append(confirmedByKind[signal.Kind], signal)
		case ContextSignalClassificationDetectedFact:
			detectedByKind[signal.Kind] = append(detectedByKind[signal.Kind], signal)
		}
	}

	context := PromptProjectContext{policy: assembler.policy}
	confirmedValues := make(map[ContextSignalKind]map[string]bool)
	for _, signal := range signals {
		if signal.Classification != ContextSignalClassificationUserConfirmedContext ||
			!isPromptRelevantContextSignal(signal) {
			continue
		}
		context.addUserConfirmed(signal, assembler.policy)
		if confirmedValues[signal.Kind] == nil {
			confirmedValues[signal.Kind] = make(map[string]bool)
		}
		confirmedValues[signal.Kind][normalizePromptContextValue(signal.Value)] = true
	}

	for _, signal := range signals {
		if signal.Classification != ContextSignalClassificationDetectedFact ||
			!isPromptRelevantContextSignal(signal) {
			continue
		}
		if values := confirmedValues[signal.Kind]; len(values) > 0 {
			if values[normalizePromptContextValue(signal.Value)] {
				context.omitted = true
				continue
			}
			context.addConflict(signal, confirmedByKind[signal.Kind], assembler.policy)
		}
		context.addDetectedFact(signal, assembler.policy)
	}

	for _, signal := range signals {
		if signal.Classification != ContextSignalClassificationSuggestedAssumption ||
			!isPromptRelevantContextSignal(signal) {
			continue
		}
		if len(confirmedByKind[signal.Kind]) > 0 || len(detectedByKind[signal.Kind]) > 0 {
			context.omitted = true
			continue
		}
		context.addSuggestedAssumption(signal, assembler.policy)
	}

	return context
}

func (context *PromptProjectContext) addUserConfirmed(signal ContextSignal, policy PromptContextRenderPolicy) {
	if promptContextItemsContain(context.userConfirmed, signal) {
		return
	}
	if len(context.userConfirmed) >= policy.MaxUserConfirmedItems {
		context.omitted = true
		return
	}
	context.userConfirmed = append(context.userConfirmed, promptContextItemFromSignal(signal, policy))
}

func (context *PromptProjectContext) addDetectedFact(signal ContextSignal, policy PromptContextRenderPolicy) {
	if promptContextItemsContain(context.detectedFacts, signal) {
		return
	}
	if promptContextKindCount(context.detectedFacts, signal.Kind) >= maxPromptContextItemsPerKind(signal.Kind) {
		context.omitted = true
		return
	}
	if len(context.detectedFacts) >= policy.MaxDetectedFactItems {
		context.omitted = true
		return
	}
	context.detectedFacts = append(context.detectedFacts, promptContextItemFromSignal(signal, policy))
}

func (context *PromptProjectContext) addSuggestedAssumption(signal ContextSignal, policy PromptContextRenderPolicy) {
	if promptContextItemsContain(context.suggestedAssumptions, signal) {
		return
	}
	if promptContextKindCount(context.suggestedAssumptions, signal.Kind) >= maxPromptContextItemsPerKind(signal.Kind) {
		context.omitted = true
		return
	}
	if len(context.suggestedAssumptions) >= policy.MaxSuggestedAssumptionItems {
		context.omitted = true
		return
	}
	context.suggestedAssumptions = append(context.suggestedAssumptions, promptContextItemFromSignal(signal, policy))
}

func (context *PromptProjectContext) addConflict(
	detected ContextSignal,
	confirmed []ContextSignal,
	policy PromptContextRenderPolicy,
) {
	if len(confirmed) == 0 {
		return
	}
	confirmedValue := truncatePromptContextText(confirmed[0].Value, policy.MaxValueLength)
	detectedValue := truncatePromptContextText(detected.Value, policy.MaxValueLength)
	confirmedLabel := promptContextLabelForSignal(confirmed[0])
	detectedLabel := promptContextLabelForSignal(detected)
	message := fmt.Sprintf(
		"Confirmed %s is %s from .specharbor/project-brief.md; detected %s includes %s from %s. Prefer the confirmed value unless the user updates the brief.",
		confirmedLabel,
		confirmedValue,
		detectedLabel,
		detectedValue,
		formatPromptContextSource(detected.Source, policy),
	)
	for _, conflict := range context.conflicts {
		if conflict.Message == message {
			return
		}
	}
	context.conflicts = append(context.conflicts, PromptContextConflict{Message: message})
}

func promptContextItemFromSignal(signal ContextSignal, policy PromptContextRenderPolicy) PromptContextItem {
	return PromptContextItem{
		Kind:       signal.Kind,
		Label:      promptContextLabelForSignal(signal),
		Value:      truncatePromptContextText(signal.Value, policy.MaxValueLength),
		Source:     truncatePromptContextSource(signal.Source, policy.MaxSourceLength),
		Confidence: signal.Confidence,
	}
}

func renderPromptProjectContext(context PromptProjectContext) string {
	policy := normalizePromptContextRenderPolicy(context.policy)
	var builder strings.Builder
	builder.WriteString("## Project Context\n\n")
	builder.WriteString("Use the context below as guidance, but do not treat assumptions as facts.\n\n")

	if !context.HasContextSignals() {
		builder.WriteString(promptContextMissingInstructions)
		builder.WriteString("\n\n")
	}

	if len(context.userConfirmed) > 0 {
		builder.WriteString("### User-confirmed context\n\n")
		for _, item := range context.userConfirmed {
			builder.WriteString("- ")
			builder.WriteString(item.Label)
			builder.WriteString(": ")
			builder.WriteString(item.Value)
			builder.WriteString("\n")
		}
		builder.WriteString("\n")
	}

	if len(context.detectedFacts) > 0 {
		builder.WriteString("### Detected facts\n\n")
		for _, item := range context.detectedFacts {
			builder.WriteString("- ")
			builder.WriteString(item.Label)
			builder.WriteString(": ")
			builder.WriteString(item.Value)
			builder.WriteString("\n")
			builder.WriteString("  Source: ")
			builder.WriteString(formatPromptContextSource(item.Source, policy))
			builder.WriteString("\n")
			builder.WriteString("  Confidence: ")
			builder.WriteString(string(item.Confidence))
			builder.WriteString("\n\n")
		}
	}

	if len(context.suggestedAssumptions) > 0 {
		builder.WriteString("### Suggested assumptions\n\n")
		for _, item := range context.suggestedAssumptions {
			builder.WriteString("- ")
			builder.WriteString(item.Label)
			builder.WriteString(" may be `")
			builder.WriteString(item.Value)
			builder.WriteString("`\n")
			builder.WriteString("  Source: ")
			builder.WriteString(formatPromptContextSource(item.Source, policy))
			builder.WriteString("\n")
			builder.WriteString("  Confidence: ")
			builder.WriteString(string(item.Confidence))
			builder.WriteString("\n\n")
		}
	}

	if len(context.conflicts) > 0 {
		builder.WriteString("### Conflict notes\n\n")
		for _, conflict := range context.conflicts {
			builder.WriteString("- ")
			builder.WriteString(conflict.Message)
			builder.WriteString("\n")
		}
		builder.WriteString("\n")
	}

	if context.omitted {
		builder.WriteString(promptContextOmittedNotice)
		builder.WriteString("\n\n")
	}

	builder.WriteString("Rules:\n")
	builder.WriteString("- Prefer user-confirmed context over detected facts.\n")
	builder.WriteString("- Do not treat suggested assumptions as facts.\n")
	builder.WriteString("- Ask before making major architecture, persistence, or workflow decisions when context is missing, ambiguous, or conflicting.\n")
	builder.WriteString("- Do not invent stack, architecture, commands, persistence decisions, workflow decisions, or project direction.\n")
	return builder.String()
}

func normalizePromptContextRenderPolicy(policy PromptContextRenderPolicy) PromptContextRenderPolicy {
	defaults := DefaultPromptContextRenderPolicy()
	if policy.MaxUserConfirmedItems <= 0 {
		policy.MaxUserConfirmedItems = defaults.MaxUserConfirmedItems
	}
	if policy.MaxDetectedFactItems <= 0 {
		policy.MaxDetectedFactItems = defaults.MaxDetectedFactItems
	}
	if policy.MaxSuggestedAssumptionItems <= 0 {
		policy.MaxSuggestedAssumptionItems = defaults.MaxSuggestedAssumptionItems
	}
	if policy.MaxValueLength <= 0 {
		policy.MaxValueLength = defaults.MaxValueLength
	}
	if policy.MaxSourceLength <= 0 {
		policy.MaxSourceLength = defaults.MaxSourceLength
	}
	if policy.MaxTotalCharacters <= 0 {
		policy.MaxTotalCharacters = defaults.MaxTotalCharacters
	}
	return policy
}

func isPromptRelevantContextSignal(signal ContextSignal) bool {
	switch signal.Kind {
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
	default:
		return false
	}

	switch signal.Kind {
	case ContextSignalKindDocumentationSource:
		return signal.Value == "README.md" || signal.Value == "CONTRIBUTING.md" || signal.Value == "docs/"
	case ContextSignalKindAgentInstructionSource:
		if signal.Classification == ContextSignalClassificationUserConfirmedContext &&
			signal.Source.Category == ContextSourceCategoryProjectBrief {
			return true
		}
		return signal.Value == "AGENTS.md" || signal.Value == ".specharbor/rules/"
	case ContextSignalKindOpenSpecSource:
		return signal.Value == "openspec/project.md" || signal.Value == "openspec/specs/"
	default:
		return true
	}
}

func promptContextLabelForSignal(signal ContextSignal) string {
	if signal.Kind == ContextSignalKindAgentInstructionSource &&
		signal.Classification == ContextSignalClassificationUserConfirmedContext &&
		signal.Source.Category == ContextSourceCategoryProjectBrief {
		return "Agent behavior"
	}
	return signal.Kind.Label()
}

func promptContextKindPriority(kind ContextSignalKind) int {
	order := []ContextSignalKind{
		ContextSignalKindProjectType,
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
		ContextSignalKindWorkflowSignal,
	}
	for index, candidate := range order {
		if kind == candidate {
			return index
		}
	}
	return len(order)
}

func promptContextConfidencePriority(confidence ContextConfidence) int {
	switch confidence {
	case ContextConfidenceHigh:
		return 0
	case ContextConfidenceMedium:
		return 1
	case ContextConfidenceLow:
		return 2
	default:
		return 3
	}
}

func promptContextItemsContain(items []PromptContextItem, signal ContextSignal) bool {
	normalizedValue := normalizePromptContextValue(signal.Value)
	for _, item := range items {
		if item.Kind == signal.Kind && normalizePromptContextValue(item.Value) == normalizedValue {
			return true
		}
	}
	return false
}

func promptContextKindCount(items []PromptContextItem, kind ContextSignalKind) int {
	count := 0
	for _, item := range items {
		if item.Kind == kind {
			count++
		}
	}
	return count
}

func maxPromptContextItemsPerKind(kind ContextSignalKind) int {
	switch kind {
	case ContextSignalKindPurposeSummary,
		ContextSignalKindTestCommand,
		ContextSignalKindBuildCommand,
		ContextSignalKindRunCommand,
		ContextSignalKindDocumentationSource,
		ContextSignalKindAgentInstructionSource,
		ContextSignalKindOpenSpecSource:
		return 2
	case ContextSignalKindStack,
		ContextSignalKindFramework,
		ContextSignalKindArchitectureHint,
		ContextSignalKindWorkflowSignal:
		return 3
	default:
		return 1
	}
}

func formatPromptContextSource(source ContextSource, policy PromptContextRenderPolicy) string {
	trimmed := truncatePromptContextText(source.Path, policy.MaxSourceLength)
	if source.Evidence == "" {
		return trimmed
	}
	evidence := truncatePromptContextText(source.Evidence, policy.MaxSourceLength)
	return fmt.Sprintf("%s (%s)", trimmed, evidence)
}

func truncatePromptContextSource(source ContextSource, maxLength int) ContextSource {
	source.Path = truncatePromptContextText(source.Path, maxLength)
	source.Evidence = truncatePromptContextText(source.Evidence, maxLength)
	return source
}

func truncatePromptContextText(value string, maxLength int) string {
	trimmed := strings.TrimSpace(value)
	if maxLength <= 0 || len(trimmed) <= maxLength {
		return trimmed
	}
	if maxLength <= 3 {
		return trimmed[:maxLength]
	}
	return strings.TrimSpace(trimmed[:maxLength-3]) + "..."
}

func normalizePromptContextValue(value string) string {
	return strings.Join(strings.Fields(strings.ToLower(strings.TrimSpace(value))), " ")
}
