package config

import (
	"fmt"

	"github.com/guferreira1/spec-harbor/internal/core/domain"
	"gopkg.in/yaml.v3"
)

type YAMLParser struct{}

func NewYAMLParser() *YAMLParser {
	return &YAMLParser{}
}

func (parser *YAMLParser) ParseLocalConfig(contents string) (domain.LocalConfig, error) {
	var dto localConfigYAML
	if err := yaml.Unmarshal([]byte(contents), &dto); err != nil {
		return domain.LocalConfig{}, fmt.Errorf("parse local config YAML: %w", err)
	}

	return dto.toDomain(), nil
}

type localConfigYAML struct {
	Version    int                  `yaml:"version"`
	Defaults   configDefaultsYAML   `yaml:"defaults"`
	Validation configValidationYAML `yaml:"validation"`
	Review     configReviewYAML     `yaml:"review"`
	Archive    configArchiveYAML    `yaml:"archive"`
	Scan       configScanYAML       `yaml:"scan"`
	Output     configOutputYAML     `yaml:"output"`
}

type configDefaultsYAML struct {
	AgentRole      string `yaml:"agent_role"`
	GenerationMode string `yaml:"generation_mode"`
}

type configValidationYAML struct {
	RequireAllChangeFiles bool `yaml:"require_all_change_files"`
}

type configReviewYAML struct {
	RequireCompletedTasks bool `yaml:"require_completed_tasks"`
}

type configArchiveYAML struct {
	DateLayout string `yaml:"date_layout"`
}

type configScanYAML struct {
	IncludeCommonProjectFiles bool `yaml:"include_common_project_files"`
}

type configOutputYAML struct {
	Format string `yaml:"format"`
}

func (dto localConfigYAML) toDomain() domain.LocalConfig {
	return domain.LocalConfig{
		Version: dto.Version,
		Defaults: domain.ConfigDefaults{
			AgentRole:      dto.Defaults.AgentRole,
			GenerationMode: dto.Defaults.GenerationMode,
		},
		Validation: domain.ConfigValidation{
			RequireAllChangeFiles: dto.Validation.RequireAllChangeFiles,
		},
		Review: domain.ConfigReview{
			RequireCompletedTasks: dto.Review.RequireCompletedTasks,
		},
		Archive: domain.ConfigArchive{
			DateLayout: dto.Archive.DateLayout,
		},
		Scan: domain.ConfigScan{
			IncludeCommonProjectFiles: dto.Scan.IncludeCommonProjectFiles,
		},
		Output: domain.ConfigOutput{
			Format: dto.Output.Format,
		},
	}
}
