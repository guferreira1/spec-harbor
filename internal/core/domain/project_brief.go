package domain

import (
	"errors"
	"fmt"
	"strings"
)

const ProjectBriefCustomOptionLabel = "Other / custom"

type ProjectBriefQuestionID string

const (
	ProjectBriefQuestionProjectType   ProjectBriefQuestionID = "project_type"
	ProjectBriefQuestionPurpose       ProjectBriefQuestionID = "purpose"
	ProjectBriefQuestionTargetUsers   ProjectBriefQuestionID = "target_users"
	ProjectBriefQuestionStack         ProjectBriefQuestionID = "stack"
	ProjectBriefQuestionArchitecture  ProjectBriefQuestionID = "architecture"
	ProjectBriefQuestionInstall       ProjectBriefQuestionID = "install_command"
	ProjectBriefQuestionTest          ProjectBriefQuestionID = "test_command"
	ProjectBriefQuestionBuild         ProjectBriefQuestionID = "build_command"
	ProjectBriefQuestionRun           ProjectBriefQuestionID = "run_command"
	ProjectBriefQuestionAgentBehavior ProjectBriefQuestionID = "agent_behavior"
)

type ProjectBriefAnswerSource string

const (
	ProjectBriefAnswerSourceUserProvided    ProjectBriefAnswerSource = "user_provided"
	ProjectBriefAnswerSourceDetectedContext ProjectBriefAnswerSource = "detected_context"
	ProjectBriefAnswerSourceAssumption      ProjectBriefAnswerSource = "assumption"
)

type ProjectBriefOption struct {
	Label string
}

type ProjectBriefQuestion struct {
	ID      ProjectBriefQuestionID
	Prompt  string
	Options []ProjectBriefOption
}

type ProjectBriefAnswer struct {
	Value  string
	Source ProjectBriefAnswerSource
}

type ProjectBriefCommandSet struct {
	Install ProjectBriefAnswer
	Test    ProjectBriefAnswer
	Build   ProjectBriefAnswer
	Run     ProjectBriefAnswer
}

type ProjectBriefAnswers struct {
	ProjectType   ProjectBriefAnswer
	Purpose       ProjectBriefAnswer
	TargetUsers   ProjectBriefAnswer
	Stack         ProjectBriefAnswer
	Architecture  ProjectBriefAnswer
	Commands      ProjectBriefCommandSet
	AgentBehavior ProjectBriefAnswer
}

type ProjectBriefContextSource struct {
	Label  string
	Value  string
	Source ProjectBriefAnswerSource
}

type ProjectBriefAssumption struct {
	Description string
	Source      ProjectBriefAnswerSource
}

type ProjectBrief struct {
	ProjectType   ProjectBriefAnswer
	Purpose       ProjectBriefAnswer
	TargetUsers   ProjectBriefAnswer
	Stack         ProjectBriefAnswer
	Architecture  ProjectBriefAnswer
	Commands      ProjectBriefCommandSet
	AgentBehavior ProjectBriefAnswer

	contextSources []ProjectBriefContextSource
	assumptions    []ProjectBriefAssumption
}

