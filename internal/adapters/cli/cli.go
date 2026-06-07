package cli

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/guferreira1/spec-harbor/internal/adapters/filesystem"
	"github.com/guferreira1/spec-harbor/internal/adapters/templates"
	"github.com/guferreira1/spec-harbor/internal/core/domain"
	"github.com/guferreira1/spec-harbor/internal/core/usecase"
	"github.com/guferreira1/spec-harbor/internal/platform/version"
)

type CommandContext struct {
	Args   []string
	Output io.Writer
}

type CommandHandler func(ctx CommandContext) error

type ExitError struct {
	Code int
}

func (err ExitError) Error() string {
	return fmt.Sprintf("exit status %d", err.Code)
}

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
		"scan":     scanCommand,
		"generate": generateCommand,
		"prompt":   promptCommand,
		"validate": validateCommand,
		"review":   reviewCommand,
		"archive":  archiveCommand,
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

func scanCommand(ctx CommandContext) error {
	if err := parseScanArguments(ctx.Args); err != nil {
		return err
	}

	root, err := os.Getwd()
	if err != nil {
		return err
	}

	fileSystem := filesystem.NewLocalFileSystem()
	scanProject := usecase.NewScanProject(fileSystem)

	result, err := scanProject.Execute(usecase.ScanProjectInput{ProjectRoot: root})
	if err != nil {
		return err
	}

	printScanReport(ctx.Output, result)
	return nil
}

func parseScanArguments(args []string) error {
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			return fmt.Errorf("unsupported flag: %s", arg)
		}
		return fmt.Errorf("unexpected argument: %s", arg)
	}
	return nil
}

func printScanReport(output io.Writer, result domain.ScanResult) {
	fmt.Fprintln(output, "SpecHarbor project scan completed.")
	fmt.Fprintf(output, "Project root: %s\n", result.ProjectRoot)

	printScanDetectionSection(output, "Detected ecosystems:", result.Ecosystems)
	printScanDetectionSection(output, "Package managers:", result.PackageManagers)
	printScanLineSection(output, "Test command hints:", result.TestCommandHints)
	printScanDetectionSection(output, "CI:", result.CIProviders)
	printScanDetectionSection(output, "Containers/deployment:", result.ContainerDeployments)
	printScanDetectionSection(output, "SpecHarbor/OpenSpec:", result.SpecHarborSignals)
	printScanLineSection(output, "Notes:", result.Notes)
}

func printScanDetectionSection(output io.Writer, heading string, detections []domain.ScanDetection) {
	fmt.Fprintln(output)
	fmt.Fprintln(output, heading)

	if len(detections) == 0 {
		fmt.Fprintln(output, "- none detected")
		return
	}

	for _, detection := range detections {
		if detection.Name != "" {
			fmt.Fprintf(output, "- %s: %s\n", detection.Name, detection.Signal)
			continue
		}
		fmt.Fprintf(output, "- %s\n", detection.Signal)
	}
}

func printScanLineSection(output io.Writer, heading string, lines []string) {
	fmt.Fprintln(output)
	fmt.Fprintln(output, heading)

	if len(lines) == 0 {
		fmt.Fprintln(output, "- none detected")
		return
	}

	for _, line := range lines {
		fmt.Fprintf(output, "- %s\n", line)
	}
}

func generateCommand(ctx CommandContext) error {
	arguments, err := parseGenerateArguments(ctx.Args)
	if err != nil {
		return err
	}

	root, err := os.Getwd()
	if err != nil {
		return err
	}

	fileSystem := filesystem.NewLocalFileSystem()
	content := templates.NewDefaultBlankChangeContent()
	generateChange := usecase.NewGenerateChange(fileSystem, content)

	result, err := generateChange.Execute(usecase.GenerateChangeInput{
		ProjectRoot: root,
		ChangeID:    arguments.changeID,
		Mode:        domain.BlankMode,
	})
	if err != nil {
		return err
	}

	printGenerationReport(ctx.Output, result)
	return nil
}

type generateArguments struct {
	changeID string
}

