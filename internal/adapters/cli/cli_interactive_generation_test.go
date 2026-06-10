package cli

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/guferreira1/spec-harbor/internal/core/domain"
)

func TestParseGenerateInteractiveArguments(t *testing.T) {
	arguments, err := parseGenerateArguments([]string{"change", "--interactive"})
	if err != nil {
		t.Fatalf("parseGenerateArguments() error = %v", err)
	}
	if !arguments.interactive {
		t.Fatalf("interactive = false, want true")
	}
	if arguments.changeID != "change" {
		t.Fatalf("changeID = %q, want change", arguments.changeID)
	}

	arguments, err = parseGenerateArguments([]string{"--interactive", "change"})
	if err != nil {
		t.Fatalf("parseGenerateArguments(flag first) error = %v", err)
	}
	if !arguments.interactive || arguments.changeID != "change" {
		t.Fatalf("arguments = %+v, want interactive change", arguments)
	}
}

func TestParseGenerateInteractiveRejectsInvalidArguments(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "duplicate interactive", args: []string{"change", "--interactive", "--interactive"}, want: "interactive generation flag specified more than once"},
		{name: "missing change id", args: []string{"--interactive"}, want: "change id is required"},
		{name: "extra argument", args: []string{"change", "extra", "--interactive"}, want: "unexpected argument: extra"},
		{name: "blank conflict", args: []string{"change", "--interactive", "--blank"}, want: "interactive and blank generation flags cannot be used together"},
		{name: "template conflict", args: []string{"change", "--interactive", "--template", "feature"}, want: "interactive and template generation flags cannot be used together"},
		{name: "custom template conflict", args: []string{"change", "--interactive", "--custom-template", "api-feature"}, want: "interactive and custom-template generation flags cannot be used together"},
		{name: "config template conflict", args: []string{"change", "--interactive", "--config-template", "api-feature"}, want: "interactive and config-template generation flags cannot be used together"},
		{name: "guided conflict", args: []string{"change", "--interactive", "--guided"}, want: "interactive and guided generation flags cannot be used together"},
		{name: "hybrid conflict", args: []string{"change", "--interactive", "--hybrid"}, want: "interactive and hybrid generation flags cannot be used together"},
		{name: "ai assisted conflict", args: []string{"change", "--interactive", "--ai-assisted"}, want: "interactive and ai-assisted generation flags cannot be used together"},
		{name: "agent assisted conflict", args: []string{"change", "--interactive", "--agent-assisted"}, want: "interactive and agent-assisted generation flags cannot be used together"},
		{name: "type conflict", args: []string{"change", "--interactive", "--type", "feature"}, want: "interactive and type flags cannot be used together"},
		{name: "title conflict", args: []string{"change", "--interactive", "--title", "Title"}, want: "interactive and title flags cannot be used together"},
		{name: "summary conflict", args: []string{"change", "--interactive", "--summary", "Summary"}, want: "interactive and summary flags cannot be used together"},
		{name: "from file conflict", args: []string{"change", "--interactive", "--from-file", "agent-output.txt"}, want: "interactive and from-file flags cannot be used together"},
		{name: "agent conflict", args: []string{"change", "--interactive", "--agent", "codex"}, want: "interactive and agent flags cannot be used together"},
		{name: "execute conflict", args: []string{"change", "--interactive", "--execute"}, want: "interactive and execute flags cannot be used together"},
		{name: "overwrite conflict", args: []string{"change", "--interactive", "--overwrite"}, want: "interactive and overwrite flags cannot be used together"},
		{name: "missing template value keeps existing style", args: []string{"change", "--interactive", "--template"}, want: "template name is required"},
		{name: "unsupported flag keeps existing style", args: []string{"change", "--interactive", "--force"}, want: "unsupported flag: --force"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := parseGenerateArguments(test.args)
			if err == nil {
				t.Fatalf("parseGenerateArguments(%v) error = nil, want %q", test.args, test.want)
			}
			if err.Error() != test.want {
				t.Fatalf("parseGenerateArguments(%v) error = %q, want %q", test.args, err.Error(), test.want)
			}
		})
	}
}

