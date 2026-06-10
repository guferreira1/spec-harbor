package cli

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/guferreira1/spec-harbor/internal/adapters/agentrunner"
	configadapter "github.com/guferreira1/spec-harbor/internal/adapters/config"
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
		"config":   configCommand,

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

	if arguments.mode == domain.AgentAssistedMode {
		promptTemplate := templates.NewAgentAssistedAuthoringPromptTemplate()
		localRunner := agentrunner.NewLocalCommandRunner()
		agentAssistedAuthoring := usecase.NewAgentAssistedAuthoringWithRunner(promptTemplate, localRunner)

		result, err := agentAssistedAuthoring.Execute(usecase.AgentAssistedAuthoringInput{
			ProjectRoot:   root,
			ChangeID:      arguments.changeID,
			AgentName:     arguments.agentName,
			AuthoringType: arguments.guidedType,
			Title:         arguments.title,
			Summary:       arguments.summary,
			Execute:       arguments.execute,
		})
		if err != nil {
			return err
		}

		printAgentAssistedAuthoringReport(ctx.Output, result)
		if runnerResult, ok := result.RunnerResult(); ok && runnerResult.Status == domain.AgentRunStatusNonZeroExit {
			return ExitError{Code: runnerResult.ExitCode}
		}
		return nil
	}

	fileSystem := filesystem.NewLocalFileSystem()
	blankContent := templates.NewDefaultBlankChangeContent()
	templateContent := templates.NewBuiltInChangeTemplates()
	guidedContent := templates.NewGuidedChangeTemplates()
	generateChange := usecase.NewGenerateChangeWithContent(fileSystem, blankContent, templateContent, guidedContent)

	result, err := generateChange.Execute(usecase.GenerateChangeInput{
		ProjectRoot:  root,
		ChangeID:     arguments.changeID,
		Mode:         arguments.mode,
		TemplateName: arguments.templateName,
		GuidedType:   arguments.guidedType,
		Title:        arguments.title,
		Summary:      arguments.summary,
	})
	if err != nil {
		return err
	}

	printGenerationReport(ctx.Output, result)
	return nil
}

type generateArguments struct {
	changeID     string
	mode         domain.GenerationMode
	templateName string
	guidedType   string
	agentName    string
	title        string
	summary      string
	execute      bool
}

