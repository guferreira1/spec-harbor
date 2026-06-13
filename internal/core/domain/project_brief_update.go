package domain

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

type ProjectBriefSourceCategory string

const (
	ProjectBriefSourceCategoryUserConfirmedContext ProjectBriefSourceCategory = "user_confirmed_context"
	ProjectBriefSourceCategoryDetectedFact         ProjectBriefSourceCategory = "detected_fact"
	ProjectBriefSourceCategorySuggestedAssumption  ProjectBriefSourceCategory = "suggested_assumption"
)

type ProjectBriefFieldID string

const (
	ProjectBriefFieldProjectType   ProjectBriefFieldID = "project_type"
	ProjectBriefFieldPurpose       ProjectBriefFieldID = "purpose"
	ProjectBriefFieldTargetUsers   ProjectBriefFieldID = "target_users"
	ProjectBriefFieldStack         ProjectBriefFieldID = "stack"
	ProjectBriefFieldArchitecture  ProjectBriefFieldID = "architecture"
	ProjectBriefFieldInstall       ProjectBriefFieldID = "install_command"
	ProjectBriefFieldTest          ProjectBriefFieldID = "test_command"
	ProjectBriefFieldBuild         ProjectBriefFieldID = "build_command"
	ProjectBriefFieldRun           ProjectBriefFieldID = "run_command"
	ProjectBriefFieldAgentBehavior ProjectBriefFieldID = "agent_behavior"
)

type ProjectBriefCandidate struct {
	Field      ProjectBriefFieldID
	Label      string
	Value      string
	Category   ProjectBriefSourceCategory
	Source     ContextSource
	Confidence ContextConfidence
}

type ProjectBriefConflict struct {
	Field          ProjectBriefFieldID
	ExistingValue  string
	CandidateValue string
	Source         ContextSource
	Confidence     ContextConfidence
}

type ProjectBriefStaleRecord struct {
	Field string
	Value string
}

type ProjectBriefFieldProposal struct {
	Field                ProjectBriefFieldID
	Label                string
	ExistingValue        string
	DetectedFacts        []ProjectBriefCandidate
	SuggestedAssumptions []ProjectBriefCandidate
	Conflicts            []ProjectBriefConflict
	Stale                bool
}

type ProjectBriefUpdateProposal struct {
	Parsed                      ParsedProjectBrief
	Fields                      []ProjectBriefFieldProposal
	StaleDetectedContext        []ProjectBriefStaleRecord
	StaleAssumptions            []ProjectBriefStaleRecord
	CurrentDetectedFacts        []ContextSignal
	CurrentSuggestedAssumptions []ContextSignal
}

type ProjectBriefMergeDecisionKind string

const (
	ProjectBriefMergeDecisionKeepExisting              ProjectBriefMergeDecisionKind = "keep_existing"
	ProjectBriefMergeDecisionReplaceWithCustom         ProjectBriefMergeDecisionKind = "replace_with_custom"
	ProjectBriefMergeDecisionAcceptDetectedFact        ProjectBriefMergeDecisionKind = "accept_detected_fact"
	ProjectBriefMergeDecisionAcceptSuggestedAssumption ProjectBriefMergeDecisionKind = "accept_suggested_assumption"
	ProjectBriefMergeDecisionIgnoreDetectedFact        ProjectBriefMergeDecisionKind = "ignore_detected_fact"
)

type ProjectBriefMergeDecision struct {
	Field ProjectBriefFieldID
	Kind  ProjectBriefMergeDecisionKind
	Value string
}

type ProjectBriefUpdateDecisions struct {
	FieldDecisions         []ProjectBriefMergeDecision
	RemoveStaleAssumptions bool
}

type ParsedProjectBrief struct {
	Answers        ProjectBriefAnswers
	ContextSources []ProjectBriefContextSource
	Assumptions    []ProjectBriefAssumption
}

