package usecase

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/guferreira1/spec-harbor/internal/core/domain"
	"github.com/guferreira1/spec-harbor/internal/core/ports"
)

func TestDiscoverProjectContextDetectsGoProjectContext(t *testing.T) {
	fileSystem := newFakeContextDiscoveryFileSystem()
	fileSystem.files["go.mod"] = "module example.com/project\n"

	result, err := NewDiscoverProjectContext(fileSystem).Execute(DiscoverProjectContextInput{ProjectRoot: "/project"})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	assertContextSignal(t, result, domain.ContextSignalKindStack, "Go", domain.ContextSignalClassificationDetectedFact, "go.mod")
	assertContextSignal(t, result, domain.ContextSignalKindLanguage, "Go", domain.ContextSignalClassificationDetectedFact, "go.mod")
	assertContextSignal(t, result, domain.ContextSignalKindPackageManager, "Go modules", domain.ContextSignalClassificationDetectedFact, "go.mod")
	assertContextSignal(t, result, domain.ContextSignalKindTestCommand, "go test ./...", domain.ContextSignalClassificationSuggestedAssumption, "go.mod")
	assertContextSignal(t, result, domain.ContextSignalKindBuildCommand, "go build ./...", domain.ContextSignalClassificationSuggestedAssumption, "go.mod")
}

func TestDiscoverProjectContextDetectsNodeProjectContext(t *testing.T) {
	fileSystem := newFakeContextDiscoveryFileSystem()
	fileSystem.files["package.json"] = `{
  "packageManager": "pnpm@9.0.0",
  "scripts": {
    "test": "vitest",
    "build": "vite build",
    "dev": "vite --host"
  },
  "dependencies": {
    "react": "latest",
    "express": "latest"
  },
  "devDependencies": {
    "typescript": "latest"
  }
}`

	result, err := NewDiscoverProjectContext(fileSystem).Execute(DiscoverProjectContextInput{ProjectRoot: "/project"})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	assertContextSignal(t, result, domain.ContextSignalKindStack, "Node.js", domain.ContextSignalClassificationDetectedFact, "package.json")
	assertContextSignal(t, result, domain.ContextSignalKindPackageManager, "pnpm", domain.ContextSignalClassificationDetectedFact, "package.json")
	assertContextSignal(t, result, domain.ContextSignalKindTestCommand, "pnpm test", domain.ContextSignalClassificationDetectedFact, "package.json")
	assertContextSignal(t, result, domain.ContextSignalKindBuildCommand, "pnpm run build", domain.ContextSignalClassificationDetectedFact, "package.json")
	assertContextSignal(t, result, domain.ContextSignalKindRunCommand, "pnpm run dev", domain.ContextSignalClassificationDetectedFact, "package.json")
	assertContextSignal(t, result, domain.ContextSignalKindFramework, "React", domain.ContextSignalClassificationDetectedFact, "package.json")
	assertContextSignal(t, result, domain.ContextSignalKindFramework, "Express", domain.ContextSignalClassificationDetectedFact, "package.json")
	assertContextSignal(t, result, domain.ContextSignalKindLanguage, "TypeScript", domain.ContextSignalClassificationDetectedFact, "package.json")
}

func TestDiscoverProjectContextDetectsJavaPythonAndRustContext(t *testing.T) {
	fileSystem := newFakeContextDiscoveryFileSystem()
	fileSystem.files["pom.xml"] = `<project><artifactId>spring-boot-starter-web</artifactId></project>`
	fileSystem.files["build.gradle"] = `plugins { id "org.springframework.boot" version "3.0.0" }`
	fileSystem.files["pyproject.toml"] = "[tool.poetry]\nname = \"service\"\ndependencies = [\"fastapi\"]\n"
	fileSystem.files["requirements.txt"] = "Flask\n"
	fileSystem.files["Cargo.toml"] = "[package]\nname = \"cli\"\n"

	result, err := NewDiscoverProjectContext(fileSystem).Execute(DiscoverProjectContextInput{ProjectRoot: "/project"})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	assertContextSignal(t, result, domain.ContextSignalKindPackageManager, "Maven", domain.ContextSignalClassificationDetectedFact, "pom.xml")
	assertContextSignal(t, result, domain.ContextSignalKindPackageManager, "Gradle", domain.ContextSignalClassificationDetectedFact, "build.gradle")
	assertContextSignal(t, result, domain.ContextSignalKindFramework, "Spring Boot", domain.ContextSignalClassificationDetectedFact, "pom.xml")
	assertContextSignal(t, result, domain.ContextSignalKindStack, "Python", domain.ContextSignalClassificationDetectedFact, "pyproject.toml")
	assertContextSignal(t, result, domain.ContextSignalKindPackageManager, "Poetry", domain.ContextSignalClassificationDetectedFact, "pyproject.toml")
	assertContextSignal(t, result, domain.ContextSignalKindFramework, "FastAPI", domain.ContextSignalClassificationDetectedFact, "pyproject.toml")
	assertContextSignal(t, result, domain.ContextSignalKindFramework, "Flask", domain.ContextSignalClassificationDetectedFact, "requirements.txt")
	assertContextSignal(t, result, domain.ContextSignalKindStack, "Rust", domain.ContextSignalClassificationDetectedFact, "Cargo.toml")
	assertContextSignal(t, result, domain.ContextSignalKindPackageManager, "Cargo", domain.ContextSignalClassificationDetectedFact, "Cargo.toml")
	assertNoteContains(t, result, "Multiple stack signals detected")
}

