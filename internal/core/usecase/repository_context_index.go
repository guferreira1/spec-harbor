package usecase

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/guferreira1/spec-harbor/internal/core/domain"
	"github.com/guferreira1/spec-harbor/internal/core/ports"
)

type RepositoryContextIndexInput struct {
	ProjectRoot string
	Mode        domain.RepositoryContextIndexMode
}

type BuildRepositoryContextIndex struct {
	fileSystem ports.RepositoryContextIndexFileSystem
	limits     domain.RepositoryContextIndexLimits
}

func NewBuildRepositoryContextIndex(
	fileSystem ports.RepositoryContextIndexFileSystem,
) *BuildRepositoryContextIndex {
	return &BuildRepositoryContextIndex{
		fileSystem: fileSystem,
		limits:     domain.DefaultRepositoryContextIndexLimits(),
	}
}

func NewBuildRepositoryContextIndexWithLimits(
	fileSystem ports.RepositoryContextIndexFileSystem,
	limits domain.RepositoryContextIndexLimits,
) *BuildRepositoryContextIndex {
	return &BuildRepositoryContextIndex{
		fileSystem: fileSystem,
		limits:     limits,
	}
}

func (useCase *BuildRepositoryContextIndex) Execute(
	input RepositoryContextIndexInput,
) (domain.RepositoryContextIndexReport, error) {
	if useCase == nil {
		return domain.RepositoryContextIndexReport{}, errors.New("repository context index use case is required")
	}
	if useCase.fileSystem == nil {
		return domain.RepositoryContextIndexReport{}, errors.New("repository context index filesystem is required")
	}
	projectRoot := strings.TrimSpace(input.ProjectRoot)
	if projectRoot == "" {
		return domain.RepositoryContextIndexReport{}, errors.New("project root is required")
	}
	if err := useCase.limits.Validate(); err != nil {
		return domain.RepositoryContextIndexReport{}, err
	}

	mode := input.Mode
	if mode == "" {
		mode = domain.RepositoryContextIndexModeReport
	}
	if !isSupportedRepositoryContextIndexMode(mode) {
		return domain.RepositoryContextIndexReport{}, fmt.Errorf("unsupported repository context index mode: %s", mode)
	}

	current, err := useCase.buildCurrentIndex(projectRoot)
	if err != nil {
		return domain.RepositoryContextIndexReport{}, err
	}

	report := domain.RepositoryContextIndexReport{
		Mode:      mode,
		IndexPath: domain.RepositoryContextIndexPath,
		Index:     current,
	}

	switch mode {
	case domain.RepositoryContextIndexModeReport:
		report.Status = domain.RepositoryContextIndexStatusBuilt
		if current.Truncated {
			report.Status = domain.RepositoryContextIndexStatusTruncated
		}
		return report, nil
	case domain.RepositoryContextIndexModeWrite:
		contents, err := encodeRepositoryContextIndex(current)
		if err != nil {
			return domain.RepositoryContextIndexReport{}, err
		}
		if err := useCase.fileSystem.CreateDirectory(projectRoot, ".specharbor"); err != nil {
			return domain.RepositoryContextIndexReport{}, err
		}
		if err := useCase.fileSystem.WriteFileSafely(projectRoot, domain.RepositoryContextIndexPath, contents); err != nil {
			return domain.RepositoryContextIndexReport{}, err
		}
		report.Status = domain.RepositoryContextIndexStatusWritten
		return report, nil
	case domain.RepositoryContextIndexModeCheck:
		return useCase.checkStoredIndex(projectRoot, current, report)
	default:
		return domain.RepositoryContextIndexReport{}, fmt.Errorf("unsupported repository context index mode: %s", mode)
	}
}