func ParseProjectBriefMarkdown(contents string) (ParsedProjectBrief, error) {
	if strings.TrimSpace(contents) == "" {
		return ParsedProjectBrief{}, errors.New("project brief is empty")
	}

	lines := strings.Split(contents, "\n")
	answers := make(map[ProjectBriefFieldID]string)
	var contextSources []ProjectBriefContextSource
	var assumptions []ProjectBriefAssumption
	var currentField ProjectBriefFieldID
	inDetectedContext := false
	inAssumptions := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(trimmed, "## ") && !strings.HasPrefix(trimmed, "### "):
			inDetectedContext = false
			inAssumptions = false
			currentField = projectBriefFieldFromHeading(strings.TrimSpace(strings.TrimPrefix(trimmed, "## ")))
			if strings.EqualFold(strings.TrimSpace(strings.TrimPrefix(trimmed, "## ")), "Assumptions") {
				inAssumptions = true
			}
		case strings.HasPrefix(trimmed, "### "):
			heading := strings.TrimSpace(strings.TrimPrefix(trimmed, "### "))
			currentField = projectBriefFieldFromHeading(heading)
			inDetectedContext = strings.EqualFold(heading, "Detected context")
			inAssumptions = false
		case currentField != "" && strings.HasPrefix(trimmed, "Answer:"):
			value := strings.TrimSpace(strings.TrimPrefix(trimmed, "Answer:"))
			if value != "" {
				answers[currentField] = value
			}
		case inDetectedContext && strings.HasPrefix(trimmed, "- "):
			contextSource, ok := parseProjectBriefContextSourceLine(trimmed)
			if ok {
				contextSources = append(contextSources, contextSource)
			}
		case inAssumptions && strings.HasPrefix(trimmed, "- "):
			assumption, ok := parseProjectBriefAssumptionLine(trimmed)
			if ok {
				assumptions = append(assumptions, assumption)
			}
		}
	}

	briefAnswers, err := projectBriefAnswersFromParsedValues(answers)
	if err != nil {
		return ParsedProjectBrief{}, err
	}
	brief, err := NewProjectBrief(briefAnswers, contextSources, assumptions)
	if err != nil {
		return ParsedProjectBrief{}, err
	}

	return ParsedProjectBrief{
		Answers:        briefAnswers,
		ContextSources: brief.ContextSources(),
		Assumptions:    brief.Assumptions(),
	}, nil
}

func NewProjectBriefUpdateProposal(
	parsed ParsedProjectBrief,
	discovery ContextDiscoveryResult,
) ProjectBriefUpdateProposal {
	detectedFacts := discovery.SignalsByClassification(ContextSignalClassificationDetectedFact)
	suggestedAssumptions := discovery.SignalsByClassification(ContextSignalClassificationSuggestedAssumption)
	proposal := ProjectBriefUpdateProposal{
		Parsed:                      parsed,
		CurrentDetectedFacts:        detectedFacts,
		CurrentSuggestedAssumptions: suggestedAssumptions,
	}

	for _, field := range projectBriefFieldOrder() {
		existing := projectBriefParsedFieldValue(parsed.Answers, field)
		fieldProposal := ProjectBriefFieldProposal{
			Field:         field,
			Label:         projectBriefFieldLabel(field),
			ExistingValue: existing,
		}
		for _, signal := range detectedFacts {
			if !projectBriefSignalMatchesField(signal, field) {
				continue
			}
			candidate := projectBriefCandidateFromSignal(field, signal, ProjectBriefSourceCategoryDetectedFact)
			fieldProposal.DetectedFacts = append(fieldProposal.DetectedFacts, candidate)
			if !sameProjectBriefValue(existing, signal.Value) {
				fieldProposal.Conflicts = append(fieldProposal.Conflicts, ProjectBriefConflict{
					Field:          field,
					ExistingValue:  existing,
					CandidateValue: signal.Value,
					Source:         signal.Source,
					Confidence:     signal.Confidence,
				})
			}
		}
		for _, signal := range suggestedAssumptions {
			if !projectBriefSignalMatchesField(signal, field) {
				continue
			}
			fieldProposal.SuggestedAssumptions = append(
				fieldProposal.SuggestedAssumptions,
				projectBriefCandidateFromSignal(field, signal, ProjectBriefSourceCategorySuggestedAssumption),
			)
		}
		if len(fieldProposal.DetectedFacts) > 0 {
			fieldProposal.Stale = !projectBriefCandidatesContainValue(fieldProposal.DetectedFacts, existing)
		} else if field != ProjectBriefFieldAgentBehavior {
			fieldProposal.Stale = true
		}
		proposal.Fields = append(proposal.Fields, fieldProposal)
	}

	proposal.StaleDetectedContext = staleProjectBriefContextSources(parsed.ContextSources, detectedFacts)
	proposal.StaleAssumptions = staleProjectBriefAssumptions(parsed.Assumptions, suggestedAssumptions)
	return proposal
}

