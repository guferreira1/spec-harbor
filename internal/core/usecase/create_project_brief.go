package usecase

import (
	"errors"
	"fmt"
	"strings"

	"github.com/guferreira1/spec-harbor/internal/core/domain"
	"github.com/guferreira1/spec-harbor/internal/core/ports"
)

const (
	projectBriefDirectory    = ".specharbor"
	projectBriefRelativePath = ".specharbor/project-brief.md"
)

type CreateProjectBriefInput struct {
	ProjectRoot    string
	Answers        domain.ProjectBriefAnswers
	ContextSources []domain.ProjectBriefContextSource
	Assumptions    []domain.ProjectBriefAssumption
}

type CreateProjectBriefResult struct {
	TargetPath       string
	DirectoryCreated bool
	FileWritten      bool
}

type CreateProjectBrief struct {
	fileSystem ports.ProjectBriefFileSystem
}

func NewCreateProjectBrief(fileSystem ports.ProjectBriefFileSystem) *CreateProjectBrief {
	return &CreateProjectBrief{fileSystem: fileSystem}
}

func (useCase *CreateProjectBrief) Execute(input CreateProjectBriefInput) (CreateProjectBriefResult, error) {
	if useCase == nil {
		return CreateProjectBriefResult{}, errors.New("create project brief use case is required")
	}
	if useCase.fileSystem == nil {
		return CreateProjectBriefResult{}, errors.New("project brief filesystem is required")
	}

	projectRoot := strings.TrimSpace(input.ProjectRoot)
	if projectRoot == "" {
		return CreateProjectBriefResult{}, errors.New("project root is required")
	}

	brief, err := domain.NewProjectBrief(input.Answers, input.ContextSources, input.Assumptions)
	if err != nil {
		return CreateProjectBriefResult{}, err
	}

	exists, err := useCase.fileSystem.FileExists(projectRoot, projectBriefRelativePath)
	if err != nil {
		return CreateProjectBriefResult{}, fmt.Errorf("check file %s: %w", projectBriefRelativePath, err)
	}
	if exists {
		return CreateProjectBriefResult{}, existingProjectBriefError()
	}

	directoryCreated, err := useCase.ensureProjectBriefDirectory(projectRoot)
	if err != nil {
		return CreateProjectBriefResult{}, err
	}

	written, err := useCase.fileSystem.WriteFileIfAbsent(projectRoot, projectBriefRelativePath, brief.RenderMarkdown())
	if err != nil {
		return CreateProjectBriefResult{}, fmt.Errorf("write file %s: %w", projectBriefRelativePath, err)
	}
	if !written {
		return CreateProjectBriefResult{}, existingProjectBriefError()
	}

	return CreateProjectBriefResult{
		TargetPath:       projectBriefRelativePath,
		DirectoryCreated: directoryCreated,
		FileWritten:      written,
	}, nil
}

func (useCase *CreateProjectBrief) ensureProjectBriefDirectory(projectRoot string) (bool, error) {
	exists, err := useCase.fileSystem.DirectoryExists(projectRoot, projectBriefDirectory)
	if err != nil {
		return false, fmt.Errorf("check directory %s: %w", projectBriefDirectory, err)
	}
	if exists {
		return false, nil
	}

	if err := useCase.fileSystem.CreateDirectory(projectRoot, projectBriefDirectory); err != nil {
		return false, fmt.Errorf("create directory %s: %w", projectBriefDirectory, err)
	}
	return true, nil
}

func existingProjectBriefError() error {
	return fmt.Errorf("project brief already exists at %s; update or merge behavior is out of scope for this change", projectBriefRelativePath)
}
