package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/guferreira1/spec-harbor/internal/core/domain"
)

const interactivePromptMaxAttempts = 3

var errInteractiveOperationCancelled = errors.New("operation cancelled")

type interactiveTerminal interface {
	IsInputTerminal() bool
	ReadLine() (string, error)
	WriteString(string) error
}

type osInteractiveTerminal struct {
	input  *os.File
	output io.Writer
	reader *bufio.Reader
}

func newOSInteractiveTerminal(output io.Writer) interactiveTerminal {
	return &osInteractiveTerminal{
		input:  os.Stdin,
		output: output,
		reader: bufio.NewReader(os.Stdin),
	}
}

func (terminal *osInteractiveTerminal) IsInputTerminal() bool {
	if terminal == nil || terminal.input == nil {
		return false
	}
	return isInputTerminalFile(terminal.input)
}

func (terminal *osInteractiveTerminal) ReadLine() (string, error) {
	if terminal == nil || terminal.reader == nil {
		return "", io.EOF
	}
	line, err := terminal.reader.ReadString('\n')
	if err != nil {
		if errors.Is(err, io.EOF) && line != "" {
			return trimLineBreak(line), nil
		}
		return "", err
	}
	return trimLineBreak(line), nil
}

func (terminal *osInteractiveTerminal) WriteString(value string) error {
	if terminal == nil || terminal.output == nil {
		return errors.New("interactive output is required")
	}
	_, err := io.WriteString(terminal.output, value)
	return err
}

func trimLineBreak(value string) string {
	return strings.TrimSuffix(strings.TrimSuffix(value, "\n"), "\r")
}

type interactiveGenerationPath string

const (
	interactivePathBlank          interactiveGenerationPath = "blank"
	interactivePathBuiltIn        interactiveGenerationPath = "built-in template"
	interactivePathCustomTemplate interactiveGenerationPath = "custom template"
	interactivePathConfigTemplate interactiveGenerationPath = "config template"
	interactivePathHybrid         interactiveGenerationPath = "hybrid"
)

type interactiveHybridSource string

const (
	interactiveHybridSourceBuiltIn interactiveHybridSource = "built-in template"
	interactiveHybridSourceCustom  interactiveHybridSource = "custom template"
	interactiveHybridSourceConfig  interactiveHybridSource = "config template"
)

type interactiveGenerationSelection struct {
	changeID            string
	path                interactiveGenerationPath
	hybridSource        interactiveHybridSource
	templateName        string
	customTemplateName  string
	configTemplateAlias string
	title               string
	summary             string
	generationType      string
}

func promptInteractiveGeneration(terminal interactiveTerminal, changeID string) (generateArguments, error) {
	if terminal == nil {
		return generateArguments{}, errors.New("interactive terminal is required")
	}
	if !terminal.IsInputTerminal() {
		return generateArguments{}, errors.New("interactive mode requires a TTY")
	}

	selection := interactiveGenerationSelection{changeID: changeID}
	path, err := promptGenerationPath(terminal)
	if err != nil {
		return generateArguments{}, err
	}
	selection.path = path

	switch path {
	case interactivePathBlank:
	case interactivePathBuiltIn:
		templateName, err := promptBuiltInTemplateName(terminal)
		if err != nil {
			return generateArguments{}, err
		}
		selection.templateName = templateName
	case interactivePathCustomTemplate:
		customTemplateName, err := promptCustomTemplateName(terminal)
		if err != nil {
			return generateArguments{}, err
		}
		selection.customTemplateName = customTemplateName
		selection.title, err = promptOptionalValue(terminal, "Title (optional): ")
		if err != nil {
			return generateArguments{}, err
		}
		selection.summary, err = promptOptionalValue(terminal, "Summary (optional): ")
		if err != nil {
			return generateArguments{}, err
		}
	case interactivePathConfigTemplate:
		configTemplateAlias, err := promptConfigTemplateAlias(terminal)
		if err != nil {
			return generateArguments{}, err
		}
		selection.configTemplateAlias = configTemplateAlias
		selection.title, err = promptOptionalValue(terminal, "Title (optional): ")
		if err != nil {
			return generateArguments{}, err
		}
		selection.summary, err = promptOptionalValue(terminal, "Summary (optional): ")
		if err != nil {
			return generateArguments{}, err
		}
	case interactivePathHybrid:
		if err := promptHybridValues(terminal, &selection); err != nil {
			return generateArguments{}, err
		}
	default:
		return generateArguments{}, fmt.Errorf("unsupported interactive generation path: %s", path)
	}

	if err := printInteractiveGenerationSummary(terminal, selection); err != nil {
		return generateArguments{}, err
	}
	confirmed, err := promptInteractiveConfirmation(terminal)
	if err != nil {
		return generateArguments{}, err
	}
	if !confirmed {
		return generateArguments{}, errInteractiveOperationCancelled
	}

	return interactiveSelectionToGenerateArguments(selection), nil
}

