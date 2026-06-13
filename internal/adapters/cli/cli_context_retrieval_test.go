package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExecuteContextRetrievePrintsSourceAttributedResults(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	createOpenSpecProject(t, root)
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("# Project\n\nHexagonal Architecture keeps boundaries clear.\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(README.md) error = %v", err)
	}
	writeContextIndexForCLITest(t)

	var output bytes.Buffer
	if err := execute([]string{"context", "retrieve", "--query", "hexagonal architecture"}, &output); err != nil {
		t.Fatalf("execute(context retrieve) error = %v\noutput:\n%s", err, output.String())
	}
	got := output.String()
	for _, want := range []string{
		"Local context retrieval:",
		"Query: hexagonal architecture",
		"Normalized terms: hexagonal, architecture",
		"Index: .specharbor/context-index.json",
		"Index status: current",
		"Results:",
		"README.md",
		"Category: readme",
		"Score:",
		"Lines:",
		"Snippet:",
		"Hexagonal Architecture keeps boundaries clear.",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("retrieve output = %q, want %q", got, want)
		}
	}
}

func TestExecuteContextRetrieveReportsNoResults(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	createOpenSpecProject(t, root)
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("# Project\nNo matching text.\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(README.md) error = %v", err)
	}
	writeContextIndexForCLITest(t)

	var output bytes.Buffer
	if err := execute([]string{"context", "retrieve", "--query", "nonexistentterm"}, &output); err != nil {
		t.Fatalf("execute(context retrieve no results) error = %v\noutput:\n%s", err, output.String())
	}
	if !strings.Contains(output.String(), "No matching local context found.") {
		t.Fatalf("output = %q, want no-results message", output.String())
	}
}

func TestExecuteContextRetrieveRejectsInvalidQueriesAndFlags(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "missing query", args: []string{"context", "retrieve"}, want: "context retrieve query is required"},
		{name: "empty query", args: []string{"context", "retrieve", "--query", ""}, want: "retrieval query is required"},
		{name: "oversized query", args: []string{"context", "retrieve", "--query", strings.Repeat("a", 513)}, want: "retrieval query must be at most 512 characters"},
		{name: "positional query", args: []string{"context", "retrieve", "architecture"}, want: "unexpected argument: architecture"},
		{name: "unsupported flag", args: []string{"context", "retrieve", "--query", "architecture", "--rag"}, want: "unsupported flag: --rag"},
		{name: "duplicate query", args: []string{"context", "retrieve", "--query", "one", "--query", "two"}, want: "context retrieve query flag specified more than once"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var output bytes.Buffer
			err := execute(test.args, &output)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("execute(%v) error = %v, want %q", test.args, err, test.want)
			}
			if output.String() != "" {
				t.Fatalf("output = %q, want empty", output.String())
			}
		})
	}
}

func TestExecuteContextRetrieveFailsForMissingAndStaleIndex(t *testing.T) {
	t.Run("missing", func(t *testing.T) {
		root := t.TempDir()
		t.Chdir(root)
		createOpenSpecProject(t, root)
		if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("# Project\nArchitecture\n"), 0o644); err != nil {
			t.Fatalf("WriteFile(README.md) error = %v", err)
		}

		var output bytes.Buffer
		err := execute([]string{"context", "retrieve", "--query", "architecture"}, &output)
		if err == nil {
			t.Fatalf("execute(context retrieve) error = nil, want missing index failure")
		}
		if !strings.Contains(output.String(), "Index status: missing_index") ||
			!strings.Contains(output.String(), "specharbor context index --write") {
			t.Fatalf("output = %q, want missing index report", output.String())
		}
	})

	t.Run("stale", func(t *testing.T) {
		root := t.TempDir()
		t.Chdir(root)
		createOpenSpecProject(t, root)
		readmePath := filepath.Join(root, "README.md")
		if err := os.WriteFile(readmePath, []byte("# Project\nArchitecture\n"), 0o644); err != nil {
			t.Fatalf("WriteFile(README.md) error = %v", err)
		}
		writeContextIndexForCLITest(t)
		if err := os.WriteFile(readmePath, []byte("# Project\nArchitecture changed\n"), 0o644); err != nil {
			t.Fatalf("WriteFile(README.md changed) error = %v", err)
		}

		var output bytes.Buffer
		err := execute([]string{"context", "retrieve", "--query", "architecture"}, &output)
		if err == nil {
			t.Fatalf("execute(context retrieve) error = nil, want stale index failure")
		}
		if !strings.Contains(output.String(), "Index status: stale_index") ||
			!strings.Contains(output.String(), "Stale reasons:") {
			t.Fatalf("output = %q, want stale index report", output.String())
		}
	})
}

func writeContextIndexForCLITest(t *testing.T) {
	t.Helper()
	var output bytes.Buffer
	if err := execute([]string{"context", "index", "--write"}, &output); err != nil {
		t.Fatalf("execute(context index --write) error = %v\noutput:\n%s", err, output.String())
	}
}