func TestExecuteGenerateInteractiveNonTTYFailsWithoutPromptOrWrites(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	createOpenSpecProject(t, root)

	var output bytes.Buffer
	terminal := &cliFakeInteractiveTerminal{isTTY: false, output: &output, inputs: []string{"blank", "y"}}
	err := executeWithTerminal([]string{"generate", "interactive-change", "--interactive"}, &output, terminal)
	if err == nil || err.Error() != "interactive mode requires a TTY" {
		t.Fatalf("execute interactive non-tty error = %v, want TTY error", err)
	}
	if output.String() != "" {
		t.Fatalf("output = %q, want empty", output.String())
	}
	if terminal.reads != 0 {
		t.Fatalf("terminal reads = %d, want 0", terminal.reads)
	}
	assertPathDoesNotExist(t, root, "openspec/changes/interactive-change")
}

func TestExecuteGenerateInteractiveValidatesChangeIDBeforePrompting(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	createOpenSpecProject(t, root)

	var output bytes.Buffer
	terminal := &cliFakeInteractiveTerminal{isTTY: true, output: &output, inputs: []string{"blank", "y"}}
	err := executeWithTerminal([]string{"generate", "bad/id", "--interactive"}, &output, terminal)
	if err == nil || err.Error() != "change id must be a single path segment" {
		t.Fatalf("execute interactive invalid change id error = %v, want ChangeID validation error", err)
	}
	if output.String() != "" {
		t.Fatalf("output = %q, want empty", output.String())
	}
	if terminal.reads != 0 {
		t.Fatalf("terminal reads = %d, want 0", terminal.reads)
	}
	assertPathDoesNotExist(t, root, "openspec/changes/bad")
}

func TestExecuteGenerateInteractivePromptFlows(t *testing.T) {
	tests := []struct {
		name       string
		changeID   string
		setup      func(t *testing.T, root string)
		inputs     []string
		want       []string
		forbidden  []string
		assertFile func(t *testing.T, root string, changeID string)
	}{
		{
			name:     "blank",
			changeID: "interactive-blank",
			inputs:   []string{"1", "y"},
			want: []string{
				"Select generation path:",
				"Generation path: blank",
				"Expected write target: openspec/changes/interactive-blank/",
				"Validation: automatic no",
				"SpecHarbor blank change generated.",
				"Created files: 5",
			},
			forbidden: []string{"Select built-in template:", "Custom template name:", "Config template alias:", "Title:"},
		},
		{
			name:     "built-in template",
			changeID: "interactive-built-in",
			inputs:   []string{"template", "2", "Y"},
			want: []string{
				"Select built-in template:",
				"Generation path: built-in template",
				"Template: bugfix",
				"Validation: automatic no",
				"SpecHarbor template change generated.",
				"Template: bugfix",
			},
			forbidden: []string{"Custom template name:", "Config template alias:"},
		},
		{
			name:     "custom template",
			changeID: "interactive-custom",
			setup: func(t *testing.T, root string) {
				createCustomTemplateDirectory(t, root, "api-feature")
			},
			inputs: []string{"custom", "api-feature", "Add payments", "Adds a payment flow.", "YES"},
			want: []string{
				"Generation path: custom template",
				"Custom template: api-feature",
				"Title: Add payments",
				"Summary: Adds a payment flow.",
				"Validation: automatic no",
				"SpecHarbor custom template change generated.",
			},
			forbidden: []string{"Select built-in template:", "Config template alias:"},
			assertFile: func(t *testing.T, root string, changeID string) {
				proposal := readInteractiveGeneratedFile(t, root, changeID, "proposal.md")
				if !strings.Contains(proposal, "Title: Add payments") || !strings.Contains(proposal, "Summary: Adds a payment flow.") {
					t.Fatalf("proposal.md = %q, want title and summary substitutions", proposal)
				}
			},
		},
		{
			name:     "config template",
			changeID: "interactive-config",
			setup: func(t *testing.T, root string) {
				writeConfigTemplateConfig(t, root, `
    default-feature:
      source: builtin
      template: feature
`)
			},
			inputs: []string{"config", "default-feature", "", "", "Yes"},
			want: []string{
				"Generation path: config template",
				"Config alias: default-feature",
				"Validation: automatic no",
				"SpecHarbor config template change generated.",
				"Config template: default-feature",
				"Resolved source: builtin",
			},
			forbidden: []string{"Custom template name:", "Title: ", "Summary: "},
		},
		{
			name:     "hybrid built-in",
			changeID: "interactive-hybrid",
			inputs: []string{
				"hybrid",
				"built-in template",
				"feature",
				"Add login",
				"Add login support",
				"",
				"yes",
			},
			want: []string{
				"Select hybrid source:",
				"Hybrid source: built-in template",
				"Template: feature",
				"Title: Add login",
				"Summary: Add login support",
				"Validation: automatic yes",
				"SpecHarbor hybrid change generated.",
				"Validation:",
				"Status: valid",
			},
			forbidden: []string{"Custom template name:", "Config template alias:"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			t.Chdir(root)
			createOpenSpecProject(t, root)
			if test.setup != nil {
				test.setup(t, root)
			}

			output, err := executeInteractiveGenerate(t, test.changeID, test.inputs)
			if err != nil {
				t.Fatalf("execute interactive %s error = %v\noutput:\n%s", test.name, err, output)
			}
			for _, want := range test.want {
				if !strings.Contains(output, want) {
					t.Fatalf("interactive output = %q, want %q", output, want)
				}
			}
			for _, forbidden := range test.forbidden {
				if strings.Contains(output, forbidden) {
					t.Fatalf("interactive output = %q, must not contain %q", output, forbidden)
				}
			}
			assertInteractiveSummarySafety(t, output)
			for _, requiredFile := range domain.RequiredOpenSpecChangeFiles() {
				assertPathExists(t, root, filepath.ToSlash(filepath.Join("openspec", "changes", test.changeID, requiredFile)))
			}
			if test.assertFile != nil {
				test.assertFile(t, root, test.changeID)
			}
		})
	}
}

