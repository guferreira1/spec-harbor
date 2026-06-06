package templates

import (
	"embed"
	"fmt"
)

//go:embed defaults/project.md defaults/config.yml defaults/rules/*.md
var initializationDefaults embed.FS

type DefaultInitializationTemplates struct{}

func NewDefaultInitializationTemplates() *DefaultInitializationTemplates {
	return &DefaultInitializationTemplates{}
}

func (templates *DefaultInitializationTemplates) ContentFor(relativePath string) (string, error) {
	templatePath := initializationTemplatePathFor(relativePath)
	if templatePath == "" {
		return "", fmt.Errorf("no initialization default for %s", relativePath)
	}

	contents, err := initializationDefaults.ReadFile(templatePath)
	if err != nil {
		return "", err
	}
	return string(contents), nil
}

func initializationTemplatePathFor(relativePath string) string {
	switch relativePath {
	case "openspec/project.md":
		return "defaults/project.md"
	case ".specharbor/config.yml":
		return "defaults/config.yml"
	case ".specharbor/rules/global.md":
		return "defaults/rules/global.md"
	case ".specharbor/rules/spec-author.md":
		return "defaults/rules/spec-author.md"
	case ".specharbor/rules/implementer.md":
		return "defaults/rules/implementer.md"
	case ".specharbor/rules/architecture-reviewer.md":
		return "defaults/rules/architecture-reviewer.md"
	case ".specharbor/rules/test-engineer.md":
		return "defaults/rules/test-engineer.md"
	case ".specharbor/rules/change-reviewer.md":
		return "defaults/rules/change-reviewer.md"
	default:
		return ""
	}
}
