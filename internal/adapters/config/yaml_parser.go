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

	templateAliases, err := parseConfigTemplateAliases(contents)
	if err != nil {
		return domain.LocalConfig{}, err
	}

	return dto.toDomain(templateAliases), nil
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

func (dto localConfigYAML) toDomain(templateAliases domain.ConfigTemplateAliases) domain.LocalConfig {
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
		Templates: domain.NewConfigTemplates(templateAliases),
	}
}

func parseConfigTemplateAliases(contents string) (domain.ConfigTemplateAliases, error) {
	var document yaml.Node
	if err := yaml.Unmarshal([]byte(contents), &document); err != nil {
		return domain.ConfigTemplateAliases{}, fmt.Errorf("parse local config YAML: %w", err)
	}
	if len(document.Content) == 0 {
		return domain.EmptyConfigTemplateAliases(), nil
	}

	root := document.Content[0]
	if root.Kind == 0 || root.Tag == "!!null" {
		return domain.EmptyConfigTemplateAliases(), nil
	}
	if root.Kind != yaml.MappingNode {
		return domain.ConfigTemplateAliases{}, fmt.Errorf("local config must be a mapping")
	}

	templatesNode := mappingValue(root, "templates")
	if nodeIsMissingOrNull(templatesNode) {
		return domain.EmptyConfigTemplateAliases(), nil
	}
	if templatesNode.Kind != yaml.MappingNode {
		return domain.ConfigTemplateAliases{}, fmt.Errorf("templates must be a mapping")
	}

	aliasesNode := mappingValue(templatesNode, "aliases")
	if nodeIsMissingOrNull(aliasesNode) {
		return domain.EmptyConfigTemplateAliases(), nil
	}
	if aliasesNode.Kind != yaml.MappingNode {
		return domain.ConfigTemplateAliases{}, fmt.Errorf("templates.aliases must be a mapping")
	}

	references := make([]domain.ConfigTemplateReference, 0, len(aliasesNode.Content)/2)
	for index := 0; index < len(aliasesNode.Content); index += 2 {
		aliasNode := aliasesNode.Content[index]
		entryNode := aliasesNode.Content[index+1]
		if aliasNode.Kind != yaml.ScalarNode {
			return domain.ConfigTemplateAliases{}, fmt.Errorf("templates.aliases keys must be strings")
		}

		alias, err := domain.NewConfigTemplateAlias(aliasNode.Value)
		if err != nil {
			return domain.ConfigTemplateAliases{}, fmt.Errorf("invalid config template alias %q: %w", aliasNode.Value, err)
		}

		reference, err := parseConfigTemplateAliasEntry(alias, entryNode)
		if err != nil {
			return domain.ConfigTemplateAliases{}, fmt.Errorf("invalid config template alias %q: %w", alias, err)
		}
		references = append(references, reference)
	}

	return domain.NewConfigTemplateAliases(references), nil
}

func parseConfigTemplateAliasEntry(
	alias domain.ConfigTemplateAlias,
	entryNode *yaml.Node,
) (domain.ConfigTemplateReference, error) {
	if entryNode.Kind != yaml.MappingNode {
		return domain.ConfigTemplateReference{}, fmt.Errorf("config template alias entry must be a mapping")
	}

	var source string
	var template string
	var unsupportedFields []string
	for index := 0; index < len(entryNode.Content); index += 2 {
		fieldNode := entryNode.Content[index]
		valueNode := entryNode.Content[index+1]
		fieldName := fieldNode.Value

		switch fieldName {
		case "source":
			value, err := stringYAMLValue(valueNode, "source")
			if err != nil {
				return domain.ConfigTemplateReference{}, err
			}
			source = value
		case "template":
			value, err := stringYAMLValue(valueNode, "template")
			if err != nil {
				return domain.ConfigTemplateReference{}, err
			}
			template = value
		default:
			unsupportedFields = append(unsupportedFields, fieldName)
		}
	}

	return domain.NewConfigTemplateReference(domain.ConfigTemplateReferenceInput{
		Alias:             alias,
		Source:            source,
		Template:          template,
		UnsupportedFields: unsupportedFields,
	})
}

func mappingValue(mapping *yaml.Node, key string) *yaml.Node {
	for index := 0; index < len(mapping.Content); index += 2 {
		if mapping.Content[index].Value == key {
			return mapping.Content[index+1]
		}
	}
	return nil
}

func nodeIsMissingOrNull(node *yaml.Node) bool {
	return node == nil || node.Kind == 0 || node.Tag == "!!null"
}

func stringYAMLValue(node *yaml.Node, fieldName string) (string, error) {
	if nodeIsMissingOrNull(node) {
		return "", nil
	}
	if node.Kind != yaml.ScalarNode || node.Tag != "!!str" {
		return "", fmt.Errorf("config template %s must be a string", fieldName)
	}
	return node.Value, nil
}
