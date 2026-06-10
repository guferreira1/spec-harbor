package ports

import "github.com/guferreira1/spec-harbor/internal/core/domain"

// GenerationFileSystem provides only the filesystem operations required by
// OpenSpec change generation.
type GenerationFileSystem interface {
	DirectoryExists(root string, relativePath string) (bool, error)
	FileExists(root string, relativePath string) (bool, error)
	CreateDirectory(root string, relativePath string) error
	WriteFileIfAbsent(root string, relativePath string, contents string) (bool, error)
}

// AIAssistedGenerationFileSystem provides only the local read and bounded
// OpenSpec write operations required by AI-assisted from-file generation.
type AIAssistedGenerationFileSystem interface {
	ReadSourceFile(path string) (string, error)
	DirectoryExists(root string, relativePath string) (bool, error)
	FileExists(root string, relativePath string) (bool, error)
	PathExists(root string, relativePath string) (bool, error)
	CreateDirectory(root string, relativePath string) error
	EnsureSafeWriteTarget(root string, relativePath string) error
	WriteFileIfAbsent(root string, relativePath string, contents string) (bool, error)
	WriteFile(root string, relativePath string, contents string) error
}

// CustomTemplateFileSystem provides only the filesystem reads required to
// load a project-local custom template.
type CustomTemplateFileSystem interface {
	DirectoryExists(root string, relativePath string) (bool, error)
	FileExists(root string, relativePath string) (bool, error)
	ReadFile(root string, relativePath string) (string, error)
}

// BlankChangeContent provides deterministic starter content for blank
// OpenSpec change files.
type BlankChangeContent interface {
	ContentFor(relativePath string) (string, error)
}

// TemplateChangeContent provides deterministic starter content for built-in
// OpenSpec change templates.
type TemplateChangeContent interface {
	ContentFor(templateName domain.TemplateName, relativePath string) (string, error)
}

// GuidedChangeContent provides deterministic starter content for guided
// OpenSpec change files.
type GuidedChangeContent interface {
	ContentFor(guidedType domain.GuidedType, title string, summary string, relativePath string) (string, error)
}

// RemoteTemplateFetcher downloads one already-validated remote template
// reference and returns downloaded bytes plus safe response metadata.
type RemoteTemplateFetcher interface {
	FetchRemoteTemplate(request domain.RemoteTemplateFetchRequest) (domain.RemoteTemplateFetchResult, error)
}

// RemoteTemplateBundleReader decodes verified remote template bytes into the
// approved OpenSpec change files without exposing archive implementation types.
type RemoteTemplateBundleReader interface {
	ReadRemoteTemplateBundle(contents []byte, policy domain.RemoteTemplateArchivePolicy) (domain.RemoteTemplateBundle, error)
}

// AgentAssistedAuthoringPromptRenderer renders deterministic dry-run prompts
// for agent-assisted OpenSpec authoring.
type AgentAssistedAuthoringPromptRenderer interface {
	Render(request domain.AgentAssistedAuthoringPromptRequest) (string, error)
}

// AgentRunner runs an already-resolved local agent command for explicit
// run-and-report agent-assisted OpenSpec authoring.
type AgentRunner interface {
	Run(request domain.AgentRunRequest) (domain.AgentRunResult, error)
}
