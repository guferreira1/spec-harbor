package cli

import (
	"fmt"
	"strings"

	"github.com/guferreira1/spec-harbor/internal/version"
)

func Execute(args []string) error {
	if len(args) == 0 {
		printHelp()
		return nil
	}

	switch args[0] {
	case "version":
		fmt.Println(version.Version)
		return nil
	case "init":
		fmt.Println("specharbor init: not implemented yet")
		return nil
	case "scan":
		fmt.Println("specharbor scan: not implemented yet")
		return nil
	case "generate":
		fmt.Printf("specharbor generate: %s\n", strings.Join(args[1:], " "))
		return nil
	case "prompt":
		fmt.Println("specharbor prompt: not implemented yet")
		return nil
	case "validate":
		fmt.Println("specharbor validate: not implemented yet")
		return nil
	case "review":
		fmt.Println("specharbor review: not implemented yet")
		return nil
	case "archive":
		fmt.Println("specharbor archive: not implemented yet")
		return nil
	case "config":
		fmt.Println("specharbor config: not implemented yet")
		return nil
	case "help", "-h", "--help":
		printHelp()
		return nil
	default:
		return fmt.Errorf("unknown command: %s", args[0])
	}
}

func printHelp() {
	fmt.Println(`SpecHarbor

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