func TestDiscoverProjectContextDetectsContainersTaskRunnersWorkflowsAndDocs(t *testing.T) {
	fileSystem := newFakeContextDiscoveryFileSystem()
	fileSystem.files["Dockerfile"] = "FROM scratch\n"
	fileSystem.files["docker-compose.yml"] = "services:\n  app:\n    build: .\n"
	fileSystem.files["Makefile"] = "test:\n\tgo test ./...\nbuild:\n\tgo build ./...\nrun:\n\tgo run ./cmd/app\n"
	fileSystem.files["Taskfile.yml"] = "version: '3'\ntasks:\n  test:\n    cmds: [go test ./...]\n  build:\n    cmds: [go build ./...]\n  run:\n    cmds: [go run ./cmd/app]\n"
	fileSystem.files["README.md"] = "# Project\n\nA CLI for project work.\n"
	fileSystem.files["CONTRIBUTING.md"] = "# Contributing\n"
	fileSystem.directories["docs"] = []ports.ContextDiscoveryDirectoryEntry{{Name: "usage.md", IsRegular: true}}
	fileSystem.files["docs/usage.md"] = "# Usage\n"
	fileSystem.directories[".github/workflows"] = []ports.ContextDiscoveryDirectoryEntry{{Name: "ci.yml", IsRegular: true}}
	fileSystem.files[".github/workflows/ci.yml"] = "steps:\n  - run: go test ./...\n"

	result, err := NewDiscoverProjectContext(fileSystem).Execute(DiscoverProjectContextInput{ProjectRoot: "/project"})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	assertContextSignal(t, result, domain.ContextSignalKindContainerSignal, "Dockerfile", domain.ContextSignalClassificationDetectedFact, "Dockerfile")
	assertContextSignal(t, result, domain.ContextSignalKindContainerSignal, "Docker Compose", domain.ContextSignalClassificationDetectedFact, "docker-compose.yml")
	assertContextSignal(t, result, domain.ContextSignalKindTestCommand, "make test", domain.ContextSignalClassificationDetectedFact, "Makefile")
	assertContextSignal(t, result, domain.ContextSignalKindBuildCommand, "make build", domain.ContextSignalClassificationDetectedFact, "Makefile")
	assertContextSignal(t, result, domain.ContextSignalKindRunCommand, "make run", domain.ContextSignalClassificationDetectedFact, "Makefile")
	assertContextSignal(t, result, domain.ContextSignalKindDocumentationSource, "README.md", domain.ContextSignalClassificationDetectedFact, "README.md")
	assertContextSignal(t, result, domain.ContextSignalKindDocumentationSource, "CONTRIBUTING.md", domain.ContextSignalClassificationDetectedFact, "CONTRIBUTING.md")
	assertContextSignal(t, result, domain.ContextSignalKindDocumentationSource, "docs/usage.md", domain.ContextSignalClassificationDetectedFact, "docs/usage.md")
	assertContextSignal(t, result, domain.ContextSignalKindWorkflowSignal, "GitHub Actions", domain.ContextSignalClassificationDetectedFact, ".github/workflows")
	assertContextSignal(t, result, domain.ContextSignalKindWorkflowSignal, ".github/workflows/ci.yml", domain.ContextSignalClassificationDetectedFact, ".github/workflows/ci.yml")
}

