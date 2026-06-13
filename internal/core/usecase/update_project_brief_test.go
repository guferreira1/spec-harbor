package usecase

import (
	"errors"
	"strings"
	"testing"

	"github.com/guferreira1/spec-harbor/internal/core/domain"
)

func TestUpdateProjectBriefPrepareRequiresExistingBrief(t *testing.T) {
	fileSystem := newFakeGenerationFileSystem()
	useCase := NewUpdateProjectBrief(fileSystem, fakeProjectBriefUpdateContextProvider{})

	_, err := useCase.Prepare(PrepareProjectBriefUpdateInput{ProjectRoot: "/project"})
	if err == nil || !strings.Contains(err.Error(), "project brief does not exist at .specharbor/project-brief.md") {
		t.Fatalf("Prepare() error = %v, want missing brief error", err)
	}
}

func TestUpdateProjectBriefFinalConfirmedFalseWritesNothing(t *testing.T) {
	fileSystem := newFakeGenerationFileSystem()
	fileSystem.files[projectBriefRelativePath] = renderedProjectBriefForUpdateTest(t)
	useCase := NewUpdateProjectBrief(fileSystem, fakeProjectBriefUpdateContextProvider{
		result: domain.NewContextDiscoveryResult([]domain.ContextSignal{
			updateTestSignal(t, domain.ContextSignalKindStack, "Node.js", domain.ContextSignalClassificationDetectedFact),
		}, nil),
	})

	result, err := useCase.Execute(UpdateProjectBriefInput{
		ProjectRoot: "/project",
		Decisions: domain.ProjectBriefUpdateDecisions{
			FieldDecisions: []domain.ProjectBriefMergeDecision{{
				Field: domain.ProjectBriefFieldStack,
				Kind:  domain.ProjectBriefMergeDecisionAcceptDetectedFact,
				Value: "Node.js",
			}},
		},
		Confirmed: false,
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if result.FileWritten {
		t.Fatalf("FileWritten = true, want false")
	}
	if len(fileSystem.writtenFiles) != 0 {
		t.Fatalf("written files = %v, want none", fileSystem.writtenFiles)
	}
	if len(fileSystem.safelyReadFiles) != 1 || fileSystem.safelyReadFiles[0] != projectBriefRelativePath {
		t.Fatalf("safely read files = %v, want project brief safe read", fileSystem.safelyReadFiles)
	}
	if !strings.Contains(result.Markdown, "Answer: Node.js") {
		t.Fatalf("preview markdown = %q, want accepted value", result.Markdown)
	}
}

func TestUpdateProjectBriefWritesConfirmedUpdate(t *testing.T) {
	fileSystem := newFakeGenerationFileSystem()
	fileSystem.files[projectBriefRelativePath] = renderedProjectBriefForUpdateTest(t)
	useCase := NewUpdateProjectBrief(fileSystem, fakeProjectBriefUpdateContextProvider{
		result: domain.NewContextDiscoveryResult([]domain.ContextSignal{
			updateTestSignal(t, domain.ContextSignalKindStack, "Node.js", domain.ContextSignalClassificationDetectedFact),
		}, nil),
	})

	result, err := useCase.Execute(UpdateProjectBriefInput{
		ProjectRoot: "/project",
		Decisions: domain.ProjectBriefUpdateDecisions{
			FieldDecisions: []domain.ProjectBriefMergeDecision{{
				Field: domain.ProjectBriefFieldStack,
				Kind:  domain.ProjectBriefMergeDecisionAcceptDetectedFact,
				Value: "Node.js",
			}},
		},
		Confirmed: true,
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if !result.FileWritten {
		t.Fatalf("FileWritten = false, want true")
	}
	if !strings.Contains(fileSystem.files[projectBriefRelativePath], "Answer: Node.js") {
		t.Fatalf("updated brief = %q, want Node.js stack", fileSystem.files[projectBriefRelativePath])
	}
}

func TestUpdateProjectBriefSupportsCustomValueAndIgnoreDetectedFact(t *testing.T) {
	fileSystem := newFakeGenerationFileSystem()
	fileSystem.files[projectBriefRelativePath] = renderedProjectBriefForUpdateTest(t)
	useCase := NewUpdateProjectBrief(fileSystem, fakeProjectBriefUpdateContextProvider{
		result: domain.NewContextDiscoveryResult([]domain.ContextSignal{
			updateTestSignal(t, domain.ContextSignalKindBuildCommand, "make build", domain.ContextSignalClassificationDetectedFact),
		}, nil),
	})

	result, err := useCase.Execute(UpdateProjectBriefInput{
		ProjectRoot: "/project",
		Decisions: domain.ProjectBriefUpdateDecisions{
			FieldDecisions: []domain.ProjectBriefMergeDecision{
				{Field: domain.ProjectBriefFieldPurpose, Kind: domain.ProjectBriefMergeDecisionReplaceWithCustom, Value: "Coordinate specs"},
				{Field: domain.ProjectBriefFieldBuild, Kind: domain.ProjectBriefMergeDecisionIgnoreDetectedFact},
			},
		},
		Confirmed: true,
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if !strings.Contains(result.Markdown, "Answer: Coordinate specs") {
		t.Fatalf("updated markdown = %q, want custom purpose", result.Markdown)
	}
	if strings.Contains(result.Markdown, "make build (Source: detected context)") {
		t.Fatalf("ignored detected fact was rendered:\n%s", result.Markdown)
	}
}

func TestUpdateProjectBriefWriteFailurePreservesOriginal(t *testing.T) {
	fileSystem := newFakeGenerationFileSystem()
	original := renderedProjectBriefForUpdateTest(t)
	fileSystem.files[projectBriefRelativePath] = original
	fileSystem.writeErrors[projectBriefRelativePath] = errors.New("disk full")
	useCase := NewUpdateProjectBrief(fileSystem, fakeProjectBriefUpdateContextProvider{
		result: domain.NewContextDiscoveryResult([]domain.ContextSignal{
			updateTestSignal(t, domain.ContextSignalKindStack, "Node.js", domain.ContextSignalClassificationDetectedFact),
		}, nil),
	})

	_, err := useCase.Execute(UpdateProjectBriefInput{
		ProjectRoot: "/project",
		Decisions: domain.ProjectBriefUpdateDecisions{
			FieldDecisions: []domain.ProjectBriefMergeDecision{{
				Field: domain.ProjectBriefFieldStack,
				Kind:  domain.ProjectBriefMergeDecisionAcceptDetectedFact,
				Value: "Node.js",
			}},
		},
		Confirmed: true,
	})
	if err == nil || !strings.Contains(err.Error(), "disk full") {
		t.Fatalf("Execute() error = %v, want disk full", err)
	}
	if fileSystem.files[projectBriefRelativePath] != original {
		t.Fatalf("brief changed after write failure:\n%s", fileSystem.files[projectBriefRelativePath])
	}
}

func (fileSystem *fakeGenerationFileSystem) WriteFileSafely(_ string, relativePath string, contents string) error {
	fileSystem.writtenFiles = append(fileSystem.writtenFiles, relativePath)
	if err := fileSystem.writeErrors[relativePath]; err != nil {
		return err
	}
	fileSystem.files[relativePath] = contents
	return nil
}

type fakeProjectBriefUpdateContextProvider struct {
	result domain.ContextDiscoveryResult
	err    error
}

func (provider fakeProjectBriefUpdateContextProvider) DiscoverPromptContext(_ string) (domain.ContextDiscoveryResult, error) {
	if provider.err != nil {
		return domain.ContextDiscoveryResult{}, provider.err
	}
	return provider.result, nil
}

func renderedProjectBriefForUpdateTest(t *testing.T) string {
	t.Helper()

	brief, err := domain.NewProjectBrief(validProjectBriefAnswers(t), nil, nil)
	if err != nil {
		t.Fatalf("NewProjectBrief() error = %v", err)
	}
	return brief.RenderMarkdown()
}

func updateTestSignal(
	t *testing.T,
	kind domain.ContextSignalKind,
	value string,
	classification domain.ContextSignalClassification,
) domain.ContextSignal {
	t.Helper()

	signal, err := domain.NewContextSignal(domain.ContextSignalInput{
		Kind:           kind,
		Value:          value,
		Classification: classification,
		Confidence:     domain.ContextConfidenceHigh,
		Source: domain.ContextSource{
			Path:     "go.mod",
			Category: domain.ContextSourceCategoryPackageManifest,
		},
	})
	if err != nil {
		t.Fatalf("NewContextSignal() error = %v", err)
	}
	return signal
}
