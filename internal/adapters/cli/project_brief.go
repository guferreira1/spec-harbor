package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/guferreira1/spec-harbor/internal/adapters/filesystem"
	"github.com/guferreira1/spec-harbor/internal/core/domain"
	"github.com/guferreira1/spec-harbor/internal/core/usecase"
)

func briefCommand(ctx CommandContext) error {
	arguments, err := parseBriefArguments(ctx.Args)
	if err != nil {
		return err
	}

	root, err := os.Getwd()
	if err != nil {
		return err
	}

	terminal := ctx.terminal
	if terminal == nil {
		terminal = newOSInteractiveTerminal(ctx.Output)
	}
	if !terminal.IsInputTerminal() {
		return errors.New("brief requires an interactive TTY")
	}
	if arguments.update {
		return briefUpdateCommand(ctx.Output, terminal, root)
	}

	discoveredContext := discoverProjectBriefSuggestionContext(root)

	answers, err := promptProjectBriefAnswers(terminal, discoveredContext)
	if err != nil {
		return err
	}
	if err := printProjectBriefWriteSummary(terminal); err != nil {
		return err
	}
	confirmed, err := promptProjectBriefConfirmation(terminal)
	if err != nil {
		return err
	}
	if !confirmed {
		return errInteractiveOperationCancelled
	}

	fileSystem := filesystem.NewLocalFileSystem()
	createBrief := usecase.NewCreateProjectBrief(fileSystem)
	result, err := createBrief.Execute(usecase.CreateProjectBriefInput{
		ProjectRoot:    root,
		Answers:        answers,
		ContextSources: contextSourcesFromDiscovery(discoveredContext),
		Assumptions:    assumptionsFromDiscovery(discoveredContext, answers),
	})
	if err != nil {
		return err
	}

	printProjectBriefReport(ctx.Output, result)
	return nil
}

type briefArguments struct {
	update bool
}

func parseBriefArguments(args []string) (briefArguments, error) {
	var arguments briefArguments
	for _, arg := range args {
		if arg == "--update" {
			if arguments.update {
				return briefArguments{}, fmt.Errorf("duplicate flag: %s", arg)
			}
			arguments.update = true
			continue
		}
		if strings.HasPrefix(arg, "-") {
			return briefArguments{}, fmt.Errorf("unsupported flag: %s", arg)
		}
		return briefArguments{}, fmt.Errorf("unexpected argument: %s", arg)
	}
	return arguments, nil
}

func briefUpdateCommand(output io.Writer, terminal interactiveTerminal, root string) error {
	fileSystem := filesystem.NewLocalFileSystem()
	contextFileSystem := filesystem.NewContextDiscoveryFileSystem()
	discoverContext := usecase.NewDiscoverProjectContext(contextFileSystem)
	updateBrief := usecase.NewUpdateProjectBrief(fileSystem, discoverContext)

	plan, err := updateBrief.Prepare(usecase.PrepareProjectBriefUpdateInput{ProjectRoot: root})
	if err != nil {
		return err
	}
	if err := printProjectBriefUpdateSummary(terminal, plan); err != nil {
		return err
	}
	decisions, err := promptProjectBriefUpdateDecisions(terminal, plan.Proposal)
	if err != nil {
		return err
	}

	preview, err := updateBrief.Execute(usecase.UpdateProjectBriefInput{
		ProjectRoot: root,
		Decisions:   decisions,
		Confirmed:   false,
	})
	if err != nil {
		return err
	}
	if err := terminal.WriteString("\n" + preview.Preview + "\n"); err != nil {
		return err
	}
	confirmed, err := promptProjectBriefUpdateConfirmation(terminal)
	if err != nil {
		return err
	}
	if !confirmed {
		return errInteractiveOperationCancelled
	}

	result, err := updateBrief.Execute(usecase.UpdateProjectBriefInput{
		ProjectRoot: root,
		Decisions:   decisions,
		Confirmed:   true,
	})
	if err != nil {
		return err
	}
	printProjectBriefUpdateReport(output, result)
	return nil
}