func parseGenerateArguments(args []string) (generateArguments, error) {
	var positionals []string
	blankProvided := false
	templateProvided := false
	guidedProvided := false
	agentAssistedProvided := false
	guidedTypeProvided := false
	agentProvided := false
	titleProvided := false
	summaryProvided := false
	executeProvided := false
	var templateName string
	var guidedType string
	var agentName string
	var title string
	var summary string

	for index := 0; index < len(args); index++ {
		arg := args[index]
		if arg == "--blank" {
			if blankProvided {
				return generateArguments{}, fmt.Errorf("blank generation flag specified more than once")
			}
			blankProvided = true
			continue
		}

		if arg == "--guided" {
			if guidedProvided {
				return generateArguments{}, fmt.Errorf("guided generation flag specified more than once")
			}
			guidedProvided = true
			continue
		}

		if arg == "--agent-assisted" {
			if agentAssistedProvided {
				return generateArguments{}, fmt.Errorf("agent-assisted generation flag specified more than once")
			}
			agentAssistedProvided = true
			continue
		}

		if arg == "--template" {
			if templateProvided {
				return generateArguments{}, fmt.Errorf("template generation flag specified more than once")
			}
			if index+1 >= len(args) {
				return generateArguments{}, fmt.Errorf("template name is required")
			}
			if strings.HasPrefix(args[index+1], "-") {
				return generateArguments{}, fmt.Errorf("template name is required")
			}

			templateName = args[index+1]
			templateProvided = true
			index++
			continue
		}

		if arg == "--agent" {
			if agentProvided {
				return generateArguments{}, fmt.Errorf("agent flag specified more than once")
			}
			if index+1 >= len(args) {
				return generateArguments{}, fmt.Errorf("agent name is required")
			}
			if strings.HasPrefix(args[index+1], "-") {
				return generateArguments{}, fmt.Errorf("agent name is required")
			}

			agentName = args[index+1]
			agentProvided = true
			index++
			continue
		}

		if arg == "--type" {
			if guidedTypeProvided {
				return generateArguments{}, duplicateTypeFlagError(agentAssistedProvided)
			}
			if index+1 >= len(args) {
				return generateArguments{}, requiredTypeError(agentAssistedProvided)
			}
			if strings.HasPrefix(args[index+1], "-") {
				return generateArguments{}, requiredTypeError(agentAssistedProvided)
			}

			guidedType = args[index+1]
			guidedTypeProvided = true
			index++
			continue
		}

		if arg == "--title" {
			if titleProvided {
				return generateArguments{}, duplicateTitleFlagError(agentAssistedProvided)
			}
			if index+1 >= len(args) {
				return generateArguments{}, requiredTitleError(agentAssistedProvided)
			}
			if strings.HasPrefix(args[index+1], "-") {
				return generateArguments{}, requiredTitleError(agentAssistedProvided)
			}

			title = args[index+1]
			titleProvided = true
			index++
			continue
		}

		if arg == "--summary" {
			if summaryProvided {
				return generateArguments{}, duplicateSummaryFlagError(agentAssistedProvided)
			}
			if index+1 >= len(args) {
				return generateArguments{}, requiredSummaryError(agentAssistedProvided)
			}
			if strings.HasPrefix(args[index+1], "-") {
				return generateArguments{}, requiredSummaryError(agentAssistedProvided)
			}

			summary = args[index+1]
			summaryProvided = true
			index++
			continue
		}

		if arg == "--execute" {
			if executeProvided {
				return generateArguments{}, fmt.Errorf("execute flag specified more than once")
			}
			executeProvided = true
			continue
		}

		if strings.HasPrefix(arg, "-") {
			return generateArguments{}, fmt.Errorf("unsupported flag: %s", arg)
		}

		positionals = append(positionals, arg)
	}

	if blankProvided && templateProvided {
		return generateArguments{}, fmt.Errorf("blank and template generation flags cannot be used together")
	}
	if guidedProvided && blankProvided {
		return generateArguments{}, fmt.Errorf("guided and blank generation flags cannot be used together")
	}
	if guidedProvided && templateProvided {
		return generateArguments{}, fmt.Errorf("guided and template generation flags cannot be used together")
	}
	if agentAssistedProvided && blankProvided {
		return generateArguments{}, fmt.Errorf("agent-assisted and blank generation flags cannot be used together")
	}
	if agentAssistedProvided && templateProvided {
		return generateArguments{}, fmt.Errorf("agent-assisted and template generation flags cannot be used together")
	}
	if agentAssistedProvided && guidedProvided {
		return generateArguments{}, fmt.Errorf("agent-assisted and guided generation flags cannot be used together")
	}
	if !agentAssistedProvided && executeProvided {
		return generateArguments{}, fmt.Errorf("unsupported flag: --execute")
	}
	if !agentAssistedProvided && agentProvided {
		return generateArguments{}, fmt.Errorf("agent-assisted input flags require --agent-assisted")
	}
	if !guidedProvided && !agentAssistedProvided && (guidedTypeProvided || titleProvided || summaryProvided) {
		return generateArguments{}, fmt.Errorf("guided input flags require --guided")
	}
	if len(positionals) == 0 {
		return generateArguments{}, fmt.Errorf("change id is required")
	}
	if len(positionals) > 1 {
		return generateArguments{}, fmt.Errorf("unexpected argument: %s", positionals[1])
	}
	if !blankProvided && !templateProvided && !guidedProvided && !agentAssistedProvided {
		return generateArguments{}, fmt.Errorf("generation mode flag is required")
	}
	if agentAssistedProvided {
		if !agentProvided || strings.TrimSpace(agentName) == "" {
			return generateArguments{}, fmt.Errorf("agent name is required")
		}
		if _, err := domain.ParseAgentName(agentName); err != nil {
			return generateArguments{}, err
		}
		if !guidedTypeProvided || strings.TrimSpace(guidedType) == "" {
			return generateArguments{}, fmt.Errorf("agent-assisted authoring type is required")
		}
		if _, err := domain.ParseAgentAssistedAuthoringType(guidedType); err != nil {
			return generateArguments{}, err
		}
		if !titleProvided || strings.TrimSpace(title) == "" {
			return generateArguments{}, fmt.Errorf("agent-assisted title is required")
		}
		if !summaryProvided || strings.TrimSpace(summary) == "" {
			return generateArguments{}, fmt.Errorf("agent-assisted summary is required")
		}
		return generateArguments{
			changeID:   positionals[0],
			mode:       domain.AgentAssistedMode,
			guidedType: guidedType,
			agentName:  agentName,
			title:      title,
			summary:    summary,
			execute:    executeProvided,
		}, nil
	}
	if templateProvided {
		if strings.TrimSpace(templateName) == "" {
			return generateArguments{}, fmt.Errorf("template name is required")
		}
		return generateArguments{
			changeID:     positionals[0],
			mode:         domain.TemplateMode,
			templateName: templateName,
		}, nil
	}
	if guidedProvided {
		if !guidedTypeProvided || strings.TrimSpace(guidedType) == "" {
			return generateArguments{}, fmt.Errorf("guided type is required")
		}
		if _, err := domain.ParseGuidedType(guidedType); err != nil {
			return generateArguments{}, err
		}
		if !titleProvided || strings.TrimSpace(title) == "" {
			return generateArguments{}, fmt.Errorf("guided title is required")
		}
		if !summaryProvided || strings.TrimSpace(summary) == "" {
			return generateArguments{}, fmt.Errorf("guided summary is required")
		}
		return generateArguments{
			changeID:   positionals[0],
			mode:       domain.GuidedMode,
			guidedType: guidedType,
			title:      title,
			summary:    summary,
		}, nil
	}

	return generateArguments{
		changeID: positionals[0],
		mode:     domain.BlankMode,
	}, nil
}

