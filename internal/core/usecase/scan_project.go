package usecase

import (
	"errors"
	"fmt"
	"strings"

	"github.com/guferreira1/spec-harbor/internal/core/domain"
	"github.com/guferreira1/spec-harbor/internal/core/ports"
)

const scanProjectRootListing = "."

type ScanProjectInput struct {
	ProjectRoot string
}

type ScanProject struct {
	fileSystem ports.ScanFileSystem
}

func NewScanProject(fileSystem ports.ScanFileSystem) *ScanProject {
	return &ScanProject{fileSystem: fileSystem}
}

func (useCase *ScanProject) Execute(input ScanProjectInput) (domain.ScanResult, error) {
	if useCase == nil {
		return domain.ScanResult{}, errors.New("scan project use case is required")
	}
	if useCase.fileSystem == nil {
		return domain.ScanResult{}, errors.New("scan filesystem is required")
	}

	projectRoot := strings.TrimSpace(input.ProjectRoot)
	if projectRoot == "" {
		return domain.ScanResult{}, errors.New("project root is required")
	}

	topLevelEntries, err := useCase.fileSystem.ListDirectoryNames(projectRoot, scanProjectRootListing)
	if err != nil {
		return domain.ScanResult{}, fmt.Errorf("list project root: %w", err)
	}

	var matchedRules []domain.ScanSignalRule
	for _, rule := range domain.ScanSignalRules() {
		matched, err := useCase.matchRule(projectRoot, rule, topLevelEntries)
		if err != nil {
			return domain.ScanResult{}, err
		}
		if matched {
			matchedRules = append(matchedRules, rule)
		}
	}

	return domain.AssembleScanResult(projectRoot, matchedRules), nil
}

func (useCase *ScanProject) matchRule(projectRoot string, rule domain.ScanSignalRule, topLevelEntries []string) (bool, error) {
	switch rule.Kind {
	case domain.ScanProbeKindFile:
		exists, err := useCase.fileSystem.FileExists(projectRoot, rule.Path)
		if err != nil {
			return false, fmt.Errorf("check file %s: %w", rule.Path, err)
		}
		return exists, nil
	case domain.ScanProbeKindDirectory:
		exists, err := useCase.fileSystem.DirectoryExists(projectRoot, rule.Path)
		if err != nil {
			return false, fmt.Errorf("check directory %s: %w", rule.Path, err)
		}
		return exists, nil
	case domain.ScanProbeKindFileSuffix:
		for _, entry := range topLevelEntries {
			if strings.HasSuffix(entry, rule.Path) {
				return true, nil
			}
		}
		return false, nil
	default:
		return false, nil
	}
}