func promptProjectBriefAnswers(
	terminal interactiveTerminal,
	discoveredContext domain.ContextDiscoveryResult,
) (domain.ProjectBriefAnswers, error) {
	if terminal == nil {
		return domain.ProjectBriefAnswers{}, errors.New("interactive terminal is required")
	}
	if !terminal.IsInputTerminal() {
		return domain.ProjectBriefAnswers{}, errors.New("brief requires an interactive TTY")
	}
	if err := terminal.WriteString("SpecHarbor could not find enough confirmed project context.\n\n"); err != nil {
		return domain.ProjectBriefAnswers{}, err
	}

	var answers domain.ProjectBriefAnswers
	for _, question := range domain.DefaultProjectBriefQuestions() {
		question = projectBriefQuestionWithContextSuggestions(question, discoveredContext)
		if err := question.Validate(); err != nil {
			return domain.ProjectBriefAnswers{}, err
		}
		answer, err := promptProjectBriefQuestion(terminal, question)
		if err != nil {
			return domain.ProjectBriefAnswers{}, err
		}
		if err := assignProjectBriefAnswer(&answers, question.ID, answer); err != nil {
			return domain.ProjectBriefAnswers{}, err
		}
	}
	return answers, nil
}

func discoverProjectBriefSuggestionContext(root string) domain.ContextDiscoveryResult {
	fileSystem := filesystem.NewContextDiscoveryFileSystem()
	discoverContext := usecase.NewDiscoverProjectContext(fileSystem)
	result, err := discoverContext.Execute(usecase.DiscoverProjectContextInput{ProjectRoot: root})
	if err != nil {
		return domain.NewContextDiscoveryResult(nil, nil)
	}
	return result
}

func projectBriefQuestionWithContextSuggestions(
	question domain.ProjectBriefQuestion,
	discoveredContext domain.ContextDiscoveryResult,
) domain.ProjectBriefQuestion {
	kinds := projectBriefSuggestionKinds(question.ID)
	if len(kinds) == 0 {
		return question
	}

	suggestions := projectBriefSuggestionValues(discoveredContext, kinds)
	if len(suggestions) == 0 {
		return question
	}

	optionLabels := make([]string, 0, 5)
	seen := make(map[string]bool)
	addOption := func(label string) {
		trimmed := strings.TrimSpace(label)
		if trimmed == "" || trimmed == domain.ProjectBriefCustomOptionLabel || seen[trimmed] || len(optionLabels) >= 4 {
			return
		}
		seen[trimmed] = true
		optionLabels = append(optionLabels, trimmed)
	}

	for _, suggestion := range suggestions {
		addOption(suggestion)
	}
	for _, option := range question.Options {
		addOption(option.Label)
	}
	optionLabels = append(optionLabels, domain.ProjectBriefCustomOptionLabel)

	options := make([]domain.ProjectBriefOption, 0, len(optionLabels))
	for _, label := range optionLabels {
		options = append(options, domain.ProjectBriefOption{Label: label})
	}
	question.Options = options
	return question
}

