package domain

import (
	"fmt"
	"strings"
)

// AllowedCustomTemplateFiles lists the only filenames read from a custom
// template directory and written into a generated change.
func AllowedCustomTemplateFiles() []string {
	return RequiredOpenSpecChangeFiles()
}

// CustomTemplate is a validated project-local change template holding the
// contents of the five required OpenSpec change files.
type CustomTemplate struct {
	name  CustomTemplateName
	files map[string]string
}

func NewCustomTemplate(name CustomTemplateName, files map[string]string) (CustomTemplate, error) {
	if name.String() == "" {
		return CustomTemplate{}, fmt.Errorf("custom template name is required")
	}

	var missingFiles []string
	for _, requiredFile := range AllowedCustomTemplateFiles() {
		if _, exists := files[requiredFile]; !exists {
			missingFiles = append(missingFiles, requiredFile)
		}
	}
	if len(missingFiles) > 0 {
		return CustomTemplate{}, fmt.Errorf(
			"custom template %s is missing required files: %s",
			name,
			strings.Join(missingFiles, ", "),
		)
	}

	copied := make(map[string]string, len(AllowedCustomTemplateFiles()))
	for _, requiredFile := range AllowedCustomTemplateFiles() {
		contents := files[requiredFile]
		if strings.TrimSpace(contents) == "" {
			return CustomTemplate{}, fmt.Errorf("custom template file %s/%s is empty", name, requiredFile)
		}
		copied[requiredFile] = contents
	}

	return CustomTemplate{name: name, files: copied}, nil
}

func (template CustomTemplate) Name() CustomTemplateName {
	return template.name
}

func (template CustomTemplate) Files() map[string]string {
	copied := make(map[string]string, len(template.files))
	for file, contents := range template.files {
		copied[file] = contents
	}
	return copied
}

func (template CustomTemplate) Render(changeID string, title string, summary string) map[string]string {
	rendered := make(map[string]string, len(template.files))
	for file, contents := range template.files {
		rendered[file] = RenderCustomTemplateContent(contents, changeID, title, summary)
	}
	return rendered
}

// RenderCustomTemplateContent performs deterministic variable substitution.
// {{change_id}} is always replaced; {{title}} and {{summary}} only when a
// non-empty trimmed value is provided. All other tokens stay verbatim.
func RenderCustomTemplateContent(source string, changeID string, title string, summary string) string {
	replacements := []string{"{{change_id}}", changeID}
	if trimmedTitle := strings.TrimSpace(title); trimmedTitle != "" {
		replacements = append(replacements, "{{title}}", trimmedTitle)
	}
	if trimmedSummary := strings.TrimSpace(summary); trimmedSummary != "" {
		replacements = append(replacements, "{{summary}}", trimmedSummary)
	}
	return strings.NewReplacer(replacements...).Replace(source)
}