func (useCase *BuildRepositoryContextIndex) checkStoredIndex(
	projectRoot string,
	current domain.RepositoryContextIndex,
	report domain.RepositoryContextIndexReport,
) (domain.RepositoryContextIndexReport, error) {
	exists, err := useCase.fileSystem.FileExists(projectRoot, domain.RepositoryContextIndexPath)
	if err != nil {
		return domain.RepositoryContextIndexReport{}, err
	}
	if !exists {
		report.Status = domain.RepositoryContextIndexStatusMissing
		report.ErrorMessage = "repository context index is missing"
		return report, nil
	}

	contents, err := useCase.fileSystem.ReadFileSafely(projectRoot, domain.RepositoryContextIndexPath)
	if err != nil {
		report.Status = domain.RepositoryContextIndexStatusInvalid
		report.ErrorMessage = err.Error()
		return report, nil
	}

	var stored domain.RepositoryContextIndex
	if err := json.Unmarshal([]byte(contents), &stored); err != nil {
		report.Status = domain.RepositoryContextIndexStatusInvalid
		report.ErrorMessage = err.Error()
		return report, nil
	}

	stored, err = domain.NormalizeRepositoryContextIndex(stored)
	if err != nil {
		report.Status = domain.RepositoryContextIndexStatusInvalid
		report.ErrorMessage = err.Error()
		return report, nil
	}
	reasons := domain.CompareRepositoryContextIndexes(stored, current)
	if len(reasons) > 0 {
		report.Status = domain.RepositoryContextIndexStatusStale
		report.StaleReasons = reasons
		return report, nil
	}

	report.Status = domain.RepositoryContextIndexStatusCurrent
	return report, nil
}

func (useCase *BuildRepositoryContextIndex) buildCurrentIndex(
	projectRoot string,
) (domain.RepositoryContextIndex, error) {
	builder := &repositoryContextIndexBuilder{
		projectRoot: projectRoot,
		fileSystem:  useCase.fileSystem,
		limits:      useCase.limits,
	}
	return builder.build()
}

type repositoryContextIndexBuilder struct {
	projectRoot string
	fileSystem  ports.RepositoryContextIndexFileSystem
	limits      domain.RepositoryContextIndexLimits

	entries        []domain.RepositoryContextIndexEntry
	skipped        []domain.RepositoryContextIndexSkipped
	totalFileBytes int64
	truncated      bool
}

type repositoryContextSourceSpec struct {
	relativePath           string
	category               domain.ContextSourceCategory
	fileType               domain.RepositoryContextIndexFileType
	languageOrEcosystem    string
	classificationHints    []domain.RepositoryContextIndexClassificationHint
	sourceEvidenceCategory string
}

type repositoryContextDirectorySpec struct {
	relativePath string
	extensions   []string
	category     domain.ContextSourceCategory
	fileType     domain.RepositoryContextIndexFileType
	language     string
	evidence     string
}

func (builder *repositoryContextIndexBuilder) build() (domain.RepositoryContextIndex, error) {
	for _, spec := range repositoryContextFixedSources() {
		if err := builder.addFile(spec); err != nil {
			return domain.RepositoryContextIndex{}, err
		}
	}
	for _, spec := range repositoryContextDirectorySources() {
		if err := builder.collectDirectory(spec.relativePath, spec, builder.limits.MaxDirectoryDepth); err != nil {
			return domain.RepositoryContextIndex{}, err
		}
	}

	rootMarker := "none"
	exists, err := builder.fileSystem.FileExists(builder.projectRoot, "openspec/project.md")
	if err != nil {
		return domain.RepositoryContextIndex{}, fmt.Errorf("check file openspec/project.md: %w", err)
	}
	if exists {
		rootMarker = "openspec/project.md"
	}
	return domain.NewRepositoryContextIndex(rootMarker, builder.limits, builder.entries, builder.skipped, builder.truncated)
}