func promptGenerationPath(terminal interactiveTerminal) (interactiveGenerationPath, error) {
	menu := "Select generation path:\n" +
		"1. blank\n" +
		"2. built-in template\n" +
		"3. custom template\n" +
		"4. config template\n" +
		"5. hybrid\n" +
		"Choice: "
	return promptRequiredValue(terminal, menu, parseInteractiveGenerationPath, "generation path retry limit exceeded")
}

func parseInteractiveGenerationPath(value string) (interactiveGenerationPath, error) {
	normalized := strings.ToLower(strings.TrimSpace(value))
	switch normalized {
	case "1", "blank":
		return interactivePathBlank, nil
	case "2", "built-in template", "built-in", "builtin", "template":
		return interactivePathBuiltIn, nil
	case "3", "custom template", "custom":
		return interactivePathCustomTemplate, nil
	case "4", "config template", "config":
		return interactivePathConfigTemplate, nil
	case "5", "hybrid":
		return interactivePathHybrid, nil
	case "":
		return "", errors.New("generation path is required")
	default:
		return "", fmt.Errorf("unsupported generation path: %s", strings.TrimSpace(value))
	}
}

func promptBuiltInTemplateName(terminal interactiveTerminal) (string, error) {
	menu := "Select built-in template:\n" +
		"1. feature\n" +
		"2. bugfix\n" +
		"3. docs\n" +
		"4. refactor\n" +
		"Choice: "
	return promptRequiredValue(terminal, menu, parseBuiltInTemplateChoice, "built-in template retry limit exceeded")
}

func parseBuiltInTemplateChoice(value string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(value))
	switch normalized {
	case "1":
		normalized = string(domain.FeatureTemplate)
	case "2":
		normalized = string(domain.BugfixTemplate)
	case "3":
		normalized = string(domain.DocsTemplate)
	case "4":
		normalized = string(domain.RefactorTemplate)
	}
	templateName, err := domain.ParseTemplateName(normalized)
	if err != nil {
		return "", err
	}
	return string(templateName), nil
}

func promptCustomTemplateName(terminal interactiveTerminal) (string, error) {
	return promptRequiredValue(terminal, "Custom template name: ", func(value string) (string, error) {
		customTemplateName, err := domain.NewCustomTemplateName(value)
		if err != nil {
			return "", err
		}
		return customTemplateName.String(), nil
	}, "custom template retry limit exceeded")
}

func promptConfigTemplateAlias(terminal interactiveTerminal) (string, error) {
	return promptRequiredValue(terminal, "Config template alias: ", func(value string) (string, error) {
		configTemplateAlias, err := domain.NewConfigTemplateAlias(value)
		if err != nil {
			return "", err
		}
		return configTemplateAlias.String(), nil
	}, "config template retry limit exceeded")
}

func promptOptionalValue(terminal interactiveTerminal, prompt string) (string, error) {
	if err := terminal.WriteString(prompt); err != nil {
		return "", err
	}
	value, err := terminal.ReadLine()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return "", errInteractiveOperationCancelled
		}
		return "", err
	}
	return strings.TrimSpace(value), nil
}

func promptHybridValues(terminal interactiveTerminal, selection *interactiveGenerationSelection) error {
	source, err := promptHybridSource(terminal)
	if err != nil {
		return err
	}
	selection.hybridSource = source

	switch source {
	case interactiveHybridSourceBuiltIn:
		templateName, err := promptBuiltInTemplateName(terminal)
		if err != nil {
			return err
		}
		selection.templateName = templateName
	case interactiveHybridSourceCustom:
		customTemplateName, err := promptCustomTemplateName(terminal)
		if err != nil {
			return err
		}
		selection.customTemplateName = customTemplateName
	case interactiveHybridSourceConfig:
		configTemplateAlias, err := promptConfigTemplateAlias(terminal)
		if err != nil {
			return err
		}
		selection.configTemplateAlias = configTemplateAlias
	default:
		return fmt.Errorf("unsupported hybrid source: %s", source)
	}

	title, err := promptRequiredValue(terminal, "Title: ", func(value string) (string, error) {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			return "", errors.New("hybrid title is required")
		}
		return trimmed, nil
	}, "hybrid title retry limit exceeded")
	if err != nil {
		return err
	}
	selection.title = title

	summary, err := promptRequiredValue(terminal, "Summary: ", func(value string) (string, error) {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			return "", errors.New("hybrid summary is required")
		}
		return trimmed, nil
	}, "hybrid summary retry limit exceeded")
	if err != nil {
		return err
	}
	selection.summary = summary

	generationType, err := promptOptionalHybridType(terminal)
	if err != nil {
		return err
	}
	selection.generationType = generationType

	if _, err := domain.NewHybridMetadata(selection.title, selection.summary, selection.generationType); err != nil {
		return err
	}
	if _, err := domain.NewHybridSourceSelection(selection.templateName, selection.customTemplateName, selection.configTemplateAlias); err != nil {
		return err
	}
	return nil
}

func promptHybridSource(terminal interactiveTerminal) (interactiveHybridSource, error) {
	menu := "Select hybrid source:\n" +
		"1. built-in template\n" +
		"2. custom template\n" +
		"3. config template\n" +
		"Choice: "
	return promptRequiredValue(terminal, menu, parseHybridSourceChoice, "hybrid source retry limit exceeded")
}

