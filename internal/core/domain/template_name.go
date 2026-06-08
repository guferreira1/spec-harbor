package domain

import (
	"fmt"
	"strings"
)

type TemplateName string

const (
	FeatureTemplate  TemplateName = "feature"
	BugfixTemplate   TemplateName = "bugfix"
	DocsTemplate     TemplateName = "docs"
	RefactorTemplate TemplateName = "refactor"
)

func ParseTemplateName(value string) (TemplateName, error) {
	name := TemplateName(strings.TrimSpace(value))
	if name == "" {
		return "", fmt.Errorf("template name is required")
	}
	if !name.IsSupported() {
		return "", fmt.Errorf("unknown template name: %s", name)
	}
	return name, nil
}

func SupportedTemplateNames() []TemplateName {
	return []TemplateName{
		FeatureTemplate,
		BugfixTemplate,
		DocsTemplate,
		RefactorTemplate,
	}
}

func (name TemplateName) IsSupported() bool {
	switch name {
	case FeatureTemplate, BugfixTemplate, DocsTemplate, RefactorTemplate:
		return true
	default:
		return false
	}
}