func duplicateTypeFlagError(agentAssistedProvided bool) error {
	if agentAssistedProvided {
		return fmt.Errorf("agent-assisted authoring type flag specified more than once")
	}
	return fmt.Errorf("guided type flag specified more than once")
}

func requiredTypeError(agentAssistedProvided bool) error {
	if agentAssistedProvided {
		return fmt.Errorf("agent-assisted authoring type is required")
	}
	return fmt.Errorf("guided type is required")
}

func duplicateTitleFlagError(agentAssistedProvided bool) error {
	if agentAssistedProvided {
		return fmt.Errorf("agent-assisted title flag specified more than once")
	}
	return fmt.Errorf("guided title flag specified more than once")
}

func requiredTitleError(agentAssistedProvided bool) error {
	if agentAssistedProvided {
		return fmt.Errorf("agent-assisted title is required")
	}
	return fmt.Errorf("guided title is required")
}

func duplicateSummaryFlagError(agentAssistedProvided bool) error {
	if agentAssistedProvided {
		return fmt.Errorf("agent-assisted summary flag specified more than once")
	}
	return fmt.Errorf("guided summary flag specified more than once")
}

func requiredSummaryError(agentAssistedProvided bool) error {
	if agentAssistedProvided {
		return fmt.Errorf("agent-assisted summary is required")
	}
	return fmt.Errorf("guided summary is required")
}

