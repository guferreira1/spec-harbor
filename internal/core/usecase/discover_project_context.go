package usecase

import (
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"sort"
	"strings"
	"unicode"

	"github.com/guferreira1/spec-harbor/internal/core/domain"
	"github.com/guferreira1/spec-harbor/internal/core/ports"
)

const (
	contextDiscoveryDefaultReadLimitBytes int64 = 128 * 1024
	contextDiscoveryBriefReadLimitBytes   int64 = 64 * 1024
	contextDiscoveryMaxDirectoryFiles           = 20
	contextDiscoveryMaxDirectoryDepth           = 2
)

type DiscoverProjectContextInput struct {
	ProjectRoot string
}

type DiscoverProjectContext struct {
	fileSystem ports.ContextDiscoveryFileSystem
}

func NewDiscoverProjectContext(fileSystem ports.ContextDiscoveryFileSystem) *DiscoverProjectContext {
	return &DiscoverProjectContext{fileSystem: fileSystem}
}

func (useCase *DiscoverProjectContext) DiscoverPromptContext(
	projectRoot string,
) (domain.ContextDiscoveryResult, error) {
	return useCase.Execute(DiscoverProjectContextInput{ProjectRoot: projectRoot})
}

func (useCase *DiscoverProjectContext) ProjectBriefExists(projectRoot string) (bool, error) {
	if useCase == nil {
		return false, errors.New("discover project context use case is required")
	}
	if useCase.fileSystem == nil {
		return false, errors.New("context discovery filesystem is required")
	}
	trimmedRoot := strings.TrimSpace(projectRoot)
	if trimmedRoot == "" {
		return false, errors.New("project root is required")
	}
	exists, err := useCase.fileSystem.FileExists(trimmedRoot, ".specharbor/project-brief.md")
	if err != nil {
		return false, fmt.Errorf("check file .specharbor/project-brief.md: %w", err)
	}
	return exists, nil
}

func (useCase *DiscoverProjectContext) Execute(
	input DiscoverProjectContextInput,
) (domain.ContextDiscoveryResult, error) {
	if useCase == nil {
		return domain.ContextDiscoveryResult{}, errors.New("discover project context use case is required")
	}
	if useCase.fileSystem == nil {
		return domain.ContextDiscoveryResult{}, errors.New("context discovery filesystem is required")
	}

	projectRoot := strings.TrimSpace(input.ProjectRoot)
	if projectRoot == "" {
		return domain.ContextDiscoveryResult{}, errors.New("project root is required")
	}

	builder := newContextDiscoveryBuilder(projectRoot, useCase.fileSystem)
	if err := builder.discover(); err != nil {
		return domain.ContextDiscoveryResult{}, err
	}
	return builder.result(), nil
}

type contextDiscoveryBuilder struct {
	projectRoot string
	fileSystem  ports.ContextDiscoveryFileSystem

	signals []domain.ContextSignal
	notes   []domain.ContextDiscoveryNote
	seen    map[string]bool

	goSource            string
	nodeSource          string
	nodePackageManager  string
	mavenSource         string
	gradleSource        string
	pythonSource        string
	rustSource          string
	dockerfileSource    string
	dockerComposeSource string
	cliEntrypoints      []string
	stackValues         map[string]bool
}

func newContextDiscoveryBuilder(
	projectRoot string,
	fileSystem ports.ContextDiscoveryFileSystem,
) *contextDiscoveryBuilder {
	return &contextDiscoveryBuilder{
		projectRoot:        projectRoot,
		fileSystem:         fileSystem,
		seen:               make(map[string]bool),
		stackValues:        make(map[string]bool),
		nodePackageManager: "npm",
	}
}

func (builder *contextDiscoveryBuilder) discover() error {
	steps := []func() error{
		builder.detectProjectBrief,
		builder.detectAgentInstructions,
		builder.detectDocumentation,
		builder.detectOpenSpecSources,
		builder.detectGo,
		builder.detectNode,
		builder.detectJava,
		builder.detectRust,
		builder.detectPython,
		builder.detectContainers,
		builder.detectTaskRunners,
		builder.detectWorkflows,
		builder.detectCLIEntrypoints,
	}

	for _, step := range steps {
		if err := step(); err != nil {
			return err
		}
	}

	builder.addConventionalAssumptions()
	builder.addAmbiguityNotes()
	if len(builder.signals) == 0 && len(builder.notes) == 0 {
		builder.addNote("No supported context sources detected.")
	}
	return nil
}

func (builder *contextDiscoveryBuilder) result() domain.ContextDiscoveryResult {
	return domain.NewContextDiscoveryResult(builder.signals, builder.notes)
}

func (builder *contextDiscoveryBuilder) detectProjectBrief() error {
	const sourcePath = ".specharbor/project-brief.md"
	contents, exists, err := builder.readFile(sourcePath, contextDiscoveryBriefReadLimitBytes)
	if err != nil || !exists {
		return err
	}

	for _, parsed := range parseProjectBriefSignals(contents) {
		builder.addSignal(parsed.kind, parsed.value, domain.ContextSignalClassificationUserConfirmedContext, domain.ContextConfidenceHigh, domain.ContextSource{
			Path:     sourcePath,
			Category: domain.ContextSourceCategoryProjectBrief,
			Evidence: parsed.evidence,
		})
	}
	return nil
}

func (builder *contextDiscoveryBuilder) detectAgentInstructions() error {
	const agentsPath = "AGENTS.md"
	contents, exists, err := builder.readFile(agentsPath, contextDiscoveryDefaultReadLimitBytes)
	if err != nil {
		return err
	}
	if exists {
		builder.addDetected(domain.ContextSignalKindAgentInstructionSource, agentsPath, domain.ContextConfidenceHigh, domain.ContextSource{
			Path:     agentsPath,
			Category: domain.ContextSourceCategoryAgentInstruction,
		})
		builder.detectArchitectureFromText(agentsPath, domain.ContextSourceCategoryAgentInstruction, contents)
		builder.detectDocumentedCommands(agentsPath, domain.ContextSourceCategoryAgentInstruction, contents)
	}

	const rulesDirectory = ".specharbor/rules"
	exists, err = builder.directoryExists(rulesDirectory)
	if err != nil || !exists {
		return err
	}
	builder.addDetected(domain.ContextSignalKindAgentInstructionSource, rulesDirectory+"/", domain.ContextConfidenceHigh, domain.ContextSource{
		Path:     rulesDirectory,
		Category: domain.ContextSourceCategorySpecHarborRules,
	})

	ruleFiles, err := builder.collectFiles(rulesDirectory, []string{".md"}, 1, contextDiscoveryMaxDirectoryFiles)
	if err != nil {
		return err
	}
	for _, ruleFile := range ruleFiles {
		builder.addDetected(domain.ContextSignalKindAgentInstructionSource, ruleFile, domain.ContextConfidenceHigh, domain.ContextSource{
			Path:     ruleFile,
			Category: domain.ContextSourceCategorySpecHarborRules,
		})
	}
	return nil
}