func TestDiscoverProjectContextDetectsReadmeMarkdownCommands(t *testing.T) {
	fileSystem := newFakeContextDiscoveryFileSystem()
	fileSystem.files["README.md"] = "# Project\n\nSpecHarbor is a CLI for OpenSpec workflows.\n\n```bash\ngo test ./...\ngo build ./...\ngo run ./cmd/fixture\n```\n"

	result, err := NewDiscoverProjectContext(fileSystem).Execute(DiscoverProjectContextInput{ProjectRoot: "/project"})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	assertContextSignal(t, result, domain.ContextSignalKindTestCommand, "go test ./...", domain.ContextSignalClassificationDetectedFact, "README.md")
	assertContextSignal(t, result, domain.ContextSignalKindBuildCommand, "go build ./...", domain.ContextSignalClassificationDetectedFact, "README.md")
	assertContextSignal(t, result, domain.ContextSignalKindRunCommand, "go run ./cmd/fixture", domain.ContextSignalClassificationDetectedFact, "README.md")
}

func TestDiscoverProjectContextDetectsContributingMarkdownCommands(t *testing.T) {
	fileSystem := newFakeContextDiscoveryFileSystem()
	fileSystem.files["CONTRIBUTING.md"] = "# Contributing\n\nBefore opening a change, run:\n\n- `npm run test`\n- `npm run build`\n- `npm run dev`\n"

	result, err := NewDiscoverProjectContext(fileSystem).Execute(DiscoverProjectContextInput{ProjectRoot: "/project"})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	assertContextSignal(t, result, domain.ContextSignalKindTestCommand, "npm run test", domain.ContextSignalClassificationDetectedFact, "CONTRIBUTING.md")
	assertContextSignal(t, result, domain.ContextSignalKindBuildCommand, "npm run build", domain.ContextSignalClassificationDetectedFact, "CONTRIBUTING.md")
	assertContextSignal(t, result, domain.ContextSignalKindRunCommand, "npm run dev", domain.ContextSignalClassificationDetectedFact, "CONTRIBUTING.md")
}

func TestDiscoverProjectContextDoesNotDetectPurposeFromContributingSections(t *testing.T) {
	tests := []struct {
		name     string
		contents string
	}{
		{
			name:     "overview section",
			contents: "# Contributing\n\n## Overview\n\nThis guide explains how contributors should prepare changes.\n",
		},
		{
			name:     "purpose section",
			contents: "# Contributing\n\n## Purpose\n\nThis guide explains how contributors should prepare changes.\n",
		},
		{
			name:     "about section",
			contents: "# Contributing\n\n## About\n\nThis guide explains how contributors should prepare changes.\n",
		},
		{
			name:     "description section",
			contents: "# Contributing\n\n## Description\n\nThis guide explains how contributors should prepare changes.\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fileSystem := newFakeContextDiscoveryFileSystem()
			fileSystem.files["CONTRIBUTING.md"] = test.contents

			result, err := NewDiscoverProjectContext(fileSystem).Execute(DiscoverProjectContextInput{ProjectRoot: "/project"})
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			assertContextSignal(t, result, domain.ContextSignalKindDocumentationSource, "CONTRIBUTING.md", domain.ContextSignalClassificationDetectedFact, "CONTRIBUTING.md")
			assertNoContextSignalKind(t, result, domain.ContextSignalKindPurposeSummary)
		})
	}
}

func TestDiscoverProjectContextDoesNotUseGenericMarkdownProseAsPurpose(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		contents string
	}{
		{
			name:     "contributing procedure",
			path:     "CONTRIBUTING.md",
			contents: "# Contributing\n\nBefore opening a change, run:\n",
		},
		{
			name:     "readme procedure",
			path:     "README.md",
			contents: "# Project\n\nRun the tests before submitting.\n",
		},
		{
			name:     "readme command introduction",
			path:     "README.md",
			contents: "# Project\n\nUse this command to start the service.\n",
		},
		{
			name:     "readme ordered procedure",
			path:     "README.md",
			contents: "# Project\n\n## Overview\n\n1. Run the tests before submitting.\n",
		},
		{
			name:     "readme guide procedure",
			path:     "README.md",
			contents: "# Project\n\n## Overview\n\nThis guide explains how contributors should prepare changes.\n",
		},
		{
			name:     "unknown vague prose",
			path:     "README.md",
			contents: "# Notes\n\nUseful notes for contributors live here.\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fileSystem := newFakeContextDiscoveryFileSystem()
			fileSystem.files[test.path] = test.contents

			result, err := NewDiscoverProjectContext(fileSystem).Execute(DiscoverProjectContextInput{ProjectRoot: "/project"})
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			assertNoContextSignalKind(t, result, domain.ContextSignalKindPurposeSummary)
		})
	}
}