func TestExecuteGenerateInteractiveHybridCustomAndConfigFlows(t *testing.T) {
	tests := []struct {
		name     string
		changeID string
		setup    func(t *testing.T, root string)
		inputs   []string
		want     []string
	}{
		{
			name:     "custom source",
			changeID: "interactive-hybrid-custom",
			setup: func(t *testing.T, root string) {
				createHybridCustomTemplateDirectory(t, root, "api-feature", true)
			},
			inputs: []string{"5", "2", "api-feature", "Add payments", "Adds payments.", "docs", "y"},
			want: []string{
				"Hybrid source: custom template",
				"Custom template: api-feature",
				"Type: docs",
				"Source kind: custom",
				"Source: api-feature",
				"Type: docs",
			},
		},
		{
			name:     "config source",
			changeID: "interactive-hybrid-config",
			setup: func(t *testing.T, root string) {
				writeConfigTemplateConfig(t, root, `
    default-feature:
      source: builtin
      template: feature
`)
			},
			inputs: []string{"5", "3", "default-feature", "Add login", "Adds login.", "", "y"},
			want: []string{
				"Hybrid source: config template",
				"Config alias: default-feature",
				"Source kind: config",
				"Source: default-feature",
				"Resolved source: builtin",
				"Type: feature",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			t.Chdir(root)
			createOpenSpecProject(t, root)
			test.setup(t, root)

			output, err := executeInteractiveGenerate(t, test.changeID, test.inputs)
			if err != nil {
				t.Fatalf("execute interactive hybrid error = %v\noutput:\n%s", err, output)
			}
			for _, want := range test.want {
				if !strings.Contains(output, want) {
					t.Fatalf("interactive hybrid output = %q, want %q", output, want)
				}
			}
			if !strings.Contains(output, "Validation: automatic yes") {
				t.Fatalf("interactive hybrid output = %q, want automatic validation summary", output)
			}
		})
	}
}