func (builder *repositoryContextIndexBuilder) addFile(spec repositoryContextSourceSpec) error {
	relativePath, err := domain.SafeRepositoryContextIndexPath(spec.relativePath)
	if err != nil {
		builder.addSkipFallback(spec.relativePath, domain.RepositoryContextIndexSkipUnsafePath)
		return nil
	}
	if relativePath == domain.RepositoryContextIndexPath {
		return nil
	}
	if reason, skip := domain.ShouldSkipRepositoryContextIndexPath(relativePath); skip {
		builder.addSkip(relativePath, reason)
		return nil
	}

	exists, err := builder.fileSystem.FileExists(builder.projectRoot, relativePath)
	if err != nil {
		return fmt.Errorf("check file %s: %w", relativePath, err)
	}
	if !exists {
		return nil
	}
	if len(builder.entries) >= builder.limits.MaxIndexedFiles {
		builder.addSkip(relativePath, domain.RepositoryContextIndexSkipMaxFilesReached)
		builder.truncated = true
		return nil
	}

	metadata, err := builder.fileSystem.FileMetadata(builder.projectRoot, relativePath)
	if err != nil {
		return fmt.Errorf("read metadata %s: %w", relativePath, err)
	}
	if metadata.SizeBytes > builder.limits.MaxFileSizeBytes {
		builder.addSkip(relativePath, domain.RepositoryContextIndexSkipFileTooLarge)
		builder.truncated = true
		return nil
	}
	if builder.totalFileBytes+metadata.SizeBytes > builder.limits.MaxTotalFileBytes {
		builder.addSkip(relativePath, domain.RepositoryContextIndexSkipMaxTotalBytesReached)
		builder.truncated = true
		return nil
	}

	contents, err := builder.fileSystem.ReadFileBytes(builder.projectRoot, relativePath, builder.limits.MaxFileSizeBytes)
	if err != nil {
		return fmt.Errorf("read file bytes %s: %w", relativePath, err)
	}
	entry, err := domain.NewRepositoryContextIndexEntry(domain.RepositoryContextIndexEntryInput{
		Path:                   relativePath,
		SourceCategory:         spec.category,
		FileType:               spec.fileType,
		LanguageOrEcosystem:    spec.languageOrEcosystem,
		SizeBytes:              metadata.SizeBytes,
		ContentHash:            repositoryContextIndexHash(contents),
		ModifiedTime:           metadata.ModifiedTime.UTC().Format(time.RFC3339Nano),
		SupportedForRetrieval:  true,
		ClassificationHints:    spec.classificationHints,
		SourceEvidenceCategory: spec.sourceEvidenceCategory,
	})
	if err != nil {
		return err
	}
	builder.entries = append(builder.entries, entry)
	builder.totalFileBytes += metadata.SizeBytes
	return nil
}

