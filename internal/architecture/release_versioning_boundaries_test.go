package architecture_test

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestVersionPlatformPackageAvoidsRuntimeDiscoveryDependencies(t *testing.T) {
	versionDirectory := filepath.Join("..", "platform", "version")

	entries, err := os.ReadDir(versionDirectory)
	if err != nil {
		t.Fatalf("ReadDir(%q) error = %v", versionDirectory, err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}

		sourcePath := filepath.Join(versionDirectory, entry.Name())
		t.Run(filepath.ToSlash(sourcePath), func(t *testing.T) {
			parsedFile, err := parser.ParseFile(token.NewFileSet(), sourcePath, nil, parser.ImportsOnly)
			if err != nil {
				t.Fatalf("ParseFile(%q) error = %v", sourcePath, err)
			}

			for _, imported := range parsedFile.Imports {
				importPath := strings.Trim(imported.Path.Value, `"`)
				if forbiddenVersionRuntimeImport(importPath) {
					t.Fatalf("%s imports forbidden version runtime dependency %q", sourcePath, importPath)
				}
			}

			source := mustReadArchitectureFile(t, sourcePath)
			for _, forbidden := range []string{
				".git",
				"git tag",
				"git describe",
				"exec.Command",
				"http.Get",
				"http.Post",
				"net.Dial",
				"TrimPrefix",
				"TrimSuffix",
			} {
				if strings.Contains(source, forbidden) {
					t.Fatalf("%s contains forbidden version runtime behavior %q", sourcePath, forbidden)
				}
			}
		})
	}
}

func TestVersionCommandDoesNotWriteFilesOrUseRuntimeDiscovery(t *testing.T) {
	sourcePath := filepath.Join("..", "adapters", "cli", "cli.go")
	source := mustReadArchitectureFile(t, sourcePath)
	versionCommand := sourceBetween(t, source, "func versionCommand", "\nfunc parseVersionArguments")

	for _, forbidden := range []string{
		"os.",
		"WriteFile",
		"Mkdir",
		"Create",
		"Remove",
		"Rename",
		"exec.Command",
		"http.Get",
		"http.Post",
		".git",
		"git ",
	} {
		if strings.Contains(versionCommand, forbidden) {
			t.Fatalf("versionCommand contains forbidden behavior %q", forbidden)
		}
	}
}