func ApplyProjectBriefUpdateDecisions(
	proposal ProjectBriefUpdateProposal,
	decisions ProjectBriefUpdateDecisions,
) (ProjectBrief, error) {
	answers := proposal.Parsed.Answers
	decisionByField := make(map[ProjectBriefFieldID]ProjectBriefMergeDecision)
	ignoredFields := make(map[ProjectBriefFieldID]bool)
	for _, decision := range decisions.FieldDecisions {
		if decision.Field == "" {
			return ProjectBrief{}, errors.New("project brief update decision field is required")
		}
		if _, ok := decisionByField[decision.Field]; ok {
			return ProjectBrief{}, fmt.Errorf("duplicate project brief update decision for %s", decision.Field)
		}
		decisionByField[decision.Field] = decision
	}

	for _, field := range projectBriefFieldOrder() {
		decision, ok := decisionByField[field]
		if !ok {
			continue
		}
		switch decision.Kind {
		case ProjectBriefMergeDecisionKeepExisting:
		case ProjectBriefMergeDecisionIgnoreDetectedFact:
			ignoredFields[field] = true
		case ProjectBriefMergeDecisionReplaceWithCustom:
			answer, err := NewUserProvidedProjectBriefAnswer(decision.Value)
			if err != nil {
				return ProjectBrief{}, err
			}
			assignProjectBriefFieldAnswer(&answers, field, answer)
		case ProjectBriefMergeDecisionAcceptDetectedFact:
			if !proposalHasCandidateValue(proposal, field, decision.Value, ProjectBriefSourceCategoryDetectedFact) {
				return ProjectBrief{}, fmt.Errorf("detected fact is not available for %s: %s", field, strings.TrimSpace(decision.Value))
			}
			answer, err := NewUserProvidedProjectBriefAnswer(decision.Value)
			if err != nil {
				return ProjectBrief{}, err
			}
			assignProjectBriefFieldAnswer(&answers, field, answer)
		case ProjectBriefMergeDecisionAcceptSuggestedAssumption:
			if !proposalHasCandidateValue(proposal, field, decision.Value, ProjectBriefSourceCategorySuggestedAssumption) {
				return ProjectBrief{}, fmt.Errorf("suggested assumption is not available for %s: %s", field, strings.TrimSpace(decision.Value))
			}
			answer, err := NewUserProvidedProjectBriefAnswer(decision.Value)
			if err != nil {
				return ProjectBrief{}, err
			}
			assignProjectBriefFieldAnswer(&answers, field, answer)
		default:
			return ProjectBrief{}, fmt.Errorf("unsupported project brief merge decision: %s", decision.Kind)
		}
	}

	contextSources := projectBriefContextSourcesForUpdate(proposal, ignoredFields)
	assumptions := projectBriefAssumptionsForUpdate(proposal, decisions)
	return NewProjectBrief(answers, contextSources, assumptions)
}