func DefaultProjectBriefQuestions() []ProjectBriefQuestion {
	return []ProjectBriefQuestion{
		newProjectBriefQuestion(ProjectBriefQuestionProjectType, "What type of project is this?", []string{
			"Backend API",
			"Full-stack web application",
			"CLI/tooling project",
			"Library/package",
			ProjectBriefCustomOptionLabel,
		}),
		newProjectBriefQuestion(ProjectBriefQuestionPurpose, "What is the primary purpose of this project?", []string{
			"Customer-facing product",
			"Internal operations tool",
			"Developer productivity tool",
			"Data or automation pipeline",
			ProjectBriefCustomOptionLabel,
		}),
		newProjectBriefQuestion(ProjectBriefQuestionTargetUsers, "Who are the target users?", []string{
			"External customers",
			"Internal business users",
			"Developers or platform engineers",
			"Operators or support teams",
			ProjectBriefCustomOptionLabel,
		}),
		newProjectBriefQuestion(ProjectBriefQuestionStack, "What stack should agents assume only after confirmation?", []string{
			"Go",
			"Node.js / TypeScript",
			"Python",
			"Multi-stack or monorepo",
			ProjectBriefCustomOptionLabel,
		}),
		newProjectBriefQuestion(ProjectBriefQuestionArchitecture, "What architecture should agents preserve?", []string{
			"MVC / layered architecture",
			"Clean Architecture / Hexagonal",
			"DDD-oriented modules",
			"Simple modular structure",
			ProjectBriefCustomOptionLabel,
		}),
		newProjectBriefQuestion(ProjectBriefQuestionInstall, "What install command should agents use?", []string{
			"No install command",
			"npm install",
			"pnpm install",
			"go mod download",
			ProjectBriefCustomOptionLabel,
		}),
		newProjectBriefQuestion(ProjectBriefQuestionTest, "What test command should agents use?", []string{
			"No test command",
			"go test ./...",
			"npm test",
			"pytest",
			ProjectBriefCustomOptionLabel,
		}),
		newProjectBriefQuestion(ProjectBriefQuestionBuild, "What build command should agents use?", []string{
			"No build command",
			"go build ./...",
			"npm run build",
			"make build",
			ProjectBriefCustomOptionLabel,
		}),
		newProjectBriefQuestion(ProjectBriefQuestionRun, "What run command should agents use?", []string{
			"No run command",
			"go run ./cmd/<name>",
			"npm run dev",
			"make run",
			ProjectBriefCustomOptionLabel,
		}),
		newProjectBriefQuestion(ProjectBriefQuestionAgentBehavior, "What should agents do when context is missing?", []string{
			"Ask before assuming",
			"Suggest assumptions and ask for confirmation",
			"Proceed with best-effort suggestions",
			ProjectBriefCustomOptionLabel,
		}),
	}
}

func newProjectBriefQuestion(id ProjectBriefQuestionID, prompt string, optionLabels []string) ProjectBriefQuestion {
	options := make([]ProjectBriefOption, 0, len(optionLabels))
	for _, label := range optionLabels {
		options = append(options, ProjectBriefOption{Label: label})
	}
	return ProjectBriefQuestion{
		ID:      id,
		Prompt:  prompt,
		Options: options,
	}
}

func (question ProjectBriefQuestion) Validate() error {
	if question.ID == "" {
		return errors.New("project brief question id is required")
	}
	if strings.TrimSpace(question.Prompt) == "" {
		return fmt.Errorf("project brief question %s prompt is required", question.ID)
	}
	if len(question.Options) < 3 || len(question.Options) > 5 {
		return fmt.Errorf("project brief question %s must have three to five options", question.ID)
	}
	for index, option := range question.Options {
		if strings.TrimSpace(option.Label) == "" {
			return fmt.Errorf("project brief question %s option %d is required", question.ID, index+1)
		}
	}
	if question.Options[len(question.Options)-1].Label != ProjectBriefCustomOptionLabel {
		return fmt.Errorf("project brief question %s last option must be %q", question.ID, ProjectBriefCustomOptionLabel)
	}
	return nil
}

func NewProjectBriefAnswer(value string, source ProjectBriefAnswerSource) (ProjectBriefAnswer, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ProjectBriefAnswer{}, errors.New("project brief answer value is required")
	}
	if !source.IsSupported() {
		return ProjectBriefAnswer{}, fmt.Errorf("unsupported project brief answer source: %s", source)
	}
	return ProjectBriefAnswer{Value: trimmed, Source: source}, nil
}

func NewUserProvidedProjectBriefAnswer(value string) (ProjectBriefAnswer, error) {
	return NewProjectBriefAnswer(value, ProjectBriefAnswerSourceUserProvided)
}

func (source ProjectBriefAnswerSource) IsSupported() bool {
	switch source {
	case ProjectBriefAnswerSourceUserProvided, ProjectBriefAnswerSourceDetectedContext, ProjectBriefAnswerSourceAssumption:
		return true
	default:
		return false
	}
}

func (source ProjectBriefAnswerSource) Label() string {
	switch source {
	case ProjectBriefAnswerSourceUserProvided:
		return "user-provided answer"
	case ProjectBriefAnswerSourceDetectedContext:
		return "detected context"
	case ProjectBriefAnswerSourceAssumption:
		return "assumption"
	default:
		return string(source)
	}
}