func (builder *contextDiscoveryBuilder) detectDocumentation() error {
	for _, doc := range []struct {
		path          string
		category      domain.ContextSourceCategory
		detectPurpose bool
	}{
		{path: "README.md", category: domain.ContextSourceCategoryReadme, detectPurpose: true},
		{path: "CONTRIBUTING.md", category: domain.ContextSourceCategoryContributing},
	} {
		contents, exists, err := builder.readFile(doc.path, contextDiscoveryDefaultReadLimitBytes)
		if err != nil {
			return err
		}
		if !exists {
			continue
		}
		builder.addDetected(domain.ContextSignalKindDocumentationSource, doc.path, domain.ContextConfidenceHigh, domain.ContextSource{
			Path:     doc.path,
			Category: doc.category,
		})
		if doc.detectPurpose {
			builder.detectPurposeFromMarkdown(doc.path, doc.category, contents)
		}
		builder.detectProjectTypeFromText(doc.path, doc.category, contents)
		builder.detectArchitectureFromText(doc.path, doc.category, contents)
		builder.detectDocumentedCommands(doc.path, doc.category, contents)
	}

	const docsDirectory = "docs"
	exists, err := builder.directoryExists(docsDirectory)
	if err != nil || !exists {
		return err
	}
	builder.addDetected(domain.ContextSignalKindDocumentationSource, docsDirectory+"/", domain.ContextConfidenceHigh, domain.ContextSource{
		Path:     docsDirectory,
		Category: domain.ContextSourceCategoryDocumentation,
	})
	docFiles, err := builder.collectFiles(docsDirectory, []string{".md"}, contextDiscoveryMaxDirectoryDepth, contextDiscoveryMaxDirectoryFiles)
	if err != nil {
		return err
	}
	for _, docFile := range docFiles {
		builder.addDetected(domain.ContextSignalKindDocumentationSource, docFile, domain.ContextConfidenceMedium, domain.ContextSource{
			Path:     docFile,
			Category: domain.ContextSourceCategoryDocumentation,
		})
		contents, exists, err := builder.readFile(docFile, contextDiscoveryDefaultReadLimitBytes)
		if err != nil {
			return err
		}
		if exists {
			builder.detectArchitectureFromText(docFile, domain.ContextSourceCategoryDocumentation, contents)
			builder.detectDocumentedCommands(docFile, domain.ContextSourceCategoryDocumentation, contents)
		}
	}
	return nil
}

func (builder *contextDiscoveryBuilder) detectOpenSpecSources() error {
	const projectPath = "openspec/project.md"
	contents, exists, err := builder.readFile(projectPath, contextDiscoveryDefaultReadLimitBytes)
	if err != nil {
		return err
	}
	if exists {
		builder.addDetected(domain.ContextSignalKindOpenSpecSource, projectPath, domain.ContextConfidenceHigh, domain.ContextSource{
			Path:     projectPath,
			Category: domain.ContextSourceCategoryOpenSpecProject,
		})
		builder.detectPurposeFromOpenSpecProject(contents)
		builder.detectOpenSpecProjectStack(contents)
		builder.detectProjectTypeFromText(projectPath, domain.ContextSourceCategoryOpenSpecProject, contents)
		builder.detectArchitectureFromText(projectPath, domain.ContextSourceCategoryOpenSpecProject, contents)
	}

	const specsDirectory = "openspec/specs"
	exists, err = builder.directoryExists(specsDirectory)
	if err != nil || !exists {
		return err
	}
	builder.addDetected(domain.ContextSignalKindOpenSpecSource, specsDirectory+"/", domain.ContextConfidenceHigh, domain.ContextSource{
		Path:     specsDirectory,
		Category: domain.ContextSourceCategoryOpenSpecSpec,
	})
	specFiles, err := builder.collectFiles(specsDirectory, []string{".md"}, contextDiscoveryMaxDirectoryDepth, contextDiscoveryMaxDirectoryFiles)
	if err != nil {
		return err
	}
	for _, specFile := range specFiles {
		builder.addDetected(domain.ContextSignalKindOpenSpecSource, specFile, domain.ContextConfidenceHigh, domain.ContextSource{
			Path:     specFile,
			Category: domain.ContextSourceCategoryOpenSpecSpec,
		})
		contents, exists, err := builder.readFile(specFile, contextDiscoveryDefaultReadLimitBytes)
		if err != nil {
			return err
		}
		if exists {
			builder.detectArchitectureFromText(specFile, domain.ContextSourceCategoryOpenSpecSpec, contents)
		}
	}
	return nil
}

func (builder *contextDiscoveryBuilder) detectGo() error {
	const sourcePath = "go.mod"
	exists, err := builder.fileExists(sourcePath)
	if err != nil || !exists {
		return err
	}
	builder.goSource = sourcePath
	source := domain.ContextSource{Path: sourcePath, Category: domain.ContextSourceCategoryPackageManifest}
	builder.addStack("Go", domain.ContextConfidenceHigh, source)
	builder.addDetected(domain.ContextSignalKindLanguage, "Go", domain.ContextConfidenceHigh, source)
	builder.addDetected(domain.ContextSignalKindPackageManager, "Go modules", domain.ContextConfidenceHigh, source)
	return nil
}

func (builder *contextDiscoveryBuilder) detectNode() error {
	const sourcePath = "package.json"
	contents, exists, err := builder.readFile(sourcePath, contextDiscoveryDefaultReadLimitBytes)
	if err != nil || !exists {
		return err
	}
	builder.nodeSource = sourcePath
	source := domain.ContextSource{Path: sourcePath, Category: domain.ContextSourceCategoryPackageManifest}
	builder.addStack("Node.js", domain.ContextConfidenceHigh, source)
	builder.addDetected(domain.ContextSignalKindLanguage, "JavaScript", domain.ContextConfidenceMedium, source)

	var manifest packageJSONManifest
	if err := json.Unmarshal([]byte(contents), &manifest); err != nil {
		builder.addNote("Could not parse package.json; only file presence was used for Node.js context.")
		return nil
	}

	if strings.TrimSpace(manifest.PackageManager) != "" {
		builder.nodePackageManager = packageManagerName(manifest.PackageManager)
		builder.addDetected(domain.ContextSignalKindPackageManager, builder.nodePackageManager, domain.ContextConfidenceHigh, domain.ContextSource{
			Path:     sourcePath,
			Category: domain.ContextSourceCategoryPackageManifest,
			Evidence: "packageManager",
		})
	}
	builder.detectPackageJSONScripts(manifest.Scripts, sourcePath)
	builder.detectPackageJSONFrameworks(manifest)
	return nil
}

func (builder *contextDiscoveryBuilder) detectJava() error {
	if exists, err := builder.fileExists("pom.xml"); err != nil || !exists {
		if err != nil {
			return err
		}
	} else {
		builder.mavenSource = "pom.xml"
		source := domain.ContextSource{Path: "pom.xml", Category: domain.ContextSourceCategoryBuildManifest}
		builder.addStack("Java", domain.ContextConfidenceHigh, source)
		builder.addDetected(domain.ContextSignalKindLanguage, "Java", domain.ContextConfidenceHigh, source)
		builder.addDetected(domain.ContextSignalKindPackageManager, "Maven", domain.ContextConfidenceHigh, source)
		contents, exists, err := builder.readFile("pom.xml", contextDiscoveryDefaultReadLimitBytes)
		if err != nil {
			return err
		}
		if exists && strings.Contains(strings.ToLower(contents), "spring-boot") {
			builder.addDetected(domain.ContextSignalKindFramework, "Spring Boot", domain.ContextConfidenceMedium, source)
		}
	}

	for _, gradlePath := range []string{"build.gradle", "build.gradle.kts"} {
		contents, exists, err := builder.readFile(gradlePath, contextDiscoveryDefaultReadLimitBytes)
		if err != nil {
			return err
		}
		if !exists {
			continue
		}
		builder.gradleSource = gradlePath
		source := domain.ContextSource{Path: gradlePath, Category: domain.ContextSourceCategoryBuildManifest}
		builder.addStack("JVM", domain.ContextConfidenceHigh, source)
		builder.addDetected(domain.ContextSignalKindPackageManager, "Gradle", domain.ContextConfidenceHigh, source)
		if strings.Contains(strings.ToLower(contents), "kotlin") {
			builder.addDetected(domain.ContextSignalKindLanguage, "Kotlin", domain.ContextConfidenceMedium, source)
		} else {
			builder.addDetected(domain.ContextSignalKindLanguage, "Java", domain.ContextConfidenceMedium, source)
		}
		if strings.Contains(strings.ToLower(contents), "org.springframework.boot") {
			builder.addDetected(domain.ContextSignalKindFramework, "Spring Boot", domain.ContextConfidenceMedium, source)
		}
		break
	}
	return nil
}