func RenderProjectBriefUpdatePreview(proposal ProjectBriefUpdateProposal, decisions ProjectBriefUpdateDecisions) string {
	var builder strings.Builder
	builder.WriteString("Project brief update preview:\n\n")
	builder.WriteString("Confirmed context:\n")
	decisionByField := make(map[ProjectBriefFieldID]ProjectBriefMergeDecision)
	for _, decision := range decisions.FieldDecisions {
		decisionByField[decision.Field] = decision
	}
	for _, field := range projectBriefFieldOrder() {
		label := projectBriefFieldLabel(field)
		existing := projectBriefParsedFieldValue(proposal.Parsed.Answers, field)
		decision, ok := decisionByField[field]
		if !ok {
			builder.WriteString("- ")
			builder.WriteString(label)
			builder.WriteString(": keep existing `")
			builder.WriteString(existing)
			builder.WriteString("`\n")
			continue
		}
		builder.WriteString("- ")
		builder.WriteString(label)
		builder.WriteString(": ")
		switch decision.Kind {
		case ProjectBriefMergeDecisionReplaceWithCustom:
			builder.WriteString("replace with custom `")
			builder.WriteString(strings.TrimSpace(decision.Value))
			builder.WriteString("`")
		case ProjectBriefMergeDecisionAcceptDetectedFact:
			builder.WriteString("replace with detected fact `")
			builder.WriteString(strings.TrimSpace(decision.Value))
			builder.WriteString("`")
		case ProjectBriefMergeDecisionAcceptSuggestedAssumption:
			builder.WriteString("replace with explicitly confirmed assumption `")
			builder.WriteString(strings.TrimSpace(decision.Value))
			builder.WriteString("`")
		case ProjectBriefMergeDecisionIgnoreDetectedFact:
			builder.WriteString("keep existing `")
			builder.WriteString(existing)
			builder.WriteString("` and ignore detected fact")
		default:
			builder.WriteString("keep existing `")
			builder.WriteString(existing)
			builder.WriteString("`")
		}
		builder.WriteByte('\n')
	}

	conflicts := proposal.Conflicts()
	builder.WriteString("\nConflicts:\n")
	if len(conflicts) == 0 {
		builder.WriteString("- none\n")
	} else {
		for _, conflict := range conflicts {
			builder.WriteString("- ")
			builder.WriteString(projectBriefFieldLabel(conflict.Field))
			builder.WriteString(": existing `")
			builder.WriteString(conflict.ExistingValue)
			builder.WriteString("`; detected `")
			builder.WriteString(conflict.CandidateValue)
			builder.WriteString("` from ")
			builder.WriteString(formatProjectBriefContextSource(conflict.Source))
			builder.WriteByte('\n')
		}
	}

	builder.WriteString("\nStale assumptions:\n")
	if len(proposal.StaleAssumptions) == 0 {
		builder.WriteString("- none\n")
	} else {
		action := "kept"
		if decisions.RemoveStaleAssumptions {
			action = "removed"
		}
		for _, stale := range proposal.StaleAssumptions {
			builder.WriteString("- ")
			builder.WriteString(action)
			builder.WriteString(": ")
			builder.WriteString(stale.Value)
			builder.WriteByte('\n')
		}
	}

	return builder.String()
}

func (proposal ProjectBriefUpdateProposal) Conflicts() []ProjectBriefConflict {
	var conflicts []ProjectBriefConflict
	for _, field := range proposal.Fields {
		conflicts = append(conflicts, field.Conflicts...)
	}
	return conflicts
}

func (category ProjectBriefSourceCategory) IsSupported() bool {
	switch category {
	case ProjectBriefSourceCategoryUserConfirmedContext,
		ProjectBriefSourceCategoryDetectedFact,
		ProjectBriefSourceCategorySuggestedAssumption:
		return true
	default:
		return false
	}
}

func projectBriefFieldFromHeading(heading string) ProjectBriefFieldID {
	switch strings.ToLower(strings.TrimSpace(heading)) {
	case "project type":
		return ProjectBriefFieldProjectType
	case "purpose":
		return ProjectBriefFieldPurpose
	case "target users":
		return ProjectBriefFieldTargetUsers
	case "stack":
		return ProjectBriefFieldStack
	case "architecture":
		return ProjectBriefFieldArchitecture
	case "install":
		return ProjectBriefFieldInstall
	case "test":
		return ProjectBriefFieldTest
	case "build":
		return ProjectBriefFieldBuild
	case "run":
		return ProjectBriefFieldRun
	case "agent behavior":
		return ProjectBriefFieldAgentBehavior
	default:
		return ""
	}
}

func projectBriefAnswersFromParsedValues(values map[ProjectBriefFieldID]string) (ProjectBriefAnswers, error) {
	var answers ProjectBriefAnswers
	for _, field := range projectBriefFieldOrder() {
		value := strings.TrimSpace(values[field])
		if value == "" {
			return ProjectBriefAnswers{}, fmt.Errorf("project brief field %s is required", projectBriefFieldLabel(field))
		}
		answer, err := NewUserProvidedProjectBriefAnswer(value)
		if err != nil {
			return ProjectBriefAnswers{}, err
		}
		assignProjectBriefFieldAnswer(&answers, field, answer)
	}
	return answers, nil
}