func TestGoReleaserConfigBuildsApprovedArchivesAndChecksums(t *testing.T) {
	root := filepath.Join("..", "..")
	configPath := filepath.Join(root, ".goreleaser.yaml")

	assertArchitecturePathExists(t, configPath)
	assertArchitecturePathDoesNotExist(t, filepath.Join(root, ".goreleaser.yml"))

	config := mustReadYAMLMap(t, configPath)
	assertExactYAMLKeys(t, "GoReleaser top-level config", config, []string{
		"version",
		"project_name",
		"dist",
		"builds",
		"archives",
		"checksum",
	})
	assertYAMLValue(t, "version", config["version"], 2)
	assertYAMLValue(t, "project_name", config["project_name"], "specharbor")
	assertYAMLValue(t, "dist", config["dist"], "dist")

	builds := mustYAMLList(t, config["builds"], "builds")
	if len(builds) != 1 {
		t.Fatalf("builds length = %d, want 1", len(builds))
	}
	build := mustYAMLMap(t, builds[0], "builds[0]")
	assertExactYAMLKeys(t, "builds[0]", build, []string{
		"id",
		"main",
		"binary",
		"env",
		"goos",
		"goarch",
		"ldflags",
	})
	assertYAMLValue(t, "builds[0].id", build["id"], "specharbor")
	assertYAMLValue(t, "builds[0].main", build["main"], "./cmd/specharbor")
	assertYAMLValue(t, "builds[0].binary", build["binary"], "specharbor")
	assertStringList(t, "builds[0].env", mustYAMLStringList(t, build["env"], "builds[0].env"), []string{"CGO_ENABLED=0"})
	assertStringList(t, "builds[0].goos", mustYAMLStringList(t, build["goos"], "builds[0].goos"), []string{"linux", "darwin", "windows"})
	assertStringList(t, "builds[0].goarch", mustYAMLStringList(t, build["goarch"], "builds[0].goarch"), []string{"amd64", "arm64"})

	archives := mustYAMLList(t, config["archives"], "archives")
	if len(archives) != 1 {
		t.Fatalf("archives length = %d, want 1", len(archives))
	}
	archive := mustYAMLMap(t, archives[0], "archives[0]")
	assertExactYAMLKeys(t, "archives[0]", archive, []string{
		"id",
		"ids",
		"formats",
		"format_overrides",
		"name_template",
	})
	assertYAMLValue(t, "archives[0].id", archive["id"], "specharbor")
	assertStringList(t, "archives[0].ids", mustYAMLStringList(t, archive["ids"], "archives[0].ids"), []string{"specharbor"})
	assertStringList(t, "archives[0].formats", mustYAMLStringList(t, archive["formats"], "archives[0].formats"), []string{"tar.gz"})

	formatOverrides := mustYAMLList(t, archive["format_overrides"], "archives[0].format_overrides")
	if len(formatOverrides) != 1 {
		t.Fatalf("archives[0].format_overrides length = %d, want 1", len(formatOverrides))
	}
	windowsOverride := mustYAMLMap(t, formatOverrides[0], "archives[0].format_overrides[0]")
	assertExactYAMLKeys(t, "archives[0].format_overrides[0]", windowsOverride, []string{"goos", "formats"})
	assertYAMLValue(t, "archives[0].format_overrides[0].goos", windowsOverride["goos"], "windows")
	assertStringList(t, "archives[0].format_overrides[0].formats", mustYAMLStringList(t, windowsOverride["formats"], "archives[0].format_overrides[0].formats"), []string{"zip"})

	nameTemplate := mustYAMLString(t, archive["name_template"], "archives[0].name_template")
	for _, want := range []string{
		"{{ .ProjectName }}",
		"{{ title .Os }}",
		`{{ if eq .Arch "amd64" }}x86_64{{ else }}{{ .Arch }}{{ end }}`,
	} {
		if !strings.Contains(nameTemplate, want) {
			t.Fatalf("archives[0].name_template = %q, want snippet %q", nameTemplate, want)
		}
	}

	checksum := mustYAMLMap(t, config["checksum"], "checksum")
	assertExactYAMLKeys(t, "checksum", checksum, []string{"name_template", "algorithm"})
	assertYAMLValue(t, "checksum.name_template", checksum["name_template"], "checksums.txt")
	assertYAMLValue(t, "checksum.algorithm", checksum["algorithm"], "sha256")
}

func TestGoReleaserLdflagsInjectApprovedVersionMetadata(t *testing.T) {
	root := filepath.Join("..", "..")
	config := mustReadYAMLMap(t, filepath.Join(root, ".goreleaser.yaml"))
	build := mustYAMLMap(t, mustYAMLList(t, config["builds"], "builds")[0], "builds[0]")
	ldflags := strings.Join(mustYAMLStringList(t, build["ldflags"], "builds[0].ldflags"), " ")

	versionPackage := "github.com/guferreira1/spec-harbor/internal/platform/version"
	requiredLdflags := map[string]string{
		versionPackage + ".Version": "{{ .Version }}",
		versionPackage + ".Commit":  "{{ .FullCommit }}",
		versionPackage + ".Date":    "{{ .Date }}",
		versionPackage + ".Dirty":   "{{ .IsGitDirty }}",
	}
	for variable, template := range requiredLdflags {
		snippet := "-X " + variable + "=" + template
		if !strings.Contains(ldflags, snippet) {
			t.Fatalf("ldflags missing %q in %q", snippet, ldflags)
		}
	}
	if count := strings.Count(ldflags, "-X "); count != len(requiredLdflags) {
		t.Fatalf("ldflags contain %d -X entries, want %d: %q", count, len(requiredLdflags), ldflags)
	}
	if count := strings.Count(ldflags, versionPackage+"."); count != len(requiredLdflags) {
		t.Fatalf("ldflags target %d version package variables, want %d: %q", count, len(requiredLdflags), ldflags)
	}
	for _, forbidden := range []string{
		".Tag",
		".ShortCommit",
		"TrimPrefix",
		"trimPrefix",
		"replace",
		"Version=v",
	} {
		if strings.Contains(ldflags, forbidden) {
			t.Fatalf("ldflags contain forbidden version injection template %q: %q", forbidden, ldflags)
		}
	}
}