func TestExecuteGenerateInteractiveHybridTypeValidationMatchesDirectHybrid(t *testing.T) {
	t.Run("rejects uppercase and mixed case like direct hybrid", func(t *testing.T) {
		tests := []struct {
			name       string
			typeAnswer string
		}{
			{name: "mixed case", typeAnswer: "Feature"},
			{name: "uppercase", typeAnswer: "FEATURE"},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				directError := directHybridTypeValidationError(t, test.typeAnswer)
				root := t.TempDir()
				t.Chdir(root)
				createOpenSpecProject(t, root)

				output, err := executeInteractiveGenerate(t, "hybrid-type-"+strings.ReplaceAll(test.name, " ", "-"), []string{
					"hybrid",
					"template",
					"feature",
					"Add login",
					"Add login support",
					test.typeAnswer,
					test.typeAnswer,
					test.typeAnswer,
				})
				if err == nil || err.Error() != "hybrid type retry limit exceeded" {
					t.Fatalf("execute interactive hybrid type error = %v, want retry limit\noutput:\n%s", err, output)
				}
				if !strings.Contains(output, "Invalid answer: "+directError) {
					t.Fatalf("interactive output = %q, want direct validation error %q", output, directError)
				}
				assertPathDoesNotExist(t, root, "openspec/changes/hybrid-type-"+strings.ReplaceAll(test.name, " ", "-"))
			})
		}
	})

	t.Run("accepts exact lowercase supported type", func(t *testing.T) {
		root := t.TempDir()
		t.Chdir(root)
		createOpenSpecProject(t, root)

		output, err := executeInteractiveGenerate(t, "hybrid-type-lowercase", []string{
			"hybrid",
			"template",
			"feature",
			"Add login",
			"Add login support",
			"feature",
			"y",
		})
		if err != nil {
			t.Fatalf("execute interactive lowercase hybrid type error = %v\noutput:\n%s", err, output)
		}
		if !strings.Contains(output, "Type: feature") {
			t.Fatalf("interactive output = %q, want lowercase type summary/report", output)
		}
		assertPathExists(t, root, "openspec/changes/hybrid-type-lowercase/proposal.md")
	})
}

func TestExecuteGenerateInteractiveConfirmationProceedIsCaseInsensitive(t *testing.T) {
	tests := []struct {
		name   string
		answer string
	}{
		{name: "Y", answer: "Y"},
		{name: "YES", answer: "YES"},
		{name: "Yes", answer: "Yes"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			t.Chdir(root)
			createOpenSpecProject(t, root)

			changeID := "proceed-" + strings.ToLower(test.name)
			output, err := executeInteractiveGenerate(t, changeID, []string{"blank", test.answer})
			if err != nil {
				t.Fatalf("execute interactive proceed %q error = %v\noutput:\n%s", test.answer, err, output)
			}
			assertPathExists(t, root, "openspec/changes/"+changeID+"/proposal.md")
		})
	}
}

func TestExecuteGenerateInteractiveConfirmationTrimsWhitespace(t *testing.T) {
	t.Run("proceed variants generate", func(t *testing.T) {
		tests := []struct {
			name   string
			answer string
		}{
			{name: "y", answer: " Y "},
			{name: "yes upper", answer: " YES "},
			{name: "yes mixed", answer: " Yes "},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				root := t.TempDir()
				t.Chdir(root)
				createOpenSpecProject(t, root)

				changeID := "trimmed-proceed-" + strings.ReplaceAll(test.name, " ", "-")
				output, err := executeInteractiveGenerate(t, changeID, []string{"blank", test.answer})
				if err != nil {
					t.Fatalf("execute interactive proceed %q error = %v\noutput:\n%s", test.answer, err, output)
				}
				assertPathExists(t, root, "openspec/changes/"+changeID+"/proposal.md")
			})
		}
	})

	t.Run("cancel variants write nothing", func(t *testing.T) {
		tests := []struct {
			name   string
			answer string
		}{
			{name: "n", answer: " N "},
			{name: "no upper", answer: " NO "},
			{name: "no mixed", answer: " No "},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				root := t.TempDir()
				t.Chdir(root)
				createOpenSpecProject(t, root)

				changeID := "trimmed-cancel-" + strings.ReplaceAll(test.name, " ", "-")
				output, err := executeInteractiveGenerate(t, changeID, []string{"blank", test.answer})
				if err == nil || err.Error() != "operation cancelled" {
					t.Fatalf("execute interactive cancel %q error = %v, want operation cancelled\noutput:\n%s", test.answer, err, output)
				}
				assertPathDoesNotExist(t, root, "openspec/changes/"+changeID)
			})
		}
	})
}

