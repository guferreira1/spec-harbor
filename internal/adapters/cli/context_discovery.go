package cli

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/guferreira1/spec-harbor/internal/adapters/filesystem"
	"github.com/guferreira1/spec-harbor/internal/core/domain"
	"github.com/guferreira1/spec-harbor/internal/core/usecase"
)

func contextCommand(ctx CommandContext) error {
	if err := parseContextArguments(ctx.Args); err != nil {
		return err
	}

	root, err := os.Getwd()
	if err != nil {
		return err
	}

	fileSystem := filesystem.NewContextDiscoveryFileSystem()
	discoverContext := usecase.NewDiscoverProjectContext(fileSystem)
	result, err := discoverContext.Execute(usecase.DiscoverProjectContextInput{ProjectRoot: root})
	if err != nil {
		return err
	}

	printContextDiscoveryReport(ctx.Output, result)
	return nil
}

func parseContextArguments(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("context subcommand is required: discover")
	}
	if strings.HasPrefix(args[0], "-") {
		return fmt.Errorf("unsupported flag: %s", args[0])
	}
	if args[0] != "discover" {
		return fmt.Errorf("unsupported context subcommand: %s", args[0])
	}

	for _, arg := range args[1:] {
		if strings.HasPrefix(arg, "-") {
			return fmt.Errorf("unsupported flag: %s", arg)
		}
		return fmt.Errorf("unexpected argument: %s", arg)
	}
	return nil
}

func printContextDiscoveryReport(output io.Writer, result domain.ContextDiscoveryResult) {
	fmt.Fprintln(output, "Detected project context:")
	printContextSignalSection(
		output,
		"User-confirmed context:",
		result.SignalsByClassification(domain.ContextSignalClassificationUserConfirmedContext),
	)
	printContextSignalSection(
		output,
		"Detected facts:",
		result.SignalsByClassification(domain.ContextSignalClassificationDetectedFact),
	)
	printContextSignalSection(
		output,
		"Suggested assumptions:",
		result.SignalsByClassification(domain.ContextSignalClassificationSuggestedAssumption),
	)
	printContextNotesSection(output, result.Notes())
}

func printContextSignalSection(output io.Writer, heading string, signals []domain.ContextSignal) {
	fmt.Fprintln(output)
	fmt.Fprintln(output, heading)
	if len(signals) == 0 {
		fmt.Fprintln(output, "- none detected")
		return
	}
	for _, signal := range signals {
		fmt.Fprintf(output, "- %s: %s\n", signal.Kind.Label(), signal.Value)
		fmt.Fprintf(output, "  Source: %s\n", formatContextSource(signal.Source))
		fmt.Fprintf(output, "  Classification: %s\n", signal.Classification)
		fmt.Fprintf(output, "  Confidence: %s\n", signal.Confidence)
	}
}

func printContextNotesSection(output io.Writer, notes []domain.ContextDiscoveryNote) {
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Notes:")
	if len(notes) == 0 {
		fmt.Fprintln(output, "- none detected")
		return
	}
	for _, note := range notes {
		fmt.Fprintf(output, "- %s\n", note.Message)
	}
}

func formatContextSource(source domain.ContextSource) string {
	if source.Evidence == "" {
		return source.Path
	}
	return fmt.Sprintf("%s (%s)", source.Path, source.Evidence)
}
