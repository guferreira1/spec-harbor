package domain

import "testing"

func TestSupportedLocalConfigVersionIsOne(t *testing.T) {
	if SupportedLocalConfigVersion != 1 {
		t.Fatalf("SupportedLocalConfigVersion = %d, want 1", SupportedLocalConfigVersion)
	}
}

func TestIsSupportedLocalConfigVersion(t *testing.T) {
	tests := []struct {
		name    string
		version int
		want    bool
	}{
		{name: "supported", version: 1, want: true},
		{name: "missing", version: 0, want: false},
		{name: "zero", version: 0, want: false},
		{name: "negative", version: -1, want: false},
		{name: "future", version: 2, want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := IsSupportedLocalConfigVersion(test.version); got != test.want {
				t.Fatalf("IsSupportedLocalConfigVersion(%d) = %v, want %v", test.version, got, test.want)
			}
		})
	}
}

func TestConfigResultStoresLocalConfigValues(t *testing.T) {
	config := LocalConfig{
		Version: 1,
		Defaults: ConfigDefaults{
			AgentRole:      "implementer",
			GenerationMode: "blank",
		},
		Validation: ConfigValidation{
			RequireAllChangeFiles: true,
		},
		Review: ConfigReview{
			RequireCompletedTasks: true,
		},
		Archive: ConfigArchive{
			DateLayout: "2006-01-02",
		},
		Scan: ConfigScan{
			IncludeCommonProjectFiles: true,
		},
		Output: ConfigOutput{
			Format: "text",
		},
	}
	result := ConfigResult{
		Path:   ".specharbor/config.yml",
		Config: config,
	}

	if result.Path != ".specharbor/config.yml" {
		t.Fatalf("Path = %q, want .specharbor/config.yml", result.Path)
	}
	if result.Config.Version != 1 {
		t.Fatalf("Version = %d, want 1", result.Config.Version)
	}
	if result.Config.Defaults.AgentRole != "implementer" {
		t.Fatalf("Defaults.AgentRole = %q, want implementer", result.Config.Defaults.AgentRole)
	}
	if result.Config.Defaults.GenerationMode != "blank" {
		t.Fatalf("Defaults.GenerationMode = %q, want blank", result.Config.Defaults.GenerationMode)
	}
	if !result.Config.Validation.RequireAllChangeFiles {
		t.Fatalf("Validation.RequireAllChangeFiles = false, want true")
	}
	if !result.Config.Review.RequireCompletedTasks {
		t.Fatalf("Review.RequireCompletedTasks = false, want true")
	}
	if result.Config.Archive.DateLayout != "2006-01-02" {
		t.Fatalf("Archive.DateLayout = %q, want 2006-01-02", result.Config.Archive.DateLayout)
	}
	if !result.Config.Scan.IncludeCommonProjectFiles {
		t.Fatalf("Scan.IncludeCommonProjectFiles = false, want true")
	}
	if result.Config.Output.Format != "text" {
		t.Fatalf("Output.Format = %q, want text", result.Config.Output.Format)
	}
}
