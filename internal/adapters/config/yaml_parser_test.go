package config

import (
	"reflect"
	"strings"
	"testing"

	"github.com/guferreira1/spec-harbor/internal/core/domain"
)

func TestYAMLParserParsesCompleteVersionOneConfig(t *testing.T) {
	config, err := NewYAMLParser().ParseLocalConfig(completeVersionOneConfigYAML())
	if err != nil {
		t.Fatalf("ParseLocalConfig() error = %v", err)
	}

	want := domain.LocalConfig{
		Version: 1,
		Defaults: domain.ConfigDefaults{
			AgentRole:      "implementer",
			GenerationMode: "blank",
		},
		Validation: domain.ConfigValidation{
			RequireAllChangeFiles: true,
		},
		Review: domain.ConfigReview{
			RequireCompletedTasks: true,
		},
		Archive: domain.ConfigArchive{
			DateLayout: "2006-01-02",
		},
		Scan: domain.ConfigScan{
			IncludeCommonProjectFiles: true,
		},
		Output: domain.ConfigOutput{
			Format: "text",
		},
		Templates: domain.NewConfigTemplates(domain.EmptyConfigTemplateAliases()),
	}

	if !reflect.DeepEqual(config, want) {
		t.Fatalf("ParseLocalConfig() = %#v, want %#v", config, want)
	}
}

func TestYAMLParserReturnsErrorForInvalidYAML(t *testing.T) {
	_, err := NewYAMLParser().ParseLocalConfig("version: [\n")
	if err == nil {
		t.Fatalf("ParseLocalConfig() error = nil, want parse error")
	}
	if !strings.Contains(err.Error(), "parse local config YAML") {
		t.Fatalf("ParseLocalConfig() error = %q, want parse local config YAML", err.Error())
	}
}

func TestYAMLParserReturnsErrorForIncompatibleScalarTypes(t *testing.T) {
	contents := `version: 1
validation:
  require_all_change_files: not-a-bool
`

	_, err := NewYAMLParser().ParseLocalConfig(contents)
	if err == nil {
		t.Fatalf("ParseLocalConfig() error = nil, want decode error")
	}
	if !strings.Contains(err.Error(), "parse local config YAML") {
		t.Fatalf("ParseLocalConfig() error = %q, want parse local config YAML", err.Error())
	}
}

func TestYAMLParserIgnoresUnknownKeys(t *testing.T) {
	contents := `version: 1
unknown_top_level: ignored

defaults:
  agent_role: implementer
  generation_mode: blank
  unknown_default: ignored

validation:
  require_all_change_files: true

review:
  require_completed_tasks: true

archive:
  date_layout: "2006-01-02"

scan:
  include_common_project_files: true

output:
  format: text

future:
  nested: ignored
`

	config, err := NewYAMLParser().ParseLocalConfig(contents)
	if err != nil {
		t.Fatalf("ParseLocalConfig() error = %v", err)
	}
	if config.Version != 1 {
		t.Fatalf("Version = %d, want 1", config.Version)
	}
	if config.Defaults.AgentRole != "implementer" {
		t.Fatalf("Defaults.AgentRole = %q, want implementer", config.Defaults.AgentRole)
	}
	if config.Output.Format != "text" {
		t.Fatalf("Output.Format = %q, want text", config.Output.Format)
	}
}

func TestYAMLParserUsesZeroValuesForOmittedOptionalSectionsAndFields(t *testing.T) {
	config, err := NewYAMLParser().ParseLocalConfig("version: 1\n")
	if err != nil {
		t.Fatalf("ParseLocalConfig() error = %v", err)
	}

	if config.Version != 1 {
		t.Fatalf("Version = %d, want 1", config.Version)
	}
	if config.Defaults != (domain.ConfigDefaults{}) {
		t.Fatalf("Defaults = %#v, want zero value", config.Defaults)
	}
	if config.Validation != (domain.ConfigValidation{}) {
		t.Fatalf("Validation = %#v, want zero value", config.Validation)
	}
	if config.Review != (domain.ConfigReview{}) {
		t.Fatalf("Review = %#v, want zero value", config.Review)
	}
	if config.Archive != (domain.ConfigArchive{}) {
		t.Fatalf("Archive = %#v, want zero value", config.Archive)
	}
	if config.Scan != (domain.ConfigScan{}) {
		t.Fatalf("Scan = %#v, want zero value", config.Scan)
	}
	if config.Output != (domain.ConfigOutput{}) {
		t.Fatalf("Output = %#v, want zero value", config.Output)
	}
	if config.Templates.Aliases().Len() != 0 {
		t.Fatalf("Template aliases = %d, want 0", config.Templates.Aliases().Len())
	}
}