func parseGenerateArguments(args []string) (generateArguments, error) {
	var positionals []string
	blankCount := 0

	for _, arg := range args {
		if arg == "--blank" {
			blankCount++
			if blankCount > 1 {
				return generateArguments{}, fmt.Errorf("blank generation flag specified more than once")
			}
			continue
		}

		if strings.HasPrefix(arg, "-") {
			return generateArguments{}, fmt.Errorf("unsupported flag: %s", arg)
		}

		positionals = append(positionals, arg)
	}

	if len(positionals) == 0 {
		return generateArguments{}, fmt.Errorf("change id is required")
	}
	if len(positionals) > 1 {
		return generateArguments{}, fmt.Errorf("unexpected argument: %s", positionals[1])
	}
	if blankCount == 0 {
		return generateArguments{}, fmt.Errorf("blank generation flag is required")
	}

	return generateArguments{changeID: positionals[0]}, nil
}

func printGenerationReport(output io.Writer, result domain.GenerationResult) {
	createdFiles := result.CreatedFiles()
	skippedExistingFiles := result.SkippedExistingFiles()
	directoryStatus := "existing"
	if result.ChangeDirectoryCreated {
		directoryStatus = "created"
	}

	fmt.Fprintln(output, "SpecHarbor blank change generated.")
	fmt.Fprintf(output, "Change: %s\n", result.ChangeID)
	fmt.Fprintf(output, "Path: %s\n", result.ChangePath)
	fmt.Fprintf(output, "Directory: %s\n", directoryStatus)
	fmt.Fprintf(output, "Created files: %d\n", len(createdFiles))
	fmt.Fprintf(output, "Skipped existing files: %d\n", len(skippedExistingFiles))

	if len(createdFiles) > 0 {
		fmt.Fprintln(output)
		fmt.Fprintln(output, "Created:")
		for _, file := range createdFiles {
			fmt.Fprintf(output, "- %s\n", file)
		}
	}

	if len(skippedExistingFiles) > 0 {
		fmt.Fprintln(output)
		fmt.Fprintln(output, "Skipped existing:")
		for _, file := range skippedExistingFiles {
			fmt.Fprintf(output, "- %s\n", file)
		}
	}
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

type validateArguments struct {
	changeID string
}

func validateCommand(ctx CommandContext) error {
	arguments, err := parseValidateArguments(ctx.Args)
	if err != nil {
		return err
	}

	root, err := os.Getwd()
	if err != nil {
		return err
	}

	fileSystem := filesystem.NewLocalFileSystem()
	validateChange := usecase.NewValidateChange(fileSystem)

	result, err := validateChange.Execute(usecase.ValidateChangeInput{
		ProjectRoot: root,
		ChangeID:    arguments.changeID,
	})
	if err != nil {
		return err
	}

	printValidationReport(ctx.Output, result)
	if result.Status == domain.ValidationStatusInvalid {
		return ExitError{Code: 1}
	}

	return nil
}

func parseValidateArguments(args []string) (validateArguments, error) {
	var positionals []string

	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			return validateArguments{}, fmt.Errorf("unsupported flag: %s", arg)
		}

		positionals = append(positionals, arg)
	}

	if len(positionals) == 0 {
		return validateArguments{}, fmt.Errorf("change id is required")
	}
	if len(positionals) > 1 {
		return validateArguments{}, fmt.Errorf("unexpected argument: %s", positionals[1])
	}

	return validateArguments{changeID: positionals[0]}, nil
}

func printValidationReport(output io.Writer, result domain.ValidationResult) {
	if result.Status == domain.ValidationStatusValid {
		fmt.Fprintln(output, "SpecHarbor change is valid.")
		fmt.Fprintf(output, "Change: %s\n", result.ChangeID)
		fmt.Fprintf(output, "Checked path: %s\n", result.CheckedPath)
		fmt.Fprintf(output, "Required files: %d\n", len(result.RequiredFiles))
		fmt.Fprintf(output, "Findings: %d\n", len(result.Findings))
		return
	}

	fmt.Fprintln(output, "SpecHarbor change is invalid.")
	fmt.Fprintf(output, "Change: %s\n", result.ChangeID)
	fmt.Fprintf(output, "Checked path: %s\n", result.CheckedPath)
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Findings:")
	for _, finding := range result.Findings {
		fmt.Fprintf(output, "- [%s] %s: %s\n", finding.Severity, finding.Code, finding.Message)
	}
}

type reviewArguments struct {
	changeID string
}