func NewDetectedProjectBriefContextSource(label string, value string) (ProjectBriefContextSource, error) {
	trimmedLabel := strings.TrimSpace(label)
	if trimmedLabel == "" {
		return ProjectBriefContextSource{}, errors.New("project brief context source label is required")
	}
	trimmedValue := strings.TrimSpace(value)
	if trimmedValue == "" {
		return ProjectBriefContextSource{}, errors.New("project brief context source value is required")
	}
	return ProjectBriefContextSource{
		Label:  trimmedLabel,
		Value:  trimmedValue,
		Source: ProjectBriefAnswerSourceDetectedContext,
	}, nil
}

func NewProjectBriefAssumption(description string) (ProjectBriefAssumption, error) {
	trimmed := strings.TrimSpace(description)
	if trimmed == "" {
		return ProjectBriefAssumption{}, errors.New("project brief assumption description is required")
	}
	return ProjectBriefAssumption{
		Description: trimmed,
		Source:      ProjectBriefAnswerSourceAssumption,
	}, nil
}

func NewProjectBrief(
	answers ProjectBriefAnswers,
	contextSources []ProjectBriefContextSource,
	assumptions []ProjectBriefAssumption,
) (ProjectBrief, error) {
	if err := validateConfirmedAnswer("project type", answers.ProjectType); err != nil {
		return ProjectBrief{}, err
	}
	if err := validateConfirmedAnswer("purpose", answers.Purpose); err != nil {
		return ProjectBrief{}, err
	}
	if err := validateConfirmedAnswer("target users", answers.TargetUsers); err != nil {
		return ProjectBrief{}, err
	}
	if err := validateConfirmedAnswer("stack", answers.Stack); err != nil {
		return ProjectBrief{}, err
	}
	if err := validateConfirmedAnswer("architecture", answers.Architecture); err != nil {
		return ProjectBrief{}, err
	}
	if err := validateConfirmedAnswer("install command", answers.Commands.Install); err != nil {
		return ProjectBrief{}, err
	}
	if err := validateConfirmedAnswer("test command", answers.Commands.Test); err != nil {
		return ProjectBrief{}, err
	}
	if err := validateConfirmedAnswer("build command", answers.Commands.Build); err != nil {
		return ProjectBrief{}, err
	}
	if err := validateConfirmedAnswer("run command", answers.Commands.Run); err != nil {
		return ProjectBrief{}, err
	}
	if err := validateConfirmedAnswer("agent behavior", answers.AgentBehavior); err != nil {
		return ProjectBrief{}, err
	}
	if err := validateContextSources(contextSources); err != nil {
		return ProjectBrief{}, err
	}
	if err := validateAssumptions(assumptions); err != nil {
		return ProjectBrief{}, err
	}

	return ProjectBrief{
		ProjectType:    answers.ProjectType,
		Purpose:        answers.Purpose,
		TargetUsers:    answers.TargetUsers,
		Stack:          answers.Stack,
		Architecture:   answers.Architecture,
		Commands:       answers.Commands,
		AgentBehavior:  answers.AgentBehavior,
		contextSources: append([]ProjectBriefContextSource(nil), contextSources...),
		assumptions:    append([]ProjectBriefAssumption(nil), assumptions...),
	}, nil
}

func validateConfirmedAnswer(name string, answer ProjectBriefAnswer) error {
	if strings.TrimSpace(answer.Value) == "" {
		return fmt.Errorf("%s is required", name)
	}
	if answer.Source != ProjectBriefAnswerSourceUserProvided {
		return fmt.Errorf("%s must be a user-provided answer", name)
	}
	return nil
}

func validateContextSources(contextSources []ProjectBriefContextSource) error {
	for index, contextSource := range contextSources {
		if strings.TrimSpace(contextSource.Label) == "" {
			return fmt.Errorf("detected context source %d label is required", index+1)
		}
		if strings.TrimSpace(contextSource.Value) == "" {
			return fmt.Errorf("detected context source %d value is required", index+1)
		}
		if contextSource.Source != ProjectBriefAnswerSourceDetectedContext {
			return fmt.Errorf("detected context source %d must use detected context source", index+1)
		}
	}
	return nil
}

func validateAssumptions(assumptions []ProjectBriefAssumption) error {
	for index, assumption := range assumptions {
		if strings.TrimSpace(assumption.Description) == "" {
			return fmt.Errorf("assumption %d description is required", index+1)
		}
		if assumption.Source != ProjectBriefAnswerSourceAssumption {
			return fmt.Errorf("assumption %d must use assumption source", index+1)
		}
	}
	return nil
}