func TestDiscoverProjectContextDetectsExplicitReadmePurposeEvidence(t *testing.T) {
	tests := []struct {
		name     string
		contents string
		want     string
	}{
		{
			name:     "project name sentence",
			contents: "# SpecHarbor\n\nSpecHarbor is an open source CLI for OpenSpec workflows.\n",
			want:     "SpecHarbor is an open source CLI for OpenSpec workflows.",
		},
		{
			name:     "project title description",
			contents: "# SpecHarbor\n\nA Go CLI for OpenSpec workflows.\n",
			want:     "A Go CLI for OpenSpec workflows.",
		},
		{
			name:     "purpose heading",
			contents: "# Project\n\n## Purpose\n\nSpecHarbor helps teams prepare OpenSpec changes.\n",
			want:     "SpecHarbor helps teams prepare OpenSpec changes.",
		},
		{
			name:     "overview heading",
			contents: "# Project\n\n## Overview\n\nA local CLI for OpenSpec workflows.\n",
			want:     "A local CLI for OpenSpec workflows.",
		},
		{
			name:     "description label",
			contents: "# Project\n\nDescription: SpecHarbor is a local CLI for OpenSpec workflows.\n",
			want:     "SpecHarbor is a local CLI for OpenSpec workflows.",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fileSystem := newFakeContextDiscoveryFileSystem()
			fileSystem.files["README.md"] = test.contents

			result, err := NewDiscoverProjectContext(fileSystem).Execute(DiscoverProjectContextInput{ProjectRoot: "/project"})
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			assertContextSignal(t, result, domain.ContextSignalKindPurposeSummary, test.want, domain.ContextSignalClassificationDetectedFact, "README.md")
		})
	}
}

func TestDiscoverProjectContextDetectsDocsMarkdownCommandsAndArchitecture(t *testing.T) {
	fileSystem := newFakeContextDiscoveryFileSystem()
	fileSystem.directories["docs"] = []ports.ContextDiscoveryDirectoryEntry{{Name: "architecture.md", IsRegular: true}}
	fileSystem.files["docs/architecture.md"] = "# Architecture\n\nHexagonal Architecture\n\n```bash\ngo test ./...\n```\n\nPrivate deployment note that must not be copied into context.\n"

	result, err := NewDiscoverProjectContext(fileSystem).Execute(DiscoverProjectContextInput{ProjectRoot: "/project"})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	assertContextSignal(t, result, domain.ContextSignalKindArchitectureHint, "Hexagonal Architecture", domain.ContextSignalClassificationDetectedFact, "docs/architecture.md")
	assertContextSignal(t, result, domain.ContextSignalKindTestCommand, "go test ./...", domain.ContextSignalClassificationDetectedFact, "docs/architecture.md")
	assertNoContextSignalText(t, result, "Private deployment note")
}

func TestDiscoverProjectContextDoesNotUseVagueMarkdownAsArchitecture(t *testing.T) {
	tests := []struct {
		name            string
		contents        string
		unwantedValue   string
		requireNoSignal bool
	}{
		{
			name:          "solid test coverage",
			contents:      "# Project\n\nsolid test coverage keeps releases safe.\n",
			unwantedValue: "SOLID",
		},
		{
			name:          "docs are solid",
			contents:      "# Project\n\nThe docs are solid.\n",
			unwantedValue: "SOLID",
		},
		{
			name:          "mvc inside another word",
			contents:      "# Project\n\nThe amvc architecture note is a typo.\n",
			unwantedValue: "MVC",
		},
		{
			name:            "vague markdown prose",
			contents:        "# Project\n\nThis section describes project organization.\n",
			requireNoSignal: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fileSystem := newFakeContextDiscoveryFileSystem()
			fileSystem.files["README.md"] = test.contents

			result, err := NewDiscoverProjectContext(fileSystem).Execute(DiscoverProjectContextInput{ProjectRoot: "/project"})
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			if test.requireNoSignal {
				assertNoContextSignalKind(t, result, domain.ContextSignalKindArchitectureHint)
				return
			}
			assertNoContextSignal(t, result, domain.ContextSignalKindArchitectureHint, test.unwantedValue, domain.ContextSignalClassificationDetectedFact)
		})
	}
}