func TestYAMLParserDoesNotDefaultMissingVersion(t *testing.T) {
	config, err := NewYAMLParser().ParseLocalConfig(`defaults:
  agent_role: implementer
`)
	if err != nil {
		t.Fatalf("ParseLocalConfig() error = %v", err)
	}

	if config.Version != 0 {
		t.Fatalf("Version = %d, want 0", config.Version)
	}
	if config.Defaults.AgentRole != "implementer" {
		t.Fatalf("Defaults.AgentRole = %q, want implementer", config.Defaults.AgentRole)
	}
}

func TestYAMLParserParsesTemplateAliases(t *testing.T) {
	contents := `version: 1

templates:
  aliases:
    api-feature:
      source: custom
      template: api-feature
    default-feature:
      source: builtin
      template: feature
    remote-feature:
      source: remote
      url: https://example.com/templates/feature.zip
      checksum: sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef
      format: zip
`

	config, err := NewYAMLParser().ParseLocalConfig(contents)
	if err != nil {
		t.Fatalf("ParseLocalConfig() error = %v", err)
	}

	aliases := config.Templates.Aliases()
	if aliases.Len() != 3 {
		t.Fatalf("alias count = %d, want 3", aliases.Len())
	}

	customReference, err := aliases.Lookup(mustConfigTemplateAlias(t, "api-feature"))
	if err != nil {
		t.Fatalf("Lookup(api-feature) error = %v", err)
	}
	if customReference.SourceKind() != domain.ConfigTemplateSourceCustom || customReference.Template() != "api-feature" {
		t.Fatalf("custom reference = %#v, want custom api-feature", customReference)
	}

	builtInReference, err := aliases.Lookup(mustConfigTemplateAlias(t, "default-feature"))
	if err != nil {
		t.Fatalf("Lookup(default-feature) error = %v", err)
	}
	if builtInReference.SourceKind() != domain.ConfigTemplateSourceBuiltin || builtInReference.Template() != "feature" {
		t.Fatalf("builtin reference = %#v, want builtin feature", builtInReference)
	}

	remoteReference, err := aliases.Lookup(mustConfigTemplateAlias(t, "remote-feature"))
	if err != nil {
		t.Fatalf("Lookup(remote-feature) error = %v", err)
	}
	if remoteReference.SourceKind() != domain.ConfigTemplateSourceRemote {
		t.Fatalf("remote source = %q, want remote", remoteReference.SourceKind())
	}
	resolvedRemote, ok := remoteReference.RemoteTemplateReference()
	if !ok {
		t.Fatalf("RemoteTemplateReference() ok = false, want true")
	}
	if resolvedRemote.URL().Host() != "example.com" || resolvedRemote.Format() != domain.RemoteTemplateFormatZip {
		t.Fatalf("remote reference = %#v, want example.com zip", resolvedRemote)
	}
}

func TestYAMLParserTreatsOmittedTemplatesAsEmptyAliases(t *testing.T) {
	tests := []struct {
		name     string
		contents string
	}{
		{name: "omitted templates", contents: "version: 1\n"},
		{name: "omitted aliases", contents: "version: 1\ntemplates: {}\n"},
		{name: "null aliases", contents: "version: 1\ntemplates:\n  aliases:\n"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config, err := NewYAMLParser().ParseLocalConfig(test.contents)
			if err != nil {
				t.Fatalf("ParseLocalConfig() error = %v", err)
			}
			if config.Templates.Aliases().Len() != 0 {
				t.Fatalf("alias count = %d, want 0", config.Templates.Aliases().Len())
			}
		})
	}
}

