package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/guferreira1/spec-harbor/internal/adapters/agentrunner"
	configadapter "github.com/guferreira1/spec-harbor/internal/adapters/config"
	"github.com/guferreira1/spec-harbor/internal/adapters/filesystem"
	remoteadapter "github.com/guferreira1/spec-harbor/internal/adapters/remote"
	"github.com/guferreira1/spec-harbor/internal/adapters/templates"
	"github.com/guferreira1/spec-harbor/internal/core/domain"
	"github.com/guferreira1/spec-harbor/internal/core/ports"
	"github.com/guferreira1/spec-harbor/internal/core/usecase"
	"github.com/guferreira1/spec-harbor/internal/platform/version"
)

type CommandContext struct {
	Args     []string
	Output   io.Writer
	terminal interactiveTerminal
}

type CommandHandler func(ctx CommandContext) error

var newRemoteTemplateFetcher = func() ports.RemoteTemplateFetcher {
	return remoteadapter.NewHTTPTemplateFetcher()
}

var newRemoteTemplateBundleReader = func() ports.RemoteTemplateBundleReader {
	return remoteadapter.NewZIPTemplateBundleReader()
}

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
	return executeWithTerminal(args, output, nil)
}

func executeWithTerminal(args []string, output io.Writer, terminal interactiveTerminal) error {
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
		Args:     args[1:],
		Output:   output,
		terminal: terminal,
	})
}

func commandRegistry() map[string]CommandHandler {
	return map[string]CommandHandler{
		"version":  versionCommand,
		"init":     initCommand,
		"scan":     scanCommand,
		"context":  contextCommand,
		"brief":    briefCommand,
		"generate": generateCommand,
		"prompt":   promptCommand,
		"validate": validateCommand,
		"review":   reviewCommand,
		"archive":  archiveCommand,
		"config":   configCommand,
		"workflow": workflowCommand,

		"help":   helpCommand,
		"-h":     helpCommand,
		"--help": helpCommand,
	}
}

func versionCommand(ctx CommandContext) error {
	if err := parseVersionArguments(ctx.Args); err != nil {
		return err
	}

	fmt.Fprintln(ctx.Output, version.Current().Format())
	return nil
}

func parseVersionArguments(args []string) error {
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			return fmt.Errorf("unsupported flag: %s", arg)
		}
		return fmt.Errorf("unexpected argument: %s", arg)
	}
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

	if arguments.interactive {
		parsedChangeID, err := domain.NewChangeID(arguments.changeID)
		if err != nil {
			return err
		}

		terminal := ctx.terminal
		if terminal == nil {
			terminal = newOSInteractiveTerminal(ctx.Output)
		}

		arguments, err = promptInteractiveGeneration(terminal, parsedChangeID.String())
		if err != nil {
			return err
		}
	}

	root, err := os.Getwd()
	if err != nil {
		return err
	}

	if arguments.mode == domain.AIAssistedMode {
		fileSystem := filesystem.NewLocalFileSystem()
		validateChange := usecase.NewValidateChange(fileSystem)
		generateAIAssisted := usecase.NewGenerateAIAssistedChange(fileSystem, validateChange)

		result, err := generateAIAssisted.Execute(usecase.GenerateAIAssistedChangeInput{
			ProjectRoot: root,
			ChangeID:    arguments.changeID,
			SourcePath:  arguments.fromFile,
			Overwrite:   arguments.overwrite,
		})
		if err != nil {
			var parseFailure *usecase.AIAssistedParseFailure
			if errors.As(err, &parseFailure) {
				printAIAssistedParseFailureReport(ctx.Output, parseFailure)
				return ExitError{Code: 1}
			}
			return err
		}

		printAIAssistedGenerationReport(ctx.Output, result)
		if validationResult, ok := result.ValidationResult(); ok && validationResult.Status == domain.ValidationStatusInvalid {
			return ExitError{Code: 1}
		}
		return nil
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
	if arguments.mode == domain.HybridMode {
		templateContent := templates.NewBuiltInChangeTemplates()
		configParser := configadapter.NewYAMLParser()
		validateChange := usecase.NewValidateChange(fileSystem)
		generateHybrid := usecase.NewGenerateHybridChange(
			fileSystem,
			templateContent,
			fileSystem,
			fileSystem,
			configParser,
			newRemoteTemplateFetcher(),
			newRemoteTemplateBundleReader(),
			validateChange,
		)

		result, err := generateHybrid.Execute(usecase.GenerateHybridChangeInput{
			ProjectRoot:         root,
			ChangeID:            arguments.changeID,
			TemplateName:        arguments.templateName,
			CustomTemplateName:  arguments.customTemplateName,
			ConfigTemplateAlias: arguments.configTemplateAlias,
			Title:               arguments.title,
			Summary:             arguments.summary,
			Type:                arguments.guidedType,
		})
		if err != nil {
			return err
		}

		printHybridGenerationReport(ctx.Output, result)
		if validationResult, ok := result.ValidationResult(); ok && validationResult.Status == domain.ValidationStatusInvalid {
			return ExitError{Code: 1}
		}
		return nil
	}

	blankContent := templates.NewDefaultBlankChangeContent()
	templateContent := templates.NewBuiltInChangeTemplates()
	guidedContent := templates.NewGuidedChangeTemplates()
	configParser := configadapter.NewYAMLParser()
	generateChange := usecase.NewGenerateChangeWithRemoteConfigTemplates(
		fileSystem,
		blankContent,
		templateContent,
		guidedContent,
		fileSystem,
		fileSystem,
		configParser,
		newRemoteTemplateFetcher(),
		newRemoteTemplateBundleReader(),
	)

	generateInput := usecase.GenerateChangeInput{
		ProjectRoot:  root,
		ChangeID:     arguments.changeID,
		Mode:         arguments.mode,
		TemplateName: arguments.templateName,
		GuidedType:   arguments.guidedType,
		Title:        arguments.title,
		Summary:      arguments.summary,
	}
	if arguments.customTemplate {
		generateInput.TemplateSource = domain.CustomTemplateSource
		generateInput.CustomTemplateName = arguments.customTemplateName
	}
	if arguments.configTemplate {
		generateInput.ConfigTemplate = true
		generateInput.ConfigTemplateAlias = arguments.configTemplateAlias
	}

	result, err := generateChange.Execute(generateInput)
	if err != nil {
		return err
	}

	printGenerationReport(ctx.Output, result)
	return nil
}