func parseHybridSourceChoice(value string) (interactiveHybridSource, error) {
	normalized := strings.ToLower(strings.TrimSpace(value))
	switch normalized {
	case "1", "built-in template", "built-in", "builtin", "template":
		return interactiveHybridSourceBuiltIn, nil
	case "2", "custom template", "custom":
		return interactiveHybridSourceCustom, nil
	case "3", "config template", "config":
		return interactiveHybridSourceConfig, nil
	case "":
		return "", errors.New("hybrid source is required")
	default:
		return "", fmt.Errorf("unsupported hybrid source: %s", strings.TrimSpace(value))
	}
}

func promptOptionalHybridType(terminal interactiveTerminal) (string, error) {
	return promptRequiredValue(terminal, "Type (optional): ", func(value string) (string, error) {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			return "", nil
		}
		hybridType, err := domain.ParseHybridType(trimmed)
		if err != nil {
			return "", err
		}
		return string(hybridType), nil
	}, "hybrid type retry limit exceeded")
}

func promptRequiredValue[T any](
	terminal interactiveTerminal,
	prompt string,
	parse func(string) (T, error),
	exhaustedMessage string,
) (T, error) {
	var zero T
	for attempt := 0; attempt < interactivePromptMaxAttempts; attempt++ {
		if err := terminal.WriteString(prompt); err != nil {
			return zero, err
		}
		value, err := terminal.ReadLine()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return zero, errInteractiveOperationCancelled
			}
			return zero, err
		}

		parsed, err := parse(value)
		if err == nil {
			return parsed, nil
		}
		if err := terminal.WriteString(fmt.Sprintf("Invalid answer: %s\n", err.Error())); err != nil {
			return zero, err
		}
	}
	return zero, errors.New(exhaustedMessage)
}

func printInteractiveGenerationSummary(terminal interactiveTerminal, selection interactiveGenerationSelection) error {
	lines := []string{
		"",
		"Interactive generation summary:",
		"Change: " + selection.changeID,
		"Generation path: " + string(selection.path),
	}

	if selection.path == interactivePathHybrid {
		lines = append(lines, "Hybrid source: "+string(selection.hybridSource))
	}
	if selection.templateName != "" {
		lines = append(lines, "Template: "+selection.templateName)
	}
	if selection.customTemplateName != "" {
		lines = append(lines, "Custom template: "+selection.customTemplateName)
	}
	if selection.configTemplateAlias != "" {
		lines = append(lines, "Config alias: "+selection.configTemplateAlias)
	}
	if selection.title != "" {
		lines = append(lines, "Title: "+selection.title)
	}
	if selection.summary != "" {
		lines = append(lines, "Summary: "+selection.summary)
	}
	if selection.generationType != "" {
		lines = append(lines, "Type: "+selection.generationType)
	}

	lines = append(lines,
		"Expected write target: openspec/changes/"+selection.changeID+"/",
		"Files: proposal.md, design.md, tasks.md, acceptance-criteria.md, risks.md",
		"Validation: automatic "+interactiveValidationBehavior(selection.path),
		"Safety:",
		"- Writes are limited to OpenSpec change files.",
		"- Production code will not be modified.",
		"- Source-control commands will not be run.",
		"- Workflow automation will not be triggered.",
		"- Provider, LLM, and agent APIs will not be called.",
		"- No auto-commit, auto-push, PR creation, merge, or archive will be performed.",
		"",
	)

	return terminal.WriteString(strings.Join(lines, "\n"))
}

func interactiveValidationBehavior(path interactiveGenerationPath) string {
	if path == interactivePathHybrid {
		return "yes"
	}
	return "no"
}

func promptInteractiveConfirmation(terminal interactiveTerminal) (bool, error) {
	for attempt := 0; attempt < interactivePromptMaxAttempts; attempt++ {
		if err := terminal.WriteString("Proceed? [y/N]: "); err != nil {
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

func interactiveSelectionToGenerateArguments(selection interactiveGenerationSelection) generateArguments {
	arguments := generateArguments{
		changeID: selection.changeID,
		title:    selection.title,
		summary:  selection.summary,
	}

	switch selection.path {
	case interactivePathBlank:
		arguments.mode = domain.BlankMode
	case interactivePathBuiltIn:
		arguments.mode = domain.TemplateMode
		arguments.templateName = selection.templateName
	case interactivePathCustomTemplate:
		arguments.mode = domain.TemplateMode
		arguments.customTemplate = true
		arguments.customTemplateName = selection.customTemplateName
	case interactivePathConfigTemplate:
		arguments.mode = domain.TemplateMode
		arguments.configTemplate = true
		arguments.configTemplateAlias = selection.configTemplateAlias
	case interactivePathHybrid:
		arguments.mode = domain.HybridMode
		arguments.templateName = selection.templateName
		arguments.customTemplate = selection.customTemplateName != ""
		arguments.customTemplateName = selection.customTemplateName
		arguments.configTemplate = selection.configTemplateAlias != ""
		arguments.configTemplateAlias = selection.configTemplateAlias
		arguments.guidedType = selection.generationType
	}

	return arguments
}
