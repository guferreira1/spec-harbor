package templates

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/guferreira1/spec-harbor/internal/core/domain"
)

type RolePromptTemplates struct{}

func NewRolePromptTemplates() *RolePromptTemplates {
	return &RolePromptTemplates{}
}

func (templates *RolePromptTemplates) TemplateForRole(projectRoot string, role domain.PromptRole) (string, error) {
	templatePath := filepath.Join(projectRoot, "agent-prompts", "roles", string(role)+".md.tmpl")
	contents, err := os.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("read role prompt template %s: %w", role, err)
	}
	return string(contents), nil
}

type PromptTemplateRenderer struct{}

func NewPromptTemplateRenderer() *PromptTemplateRenderer {
	return &PromptTemplateRenderer{}
}

func (renderer *PromptTemplateRenderer) Render(templateSource string, data map[string]string) (string, error) {
	parsed, err := template.New("prompt").Funcs(templateFunctions(data)).Option("missingkey=error").Parse(templateSource)
	if err != nil {
		return "", err
	}

	var output bytes.Buffer
	if err := parsed.Execute(&output, data); err != nil {
		return "", err
	}
	return output.String(), nil
}

func templateFunctions(data map[string]string) template.FuncMap {
	functions := make(template.FuncMap, len(data))
	for key, value := range data {
		value := value
		functions[key] = func() string {
			return value
		}
	}
	return functions
}
