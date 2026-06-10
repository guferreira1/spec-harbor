package templates

import (
	"strings"
	"testing"
)

func TestDefaultInitializationTemplatesProvideRequiredDefaults(t *testing.T) {
	defaults := NewDefaultInitializationTemplates()

	requiredPaths := []string{
		"openspec/project.md",
		".specharbor/config.yml",
		".specharbor/rules/global.md",
		".specharbor/rules/spec-author.md",
		".specharbor/rules/implementer.md",
		".specharbor/rules/architecture-reviewer.md",
		".specharbor/rules/test-engineer.md",
		".specharbor/rules/change-reviewer.md",
	}

	for _, path := range requiredPaths {
		contents, err := defaults.ContentFor(path)
		if err != nil {
			t.Fatalf("ContentFor(%q) error = %v", path, err)
		}
		if strings.TrimSpace(contents) == "" {
			t.Fatalf("ContentFor(%q) returned empty contents", path)
		}
	}
}

func TestDefaultInitializationConfigDoesNotContainCredentials(t *testing.T) {
	defaults := NewDefaultInitializationTemplates()

	contents, err := defaults.ContentFor(".specharbor/config.yml")
	if err != nil {
		t.Fatalf("ContentFor() error = %v", err)
	}

	for _, disallowed := range []string{"api_key:", "token:", "secret:", "sk-"} {
		if strings.Contains(strings.ToLower(contents), disallowed) {
			t.Fatalf("config contains disallowed credential marker %q", disallowed)
		}
	}
	if !strings.Contains(contents, "api_key_env: SPECHARBOR_AI_API_KEY") {
		t.Fatalf("config does not reference the expected API key environment variable")
	}
}

func TestDefaultInitializationConfigIncludesTemplateAliasSchema(t *testing.T) {
	defaults := NewDefaultInitializationTemplates()

	contents, err := defaults.ContentFor(".specharbor/config.yml")
	if err != nil {
		t.Fatalf("ContentFor() error = %v", err)
	}

	for _, want := range []string{
		"version: 1",
		"templates:",
		"  aliases: {}",
	} {
		if !strings.Contains(contents, want) {
			t.Fatalf("default config = %q, want %q", contents, want)
		}
	}
}

func TestDefaultInitializationTemplatesRejectUnknownPath(t *testing.T) {
	defaults := NewDefaultInitializationTemplates()

	if _, err := defaults.ContentFor("openspec/changes/example.md"); err == nil {
		t.Fatalf("ContentFor() error = nil, want error")
	}
}