func TestYAMLParserRejectsInvalidTemplateAliasConfig(t *testing.T) {
	tests := []struct {
		name     string
		contents string
		want     string
	}{
		{
			name: "templates not mapping",
			contents: `version: 1
templates: []
`,
			want: "templates must be a mapping",
		},
		{
			name: "aliases not mapping",
			contents: `version: 1
templates:
  aliases: []
`,
			want: "templates.aliases must be a mapping",
		},
		{
			name: "invalid alias name",
			contents: `version: 1
templates:
  aliases:
    nested/template:
      source: builtin
      template: feature
`,
			want: `invalid config template alias "nested/template": config template alias must be a single path segment`,
		},
		{
			name: "alias entry not mapping",
			contents: `version: 1
templates:
  aliases:
    api-feature: true
`,
			want: `invalid config template alias "api-feature": config template alias entry must be a mapping`,
		},
		{
			name: "missing source",
			contents: `version: 1
templates:
  aliases:
    api-feature:
      template: feature
`,
			want: `invalid config template alias "api-feature": config template source is required`,
		},
		{
			name: "missing template",
			contents: `version: 1
templates:
  aliases:
    api-feature:
      source: builtin
`,
			want: `invalid config template alias "api-feature": config template reference template is required`,
		},
		{
			name: "missing custom template",
			contents: `version: 1
templates:
  aliases:
    api-feature:
      source: custom
`,
			want: `invalid config template alias "api-feature": config template reference template is required`,
		},
		{
			name: "unsupported source",
			contents: `version: 1
templates:
  aliases:
    api-feature:
      source: local
      template: feature
`,
			want: `invalid config template alias "api-feature": unsupported config template source: local`,
		},
		{
			name: "remote missing url",
			contents: `version: 1
templates:
  aliases:
    api-feature:
      source: remote
      checksum: sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef
      format: zip
`,
			want: `invalid config template alias "api-feature": remote template URL is required`,
		},
		{
			name: "remote missing checksum",
			contents: `version: 1
templates:
  aliases:
    api-feature:
      source: remote
      url: https://example.com/template.zip
      format: zip
`,
			want: `invalid config template alias "api-feature": remote template checksum is required`,
		},
		{
			name: "remote missing format",
			contents: `version: 1
templates:
  aliases:
    api-feature:
      source: remote
      url: https://example.com/template.zip
      checksum: sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef
`,
			want: `invalid config template alias "api-feature": remote template format is required`,
		},
		{
			name: "remote unsupported format",
			contents: `version: 1
templates:
  aliases:
    api-feature:
      source: remote
      url: https://example.com/template.zip
      checksum: sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef
      format: tar
`,
			want: `invalid config template alias "api-feature": unsupported remote template format: tar`,
		},
		{
			name: "remote rejects template",
			contents: `version: 1
templates:
  aliases:
    api-feature:
      source: remote
      template: feature
      url: https://example.com/template.zip
      checksum: sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef
      format: zip
`,
			want: `invalid config template alias "api-feature": unsupported config template field "template"`,
		},
		{
			name: "remote rejects unknown field",
			contents: `version: 1
templates:
  aliases:
    api-feature:
      source: remote
      url: https://example.com/template.zip
      checksum: sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef
      format: zip
      headers:
        Authorization: secret
`,
			want: `invalid config template alias "api-feature": unsupported config template field "headers"`,
		},
		{
			name: "unsupported path field",
			contents: `version: 1
templates:
  aliases:
    api-feature:
      source: builtin
      template: feature
      path: ../templates/feature
`,
			want: `invalid config template alias "api-feature": unsupported config template field "path"`,
		},
		{
			name: "unsupported url field",
			contents: `version: 1
templates:
  aliases:
    api-feature:
      source: builtin
      template: feature
      url: https://example.invalid/template
`,
			want: `invalid config template alias "api-feature": unsupported config template field "url"`,
		},
		{
			name: "unsupported checksum field on builtin",
			contents: `version: 1
templates:
  aliases:
    api-feature:
      source: builtin
      template: feature
      checksum: sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef
`,
			want: `invalid config template alias "api-feature": unsupported config template field "checksum"`,
		},
		{
			name: "unsupported format field on builtin",
			contents: `version: 1
templates:
  aliases:
    api-feature:
      source: builtin
      template: feature
      format: zip
`,
			want: `invalid config template alias "api-feature": unsupported config template field "format"`,
		},
		{
			name: "unsupported url field on custom",
			contents: `version: 1
templates:
  aliases:
    api-feature:
      source: custom
      template: api-feature
      url: https://example.invalid/template
`,
			want: `invalid config template alias "api-feature": unsupported config template field "url"`,
		},
		{
			name: "unsupported checksum field on custom",
			contents: `version: 1
templates:
  aliases:
    api-feature:
      source: custom
      template: api-feature
      checksum: sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef
`,
			want: `invalid config template alias "api-feature": unsupported config template field "checksum"`,
		},
		{
			name: "unsupported format field on custom",
			contents: `version: 1
templates:
  aliases:
    api-feature:
      source: custom
      template: api-feature
      format: zip
`,
			want: `invalid config template alias "api-feature": unsupported config template field "format"`,
		},
		{
			name: "source not string",
			contents: `version: 1
templates:
  aliases:
    api-feature:
      source: 1
      template: feature
`,
			want: `invalid config template alias "api-feature": config template source must be a string`,
		},
		{
			name: "template not string",
			contents: `version: 1
templates:
  aliases:
    api-feature:
      source: builtin
      template: 1
`,
			want: `invalid config template alias "api-feature": config template template must be a string`,
		},
		{
			name: "unknown builtin template",
			contents: `version: 1
templates:
  aliases:
    api-feature:
      source: builtin
      template: maintenance
`,
			want: `invalid config template alias "api-feature": unknown template name: maintenance`,
		},
		{
			name: "invalid custom template reference",
			contents: `version: 1
templates:
  aliases:
    api-feature:
      source: custom
      template: ../escape
`,
			want: `invalid config template alias "api-feature": custom template name must be a single path segment`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewYAMLParser().ParseLocalConfig(test.contents)
			if err == nil {
				t.Fatalf("ParseLocalConfig() error = nil, want %q", test.want)
			}
			if err.Error() != test.want {
				t.Fatalf("ParseLocalConfig() error = %q, want %q", err.Error(), test.want)
			}
		})
	}
}

