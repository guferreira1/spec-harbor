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
	if err := parseBriefArguments(ctx.Args); err != nil {
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

	answers, err := promptProjectBriefAnswers(terminal)
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
		ProjectRoot: root,
		Answers:     answers,
	})
	if err != nil {
		return err
	}

	printProjectBriefReport(ctx.Output, result)
	return nil
}

func parseBriefArguments(args []string) error {
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			return fmt.Errorf("unsupported flag: %s", arg)
		}
		return fmt.Errorf("unexpected argument: %s", arg)
	}
	return nil
}

func promptProjectBriefAnswers(terminal interactiveTerminal) (domain.ProjectBriefAnswers, error) {
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