func (builder *contextDiscoveryBuilder) detectRust() error {
	const sourcePath = "Cargo.toml"
	exists, err := builder.fileExists(sourcePath)
	if err != nil || !exists {
		return err
	}
	builder.rustSource = sourcePath
	source := domain.ContextSource{Path: sourcePath, Category: domain.ContextSourceCategoryPackageManifest}
	builder.addStack("Rust", domain.ContextConfidenceHigh, source)
	builder.addDetected(domain.ContextSignalKindLanguage, "Rust", domain.ContextConfidenceHigh, source)
	builder.addDetected(domain.ContextSignalKindPackageManager, "Cargo", domain.ContextConfidenceHigh, source)
	return nil
}

func (builder *contextDiscoveryBuilder) detectPython() error {
	for _, pythonPath := range []string{"pyproject.toml", "requirements.txt"} {
		contents, exists, err := builder.readFile(pythonPath, contextDiscoveryDefaultReadLimitBytes)
		if err != nil {
			return err
		}
		if !exists {
			continue
		}
		if builder.pythonSource == "" {
			builder.pythonSource = pythonPath
		}
		category := domain.ContextSourceCategoryDependencyManifest
		if pythonPath == "pyproject.toml" {
			category = domain.ContextSourceCategoryBuildManifest
		}
		source := domain.ContextSource{Path: pythonPath, Category: category}
		builder.addStack("Python", domain.ContextConfidenceHigh, source)
		builder.addDetected(domain.ContextSignalKindLanguage, "Python", domain.ContextConfidenceHigh, source)
		builder.detectPythonPackageHints(contents, source)
	}
	return nil
}

func (builder *contextDiscoveryBuilder) detectContainers() error {
	for _, containerPath := range []string{"Dockerfile", "docker-compose.yml", "docker-compose.yaml"} {
		exists, err := builder.fileExists(containerPath)
		if err != nil || !exists {
			if err != nil {
				return err
			}
			continue
		}
		source := domain.ContextSource{Path: containerPath, Category: domain.ContextSourceCategoryContainerConfig}
		if containerPath == "Dockerfile" {
			builder.dockerfileSource = containerPath
			builder.addDetected(domain.ContextSignalKindContainerSignal, "Dockerfile", domain.ContextConfidenceHigh, source)
			continue
		}
		if builder.dockerComposeSource == "" {
			builder.dockerComposeSource = containerPath
		}
		builder.addDetected(domain.ContextSignalKindContainerSignal, "Docker Compose", domain.ContextConfidenceHigh, source)
	}
	return nil
}

func (builder *contextDiscoveryBuilder) detectTaskRunners() error {
	makefileContents, exists, err := builder.readFile("Makefile", contextDiscoveryDefaultReadLimitBytes)
	if err != nil {
		return err
	}
	if exists {
		builder.detectMakefileTargets(makefileContents)
	}

	for _, taskfilePath := range []string{"Taskfile.yml", "Taskfile.yaml"} {
		taskfileContents, exists, err := builder.readFile(taskfilePath, contextDiscoveryDefaultReadLimitBytes)
		if err != nil {
			return err
		}
		if !exists {
			continue
		}
		builder.detectTaskfileTasks(taskfilePath, taskfileContents)
		break
	}
	return nil
}

func (builder *contextDiscoveryBuilder) detectWorkflows() error {
	const workflowDirectory = ".github/workflows"
	exists, err := builder.directoryExists(workflowDirectory)
	if err != nil || !exists {
		return err
	}
	builder.addDetected(domain.ContextSignalKindWorkflowSignal, "GitHub Actions", domain.ContextConfidenceHigh, domain.ContextSource{
		Path:     workflowDirectory,
		Category: domain.ContextSourceCategoryWorkflowConfig,
	})
	workflowFiles, err := builder.collectFiles(workflowDirectory, []string{".yml", ".yaml"}, 1, contextDiscoveryMaxDirectoryFiles)
	if err != nil {
		return err
	}
	for _, workflowFile := range workflowFiles {
		builder.addDetected(domain.ContextSignalKindWorkflowSignal, workflowFile, domain.ContextConfidenceHigh, domain.ContextSource{
			Path:     workflowFile,
			Category: domain.ContextSourceCategoryWorkflowConfig,
		})
		contents, exists, err := builder.readFile(workflowFile, contextDiscoveryDefaultReadLimitBytes)
		if err != nil {
			return err
		}
		if exists {
			builder.detectWorkflowCommands(workflowFile, contents)
		}
	}
	return nil
}

func (builder *contextDiscoveryBuilder) detectCLIEntrypoints() error {
	const cmdDirectory = "cmd"
	exists, err := builder.directoryExists(cmdDirectory)
	if err != nil || !exists {
		return err
	}
	entries, err := builder.listDirectory(cmdDirectory)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsSymlink || !entry.IsDirectory {
			continue
		}
		relativePath := path.Join(cmdDirectory, entry.Name)
		if domain.ShouldSkipContextDiscoveryPath(relativePath) {
			continue
		}
		source := domain.ContextSource{Path: relativePath, Category: domain.ContextSourceCategoryRepositoryLayout}
		builder.addDetected(domain.ContextSignalKindCLIEntrypoint, relativePath, domain.ContextConfidenceMedium, source)
		builder.cliEntrypoints = append(builder.cliEntrypoints, relativePath)
	}
	sort.Strings(builder.cliEntrypoints)
	return nil
}

func (builder *contextDiscoveryBuilder) addConventionalAssumptions() {
	if !builder.hasConfirmedOrDetected(domain.ContextSignalKindPackageManager) && builder.nodeSource != "" {
		builder.addAssumption(domain.ContextSignalKindPackageManager, "npm", domain.ContextConfidenceMedium, domain.ContextSource{
			Path:     builder.nodeSource,
			Category: domain.ContextSourceCategoryPackageManifest,
			Evidence: "package.json convention",
		})
	}

	if !builder.hasConfirmedOrDetected(domain.ContextSignalKindTestCommand) {
		for _, suggestion := range builder.testCommandAssumptions() {
			builder.addAssumption(domain.ContextSignalKindTestCommand, suggestion.value, domain.ContextConfidenceMedium, suggestion.source)
		}
	}
	if !builder.hasConfirmedOrDetected(domain.ContextSignalKindBuildCommand) {
		for _, suggestion := range builder.buildCommandAssumptions() {
			builder.addAssumption(domain.ContextSignalKindBuildCommand, suggestion.value, domain.ContextConfidenceMedium, suggestion.source)
		}
	}
	if !builder.hasConfirmedOrDetected(domain.ContextSignalKindRunCommand) {
		for _, suggestion := range builder.runCommandAssumptions() {
			builder.addAssumption(domain.ContextSignalKindRunCommand, suggestion.value, domain.ContextConfidenceMedium, suggestion.source)
		}
	}
}

