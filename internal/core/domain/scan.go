package domain

// ScanCategory groups a detectable project signal.
type ScanCategory string

const (
	ScanCategoryEcosystem           ScanCategory = "ecosystem"
	ScanCategoryPackageManager      ScanCategory = "package-manager"
	ScanCategoryCI                  ScanCategory = "ci"
	ScanCategoryContainerDeployment ScanCategory = "container-deployment"
	ScanCategorySpecHarbor          ScanCategory = "specharbor"
)

// ScanProbeKind describes how a signal rule is detected.
type ScanProbeKind string

const (
	ScanProbeKindFile       ScanProbeKind = "file"
	ScanProbeKindDirectory  ScanProbeKind = "directory"
	ScanProbeKindFileSuffix ScanProbeKind = "file-suffix"
)

// ScanSignalRule describes one deterministic project signal to detect.
type ScanSignalRule struct {
	Category    ScanCategory
	Name        string
	Path        string
	Kind        ScanProbeKind
	TestCommand string
}

// ScanDetection is a matched signal rendered for reporting.
type ScanDetection struct {
	Name   string
	Signal string
}

// ScanResult is the deterministic, structured outcome of a project scan.
type ScanResult struct {
	ProjectRoot          string
	Ecosystems           []ScanDetection
	PackageManagers      []ScanDetection
	TestCommandHints     []string
	CIProviders          []ScanDetection
	ContainerDeployments []ScanDetection
	SpecHarborSignals    []ScanDetection
	Notes                []string
}

// NewScanResult builds a scan result, defensively copying every slice so callers
// cannot mutate the result through the slices they passed in.
func NewScanResult(
	projectRoot string,
	ecosystems []ScanDetection,
	packageManagers []ScanDetection,
	testCommandHints []string,
	ciProviders []ScanDetection,
	containerDeployments []ScanDetection,
	specHarborSignals []ScanDetection,
	notes []string,
) ScanResult {
	return ScanResult{
		ProjectRoot:          projectRoot,
		Ecosystems:           append([]ScanDetection(nil), ecosystems...),
		PackageManagers:      append([]ScanDetection(nil), packageManagers...),
		TestCommandHints:     append([]string(nil), testCommandHints...),
		CIProviders:          append([]ScanDetection(nil), ciProviders...),
		ContainerDeployments: append([]ScanDetection(nil), containerDeployments...),
		SpecHarborSignals:    append([]ScanDetection(nil), specHarborSignals...),
		Notes:                append([]string(nil), notes...),
	}
}