func TestExecuteGenerateInteractiveCancellationWritesNothing(t *testing.T) {
	tests := []struct {
		name   string
		inputs []string
	}{
		{name: "n", inputs: []string{"blank", "n"}},
		{name: "N", inputs: []string{"blank", "N"}},
		{name: "no", inputs: []string{"blank", "no"}},
		{name: "NO", inputs: []string{"blank", "NO"}},
		{name: "No", inputs: []string{"blank", "No"}},
		{name: "empty confirmation", inputs: []string{"blank", ""}},
		{name: "EOF confirmation", inputs: []string{"blank"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			t.Chdir(root)
			createOpenSpecProject(t, root)

			changeID := "cancel-" + strings.ReplaceAll(strings.ToLower(test.name), " ", "-")
			output, err := executeInteractiveGenerate(t, changeID, test.inputs)
			if err == nil || err.Error() != "operation cancelled" {
				t.Fatalf("execute interactive cancel error = %v, want operation cancelled\noutput:\n%s", err, output)
			}
			if !strings.Contains(output, "Proceed? [y/N]:") {
				t.Fatalf("cancel output = %q, want confirmation prompt", output)
			}
			assertPathDoesNotExist(t, root, "openspec/changes/"+changeID)
		})
	}
}

func TestExecuteGenerateInteractiveInvalidInputRetryAndExhaustion(t *testing.T) {
	t.Run("invalid menu retries then succeeds", func(t *testing.T) {
		root := t.TempDir()
		t.Chdir(root)
		createOpenSpecProject(t, root)

		output, err := executeInteractiveGenerate(t, "retry-menu", []string{"unknown", "blank", "y"})
		if err != nil {
			t.Fatalf("execute interactive retry menu error = %v\noutput:\n%s", err, output)
		}
		if !strings.Contains(output, "Invalid answer: unsupported generation path: unknown") {
			t.Fatalf("output = %q, want invalid menu retry", output)
		}
		assertPathExists(t, root, "openspec/changes/retry-menu/proposal.md")
	})

	t.Run("invalid value retries then succeeds", func(t *testing.T) {
		root := t.TempDir()
		t.Chdir(root)
		createOpenSpecProject(t, root)

		output, err := executeInteractiveGenerate(t, "retry-template", []string{"2", "maintenance", "feature", "y"})
		if err != nil {
			t.Fatalf("execute interactive retry value error = %v\noutput:\n%s", err, output)
		}
		if !strings.Contains(output, "Invalid answer: unknown template name: maintenance") {
			t.Fatalf("output = %q, want invalid template retry", output)
		}
		assertPathExists(t, root, "openspec/changes/retry-template/proposal.md")
	})

	t.Run("invalid config alias retries then succeeds", func(t *testing.T) {
		root := t.TempDir()
		t.Chdir(root)
		createOpenSpecProject(t, root)
		writeConfigTemplateConfig(t, root, `
    default-feature:
      source: builtin
      template: feature
`)

		output, err := executeInteractiveGenerate(t, "retry-config", []string{"config", "../escape", "default-feature", "", "", "y"})
		if err != nil {
			t.Fatalf("execute interactive retry config error = %v\noutput:\n%s", err, output)
		}
		if !strings.Contains(output, "Invalid answer: config template alias must be a single path segment") {
			t.Fatalf("output = %q, want invalid config alias retry", output)
		}
		assertPathExists(t, root, "openspec/changes/retry-config/proposal.md")
	})

	t.Run("invalid hybrid type retries then succeeds", func(t *testing.T) {
		root := t.TempDir()
		t.Chdir(root)
		createOpenSpecProject(t, root)

		output, err := executeInteractiveGenerate(t, "retry-hybrid-type", []string{
			"hybrid",
			"template",
			"feature",
			"Add login",
			"Add login support",
			"maintenance",
			"",
			"y",
		})
		if err != nil {
			t.Fatalf("execute interactive retry hybrid type error = %v\noutput:\n%s", err, output)
		}
		if !strings.Contains(output, "Invalid answer: unsupported hybrid type: maintenance") {
			t.Fatalf("output = %q, want invalid hybrid type retry", output)
		}
		assertPathExists(t, root, "openspec/changes/retry-hybrid-type/proposal.md")
	})

	t.Run("empty hybrid title retries then succeeds", func(t *testing.T) {
		root := t.TempDir()
		t.Chdir(root)
		createOpenSpecProject(t, root)

		output, err := executeInteractiveGenerate(t, "retry-hybrid-title", []string{
			"hybrid",
			"template",
			"feature",
			"",
			"Add login",
			"Add login support",
			"",
			"y",
		})
		if err != nil {
			t.Fatalf("execute interactive retry hybrid title error = %v\noutput:\n%s", err, output)
		}
		if !strings.Contains(output, "Invalid answer: hybrid title is required") {
			t.Fatalf("output = %q, want invalid hybrid title retry", output)
		}
		assertPathExists(t, root, "openspec/changes/retry-hybrid-title/proposal.md")
	})

	t.Run("empty hybrid title exhaustion writes nothing", func(t *testing.T) {
		root := t.TempDir()
		t.Chdir(root)
		createOpenSpecProject(t, root)

		output, err := executeInteractiveGenerate(t, "hybrid-title-exhausted", []string{
			"hybrid",
			"template",
			"feature",
			"",
			" ",
			"\t",
		})
		if err == nil || err.Error() != "hybrid title retry limit exceeded" {
			t.Fatalf("execute interactive hybrid title exhaustion error = %v, want retry limit\noutput:\n%s", err, output)
		}
		assertPathDoesNotExist(t, root, "openspec/changes/hybrid-title-exhausted")
	})

	t.Run("empty hybrid summary retries then succeeds", func(t *testing.T) {
		root := t.TempDir()
		t.Chdir(root)
		createOpenSpecProject(t, root)

		output, err := executeInteractiveGenerate(t, "retry-hybrid-summary", []string{
			"hybrid",
			"template",
			"feature",
			"Add login",
			"",
			"Add login support",
			"",
			"y",
		})
		if err != nil {
			t.Fatalf("execute interactive retry hybrid summary error = %v\noutput:\n%s", err, output)
		}
		if !strings.Contains(output, "Invalid answer: hybrid summary is required") {
			t.Fatalf("output = %q, want invalid hybrid summary retry", output)
		}
		assertPathExists(t, root, "openspec/changes/retry-hybrid-summary/proposal.md")
	})

	t.Run("empty hybrid summary exhaustion writes nothing", func(t *testing.T) {
		root := t.TempDir()
		t.Chdir(root)
		createOpenSpecProject(t, root)

		output, err := executeInteractiveGenerate(t, "hybrid-summary-exhausted", []string{
			"hybrid",
			"template",
			"feature",
			"Add login",
			"",
			" ",
			"\t",
		})
		if err == nil || err.Error() != "hybrid summary retry limit exceeded" {
			t.Fatalf("execute interactive hybrid summary exhaustion error = %v, want retry limit\noutput:\n%s", err, output)
		}
		assertPathDoesNotExist(t, root, "openspec/changes/hybrid-summary-exhausted")
	})

	t.Run("menu exhaustion writes nothing", func(t *testing.T) {
		root := t.TempDir()
		t.Chdir(root)
		createOpenSpecProject(t, root)

		output, err := executeInteractiveGenerate(t, "retry-exhausted", []string{"bad", "worse", ""})
		if err == nil || err.Error() != "generation path retry limit exceeded" {
			t.Fatalf("execute interactive exhaustion error = %v, want retry limit\noutput:\n%s", err, output)
		}
		assertPathDoesNotExist(t, root, "openspec/changes/retry-exhausted")
	})

	t.Run("value exhaustion writes nothing", func(t *testing.T) {
		root := t.TempDir()
		t.Chdir(root)
		createOpenSpecProject(t, root)

		output, err := executeInteractiveGenerate(t, "value-exhausted", []string{"custom", "../escape", "../escape", "../escape"})
		if err == nil || err.Error() != "custom template retry limit exceeded" {
			t.Fatalf("execute interactive value exhaustion error = %v, want retry limit\noutput:\n%s", err, output)
		}
		assertPathDoesNotExist(t, root, "openspec/changes/value-exhausted")
	})
}