func projectBriefSuggestionKinds(questionID domain.ProjectBriefQuestionID) []domain.ContextSignalKind {
	switch questionID {
	case domain.ProjectBriefQuestionProjectType:
		return []domain.ContextSignalKind{domain.ContextSignalKindProjectType}
	case domain.ProjectBriefQuestionPurpose:
		return []domain.ContextSignalKind{domain.ContextSignalKindPurposeSummary}
	case domain.ProjectBriefQuestionTargetUsers:
		return []domain.ContextSignalKind{domain.ContextSignalKindTargetUsers}
	case domain.ProjectBriefQuestionStack:
		return []domain.ContextSignalKind{
			domain.ContextSignalKindStack,
			domain.ContextSignalKindLanguage,
			domain.ContextSignalKindFramework,
		}
	case domain.ProjectBriefQuestionArchitecture:
		return []domain.ContextSignalKind{domain.ContextSignalKindArchitectureHint}
	case domain.ProjectBriefQuestionInstall:
		return []domain.ContextSignalKind{domain.ContextSignalKindInstallCommand}
	case domain.ProjectBriefQuestionTest:
		return []domain.ContextSignalKind{domain.ContextSignalKindTestCommand}
	case domain.ProjectBriefQuestionBuild:
		return []domain.ContextSignalKind{domain.ContextSignalKindBuildCommand}
	case domain.ProjectBriefQuestionRun:
		return []domain.ContextSignalKind{domain.ContextSignalKindRunCommand}
	default:
		return nil
	}
}

func projectBriefSuggestionValues(
	discoveredContext domain.ContextDiscoveryResult,
	kinds []domain.ContextSignalKind,
) []string {
	kindSet := make(map[domain.ContextSignalKind]bool)
	for _, kind := range kinds {
		kindSet[kind] = true
	}

	var values []string
	seen := make(map[string]bool)
	for _, classification := range []domain.ContextSignalClassification{
		domain.ContextSignalClassificationUserConfirmedContext,
		domain.ContextSignalClassificationDetectedFact,
		domain.ContextSignalClassificationSuggestedAssumption,
	} {
		for _, signal := range discoveredContext.SignalsByClassification(classification) {
			if !kindSet[signal.Kind] || seen[signal.Value] {
				continue
			}
			seen[signal.Value] = true
			values = append(values, signal.Value)
		}
	}
	return values
}

func contextSourcesFromDiscovery(discoveredContext domain.ContextDiscoveryResult) []domain.ProjectBriefContextSource {
	var contextSources []domain.ProjectBriefContextSource
	for _, signal := range discoveredContext.SignalsByClassification(domain.ContextSignalClassificationDetectedFact) {
		contextSource, err := domain.NewDetectedProjectBriefContextSource(
			signal.Kind.Label()+" from "+formatContextSource(signal.Source),
			signal.Value,
		)
		if err != nil {
			continue
		}
		contextSources = append(contextSources, contextSource)
	}
	return contextSources
}

func assumptionsFromDiscovery(
	discoveredContext domain.ContextDiscoveryResult,
	answers domain.ProjectBriefAnswers,
) []domain.ProjectBriefAssumption {
	var assumptions []domain.ProjectBriefAssumption
	confirmedValues := confirmedProjectBriefAnswerValues(answers)
	for _, signal := range discoveredContext.SignalsByClassification(domain.ContextSignalClassificationSuggestedAssumption) {
		if confirmedValues[signal.Value] {
			continue
		}
		assumption, err := domain.NewProjectBriefAssumption(
			signal.Kind.Label() + ": " + signal.Value + " (Source: " + formatContextSource(signal.Source) + ")",
		)
		if err != nil {
			continue
		}
		assumptions = append(assumptions, assumption)
	}
	return assumptions
}

func confirmedProjectBriefAnswerValues(answers domain.ProjectBriefAnswers) map[string]bool {
	values := make(map[string]bool)
	for _, answer := range []domain.ProjectBriefAnswer{
		answers.ProjectType,
		answers.Purpose,
		answers.TargetUsers,
		answers.Stack,
		answers.Architecture,
		answers.Commands.Install,
		answers.Commands.Test,
		answers.Commands.Build,
		answers.Commands.Run,
		answers.AgentBehavior,
	} {
		values[answer.Value] = true
	}
	return values
}

