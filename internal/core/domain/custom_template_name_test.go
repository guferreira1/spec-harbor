package domain

import (
	"strings"
	"testing"
)

func TestNewCustomTemplateNameAcceptsSafeNames(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{name: "kebab-case", raw: "api-feature", want: "api-feature"},
		{name: "snake_case", raw: "data_migration", want: "data_migration"},
		{name: "internal dot", raw: "team.compliance", want: "team.compliance"},
		{name: "digits", raw: "feature2", want: "feature2"},
		{name: "mixed case", raw: "ApiFeature", want: "ApiFeature"},
		{name: "trimmed whitespace", raw: "  api-feature  ", want: "api-feature"},
		{name: "max length", raw: strings.Repeat("a", 128), want: strings.Repeat("a", 128)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			templateName, err := NewCustomTemplateName(test.raw)
			if err != nil {
				t.Fatalf("NewCustomTemplateName(%q) error = %v", test.raw, err)
			}
			if templateName.String() != test.want {
				t.Fatalf("String() = %q, want %q", templateName.String(), test.want)
			}
		})
	}
}

func TestNewCustomTemplateNameRejectsUnsafeNames(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{name: "empty", raw: "", want: "custom template name is required"},
		{name: "whitespace only", raw: "   ", want: "custom template name is required"},
		{name: "forward slash", raw: "nested/template", want: "custom template name must be a single path segment"},
		{name: "backslash", raw: "nested\\template", want: "custom template name must be a single path segment"},
		{name: "absolute path", raw: "/etc/passwd", want: "custom template name must be a single path segment"},
		{name: "traversal with separator", raw: "../escape", want: "custom template name must be a single path segment"},
		{name: "single dot", raw: ".", want: "custom template name must not contain '.' or '..' path sequences"},
		{name: "double dot", raw: "..", want: "custom template name must not contain '.' or '..' path sequences"},
		{name: "embedded dot dot", raw: "a..b", want: "custom template name must not contain '.' or '..' path sequences"},
		{name: "leading dot", raw: ".hidden", want: "custom template name must not start with '.'"},
		{name: "leading dash", raw: "-flag-like", want: "custom template name must not start with '-'"},
		{name: "over max length", raw: strings.Repeat("a", 129), want: "custom template name must be at most 128 characters"},
		{name: "space inside", raw: "api feature", want: `custom template name contains unsupported character ' '`},
		{name: "colon", raw: "api:feature", want: `custom template name contains unsupported character ':'`},
		{name: "shell metacharacter", raw: "api;feature", want: `custom template name contains unsupported character ';'`},
		{name: "template token", raw: "api{feature", want: `custom template name contains unsupported character '{'`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewCustomTemplateName(test.raw)
			if err == nil {
				t.Fatalf("NewCustomTemplateName(%q) error = nil, want %q", test.raw, test.want)
			}
			if err.Error() != test.want {
				t.Fatalf("NewCustomTemplateName(%q) error = %q, want %q", test.raw, err.Error(), test.want)
			}
		})
	}
}
