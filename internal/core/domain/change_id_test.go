package domain

import (
	"strings"
	"testing"
)

func TestNewChangeIDAcceptsSafeIdentifiers(t *testing.T) {
	accepted := []string{
		"implement-advanced-validation",
		"add-example-feature",
		"change.v1",
		"change_v2",
		"Change-123",
		"a",
		strings.Repeat("a", 128),
	}

	for _, raw := range accepted {
		t.Run(raw, func(t *testing.T) {
			changeID, err := NewChangeID(raw)
			if err != nil {
				t.Fatalf("NewChangeID(%q) error = %v, want nil", raw, err)
			}
			if changeID.String() != raw {
				t.Fatalf("String() = %q, want %q", changeID.String(), raw)
			}
		})
	}
}

func TestNewChangeIDTrimsSurroundingWhitespace(t *testing.T) {
	changeID, err := NewChangeID("  change-v1  ")
	if err != nil {
		t.Fatalf("NewChangeID() error = %v, want nil", err)
	}
	if changeID.String() != "change-v1" {
		t.Fatalf("String() = %q, want %q", changeID.String(), "change-v1")
	}
}

func TestNewChangeIDRejectsUnsafeIdentifiers(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{name: "empty", raw: "", want: "change id is required"},
		{name: "whitespace only", raw: "   ", want: "change id is required"},
		{name: "single dot", raw: ".", want: "change id must not contain '.' or '..' path sequences"},
		{name: "double dot", raw: "..", want: "change id must not contain '.' or '..' path sequences"},
		{name: "internal double dot", raw: "a..b", want: "change id must not contain '.' or '..' path sequences"},
		{name: "traversal with separator", raw: "a/../b", want: "change id must be a single path segment"},
		{name: "relative traversal", raw: "../outside", want: "change id must be a single path segment"},
		{name: "forward slash", raw: "a/b", want: "change id must be a single path segment"},
		{name: "backslash", raw: `a\b`, want: "change id must be a single path segment"},
		{name: "absolute path", raw: "/absolute", want: "change id must be a single path segment"},
		{name: "leading dot", raw: ".hidden", want: "change id must not start with '.'"},
		{name: "leading dash", raw: "-flag", want: "change id must not start with '-'"},
		{name: "internal whitespace", raw: "change id", want: "change id contains unsupported character"},
		{name: "tab character", raw: "change\tid", want: "change id contains unsupported character"},
		{name: "unsafe character", raw: "change$id", want: "change id contains unsupported character"},
		{name: "non-ascii character", raw: "chänge", want: "change id contains unsupported character"},
		{name: "over length", raw: strings.Repeat("a", 129), want: "change id must be at most 128 characters"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewChangeID(test.raw)
			if err == nil {
				t.Fatalf("NewChangeID(%q) error = nil, want %q", test.raw, test.want)
			}
			if !strings.Contains(err.Error(), test.want) {
				t.Fatalf("NewChangeID(%q) error = %q, want %q", test.raw, err.Error(), test.want)
			}
		})
	}
}