func (builder *contextDiscoveryBuilder) testCommandAssumptions() []contextCommandSuggestion {
	var suggestions []contextCommandSuggestion
	suggestions = appendCommandSuggestion(suggestions, builder.goSource, "go test ./...", domain.ContextSourceCategoryPackageManifest, "go.mod convention")
	suggestions = appendCommandSuggestion(suggestions, builder.nodeSource, builder.nodePackageManager+" test", domain.ContextSourceCategoryPackageManifest, "package.json convention")
	suggestions = appendCommandSuggestion(suggestions, builder.mavenSource, "mvn test", domain.ContextSourceCategoryBuildManifest, "Maven convention")
	suggestions = appendCommandSuggestion(suggestions, builder.gradleSource, "gradle test", domain.ContextSourceCategoryBuildManifest, "Gradle convention")
	suggestions = appendCommandSuggestion(suggestions, builder.pythonSource, "pytest", sourceCategoryForPython(builder.pythonSource), "Python convention")
	suggestions = appendCommandSuggestion(suggestions, builder.rustSource, "cargo test", domain.ContextSourceCategoryPackageManifest, "Cargo convention")
	return suggestions
}

func (builder *contextDiscoveryBuilder) buildCommandAssumptions() []contextCommandSuggestion {
	var suggestions []contextCommandSuggestion
	suggestions = appendCommandSuggestion(suggestions, builder.goSource, "go build ./...", domain.ContextSourceCategoryPackageManifest, "go.mod convention")
	suggestions = appendCommandSuggestion(suggestions, builder.nodeSource, builder.nodePackageManager+" run build", domain.ContextSourceCategoryPackageManifest, "package.json convention")
	suggestions = appendCommandSuggestion(suggestions, builder.mavenSource, "mvn package", domain.ContextSourceCategoryBuildManifest, "Maven convention")
	suggestions = appendCommandSuggestion(suggestions, builder.gradleSource, "gradle build", domain.ContextSourceCategoryBuildManifest, "Gradle convention")
	suggestions = appendCommandSuggestion(suggestions, builder.rustSource, "cargo build", domain.ContextSourceCategoryPackageManifest, "Cargo convention")
	suggestions = appendCommandSuggestion(suggestions, builder.dockerfileSource, "docker build .", domain.ContextSourceCategoryContainerConfig, "Dockerfile convention")
	return suggestions
}

func (builder *contextDiscoveryBuilder) runCommandAssumptions() []contextCommandSuggestion {
	var suggestions []contextCommandSuggestion
	suggestions = appendCommandSuggestion(suggestions, builder.dockerComposeSource, "docker compose up", domain.ContextSourceCategoryContainerConfig, "Docker Compose convention")
	for _, entrypoint := range builder.cliEntrypoints {
		suggestions = appendCommandSuggestion(suggestions, entrypoint, "go run ./"+entrypoint, domain.ContextSourceCategoryRepositoryLayout, "Go CLI layout convention")
	}
	return suggestions
}

func (builder *contextDiscoveryBuilder) addAmbiguityNotes() {
	if len(builder.stackValues) > 1 {
		builder.addNote("Multiple stack signals detected; review detected facts before treating the project as single-stack.")
	}
}

func (builder *contextDiscoveryBuilder) detectPurposeFromOpenSpecProject(contents string) {
	section := markdownSection(contents, "Purpose")
	if summary := firstPlainMarkdownLine(section); summary != "" {
		builder.addDetected(domain.ContextSignalKindPurposeSummary, summary, domain.ContextConfidenceHigh, domain.ContextSource{
			Path:     "openspec/project.md",
			Category: domain.ContextSourceCategoryOpenSpecProject,
			Evidence: "Purpose",
		})
	}
}

func (builder *contextDiscoveryBuilder) detectPurposeFromMarkdown(
	sourcePath string,
	category domain.ContextSourceCategory,
	contents string,
) {
	summary, evidence, ok := purposeSummaryFromMarkdown(sourcePath, contents)
	if ok {
		builder.addDetected(domain.ContextSignalKindPurposeSummary, summary, domain.ContextConfidenceMedium, domain.ContextSource{
			Path:     sourcePath,
			Category: category,
			Evidence: evidence,
		})
	}
}

func (builder *contextDiscoveryBuilder) detectOpenSpecProjectStack(contents string) {
	section := markdownSection(contents, "Technical stack")
	for _, line := range strings.Split(section, "\n") {
		trimmed := strings.TrimSpace(strings.TrimPrefix(line, "-"))
		label, value, ok := strings.Cut(trimmed, ":")
		if !ok {
			continue
		}
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		source := domain.ContextSource{
			Path:     "openspec/project.md",
			Category: domain.ContextSourceCategoryOpenSpecProject,
			Evidence: strings.TrimSpace(label),
		}
		switch strings.ToLower(strings.TrimSpace(label)) {
		case "language":
			builder.addStack(value, domain.ContextConfidenceHigh, source)
			builder.addDetected(domain.ContextSignalKindLanguage, value, domain.ContextConfidenceHigh, source)
		case "interface":
			if strings.Contains(strings.ToLower(value), "cli") {
				builder.addDetected(domain.ContextSignalKindProjectType, "CLI/tooling project", domain.ContextConfidenceHigh, source)
			}
		case "ci":
			builder.addDetected(domain.ContextSignalKindWorkflowSignal, value, domain.ContextConfidenceMedium, source)
		default:
			builder.addDetected(domain.ContextSignalKindStack, value, domain.ContextConfidenceMedium, source)
		}
	}
}

func (builder *contextDiscoveryBuilder) detectProjectTypeFromText(
	sourcePath string,
	category domain.ContextSourceCategory,
	contents string,
) {
	lower := strings.ToLower(contents)
	if strings.Contains(lower, " cli ") || strings.Contains(lower, " cli for ") ||
		strings.Contains(lower, "command-line") || strings.Contains(lower, "tooling") {
		builder.addDetected(domain.ContextSignalKindProjectType, "CLI/tooling project", domain.ContextConfidenceMedium, domain.ContextSource{
			Path:     sourcePath,
			Category: category,
			Evidence: "CLI/tooling wording",
		})
	}
}

func (builder *contextDiscoveryBuilder) detectArchitectureFromText(
	sourcePath string,
	category domain.ContextSourceCategory,
	contents string,
) {
	seen := make(map[string]bool)
	for _, line := range markdownEvidenceLines(contents) {
		for _, hint := range architectureHintsFromLine(line) {
			if seen[hint.value] {
				continue
			}
			seen[hint.value] = true
			builder.addDetected(domain.ContextSignalKindArchitectureHint, hint.value, domain.ContextConfidenceHigh, domain.ContextSource{
				Path:     sourcePath,
				Category: category,
				Evidence: hint.value,
			})
		}
	}
}

func (builder *contextDiscoveryBuilder) detectDocumentedCommands(
	sourcePath string,
	category domain.ContextSourceCategory,
	contents string,
) {
	seenByKind := make(map[domain.ContextSignalKind]bool)
	for _, line := range strings.Split(contents, "\n") {
		command := normalizeCommandLine(line)
		kind, ok := classifyCommand(command)
		if !ok || seenByKind[kind] {
			continue
		}
		seenByKind[kind] = true
		builder.addDetected(kind, command, domain.ContextConfidenceHigh, domain.ContextSource{
			Path:     sourcePath,
			Category: category,
			Evidence: "documented command",
		})
	}
}

func (builder *contextDiscoveryBuilder) detectPackageJSONScripts(scripts map[string]string, sourcePath string) {
	if len(scripts) == 0 {
		return
	}
	source := func(script string) domain.ContextSource {
		return domain.ContextSource{
			Path:     sourcePath,
			Category: domain.ContextSourceCategoryPackageManifest,
			Evidence: "scripts." + script,
		}
	}
	if _, ok := scripts["test"]; ok {
		builder.addDetected(domain.ContextSignalKindTestCommand, builder.nodePackageManager+" test", domain.ContextConfidenceHigh, source("test"))
	}
	if _, ok := scripts["build"]; ok {
		builder.addDetected(domain.ContextSignalKindBuildCommand, builder.nodePackageManager+" run build", domain.ContextConfidenceHigh, source("build"))
	}
	if _, ok := scripts["dev"]; ok {
		builder.addDetected(domain.ContextSignalKindRunCommand, builder.nodePackageManager+" run dev", domain.ContextConfidenceHigh, source("dev"))
		return
	}
	if _, ok := scripts["start"]; ok {
		builder.addDetected(domain.ContextSignalKindRunCommand, builder.nodePackageManager+" start", domain.ContextConfidenceHigh, source("start"))
	}
}