func TestExecuteGenerateInteractiveInvalidConfirmationRetryAndExhaustion(t *testing.T) {
	t.Run("invalid confirmation retries then succeeds", func(t *testing.T) {
		root := t.TempDir()
		t.Chdir(root)
		createOpenSpecProject(t, root)

		output, err := executeInteractiveGenerate(t, "confirmation-retry", []string{"blank", "maybe", "y"})
		if err != nil {
			t.Fatalf("execute interactive confirmation retry error = %v\noutput:\n%s", err, output)
		}
		if !strings.Contains(output, "Invalid answer: enter y/yes or n/no.") {
			t.Fatalf("output = %q, want invalid confirmation retry", output)
		}
		assertPathExists(t, root, "openspec/changes/confirmation-retry/proposal.md")
	})

	t.Run("invalid confirmation exhaustion writes nothing", func(t *testing.T) {
		root := t.TempDir()
		t.Chdir(root)
		createOpenSpecProject(t, root)

		output, err := executeInteractiveGenerate(t, "confirmation-exhausted", []string{"blank", "maybe", "later", "go"})
		if err == nil || err.Error() != "confirmation retry limit exceeded" {
			t.Fatalf("execute interactive confirmation exhaustion error = %v, want retry limit\noutput:\n%s", err, output)
		}
		assertPathDoesNotExist(t, root, "openspec/changes/confirmation-exhausted")
	})
}