func TestDiscoverProjectContextDetectsExplicitMarkdownArchitectureEvidence(t *testing.T) {
	tests := []struct {
		name     string
		contents string
		want     string
	}{
		{
			name:     "mvc architecture",
			contents: "# Project\n\nMVC architecture\n",
			want:     "MVC",
		},
		{
			name:     "hexagonal architecture",
			contents: "# Project\n\nThe project follows Hexagonal Architecture.\n",
			want:     "Hexagonal Architecture",
		},
		{
			name:     "clean architecture",
			contents: "# Project\n\nArchitecture: Clean Architecture\n",
			want:     "Clean Architecture",
		},
		{
			name:     "domain driven design",
			contents: "# Project\n\nDomain-Driven Design\n",
			want:     "Domain-Driven Design",
		},
		{
			name:     "ddd architecture",
			contents: "# Project\n\nDDD architecture\n",
			want:     "Domain-Driven Design",
		},
		{
			name:     "layered architecture",
			contents: "# Project\n\nWe use layered architecture for service boundaries.\n",
			want:     "Layered Architecture",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fileSystem := newFakeContextDiscoveryFileSystem()
			fileSystem.files["README.md"] = test.contents

			result, err := NewDiscoverProjectContext(fileSystem).Execute(DiscoverProjectContextInput{ProjectRoot: "/project"})
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			assertContextSignal(t, result, domain.ContextSignalKindArchitectureHint, test.want, domain.ContextSignalClassificationDetectedFact, "README.md")
		})
	}
}

func TestDiscoverProjectContextMarkdownOutputOrderingIsDeterministic(t *testing.T) {
	fileSystem := newFakeContextDiscoveryFileSystem()
	fileSystem.directories["docs"] = []ports.ContextDiscoveryDirectoryEntry{
		{Name: "b.md", IsRegular: true},
		{Name: "a.md", IsRegular: true},
	}
	fileSystem.files["docs/a.md"] = "# A\n\nClean Architecture\n"
	fileSystem.files["docs/b.md"] = "# B\n\nHexagonal Architecture\n"

	first, err := NewDiscoverProjectContext(fileSystem).Execute(DiscoverProjectContextInput{ProjectRoot: "/project"})
	if err != nil {
		t.Fatalf("first Execute() error = %v", err)
	}
	second, err := NewDiscoverProjectContext(fileSystem).Execute(DiscoverProjectContextInput{ProjectRoot: "/project"})
	if err != nil {
		t.Fatalf("second Execute() error = %v", err)
	}
	if !reflect.DeepEqual(first.Signals(), second.Signals()) {
		t.Fatalf("signals are not deterministic:\nfirst=%+v\nsecond=%+v", first.Signals(), second.Signals())
	}

	var docSources []string
	for _, signal := range first.Signals() {
		if signal.Kind == domain.ContextSignalKindDocumentationSource && strings.HasPrefix(signal.Source.Path, "docs/") {
			docSources = append(docSources, signal.Source.Path)
		}
	}
	wantDocSources := []string{"docs/a.md", "docs/b.md"}
	if !reflect.DeepEqual(docSources, wantDocSources) {
		t.Fatalf("documentation source order = %v, want %v", docSources, wantDocSources)
	}
}

func TestDiscoverProjectContextDetectsAgentAndOpenSpecSources(t *testing.T) {
	fileSystem := newFakeContextDiscoveryFileSystem()
	fileSystem.files["AGENTS.md"] = "# AGENTS\n\nUse Hexagonal Architecture.\n\ngo test ./...\n"
	fileSystem.files["openspec/project.md"] = "# Project\n\n## Purpose\n\nBuild a CLI for OpenSpec workflows.\n\n## Technical stack\n\n- Language: Go\n- Interface: CLI\n- CI: GitHub Actions\n"
	fileSystem.directories[".specharbor/rules"] = []ports.ContextDiscoveryDirectoryEntry{{Name: "global.md", IsRegular: true}}
	fileSystem.files[".specharbor/rules/global.md"] = "# Global\n"
	fileSystem.directories["openspec/specs"] = []ports.ContextDiscoveryDirectoryEntry{{Name: "architecture", IsDirectory: true}}
	fileSystem.directories["openspec/specs/architecture"] = []ports.ContextDiscoveryDirectoryEntry{{Name: "spec.md", IsRegular: true}}
	fileSystem.files["openspec/specs/architecture/spec.md"] = "# Architecture\n\nThe project follows Clean Architecture and SOLID.\n"

	result, err := NewDiscoverProjectContext(fileSystem).Execute(DiscoverProjectContextInput{ProjectRoot: "/project"})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	assertContextSignal(t, result, domain.ContextSignalKindAgentInstructionSource, "AGENTS.md", domain.ContextSignalClassificationDetectedFact, "AGENTS.md")
	assertContextSignal(t, result, domain.ContextSignalKindAgentInstructionSource, ".specharbor/rules/", domain.ContextSignalClassificationDetectedFact, ".specharbor/rules")
	assertContextSignal(t, result, domain.ContextSignalKindOpenSpecSource, "openspec/project.md", domain.ContextSignalClassificationDetectedFact, "openspec/project.md")
	assertContextSignal(t, result, domain.ContextSignalKindOpenSpecSource, "openspec/specs/architecture/spec.md", domain.ContextSignalClassificationDetectedFact, "openspec/specs/architecture/spec.md")
	assertContextSignal(t, result, domain.ContextSignalKindPurposeSummary, "Build a CLI for OpenSpec workflows.", domain.ContextSignalClassificationDetectedFact, "openspec/project.md")
	assertContextSignal(t, result, domain.ContextSignalKindArchitectureHint, "Hexagonal Architecture", domain.ContextSignalClassificationDetectedFact, "AGENTS.md")
	assertContextSignal(t, result, domain.ContextSignalKindArchitectureHint, "Clean Architecture", domain.ContextSignalClassificationDetectedFact, "openspec/specs/architecture/spec.md")
	assertContextSignal(t, result, domain.ContextSignalKindWorkflowSignal, "GitHub Actions", domain.ContextSignalClassificationDetectedFact, "openspec/project.md")
}