func (builder *contextDiscoveryBuilder) detectPackageJSONFrameworks(manifest packageJSONManifest) {
	dependencies := make(map[string]string)
	for name := range manifest.Dependencies {
		dependencies[strings.ToLower(name)] = name
	}
	for name := range manifest.DevDependencies {
		dependencies[strings.ToLower(name)] = name
	}
	source := domain.ContextSource{Path: "package.json", Category: domain.ContextSourceCategoryPackageManifest}
	for packageName, framework := range map[string]string{
		"react":         "React",
		"vue":           "Vue",
		"@angular/core": "Angular",
		"next":          "Next.js",
		"nuxt":          "Nuxt",
		"express":       "Express",
		"fastify":       "Fastify",
		"@nestjs/core":  "NestJS",
		"svelte":        "Svelte",
		"vite":          "Vite",
		"hono":          "Hono",
		"typescript":    "TypeScript",
		"@types/node":   "TypeScript",
		"tsx":           "TypeScript",
		"ts-node":       "TypeScript",
	} {
		if _, ok := dependencies[packageName]; !ok {
			continue
		}
		if framework == "TypeScript" {
			builder.addDetected(domain.ContextSignalKindLanguage, "TypeScript", domain.ContextConfidenceMedium, source)
			continue
		}
		builder.addDetected(domain.ContextSignalKindFramework, framework, domain.ContextConfidenceMedium, source)
	}
}

func (builder *contextDiscoveryBuilder) detectPythonPackageHints(contents string, source domain.ContextSource) {
	lower := strings.ToLower(contents)
	if strings.Contains(lower, "[tool.poetry]") {
		builder.addDetected(domain.ContextSignalKindPackageManager, "Poetry", domain.ContextConfidenceMedium, source)
	}
	if strings.Contains(lower, "uv") && strings.Contains(lower, "[tool.uv") {
		builder.addDetected(domain.ContextSignalKindPackageManager, "uv", domain.ContextConfidenceMedium, source)
	}
	for _, framework := range []string{"Django", "FastAPI", "Flask"} {
		if strings.Contains(lower, strings.ToLower(framework)) {
			builder.addDetected(domain.ContextSignalKindFramework, framework, domain.ContextConfidenceMedium, source)
		}
	}
}

func (builder *contextDiscoveryBuilder) detectMakefileTargets(contents string) {
	targets := parseMakeTargets(contents)
	source := func(target string) domain.ContextSource {
		return domain.ContextSource{
			Path:     "Makefile",
			Category: domain.ContextSourceCategoryTaskRunner,
			Evidence: target + " target",
		}
	}
	if targets["test"] {
		builder.addDetected(domain.ContextSignalKindTestCommand, "make test", domain.ContextConfidenceHigh, source("test"))
	}
	if targets["build"] {
		builder.addDetected(domain.ContextSignalKindBuildCommand, "make build", domain.ContextConfidenceHigh, source("build"))
	}
	if targets["run"] {
		builder.addDetected(domain.ContextSignalKindRunCommand, "make run", domain.ContextConfidenceHigh, source("run"))
	}
}

func (builder *contextDiscoveryBuilder) detectTaskfileTasks(sourcePath string, contents string) {
	tasks := parseTaskfileTasks(contents)
	source := func(task string) domain.ContextSource {
		return domain.ContextSource{
			Path:     sourcePath,
			Category: domain.ContextSourceCategoryTaskRunner,
			Evidence: task + " task",
		}
	}
	if tasks["test"] {
		builder.addDetected(domain.ContextSignalKindTestCommand, "task test", domain.ContextConfidenceHigh, source("test"))
	}
	if tasks["build"] {
		builder.addDetected(domain.ContextSignalKindBuildCommand, "task build", domain.ContextConfidenceHigh, source("build"))
	}
	if tasks["run"] {
		builder.addDetected(domain.ContextSignalKindRunCommand, "task run", domain.ContextConfidenceHigh, source("run"))
	}
}

func (builder *contextDiscoveryBuilder) detectWorkflowCommands(sourcePath string, contents string) {
	for _, line := range strings.Split(contents, "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "run:") {
			continue
		}
		command := strings.TrimSpace(strings.Trim(strings.TrimSpace(strings.TrimPrefix(trimmed, "run:")), `"'`))
		kind, ok := classifyCommand(command)
		if !ok {
			continue
		}
		builder.addDetected(kind, command, domain.ContextConfidenceMedium, domain.ContextSource{
			Path:     sourcePath,
			Category: domain.ContextSourceCategoryWorkflowConfig,
			Evidence: "run step",
		})
	}
}

func (builder *contextDiscoveryBuilder) addStack(
	value string,
	confidence domain.ContextConfidence,
	source domain.ContextSource,
) {
	builder.stackValues[value] = true
	builder.addDetected(domain.ContextSignalKindStack, value, confidence, source)
}

func (builder *contextDiscoveryBuilder) addDetected(
	kind domain.ContextSignalKind,
	value string,
	confidence domain.ContextConfidence,
	source domain.ContextSource,
) {
	builder.addSignal(kind, value, domain.ContextSignalClassificationDetectedFact, confidence, source)
}

func (builder *contextDiscoveryBuilder) addAssumption(
	kind domain.ContextSignalKind,
	value string,
	confidence domain.ContextConfidence,
	source domain.ContextSource,
) {
	builder.addSignal(kind, value, domain.ContextSignalClassificationSuggestedAssumption, confidence, source)
}

func (builder *contextDiscoveryBuilder) addSignal(
	kind domain.ContextSignalKind,
	value string,
	classification domain.ContextSignalClassification,
	confidence domain.ContextConfidence,
	source domain.ContextSource,
) {
	signal, err := domain.NewContextSignal(domain.ContextSignalInput{
		Kind:           kind,
		Value:          value,
		Classification: classification,
		Confidence:     confidence,
		Source:         source,
	})
	if err != nil {
		builder.addNote(fmt.Sprintf("Skipped invalid context signal from %s.", source.Path))
		return
	}

	key := strings.Join([]string{
		string(signal.Kind),
		signal.Value,
		string(signal.Classification),
		string(signal.Confidence),
		signal.Source.Path,
		string(signal.Source.Category),
		signal.Source.Evidence,
	}, "\x00")
	if builder.seen[key] {
		return
	}
	builder.seen[key] = true
	builder.signals = append(builder.signals, signal)
}

func (builder *contextDiscoveryBuilder) addNote(message string) {
	note, err := domain.NewContextDiscoveryNote(message)
	if err != nil {
		return
	}
	for _, existing := range builder.notes {
		if existing.Message == note.Message {
			return
		}
	}
	builder.notes = append(builder.notes, note)
}

func (builder *contextDiscoveryBuilder) hasConfirmedOrDetected(kind domain.ContextSignalKind) bool {
	for _, signal := range builder.signals {
		if signal.Kind != kind {
			continue
		}
		if signal.Classification == domain.ContextSignalClassificationDetectedFact ||
			signal.Classification == domain.ContextSignalClassificationUserConfirmedContext {
			return true
		}
	}
	return false
}

func (builder *contextDiscoveryBuilder) fileExists(relativePath string) (bool, error) {
	if domain.ShouldSkipContextDiscoveryPath(relativePath) {
		return false, nil
	}
	exists, err := builder.fileSystem.FileExists(builder.projectRoot, relativePath)
	if err != nil {
		return false, fmt.Errorf("check file %s: %w", relativePath, err)
	}
	return exists, nil
}

