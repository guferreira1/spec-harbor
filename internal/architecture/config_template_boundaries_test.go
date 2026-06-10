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
		"ports.RemoteTemplateFetcher",
		"ports.RemoteTemplateBundleReader",
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
		`"archive/zip"`,
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

func TestRemoteTemplateAdaptersOwnHTTPAndZipImplementations(t *testing.T) {
	portPath := filepath.Join("..", "core", "ports", "generation.go")
	portContents, err := os.ReadFile(portPath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", portPath, err)
	}
	portSource := string(portContents)
	for _, want := range []string{
		"type RemoteTemplateFetcher interface",
		"type RemoteTemplateBundleReader interface",
		"domain.RemoteTemplateFetchRequest",
		"domain.RemoteTemplateBundle",
	} {
		if !strings.Contains(portSource, want) {
			t.Fatalf("%s missing remote template port boundary %q", portPath, want)
		}
	}
	for _, forbidden := range []string{
		`"net/http"`,
		`"archive/zip"`,
		"http.Client",
		"zip.Reader",
	} {
		if strings.Contains(portSource, forbidden) {
			t.Fatalf("%s exposes adapter implementation detail %q", portPath, forbidden)
		}
	}

	httpAdapterPath := filepath.Join("..", "adapters", "remote", "http_fetcher.go")
	httpAdapterContents, err := os.ReadFile(httpAdapterPath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", httpAdapterPath, err)
	}
	if !strings.Contains(string(httpAdapterContents), `"net/http"`) ||
		!strings.Contains(string(httpAdapterContents), "CheckRedirect") ||
		!strings.Contains(string(httpAdapterContents), "http.MethodGet") {
		t.Fatalf("%s missing expected adapter-owned HTTP behavior", httpAdapterPath)
	}

	zipAdapterPath := filepath.Join("..", "adapters", "remote", "zip_reader.go")
	zipAdapterContents, err := os.ReadFile(zipAdapterPath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", zipAdapterPath, err)
	}
	if !strings.Contains(string(zipAdapterContents), `"archive/zip"`) ||
		!strings.Contains(string(zipAdapterContents), "ValidateRemoteTemplateArchiveEntry") {
		t.Fatalf("%s missing expected adapter-owned ZIP behavior", zipAdapterPath)
	}
}
