package domain

import (
	"strings"
	"testing"
)

func TestScanSignalRulesCoversEveryCategoryDeterministically(t *testing.T) {
	rules := ScanSignalRules()
	if len(rules) == 0 {
		t.Fatalf("ScanSignalRules() is empty, want a populated catalog")
	}

	categories := map[ScanCategory]int{}
	for _, rule := range rules {
		if rule.Path == "" {
			t.Fatalf("rule %+v has empty path", rule)
		}
		switch rule.Kind {
		case ScanProbeKindFile, ScanProbeKindDirectory, ScanProbeKindFileSuffix:
		default:
			t.Fatalf("rule %+v has unknown probe kind", rule)
		}
		categories[rule.Category]++
	}

	for _, category := range []ScanCategory{
		ScanCategoryEcosystem,
		ScanCategoryPackageManager,
		ScanCategoryCI,
		ScanCategoryContainerDeployment,
		ScanCategorySpecHarbor,
	} {
		if categories[category] == 0 {
			t.Fatalf("catalog has no rules for category %q", category)
		}
	}

	if !rulesAreStable(rules, ScanSignalRules()) {
		t.Fatalf("ScanSignalRules() is not deterministic between calls")
	}
}

func TestScanSignalRulesDoesNotReuseRequiredOpenSpecChangeFiles(t *testing.T) {
	required := map[string]bool{}
	for _, file := range RequiredOpenSpecChangeFiles() {
		required[file] = true
	}

	for _, rule := range ScanSignalRules() {
		if required[rule.Path] {
			t.Fatalf("scan catalog reuses required-change-file %q as a detection path", rule.Path)
		}
	}
}

func TestAssembleScanResultGroupsMatchedRulesInCatalogOrder(t *testing.T) {
	result := AssembleScanResult("/project", []ScanSignalRule{
		{Category: ScanCategoryEcosystem, Name: "Go", Path: "go.mod", Kind: ScanProbeKindFile, TestCommand: "go test ./..."},
		{Category: ScanCategoryEcosystem, Name: "Node", Path: "package.json", Kind: ScanProbeKindFile, TestCommand: "npm test"},
		{Category: ScanCategoryPackageManager, Name: "npm", Path: "package-lock.json", Kind: ScanProbeKindFile},
		{Category: ScanCategoryCI, Name: "GitHub Actions", Path: ".github/workflows", Kind: ScanProbeKindDirectory},
		{Category: ScanCategoryContainerDeployment, Path: "Dockerfile", Kind: ScanProbeKindFile},
		{Category: ScanCategorySpecHarbor, Path: "openspec/project.md", Kind: ScanProbeKindFile},
	})

	if result.ProjectRoot != "/project" {
		t.Fatalf("ProjectRoot = %q, want /project", result.ProjectRoot)
	}
	assertDetectionSignals(t, "ecosystems", result.Ecosystems, []string{"go.mod", "package.json"})
	assertDetectionSignals(t, "package managers", result.PackageManagers, []string{"package-lock.json"})
	assertDetectionSignals(t, "ci", result.CIProviders, []string{".github/workflows/"})
	assertDetectionSignals(t, "containers", result.ContainerDeployments, []string{"Dockerfile"})
	assertDetectionSignals(t, "specharbor", result.SpecHarborSignals, []string{"openspec/project.md"})
}

func TestAssembleScanResultBuildsSignalDisplayPerProbeKind(t *testing.T) {
	result := AssembleScanResult("/project", []ScanSignalRule{
		{Category: ScanCategoryEcosystem, Name: "Go", Path: "go.mod", Kind: ScanProbeKindFile},
		{Category: ScanCategoryEcosystem, Name: ".NET", Path: ".csproj", Kind: ScanProbeKindFileSuffix},
		{Category: ScanCategoryCI, Name: "GitHub Actions", Path: ".github/workflows", Kind: ScanProbeKindDirectory},
	})

	if result.Ecosystems[0].Signal != "go.mod" {
		t.Fatalf("file signal = %q, want go.mod", result.Ecosystems[0].Signal)
	}
	if result.Ecosystems[1].Signal != ".csproj" {
		t.Fatalf("file-suffix signal = %q, want .csproj", result.Ecosystems[1].Signal)
	}
	if result.CIProviders[0].Signal != ".github/workflows/" {
		t.Fatalf("directory signal = %q, want .github/workflows/", result.CIProviders[0].Signal)
	}
}