func printGenerationReport(output io.Writer, result domain.GenerationResult) {
	createdFiles := result.CreatedFiles()
	skippedExistingFiles := result.SkippedExistingFiles()
	directoryStatus := "existing"
	if result.ChangeDirectoryCreated {
		directoryStatus = "created"
	}

	if result.Mode == domain.TemplateMode {
		fmt.Fprintln(output, "SpecHarbor template change generated.")
	} else if result.Mode == domain.GuidedMode {
		fmt.Fprintln(output, "SpecHarbor guided change generated.")
	} else {
		fmt.Fprintln(output, "SpecHarbor blank change generated.")
	}
	fmt.Fprintf(output, "Change: %s\n", result.ChangeID)
	if result.Mode == domain.TemplateMode {
		fmt.Fprintf(output, "Template: %s\n", result.TemplateName)
	}
	if result.Mode == domain.GuidedMode {
		fmt.Fprintf(output, "Guided type: %s\n", result.GuidedType)
		fmt.Fprintf(output, "Title: %s\n", result.GuidedTitle)
	}
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

func printAgentAssistedAuthoringReport(output io.Writer, result domain.AgentAssistedAuthoringResult) {
	if result.Execute {
		printAgentAssistedAuthoringExecutionReport(output, result)
		return
	}

	fmt.Fprintln(output, "SpecHarbor agent-assisted spec authoring dry run.")
	fmt.Fprintf(output, "Change: %s\n", result.ChangeID)
	fmt.Fprintf(output, "Agent: %s\n", result.AgentName)
	fmt.Fprintf(output, "Authoring type: %s\n", result.AuthoringType)
	fmt.Fprintf(output, "Title: %s\n", result.Title)
	fmt.Fprintf(output, "Summary: %s\n", result.Summary)
	fmt.Fprintf(output, "Path: %s\n", result.ChangePath)
	fmt.Fprintf(output, "Dry run: %s\n", yesNo(result.DryRun))

	fmt.Fprintln(output)
	fmt.Fprintln(output, "Required files:")
	for _, requiredFile := range result.RequiredFiles() {
		fmt.Fprintf(output, "- %s\n", requiredFile)
	}

	fmt.Fprintln(output)
	fmt.Fprintln(output, "Plan:")
	for _, planItem := range result.Plan() {
		fmt.Fprintf(output, "- %s\n", planItem)
	}

	fmt.Fprintln(output)
	fmt.Fprintln(output, "Status:")
	fmt.Fprintf(output, "- No files written: %s\n", yesNo(result.NoFilesWritten))
	fmt.Fprintf(output, "- No prompt file written: %s\n", yesNo(result.NoPromptFileWritten))
	fmt.Fprintf(output, "- No agent executed: %s\n", yesNo(result.NoAgentExecuted))
	fmt.Fprintf(output, "- No external command executed: %s\n", yesNo(result.NoExternalCommandExecuted))
	fmt.Fprintf(output, "- No agent output parsed or applied: %s\n", yesNo(result.NoAgentOutputParsedOrApplied))

	fmt.Fprintln(output)
	fmt.Fprintln(output, "Generated prompt:")
	fmt.Fprintln(output)
	fmt.Fprint(output, result.Prompt)
	if !strings.HasSuffix(result.Prompt, "\n") {
		fmt.Fprintln(output)
	}
}

func printAgentAssistedAuthoringExecutionReport(output io.Writer, result domain.AgentAssistedAuthoringResult) {
	runnerResult, ok := result.RunnerResult()
	if !ok {
		return
	}

	fmt.Fprintln(output, "SpecHarbor agent-assisted spec authoring execute run.")
	fmt.Fprintln(output, "Mode: execute")
	fmt.Fprintf(output, "Change: %s\n", result.ChangeID)
	fmt.Fprintf(output, "Agent target id: %s\n", result.AgentName)
	fmt.Fprintf(output, "Agent display name: %s\n", result.AgentDisplayName)
	fmt.Fprintf(output, "Resolved command: %s\n", result.ResolvedCommandName)
	printResolvedArgs(output, result.ResolvedCommandFixedArgs())
	fmt.Fprintf(output, "Authoring type: %s\n", result.AuthoringType)
	fmt.Fprintf(output, "Title: %s\n", result.Title)
	fmt.Fprintf(output, "Summary: %s\n", result.Summary)
	fmt.Fprintf(output, "Path: %s\n", result.ChangePath)
	fmt.Fprintf(output, "Working directory: project root (%s)\n", result.WorkingDirectory)
	fmt.Fprintf(output, "Prompt sent through stdin: %s\n", yesNo(result.PromptSentToRunner))
	fmt.Fprintf(output, "Execution status: %s\n", runnerResult.Status)
	fmt.Fprintf(output, "Exit code: %d\n", runnerResult.ExitCode)

	fmt.Fprintln(output)
	fmt.Fprintln(output, "Required files:")
	for _, requiredFile := range result.RequiredFiles() {
		fmt.Fprintf(output, "- %s\n", requiredFile)
	}

	fmt.Fprintln(output)
	fmt.Fprintln(output, "Plan:")
	for _, planItem := range result.Plan() {
		fmt.Fprintf(output, "- %s\n", planItem)
	}

	printCapturedOutput(output, "Stdout:", runnerResult.Stdout)
	printCapturedOutput(output, "Stderr:", runnerResult.Stderr)

	fmt.Fprintln(output)
	fmt.Fprintln(output, "Status:")
	fmt.Fprintf(output, "- Output parsed or applied by SpecHarbor: %s\n", yesNo(!result.NoAgentOutputParsedOrApplied))
	fmt.Fprintf(output, "- OpenSpec files written from runner output: %s\n", yesNo(!result.NoOpenSpecFilesWrittenFromOutput))
	fmt.Fprintf(output, "- Production code modified by SpecHarbor: %s\n", yesNo(!result.NoProductionCodeModifiedBySpecHarbor))
	fmt.Fprintf(output, "- Auto-commit, auto-push, or auto-merge by SpecHarbor: %s\n", yesNo(!result.NoAutoCommitPushMerge))
}

func printResolvedArgs(output io.Writer, args []string) {
	if len(args) == 0 {
		fmt.Fprintln(output, "Resolved args: (none)")
		return
	}
	fmt.Fprintf(output, "Resolved args: %s\n", strings.Join(args, " "))
}

func printCapturedOutput(output io.Writer, heading string, contents string) {
	fmt.Fprintln(output)
	fmt.Fprintln(output, heading)
	if contents == "" {
		fmt.Fprintln(output, "(empty)")
		return
	}
	fmt.Fprint(output, contents)
	if !strings.HasSuffix(contents, "\n") {
		fmt.Fprintln(output)
	}
}

func yesNo(value bool) string {
	if value {
		return "yes"
	}
	return "no"
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

func configCommand(ctx CommandContext) error {
	if err := parseConfigArguments(ctx.Args); err != nil {
		return err
	}

	root, err := os.Getwd()
	if err != nil {
		return err
	}

	fileSystem := filesystem.NewLocalFileSystem()
	parser := configadapter.NewYAMLParser()
	showConfig := usecase.NewShowConfig(fileSystem, parser)

	result, err := showConfig.Execute(usecase.ShowConfigInput{ProjectRoot: root})
	if err != nil {
		return err
	}

	printConfigReport(ctx.Output, result)
	return nil
}

func parseConfigArguments(args []string) error {
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			return fmt.Errorf("unsupported flag: %s", arg)
		}
	}

	if len(args) == 0 {
		return nil
	}
	if args[0] != "show" {
		return fmt.Errorf("unsupported config subcommand: %s", args[0])
	}
	if len(args) > 1 {
		return fmt.Errorf("unexpected argument: %s", args[1])
	}

	return nil
}