type generateArguments struct {
	changeID            string
	mode                domain.GenerationMode
	interactive         bool
	templateName        string
	customTemplate      bool
	customTemplateName  string
	configTemplate      bool
	configTemplateAlias string
	guidedType          string
	agentName           string
	fromFile            string
	title               string
	summary             string
	execute             bool
	overwrite           bool
}

func parseGenerateArguments(args []string) (generateArguments, error) {
	var positionals []string
	interactiveProvided := false
	hybridProvided := false
	blankProvided := false
	templateProvided := false
	customTemplateProvided := false
	configTemplateProvided := false
	guidedProvided := false
	aiAssistedProvided := false
	agentAssistedProvided := false
	guidedTypeProvided := false
	agentProvided := false
	fromFileProvided := false
	titleProvided := false
	summaryProvided := false
	executeProvided := false
	overwriteProvided := false
	var templateName string
	var customTemplateName string
	var configTemplateAlias string
	var guidedType string
	var agentName string
	var fromFile string
	var title string
	var summary string

	for index := 0; index < len(args); index++ {
		arg := args[index]
		if arg == "--interactive" {
			if interactiveProvided {
				return generateArguments{}, fmt.Errorf("interactive generation flag specified more than once")
			}
			interactiveProvided = true
			continue
		}

		if arg == "--hybrid" {
			if hybridProvided {
				return generateArguments{}, fmt.Errorf("hybrid generation flag specified more than once")
			}
			hybridProvided = true
			continue
		}

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

		if arg == "--ai-assisted" {
			if aiAssistedProvided {
				return generateArguments{}, fmt.Errorf("ai-assisted generation flag specified more than once")
			}
			aiAssistedProvided = true
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

		if arg == "--custom-template" {
			if customTemplateProvided {
				return generateArguments{}, fmt.Errorf("custom-template generation flag specified more than once")
			}
			if index+1 >= len(args) {
				return generateArguments{}, fmt.Errorf("custom template name is required")
			}
			if strings.HasPrefix(args[index+1], "-") {
				return generateArguments{}, fmt.Errorf("custom template name is required")
			}

			customTemplateName = args[index+1]
			customTemplateProvided = true
			index++
			continue
		}

		if arg == "--config-template" {
			if configTemplateProvided {
				return generateArguments{}, fmt.Errorf("config-template generation flag specified more than once")
			}
			if index+1 >= len(args) {
				return generateArguments{}, fmt.Errorf("config template alias is required")
			}
			if strings.HasPrefix(args[index+1], "-") {
				return generateArguments{}, fmt.Errorf("config template alias is required")
			}

			configTemplateAlias = args[index+1]
			configTemplateProvided = true
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

		if arg == "--from-file" {
			if fromFileProvided {
				return generateArguments{}, fmt.Errorf("from-file flag specified more than once")
			}
			if index+1 >= len(args) {
				return generateArguments{}, fmt.Errorf("source file is required")
			}
			if strings.HasPrefix(args[index+1], "-") {
				return generateArguments{}, fmt.Errorf("source file is required")
			}

			fromFile = args[index+1]
			fromFileProvided = true
			index++
			continue
		}

		if arg == "--type" {
			if guidedTypeProvided {
				return generateArguments{}, duplicateTypeFlagError(agentAssistedProvided, hybridProvided)
			}
			if index+1 >= len(args) {
				return generateArguments{}, requiredTypeError(agentAssistedProvided, hybridProvided)
			}
			if strings.HasPrefix(args[index+1], "-") {
				return generateArguments{}, requiredTypeError(agentAssistedProvided, hybridProvided)
			}

			guidedType = args[index+1]
			guidedTypeProvided = true
			index++
			continue
		}

		if arg == "--title" {
			if titleProvided {
				return generateArguments{}, duplicateTitleFlagError(agentAssistedProvided, configTemplateProvided, hybridProvided)
			}
			if index+1 >= len(args) {
				return generateArguments{}, requiredTitleError(agentAssistedProvided, configTemplateProvided, hybridProvided)
			}
			if strings.HasPrefix(args[index+1], "-") {
				return generateArguments{}, requiredTitleError(agentAssistedProvided, configTemplateProvided, hybridProvided)
			}

			title = args[index+1]
			titleProvided = true
			index++
			continue
		}

		if arg == "--summary" {
			if summaryProvided {
				return generateArguments{}, duplicateSummaryFlagError(agentAssistedProvided, configTemplateProvided, hybridProvided)
			}
			if index+1 >= len(args) {
				return generateArguments{}, requiredSummaryError(agentAssistedProvided, configTemplateProvided, hybridProvided)
			}
			if strings.HasPrefix(args[index+1], "-") {
				return generateArguments{}, requiredSummaryError(agentAssistedProvided, configTemplateProvided, hybridProvided)
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

		if arg == "--overwrite" {
			if overwriteProvided {
				return generateArguments{}, fmt.Errorf("overwrite flag specified more than once")
			}
			overwriteProvided = true
			continue
		}

		if strings.HasPrefix(arg, "-") {
			return generateArguments{}, fmt.Errorf("unsupported flag: %s", arg)
		}

		positionals = append(positionals, arg)
	}

	if interactiveProvided {
		if blankProvided {
			return generateArguments{}, fmt.Errorf("interactive and blank generation flags cannot be used together")
		}
		if templateProvided {
			return generateArguments{}, fmt.Errorf("interactive and template generation flags cannot be used together")
		}
		if customTemplateProvided {
			return generateArguments{}, fmt.Errorf("interactive and custom-template generation flags cannot be used together")
		}
		if configTemplateProvided {
			return generateArguments{}, fmt.Errorf("interactive and config-template generation flags cannot be used together")
		}
		if guidedProvided {
			return generateArguments{}, fmt.Errorf("interactive and guided generation flags cannot be used together")
		}
		if hybridProvided {
			return generateArguments{}, fmt.Errorf("interactive and hybrid generation flags cannot be used together")
		}
		if aiAssistedProvided {
			return generateArguments{}, fmt.Errorf("interactive and ai-assisted generation flags cannot be used together")
		}
		if agentAssistedProvided {
			return generateArguments{}, fmt.Errorf("interactive and agent-assisted generation flags cannot be used together")
		}
		if guidedTypeProvided {
			return generateArguments{}, fmt.Errorf("interactive and type flags cannot be used together")
		}
		if titleProvided {
			return generateArguments{}, fmt.Errorf("interactive and title flags cannot be used together")
		}
		if summaryProvided {
			return generateArguments{}, fmt.Errorf("interactive and summary flags cannot be used together")
		}
		if fromFileProvided {
			return generateArguments{}, fmt.Errorf("interactive and from-file flags cannot be used together")
		}
		if agentProvided {
			return generateArguments{}, fmt.Errorf("interactive and agent flags cannot be used together")
		}
		if executeProvided {
			return generateArguments{}, fmt.Errorf("interactive and execute flags cannot be used together")
		}
		if overwriteProvided {
			return generateArguments{}, fmt.Errorf("interactive and overwrite flags cannot be used together")
		}
		if len(positionals) == 0 {
			return generateArguments{}, fmt.Errorf("change id is required")
		}
		if len(positionals) > 1 {
			return generateArguments{}, fmt.Errorf("unexpected argument: %s", positionals[1])
		}
		return generateArguments{
			changeID:    positionals[0],
			interactive: true,
		}, nil
	}

	if hybridProvided {
		return parseHybridGenerateArguments(positionals, hybridArgumentState{
			blankProvided:          blankProvided,
			templateProvided:       templateProvided,
			customTemplateProvided: customTemplateProvided,
			configTemplateProvided: configTemplateProvided,
			guidedProvided:         guidedProvided,
			aiAssistedProvided:     aiAssistedProvided,
			agentAssistedProvided:  agentAssistedProvided,
			agentProvided:          agentProvided,
			fromFileProvided:       fromFileProvided,
			titleProvided:          titleProvided,
			summaryProvided:        summaryProvided,
			guidedTypeProvided:     guidedTypeProvided,
			executeProvided:        executeProvided,
			overwriteProvided:      overwriteProvided,
			templateName:           templateName,
			customTemplateName:     customTemplateName,
			configTemplateAlias:    configTemplateAlias,
			guidedType:             guidedType,
			title:                  title,
			summary:                summary,
		})
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
	if aiAssistedProvided && blankProvided {
		return generateArguments{}, fmt.Errorf("ai-assisted and blank generation flags cannot be used together")
	}
	if aiAssistedProvided && templateProvided {
		return generateArguments{}, fmt.Errorf("ai-assisted and template generation flags cannot be used together")
	}
	if aiAssistedProvided && customTemplateProvided {
		return generateArguments{}, fmt.Errorf("ai-assisted and custom-template generation flags cannot be used together")
	}
	if aiAssistedProvided && guidedProvided {
		return generateArguments{}, fmt.Errorf("ai-assisted and guided generation flags cannot be used together")
	}
	if aiAssistedProvided && agentAssistedProvided {
		return generateArguments{}, fmt.Errorf("ai-assisted and agent-assisted generation flags cannot be used together")
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
	if customTemplateProvided && blankProvided {
		return generateArguments{}, fmt.Errorf("custom-template and blank generation flags cannot be used together")
	}
	if customTemplateProvided && templateProvided {
		return generateArguments{}, fmt.Errorf("custom-template and template generation flags cannot be used together")
	}
	if customTemplateProvided && guidedProvided {
		return generateArguments{}, fmt.Errorf("custom-template and guided generation flags cannot be used together")
	}
	if customTemplateProvided && agentAssistedProvided {
		return generateArguments{}, fmt.Errorf("custom-template and agent-assisted generation flags cannot be used together")
	}
	if configTemplateProvided && blankProvided {
		return generateArguments{}, fmt.Errorf("config-template and blank generation flags cannot be used together")
	}
	if configTemplateProvided && templateProvided {
		return generateArguments{}, fmt.Errorf("config-template and template generation flags cannot be used together")
	}
	if configTemplateProvided && customTemplateProvided {
		return generateArguments{}, fmt.Errorf("config-template and custom-template generation flags cannot be used together")
	}
	if configTemplateProvided && guidedProvided {
		return generateArguments{}, fmt.Errorf("config-template and guided generation flags cannot be used together")
	}
	if configTemplateProvided && agentAssistedProvided {
		return generateArguments{}, fmt.Errorf("config-template and agent-assisted generation flags cannot be used together")
	}
	if configTemplateProvided && aiAssistedProvided {
		return generateArguments{}, fmt.Errorf("config-template and ai-assisted generation flags cannot be used together")
	}
	if configTemplateProvided && executeProvided {
		return generateArguments{}, fmt.Errorf("config-template and execute flags cannot be used together")
	}
	if configTemplateProvided && guidedTypeProvided {
		return generateArguments{}, fmt.Errorf("config-template and type flags cannot be used together")
	}
	if configTemplateProvided && agentProvided {
		return generateArguments{}, fmt.Errorf("config-template and agent flags cannot be used together")
	}
	if configTemplateProvided && fromFileProvided {
		return generateArguments{}, fmt.Errorf("config-template and from-file flags cannot be used together")
	}
	if configTemplateProvided && overwriteProvided {
		return generateArguments{}, fmt.Errorf("config-template and overwrite flags cannot be used together")
	}
	if aiAssistedProvided && executeProvided {
		return generateArguments{}, fmt.Errorf("ai-assisted and execute flags cannot be used together")
	}
	if !agentAssistedProvided && executeProvided {
		return generateArguments{}, fmt.Errorf("unsupported flag: --execute")
	}
	if !aiAssistedProvided && fromFileProvided {
		return generateArguments{}, fmt.Errorf("from-file requires --ai-assisted")
	}
	if !aiAssistedProvided && overwriteProvided {
		return generateArguments{}, fmt.Errorf("overwrite requires --ai-assisted")
	}
	if aiAssistedProvided && agentProvided {
		return generateArguments{}, fmt.Errorf("agent-assisted input flags cannot be used with --ai-assisted")
	}
	if aiAssistedProvided && (guidedTypeProvided || titleProvided || summaryProvided) {
		return generateArguments{}, fmt.Errorf("guided input flags cannot be used with --ai-assisted")
	}
	if !agentAssistedProvided && agentProvided {
		return generateArguments{}, fmt.Errorf("agent-assisted input flags require --agent-assisted")
	}
	if !guidedProvided && !agentAssistedProvided && !aiAssistedProvided && guidedTypeProvided {
		return generateArguments{}, fmt.Errorf("guided input flags require --guided")
	}
	if !guidedProvided && !agentAssistedProvided && !aiAssistedProvided && !customTemplateProvided && !configTemplateProvided && (titleProvided || summaryProvided) {
		return generateArguments{}, fmt.Errorf("guided input flags require --guided")
	}
	if len(positionals) == 0 {
		return generateArguments{}, fmt.Errorf("change id is required")
	}
	if len(positionals) > 1 {
		return generateArguments{}, fmt.Errorf("unexpected argument: %s", positionals[1])
	}
	if !blankProvided && !templateProvided && !customTemplateProvided && !configTemplateProvided && !guidedProvided && !aiAssistedProvided && !agentAssistedProvided {
		return generateArguments{}, fmt.Errorf("generation mode flag is required")
	}
	if aiAssistedProvided {
		if !fromFileProvided || strings.TrimSpace(fromFile) == "" {
			return generateArguments{}, fmt.Errorf("source file is required")
		}
		return generateArguments{
			changeID:  positionals[0],
			mode:      domain.AIAssistedMode,
			fromFile:  fromFile,
			overwrite: overwriteProvided,
		}, nil
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
	if customTemplateProvided {
		if strings.TrimSpace(customTemplateName) == "" {
			return generateArguments{}, fmt.Errorf("custom template name is required")
		}
		return generateArguments{
			changeID:           positionals[0],
			mode:               domain.TemplateMode,
			customTemplate:     true,
			customTemplateName: customTemplateName,
			title:              title,
			summary:            summary,
		}, nil
	}
	if configTemplateProvided {
		if strings.TrimSpace(configTemplateAlias) == "" {
			return generateArguments{}, fmt.Errorf("config template alias is required")
		}
		return generateArguments{
			changeID:            positionals[0],
			mode:                domain.TemplateMode,
			configTemplate:      true,
			configTemplateAlias: configTemplateAlias,
			title:               title,
			summary:             summary,
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

type hybridArgumentState struct {
	blankProvided          bool
	templateProvided       bool
	customTemplateProvided bool
	configTemplateProvided bool
	guidedProvided         bool
	aiAssistedProvided     bool
	agentAssistedProvided  bool
	agentProvided          bool
	fromFileProvided       bool
	titleProvided          bool
	summaryProvided        bool
	guidedTypeProvided     bool
	executeProvided        bool
	overwriteProvided      bool
	templateName           string
	customTemplateName     string
	configTemplateAlias    string
	guidedType             string
	title                  string
	summary                string
}

func parseHybridGenerateArguments(positionals []string, state hybridArgumentState) (generateArguments, error) {
	if state.blankProvided {
		return generateArguments{}, fmt.Errorf("hybrid and blank generation flags cannot be used together")
	}
	if state.guidedProvided {
		return generateArguments{}, fmt.Errorf("hybrid and guided generation flags cannot be used together")
	}
	if state.aiAssistedProvided {
		return generateArguments{}, fmt.Errorf("hybrid and ai-assisted generation flags cannot be used together")
	}
	if state.agentAssistedProvided {
		return generateArguments{}, fmt.Errorf("hybrid and agent-assisted generation flags cannot be used together")
	}
	if state.fromFileProvided {
		return generateArguments{}, fmt.Errorf("hybrid and from-file flags cannot be used together")
	}
	if state.overwriteProvided {
		return generateArguments{}, fmt.Errorf("hybrid and overwrite flags cannot be used together")
	}
	if state.agentProvided {
		return generateArguments{}, fmt.Errorf("hybrid and agent flags cannot be used together")
	}
	if state.executeProvided {
		return generateArguments{}, fmt.Errorf("hybrid and execute flags cannot be used together")
	}

	sourceCount := 0
	if state.templateProvided {
		sourceCount++
	}
	if state.customTemplateProvided {
		sourceCount++
	}
	if state.configTemplateProvided {
		sourceCount++
	}
	if sourceCount == 0 {
		return generateArguments{}, fmt.Errorf("hybrid source selector is required")
	}
	if sourceCount > 1 {
		return generateArguments{}, fmt.Errorf("hybrid requires exactly one source selector")
	}

	if len(positionals) == 0 {
		return generateArguments{}, fmt.Errorf("change id is required")
	}
	if len(positionals) > 1 {
		return generateArguments{}, fmt.Errorf("unexpected argument: %s", positionals[1])
	}
	if !state.titleProvided || strings.TrimSpace(state.title) == "" {
		return generateArguments{}, fmt.Errorf("hybrid title is required")
	}
	if !state.summaryProvided || strings.TrimSpace(state.summary) == "" {
		return generateArguments{}, fmt.Errorf("hybrid summary is required")
	}
	if state.guidedTypeProvided {
		if _, err := domain.ParseHybridType(state.guidedType); err != nil {
			return generateArguments{}, err
		}
	}

	arguments := generateArguments{
		changeID:   positionals[0],
		mode:       domain.HybridMode,
		guidedType: state.guidedType,
		title:      state.title,
		summary:    state.summary,
	}
	if state.templateProvided {
		if strings.TrimSpace(state.templateName) == "" {
			return generateArguments{}, fmt.Errorf("template name is required")
		}
		arguments.templateName = state.templateName
	}
	if state.customTemplateProvided {
		if strings.TrimSpace(state.customTemplateName) == "" {
			return generateArguments{}, fmt.Errorf("custom template name is required")
		}
		arguments.customTemplate = true
		arguments.customTemplateName = state.customTemplateName
	}
	if state.configTemplateProvided {
		if strings.TrimSpace(state.configTemplateAlias) == "" {
			return generateArguments{}, fmt.Errorf("config template alias is required")
		}
		arguments.configTemplate = true
		arguments.configTemplateAlias = state.configTemplateAlias
	}
	return arguments, nil
}

func duplicateTypeFlagError(agentAssistedProvided bool, hybridProvided bool) error {
	if agentAssistedProvided {
		return fmt.Errorf("agent-assisted authoring type flag specified more than once")
	}
	if hybridProvided {
		return fmt.Errorf("hybrid type flag specified more than once")
	}
	return fmt.Errorf("guided type flag specified more than once")
}

func requiredTypeError(agentAssistedProvided bool, hybridProvided bool) error {
	if agentAssistedProvided {
		return fmt.Errorf("agent-assisted authoring type is required")
	}
	if hybridProvided {
		return fmt.Errorf("hybrid type is required")
	}
	return fmt.Errorf("guided type is required")
}

func duplicateTitleFlagError(agentAssistedProvided bool, configTemplateProvided bool, hybridProvided bool) error {
	if agentAssistedProvided {
		return fmt.Errorf("agent-assisted title flag specified more than once")
	}
	if hybridProvided {
		return fmt.Errorf("hybrid title flag specified more than once")
	}
	if configTemplateProvided {
		return fmt.Errorf("config-template title flag specified more than once")
	}
	return fmt.Errorf("guided title flag specified more than once")
}

func requiredTitleError(agentAssistedProvided bool, configTemplateProvided bool, hybridProvided bool) error {
	if agentAssistedProvided {
		return fmt.Errorf("agent-assisted title is required")
	}
	if hybridProvided {
		return fmt.Errorf("hybrid title is required")
	}
	if configTemplateProvided {
		return fmt.Errorf("config-template title is required")
	}
	return fmt.Errorf("guided title is required")
}

func duplicateSummaryFlagError(agentAssistedProvided bool, configTemplateProvided bool, hybridProvided bool) error {
	if agentAssistedProvided {
		return fmt.Errorf("agent-assisted summary flag specified more than once")
	}
	if hybridProvided {
		return fmt.Errorf("hybrid summary flag specified more than once")
	}
	if configTemplateProvided {
		return fmt.Errorf("config-template summary flag specified more than once")
	}
	return fmt.Errorf("guided summary flag specified more than once")
}

func requiredSummaryError(agentAssistedProvided bool, configTemplateProvided bool, hybridProvided bool) error {
	if agentAssistedProvided {
		return fmt.Errorf("agent-assisted summary is required")
	}
	if hybridProvided {
		return fmt.Errorf("hybrid summary is required")
	}
	if configTemplateProvided {
		return fmt.Errorf("config-template summary is required")
	}
	return fmt.Errorf("guided summary is required")
}

func printGenerationReport(output io.Writer, result domain.GenerationResult) {
	if result.ConfigTemplateAlias != "" {
		printConfigTemplateGenerationReport(output, result)
		return
	}
	if result.Mode == domain.TemplateMode && result.TemplateSource == domain.CustomTemplateSource {
		printCustomTemplateGenerationReport(output, result)
		return
	}

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

func printConfigTemplateGenerationReport(output io.Writer, result domain.GenerationResult) {
	createdFiles := result.CreatedFiles()
	skippedExistingFiles := result.SkippedExistingFiles()
	directoryStatus := "existing"
	if result.ChangeDirectoryCreated {
		directoryStatus = "created"
	}

	fmt.Fprintln(output, "SpecHarbor config template change generated.")
	fmt.Fprintf(output, "Change: %s\n", result.ChangeID)
	fmt.Fprintf(output, "Config template: %s\n", result.ConfigTemplateAlias)
	fmt.Fprintf(output, "Resolved source: %s\n", result.ConfigTemplateSource)
	if result.ConfigTemplateSource == domain.ConfigTemplateSourceRemote {
		fmt.Fprintf(output, "Remote host: %s\n", result.RemoteTemplateHost)
		fmt.Fprintf(output, "Remote format: %s\n", result.RemoteTemplateFormat)
		fmt.Fprintf(output, "Checksum: %s\n", result.ChecksumAlgorithm)
	} else {
		fmt.Fprintf(output, "Resolved template: %s\n", result.ConfigTemplateName)
	}
	if result.ConfigTemplateSource == domain.ConfigTemplateSourceCustom {
		fmt.Fprintf(output, "Template source: %s\n", result.TemplatePath)
	}
	fmt.Fprintf(output, "Change path: %s\n", result.ChangePath)
	fmt.Fprintf(output, "Change directory: %s\n", directoryStatus)

	if len(createdFiles) > 0 {
		fmt.Fprintln(output, "Created files:")
		for _, file := range createdFiles {
			fmt.Fprintf(output, "- %s\n", file)
		}
	}

	if len(skippedExistingFiles) > 0 {
		fmt.Fprintln(output, "Skipped existing files:")
		for _, file := range skippedExistingFiles {
			fmt.Fprintf(output, "- %s\n", file)
		}
	}

	if result.ConfigTemplateSource == domain.ConfigTemplateSourceRemote {
		fmt.Fprintln(output, "Safety:")
		fmt.Fprintln(output, "- Remote access used only the explicit configured alias.")
		fmt.Fprintln(output, "- Checksum was verified before archive parsing.")
		fmt.Fprintf(output, "- Only OpenSpec change files under %s/ were written.\n", result.ChangePath)
		return
	}
	fmt.Fprintf(output, "Only OpenSpec change files under %s/ were written.\n", result.ChangePath)
}

func printHybridGenerationReport(output io.Writer, result domain.HybridGenerationResult) {
	directoryStatus := "existing"
	if result.ChangeDirectoryCreated {
		directoryStatus = "created"
	}

	fmt.Fprintln(output, "SpecHarbor hybrid change generated.")
	fmt.Fprintf(output, "Change: %s\n", result.ChangeID)
	fmt.Fprintln(output, "Mode: hybrid")
	fmt.Fprintf(output, "Source kind: %s\n", result.SelectedSourceKind)
	fmt.Fprintf(output, "Source: %s\n", result.SelectedSourceName)
	fmt.Fprintf(output, "Resolved source: %s\n", result.ResolvedSourceKind)
	if result.ResolvedSourceKind == domain.HybridResolvedSourceRemote {
		fmt.Fprintf(output, "Resolved source name: %s\n", result.ResolvedSourceName)
		fmt.Fprintf(output, "Remote host: %s\n", result.RemoteFacts.Host)
		fmt.Fprintf(output, "Remote format: %s\n", result.RemoteFacts.Format)
		fmt.Fprintf(output, "Checksum: %s\n", result.RemoteFacts.ChecksumAlgorithm)
	} else {
		fmt.Fprintf(output, "Resolved template: %s\n", result.ResolvedSourceName)
	}
	fmt.Fprintf(output, "Title: %s\n", result.Title)
	fmt.Fprintf(output, "Summary: %s\n", result.Summary)
	if result.TypeAvailable {
		fmt.Fprintf(output, "Type: %s\n", result.EffectiveType)
	}
	fmt.Fprintf(output, "Change path: %s\n", result.ChangePath)
	fmt.Fprintf(output, "Change directory: %s\n", directoryStatus)
	printHybridFileList(output, "Created files:", result.CreatedFiles())
	printHybridFileList(output, "Skipped existing files:", result.SkippedExistingFiles())

	if validationResult, ok := result.ValidationResult(); ok {
		fmt.Fprintln(output)
		fmt.Fprintln(output, "Validation:")
		fmt.Fprintf(output, "Status: %s\n", validationResult.Status)
		fmt.Fprintf(output, "Required files: %d\n", len(validationResult.RequiredFiles))
		fmt.Fprintf(output, "Errors: %d\n", validationResult.ErrorCount())
		fmt.Fprintf(output, "Warnings: %d\n", validationResult.WarningCount())
		printValidationFindingGroup(output, "Errors:", findingsBySeverity(validationResult, domain.ValidationFindingSeverityError))
		printValidationFindingGroup(output, "Warnings:", findingsBySeverity(validationResult, domain.ValidationFindingSeverityWarning))
	}

	fmt.Fprintln(output)
	fmt.Fprintln(output, "Safety:")
	fmt.Fprintln(output, "- Provider APIs called: no")
	fmt.Fprintln(output, "- LLM APIs called: no")
	fmt.Fprintln(output, "- Agent commands executed: no")
	fmt.Fprintln(output, "- AI output file imported: no")
	fmt.Fprintln(output, "- Production code modified: no")
	fmt.Fprintln(output, "- Source-control commands run: no")
	fmt.Fprintln(output, "- Auto-commit, auto-push, PR, merge, or archive: no")
}

func printHybridFileList(output io.Writer, title string, files []string) {
	fmt.Fprintln(output, title)
	if len(files) == 0 {
		fmt.Fprintln(output, "- none")
		return
	}
	for _, file := range files {
		fmt.Fprintf(output, "- %s\n", file)
	}
}

func printCustomTemplateGenerationReport(output io.Writer, result domain.GenerationResult) {
	createdFiles := result.CreatedFiles()
	skippedExistingFiles := result.SkippedExistingFiles()
	directoryStatus := "existing"
	if result.ChangeDirectoryCreated {
		directoryStatus = "created"
	}

	fmt.Fprintln(output, "SpecHarbor custom template change generated.")
	fmt.Fprintf(output, "Change: %s\n", result.ChangeID)
	fmt.Fprintf(output, "Template: %s (custom)\n", result.CustomTemplateName)
	fmt.Fprintf(output, "Template source: %s\n", result.TemplatePath)
	fmt.Fprintf(output, "Change path: %s\n", result.ChangePath)
	fmt.Fprintf(output, "Change directory: %s\n", directoryStatus)

	if len(createdFiles) > 0 {
		fmt.Fprintln(output, "Created files:")
		for _, file := range createdFiles {
			fmt.Fprintf(output, "- %s\n", file)
		}
	}

	if len(skippedExistingFiles) > 0 {
		fmt.Fprintln(output, "Skipped existing files:")
		for _, file := range skippedExistingFiles {
			fmt.Fprintf(output, "- %s\n", file)
		}
	}

	fmt.Fprintf(output, "Only OpenSpec change files under %s/ were written.\n", result.ChangePath)
}

func printAIAssistedGenerationReport(output io.Writer, result domain.AIAssistedGenerationResult) {
	generatedFiles := result.GeneratedFiles()
	skippedFiles := result.SkippedFiles()
	overwrittenFiles := result.OverwrittenFiles()
	directoryStatus := "existing"
	if result.ChangeDirectoryCreated {
		directoryStatus = "created"
	}

	fmt.Fprintln(output, "SpecHarbor AI-assisted change generated.")
	fmt.Fprintf(output, "Change: %s\n", result.ChangeID)
	fmt.Fprintf(output, "Source file: %s\n", result.SourcePath)
	fmt.Fprintf(output, "Path: %s\n", result.TargetPath)
	fmt.Fprintf(output, "Directory: %s\n", directoryStatus)
	fmt.Fprintf(output, "Overwrite: %s\n", yesNo(result.Overwrite))
	fmt.Fprintf(output, "Generated files: %d\n", len(generatedFiles))
	fmt.Fprintf(output, "Skipped existing files: %d\n", len(skippedFiles))
	fmt.Fprintf(output, "Overwritten files: %d\n", len(overwrittenFiles))

	printNamedFileList(output, "Generated:", generatedFiles)
	printNamedFileList(output, "Skipped existing:", skippedFiles)
	printNamedFileList(output, "Overwritten:", overwrittenFiles)

	if validationResult, ok := result.ValidationResult(); ok {
		fmt.Fprintln(output)
		fmt.Fprintln(output, "Validation:")
		fmt.Fprintf(output, "Status: %s\n", validationResult.Status)
		fmt.Fprintf(output, "Required files: %d\n", len(validationResult.RequiredFiles))
		fmt.Fprintf(output, "Errors: %d\n", validationResult.ErrorCount())
		fmt.Fprintf(output, "Warnings: %d\n", validationResult.WarningCount())
		printValidationFindingGroup(output, "Errors:", findingsBySeverity(validationResult, domain.ValidationFindingSeverityError))
		printValidationFindingGroup(output, "Warnings:", findingsBySeverity(validationResult, domain.ValidationFindingSeverityWarning))
	}

	printAIAssistedSafetyNotes(output)
}

func printAIAssistedParseFailureReport(output io.Writer, failure *usecase.AIAssistedParseFailure) {
	fmt.Fprintln(output, "SpecHarbor AI-assisted import failed.")
	fmt.Fprintf(output, "Change: %s\n", failure.ChangeID)
	fmt.Fprintf(output, "Source file: %s\n", failure.SourcePath)
	fmt.Fprintln(output, "Parse status: invalid")
	fmt.Fprintln(output, "Files written: 0")
	fmt.Fprintln(output, "No files written: yes")

	findings := failure.ParseResult.Findings()
	if len(findings) > 0 {
		fmt.Fprintln(output)
		fmt.Fprintln(output, "Parse errors:")
		for _, finding := range findings {
			line := fmt.Sprintf("- [%s] %s: %s", finding.Severity, finding.Code, finding.Message)
			var details []string
			if finding.FileName != "" {
				details = append(details, "file: "+finding.FileName)
			}
			if finding.Line > 0 {
				details = append(details, fmt.Sprintf("line: %d", finding.Line))
			}
			if len(details) > 0 {
				line += " (" + strings.Join(details, ", ") + ")"
			}
			fmt.Fprintln(output, line)
		}
	}

	printAIAssistedSafetyNotes(output)
}

func printNamedFileList(output io.Writer, title string, files []string) {
	if len(files) == 0 {
		return
	}

	fmt.Fprintln(output)
	fmt.Fprintln(output, title)
	for _, file := range files {
		fmt.Fprintf(output, "- %s\n", file)
	}
}

func printAIAssistedSafetyNotes(output io.Writer) {
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Safety:")
	fmt.Fprintln(output, "- Provider APIs called: no")
	fmt.Fprintln(output, "- Remote AI services called: no")
	fmt.Fprintln(output, "- Agent commands executed: no")
	fmt.Fprintln(output, "- Production code modified: no")
	fmt.Fprintln(output, "- Source-control commands run: no")
	fmt.Fprintln(output, "- Auto-commit, auto-push, PR, merge, or archive: no")
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
	contextFileSystem := filesystem.NewContextDiscoveryFileSystem()
	contextProvider := usecase.NewDiscoverProjectContext(contextFileSystem)
	renderPrompt := usecase.NewRenderPromptWithContext(roleTemplates, renderer, contextProvider)

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
		fmt.Fprintf(output, "Errors: %d\n", result.ErrorCount())
		fmt.Fprintf(output, "Warnings: %d\n", result.WarningCount())
		printValidationFindingGroup(output, "Warnings:", findingsBySeverity(result, domain.ValidationFindingSeverityWarning))
		return
	}

	fmt.Fprintln(output, "SpecHarbor change is invalid.")
	fmt.Fprintf(output, "Change: %s\n", result.ChangeID)
	fmt.Fprintf(output, "Checked path: %s\n", result.CheckedPath)
	printValidationFindingGroup(output, "Errors:", findingsBySeverity(result, domain.ValidationFindingSeverityError))
	printValidationFindingGroup(output, "Warnings:", findingsBySeverity(result, domain.ValidationFindingSeverityWarning))
}

func printValidationFindingGroup(output io.Writer, title string, findings []domain.ValidationFinding) {
	if len(findings) == 0 {
		return
	}

	fmt.Fprintln(output)
	fmt.Fprintln(output, title)
	for _, finding := range findings {
		line := fmt.Sprintf("- [%s] %s: %s", finding.Severity, finding.Code, finding.Message)
		if finding.RelativePath != "" {
			line += fmt.Sprintf(" (%s)", finding.RelativePath)
		}
		fmt.Fprintln(output, line)
	}
}

func findingsBySeverity(result domain.ValidationResult, severity domain.ValidationFindingSeverity) []domain.ValidationFinding {
	var findings []domain.ValidationFinding
	for _, finding := range result.Findings {
		if finding.Severity == severity {
			findings = append(findings, finding)
		}
	}
	return findings
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
	return now.Format("2006-01-02")
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

func workflowCommand(ctx CommandContext) error {
	if err := parseWorkflowArguments(ctx.Args); err != nil {
		return err
	}

	showWorkflow := usecase.NewShowWorkflow()
	result, err := showWorkflow.Execute()
	if err != nil {
		return err
	}

	printWorkflowReport(ctx.Output, result.Workflow)
	return nil
}

func parseWorkflowArguments(args []string) error {
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			return fmt.Errorf("unsupported flag: %s", arg)
		}
	}

	if len(args) > 0 {
		return fmt.Errorf("unexpected argument: %s", args[0])
	}

	return nil
}

func printWorkflowReport(output io.Writer, workflow domain.Workflow) {
	fmt.Fprintln(output, "SpecHarbor recommended workflow.")
	fmt.Fprintf(output, "Title: %s\n", workflow.Title)
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Steps:")

	for index, step := range workflow.Steps() {
		if index > 0 {
			fmt.Fprintln(output)
		}
		fmt.Fprintf(output, "%d. %s - %s\n", step.Order, step.ID, step.DisplayName)
		fmt.Fprintf(output, "   Mode: %s\n", step.Mode)
		fmt.Fprintf(output, "   Supported by SpecHarbor: %s\n", yesNo(step.Supported))
		fmt.Fprintf(output, "   Advisory only: %s\n", yesNo(step.AdvisoryOnly))
		fmt.Fprintf(output, "   Requires: %s\n", workflowStepRequirements(step.Requires()))
		fmt.Fprintf(output, "   Purpose: %s\n", step.Description)
		printWorkflowCommands(output, step.CommandSuggestions())
		printWorkflowSafetyNotes(output, step.SafetyNotes())
	}

	fmt.Fprintln(output)
	fmt.Fprintln(output, "Notes:")
	fmt.Fprintln(output, "- Command suggestions are advisory and are not executed by this command.")
	fmt.Fprintln(output, "- This command is read-only and does not inspect Git, GitHub, GitLab, CI, provider APIs, agent CLIs, or remote workflow state.")
}

func workflowStepRequirements(requirements []domain.WorkflowStepID) string {
	if len(requirements) == 0 {
		return "none"
	}

	values := make([]string, 0, len(requirements))
	for _, requirement := range requirements {
		values = append(values, string(requirement))
	}
	return strings.Join(values, ", ")
}

func printWorkflowCommands(output io.Writer, suggestions []domain.WorkflowCommandSuggestion) {
	fmt.Fprintln(output, "   Commands:")
	if len(suggestions) == 0 {
		fmt.Fprintln(output, "   - none")
		return
	}

	for _, suggestion := range suggestions {
		if suggestion.Description == "" {
			fmt.Fprintf(output, "   - %s\n", suggestion.Command)
			continue
		}
		fmt.Fprintf(output, "   - %s (%s)\n", suggestion.Command, suggestion.Description)
	}
}

func printWorkflowSafetyNotes(output io.Writer, safetyNotes []string) {
	if len(safetyNotes) == 0 {
		return
	}

	fmt.Fprintln(output, "   Safety:")
	for _, note := range safetyNotes {
		fmt.Fprintf(output, "   - %s\n", note)
	}
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
  context     Discover, index, and retrieve structured local project context
  brief       Collect confirmed project context
  generate    Generate a new OpenSpec change
  prompt      Generate an agent-specific implementation prompt
  validate    Validate an OpenSpec change
  review      Review implementation against a spec
  archive     Archive a completed OpenSpec change
  config      Manage SpecHarbor configuration
  workflow    Show the recommended SpecHarbor workflow
  version     Print SpecHarbor version
  help        Show this help message`)
}
