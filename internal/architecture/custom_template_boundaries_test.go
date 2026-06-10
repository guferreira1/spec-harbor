package architecture_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCustomTemplateReadPortStaysNarrowAndReadOnly(t *testing.T) {
	portPath := filepath.Join("..", "core", "ports", "generation.go")
	portContents, err := os.ReadFile(portPath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", portPath, err)
	}
	portSource := string(portContents)

	start := strings.Index(portSource, "type CustomTemplateFileSystem interface {")
	if start == -1 {
		t.Fatalf("%s does not define the core-owned CustomTemplateFileSystem port", portPath)
	}
	end := strings.Index(portSource[start:], "}")
	if end == -1 {
		t.Fatalf("%s contains an unterminated CustomTemplateFileSystem interface", portPath)
	}
	interfaceBody := portSource[start : start+end]

	for _, want := range []string{"DirectoryExists", "FileExists", "ReadFile"} {
		if !strings.Contains(interfaceBody, want) {
			t.Fatalf("CustomTemplateFileSystem port missing expected read method %q", want)
		}
	}
	for _, forbidden := range []string{
		"WriteFile",
		"CreateDirectory",
		"MoveDirectory",
		"Remove",
		"ListDirectory",
	} {
		if strings.Contains(interfaceBody, forbidden) {
			t.Fatalf("CustomTemplateFileSystem port contains forbidden write/list method %q", forbidden)
		}
	}
}

func TestCustomTemplateDomainStaysPure(t *testing.T) {
	for _, fileName := range []string{
		"custom_template_name.go",
		"custom_template.go",
		"template_source.go",
	} {
		sourcePath := filepath.Join("..", "core", "domain", fileName)
		contents, err := os.ReadFile(sourcePath)
		if err != nil {
			t.Fatalf("ReadFile(%q) error = %v", sourcePath, err)
		}
		source := string(contents)

		for _, forbidden := range []string{
			`"os"`,
			`"os/exec"`,
			`"net`,
			`"path/filepath"`,
			"internal/adapters",
			"internal/platform",
		} {
			if strings.Contains(source, forbidden) {
				t.Fatalf("%s contains forbidden dependency %q", sourcePath, forbidden)
			}
		}
	}
}

func TestCustomTemplateGenerationUsesFixedRootAndChangePathWrites(t *testing.T) {
	useCasePath := filepath.Join("..", "core", "usecase", "generate_change.go")
	contents, err := os.ReadFile(useCasePath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", useCasePath, err)
	}
	source := string(contents)

	if !strings.Contains(source, `customTemplatesDirectory = ".specharbor/templates"`) {
		t.Fatalf("%s does not pin the fixed custom template root .specharbor/templates", useCasePath)
	}
	if !strings.Contains(source, "ports.CustomTemplateFileSystem") {
		t.Fatalf("%s does not read custom templates through the core-owned port", useCasePath)
	}
	for _, forbidden := range []string{
		`"os"`,
		`"os/exec"`,
		`"net/http"`,
		"exec.Command",
		"internal/adapters",
	} {
		if strings.Contains(source, forbidden) {
			t.Fatalf("%s contains forbidden dependency %q", useCasePath, forbidden)
		}
	}
}
