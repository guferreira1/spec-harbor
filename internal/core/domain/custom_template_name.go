package domain

import (
	"errors"
	"fmt"
	"strings"
)

const maxCustomTemplateNameLength = 128

// CustomTemplateName is a validated, single-path-segment name for a
// project-local custom OpenSpec change template.
type CustomTemplateName struct {
	value string
}

func NewCustomTemplateName(raw string) (CustomTemplateName, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return CustomTemplateName{}, errors.New("custom template name is required")
	}
	if strings.ContainsAny(value, "/\\") {
		return CustomTemplateName{}, errors.New("custom template name must be a single path segment")
	}
	if value == "." || value == ".." || strings.Contains(value, "..") {
		return CustomTemplateName{}, errors.New("custom template name must not contain '.' or '..' path sequences")
	}
	if strings.HasPrefix(value, ".") {
		return CustomTemplateName{}, errors.New("custom template name must not start with '.'")
	}
	if strings.HasPrefix(value, "-") {
		return CustomTemplateName{}, errors.New("custom template name must not start with '-'")
	}
	if len(value) > maxCustomTemplateNameLength {
		return CustomTemplateName{}, fmt.Errorf("custom template name must be at most %d characters", maxCustomTemplateNameLength)
	}
	for _, character := range value {
		if !isChangeIDCharacter(character) {
			return CustomTemplateName{}, fmt.Errorf("custom template name contains unsupported character %q", character)
		}
	}

	return CustomTemplateName{value: value}, nil
}

func (name CustomTemplateName) String() string {
	return name.value
}
