package cli

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/guferreira1/spec-harbor/internal/adapters/filesystem"
	"github.com/guferreira1/spec-harbor/internal/adapters/templates"
	"github.com/guferreira1/spec-harbor/internal/core/usecase"
	"github.com/guferreira1/spec-harbor/internal/platform/version"
)

type CommandContext struct {
	Args   []string
	Output io.Writer
}

type CommandHandler func(ctx CommandContext) error

func Execute(args []string) error {
	return execute(args, os.Stdout)
}

func execute(args []string, output io.Writer) error {
	if len(args) == 0 {
		printHelp(output)
		return nil
	}

	commands := commandRegistry()

	handler, exists := commands[args[0]]
	if !exists {
		return fmt.Errorf("unknown command: %s", args[0])
	}

	return handler(CommandContext{
		Args:   args[1:],
		Output: output,
	})
}

func commandRegistry() map[string]CommandHandler {
	return map[string]CommandHandler{
		"version":  versionCommand,
		"init":     initCommand,
		"scan":     notImplementedCommand("scan"),
		"generate": generateCommand,
		"prompt":   promptCommand,
		"validate": notImplementedCommand("validate"),
		"review":   notImplementedCommand("review"),
		"archive":  notImplementedCommand("archive"),
		"config":   notImplementedCommand("config"),

		"help":   helpCommand,
		"-h":     helpCommand,
		"--help": helpCommand,
	}
}

func versionCommand(ctx CommandContext) error {
	fmt.Fprintln(ctx.Output, version.Version)
	return nil
}

func initCommand(ctx CommandContext) error {
	root, err := os.Getwd()
	if err != nil {
		return err
	}

	fileSystem := filesystem.NewLocalFileSystem()
	defaults := templates.NewDefaultInitializationTemplates()
	initializer := usecase.NewInitializeProject(fileSystem, defaults)

	result, err := initializer.Execute(usecase.InitializeProjectInput{Root: root})
	if err != nil {
		return err
	}

	if result.Status == usecase.InitializationStatusAlreadyInitialized {
		fmt.Fprintln(ctx.Output, "SpecHarbor already initialized.")
		return nil
	}

	fmt.Fprintln(ctx.Output, "SpecHarbor initialized.")
	fmt.Fprintf(ctx.Output, "Created: %d\n", len(result.Created))
	fmt.Fprintf(ctx.Output, "Skipped existing: %d\n", len(result.Skipped))

	return nil
}

func generateCommand(ctx CommandContext) error {
	fmt.Fprintf(ctx.Output, "specharbor generate: %s\n", strings.Join(ctx.Args, " "))
	return nil
}

type promptArguments struct {
	changeID string
	role     string
}

func promptCommand(ctx CommandContext) error {
	arguments, err := parsePromptArguments(ctx.Args)
	if err != nil {
		return err
	}

	root, err := os.Getwd()
	if err != nil {
		return err
	}

	roleTemplates := templates.NewRolePromptTemplates()
	renderer := templates.NewPromptTemplateRenderer()
	renderPrompt := usecase.NewRenderPrompt(roleTemplates, renderer)

	result, err := renderPrompt.Execute(usecase.RenderPromptInput{
		ProjectRoot: root,
		ChangeID:    arguments.changeID,
		Role:        arguments.role,
	})
	if err != nil {
		return err
	}

	fmt.Fprint(ctx.Output, result.Prompt)
	return nil
}

func parsePromptArguments(args []string) (promptArguments, error) {
	var positionals []string
	var role string
	roleProvided := false

	for index := 0; index < len(args); index++ {
		arg := args[index]
		if arg == "--role" {
			if roleProvided {
				return promptArguments{}, fmt.Errorf("prompt role flag specified more than once")
			}
			if index+1 >= len(args) {
				return promptArguments{}, fmt.Errorf("prompt role value is required")
			}
			if strings.HasPrefix(args[index+1], "-") {
				return promptArguments{}, fmt.Errorf("prompt role value is required")
			}

			role = args[index+1]
			roleProvided = true
			index++
			continue
		}

		if strings.HasPrefix(arg, "-") {
			return promptArguments{}, fmt.Errorf("unsupported flag: %s", arg)
		}

		positionals = append(positionals, arg)
	}

	if len(positionals) == 0 {
		return promptArguments{}, fmt.Errorf("change id is required")
	}
	if len(positionals) > 1 {
		return promptArguments{}, fmt.Errorf("unexpected argument: %s", positionals[1])
	}
	if !roleProvided || strings.TrimSpace(role) == "" {
		return promptArguments{}, fmt.Errorf("prompt role is required")
	}

	return promptArguments{
		changeID: positionals[0],
		role:     role,
	}, nil
}

func helpCommand(ctx CommandContext) error {
	printHelp(ctx.Output)
	return nil
}

func notImplementedCommand(command string) CommandHandler {
	return func(ctx CommandContext) error {
		fmt.Fprintf(ctx.Output, "specharbor %s: not implemented yet\n", command)
		return nil
	}
}

func printHelp(output io.Writer) {
	fmt.Fprintln(output, `SpecHarbor

Usage:
  specharbor <command> [arguments]

Commands:
  init        Initialize OpenSpec and SpecHarbor configuration
  scan        Detect project stack and development context
  generate    Generate a new OpenSpec change
  prompt      Generate an agent-specific implementation prompt
  validate    Validate an OpenSpec change
  review      Review implementation against a spec
  archive     Archive a completed OpenSpec change
  config      Manage SpecHarbor configuration
  version     Print SpecHarbor version
  help        Show this help message`)
}
