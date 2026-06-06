package ports

import "github.com/guferreira1/spec-harbor/internal/core/domain"

type PromptTemplateRepository interface {
	TemplateForRole(projectRoot string, role domain.PromptRole) (string, error)
}

type TemplateRenderer interface {
	Render(templateSource string, data map[string]string) (string, error)
}