func reviewCommand(ctx CommandContext) error {
	arguments, err := parseReviewArguments(ctx.Args)
	if err != nil {
		return err
	}

	root, err := os.Getwd()
	if err != nil {
		return err
	}

	fileSystem := filesystem.NewLocalFileSystem()
	reviewChange := usecase.NewReviewChange(fileSystem)

	result, err := reviewChange.Execute(usecase.ReviewChangeInput{
		ProjectRoot: root,
		ChangeID:    arguments.changeID,
	})
	if err != nil {
		return err
	}

	printReviewReport(ctx.Output, result)
	if result.Status != domain.ReviewStatusApproved {
		return ExitError{Code: 1}
	}

	return nil
}

func parseReviewArguments(args []string) (reviewArguments, error) {
	var positionals []string

	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			return reviewArguments{}, fmt.Errorf("unsupported flag: %s", arg)
		}

		positionals = append(positionals, arg)
	}

	if len(positionals) == 0 {
		return reviewArguments{}, fmt.Errorf("change id is required")
	}
	if len(positionals) > 1 {
		return reviewArguments{}, fmt.Errorf("unexpected argument: %s", positionals[1])
	}

	return reviewArguments{changeID: positionals[0]}, nil
}

func printReviewReport(output io.Writer, result domain.ReviewResult) {
	fmt.Fprintln(output, "SpecHarbor review completed.")
	fmt.Fprintf(output, "Change: %s\n", result.ChangeID)
	fmt.Fprintf(output, "Checked path: %s\n", result.CheckedPath)
	fmt.Fprintf(output, "Status: %s\n", result.Status)
	fmt.Fprintf(
		output,
		"Tasks: %d total, %d completed, %d incomplete\n",
		result.Tasks.Total,
		result.Tasks.Completed,
		result.Tasks.Incomplete,
	)

	if len(result.Findings) == 0 {
		fmt.Fprintln(output, "Findings: 0")
		return
	}

	fmt.Fprintln(output)
	fmt.Fprintln(output, "Findings:")
	for _, finding := range result.Findings {
		fmt.Fprintf(output, "- [%s] %s: %s\n", finding.Severity, finding.Code, finding.Message)
	}
}

type archiveArguments struct {
	changeID string
}

func archiveCommand(ctx CommandContext) error {
	arguments, err := parseArchiveArguments(ctx.Args)
	if err != nil {
		return err
	}

	root, err := os.Getwd()
	if err != nil {
		return err
	}

	fileSystem := filesystem.NewLocalFileSystem()
	archiveChange := usecase.NewArchiveChange(fileSystem)

	result, err := archiveChange.Execute(usecase.ArchiveChangeInput{
		ProjectRoot: root,
		ChangeID:    arguments.changeID,
		ArchiveDate: currentLocalArchiveDate(),
	})
	if err != nil {
		return err
	}

	printArchiveReport(ctx.Output, result)
	return nil
}

func parseArchiveArguments(args []string) (archiveArguments, error) {
	var positionals []string

	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			return archiveArguments{}, fmt.Errorf("unsupported flag: %s", arg)
		}

		positionals = append(positionals, arg)
	}

	if len(positionals) == 0 {
		return archiveArguments{}, fmt.Errorf("change id is required")
	}
	if len(positionals) > 1 {
		return archiveArguments{}, fmt.Errorf("unexpected argument: %s", positionals[1])
	}

	return archiveArguments{changeID: positionals[0]}, nil
}

func currentLocalArchiveDate() string {
	return formatLocalArchiveDate(time.Now())
}

func formatLocalArchiveDate(now time.Time) string {
	return now.Local().Format("2006-01-02")
}

func printArchiveReport(output io.Writer, result domain.ArchiveResult) {
	moved := "no"
	if result.Moved() {
		moved = "yes"
	}

	fmt.Fprintln(output, "SpecHarbor change archived.")
	fmt.Fprintf(output, "Change: %s\n", result.ChangeID)
	fmt.Fprintf(output, "Source: %s\n", result.SourcePath)
	fmt.Fprintf(output, "Archive: %s\n", result.ArchivePath)
	fmt.Fprintf(output, "Archive date: %s\n", result.ArchiveDate)
	fmt.Fprintf(output, "Moved: %s\n", moved)
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Moved directory:")
	fmt.Fprintf(output, "- %s -> %s\n", result.MovedDirectory.SourcePath, result.MovedDirectory.ArchivePath)
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