func (builder *contextDiscoveryBuilder) directoryExists(relativePath string) (bool, error) {
	if domain.ShouldSkipContextDiscoveryPath(relativePath) {
		return false, nil
	}
	exists, err := builder.fileSystem.DirectoryExists(builder.projectRoot, relativePath)
	if err != nil {
		return false, fmt.Errorf("check directory %s: %w", relativePath, err)
	}
	return exists, nil
}

func (builder *contextDiscoveryBuilder) readFile(relativePath string, maxBytes int64) (string, bool, error) {
	if domain.ShouldSkipContextDiscoveryPath(relativePath) {
		return "", false, nil
	}
	exists, err := builder.fileExists(relativePath)
	if err != nil || !exists {
		return "", false, err
	}
	contents, err := builder.fileSystem.ReadFile(builder.projectRoot, relativePath, maxBytes)
	if err != nil {
		if isSkippableContextReadError(err) {
			builder.addNote(fmt.Sprintf("Skipped %s: %s.", relativePath, err.Error()))
			return "", false, nil
		}
		return "", false, fmt.Errorf("read file %s: %w", relativePath, err)
	}
	return contents, true, nil
}

func (builder *contextDiscoveryBuilder) listDirectory(relativePath string) ([]ports.ContextDiscoveryDirectoryEntry, error) {
	if domain.ShouldSkipContextDiscoveryPath(relativePath) {
		return nil, nil
	}
	entries, err := builder.fileSystem.ListDirectory(builder.projectRoot, relativePath)
	if err != nil {
		return nil, fmt.Errorf("list directory %s: %w", relativePath, err)
	}
	sort.SliceStable(entries, func(i, j int) bool {
		return entries[i].Name < entries[j].Name
	})
	return entries, nil
}

func (builder *contextDiscoveryBuilder) collectFiles(
	directory string,
	extensions []string,
	maxDepth int,
	maxFiles int,
) ([]string, error) {
	var files []string
	if err := builder.collectFilesInto(directory, extensions, maxDepth, maxFiles, &files); err != nil {
		return nil, err
	}
	return files, nil
}

func (builder *contextDiscoveryBuilder) collectFilesInto(
	directory string,
	extensions []string,
	depthRemaining int,
	maxFiles int,
	files *[]string,
) error {
	if len(*files) >= maxFiles {
		return nil
	}
	entries, err := builder.listDirectory(directory)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if len(*files) >= maxFiles {
			return nil
		}
		relativePath := path.Join(directory, entry.Name)
		if entry.IsSymlink || domain.ShouldSkipContextDiscoveryPath(relativePath) {
			continue
		}
		if entry.IsDirectory {
			if depthRemaining <= 0 {
				continue
			}
			if err := builder.collectFilesInto(relativePath, extensions, depthRemaining-1, maxFiles, files); err != nil {
				return err
			}
			continue
		}
		if !entry.IsRegular || !hasSupportedExtension(relativePath, extensions) {
			continue
		}
		*files = append(*files, relativePath)
	}
	return nil
}

func isSkippableContextReadError(err error) bool {
	message := err.Error()
	return strings.Contains(message, "context discovery file exceeds") ||
		strings.Contains(message, "symlink paths are not allowed")
}

type packageJSONManifest struct {
	PackageManager  string            `json:"packageManager"`
	Scripts         map[string]string `json:"scripts"`
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
}

type parsedProjectBriefSignal struct {
	kind     domain.ContextSignalKind
	evidence string
	value    string
}

func parseProjectBriefSignals(contents string) []parsedProjectBriefSignal {
	headingToSignal := map[string]parsedProjectBriefSignal{
		"project type":   {kind: domain.ContextSignalKindProjectType, evidence: "Project type"},
		"purpose":        {kind: domain.ContextSignalKindPurposeSummary, evidence: "Purpose"},
		"target users":   {kind: domain.ContextSignalKindTargetUsers, evidence: "Target users"},
		"stack":          {kind: domain.ContextSignalKindStack, evidence: "Stack"},
		"architecture":   {kind: domain.ContextSignalKindArchitectureHint, evidence: "Architecture"},
		"install":        {kind: domain.ContextSignalKindInstallCommand, evidence: "Install"},
		"test":           {kind: domain.ContextSignalKindTestCommand, evidence: "Test"},
		"build":          {kind: domain.ContextSignalKindBuildCommand, evidence: "Build"},
		"run":            {kind: domain.ContextSignalKindRunCommand, evidence: "Run"},
		"agent behavior": {kind: domain.ContextSignalKindAgentInstructionSource, evidence: "Agent behavior"},
	}

	var current string
	seen := make(map[string]bool)
	var parsed []parsedProjectBriefSignal
	for _, line := range strings.Split(contents, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "## ") && !strings.HasPrefix(trimmed, "### ") {
			heading := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(trimmed, "## ")))
			if _, ok := headingToSignal[heading]; ok {
				current = heading
			} else {
				current = ""
			}
			continue
		}
		if strings.HasPrefix(trimmed, "### ") {
			heading := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(trimmed, "### ")))
			if _, ok := headingToSignal[heading]; ok {
				current = heading
			} else {
				current = ""
			}
			continue
		}
		if current == "" || seen[current] || !strings.HasPrefix(trimmed, "Answer:") {
			continue
		}
		value := strings.TrimSpace(strings.TrimPrefix(trimmed, "Answer:"))
		if value == "" {
			continue
		}
		signal := headingToSignal[current]
		signal.value = value
		parsed = append(parsed, signal)
		seen[current] = true
	}
	return parsed
}

func markdownSection(contents string, heading string) string {
	var lines []string
	inSection := false
	targetHeading := "## " + strings.ToLower(heading)
	for _, line := range strings.Split(contents, "\n") {
		trimmed := strings.TrimSpace(line)
		lower := strings.ToLower(trimmed)
		if strings.HasPrefix(lower, "## ") && !strings.HasPrefix(lower, "### ") {
			if lower == targetHeading {
				inSection = true
				continue
			}
			if inSection {
				break
			}
		}
		if inSection {
			lines = append(lines, line)
		}
	}
	return strings.Join(lines, "\n")
}

func purposeSummaryFromMarkdown(sourcePath string, contents string) (string, string, bool) {
	for _, heading := range []string{"Purpose", "Overview", "About", "What is this?", "Description"} {
		if summary := firstPurposeMarkdownLine(markdownSectionByHeading(contents, heading)); summary != "" {
			return summary, heading, true
		}
	}

	if summary, evidence, ok := labeledPurposeSummary(contents); ok {
		return summary, evidence, true
	}

	title := markdownDocumentTitle(contents)
	if summary := firstExplicitPurposeSentence(contents, title); summary != "" {
		return summary, "explicit purpose sentence", true
	}

	if sourcePath == "README.md" {
		if summary := readmeIntroPurposeSummary(contents, title); summary != "" {
			return summary, "introductory paragraph", true
		}
	}

	return "", "", false
}

func markdownSectionByHeading(contents string, heading string) string {
	var lines []string
	inSection := false
	sectionLevel := 0
	targetHeading := normalizeMarkdownHeading(heading)
	for _, line := range strings.Split(contents, "\n") {
		trimmed := strings.TrimSpace(line)
		level, text, isHeading := parseMarkdownHeading(trimmed)
		if isHeading {
			if inSection && level <= sectionLevel {
				break
			}
			if !inSection && normalizeMarkdownHeading(text) == targetHeading {
				inSection = true
				sectionLevel = level
				continue
			}
		}
		if inSection {
			lines = append(lines, line)
		}
	}
	return strings.Join(lines, "\n")
}

