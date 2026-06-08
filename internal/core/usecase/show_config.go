package usecase

import (
	"errors"
	"fmt"
	"strings"

	"github.com/guferreira1/spec-harbor/internal/core/domain"
	"github.com/guferreira1/spec-harbor/internal/core/ports"
)

const localConfigPath = ".specharbor/config.yml"

type ShowConfigInput struct {
	ProjectRoot string
}

type ShowConfig struct {
	fileSystem ports.ConfigFileSystem
	parser     ports.ConfigParser
}

func NewShowConfig(fileSystem ports.ConfigFileSystem, parser ports.ConfigParser) *ShowConfig {
	return &ShowConfig{
		fileSystem: fileSystem,
		parser:     parser,
	}
}

func (useCase *ShowConfig) Execute(input ShowConfigInput) (domain.ConfigResult, error) {
	if useCase == nil {
		return domain.ConfigResult{}, errors.New("show config use case is required")
	}
	if useCase.fileSystem == nil {
		return domain.ConfigResult{}, errors.New("config filesystem is required")
	}
	if useCase.parser == nil {
		return domain.ConfigResult{}, errors.New("config parser is required")
	}

	projectRoot := strings.TrimSpace(input.ProjectRoot)
	if projectRoot == "" {
		return domain.ConfigResult{}, errors.New("project root is required")
	}

	if err := useCase.requireProjectRoot(projectRoot); err != nil {
		return domain.ConfigResult{}, err
	}
	if err := useCase.requireConfigFile(projectRoot); err != nil {
		return domain.ConfigResult{}, err
	}

	contents, err := useCase.fileSystem.ReadFile(projectRoot, localConfigPath)
	if err != nil {
		return domain.ConfigResult{}, fmt.Errorf("unreadable config %s: %w", localConfigPath, err)
	}

	config, err := useCase.parser.ParseLocalConfig(contents)
	if err != nil {
		return domain.ConfigResult{}, fmt.Errorf("invalid config YAML in %s: %w", localConfigPath, err)
	}
	if !domain.IsSupportedLocalConfigVersion(config.Version) {
		return domain.ConfigResult{}, fmt.Errorf(
			"unsupported config version %d in %s: supported version is %d",
			config.Version,
			localConfigPath,
			domain.SupportedLocalConfigVersion,
		)
	}

	return domain.ConfigResult{
		Path:   localConfigPath,
		Config: config,
	}, nil
}

func (useCase *ShowConfig) requireProjectRoot(projectRoot string) error {
	exists, err := useCase.fileSystem.DirectoryExists(projectRoot, ".")
	if err != nil {
		return fmt.Errorf("check project root: %w", err)
	}
	if !exists {
		return errors.New("project root is unavailable or not a directory")
	}
	return nil
}

func (useCase *ShowConfig) requireConfigFile(projectRoot string) error {
	exists, err := useCase.fileSystem.FileExists(projectRoot, localConfigPath)
	if err != nil {
		return fmt.Errorf("check config file %s: %w", localConfigPath, err)
	}
	if !exists {
		return fmt.Errorf("missing config file: %s", localConfigPath)
	}
	return nil
}