func promptProjectBriefQuestion(
	terminal interactiveTerminal,
	question domain.ProjectBriefQuestion,
) (domain.ProjectBriefAnswer, error) {
	for attempt := 0; attempt < interactivePromptMaxAttempts; attempt++ {
		if err := terminal.WriteString(projectBriefQuestionMenu(question)); err != nil {
			return domain.ProjectBriefAnswer{}, err
		}
		value, err := terminal.ReadLine()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return domain.ProjectBriefAnswer{}, errInteractiveOperationCancelled
			}
			return domain.ProjectBriefAnswer{}, err
		}

		option, custom, err := parseProjectBriefQuestionChoice(question, value)
		if err == nil {
			if custom {
				return promptProjectBriefCustomAnswer(terminal, question.ID)
			}
			return domain.NewUserProvidedProjectBriefAnswer(option.Label)
		}
		if err := terminal.WriteString(fmt.Sprintf("Invalid answer: %s\n", err.Error())); err != nil {
			return domain.ProjectBriefAnswer{}, err
		}
	}
	return domain.ProjectBriefAnswer{}, fmt.Errorf("%s retry limit exceeded", projectBriefQuestionRetryLabel(question.ID))
}

func promptProjectBriefCustomAnswer(
	terminal interactiveTerminal,
	questionID domain.ProjectBriefQuestionID,
) (domain.ProjectBriefAnswer, error) {
	for attempt := 0; attempt < interactivePromptMaxAttempts; attempt++ {
		if err := terminal.WriteString("Custom answer: "); err != nil {
			return domain.ProjectBriefAnswer{}, err
		}
		value, err := terminal.ReadLine()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return domain.ProjectBriefAnswer{}, errInteractiveOperationCancelled
			}
			return domain.ProjectBriefAnswer{}, err
		}

		answer, err := domain.NewUserProvidedProjectBriefAnswer(value)
		if err == nil {
			return answer, nil
		}
		if err := terminal.WriteString("Invalid answer: custom answer is required\n"); err != nil {
			return domain.ProjectBriefAnswer{}, err
		}
	}
	return domain.ProjectBriefAnswer{}, fmt.Errorf("%s custom answer retry limit exceeded", projectBriefQuestionRetryLabel(questionID))
}

func projectBriefQuestionMenu(question domain.ProjectBriefQuestion) string {
	var builder strings.Builder
	builder.WriteString(question.Prompt)
	builder.WriteByte('\n')
	for index, option := range question.Options {
		builder.WriteString(strconv.Itoa(index + 1))
		builder.WriteString(". ")
		builder.WriteString(option.Label)
		builder.WriteByte('\n')
	}
	builder.WriteString("Choice: ")
	return builder.String()
}

func parseProjectBriefQuestionChoice(
	question domain.ProjectBriefQuestion,
	value string,
) (domain.ProjectBriefOption, bool, error) {
	normalized := strings.ToLower(strings.TrimSpace(value))
	if normalized == "" {
		return domain.ProjectBriefOption{}, false, errors.New("choice is required")
	}

	if choice, err := strconv.Atoi(normalized); err == nil {
		if choice < 1 || choice > len(question.Options) {
			return domain.ProjectBriefOption{}, false, fmt.Errorf("unsupported choice: %s", strings.TrimSpace(value))
		}
		option := question.Options[choice-1]
		return option, choice == len(question.Options), nil
	}

	for index, option := range question.Options {
		if strings.ToLower(option.Label) == normalized {
			return option, index == len(question.Options)-1, nil
		}
	}
	if normalized == "other" || normalized == "custom" {
		return question.Options[len(question.Options)-1], true, nil
	}

	return domain.ProjectBriefOption{}, false, fmt.Errorf("unsupported choice: %s", strings.TrimSpace(value))
}

