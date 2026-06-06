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
		"prompt":   notImplementedCommand("prompt"),
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
