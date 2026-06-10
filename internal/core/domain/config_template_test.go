package domain

import (
	"strings"
	"testing"
)

func TestNewConfigTemplateAliasAcceptsSafeAliases(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{name: "kebab case", raw: "api-feature", want: "api-feature"},
		{name: "snake case", raw: "api_feature", want: "api_feature"},
		{name: "mixed alphanumeric", raw: "Feature42", want: "Feature42"},
		{name: "internal single dot", raw: "feature.v1", want: "feature.v1"},
		{name: "trimmed whitespace", raw: "  default-feature  ", want: "default-feature"},
		{name: "max length", raw: strings.Repeat("a", 128), want: strings.Repeat("a", 128)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			alias, err := NewConfigTemplateAlias(test.raw)
			if err != nil {
				t.Fatalf("NewConfigTemplateAlias(%q) error = %v", test.raw, err)
			}
			if alias.String() != test.want {
				t.Fatalf("String() = %q, want %q", alias.String(), test.want)
			}
		})
	}
}

func TestNewConfigTemplateAliasRejectsUnsafeAliases(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{name: "empty", raw: "", want: "config template alias is required"},
		{name: "whitespace only", raw: "   ", want: "config template alias is required"},
		{name: "forward slash", raw: "nested/template", want: "config template alias must be a single path segment"},
		{name: "backslash", raw: `nested\template`, want: "config template alias must be a single path segment"},
		{name: "absolute path", raw: "/absolute", want: "config template alias must be a single path segment"},
		{name: "windows absolute path", raw: `C:\absolute`, want: "config template alias must be a single path segment"},
		{name: "path traversal", raw: "../escape", want: "config template alias must be a single path segment"},
		{name: "single dot", raw: ".", want: "config template alias must not contain '.' or '..' path sequences"},
		{name: "double dot", raw: "..", want: "config template alias must not contain '.' or '..' path sequences"},
		{name: "embedded dot dot", raw: "a..b", want: "config template alias must not contain '.' or '..' path sequences"},
		{name: "leading dot", raw: ".hidden", want: "config template alias must not start with '.'"},
		{name: "leading dash", raw: "-flag", want: "config template alias must not start with '-'"},
		{name: "space", raw: "api feature", want: `config template alias contains unsupported character ' '`},
		{name: "colon", raw: "api:feature", want: `config template alias contains unsupported character ':'`},
		{name: "shell metacharacter", raw: "api;feature", want: `config template alias contains unsupported character ';'`},
		{name: "template delimiter", raw: "api{{feature", want: `config template alias contains unsupported character '{'`},
		{name: "over length", raw: strings.Repeat("a", 129), want: "config template alias must be at most 128 characters"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewConfigTemplateAlias(test.raw)
			if err == nil {
				t.Fatalf("NewConfigTemplateAlias(%q) error = nil, want %q", test.raw, test.want)
			}
			if err.Error() != test.want {
				t.Fatalf("NewConfigTemplateAlias(%q) error = %q, want %q", test.raw, err.Error(), test.want)
			}
		})
	}
}

func TestParseConfigTemplateSourceKind(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want ConfigTemplateSourceKind
	}{
		{name: "builtin", raw: "builtin", want: ConfigTemplateSourceBuiltin},
		{name: "custom", raw: "custom", want: ConfigTemplateSourceCustom},
		{name: "trimmed", raw: " builtin ", want: ConfigTemplateSourceBuiltin},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := ParseConfigTemplateSourceKind(test.raw)
			if err != nil {
				t.Fatalf("ParseConfigTemplateSourceKind(%q) error = %v", test.raw, err)
			}
			if got != test.want {
				t.Fatalf("ParseConfigTemplateSourceKind(%q) = %q, want %q", test.raw, got, test.want)
			}
		})
	}
}

func TestParseConfigTemplateSourceKindRejectsUnsupportedValues(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{name: "empty", raw: "", want: "config template source is required"},
		{name: "remote", raw: "remote", want: "unsupported config template source: remote"},
		{name: "local", raw: "local", want: "unsupported config template source: local"},
		{name: "url like", raw: "https://example.invalid/template", want: "unsupported config template source: https://example.invalid/template"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := ParseConfigTemplateSourceKind(test.raw)
			if err == nil {
				t.Fatalf("ParseConfigTemplateSourceKind(%q) error = nil, want %q", test.raw, test.want)
			}
			if err.Error() != test.want {
				t.Fatalf("ParseConfigTemplateSourceKind(%q) error = %q, want %q", test.raw, err.Error(), test.want)
			}
		})
	}
}

func TestNewConfigTemplateReferenceValidatesBuiltinReferences(t *testing.T) {
	alias := mustConfigTemplateAlias(t, "default-feature")

	reference, err := NewConfigTemplateReference(ConfigTemplateReferenceInput{
		Alias:    alias,
		Source:   "builtin",
		Template: "feature",
	})
	if err != nil {
		t.Fatalf("NewConfigTemplateReference() error = %v", err)
	}

	if reference.Alias().String() != "default-feature" {
		t.Fatalf("Alias = %q, want default-feature", reference.Alias())
	}
	if reference.SourceKind() != ConfigTemplateSourceBuiltin {
		t.Fatalf("SourceKind = %q, want builtin", reference.SourceKind())
	}
	if reference.Template() != "feature" {
		t.Fatalf("Template = %q, want feature", reference.Template())
	}
	if name, ok := reference.BuiltInTemplateName(); !ok || name != FeatureTemplate {
		t.Fatalf("BuiltInTemplateName() = %q, %v, want feature, true", name, ok)
	}
	if _, ok := reference.CustomTemplateName(); ok {
		t.Fatalf("CustomTemplateName() ok = true, want false")
	}
}