func parseProjectBriefContextSourceLine(line string) (ProjectBriefContextSource, bool) {
	trimmed := strings.TrimSpace(strings.TrimPrefix(line, "- "))
	const suffix = " (Source: detected context)"
	if !strings.HasSuffix(trimmed, suffix) {
		return ProjectBriefContextSource{}, false
	}
	body := strings.TrimSuffix(trimmed, suffix)
	label, value, ok := strings.Cut(body, ": ")
	if !ok {
		return ProjectBriefContextSource{}, false
	}
	contextSource, err := NewDetectedProjectBriefContextSource(label, value)
	if err != nil {
		return ProjectBriefContextSource{}, false
	}
	return contextSource, true
}

func parseProjectBriefAssumptionLine(line string) (ProjectBriefAssumption, bool) {
	trimmed := strings.TrimSpace(strings.TrimPrefix(line, "- "))
	const prefix = "Assumption: "
	const suffix = " (Source: assumption)"
	if !strings.HasPrefix(trimmed, prefix) || !strings.HasSuffix(trimmed, suffix) {
		return ProjectBriefAssumption{}, false
	}
	body := strings.TrimSuffix(strings.TrimPrefix(trimmed, prefix), suffix)
	assumption, err := NewProjectBriefAssumption(body)
	if err != nil {
		return ProjectBriefAssumption{}, false
	}
	return assumption, true
}

func projectBriefFieldOrder() []ProjectBriefFieldID {
	return []ProjectBriefFieldID{
		ProjectBriefFieldProjectType,
		ProjectBriefFieldPurpose,
		ProjectBriefFieldTargetUsers,
		ProjectBriefFieldStack,
		ProjectBriefFieldArchitecture,
		ProjectBriefFieldInstall,
		ProjectBriefFieldTest,
		ProjectBriefFieldBuild,
		ProjectBriefFieldRun,
		ProjectBriefFieldAgentBehavior,
	}
}

func projectBriefFieldLabel(field ProjectBriefFieldID) string {
	switch field {
	case ProjectBriefFieldProjectType:
		return "Project type"
	case ProjectBriefFieldPurpose:
		return "Purpose"
	case ProjectBriefFieldTargetUsers:
		return "Target users"
	case ProjectBriefFieldStack:
		return "Stack"
	case ProjectBriefFieldArchitecture:
		return "Architecture"
	case ProjectBriefFieldInstall:
		return "Install command"
	case ProjectBriefFieldTest:
		return "Test command"
	case ProjectBriefFieldBuild:
		return "Build command"
	case ProjectBriefFieldRun:
		return "Run command"
	case ProjectBriefFieldAgentBehavior:
		return "Agent behavior"
	default:
		return string(field)
	}
}

func projectBriefParsedFieldValue(answers ProjectBriefAnswers, field ProjectBriefFieldID) string {
	switch field {
	case ProjectBriefFieldProjectType:
		return answers.ProjectType.Value
	case ProjectBriefFieldPurpose:
		return answers.Purpose.Value
	case ProjectBriefFieldTargetUsers:
		return answers.TargetUsers.Value
	case ProjectBriefFieldStack:
		return answers.Stack.Value
	case ProjectBriefFieldArchitecture:
		return answers.Architecture.Value
	case ProjectBriefFieldInstall:
		return answers.Commands.Install.Value
	case ProjectBriefFieldTest:
		return answers.Commands.Test.Value
	case ProjectBriefFieldBuild:
		return answers.Commands.Build.Value
	case ProjectBriefFieldRun:
		return answers.Commands.Run.Value
	case ProjectBriefFieldAgentBehavior:
		return answers.AgentBehavior.Value
	default:
		return ""
	}
}