func assignProjectBriefAnswer(
	answers *domain.ProjectBriefAnswers,
	questionID domain.ProjectBriefQuestionID,
	answer domain.ProjectBriefAnswer,
) error {
	switch questionID {
	case domain.ProjectBriefQuestionProjectType:
		answers.ProjectType = answer
	case domain.ProjectBriefQuestionPurpose:
		answers.Purpose = answer
	case domain.ProjectBriefQuestionTargetUsers:
		answers.TargetUsers = answer
	case domain.ProjectBriefQuestionStack:
		answers.Stack = answer
	case domain.ProjectBriefQuestionArchitecture:
		answers.Architecture = answer
	case domain.ProjectBriefQuestionInstall:
		answers.Commands.Install = answer
	case domain.ProjectBriefQuestionTest:
		answers.Commands.Test = answer
	case domain.ProjectBriefQuestionBuild:
		answers.Commands.Build = answer
	case domain.ProjectBriefQuestionRun:
		answers.Commands.Run = answer
	case domain.ProjectBriefQuestionAgentBehavior:
		answers.AgentBehavior = answer
	default:
		return fmt.Errorf("unsupported project brief question: %s", questionID)
	}
	return nil
}

func projectBriefQuestionRetryLabel(questionID domain.ProjectBriefQuestionID) string {
	switch questionID {
	case domain.ProjectBriefQuestionProjectType:
		return "project type"
	case domain.ProjectBriefQuestionPurpose:
		return "purpose"
	case domain.ProjectBriefQuestionTargetUsers:
		return "target users"
	case domain.ProjectBriefQuestionStack:
		return "stack"
	case domain.ProjectBriefQuestionArchitecture:
		return "architecture"
	case domain.ProjectBriefQuestionInstall:
		return "install command"
	case domain.ProjectBriefQuestionTest:
		return "test command"
	case domain.ProjectBriefQuestionBuild:
		return "build command"
	case domain.ProjectBriefQuestionRun:
		return "run command"
	case domain.ProjectBriefQuestionAgentBehavior:
		return "agent behavior"
	default:
		return "project brief question"
	}
}

func printProjectBriefWriteSummary(terminal interactiveTerminal) error {
	lines := []string{
		"",
		"SpecHarbor will create:",
		"",
		".specharbor/project-brief.md",
		"",
		"Safety:",
		"- Stack, architecture, and commands come from confirmed answers only.",
		"- Detected context remains separate from user answers.",
		"- Assumptions are not confirmed facts.",
		"- No repository indexing, RAG, provider API, agent execution, source-control automation, release, or publishing behavior will run.",
		"",
	}
	return terminal.WriteString(strings.Join(lines, "\n"))
}

func promptProjectBriefConfirmation(terminal interactiveTerminal) (bool, error) {
	for attempt := 0; attempt < interactivePromptMaxAttempts; attempt++ {
		if err := terminal.WriteString("Confirm? [y/N]: "); err != nil {
			return false, err
		}
		value, err := terminal.ReadLine()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return false, nil
			}
			return false, err
		}

		normalized := strings.ToLower(strings.TrimSpace(value))
		switch normalized {
		case "y", "yes":
			return true, nil
		case "", "n", "no":
			return false, nil
		default:
			if err := terminal.WriteString("Invalid answer: enter y/yes or n/no.\n"); err != nil {
				return false, err
			}
		}
	}
	return false, errors.New("confirmation retry limit exceeded")
}

func printProjectBriefReport(output io.Writer, result usecase.CreateProjectBriefResult) {
	directoryStatus := "existing"
	if result.DirectoryCreated {
		directoryStatus = "created"
	}

	fmt.Fprintln(output, "SpecHarbor project brief created.")
	fmt.Fprintf(output, "Path: %s\n", result.TargetPath)
	fmt.Fprintf(output, "Directory: %s\n", directoryStatus)
	fmt.Fprintf(output, "File written: %s\n", yesNo(result.FileWritten))
}