func TestAssembleScanResultUsesNamesOnlyForNamedCategories(t *testing.T) {
	result := AssembleScanResult("/project", []ScanSignalRule{
		{Category: ScanCategoryEcosystem, Name: "Go", Path: "go.mod", Kind: ScanProbeKindFile},
		{Category: ScanCategoryPackageManager, Name: "npm", Path: "package-lock.json", Kind: ScanProbeKindFile},
		{Category: ScanCategoryCI, Name: "GitLab CI", Path: ".gitlab-ci.yml", Kind: ScanProbeKindFile},
		{Category: ScanCategoryContainerDeployment, Path: "Dockerfile", Kind: ScanProbeKindFile},
		{Category: ScanCategorySpecHarbor, Path: "openspec/project.md", Kind: ScanProbeKindFile},
	})

	for _, detection := range append(append([]ScanDetection{}, result.Ecosystems...), append(result.PackageManagers, result.CIProviders...)...) {
		if detection.Name == "" {
			t.Fatalf("named-category detection %+v has empty name", detection)
		}
	}
	for _, detection := range append(append([]ScanDetection{}, result.ContainerDeployments...), result.SpecHarborSignals...) {
		if detection.Name != "" {
			t.Fatalf("path-only detection %+v has non-empty name %q", detection, detection.Name)
		}
	}
}

func TestAssembleScanResultDeduplicatesTestCommandHintsInFirstSeenOrder(t *testing.T) {
	result := AssembleScanResult("/project", []ScanSignalRule{
		{Category: ScanCategoryEcosystem, Name: "Go", Path: "go.mod", Kind: ScanProbeKindFile, TestCommand: "go test ./..."},
		{Category: ScanCategoryEcosystem, Name: "Node", Path: "package.json", Kind: ScanProbeKindFile, TestCommand: "npm test"},
		{Category: ScanCategoryEcosystem, Name: "Java", Path: "build.gradle", Kind: ScanProbeKindFile, TestCommand: "gradle test"},
		{Category: ScanCategoryEcosystem, Name: "Java", Path: "build.gradle.kts", Kind: ScanProbeKindFile, TestCommand: "gradle test"},
		{Category: ScanCategoryEcosystem, Name: "Node", Path: "tsconfig.json", Kind: ScanProbeKindFile},
	})

	want := []string{"go test ./...", "npm test", "gradle test"}
	if strings.Join(result.TestCommandHints, "|") != strings.Join(want, "|") {
		t.Fatalf("TestCommandHints = %v, want %v", result.TestCommandHints, want)
	}
}

func TestAssembleScanResultEmitsNoSignalsNote(t *testing.T) {
	result := AssembleScanResult("/project", nil)

	if len(result.Notes) != 1 || result.Notes[0] != "No known project signals detected." {
		t.Fatalf("Notes = %v, want single no-signals note", result.Notes)
	}
}

func TestAssembleScanResultEmitsKubernetesNoteWhenContainersLackKubernetes(t *testing.T) {
	result := AssembleScanResult("/project", []ScanSignalRule{
		{Category: ScanCategoryContainerDeployment, Path: "Dockerfile", Kind: ScanProbeKindFile},
	})

	if len(result.Notes) != 1 || result.Notes[0] != "No Kubernetes manifests detected." {
		t.Fatalf("Notes = %v, want single Kubernetes note", result.Notes)
	}
}