func printConfigReport(output io.Writer, result domain.ConfigResult) {
	fmt.Fprintln(output, "SpecHarbor configuration loaded.")
	fmt.Fprintf(output, "Path: %s\n", result.Path)
	fmt.Fprintf(output, "Version: %d\n", result.Config.Version)

	fmt.Fprintln(output)
	fmt.Fprintln(output, "Defaults:")
	fmt.Fprintf(output, "- agent role: %s\n", result.Config.Defaults.AgentRole)
	fmt.Fprintf(output, "- generation mode: %s\n", result.Config.Defaults.GenerationMode)

	fmt.Fprintln(output)
	fmt.Fprintln(output, "Validation:")
	fmt.Fprintf(output, "- require all change files: %t\n", result.Config.Validation.RequireAllChangeFiles)

	fmt.Fprintln(output)
	fmt.Fprintln(output, "Review:")
	fmt.Fprintf(output, "- require completed tasks: %t\n", result.Config.Review.RequireCompletedTasks)

	fmt.Fprintln(output)
	fmt.Fprintln(output, "Archive:")
	fmt.Fprintf(output, "- date layout: %s\n", result.Config.Archive.DateLayout)

	fmt.Fprintln(output)
	fmt.Fprintln(output, "Scan:")
	fmt.Fprintf(output, "- include common project files: %t\n", result.Config.Scan.IncludeCommonProjectFiles)

	fmt.Fprintln(output)
	fmt.Fprintln(output, "Output:")
	fmt.Fprintf(output, "- format: %s\n", result.Config.Output.Format)
}

func helpCommand(ctx CommandContext) error {
	printHelp(ctx.Output)
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