func printProjectBriefUpdateSummary(terminal interactiveTerminal, plan usecase.ProjectBriefUpdatePlan) error {
	detectedCount, assumptionCount := projectBriefUpdateCandidateCounts(plan.Proposal)
	staleCount := len(plan.Proposal.StaleDetectedContext) + len(plan.Proposal.StaleAssumptions)
	lines := []string{
		"SpecHarbor will update:",
		"",
		plan.TargetPath,
		"",
		fmt.Sprintf("Confirmed fields: %d", len(plan.Proposal.Fields)),
		fmt.Sprintf("Detected fact candidates: %d", detectedCount),
		fmt.Sprintf("Suggested assumption candidates: %d", assumptionCount),
		fmt.Sprintf("Conflicts: %d", len(plan.Proposal.Conflicts())),
		fmt.Sprintf("Stale records: %d", staleCount),
		"",
		"Safety:",
		"- Existing confirmed values are kept unless you explicitly replace them.",
		"- Detected facts remain separate unless you explicitly accept one.",
		"- Suggested assumptions are not confirmed facts unless you explicitly accept one.",
		"- No repository indexing, RAG, provider API, agent execution, source-control automation, release, or publishing behavior will run.",
		"",
	}
	return terminal.WriteString(strings.Join(lines, "\n"))
}

func projectBriefUpdateCandidateCounts(proposal domain.ProjectBriefUpdateProposal) (int, int) {
	var detected int
	var assumptions int
	for _, field := range proposal.Fields {
		detected += len(field.DetectedFacts)
		assumptions += len(field.SuggestedAssumptions)
	}
	return detected, assumptions
}

func promptProjectBriefUpdateDecisions(
	terminal interactiveTerminal,
	proposal domain.ProjectBriefUpdateProposal,
) (domain.ProjectBriefUpdateDecisions, error) {
	var decisions domain.ProjectBriefUpdateDecisions
	for _, field := range proposal.Fields {
		if !projectBriefUpdateFieldNeedsPrompt(field) {
			continue
		}
		decision, err := promptProjectBriefUpdateFieldDecision(terminal, field)
		if err != nil {
			return domain.ProjectBriefUpdateDecisions{}, err
		}
		decisions.FieldDecisions = append(decisions.FieldDecisions, decision)
	}
	if len(proposal.StaleAssumptions) > 0 {
		remove, err := promptProjectBriefStaleAssumptionDecision(terminal, proposal.StaleAssumptions)
		if err != nil {
			return domain.ProjectBriefUpdateDecisions{}, err
		}
		decisions.RemoveStaleAssumptions = remove
	}
	return decisions, nil
}

func projectBriefUpdateFieldNeedsPrompt(field domain.ProjectBriefFieldProposal) bool {
	return len(field.Conflicts) > 0 ||
		len(field.DetectedFacts) > 0 ||
		len(field.SuggestedAssumptions) > 0 ||
		field.Stale
}

type projectBriefUpdateChoice struct {
	label    string
	decision domain.ProjectBriefMergeDecision
	custom   bool
	cancel   bool
}

func promptProjectBriefUpdateFieldDecision(
	terminal interactiveTerminal,
	field domain.ProjectBriefFieldProposal,
) (domain.ProjectBriefMergeDecision, error) {
	choices := projectBriefUpdateChoices(field)
	for attempt := 0; attempt < interactivePromptMaxAttempts; attempt++ {
		if err := terminal.WriteString(projectBriefUpdateFieldMenu(field, choices)); err != nil {
			return domain.ProjectBriefMergeDecision{}, err
		}
		value, err := terminal.ReadLine()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return domain.ProjectBriefMergeDecision{}, errInteractiveOperationCancelled
			}
			return domain.ProjectBriefMergeDecision{}, err
		}

		choice, err := parseProjectBriefUpdateChoice(value, choices)
		if err == nil {
			if choice.cancel {
				return domain.ProjectBriefMergeDecision{}, errInteractiveOperationCancelled
			}
			if choice.custom {
				customValue, err := promptProjectBriefUpdateCustomValue(terminal)
				if err != nil {
					return domain.ProjectBriefMergeDecision{}, err
				}
				choice.decision.Value = customValue
			}
			return choice.decision, nil
		}
		if err := terminal.WriteString(fmt.Sprintf("Invalid answer: %s\n", err.Error())); err != nil {
			return domain.ProjectBriefMergeDecision{}, err
		}
	}
	return domain.ProjectBriefMergeDecision{}, fmt.Errorf("%s update retry limit exceeded", strings.ToLower(field.Label))
}

