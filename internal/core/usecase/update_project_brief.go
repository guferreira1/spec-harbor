package usecase

import (
	"errors"
	"fmt"
	"strings"

	"github.com/guferreira1/spec-harbor/internal/core/domain"
	"github.com/guferreira1/spec-harbor/internal/core/ports"
)

type ProjectBriefUpdateContextProvider interface {
	DiscoverPromptContext(projectRoot string) (domain.ContextDiscoveryResult, error)
}

type PrepareProjectBriefUpdateInput struct {
	ProjectRoot string
}

type ProjectBriefUpdatePlan struct {
	TargetPath string
	Proposal   domain.ProjectBriefUpdateProposal
}

type UpdateProjectBriefInput struct {
	ProjectRoot string
	Decisions   domain.ProjectBriefUpdateDecisions
	Confirmed   bool
}

type UpdateProjectBriefResult struct {
	TargetPath  string
	Preview     string
	Markdown    string
	FileWritten bool
}

type UpdateProjectBrief struct {
	fileSystem      ports.ProjectBriefUpdateFileSystem
	contextProvider ProjectBriefUpdateContextProvider
}

func NewUpdateProjectBrief(
	fileSystem ports.ProjectBriefUpdateFileSystem,
	contextProvider ProjectBriefUpdateContextProvider,
) *UpdateProjectBrief {
	return &UpdateProjectBrief{fileSystem: fileSystem, contextProvider: contextProvider}
}

func (useCase *UpdateProjectBrief) Prepare(
	input PrepareProjectBriefUpdateInput,
) (ProjectBriefUpdatePlan, error) {
	if err := useCase.validateDependencies(); err != nil {
		return ProjectBriefUpdatePlan{}, err
	}
	projectRoot, err := validatedProjectBriefUpdateRoot(input.ProjectRoot)
	if err != nil {
		return ProjectBriefUpdatePlan{}, err
	}

	parsed, discovery, err := useCase.readParsedBriefAndDiscovery(projectRoot)
	if err != nil {
		return ProjectBriefUpdatePlan{}, err
	}

	return ProjectBriefUpdatePlan{
		TargetPath: projectBriefRelativePath,
		Proposal:   domain.NewProjectBriefUpdateProposal(parsed, discovery),
	}, nil
}

func (useCase *UpdateProjectBrief) Execute(
	input UpdateProjectBriefInput,
) (UpdateProjectBriefResult, error) {
	if err := useCase.validateDependencies(); err != nil {
		return UpdateProjectBriefResult{}, err
	}
	projectRoot, err := validatedProjectBriefUpdateRoot(input.ProjectRoot)
	if err != nil {
		return UpdateProjectBriefResult{}, err
	}

	parsed, discovery, err := useCase.readParsedBriefAndDiscovery(projectRoot)
	if err != nil {
		return UpdateProjectBriefResult{}, err
	}
	proposal := domain.NewProjectBriefUpdateProposal(parsed, discovery)
	updatedBrief, err := domain.ApplyProjectBriefUpdateDecisions(proposal, input.Decisions)
	if err != nil {
		return UpdateProjectBriefResult{}, err
	}

	result := UpdateProjectBriefResult{
		TargetPath: projectBriefRelativePath,
		Preview:    domain.RenderProjectBriefUpdatePreview(proposal, input.Decisions),
		Markdown:   updatedBrief.RenderMarkdown(),
	}
	if !input.Confirmed {
		return result, nil
	}

	if err := useCase.fileSystem.WriteFileSafely(projectRoot, projectBriefRelativePath, result.Markdown); err != nil {
		return UpdateProjectBriefResult{}, fmt.Errorf("write file %s: %w", projectBriefRelativePath, err)
	}
	result.FileWritten = true
	return result, nil
}

func (useCase *UpdateProjectBrief) validateDependencies() error {
	if useCase == nil {
		return errors.New("update project brief use case is required")
	}
	if useCase.fileSystem == nil {
		return errors.New("project brief update filesystem is required")
	}
	if useCase.contextProvider == nil {
		return errors.New("project brief update context provider is required")
	}
	return nil
}

func (useCase *UpdateProjectBrief) readParsedBriefAndDiscovery(
	projectRoot string,
) (domain.ParsedProjectBrief, domain.ContextDiscoveryResult, error) {
	exists, err := useCase.fileSystem.FileExists(projectRoot, projectBriefRelativePath)
	if err != nil {
		return domain.ParsedProjectBrief{}, domain.ContextDiscoveryResult{}, fmt.Errorf("check file %s: %w", projectBriefRelativePath, err)
	}
	if !exists {
		return domain.ParsedProjectBrief{}, domain.ContextDiscoveryResult{}, fmt.Errorf("project brief does not exist at %s; run specharbor brief first", projectBriefRelativePath)
	}

	contents, err := useCase.fileSystem.ReadFile(projectRoot, projectBriefRelativePath)
	if err != nil {
		return domain.ParsedProjectBrief{}, domain.ContextDiscoveryResult{}, fmt.Errorf("read file %s: %w", projectBriefRelativePath, err)
	}
	parsed, err := domain.ParseProjectBriefMarkdown(contents)
	if err != nil {
		return domain.ParsedProjectBrief{}, domain.ContextDiscoveryResult{}, fmt.Errorf("parse file %s: %w", projectBriefRelativePath, err)
	}

	discovery, err := useCase.contextProvider.DiscoverPromptContext(projectRoot)
	if err != nil {
		return domain.ParsedProjectBrief{}, domain.ContextDiscoveryResult{}, fmt.Errorf("discover project context: %w", err)
	}
	return parsed, discovery, nil
}

func validatedProjectBriefUpdateRoot(projectRoot string) (string, error) {
	trimmed := strings.TrimSpace(projectRoot)
	if trimmed == "" {
		return "", errors.New("project root is required")
	}
	return trimmed, nil
}
