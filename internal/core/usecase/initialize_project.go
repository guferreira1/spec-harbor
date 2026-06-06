package usecase

import (
	"errors"
	"fmt"
	"strings"

	"github.com/guferreira1/spec-harbor/internal/core/ports"
)

type InitializationStatus string

const (
	InitializationStatusInitialized        InitializationStatus = "initialized"
	InitializationStatusAlreadyInitialized InitializationStatus = "already_initialized"
)

type InitializationItemKind string

const (
	InitializationItemKindDirectory InitializationItemKind = "directory"
	InitializationItemKindFile      InitializationItemKind = "file"
)

type InitializeProjectInput struct {
	Root string
}

type InitializationItem struct {
	Kind InitializationItemKind
	Path string
}

type InitializeProjectResult struct {
	Status  InitializationStatus
	Created []InitializationItem
	Skipped []InitializationItem
}

type InitializeProject struct {
	fileSystem ports.InitializationFileSystem
	defaults   ports.InitializationDefaults
}

func NewInitializeProject(fileSystem ports.InitializationFileSystem, defaults ports.InitializationDefaults) *InitializeProject {
	return &InitializeProject{
		fileSystem: fileSystem,
		defaults:   defaults,
	}
}

func (useCase *InitializeProject) Execute(input InitializeProjectInput) (InitializeProjectResult, error) {
	if useCase == nil {
		return InitializeProjectResult{}, errors.New("initialize project use case is required")
	}
	if useCase.fileSystem == nil {
		return InitializeProjectResult{}, errors.New("initialization filesystem is required")
	}
	if useCase.defaults == nil {
		return InitializeProjectResult{}, errors.New("initialization defaults are required")
	}
	if strings.TrimSpace(input.Root) == "" {
		return InitializeProjectResult{}, errors.New("project root is required")
	}

	result := InitializeProjectResult{
		Status: InitializationStatusInitialized,
	}
	allItemsExisted := true

	for _, directory := range requiredInitializationDirectories() {
		existed, err := useCase.fileSystem.DirectoryExists(input.Root, directory)
		if err != nil {
			return InitializeProjectResult{}, fmt.Errorf("check directory %s: %w", directory, err)
		}

		item := InitializationItem{Kind: InitializationItemKindDirectory, Path: directory}
		if existed {
			result.Skipped = append(result.Skipped, item)
			continue
		}

		allItemsExisted = false
		if err := useCase.fileSystem.CreateDirectory(input.Root, directory); err != nil {
			return InitializeProjectResult{}, fmt.Errorf("create directory %s: %w", directory, err)
		}
		result.Created = append(result.Created, item)
	}

	for _, file := range requiredInitializationFiles() {
		existed, err := useCase.fileSystem.FileExists(input.Root, file)
		if err != nil {
			return InitializeProjectResult{}, fmt.Errorf("check file %s: %w", file, err)
		}

		item := InitializationItem{Kind: InitializationItemKindFile, Path: file}
		if existed {
			result.Skipped = append(result.Skipped, item)
			continue
		}

		allItemsExisted = false
		contents, err := useCase.defaults.ContentFor(file)
		if err != nil {
			return InitializeProjectResult{}, fmt.Errorf("load default content for %s: %w", file, err)
		}

		created, err := useCase.fileSystem.WriteFileIfAbsent(input.Root, file, contents)
		if err != nil {
			return InitializeProjectResult{}, fmt.Errorf("write file %s: %w", file, err)
		}
		if created {
			result.Created = append(result.Created, item)
			continue
		}
		result.Skipped = append(result.Skipped, item)
	}

	if allItemsExisted {
		result.Status = InitializationStatusAlreadyInitialized
	}

	return result, nil
}

func requiredInitializationDirectories() []string {
	return []string{
		"openspec",
		"openspec/specs",
		"openspec/changes",
		".specharbor",
		".specharbor/rules",
	}
}

func requiredInitializationFiles() []string {
	return []string{
		"openspec/project.md",
		".specharbor/config.yml",
		".specharbor/rules/global.md",
		".specharbor/rules/spec-author.md",
		".specharbor/rules/implementer.md",
		".specharbor/rules/architecture-reviewer.md",
		".specharbor/rules/test-engineer.md",
		".specharbor/rules/change-reviewer.md",
	}
}