func TestDiscoverProjectContextParsesProjectBriefAsConfirmedContext(t *testing.T) {
	fileSystem := newFakeContextDiscoveryFileSystem()
	fileSystem.files[".specharbor/project-brief.md"] = `# Project Brief

## Project type

Answer: CLI/tooling project

## Unknown Section

Answer: Must not become context

## Stack

Answer: Go

## Commands

### Test

Answer: go test ./...

### Build

Answer: go build ./...

## Agent behavior

Answer: Ask before assuming
`
	fileSystem.files["go.mod"] = "module example.com/project\n"

	result, err := NewDiscoverProjectContext(fileSystem).Execute(DiscoverProjectContextInput{ProjectRoot: "/project"})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	assertContextSignal(t, result, domain.ContextSignalKindProjectType, "CLI/tooling project", domain.ContextSignalClassificationUserConfirmedContext, ".specharbor/project-brief.md")
	assertContextSignal(t, result, domain.ContextSignalKindStack, "Go", domain.ContextSignalClassificationUserConfirmedContext, ".specharbor/project-brief.md")
	assertContextSignal(t, result, domain.ContextSignalKindTestCommand, "go test ./...", domain.ContextSignalClassificationUserConfirmedContext, ".specharbor/project-brief.md")
	assertContextSignal(t, result, domain.ContextSignalKindBuildCommand, "go build ./...", domain.ContextSignalClassificationUserConfirmedContext, ".specharbor/project-brief.md")
	assertContextSignal(t, result, domain.ContextSignalKindAgentInstructionSource, "Ask before assuming", domain.ContextSignalClassificationUserConfirmedContext, ".specharbor/project-brief.md")
	assertContextSignal(t, result, domain.ContextSignalKindStack, "Go", domain.ContextSignalClassificationDetectedFact, "go.mod")
	assertNoContextSignalValue(t, result, "Must not become context")
	assertNoContextSignal(t, result, domain.ContextSignalKindTestCommand, "go test ./...", domain.ContextSignalClassificationSuggestedAssumption)
}

func TestDiscoverProjectContextReportsProjectBriefExistenceForPromptContext(t *testing.T) {
	fileSystem := newFakeContextDiscoveryFileSystem()
	fileSystem.files[".specharbor/project-brief.md"] = "# Project Brief\n\n## Unknown\n\nAnswer: ignored\n"
	useCase := NewDiscoverProjectContext(fileSystem)

	exists, err := useCase.ProjectBriefExists("/project")
	if err != nil {
		t.Fatalf("ProjectBriefExists() error = %v", err)
	}
	if !exists {
		t.Fatalf("ProjectBriefExists() = false, want true")
	}
}