func markdownDocumentTitle(contents string) string {
	for _, line := range strings.Split(contents, "\n") {
		level, text, ok := parseMarkdownHeading(strings.TrimSpace(line))
		if ok && level == 1 {
			return text
		}
	}
	return ""
}

func parseMarkdownHeading(line string) (int, string, bool) {
	if !strings.HasPrefix(line, "#") {
		return 0, "", false
	}
	level := 0
	for level < len(line) && line[level] == '#' {
		level++
	}
	if level == 0 || level > 6 || level >= len(line) || line[level] != ' ' {
		return 0, "", false
	}
	text := strings.TrimSpace(line[level:])
	text = strings.TrimSpace(strings.TrimRight(text, "#"))
	if text == "" {
		return 0, "", false
	}
	return level, text, true
}

func normalizeMarkdownHeading(heading string) string {
	normalized := strings.ToLower(strings.TrimSpace(heading))
	normalized = strings.Trim(normalized, " \t#?:")
	return strings.Join(strings.Fields(normalized), " ")
}

func firstPurposeMarkdownLine(contents string) string {
	inFence := false
	for _, line := range strings.Split(contents, "\n") {
		trimmed := strings.TrimSpace(line)
		if isMarkdownFence(trimmed) {
			inFence = !inFence
			continue
		}
		if inFence ||
			trimmed == "" ||
			strings.HasPrefix(trimmed, "#") ||
			strings.HasPrefix(trimmed, "-") ||
			strings.HasPrefix(trimmed, "* ") ||
			strings.HasPrefix(trimmed, "+ ") ||
			isOrderedMarkdownListLine(trimmed) ||
			strings.HasPrefix(trimmed, ">") ||
			strings.HasPrefix(trimmed, "|") {
			continue
		}
		summary := truncateContextValue(stripMarkdownInline(trimmed), 220)
		if isLikelyPurposeSummary(summary) {
			return summary
		}
	}
	return ""
}

func labeledPurposeSummary(contents string) (string, string, bool) {
	labels := map[string]string{
		"purpose":     "Purpose",
		"overview":    "Overview",
		"description": "Description",
	}
	for _, line := range markdownEvidenceLines(contents) {
		label, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		evidence, ok := labels[normalizeMarkdownHeading(label)]
		if !ok {
			continue
		}
		summary := truncateContextValue(strings.TrimSpace(value), 220)
		if isLikelyPurposeSummary(summary) {
			return summary, evidence, true
		}
	}
	return "", "", false
}

func firstExplicitPurposeSentence(contents string, title string) string {
	for _, line := range markdownEvidenceLines(contents) {
		summary := truncateContextValue(line, 220)
		if !isLikelyPurposeSummary(summary) {
			continue
		}
		lower := strings.ToLower(summary)
		if hasThisProjectPurposePrefix(lower) || hasNamedProjectPurposePrefix(summary, title) {
			return summary
		}
	}
	return ""
}

func readmeIntroPurposeSummary(contents string, title string) string {
	if isGenericMarkdownTitle(title) {
		return ""
	}
	summary := firstPurposeMarkdownLine(markdownIntroAfterTitle(contents))
	if summary == "" {
		return ""
	}
	lower := strings.ToLower(summary)
	if hasThisProjectPurposePrefix(lower) ||
		hasNamedProjectPurposePrefix(summary, title) ||
		hasArticlePurposeShape(lower) {
		return summary
	}
	return ""
}