func TestExecuteGenerateInteractiveRemoteConfigAliasRunsOnlyAfterConfirmation(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	createOpenSpecProject(t, root)
	downloadedBytes := []byte("zip bytes")
	checksum := domain.NewRemoteTemplateChecksumFromBytes(downloadedBytes).String()
	writeConfigTemplateConfig(t, root, `
    service-feature:
      source: remote
      url: https://example.com/templates/service-feature.zip
      checksum: `+checksum+`
      format: zip
`)
	fetcher := &cliFakeRemoteTemplateFetcher{result: domain.NewRemoteTemplateFetchResult(200, downloadedBytes)}
	reader := &cliFakeRemoteTemplateBundleReader{bundle: mustCLIRemoteTemplateBundle(t, "remote")}
	withCLIRemoteTemplateFactories(t, fetcher, reader)

	output, err := executeInteractiveGenerate(t, "remote-cancelled", []string{"config", "service-feature", "", "", "n"})
	if err == nil || err.Error() != "operation cancelled" {
		t.Fatalf("execute interactive remote cancel error = %v, want operation cancelled\noutput:\n%s", err, output)
	}
	if fetcher.calls != 0 || reader.calls != 0 {
		t.Fatalf("remote calls after cancellation = fetcher %d reader %d, want 0", fetcher.calls, reader.calls)
	}
	assertPathDoesNotExist(t, root, "openspec/changes/remote-cancelled")

	output, err = executeInteractiveGenerate(t, "remote-confirmed", []string{"config", "service-feature", "", "", "y"})
	if err != nil {
		t.Fatalf("execute interactive remote confirmed error = %v\noutput:\n%s", err, output)
	}
	if fetcher.calls != 1 || reader.calls != 1 {
		t.Fatalf("remote calls after confirmation = fetcher %d reader %d, want 1", fetcher.calls, reader.calls)
	}
	for _, want := range []string{"Remote host: example.com", "Remote format: zip", "Checksum: sha256"} {
		if !strings.Contains(output, want) {
			t.Fatalf("remote output = %q, want %q", output, want)
		}
	}
	for _, forbidden := range []string{"https://", "#", "token=", "user:", "supersecret"} {
		if strings.Contains(output, forbidden) {
			t.Fatalf("remote output leaked %q in %q", forbidden, output)
		}
	}
}

func TestExecuteGenerateInteractivePreservesExistingFileBehavior(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	createOpenSpecProject(t, root)

	if _, err := executeInteractiveGenerate(t, "existing-change", []string{"blank", "y"}); err != nil {
		t.Fatalf("first interactive blank error = %v", err)
	}
	proposalPath := filepath.Join(root, "openspec", "changes", "existing-change", "proposal.md")
	if err := os.WriteFile(proposalPath, []byte("manual edit"), 0o644); err != nil {
		t.Fatalf("WriteFile(proposal.md) error = %v", err)
	}

	output, err := executeInteractiveGenerate(t, "existing-change", []string{"blank", "y"})
	if err != nil {
		t.Fatalf("second interactive blank error = %v\noutput:\n%s", err, output)
	}
	if !strings.Contains(output, "Skipped existing files: 5") {
		t.Fatalf("second output = %q, want skipped existing files", output)
	}
	contents, err := os.ReadFile(proposalPath)
	if err != nil {
		t.Fatalf("ReadFile(proposal.md) error = %v", err)
	}
	if string(contents) != "manual edit" {
		t.Fatalf("proposal.md = %q, want preserved manual edit", string(contents))
	}
}