func TestDiscoverProjectContextSkipsSecretsGeneratedFoldersAndSymlinks(t *testing.T) {
	fileSystem := newFakeContextDiscoveryFileSystem()
	fileSystem.directories["docs"] = []ports.ContextDiscoveryDirectoryEntry{
		{Name: "usage.md", IsRegular: true},
		{Name: "secrets.md", IsRegular: true},
		{Name: "node_modules", IsDirectory: true},
		{Name: "linked.md", IsSymlink: true},
	}
	fileSystem.directories["docs/node_modules"] = []ports.ContextDiscoveryDirectoryEntry{{Name: "README.md", IsRegular: true}}
	fileSystem.files["docs/usage.md"] = "# Usage\n"
	fileSystem.files["docs/secrets.md"] = "# Secret\n"
	fileSystem.files["docs/node_modules/README.md"] = "# Generated\n"
	fileSystem.files["docs/linked.md"] = "# Linked\n"

	result, err := NewDiscoverProjectContext(fileSystem).Execute(DiscoverProjectContextInput{ProjectRoot: "/project"})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	assertContextSignal(t, result, domain.ContextSignalKindDocumentationSource, "docs/usage.md", domain.ContextSignalClassificationDetectedFact, "docs/usage.md")
	assertNoContextSignalValue(t, result, "docs/secrets.md")
	assertNoContextSignalValue(t, result, "docs/node_modules/README.md")
	assertNoContextSignalValue(t, result, "docs/linked.md")
	for _, readPath := range fileSystem.reads {
		if strings.Contains(readPath, "secrets") || strings.Contains(readPath, "node_modules") || strings.Contains(readPath, "linked") {
			t.Fatalf("read skipped path %q; reads=%v", readPath, fileSystem.reads)
		}
	}
}

func TestDiscoverProjectContextReturnsMissingContextNote(t *testing.T) {
	result, err := NewDiscoverProjectContext(newFakeContextDiscoveryFileSystem()).Execute(DiscoverProjectContextInput{ProjectRoot: "/project"})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if len(result.Signals()) != 0 {
		t.Fatalf("Signals() = %+v, want none", result.Signals())
	}
	assertNoteContains(t, result, "No supported context sources detected.")
}

func TestDiscoverProjectContextRejectsInvalidDependenciesAndInput(t *testing.T) {
	_, err := (*DiscoverProjectContext)(nil).Execute(DiscoverProjectContextInput{ProjectRoot: "/project"})
	if err == nil || err.Error() != "discover project context use case is required" {
		t.Fatalf("nil use case error = %v, want use case required", err)
	}

	_, err = NewDiscoverProjectContext(nil).Execute(DiscoverProjectContextInput{ProjectRoot: "/project"})
	if err == nil || err.Error() != "context discovery filesystem is required" {
		t.Fatalf("nil filesystem error = %v, want filesystem required", err)
	}

	_, err = NewDiscoverProjectContext(newFakeContextDiscoveryFileSystem()).Execute(DiscoverProjectContextInput{ProjectRoot: " "})
	if err == nil || err.Error() != "project root is required" {
		t.Fatalf("empty root error = %v, want project root required", err)
	}
}

func TestDiscoverProjectContextReturnsFilesystemErrors(t *testing.T) {
	wantErr := errors.New("permission denied")
	fileSystem := newFakeContextDiscoveryFileSystem()
	fileSystem.fileErrors["go.mod"] = wantErr

	_, err := NewDiscoverProjectContext(fileSystem).Execute(DiscoverProjectContextInput{ProjectRoot: "/project"})
	if err == nil || !errors.Is(err, wantErr) {
		t.Fatalf("Execute() error = %v, want wrapped filesystem error", err)
	}
	if !strings.Contains(err.Error(), "check file go.mod") {
		t.Fatalf("Execute() error = %q, want file context", err.Error())
	}
}

func TestDiscoverProjectContextUsesBoundedProjectBriefReads(t *testing.T) {
	fileSystem := newFakeContextDiscoveryFileSystem()
	fileSystem.files[".specharbor/project-brief.md"] = strings.Repeat("x", int(contextDiscoveryBriefReadLimitBytes)+1)

	result, err := NewDiscoverProjectContext(fileSystem).Execute(DiscoverProjectContextInput{ProjectRoot: "/project"})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	assertNoteContains(t, result, "Skipped .specharbor/project-brief.md")
	if fileSystem.readLimits[".specharbor/project-brief.md"] != contextDiscoveryBriefReadLimitBytes {
		t.Fatalf("project brief read limit = %d, want %d", fileSystem.readLimits[".specharbor/project-brief.md"], contextDiscoveryBriefReadLimitBytes)
	}
}