func markdownIntroAfterTitle(contents string) string {
	var lines []string
	seenTitle := false
	for _, line := range strings.Split(contents, "\n") {
		trimmed := strings.TrimSpace(line)
		level, _, isHeading := parseMarkdownHeading(trimmed)
		if !seenTitle {
			if isHeading && level == 1 {
				seenTitle = true
			}
			continue
		}
		if isHeading {
			break
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func isLikelyPurposeSummary(value string) bool {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" || strings.HasSuffix(trimmed, ":") {
		return false
	}
	if _, ok := classifyCommand(trimmed); ok {
		return false
	}
	if len(strings.Fields(trimmed)) < 4 {
		return false
	}
	lower := strings.ToLower(trimmed)
	procedurePrefixes := []string{
		"before ",
		"after ",
		"run ",
		"use ",
		"install ",
		"set up ",
		"setup ",
		"configure ",
		"copy ",
		"open ",
		"submit ",
		"make sure ",
		"ensure ",
		"follow ",
		"choose ",
		"select ",
		"execute ",
		"check ",
		"verify ",
		"this guide ",
		"this document ",
		"this page ",
		"this readme ",
		"the guide ",
		"the document ",
		"the page ",
		"the documentation ",
		"these docs ",
		"when ",
		"if ",
		"to run ",
		"to build ",
		"to install ",
	}
	return !startsWithAny(lower, procedurePrefixes)
}

func hasThisProjectPurposePrefix(lower string) bool {
	return startsWithAny(lower, []string{
		"this project is ",
		"this project helps ",
		"this project provides ",
		"this project enables ",
		"this project allows ",
		"this project lets ",
		"this project generates ",
		"this project creates ",
		"this project builds ",
		"this project supports ",
	})
}

func hasNamedProjectPurposePrefix(summary string, title string) bool {
	if isGenericMarkdownTitle(title) {
		return false
	}
	name := strings.ToLower(strings.TrimSpace(title))
	lower := strings.ToLower(strings.TrimSpace(summary))
	return startsWithAny(lower, []string{
		name + " is ",
		name + " helps ",
		name + " provides ",
		name + " enables ",
		name + " allows ",
		name + " lets ",
		name + " generates ",
		name + " creates ",
		name + " builds ",
		name + " supports ",
	})
}

func hasArticlePurposeShape(lower string) bool {
	if !startsWithAny(lower, []string{"a ", "an ", "the "}) || !strings.Contains(lower, " for ") {
		return false
	}
	productWords := []string{"cli", "tool", "library", "application", "app", "service", "framework", "platform", "package", "server", "api"}
	for _, word := range productWords {
		if containsNormalizedPhrase(normalizeArchitectureEvidence(lower), word) {
			return true
		}
	}
	return false
}

func isGenericMarkdownTitle(title string) bool {
	normalized := normalizeMarkdownHeading(title)
	if normalized == "" {
		return true
	}
	switch normalized {
	case "project", "readme", "contributing", "contribution", "documentation", "docs", "overview", "about", "description":
		return true
	default:
		return false
	}
}

func startsWithAny(value string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(value, prefix) {
			return true
		}
	}
	return false
}

func firstPlainMarkdownLine(contents string) string {
	for _, line := range strings.Split(contents, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" ||
			strings.HasPrefix(trimmed, "#") ||
			strings.HasPrefix(trimmed, "-") ||
			strings.HasPrefix(trimmed, "```") ||
			strings.HasPrefix(trimmed, "|") {
			continue
		}
		return truncateContextValue(stripMarkdownInline(trimmed), 220)
	}
	return ""
}

type architectureHint struct {
	value string
}

func architectureHintsFromLine(line string) []architectureHint {
	normalized := normalizeArchitectureEvidence(line)
	var hints []architectureHint
	hints = appendArchitectureHintIfContains(hints, normalized, "hexagonal architecture", "Hexagonal Architecture")
	hints = appendArchitectureHintIfContains(hints, normalized, "clean architecture", "Clean Architecture")
	hints = appendArchitectureHintIfContains(hints, normalized, "layered architecture", "Layered Architecture")
	if containsNormalizedPhrase(normalized, "domain driven design") ||
		containsNormalizedPhrase(normalized, "ddd architecture") ||
		containsNormalizedPhrase(normalized, "architecture ddd") {
		hints = appendUniqueArchitectureHint(hints, "Domain-Driven Design")
	}
	if containsNormalizedPhrase(normalized, "mvc architecture") ||
		containsNormalizedPhrase(normalized, "architecture mvc") {
		hints = appendUniqueArchitectureHint(hints, "MVC")
	}
	return hints
}

func appendArchitectureHintIfContains(
	hints []architectureHint,
	normalizedLine string,
	normalizedNeedle string,
	value string,
) []architectureHint {
	if containsNormalizedPhrase(normalizedLine, normalizedNeedle) {
		return appendUniqueArchitectureHint(hints, value)
	}
	return hints
}

func appendUniqueArchitectureHint(hints []architectureHint, value string) []architectureHint {
	for _, hint := range hints {
		if hint.value == value {
			return hints
		}
	}
	return append(hints, architectureHint{value: value})
}

func markdownEvidenceLines(contents string) []string {
	var lines []string
	inFence := false
	for _, line := range strings.Split(contents, "\n") {
		trimmed := strings.TrimSpace(line)
		if isMarkdownFence(trimmed) {
			inFence = !inFence
			continue
		}
		if inFence || trimmed == "" || strings.HasPrefix(trimmed, "|") {
			continue
		}
		line = stripMarkdownEvidencePrefix(trimmed)
		if line == "" {
			continue
		}
		if _, ok := classifyCommand(line); ok {
			continue
		}
		lines = append(lines, stripMarkdownInline(line))
	}
	return lines
}

func stripMarkdownEvidencePrefix(line string) string {
	trimmed := strings.TrimSpace(line)
	for strings.HasPrefix(trimmed, ">") {
		trimmed = strings.TrimSpace(strings.TrimPrefix(trimmed, ">"))
	}
	if _, text, ok := parseMarkdownHeading(trimmed); ok {
		return text
	}
	for _, marker := range []string{"- ", "* ", "+ "} {
		if strings.HasPrefix(trimmed, marker) {
			return strings.TrimSpace(strings.TrimPrefix(trimmed, marker))
		}
	}
	for index, char := range trimmed {
		if !unicode.IsDigit(char) {
			if isOrderedMarkdownListLine(trimmed) {
				return strings.TrimSpace(trimmed[index+2:])
			}
			break
		}
	}
	return trimmed
}

func isOrderedMarkdownListLine(line string) bool {
	for index, char := range line {
		if unicode.IsDigit(char) {
			continue
		}
		return index > 0 && index+1 < len(line) && line[index] == '.' && line[index+1] == ' '
	}
	return false
}

func isMarkdownFence(line string) bool {
	return strings.HasPrefix(line, "```") || strings.HasPrefix(line, "~~~")
}

func normalizeArchitectureEvidence(value string) string {
	var builder strings.Builder
	for _, char := range strings.ToLower(value) {
		if unicode.IsLetter(char) || unicode.IsDigit(char) {
			builder.WriteRune(char)
			continue
		}
		builder.WriteByte(' ')
	}
	return strings.Join(strings.Fields(builder.String()), " ")
}

func containsNormalizedPhrase(normalizedValue string, normalizedPhrase string) bool {
	return strings.Contains(" "+normalizedValue+" ", " "+normalizedPhrase+" ")
}

func stripMarkdownInline(value string) string {
	replacer := strings.NewReplacer("**", "", "__", "", "`", "")
	return strings.TrimSpace(replacer.Replace(value))
}

func truncateContextValue(value string, max int) string {
	if len(value) <= max {
		return value
	}
	return strings.TrimSpace(value[:max]) + "..."
}

func normalizeCommandLine(line string) string {
	trimmed := strings.TrimSpace(line)
	trimmed = strings.TrimPrefix(trimmed, "- ")
	trimmed = strings.TrimPrefix(trimmed, "$ ")
	trimmed = strings.Trim(trimmed, "`")
	return strings.TrimSpace(trimmed)
}

func classifyCommand(command string) (domain.ContextSignalKind, bool) {
	lower := strings.ToLower(strings.TrimSpace(command))
	if lower == "" || strings.Contains(lower, " | ") || strings.Contains(lower, " && ") {
		return "", false
	}
	testPrefixes := []string{"go test", "npm test", "npm run test", "pnpm test", "pnpm run test", "yarn test", "pytest", "python -m pytest", "cargo test", "mvn test", "gradle test", "./gradlew test", "make test", "task test"}
	for _, prefix := range testPrefixes {
		if commandHasPrefix(lower, prefix) {
			return domain.ContextSignalKindTestCommand, true
		}
	}
	buildPrefixes := []string{"go build", "npm run build", "pnpm run build", "pnpm build", "yarn build", "cargo build", "mvn package", "gradle build", "./gradlew build", "make build", "task build", "docker build"}
	for _, prefix := range buildPrefixes {
		if commandHasPrefix(lower, prefix) {
			return domain.ContextSignalKindBuildCommand, true
		}
	}
	runPrefixes := []string{"go run", "npm run dev", "npm start", "pnpm run dev", "pnpm dev", "yarn dev", "make run", "task run", "docker compose up", "docker-compose up"}
	for _, prefix := range runPrefixes {
		if commandHasPrefix(lower, prefix) {
			return domain.ContextSignalKindRunCommand, true
		}
	}
	return "", false
}

func commandHasPrefix(command string, prefix string) bool {
	return command == prefix || strings.HasPrefix(command, prefix+" ")
}

func packageManagerName(value string) string {
	trimmed := strings.TrimSpace(value)
	if name, _, ok := strings.Cut(trimmed, "@"); ok && strings.TrimSpace(name) != "" {
		return strings.TrimSpace(name)
	}
	return trimmed
}

func parseMakeTargets(contents string) map[string]bool {
	targets := make(map[string]bool)
	for _, line := range strings.Split(contents, "\n") {
		if line == "" || strings.HasPrefix(line, "\t") || strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}
		target, _, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		target = strings.TrimSpace(target)
		if target == "" || strings.ContainsAny(target, " \t") || strings.HasPrefix(target, ".") {
			continue
		}
		targets[target] = true
	}
	return targets
}

func parseTaskfileTasks(contents string) map[string]bool {
	tasks := make(map[string]bool)
	inTasks := false
	for _, line := range strings.Split(contents, "\n") {
		if strings.TrimSpace(line) == "tasks:" {
			inTasks = true
			continue
		}
		if !inTasks {
			continue
		}
		if strings.TrimSpace(line) == "" || strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}
		if !strings.HasPrefix(line, "  ") || strings.HasPrefix(line, "    ") {
			continue
		}
		task, _, ok := strings.Cut(strings.TrimSpace(line), ":")
		if ok && task != "" {
			tasks[task] = true
		}
	}
	return tasks
}

func hasSupportedExtension(relativePath string, extensions []string) bool {
	lower := strings.ToLower(relativePath)
	for _, extension := range extensions {
		if strings.HasSuffix(lower, strings.ToLower(extension)) {
			return true
		}
	}
	return false
}

func sourceCategoryForPython(sourcePath string) domain.ContextSourceCategory {
	if sourcePath == "pyproject.toml" {
		return domain.ContextSourceCategoryBuildManifest
	}
	return domain.ContextSourceCategoryDependencyManifest
}

type contextCommandSuggestion struct {
	value  string
	source domain.ContextSource
}

func appendCommandSuggestion(
	suggestions []contextCommandSuggestion,
	sourcePath string,
	value string,
	category domain.ContextSourceCategory,
	evidence string,
) []contextCommandSuggestion {
	if strings.TrimSpace(sourcePath) == "" || strings.TrimSpace(value) == "" {
		return suggestions
	}
	return append(suggestions, contextCommandSuggestion{
		value: value,
		source: domain.ContextSource{
			Path:     sourcePath,
			Category: category,
			Evidence: evidence,
		},
	})
}