func TestGoReleaserConfigRejectsOutOfScopePublishingSections(t *testing.T) {
	root := filepath.Join("..", "..")
	configPath := filepath.Join(root, ".goreleaser.yaml")
	config := mustReadYAMLMap(t, configPath)
	source := strings.ToLower(mustReadArchitectureFile(t, configPath))

	for _, section := range []string{
		"npm",
		"nfpms",
		"nfpm",
		"brews",
		"homebrew_casks",
		"scoops",
		"winget",
		"aurs",
		"nix",
		"dockers",
		"docker_manifests",
		"sboms",
		"signs",
		"cosign",
		"kos",
		"announce",
		"publishers",
		"blobs",
		"snapcrafts",
		"chocolateys",
		"krew",
		"notarize",
		"attestations",
	} {
		if _, ok := config[section]; ok {
			t.Fatalf(".goreleaser.yaml contains out-of-scope section %q", section)
		}
	}
	for _, forbidden := range []string{
		"goreleaser-pro",
		"cosign",
		"sbom",
		"docker",
		"homebrew",
		"npm",
		"nfpm",
		"scoop",
		"winget",
		"aur",
		"sign",
		"notar",
		"attestation",
		"token",
		"secret",
		"password",
	} {
		if strings.Contains(source, forbidden) {
			t.Fatalf(".goreleaser.yaml contains out-of-scope token %q", forbidden)
		}
	}
}

func TestReleaseWorkflowTriggersOnlyOnVersionTagsWithMinimalPermissions(t *testing.T) {
	root := filepath.Join("..", "..")
	workflowPath := filepath.Join(root, ".github", "workflows", "release.yml")

	assertArchitecturePathExists(t, workflowPath)
	assertArchitecturePathDoesNotExist(t, filepath.Join(root, ".github", "workflows", "release.yaml"))

	workflow := mustReadYAMLMap(t, workflowPath)
	assertExactYAMLKeys(t, "release workflow top-level", workflow, []string{"name", "on", "permissions", "jobs"})

	events := mustYAMLMap(t, workflow["on"], "on")
	assertExactYAMLKeys(t, "release workflow triggers", events, []string{"push"})
	push := mustYAMLMap(t, events["push"], "on.push")
	assertExactYAMLKeys(t, "release workflow push trigger", push, []string{"tags"})
	assertStringList(t, "on.push.tags", mustYAMLStringList(t, push["tags"], "on.push.tags"), []string{"v*"})

	permissions := mustYAMLMap(t, workflow["permissions"], "permissions")
	assertExactYAMLKeys(t, "release workflow permissions", permissions, []string{"contents"})
	assertYAMLValue(t, "permissions.contents", permissions["contents"], "write")
}

func TestReleaseWorkflowUsesCommunityGoReleaserAndOnlyGitHubToken(t *testing.T) {
	root := filepath.Join("..", "..")
	workflowPath := filepath.Join(root, ".github", "workflows", "release.yml")
	source := mustReadArchitectureFile(t, workflowPath)

	for _, want := range []string{
		"actions/checkout@v4",
		"fetch-depth: 0",
		"actions/setup-go@v5",
		"go-version-file: go.mod",
		"go test ./...",
		"goreleaser/goreleaser-action@v6",
		"distribution: goreleaser",
		"version: ~> v2",
		"args: release --clean",
		"GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("release workflow missing %q", want)
		}
	}
	if count := strings.Count(source, "secrets."); count != 1 {
		t.Fatalf("release workflow references %d secrets, want exactly 1", count)
	}
	if count := strings.Count(source, "\npermissions:"); count != 1 {
		t.Fatalf("release workflow defines %d top-level permissions blocks, want 1", count)
	}

	lowerSource := strings.ToLower(source)
	for _, forbidden := range []string{
		"pull_request",
		"branches:",
		"schedule:",
		"workflow_dispatch:",
		"write-all",
		"packages:",
		"id-token:",
		"pull-requests:",
		"issues:",
		"deployments:",
		"security-events:",
		"administration:",
		"goreleaser-pro",
		"npm",
		"homebrew",
		"brew",
		"package registry",
		"docker",
		"cosign",
		"signing",
		"sbom",
		"git tag",
		"git push",
		"gh pr",
		"gh repo",
	} {
		if strings.Contains(lowerSource, forbidden) {
			t.Fatalf("release workflow contains out-of-scope behavior %q", forbidden)
		}
	}
}