func assertContextSignal(
	t *testing.T,
	result domain.ContextDiscoveryResult,
	kind domain.ContextSignalKind,
	value string,
	classification domain.ContextSignalClassification,
	sourcePath string,
) {
	t.Helper()

	for _, signal := range result.Signals() {
		if signal.Kind == kind &&
			signal.Value == value &&
			signal.Classification == classification &&
			signal.Source.Path == sourcePath &&
			signal.Confidence != "" {
			return
		}
	}
	t.Fatalf("missing context signal kind=%q value=%q classification=%q source=%q in %+v", kind, value, classification, sourcePath, result.Signals())
}

func assertNoContextSignal(
	t *testing.T,
	result domain.ContextDiscoveryResult,
	kind domain.ContextSignalKind,
	value string,
	classification domain.ContextSignalClassification,
) {
	t.Helper()

	for _, signal := range result.Signals() {
		if signal.Kind == kind && signal.Value == value && signal.Classification == classification {
			t.Fatalf("unexpected context signal %+v", signal)
		}
	}
}

func assertNoContextSignalKind(
	t *testing.T,
	result domain.ContextDiscoveryResult,
	kind domain.ContextSignalKind,
) {
	t.Helper()

	for _, signal := range result.Signals() {
		if signal.Kind == kind {
			t.Fatalf("unexpected context signal kind=%q in %+v", kind, signal)
		}
	}
}

func assertNoContextSignalValue(t *testing.T, result domain.ContextDiscoveryResult, value string) {
	t.Helper()

	for _, signal := range result.Signals() {
		if signal.Value == value || signal.Source.Path == value {
			t.Fatalf("unexpected context signal value/source %q in %+v", value, signal)
		}
	}
}

func assertNoContextSignalText(t *testing.T, result domain.ContextDiscoveryResult, text string) {
	t.Helper()

	for _, signal := range result.Signals() {
		if strings.Contains(signal.Value, text) ||
			strings.Contains(signal.Source.Path, text) ||
			strings.Contains(signal.Source.Evidence, text) {
			t.Fatalf("unexpected context text %q in %+v", text, signal)
		}
	}
}

func assertNoteContains(t *testing.T, result domain.ContextDiscoveryResult, want string) {
	t.Helper()

	for _, note := range result.Notes() {
		if strings.Contains(note.Message, want) {
			return
		}
	}
	t.Fatalf("notes = %+v, want note containing %q", result.Notes(), want)
}

type fakeContextDiscoveryFileSystem struct {
	files           map[string]string
	directories     map[string][]ports.ContextDiscoveryDirectoryEntry
	fileErrors      map[string]error
	directoryErrors map[string]error
	listErrors      map[string]error
	readErrors      map[string]error
	reads           []string
	readLimits      map[string]int64
}

func newFakeContextDiscoveryFileSystem() *fakeContextDiscoveryFileSystem {
	return &fakeContextDiscoveryFileSystem{
		files:           make(map[string]string),
		directories:     make(map[string][]ports.ContextDiscoveryDirectoryEntry),
		fileErrors:      make(map[string]error),
		directoryErrors: make(map[string]error),
		listErrors:      make(map[string]error),
		readErrors:      make(map[string]error),
		readLimits:      make(map[string]int64),
	}
}

func (fileSystem *fakeContextDiscoveryFileSystem) FileExists(_ string, relativePath string) (bool, error) {
	if err := fileSystem.fileErrors[relativePath]; err != nil {
		return false, err
	}
	_, exists := fileSystem.files[relativePath]
	return exists, nil
}

func (fileSystem *fakeContextDiscoveryFileSystem) DirectoryExists(_ string, relativePath string) (bool, error) {
	if err := fileSystem.directoryErrors[relativePath]; err != nil {
		return false, err
	}
	_, exists := fileSystem.directories[relativePath]
	return exists, nil
}

func (fileSystem *fakeContextDiscoveryFileSystem) ListDirectory(_ string, relativePath string) ([]ports.ContextDiscoveryDirectoryEntry, error) {
	if err := fileSystem.listErrors[relativePath]; err != nil {
		return nil, err
	}
	return append([]ports.ContextDiscoveryDirectoryEntry(nil), fileSystem.directories[relativePath]...), nil
}

func (fileSystem *fakeContextDiscoveryFileSystem) ReadFile(_ string, relativePath string, maxBytes int64) (string, error) {
	fileSystem.reads = append(fileSystem.reads, relativePath)
	fileSystem.readLimits[relativePath] = maxBytes
	if err := fileSystem.readErrors[relativePath]; err != nil {
		return "", err
	}
	contents := fileSystem.files[relativePath]
	if int64(len(contents)) > maxBytes {
		return "", errors.New("context discovery file exceeds limit")
	}
	return contents, nil
}
