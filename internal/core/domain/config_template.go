package domain

import (
	"errors"
	"fmt"
	"strings"
)

const maxConfigTemplateAliasLength = 128

type ConfigTemplateAlias struct {
	value string
}

func NewConfigTemplateAlias(raw string) (ConfigTemplateAlias, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return ConfigTemplateAlias{}, errors.New("config template alias is required")
	}
	if strings.ContainsAny(value, "/\\") {
		return ConfigTemplateAlias{}, errors.New("config template alias must be a single path segment")
	}
	if value == "." || value == ".." || strings.Contains(value, "..") {
		return ConfigTemplateAlias{}, errors.New("config template alias must not contain '.' or '..' path sequences")
	}
	if strings.HasPrefix(value, ".") {
		return ConfigTemplateAlias{}, errors.New("config template alias must not start with '.'")
	}
	if strings.HasPrefix(value, "-") {
		return ConfigTemplateAlias{}, errors.New("config template alias must not start with '-'")
	}
	if len(value) > maxConfigTemplateAliasLength {
		return ConfigTemplateAlias{}, fmt.Errorf("config template alias must be at most %d characters", maxConfigTemplateAliasLength)
	}
	for _, character := range value {
		if !isChangeIDCharacter(character) {
			return ConfigTemplateAlias{}, fmt.Errorf("config template alias contains unsupported character %q", character)
		}
	}

	return ConfigTemplateAlias{value: value}, nil
}

func (alias ConfigTemplateAlias) String() string {
	return alias.value
}

type ConfigTemplateSourceKind string

const (
	ConfigTemplateSourceBuiltin ConfigTemplateSourceKind = "builtin"
	ConfigTemplateSourceCustom  ConfigTemplateSourceKind = "custom"
)

func ParseConfigTemplateSourceKind(raw string) (ConfigTemplateSourceKind, error) {
	value := ConfigTemplateSourceKind(strings.TrimSpace(raw))
	if value == "" {
		return "", errors.New("config template source is required")
	}
	switch value {
	case ConfigTemplateSourceBuiltin, ConfigTemplateSourceCustom:
		return value, nil
	default:
		return "", fmt.Errorf("unsupported config template source: %s", value)
	}
}

type ConfigTemplateReferenceInput struct {
	Alias             ConfigTemplateAlias
	Source            string
	Template          string
	UnsupportedFields []string
}

type ConfigTemplateReference struct {
	alias              ConfigTemplateAlias
	sourceKind         ConfigTemplateSourceKind
	template           string
	builtInTemplate    TemplateName
	customTemplateName CustomTemplateName
}

func NewConfigTemplateReference(input ConfigTemplateReferenceInput) (ConfigTemplateReference, error) {
	if input.Alias.String() == "" {
		return ConfigTemplateReference{}, errors.New("config template alias is required")
	}

	sourceKind, err := ParseConfigTemplateSourceKind(input.Source)
	if err != nil {
		return ConfigTemplateReference{}, err
	}

	if len(input.UnsupportedFields) > 0 {
		return ConfigTemplateReference{}, fmt.Errorf("unsupported config template field %q", input.UnsupportedFields[0])
	}

	template := strings.TrimSpace(input.Template)
	if template == "" {
		return ConfigTemplateReference{}, errors.New("config template reference template is required")
	}

	reference := ConfigTemplateReference{
		alias:      input.Alias,
		sourceKind: sourceKind,
		template:   template,
	}

	switch sourceKind {
	case ConfigTemplateSourceBuiltin:
		builtInTemplate, err := ParseTemplateName(template)
		if err != nil {
			return ConfigTemplateReference{}, err
		}
		reference.builtInTemplate = builtInTemplate
	case ConfigTemplateSourceCustom:
		customTemplateName, err := NewCustomTemplateName(template)
		if err != nil {
			return ConfigTemplateReference{}, err
		}
		reference.customTemplateName = customTemplateName
	}

	return reference, nil
}

func (reference ConfigTemplateReference) Alias() ConfigTemplateAlias {
	return reference.alias
}

func (reference ConfigTemplateReference) SourceKind() ConfigTemplateSourceKind {
	return reference.sourceKind
}

func (reference ConfigTemplateReference) Template() string {
	return reference.template
}

func (reference ConfigTemplateReference) BuiltInTemplateName() (TemplateName, bool) {
	if reference.sourceKind != ConfigTemplateSourceBuiltin {
		return "", false
	}
	return reference.builtInTemplate, true
}

func (reference ConfigTemplateReference) CustomTemplateName() (CustomTemplateName, bool) {
	if reference.sourceKind != ConfigTemplateSourceCustom {
		return CustomTemplateName{}, false
	}
	return reference.customTemplateName, true
}

type ConfigTemplateAliases struct {
	references map[string]ConfigTemplateReference
}

func NewConfigTemplateAliases(references []ConfigTemplateReference) ConfigTemplateAliases {
	copied := make(map[string]ConfigTemplateReference, len(references))
	for _, reference := range references {
		copied[reference.Alias().String()] = reference
	}
	return ConfigTemplateAliases{references: copied}
}

func EmptyConfigTemplateAliases() ConfigTemplateAliases {
	return NewConfigTemplateAliases(nil)
}

func (aliases ConfigTemplateAliases) Lookup(alias ConfigTemplateAlias) (ConfigTemplateReference, error) {
	if alias.String() == "" {
		return ConfigTemplateReference{}, errors.New("config template alias is required")
	}
	reference, exists := aliases.references[alias.String()]
	if !exists {
		return ConfigTemplateReference{}, fmt.Errorf("config template alias not found: %s", alias)
	}
	return reference, nil
}

func (aliases ConfigTemplateAliases) References() map[string]ConfigTemplateReference {
	copied := make(map[string]ConfigTemplateReference, len(aliases.references))
	for alias, reference := range aliases.references {
		copied[alias] = reference
	}
	return copied
}

func (aliases ConfigTemplateAliases) Copy() ConfigTemplateAliases {
	return ConfigTemplateAliases{references: aliases.References()}
}

func (aliases ConfigTemplateAliases) Len() int {
	return len(aliases.references)
}