// ScanSignalRules returns the single deterministic, stack-agnostic catalog of
// presence-detection rules. It only describes top-level or conventional paths
// and is intentionally separate from RequiredOpenSpecChangeFiles, which defines
// change-file policy rather than project presence detection.
func ScanSignalRules() []ScanSignalRule {
	return []ScanSignalRule{
		{Category: ScanCategoryEcosystem, Name: "Go", Path: "go.mod", Kind: ScanProbeKindFile, TestCommand: "go test ./..."},
		{Category: ScanCategoryEcosystem, Name: "Node", Path: "package.json", Kind: ScanProbeKindFile, TestCommand: "npm test"},
		{Category: ScanCategoryEcosystem, Name: "Node", Path: "tsconfig.json", Kind: ScanProbeKindFile},
		{Category: ScanCategoryEcosystem, Name: "Java", Path: "pom.xml", Kind: ScanProbeKindFile, TestCommand: "mvn test"},
		{Category: ScanCategoryEcosystem, Name: "Java", Path: "build.gradle", Kind: ScanProbeKindFile, TestCommand: "gradle test"},
		{Category: ScanCategoryEcosystem, Name: "Java", Path: "build.gradle.kts", Kind: ScanProbeKindFile, TestCommand: "gradle test"},
		{Category: ScanCategoryEcosystem, Name: "Java", Path: "settings.gradle", Kind: ScanProbeKindFile},
		{Category: ScanCategoryEcosystem, Name: "Java", Path: "settings.gradle.kts", Kind: ScanProbeKindFile},
		{Category: ScanCategoryEcosystem, Name: "Python", Path: "pyproject.toml", Kind: ScanProbeKindFile, TestCommand: "pytest"},
		{Category: ScanCategoryEcosystem, Name: "Python", Path: "requirements.txt", Kind: ScanProbeKindFile, TestCommand: "pytest"},
		{Category: ScanCategoryEcosystem, Name: "Python", Path: "Pipfile", Kind: ScanProbeKindFile, TestCommand: "pytest"},
		{Category: ScanCategoryEcosystem, Name: "Python", Path: "poetry.lock", Kind: ScanProbeKindFile},
		{Category: ScanCategoryEcosystem, Name: "Rust", Path: "Cargo.toml", Kind: ScanProbeKindFile, TestCommand: "cargo test"},
		{Category: ScanCategoryEcosystem, Name: ".NET", Path: ".csproj", Kind: ScanProbeKindFileSuffix, TestCommand: "dotnet test"},
		{Category: ScanCategoryEcosystem, Name: ".NET", Path: ".sln", Kind: ScanProbeKindFileSuffix, TestCommand: "dotnet test"},

		{Category: ScanCategoryPackageManager, Name: "npm", Path: "package-lock.json", Kind: ScanProbeKindFile},
		{Category: ScanCategoryPackageManager, Name: "pnpm", Path: "pnpm-lock.yaml", Kind: ScanProbeKindFile},
		{Category: ScanCategoryPackageManager, Name: "yarn", Path: "yarn.lock", Kind: ScanProbeKindFile},

		{Category: ScanCategoryCI, Name: "GitHub Actions", Path: ".github/workflows", Kind: ScanProbeKindDirectory},
		{Category: ScanCategoryCI, Name: "GitLab CI", Path: ".gitlab-ci.yml", Kind: ScanProbeKindFile},
		{Category: ScanCategoryCI, Name: "Jenkins", Path: "Jenkinsfile", Kind: ScanProbeKindFile},

		{Category: ScanCategoryContainerDeployment, Path: "Dockerfile", Kind: ScanProbeKindFile},
		{Category: ScanCategoryContainerDeployment, Path: "docker-compose.yml", Kind: ScanProbeKindFile},
		{Category: ScanCategoryContainerDeployment, Path: "docker-compose.yaml", Kind: ScanProbeKindFile},
		{Category: ScanCategoryContainerDeployment, Path: "compose.yml", Kind: ScanProbeKindFile},
		{Category: ScanCategoryContainerDeployment, Path: "compose.yaml", Kind: ScanProbeKindFile},
		{Category: ScanCategoryContainerDeployment, Path: "kubernetes", Kind: ScanProbeKindDirectory},
		{Category: ScanCategoryContainerDeployment, Path: "k8s", Kind: ScanProbeKindDirectory},
		{Category: ScanCategoryContainerDeployment, Path: "helm", Kind: ScanProbeKindDirectory},

		{Category: ScanCategorySpecHarbor, Path: "openspec/project.md", Kind: ScanProbeKindFile},
		{Category: ScanCategorySpecHarbor, Path: "openspec/changes", Kind: ScanProbeKindDirectory},
		{Category: ScanCategorySpecHarbor, Path: "openspec/specs", Kind: ScanProbeKindDirectory},
		{Category: ScanCategorySpecHarbor, Path: ".specharbor/config.yml", Kind: ScanProbeKindFile},
		{Category: ScanCategorySpecHarbor, Path: ".specharbor/rules", Kind: ScanProbeKindDirectory},
	}
}

// AssembleScanResult is the pure, deterministic assembler that turns the project
// root and the rules matched during scanning into a structured scan result.
// Matched rules are expected in catalog order, which the assembler preserves.
func AssembleScanResult(projectRoot string, matchedRules []ScanSignalRule) ScanResult {
	var ecosystems []ScanDetection
	var packageManagers []ScanDetection
	var ciProviders []ScanDetection
	var containerDeployments []ScanDetection
	var specHarborSignals []ScanDetection
	var testCommandHints []string

	seenHints := make(map[string]bool)
	kubernetesDetected := false

	for _, rule := range matchedRules {
		detection := ScanDetection{Name: rule.Name, Signal: scanSignalDisplay(rule)}

		switch rule.Category {
		case ScanCategoryEcosystem:
			ecosystems = append(ecosystems, detection)
			if rule.TestCommand != "" && !seenHints[rule.TestCommand] {
				seenHints[rule.TestCommand] = true
				testCommandHints = append(testCommandHints, rule.TestCommand)
			}
		case ScanCategoryPackageManager:
			packageManagers = append(packageManagers, detection)
		case ScanCategoryCI:
			ciProviders = append(ciProviders, detection)
		case ScanCategoryContainerDeployment:
			containerDeployments = append(containerDeployments, detection)
			if rule.Kind == ScanProbeKindDirectory && isKubernetesSignalPath(rule.Path) {
				kubernetesDetected = true
			}
		case ScanCategorySpecHarbor:
			specHarborSignals = append(specHarborSignals, detection)
		}
	}

	notes := assembleScanNotes(len(matchedRules), len(containerDeployments) > 0, kubernetesDetected)

	return NewScanResult(
		projectRoot,
		ecosystems,
		packageManagers,
		testCommandHints,
		ciProviders,
		containerDeployments,
		specHarborSignals,
		notes,
	)
}

func scanSignalDisplay(rule ScanSignalRule) string {
	if rule.Kind == ScanProbeKindDirectory {
		return rule.Path + "/"
	}
	return rule.Path
}

func isKubernetesSignalPath(path string) bool {
	return path == "kubernetes" || path == "k8s" || path == "helm"
}

func assembleScanNotes(matchedCount int, hasContainerDeployment bool, kubernetesDetected bool) []string {
	if matchedCount == 0 {
		return []string{"No known project signals detected."}
	}
	if hasContainerDeployment && !kubernetesDetected {
		return []string{"No Kubernetes manifests detected."}
	}
	return nil
}