func (builder *repositoryContextIndexBuilder) collectDirectory(
	relativePath string,
	spec repositoryContextDirectorySpec,
	depthRemaining int,
) error {
	if reason, skip := domain.ShouldSkipRepositoryContextIndexPath(relativePath); skip {
		builder.addSkip(relativePath, reason)
		return nil
	}
	exists, err := builder.fileSystem.DirectoryExists(builder.projectRoot, relativePath)
	if err != nil {
		return fmt.Errorf("check directory %s: %w", relativePath, err)
	}
	if !exists {
		return nil
	}
	entries, err := builder.fileSystem.ListDirectory(builder.projectRoot, relativePath)
	if err != nil {
		return fmt.Errorf("list directory %s: %w", relativePath, err)
	}
	sort.SliceStable(entries, func(i, j int) bool { return entries[i].Name < entries[j].Name })

	for _, entry := range entries {
		childPath := path.Join(relativePath, entry.Name)
		if childPath == domain.RepositoryContextIndexPath {
			continue
		}
		if entry.IsSymlink {
			builder.addSkip(childPath, domain.RepositoryContextIndexSkipSymlink)
			continue
		}
		if reason, skip := domain.ShouldSkipRepositoryContextIndexPath(childPath); skip {
			builder.addSkip(childPath, reason)
			continue
		}
		if entry.IsDirectory {
			if depthRemaining <= 0 {
				builder.addSkip(childPath, domain.RepositoryContextIndexSkipUnsupportedSource)
				continue
			}
			if err := builder.collectDirectory(childPath, spec, depthRemaining-1); err != nil {
				return err
			}
			continue
		}
		if !entry.IsRegular {
			builder.addSkip(childPath, domain.RepositoryContextIndexSkipUnsupportedSource)
			continue
		}
		if !repositoryContextIndexHasSupportedExtension(childPath, spec.extensions) {
			builder.addSkip(childPath, domain.RepositoryContextIndexSkipUnsupportedSource)
			continue
		}
		if err := builder.addFile(repositoryContextSourceSpec{
			relativePath:           childPath,
			category:               spec.category,
			fileType:               spec.fileType,
			languageOrEcosystem:    spec.language,
			classificationHints:    []domain.RepositoryContextIndexClassificationHint{domain.RepositoryContextIndexHintDetectedFact, domain.RepositoryContextIndexHintInventoryMetadata},
			sourceEvidenceCategory: spec.evidence,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (builder *repositoryContextIndexBuilder) addSkip(
	relativePath string,
	reason domain.RepositoryContextIndexSkipReason,
) {
	if len(builder.skipped) >= builder.limits.MaxSkippedRecords {
		builder.truncated = true
		return
	}
	skip, err := domain.NewRepositoryContextIndexSkipped(relativePath, reason)
	if err != nil {
		builder.addSkipFallback(relativePath, domain.RepositoryContextIndexSkipUnsafePath)
		return
	}
	builder.skipped = append(builder.skipped, skip)
}

func (builder *repositoryContextIndexBuilder) addSkipFallback(
	relativePath string,
	reason domain.RepositoryContextIndexSkipReason,
) {
	if len(builder.skipped) >= builder.limits.MaxSkippedRecords {
		builder.truncated = true
		return
	}
	fallbackPath := "unsafe-path"
	if strings.TrimSpace(relativePath) != "" {
		fallbackPath = "unsafe-path/" + repositoryContextIndexSafePathToken(relativePath)
	}
	skip, err := domain.NewRepositoryContextIndexSkipped(fallbackPath, reason)
	if err != nil {
		return
	}
	builder.skipped = append(builder.skipped, skip)
	builder.truncated = true
}

func repositoryContextFixedSources() []repositoryContextSourceSpec {
	detected := []domain.RepositoryContextIndexClassificationHint{
		domain.RepositoryContextIndexHintDetectedFact,
		domain.RepositoryContextIndexHintInventoryMetadata,
	}
	return []repositoryContextSourceSpec{
		sourceSpec("AGENTS.md", domain.ContextSourceCategoryAgentInstruction, domain.RepositoryContextIndexFileTypeMarkdown, "Markdown", "agent_instruction", detected),
		sourceSpec("README.md", domain.ContextSourceCategoryReadme, domain.RepositoryContextIndexFileTypeMarkdown, "Markdown", "readme", detected),
		sourceSpec("CONTRIBUTING.md", domain.ContextSourceCategoryContributing, domain.RepositoryContextIndexFileTypeMarkdown, "Markdown", "contributing", detected),
		sourceSpec(".specharbor/project-brief.md", domain.ContextSourceCategoryProjectBrief, domain.RepositoryContextIndexFileTypeMarkdown, "SpecHarbor", "project_brief", []domain.RepositoryContextIndexClassificationHint{domain.RepositoryContextIndexHintUserConfirmedContext, domain.RepositoryContextIndexHintInventoryMetadata}),
		sourceSpec("openspec/project.md", domain.ContextSourceCategoryOpenSpecProject, domain.RepositoryContextIndexFileTypeMarkdown, "OpenSpec", "openspec_project", detected),
		sourceSpec("package.json", domain.ContextSourceCategoryPackageManifest, domain.RepositoryContextIndexFileTypeJSON, "Node.js", "package_manifest", detected),
		sourceSpec("go.mod", domain.ContextSourceCategoryPackageManifest, domain.RepositoryContextIndexFileTypeGoModule, "Go", "package_manifest", detected),
		sourceSpec("pom.xml", domain.ContextSourceCategoryBuildManifest, domain.RepositoryContextIndexFileTypeXML, "JVM", "build_manifest", detected),
		sourceSpec("build.gradle", domain.ContextSourceCategoryBuildManifest, domain.RepositoryContextIndexFileTypeGradle, "JVM", "build_manifest", detected),
		sourceSpec("build.gradle.kts", domain.ContextSourceCategoryBuildManifest, domain.RepositoryContextIndexFileTypeGradle, "JVM", "build_manifest", detected),
		sourceSpec("Cargo.toml", domain.ContextSourceCategoryPackageManifest, domain.RepositoryContextIndexFileTypeTOML, "Rust", "package_manifest", detected),
		sourceSpec("pyproject.toml", domain.ContextSourceCategoryBuildManifest, domain.RepositoryContextIndexFileTypeTOML, "Python", "build_manifest", detected),
		sourceSpec("requirements.txt", domain.ContextSourceCategoryDependencyManifest, domain.RepositoryContextIndexFileTypeText, "Python", "dependency_manifest", detected),
		sourceSpec("Dockerfile", domain.ContextSourceCategoryContainerConfig, domain.RepositoryContextIndexFileTypeDockerfile, "Docker", "container_config", detected),
		sourceSpec("docker-compose.yml", domain.ContextSourceCategoryContainerConfig, domain.RepositoryContextIndexFileTypeYAML, "Docker", "container_config", detected),
		sourceSpec("docker-compose.yaml", domain.ContextSourceCategoryContainerConfig, domain.RepositoryContextIndexFileTypeYAML, "Docker", "container_config", detected),
		sourceSpec("Makefile", domain.ContextSourceCategoryTaskRunner, domain.RepositoryContextIndexFileTypeMakefile, "Make", "task_runner", detected),
		sourceSpec("Taskfile.yml", domain.ContextSourceCategoryTaskRunner, domain.RepositoryContextIndexFileTypeYAML, "Task", "task_runner", detected),
		sourceSpec("Taskfile.yaml", domain.ContextSourceCategoryTaskRunner, domain.RepositoryContextIndexFileTypeYAML, "Task", "task_runner", detected),
	}
}

func repositoryContextDirectorySources() []repositoryContextDirectorySpec {
	return []repositoryContextDirectorySpec{
		{relativePath: "docs", extensions: []string{".md"}, category: domain.ContextSourceCategoryDocumentation, fileType: domain.RepositoryContextIndexFileTypeMarkdown, language: "Markdown", evidence: "documentation"},
		{relativePath: "openspec/specs", extensions: []string{".md"}, category: domain.ContextSourceCategoryOpenSpecSpec, fileType: domain.RepositoryContextIndexFileTypeMarkdown, language: "OpenSpec", evidence: "openspec_spec"},
		{relativePath: ".specharbor/rules", extensions: []string{".md"}, category: domain.ContextSourceCategorySpecHarborRules, fileType: domain.RepositoryContextIndexFileTypeMarkdown, language: "SpecHarbor", evidence: "specharbor_rules"},
		{relativePath: ".github/workflows", extensions: []string{".yml", ".yaml"}, category: domain.ContextSourceCategoryWorkflowConfig, fileType: domain.RepositoryContextIndexFileTypeYAML, language: "GitHub Actions", evidence: "workflow_config"},
	}
}

func sourceSpec(
	relativePath string,
	category domain.ContextSourceCategory,
	fileType domain.RepositoryContextIndexFileType,
	language string,
	evidence string,
	hints []domain.RepositoryContextIndexClassificationHint,
) repositoryContextSourceSpec {
	return repositoryContextSourceSpec{
		relativePath:           relativePath,
		category:               category,
		fileType:               fileType,
		languageOrEcosystem:    language,
		classificationHints:    append([]domain.RepositoryContextIndexClassificationHint(nil), hints...),
		sourceEvidenceCategory: evidence,
	}
}

func repositoryContextIndexHash(contents []byte) string {
	sum := sha256.Sum256(contents)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func repositoryContextIndexHasSupportedExtension(relativePath string, extensions []string) bool {
	lower := strings.ToLower(relativePath)
	for _, extension := range extensions {
		if strings.HasSuffix(lower, strings.ToLower(extension)) {
			return true
		}
	}
	return false
}

func repositoryContextIndexSafePathToken(value string) string {
	var builder strings.Builder
	for _, char := range strings.ToLower(value) {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') {
			builder.WriteRune(char)
			continue
		}
		builder.WriteByte('-')
	}
	token := strings.Trim(builder.String(), "-")
	if token == "" {
		return "path"
	}
	return token
}

func encodeRepositoryContextIndex(index domain.RepositoryContextIndex) (string, error) {
	normalized, err := domain.NormalizeRepositoryContextIndex(index)
	if err != nil {
		return "", err
	}
	contents, err := json.MarshalIndent(normalized, "", "  ")
	if err != nil {
		return "", err
	}
	return string(contents) + "\n", nil
}

func isSupportedRepositoryContextIndexMode(mode domain.RepositoryContextIndexMode) bool {
	switch mode {
	case domain.RepositoryContextIndexModeReport,
		domain.RepositoryContextIndexModeWrite,
		domain.RepositoryContextIndexModeCheck:
		return true
	default:
		return false
	}
}