func (brief ProjectBrief) ContextSources() []ProjectBriefContextSource {
	return append([]ProjectBriefContextSource(nil), brief.contextSources...)
}

func (brief ProjectBrief) Assumptions() []ProjectBriefAssumption {
	return append([]ProjectBriefAssumption(nil), brief.assumptions...)
}

func (brief ProjectBrief) RenderMarkdown() string {
	var builder strings.Builder
	builder.WriteString("# Project Brief\n\n")
	renderAnswerSection(&builder, "Project type", brief.ProjectType)
	renderAnswerSection(&builder, "Purpose", brief.Purpose)
	renderAnswerSection(&builder, "Target users", brief.TargetUsers)
	renderAnswerSection(&builder, "Stack", brief.Stack)
	renderAnswerSection(&builder, "Architecture", brief.Architecture)

	builder.WriteString("## Commands\n\n")
	renderCommandAnswer(&builder, "Install", brief.Commands.Install)
	renderCommandAnswer(&builder, "Test", brief.Commands.Test)
	renderCommandAnswer(&builder, "Build", brief.Commands.Build)
	renderCommandAnswer(&builder, "Run", brief.Commands.Run)

	renderAnswerSection(&builder, "Agent behavior", brief.AgentBehavior)

	builder.WriteString("## Context sources\n\n")
	builder.WriteString("### User-provided answers\n\n")
	renderUserProvidedAnswerList(&builder, brief)
	builder.WriteString("\n### Detected context\n\n")
	if len(brief.contextSources) == 0 {
		builder.WriteString("None recorded.\n\n")
	} else {
		for _, contextSource := range brief.contextSources {
			builder.WriteString("- ")
			builder.WriteString(contextSource.Label)
			builder.WriteString(": ")
			builder.WriteString(contextSource.Value)
			builder.WriteString(" (Source: ")
			builder.WriteString(contextSource.Source.Label())
			builder.WriteString(")\n")
		}
		builder.WriteString("\n")
	}

	builder.WriteString("## Assumptions\n\n")
	if len(brief.assumptions) == 0 {
		builder.WriteString("None recorded.\n")
	} else {
		for _, assumption := range brief.assumptions {
			builder.WriteString("- Assumption: ")
			builder.WriteString(assumption.Description)
			builder.WriteString(" (Source: ")
			builder.WriteString(assumption.Source.Label())
			builder.WriteString(")\n")
		}
	}

	return builder.String()
}

func renderAnswerSection(builder *strings.Builder, title string, answer ProjectBriefAnswer) {
	builder.WriteString("## ")
	builder.WriteString(title)
	builder.WriteString("\n\n")
	builder.WriteString("Answer: ")
	builder.WriteString(answer.Value)
	builder.WriteString("\n\n")
	builder.WriteString("Source: ")
	builder.WriteString(answer.Source.Label())
	builder.WriteString("\n\n")
}

func renderCommandAnswer(builder *strings.Builder, title string, answer ProjectBriefAnswer) {
	builder.WriteString("### ")
	builder.WriteString(title)
	builder.WriteString("\n\n")
	builder.WriteString("Answer: ")
	builder.WriteString(answer.Value)
	builder.WriteString("\n\n")
	builder.WriteString("Source: ")
	builder.WriteString(answer.Source.Label())
	builder.WriteString("\n\n")
}

func renderUserProvidedAnswerList(builder *strings.Builder, brief ProjectBrief) {
	items := []struct {
		label  string
		answer ProjectBriefAnswer
	}{
		{label: "Project type", answer: brief.ProjectType},
		{label: "Purpose", answer: brief.Purpose},
		{label: "Target users", answer: brief.TargetUsers},
		{label: "Stack", answer: brief.Stack},
		{label: "Architecture", answer: brief.Architecture},
		{label: "Install command", answer: brief.Commands.Install},
		{label: "Test command", answer: brief.Commands.Test},
		{label: "Build command", answer: brief.Commands.Build},
		{label: "Run command", answer: brief.Commands.Run},
		{label: "Agent behavior", answer: brief.AgentBehavior},
	}

	for _, item := range items {
		builder.WriteString("- ")
		builder.WriteString(item.label)
		builder.WriteString(": ")
		builder.WriteString(item.answer.Value)
		builder.WriteString(" (Source: ")
		builder.WriteString(item.answer.Source.Label())
		builder.WriteString(")\n")
	}
}