func assignProjectBriefFieldAnswer(answers *ProjectBriefAnswers, field ProjectBriefFieldID, answer ProjectBriefAnswer) {
	switch field {
	case ProjectBriefFieldProjectType:
		answers.ProjectType = answer
	case ProjectBriefFieldPurpose:
		answers.Purpose = answer
	case ProjectBriefFieldTargetUsers:
		answers.TargetUsers = answer
	case ProjectBriefFieldStack:
		answers.Stack = answer
	case ProjectBriefFieldArchitecture:
		answers.Architecture = answer
	case ProjectBriefFieldInstall:
		answers.Commands.Install = answer
	case ProjectBriefFieldTest:
		answers.Commands.Test = answer
	case ProjectBriefFieldBuild:
		answers.Commands.Build = answer
	case ProjectBriefFieldRun:
		answers.Commands.Run = answer
	case ProjectBriefFieldAgentBehavior:
		answers.AgentBehavior = answer
	}
}

func projectBriefSignalMatchesField(signal ContextSignal, field ProjectBriefFieldID) bool {
	switch field {
	case ProjectBriefFieldProjectType:
		return signal.Kind == ContextSignalKindProjectType
	case ProjectBriefFieldPurpose:
		return signal.Kind == ContextSignalKindPurposeSummary
	case ProjectBriefFieldTargetUsers:
		return signal.Kind == ContextSignalKindTargetUsers
	case ProjectBriefFieldStack:
		return signal.Kind == ContextSignalKindStack ||
			signal.Kind == ContextSignalKindLanguage ||
			signal.Kind == ContextSignalKindFramework
	case ProjectBriefFieldArchitecture:
		return signal.Kind == ContextSignalKindArchitectureHint
	case ProjectBriefFieldInstall:
		return signal.Kind == ContextSignalKindInstallCommand
	case ProjectBriefFieldTest:
		return signal.Kind == ContextSignalKindTestCommand
	case ProjectBriefFieldBuild:
		return signal.Kind == ContextSignalKindBuildCommand
	case ProjectBriefFieldRun:
		return signal.Kind == ContextSignalKindRunCommand
	default:
		return false
	}
}

func projectBriefCandidateFromSignal(
	field ProjectBriefFieldID,
	signal ContextSignal,
	category ProjectBriefSourceCategory,
) ProjectBriefCandidate {
	return ProjectBriefCandidate{
		Field:      field,
		Label:      signal.Kind.Label(),
		Value:      signal.Value,
		Category:   category,
		Source:     signal.Source,
		Confidence: signal.Confidence,
	}
}

func projectBriefCandidatesContainValue(candidates []ProjectBriefCandidate, value string) bool {
	for _, candidate := range candidates {
		if sameProjectBriefValue(candidate.Value, value) {
			return true
		}
	}
	return false
}

func sameProjectBriefValue(left string, right string) bool {
	return strings.EqualFold(strings.Join(strings.Fields(left), " "), strings.Join(strings.Fields(right), " "))
}

func proposalHasCandidateValue(
	proposal ProjectBriefUpdateProposal,
	field ProjectBriefFieldID,
	value string,
	category ProjectBriefSourceCategory,
) bool {
	for _, fieldProposal := range proposal.Fields {
		if fieldProposal.Field != field {
			continue
		}
		candidates := fieldProposal.DetectedFacts
		if category == ProjectBriefSourceCategorySuggestedAssumption {
			candidates = fieldProposal.SuggestedAssumptions
		}
		for _, candidate := range candidates {
			if sameProjectBriefValue(candidate.Value, value) {
				return true
			}
		}
	}
	return false
}

func projectBriefContextSourcesForUpdate(
	proposal ProjectBriefUpdateProposal,
	ignoredFields map[ProjectBriefFieldID]bool,
) []ProjectBriefContextSource {
	var contextSources []ProjectBriefContextSource
	seen := make(map[string]bool)
	add := func(contextSource ProjectBriefContextSource) {
		key := strings.ToLower(contextSource.Label + "\x00" + contextSource.Value)
		if seen[key] {
			return
		}
		seen[key] = true
		contextSources = append(contextSources, contextSource)
	}
	for _, signal := range proposal.CurrentDetectedFacts {
		field, ok := projectBriefFieldForSignal(signal)
		if ok && ignoredFields[field] {
			continue
		}
		contextSource, err := NewDetectedProjectBriefContextSource(
			signal.Kind.Label()+" from "+formatProjectBriefContextSource(signal.Source),
			signal.Value,
		)
		if err == nil {
			add(contextSource)
		}
	}
	for _, contextSource := range proposal.Parsed.ContextSources {
		add(contextSource)
	}
	sort.SliceStable(contextSources, func(i, j int) bool {
		if contextSources[i].Label != contextSources[j].Label {
			return contextSources[i].Label < contextSources[j].Label
		}
		return contextSources[i].Value < contextSources[j].Value
	})
	return contextSources
}