func TestNewConfigTemplateReferenceValidatesCustomReferences(t *testing.T) {
	alias := mustConfigTemplateAlias(t, "api-feature")

	reference, err := NewConfigTemplateReference(ConfigTemplateReferenceInput{
		Alias:    alias,
		Source:   "custom",
		Template: "api-feature",
	})
	if err != nil {
		t.Fatalf("NewConfigTemplateReference() error = %v", err)
	}

	if reference.SourceKind() != ConfigTemplateSourceCustom {
		t.Fatalf("SourceKind = %q, want custom", reference.SourceKind())
	}
	if name, ok := reference.CustomTemplateName(); !ok || name.String() != "api-feature" {
		t.Fatalf("CustomTemplateName() = %q, %v, want api-feature, true", name.String(), ok)
	}
	if _, ok := reference.BuiltInTemplateName(); ok {
		t.Fatalf("BuiltInTemplateName() ok = true, want false")
	}
}

func TestNewConfigTemplateReferenceRejectsInvalidReferences(t *testing.T) {
	alias := mustConfigTemplateAlias(t, "api-feature")
	tests := []struct {
		name  string
		input ConfigTemplateReferenceInput
		want  string
	}{
		{
			name:  "missing alias",
			input: ConfigTemplateReferenceInput{Source: "builtin", Template: "feature"},
			want:  "config template alias is required",
		},
		{
			name:  "missing source",
			input: ConfigTemplateReferenceInput{Alias: alias, Template: "feature"},
			want:  "config template source is required",
		},
		{
			name:  "missing template",
			input: ConfigTemplateReferenceInput{Alias: alias, Source: "builtin"},
			want:  "config template reference template is required",
		},
		{
			name:  "unknown builtin",
			input: ConfigTemplateReferenceInput{Alias: alias, Source: "builtin", Template: "maintenance"},
			want:  "unknown template name: maintenance",
		},
		{
			name:  "unsafe custom",
			input: ConfigTemplateReferenceInput{Alias: alias, Source: "custom", Template: "../escape"},
			want:  "custom template name must be a single path segment",
		},
		{
			name:  "unsupported source",
			input: ConfigTemplateReferenceInput{Alias: alias, Source: "remote", Template: "feature"},
			want:  "unsupported config template source: remote",
		},
		{
			name: "unsupported path field",
			input: ConfigTemplateReferenceInput{
				Alias:             alias,
				Source:            "builtin",
				Template:          "feature",
				UnsupportedFields: []string{"path"},
			},
			want: `unsupported config template field "path"`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewConfigTemplateReference(test.input)
			if err == nil {
				t.Fatalf("NewConfigTemplateReference() error = nil, want %q", test.want)
			}
			if err.Error() != test.want {
				t.Fatalf("NewConfigTemplateReference() error = %q, want %q", err.Error(), test.want)
			}
		})
	}
}

func TestConfigTemplateAliasesLookupAndCopying(t *testing.T) {
	alias := mustConfigTemplateAlias(t, "api-feature")
	reference, err := NewConfigTemplateReference(ConfigTemplateReferenceInput{
		Alias:    alias,
		Source:   "custom",
		Template: "api-feature",
	})
	if err != nil {
		t.Fatalf("NewConfigTemplateReference() error = %v", err)
	}

	aliases := NewConfigTemplateAliases([]ConfigTemplateReference{reference})
	got, err := aliases.Lookup(alias)
	if err != nil {
		t.Fatalf("Lookup() error = %v", err)
	}
	if got.Template() != "api-feature" {
		t.Fatalf("Lookup().Template = %q, want api-feature", got.Template())
	}

	copied := aliases.References()
	delete(copied, "api-feature")
	if aliases.Len() != 1 {
		t.Fatalf("mutating References() changed alias set length to %d", aliases.Len())
	}

	copiedAliases := aliases.Copy()
	copiedReferences := copiedAliases.References()
	delete(copiedReferences, "api-feature")
	if aliases.Len() != 1 {
		t.Fatalf("mutating copied alias references changed original length to %d", aliases.Len())
	}
}

func TestConfigTemplateAliasesLookupRejectsMissingAlias(t *testing.T) {
	alias := mustConfigTemplateAlias(t, "missing")

	_, err := EmptyConfigTemplateAliases().Lookup(alias)
	if err == nil {
		t.Fatalf("Lookup() error = nil, want missing alias error")
	}
	if err.Error() != "config template alias not found: missing" {
		t.Fatalf("Lookup() error = %q, want missing alias error", err.Error())
	}
}

func mustConfigTemplateAlias(t *testing.T, value string) ConfigTemplateAlias {
	t.Helper()

	alias, err := NewConfigTemplateAlias(value)
	if err != nil {
		t.Fatalf("NewConfigTemplateAlias(%q) error = %v", value, err)
	}
	return alias
}