func TestReleaseScopeDoesNotAddOutOfScopePackageArtifacts(t *testing.T) {
	root := filepath.Join("..", "..")

	for _, relativePath := range []string{
		".goreleaser.yml",
		".github/workflows/release.yaml",
		"install.sh",
		"publish.sh",
		"release.sh",
		"scripts/install.sh",
		"scripts/publish",
		"scripts/publish.sh",
		"scripts/release",
		"scripts/release.sh",
		"package.json",
		"package-lock.json",
		"npm",
		"packages/npm",
		"Formula",
		"homebrew",
		"nfpm.yaml",
		".nfpm.yaml",
		"packaging",
		"debian",
		"rpm",
		"winget",
		"scoop",
		"chocolatey",
		"Dockerfile.release",
		"docker",
		"sbom",
		"cosign",
	} {
		assertArchitecturePathDoesNotExist(t, filepath.Join(root, filepath.FromSlash(relativePath)))
	}

	workflowsDirectory := filepath.Join(root, ".github", "workflows")
	workflowEntries, err := os.ReadDir(workflowsDirectory)
	if err != nil && !os.IsNotExist(err) {
		t.Fatalf("ReadDir(%q) error = %v", workflowsDirectory, err)
	}
	for _, entry := range workflowEntries {
		if entry.Name() == "release.yml" {
			continue
		}
		name := strings.ToLower(entry.Name())
		if strings.Contains(name, "release") {
			t.Fatalf("release-specific workflow file exists: %s", filepath.Join(workflowsDirectory, entry.Name()))
		}

		contents := mustReadArchitectureFile(t, filepath.Join(workflowsDirectory, entry.Name()))
		for _, forbidden := range []string{
			"goreleaser",
			"softprops/action-gh-release",
			"gh release",
			"npm publish",
			"brew tap",
			"docker push",
			"cosign",
		} {
			if strings.Contains(strings.ToLower(contents), forbidden) {
				t.Fatalf("%s contains release publishing behavior %q", entry.Name(), forbidden)
			}
		}
	}

	assertNoReleaseArtifacts(t, root)
}

func TestReleaseVersioningDocumentationDescribesImplementedScopeOnly(t *testing.T) {
	requiredSnippetsByDocument := map[string][]string{
		"README.md": {
			"specharbor version",
			"SpecHarbor dev",
			"`v0.1.0`",
			"`0.1.0`",
			"GoReleaser",
			"GitHub Release assets",
			"`checksums.txt`",
			"npm",
			"Homebrew",
			"future work",
		},
		"docs/usage.md": {
			"specharbor version",
			"SpecHarbor dev",
			"github.com/guferreira1/spec-harbor/internal/platform/version",
			"GoReleaser",
			"displays the injected version string as-is",
			"does not normalize",
			"`v0.1.0`",
			"`0.1.0`",
		},
		"docs/release.md": {
			"GoReleaser",
			"`vX.Y.Z`",
			"`X.Y.Z`",
			"`specharbor_Linux_x86_64.tar.gz`",
			"`specharbor_Linux_arm64.tar.gz`",
			"`specharbor_Darwin_x86_64.tar.gz`",
			"`specharbor_Darwin_arm64.tar.gz`",
			"`specharbor_Windows_x86_64.zip`",
			"`specharbor_Windows_arm64.zip`",
			"`checksums.txt`",
			"`goreleaser check`",
			"`goreleaser release --snapshot --clean`",
			"npm",
			"Homebrew",
			"install scripts",
			"SBOM",
			"Docker",
			"future work",
		},
	}

	for name, requiredSnippets := range requiredSnippetsByDocument {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join("..", "..", filepath.FromSlash(name))
			source := mustReadArchitectureFile(t, path)
			for _, snippet := range requiredSnippets {
				if !strings.Contains(source, snippet) {
					t.Fatalf("%s missing release versioning documentation snippet %q", name, snippet)
				}
			}
		})
	}
}

func forbiddenVersionRuntimeImport(importPath string) bool {
	for _, exact := range []string{
		"io/fs",
		"net",
		"net/http",
		"os",
		"os/exec",
		"path",
		"path/filepath",
	} {
		if importPath == exact {
			return true
		}
	}

	for _, prefix := range []string{
		"github.com/go-git/",
		"github.com/google/go-github",
		"github.com/xanzy/go-gitlab",
		"gopkg.in/src-d/go-git",
	} {
		if strings.HasPrefix(importPath, prefix) {
			return true
		}
	}

	return false
}