func projectBriefAssumptionsForUpdate(
	proposal ProjectBriefUpdateProposal,
	decisions ProjectBriefUpdateDecisions,
) []ProjectBriefAssumption {
	acceptedAssumptions := make(map[string]bool)
	for _, decision := range decisions.FieldDecisions {
		if decision.Kind == ProjectBriefMergeDecisionAcceptSuggestedAssumption {
			acceptedAssumptions[strings.ToLower(strings.TrimSpace(decision.Value))] = true
		}
	}

	var assumptions []ProjectBriefAssumption
	seen := make(map[string]bool)
	add := func(assumption ProjectBriefAssumption) {
		key := strings.ToLower(strings.TrimSpace(assumption.Description))
		if key == "" || seen[key] || acceptedAssumptions[key] {
			return
		}
		seen[key] = true
		assumptions = append(assumptions, assumption)
	}
	stale := make(map[string]bool)
	for _, record := range proposal.StaleAssumptions {
		stale[strings.ToLower(strings.TrimSpace(record.Value))] = true
	}
	for _, signal := range proposal.CurrentSuggestedAssumptions {
		if acceptedAssumptions[strings.ToLower(strings.TrimSpace(signal.Value))] {
			continue
		}
		assumption, err := NewProjectBriefAssumption(
			signal.Kind.Label() + ": " + signal.Value + " (Source: " + formatProjectBriefContextSource(signal.Source) + ")",
		)
		if err == nil {
			add(assumption)
		}
	}
	for _, assumption := range proposal.Parsed.Assumptions {
		if decisions.RemoveStaleAssumptions && stale[strings.ToLower(strings.TrimSpace(assumption.Description))] {
			continue
		}
		add(assumption)
	}
	sort.SliceStable(assumptions, func(i, j int) bool {
		return assumptions[i].Description < assumptions[j].Description
	})
	return assumptions
}

func projectBriefFieldForSignal(signal ContextSignal) (ProjectBriefFieldID, bool) {
	for _, field := range projectBriefFieldOrder() {
		if projectBriefSignalMatchesField(signal, field) {
			return field, true
		}
	}
	return "", false
}

func staleProjectBriefContextSources(
	existing []ProjectBriefContextSource,
	currentDetected []ContextSignal,
) []ProjectBriefStaleRecord {
	current := make(map[string]bool)
	for _, signal := range currentDetected {
		current[strings.ToLower(strings.TrimSpace(signal.Value))] = true
	}
	var stale []ProjectBriefStaleRecord
	for _, contextSource := range existing {
		if current[strings.ToLower(strings.TrimSpace(contextSource.Value))] {
			continue
		}
		stale = append(stale, ProjectBriefStaleRecord{Field: contextSource.Label, Value: contextSource.Value})
	}
	return stale
}

func staleProjectBriefAssumptions(
	existing []ProjectBriefAssumption,
	currentAssumptions []ContextSignal,
) []ProjectBriefStaleRecord {
	currentDescriptions := make(map[string]bool)
	currentValues := make(map[string]bool)
	for _, signal := range currentAssumptions {
		currentValues[strings.ToLower(strings.TrimSpace(signal.Value))] = true
		currentDescriptions[strings.ToLower(signal.Kind.Label()+": "+signal.Value+" (Source: "+formatProjectBriefContextSource(signal.Source)+")")] = true
	}
	var stale []ProjectBriefStaleRecord
	for _, assumption := range existing {
		normalized := strings.ToLower(strings.TrimSpace(assumption.Description))
		if currentDescriptions[normalized] || currentValues[normalized] {
			continue
		}
		stale = append(stale, ProjectBriefStaleRecord{Field: "Assumption", Value: assumption.Description})
	}
	return stale
}

func formatProjectBriefContextSource(source ContextSource) string {
	if source.Evidence == "" {
		return source.Path
	}
	return source.Path + " (" + source.Evidence + ")"
}