func TestAssembleScanResultSuppressesKubernetesNoteWhenKubernetesPresent(t *testing.T) {
	for _, path := range []string{"kubernetes", "k8s", "helm"} {
		t.Run(path, func(t *testing.T) {
			result := AssembleScanResult("/project", []ScanSignalRule{
				{Category: ScanCategoryContainerDeployment, Path: "Dockerfile", Kind: ScanProbeKindFile},
				{Category: ScanCategoryContainerDeployment, Path: path, Kind: ScanProbeKindDirectory},
			})

			if len(result.Notes) != 0 {
				t.Fatalf("Notes = %v, want no notes when %q directory is present", result.Notes, path)
			}
		})
	}
}

func TestAssembleScanResultEmitsNoNoteWhenSignalsDetectedWithoutContainers(t *testing.T) {
	result := AssembleScanResult("/project", []ScanSignalRule{
		{Category: ScanCategoryEcosystem, Name: "Go", Path: "go.mod", Kind: ScanProbeKindFile, TestCommand: "go test ./..."},
	})

	if len(result.Notes) != 0 {
		t.Fatalf("Notes = %v, want no notes", result.Notes)
	}
}

func TestNewScanResultCopiesSlices(t *testing.T) {
	ecosystems := []ScanDetection{{Name: "Go", Signal: "go.mod"}}
	packageManagers := []ScanDetection{{Name: "npm", Signal: "package-lock.json"}}
	hints := []string{"go test ./..."}
	ci := []ScanDetection{{Name: "GitLab CI", Signal: ".gitlab-ci.yml"}}
	containers := []ScanDetection{{Signal: "Dockerfile"}}
	specharbor := []ScanDetection{{Signal: "openspec/project.md"}}
	notes := []string{"No Kubernetes manifests detected."}

	result := NewScanResult("/project", ecosystems, packageManagers, hints, ci, containers, specharbor, notes)

	ecosystems[0].Signal = "mutated"
	packageManagers[0].Signal = "mutated"
	hints[0] = "mutated"
	ci[0].Signal = "mutated"
	containers[0].Signal = "mutated"
	specharbor[0].Signal = "mutated"
	notes[0] = "mutated"

	if result.Ecosystems[0].Signal != "go.mod" {
		t.Fatalf("Ecosystems mutated through input slice: %q", result.Ecosystems[0].Signal)
	}
	if result.PackageManagers[0].Signal != "package-lock.json" {
		t.Fatalf("PackageManagers mutated through input slice: %q", result.PackageManagers[0].Signal)
	}
	if result.TestCommandHints[0] != "go test ./..." {
		t.Fatalf("TestCommandHints mutated through input slice: %q", result.TestCommandHints[0])
	}
	if result.CIProviders[0].Signal != ".gitlab-ci.yml" {
		t.Fatalf("CIProviders mutated through input slice: %q", result.CIProviders[0].Signal)
	}
	if result.ContainerDeployments[0].Signal != "Dockerfile" {
		t.Fatalf("ContainerDeployments mutated through input slice: %q", result.ContainerDeployments[0].Signal)
	}
	if result.SpecHarborSignals[0].Signal != "openspec/project.md" {
		t.Fatalf("SpecHarborSignals mutated through input slice: %q", result.SpecHarborSignals[0].Signal)
	}
	if result.Notes[0] != "No Kubernetes manifests detected." {
		t.Fatalf("Notes mutated through input slice: %q", result.Notes[0])
	}
}

func rulesAreStable(first []ScanSignalRule, second []ScanSignalRule) bool {
	if len(first) != len(second) {
		return false
	}
	for index := range first {
		if first[index] != second[index] {
			return false
		}
	}
	return true
}

func assertDetectionSignals(t *testing.T, label string, detections []ScanDetection, want []string) {
	t.Helper()

	if len(detections) != len(want) {
		t.Fatalf("%s count = %d, want %d", label, len(detections), len(want))
	}
	for index, signal := range want {
		if detections[index].Signal != signal {
			t.Fatalf("%s[%d].Signal = %q, want %q", label, index, detections[index].Signal, signal)
		}
	}
}
