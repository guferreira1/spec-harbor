package domain

import (
	"strings"
	"testing"
)

func TestDefaultProjectBriefQuestionsAreDeterministicAndSupportCustomAnswers(t *testing.T) {
	questions := DefaultProjectBriefQuestions()

	wantOrder := []ProjectBriefQuestionID{
		ProjectBriefQuestionProjectType,
		ProjectBriefQuestionPurpose,
		ProjectBriefQuestionTargetUsers,
		ProjectBriefQuestionStack,
		ProjectBriefQuestionArchitecture,
		ProjectBriefQuestionInstall,
		ProjectBriefQuestionTest,
		ProjectBriefQuestionBuild,
		ProjectBriefQuestionRun,
		ProjectBriefQuestionAgentBehavior,
	}
	if len(questions) != len(wantOrder) {
		t.Fatalf("question count = %d, want %d", len(questions), len(wantOrder))
	}

	for index, question := range questions {
		if question.ID != wantOrder[index] {
			t.Fatalf("question %d id = %q, want %q", index, question.ID, wantOrder[index])
		}
		if err := question.Validate(); err != nil {
			t.Fatalf("question %s validation error = %v", question.ID, err)
		}
		if len(question.Options) < 3 || len(question.Options) > 5 {
			t.Fatalf("question %s option count = %d, want 3..5", question.ID, len(question.Options))
		}
		last := question.Options[len(question.Options)-1]
		if last.Label != ProjectBriefCustomOptionLabel {
			t.Fatalf("question %s last option = %q, want %q", question.ID, last.Label, ProjectBriefCustomOptionLabel)
		}
	}
}

func TestNewProjectBriefSeparatesConfirmedAnswersDetectedContextAndAssumptions(t *testing.T) {
	detected, err := NewDetectedProjectBriefContextSource("Ecosystem", "go.mod")
	if err != nil {
		t.Fatalf("NewDetectedProjectBriefContextSource() error = %v", err)
	}
	assumption, err := NewProjectBriefAssumption("Tests should run before implementation review.")
	if err != nil {
		t.Fatalf("NewProjectBriefAssumption() error = %v", err)
	}

	brief, err := NewProjectBrief(sampleProjectBriefAnswers(t), []ProjectBriefContextSource{detected}, []ProjectBriefAssumption{assumption})
	if err != nil {
		t.Fatalf("NewProjectBrief() error = %v", err)
	}

	if brief.ProjectType.Source != ProjectBriefAnswerSourceUserProvided {
		t.Fatalf("project type source = %q, want user provided", brief.ProjectType.Source)
	}
	if got := brief.ContextSources(); len(got) != 1 || got[0].Source != ProjectBriefAnswerSourceDetectedContext {
		t.Fatalf("context sources = %+v, want detected context", got)
	}
	if got := brief.Assumptions(); len(got) != 1 || got[0].Source != ProjectBriefAnswerSourceAssumption {
		t.Fatalf("assumptions = %+v, want assumption source", got)
	}
}