func TestYAMLParserRejectsUnsupportedTemplateAliasFieldsBySource(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		template string
		fields   []string
	}{
		{
			name:     "builtin",
			source:   "builtin",
			template: "feature",
			fields:   []string{"path", "headers", "auth", "branch", "tag", "ref", "command", "script"},
		},
		{
			name:     "custom",
			source:   "custom",
			template: "api-feature",
			fields:   []string{"path", "headers", "auth", "branch", "tag", "ref", "command", "script"},
		},
		{
			name:     "remote",
			source:   "remote",
			template: "",
			fields:   []string{"path", "headers", "auth", "branch", "tag", "ref", "command", "script"},
		},
	}

	for _, test := range tests {
		for _, field := range test.fields {
			t.Run(test.name+" "+field, func(t *testing.T) {
				contents := configTemplateAliasYAML(test.source, test.template, field+": unsafe\n")
				_, err := NewYAMLParser().ParseLocalConfig(contents)
				if err == nil {
					t.Fatalf("ParseLocalConfig() error = nil, want unsupported field %q", field)
				}
				want := `invalid config template alias "api-feature": unsupported config template field "` + field + `"`
				if err.Error() != want {
					t.Fatalf("ParseLocalConfig() error = %q, want %q", err.Error(), want)
				}
			})
		}
	}
}

func TestYAMLParserRejectsRemoteOnlyFieldsForBuiltinAndCustomAliases(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		template string
	}{
		{name: "builtin", source: "builtin", template: "feature"},
		{name: "custom", source: "custom", template: "api-feature"},
	}
	remoteOnlyFields := []struct {
		field string
		line  string
	}{
		{field: "url", line: "url: https://example.invalid/template.zip\n"},
		{field: "checksum", line: "checksum: sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef\n"},
		{field: "format", line: "format: zip\n"},
	}

	for _, test := range tests {
		for _, remoteOnlyField := range remoteOnlyFields {
			t.Run(test.name+" "+remoteOnlyField.field, func(t *testing.T) {
				contents := configTemplateAliasYAML(test.source, test.template, remoteOnlyField.line)
				_, err := NewYAMLParser().ParseLocalConfig(contents)
				if err == nil {
					t.Fatalf("ParseLocalConfig() error = nil, want unsupported field %q", remoteOnlyField.field)
				}
				want := `invalid config template alias "api-feature": unsupported config template field "` + remoteOnlyField.field + `"`
				if err.Error() != want {
					t.Fatalf("ParseLocalConfig() error = %q, want %q", err.Error(), want)
				}
			})
		}
	}
}

func configTemplateAliasYAML(source string, template string, extra string) string {
	contents := `version: 1
templates:
  aliases:
    api-feature:
      source: ` + source + `
`
	if template != "" {
		contents += `      template: ` + template + `
`
	}
	if source == "remote" {
		contents += `      url: https://example.com/template.zip
      checksum: sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef
      format: zip
`
	}
	return contents + "      " + extra
}

func completeVersionOneConfigYAML() string {
	return `version: 1

defaults:
  agent_role: implementer
  generation_mode: blank

validation:
  require_all_change_files: true

review:
  require_completed_tasks: true

archive:
  date_layout: "2006-01-02"

scan:
  include_common_project_files: true

output:
  format: text
`
}

func mustConfigTemplateAlias(t *testing.T, value string) domain.ConfigTemplateAlias {
	t.Helper()

	alias, err := domain.NewConfigTemplateAlias(value)
	if err != nil {
		t.Fatalf("NewConfigTemplateAlias(%q) error = %v", value, err)
	}
	return alias
}