func projectBriefUpdateChoices(field domain.ProjectBriefFieldProposal) []projectBriefUpdateChoice {
	choices := []projectBriefUpdateChoice{
		{
			label: "Keep existing value",
			decision: domain.ProjectBriefMergeDecision{
				Field: field.Field,
				Kind:  domain.ProjectBriefMergeDecisionKeepExisting,
			},
		},
		{
			label: "Enter custom replacement",
			decision: domain.ProjectBriefMergeDecision{
				Field: field.Field,
				Kind:  domain.ProjectBriefMergeDecisionReplaceWithCustom,
			},
			custom: true,
		},
	}
	if len(field.DetectedFacts) > 0 {
		choices = append(choices, projectBriefUpdateChoice{
			label: "Ignore detected facts for this field",
			decision: domain.ProjectBriefMergeDecision{
				Field: field.Field,
				Kind:  domain.ProjectBriefMergeDecisionIgnoreDetectedFact,
			},
		})
	}
	for _, candidate := range field.DetectedFacts {
		choices = append(choices, projectBriefUpdateChoice{
			label: "Accept detected fact: " + candidate.Value + " from " + formatContextSource(candidate.Source),
			decision: domain.ProjectBriefMergeDecision{
				Field: field.Field,
				Kind:  domain.ProjectBriefMergeDecisionAcceptDetectedFact,
				Value: candidate.Value,
			},
		})
	}
	for _, candidate := range field.SuggestedAssumptions {
		choices = append(choices, projectBriefUpdateChoice{
			label: "Accept suggested assumption: " + candidate.Value + " from " + formatContextSource(candidate.Source),
			decision: domain.ProjectBriefMergeDecision{
				Field: field.Field,
				Kind:  domain.ProjectBriefMergeDecisionAcceptSuggestedAssumption,
				Value: candidate.Value,
			},
		})
	}
	choices = append(choices, projectBriefUpdateChoice{label: "Cancel update", cancel: true})
	return choices
}

func projectBriefUpdateFieldMenu(
	field domain.ProjectBriefFieldProposal,
	choices []projectBriefUpdateChoice,
) string {
	var builder strings.Builder
	builder.WriteString("\nUpdate ")
	builder.WriteString(field.Label)
	builder.WriteByte('\n')
	builder.WriteString("Current: ")
	builder.WriteString(field.ExistingValue)
	builder.WriteByte('\n')
	if len(field.Conflicts) > 0 {
		builder.WriteString("Conflicts:\n")
		for _, conflict := range field.Conflicts {
			builder.WriteString("- Detected ")
			builder.WriteString(conflict.CandidateValue)
			builder.WriteString(" from ")
			builder.WriteString(formatContextSource(conflict.Source))
			builder.WriteByte('\n')
		}
	}
	if field.Stale {
		builder.WriteString("Potentially stale: yes\n")
	}
	for index, choice := range choices {
		builder.WriteString(strconv.Itoa(index + 1))
		builder.WriteString(". ")
		builder.WriteString(choice.label)
		builder.WriteByte('\n')
	}
	builder.WriteString("Choice: ")
	return builder.String()
}