func TestNewProjectBriefRejectsUnconfirmedAnswerSources(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*ProjectBriefAnswers)
		want   string
	}{
		{
			name: "detected context stack",
			mutate: func(answers *ProjectBriefAnswers) {
				answers.Stack = ProjectBriefAnswer{Value: "Go", Source: ProjectBriefAnswerSourceDetectedContext}
			},
			want: "stack must be a user-provided answer",
		},
		{
			name: "assumption architecture",
			mutate: func(answers *ProjectBriefAnswers) {
				answers.Architecture = ProjectBriefAnswer{Value: "Clean Architecture / Hexagonal", Source: ProjectBriefAnswerSourceAssumption}
			},
			want: "architecture must be a user-provided answer",
		},
		{
			name: "missing test command",
			mutate: func(answers *ProjectBriefAnswers) {
				answers.Commands.Test = ProjectBriefAnswer{Source: ProjectBriefAnswerSourceUserProvided}
			},
			want: "test command is required",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			answers := sampleProjectBriefAnswers(t)
			test.mutate(&answers)

			_, err := NewProjectBrief(answers, nil, nil)
			if err == nil || err.Error() != test.want {
				t.Fatalf("NewProjectBrief() error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestProjectBriefMarkdownRenderingIsDeterministic(t *testing.T) {
	brief, err := NewProjectBrief(sampleProjectBriefAnswers(t), nil, nil)
	if err != nil {
		t.Fatalf("NewProjectBrief() error = %v", err)
	}

	first := brief.RenderMarkdown()
	second := brief.RenderMarkdown()
	if first != second {
		t.Fatalf("RenderMarkdown() changed between calls:\nfirst:\n%s\nsecond:\n%s", first, second)
	}

	want := `# Project Brief

## Project type

Answer: CLI/tooling project

Source: user-provided answer

## Purpose

Answer: Developer productivity tool

Source: user-provided answer

## Target users

Answer: Developers or platform engineers

Source: user-provided answer

## Stack

Answer: Go

Source: user-provided answer

## Architecture

Answer: Clean Architecture / Hexagonal

Source: user-provided answer

## Commands

### Install

Answer: go mod download

Source: user-provided answer

### Test

Answer: go test ./...

Source: user-provided answer

### Build

Answer: go build ./...

Source: user-provided answer

### Run

Answer: go run ./cmd/specharbor

Source: user-provided answer

## Agent behavior

Answer: Ask before assuming

Source: user-provided answer

## Context sources

### User-provided answers

- Project type: CLI/tooling project (Source: user-provided answer)
- Purpose: Developer productivity tool (Source: user-provided answer)
- Target users: Developers or platform engineers (Source: user-provided answer)
- Stack: Go (Source: user-provided answer)
- Architecture: Clean Architecture / Hexagonal (Source: user-provided answer)
- Install command: go mod download (Source: user-provided answer)
- Test command: go test ./... (Source: user-provided answer)
- Build command: go build ./... (Source: user-provided answer)
- Run command: go run ./cmd/specharbor (Source: user-provided answer)
- Agent behavior: Ask before assuming (Source: user-provided answer)

### Detected context

None recorded.

## Assumptions

None recorded.
`
	if first != want {
		t.Fatalf("RenderMarkdown() =\n%s\nwant:\n%s", first, want)
	}
}

func TestProjectBriefMarkdownSeparatesDetectedContextAndAssumptions(t *testing.T) {
	detected, err := NewDetectedProjectBriefContextSource("Ecosystem", "go.mod")
	if err != nil {
		t.Fatalf("NewDetectedProjectBriefContextSource() error = %v", err)
	}
	assumption, err := NewProjectBriefAssumption("Build command should be checked manually before release.")
	if err != nil {
		t.Fatalf("NewProjectBriefAssumption() error = %v", err)
	}
	brief, err := NewProjectBrief(sampleProjectBriefAnswers(t), []ProjectBriefContextSource{detected}, []ProjectBriefAssumption{assumption})
	if err != nil {
		t.Fatalf("NewProjectBrief() error = %v", err)
	}

	rendered := brief.RenderMarkdown()
	for _, want := range []string{
		"## Context sources",
		"### User-provided answers",
		"### Detected context",
		"- Ecosystem: go.mod (Source: detected context)",
		"## Assumptions",
		"- Assumption: Build command should be checked manually before release. (Source: assumption)",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("RenderMarkdown() = %q, want %q", rendered, want)
		}
	}

	beforeAssumptions := strings.Split(rendered, "## Assumptions")[0]
	if strings.Contains(beforeAssumptions, "Source: assumption") {
		t.Fatalf("assumption source rendered before assumptions section:\n%s", rendered)
	}
	if strings.Contains(beforeAssumptions, "Build command should be checked manually before release.") {
		t.Fatalf("assumption rendered as a confirmed fact:\n%s", rendered)
	}
}

func sampleProjectBriefAnswers(t *testing.T) ProjectBriefAnswers {
	t.Helper()

	answer := func(value string) ProjectBriefAnswer {
		t.Helper()
		briefAnswer, err := NewUserProvidedProjectBriefAnswer(value)
		if err != nil {
			t.Fatalf("NewUserProvidedProjectBriefAnswer(%q) error = %v", value, err)
		}
		return briefAnswer
	}

	return ProjectBriefAnswers{
		ProjectType:  answer("CLI/tooling project"),
		Purpose:      answer("Developer productivity tool"),
		TargetUsers:  answer("Developers or platform engineers"),
		Stack:        answer("Go"),
		Architecture: answer("Clean Architecture / Hexagonal"),
		Commands: ProjectBriefCommandSet{
			Install: answer("go mod download"),
			Test:    answer("go test ./..."),
			Build:   answer("go build ./..."),
			Run:     answer("go run ./cmd/specharbor"),
		},
		AgentBehavior: answer("Ask before assuming"),
	}
}