func TestExecuteGenerateInteractiveWritesOnlyOpenSpecChangeFiles(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	createOpenSpecProject(t, root)

	output, err := executeInteractiveGenerate(t, "write-surface", []string{"blank", "y"})
	if err != nil {
		t.Fatalf("interactive blank error = %v\noutput:\n%s", err, output)
	}

	for _, forbidden := range []string{
		"docs",
		".github",
		".git",
		".specharbor/config.yml",
		"agent-prompts",
		"archive",
		"cmd",
		"internal",
		"production",
		"go.mod",
		"go.sum",
	} {
		assertPathDoesNotExist(t, root, forbidden)
	}
	for _, requiredFile := range domain.RequiredOpenSpecChangeFiles() {
		assertPathExists(t, root, filepath.ToSlash(filepath.Join("openspec", "changes", "write-surface", requiredFile)))
	}
}

func TestExecuteGenerateInteractiveHybridValidationBehaviorIsPreserved(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	createOpenSpecProject(t, root)
	createHybridCustomTemplateDirectory(t, root, "invalid-template", false)

	output, err := executeInteractiveGenerate(t, "invalid-hybrid", []string{
		"hybrid",
		"custom",
		"invalid-template",
		"Invalid change",
		"Uses invalid task content.",
		"",
		"y",
	})
	assertExitCode(t, err, 1)
	for _, want := range []string{"Validation: automatic yes", "Validation:", "Status: invalid", "Errors:"} {
		if !strings.Contains(output, want) {
			t.Fatalf("interactive hybrid validation output = %q, want %q", output, want)
		}
	}
}

func executeInteractiveGenerate(t *testing.T, changeID string, inputs []string) (string, error) {
	t.Helper()

	var output bytes.Buffer
	terminal := &cliFakeInteractiveTerminal{
		isTTY:  true,
		output: &output,
		inputs: inputs,
	}
	err := executeWithTerminal([]string{"generate", changeID, "--interactive"}, &output, terminal)
	return output.String(), err
}

func directHybridTypeValidationError(t *testing.T, hybridType string) string {
	t.Helper()

	_, err := parseGenerateArguments([]string{
		"change",
		"--hybrid",
		"--template",
		"feature",
		"--title",
		"Title",
		"--summary",
		"Summary",
		"--type",
		hybridType,
	})
	if err == nil {
		t.Fatalf("parseGenerateArguments accepted hybrid type %q, want validation error", hybridType)
	}
	return err.Error()
}

type cliFakeInteractiveTerminal struct {
	isTTY  bool
	output io.Writer
	inputs []string
	reads  int
}

func (terminal *cliFakeInteractiveTerminal) IsInputTerminal() bool {
	return terminal.isTTY
}

func (terminal *cliFakeInteractiveTerminal) ReadLine() (string, error) {
	if terminal.reads >= len(terminal.inputs) {
		return "", io.EOF
	}
	value := terminal.inputs[terminal.reads]
	terminal.reads++
	return value, nil
}

func (terminal *cliFakeInteractiveTerminal) WriteString(value string) error {
	_, err := io.WriteString(terminal.output, value)
	return err
}

func assertInteractiveSummarySafety(t *testing.T, output string) {
	t.Helper()

	for _, want := range []string{
		"Safety:",
		"- Writes are limited to OpenSpec change files.",
		"- Production code will not be modified.",
		"- Source-control commands will not be run.",
		"- Workflow automation will not be triggered.",
		"- Provider, LLM, and agent APIs will not be called.",
		"- No auto-commit, auto-push, PR creation, merge, or archive will be performed.",
		"Proceed? [y/N]:",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %q, want safety summary line %q", output, want)
		}
	}
	if strings.Index(output, "Safety:") > strings.Index(output, "Proceed? [y/N]:") {
		t.Fatalf("output = %q, want safety before confirmation", output)
	}
}

func readInteractiveGeneratedFile(t *testing.T, root string, changeID string, fileName string) string {
	t.Helper()

	contents, err := os.ReadFile(filepath.Join(root, "openspec", "changes", changeID, fileName))
	if err != nil {
		t.Fatalf("ReadFile(%s) error = %v", fileName, err)
	}
	return string(contents)
}