func parseProjectBriefUpdateChoice(
	value string,
	choices []projectBriefUpdateChoice,
) (projectBriefUpdateChoice, error) {
	normalized := strings.ToLower(strings.TrimSpace(value))
	if normalized == "" || normalized == "keep" || normalized == "k" {
		return choices[0], nil
	}
	if normalized == "custom" {
		for _, choice := range choices {
			if choice.custom {
				return choice, nil
			}
		}
	}
	if normalized == "cancel" || normalized == "q" {
		return projectBriefUpdateChoice{cancel: true}, nil
	}
	if choice, err := strconv.Atoi(normalized); err == nil {
		if choice < 1 || choice > len(choices) {
			return projectBriefUpdateChoice{}, fmt.Errorf("unsupported choice: %s", strings.TrimSpace(value))
		}
		return choices[choice-1], nil
	}
	return projectBriefUpdateChoice{}, fmt.Errorf("unsupported choice: %s", strings.TrimSpace(value))
}

func promptProjectBriefUpdateCustomValue(terminal interactiveTerminal) (string, error) {
	for attempt := 0; attempt < interactivePromptMaxAttempts; attempt++ {
		if err := terminal.WriteString("Custom replacement: "); err != nil {
			return "", err
		}
		value, err := terminal.ReadLine()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return "", errInteractiveOperationCancelled
			}
			return "", err
		}
		answer, err := domain.NewUserProvidedProjectBriefAnswer(value)
		if err == nil {
			return answer.Value, nil
		}
		if err := terminal.WriteString("Invalid answer: custom replacement is required\n"); err != nil {
			return "", err
		}
	}
	return "", errors.New("custom replacement retry limit exceeded")
}

func promptProjectBriefStaleAssumptionDecision(
	terminal interactiveTerminal,
	stale []domain.ProjectBriefStaleRecord,
) (bool, error) {
	for attempt := 0; attempt < interactivePromptMaxAttempts; attempt++ {
		if err := terminal.WriteString(projectBriefStaleAssumptionMenu(stale)); err != nil {
			return false, err
		}
		value, err := terminal.ReadLine()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return false, errInteractiveOperationCancelled
			}
			return false, err
		}
		switch strings.ToLower(strings.TrimSpace(value)) {
		case "", "1", "keep", "k":
			return false, nil
		case "2", "remove", "r":
			return true, nil
		case "3", "cancel", "q":
			return false, errInteractiveOperationCancelled
		default:
			if err := terminal.WriteString("Invalid answer: enter keep, remove, or cancel.\n"); err != nil {
				return false, err
			}
		}
	}
	return false, errors.New("stale assumptions retry limit exceeded")
}

func projectBriefStaleAssumptionMenu(stale []domain.ProjectBriefStaleRecord) string {
	var builder strings.Builder
	builder.WriteString("\nStale assumptions detected:\n")
	for _, record := range stale {
		builder.WriteString("- ")
		builder.WriteString(record.Value)
		builder.WriteByte('\n')
	}
	builder.WriteString("1. Keep stale assumptions\n")
	builder.WriteString("2. Remove stale assumptions\n")
	builder.WriteString("3. Cancel update\n")
	builder.WriteString("Choice: ")
	return builder.String()
}

func promptProjectBriefUpdateConfirmation(terminal interactiveTerminal) (bool, error) {
	for attempt := 0; attempt < interactivePromptMaxAttempts; attempt++ {
		if err := terminal.WriteString("Write updated project brief? [y/N]: "); err != nil {
			return false, err
		}
		value, err := terminal.ReadLine()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return false, nil
			}
			return false, err
		}

		normalized := strings.ToLower(strings.TrimSpace(value))
		switch normalized {
		case "y", "yes":
			return true, nil
		case "", "n", "no":
			return false, nil
		default:
			if err := terminal.WriteString("Invalid answer: enter y/yes or n/no.\n"); err != nil {
				return false, err
			}
		}
	}
	return false, errors.New("confirmation retry limit exceeded")
}

func printProjectBriefUpdateReport(output io.Writer, result usecase.UpdateProjectBriefResult) {
	fmt.Fprintln(output, "SpecHarbor project brief updated.")
	fmt.Fprintf(output, "Path: %s\n", result.TargetPath)
	fmt.Fprintf(output, "File written: %s\n", yesNo(result.FileWritten))
}
