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

func Execute(args []string) error {
	return execute(args, os.Stdout)
}

func execute(args []string, output io.Writer) error {
	if len(args) == 0 {
		printHelp(output)
		return nil
	}

	switch args[0] {
	case "version":
		fmt.Fprintln(output, version.Version)
		return nil
	case "init":
		return initializeProject(output)
	case "scan":
		fmt.Fprintln(output, "specharbor scan: not implemented yet")
		return nil
	case "generate":
		fmt.Fprintf(output, "specharbor generate: %s\n", strings.Join(args[1:], " "))
		return nil
	case "prompt":
		fmt.Fprintln(output, "specharbor prompt: not implemented yet")
		return nil
	case "validate":
		fmt.Fprintln(output, "specharbor validate: not implemented yet")
		return nil
	case "review":
		fmt.Fprintln(output, "specharbor review: not implemented yet")
		return nil
	case "archive":
		fmt.Fprintln(output, "specharbor archive: not implemented yet")
		return nil
	case "config":
		fmt.Fprintln(output, "specharbor config: not implemented yet")
		return nil
	case "help", "-h", "--help":
		printHelp(output)
		return nil
	default:
		return fmt.Errorf("unknown command: %s", args[0])
	}
}

func initializeProject(output io.Writer) error {
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
		fmt.Fprintln(output, "SpecHarbor already initialized.")
		return nil
	}

	fmt.Fprintln(output, "SpecHarbor initialized.")
	fmt.Fprintf(output, "Created: %d\n", len(result.Created))
	fmt.Fprintf(output, "Skipped existing: %d\n", len(result.Skipped))
	return nil
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
