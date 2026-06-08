package config

import (
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
	}

	if config != want {
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