func assertArchitecturePathDoesNotExist(t *testing.T, path string) {
	t.Helper()

	_, err := os.Stat(path)
	if err == nil {
		t.Fatalf("out-of-scope release/package artifact exists: %s", path)
	}
	if !os.IsNotExist(err) {
		t.Fatalf("Stat(%q) error = %v", path, err)
	}
}

func assertArchitecturePathExists(t *testing.T, path string) {
	t.Helper()

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected path to exist: %s: %v", path, err)
	}
}

func assertNoReleaseArtifacts(t *testing.T, root string) {
	t.Helper()

	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if entry.Name() == ".git" {
				return filepath.SkipDir
			}
			if entry.Name() == "dist" {
				return filepath.SkipDir
			}
			return nil
		}

		lowerName := strings.ToLower(entry.Name())
		for _, suffix := range []string{".tar.gz", ".tgz", ".zip", ".sha256", ".sha512"} {
			if strings.HasSuffix(lowerName, suffix) {
				t.Fatalf("generated release artifact exists: %s", path)
			}
		}
		for _, name := range []string{"checksums.txt", "sha256sums", "sha256sums.txt"} {
			if lowerName == name {
				t.Fatalf("generated checksum artifact exists: %s", path)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("WalkDir(%q) error = %v", root, err)
	}
}

func mustReadYAMLMap(t *testing.T, path string) map[string]any {
	t.Helper()

	source := []byte(mustReadArchitectureFile(t, path))
	var parsed map[string]any
	if err := yaml.Unmarshal(source, &parsed); err != nil {
		t.Fatalf("Unmarshal(%q) error = %v", path, err)
	}
	return parsed
}

func mustYAMLMap(t *testing.T, value any, context string) map[string]any {
	t.Helper()

	parsed, ok := value.(map[string]any)
	if !ok {
		t.Fatalf("%s = %#v, want YAML map", context, value)
	}
	return parsed
}

func mustYAMLList(t *testing.T, value any, context string) []any {
	t.Helper()

	parsed, ok := value.([]any)
	if !ok {
		t.Fatalf("%s = %#v, want YAML list", context, value)
	}
	return parsed
}

func mustYAMLString(t *testing.T, value any, context string) string {
	t.Helper()

	parsed, ok := value.(string)
	if !ok {
		t.Fatalf("%s = %#v, want string", context, value)
	}
	return parsed
}

func mustYAMLStringList(t *testing.T, value any, context string) []string {
	t.Helper()

	list := mustYAMLList(t, value, context)
	stringsList := make([]string, 0, len(list))
	for index, item := range list {
		stringItem, ok := item.(string)
		if !ok {
			t.Fatalf("%s[%d] = %#v, want string", context, index, item)
		}
		stringsList = append(stringsList, stringItem)
	}
	return stringsList
}

func assertYAMLValue(t *testing.T, context string, got any, want any) {
	t.Helper()

	if got != want {
		t.Fatalf("%s = %#v, want %#v", context, got, want)
	}
}

func assertStringList(t *testing.T, context string, got []string, want []string) {
	t.Helper()

	if strings.Join(got, "\x00") != strings.Join(want, "\x00") {
		t.Fatalf("%s = %#v, want %#v", context, got, want)
	}
}

func assertExactYAMLKeys(t *testing.T, context string, values map[string]any, want []string) {
	t.Helper()

	got := make([]string, 0, len(values))
	for key := range values {
		got = append(got, key)
	}
	sort.Strings(got)

	sortedWant := append([]string(nil), want...)
	sort.Strings(sortedWant)

	if strings.Join(got, "\x00") != strings.Join(sortedWant, "\x00") {
		t.Fatalf("%s keys = %#v, want %#v", context, got, sortedWant)
	}
}

func mustReadArchitectureFile(t *testing.T, path string) string {
	t.Helper()

	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}
	return string(contents)
}

func sourceBetween(t *testing.T, source string, start string, end string) string {
	t.Helper()

	startIndex := strings.Index(source, start)
	if startIndex < 0 {
		t.Fatalf("source missing start marker %q", start)
	}
	remaining := source[startIndex:]
	endIndex := strings.Index(remaining, end)
	if endIndex < 0 {
		t.Fatalf("source missing end marker %q after %q", end, start)
	}
	return remaining[:endIndex]
}
