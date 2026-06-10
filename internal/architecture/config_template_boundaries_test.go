package architecture_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigTemplateYAMLParsingStaysOutsideCore(t *testing.T) {
	for _, directory := range []string{
		filepath.Join("..", "core", "domain"),
		filepath.Join("..", "core", "ports"),
		filepath.Join("..", "core", "usecase"),
	} {
		t.Run(filepath.ToSlash(directory), func(t *testing.T) {
			err := filepath.WalkDir(directory, func(path string, entry os.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
					return nil
				}

				contents, err := os.ReadFile(path)
				if err != nil {
					return err
				}
				if strings.Contains(string(contents), "gopkg.in/yaml") {
					t.Fatalf("%s imports YAML parsing; config YAML decoding must stay outside core", path)
				}
				return nil
			})
			if err != nil {
				t.Fatalf("WalkDir(%q) error = %v", directory, err)
			}
		})
	}
}

func TestConfigTemplateGenerationUsesPortsAndNoExecutionAPIs(t *testing.T) {
	useCasePath := filepath.Join("..", "core", "usecase", "generate_change.go")
	contents, err := os.ReadFile(useCasePath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", useCasePath, err)
	}
	source := string(contents)

	for _, want := range []string{
		"ports.ConfigFileSystem",
		"ports.ConfigParser",
		"loadConfigForTemplateGeneration",
		"ConfigTemplateAlias",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("%s missing expected config-template boundary %q", useCasePath, want)
		}
	}
	for _, forbidden := range []string{
		`"os"`,
		`"os/exec"`,
		`"net/http"`,
		"exec.Command",
		"gopkg.in/yaml",
		"internal/adapters",
		"git commit",
		"git push",
	} {
		if strings.Contains(source, forbidden) {
			t.Fatalf("%s contains forbidden dependency or behavior %q", useCasePath, forbidden)
		}
	}
}
