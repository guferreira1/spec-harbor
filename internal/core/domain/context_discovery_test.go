package domain

import (
	"strings"
	"testing"
)

func TestNewContextSignalValidatesRequiredFields(t *testing.T) {
	source := ContextSource{Path: "go.mod", Category: ContextSourceCategoryPackageManifest}

	signal, err := NewContextSignal(ContextSignalInput{
		Kind:           ContextSignalKindLanguage,
		Value:          " Go ",
		Classification: ContextSignalClassificationDetectedFact,
		Confidence:     ContextConfidenceHigh,
		Source:         source,
	})
	if err != nil {
		t.Fatalf("NewContextSignal() error = %v", err)
	}
	if signal.Value != "Go" {
		t.Fatalf("Value = %q, want trimmed Go", signal.Value)
	}
	if signal.Classification != ContextSignalClassificationDetectedFact {
		t.Fatalf("Classification = %q, want detected fact", signal.Classification)
	}
	if signal.Confidence != ContextConfidenceHigh {
		t.Fatalf("Confidence = %q, want high", signal.Confidence)
	}

	tests := []struct {
		name  string
		input ContextSignalInput
		want  string
	}{
		{
			name: "unsupported kind",
			input: ContextSignalInput{
				Kind:           ContextSignalKind("unknown"),
				Value:          "Go",
				Classification: ContextSignalClassificationDetectedFact,
				Confidence:     ContextConfidenceHigh,
				Source:         source,
			},
			want: "unsupported context signal kind",
		},
		{
			name: "missing value",
			input: ContextSignalInput{
				Kind:           ContextSignalKindLanguage,
				Value:          " ",
				Classification: ContextSignalClassificationDetectedFact,
				Confidence:     ContextConfidenceHigh,
				Source:         source,
			},
			want: "context signal value is required",
		},
		{
			name: "unsupported classification",
			input: ContextSignalInput{
				Kind:           ContextSignalKindLanguage,
				Value:          "Go",
				Classification: ContextSignalClassification("fact"),
				Confidence:     ContextConfidenceHigh,
				Source:         source,
			},
			want: "unsupported context signal classification",
		},
		{
			name: "unsupported confidence",
			input: ContextSignalInput{
				Kind:           ContextSignalKindLanguage,
				Value:          "Go",
				Classification: ContextSignalClassificationDetectedFact,
				Confidence:     ContextConfidence("certain"),
				Source:         source,
			},
			want: "unsupported context confidence",
		},
		{
			name: "missing source path",
			input: ContextSignalInput{
				Kind:           ContextSignalKindLanguage,
				Value:          "Go",
				Classification: ContextSignalClassificationDetectedFact,
				Confidence:     ContextConfidenceHigh,
				Source:         ContextSource{Category: ContextSourceCategoryPackageManifest},
			},
			want: "context source path is required",
		},
		{
			name: "unsafe source path",
			input: ContextSignalInput{
				Kind:           ContextSignalKindLanguage,
				Value:          "Go",
				Classification: ContextSignalClassificationDetectedFact,
				Confidence:     ContextConfidenceHigh,
				Source:         ContextSource{Path: "../go.mod", Category: ContextSourceCategoryPackageManifest},
			},
			want: "context source path must be a safe relative path",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewContextSignal(test.input)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("NewContextSignal() error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestContextClassificationsAndConfidenceLevelsAreSupported(t *testing.T) {
	for _, classification := range []ContextSignalClassification{
		ContextSignalClassificationDetectedFact,
		ContextSignalClassificationSuggestedAssumption,
		ContextSignalClassificationUserConfirmedContext,
	} {
		if !classification.IsSupported() {
			t.Fatalf("classification %q should be supported", classification)
		}
	}
	if ContextSignalClassification("detected").IsSupported() {
		t.Fatalf("unknown classification should not be supported")
	}

	for _, confidence := range []ContextConfidence{
		ContextConfidenceHigh,
		ContextConfidenceMedium,
		ContextConfidenceLow,
	} {
		if !confidence.IsSupported() {
			t.Fatalf("confidence %q should be supported", confidence)
		}
	}
	if ContextConfidence("maximum").IsSupported() {
		t.Fatalf("unknown confidence should not be supported")
	}
}

func TestContextDiscoveryResultOrdersAndGroupsSignalsDeterministically(t *testing.T) {
	packageSource := ContextSource{Path: "package.json", Category: ContextSourceCategoryPackageManifest}
	goSource := ContextSource{Path: "go.mod", Category: ContextSourceCategoryPackageManifest}
	briefSource := ContextSource{Path: ".specharbor/project-brief.md", Category: ContextSourceCategoryProjectBrief}

	signals := []ContextSignal{
		mustContextSignal(t, ContextSignalKindTestCommand, "go test ./...", ContextSignalClassificationSuggestedAssumption, ContextConfidenceMedium, goSource),
		mustContextSignal(t, ContextSignalKindLanguage, "Node.js", ContextSignalClassificationDetectedFact, ContextConfidenceHigh, packageSource),
		mustContextSignal(t, ContextSignalKindStack, "Go", ContextSignalClassificationUserConfirmedContext, ContextConfidenceHigh, briefSource),
		mustContextSignal(t, ContextSignalKindLanguage, "Go", ContextSignalClassificationDetectedFact, ContextConfidenceHigh, goSource),
		mustContextSignal(t, ContextSignalKindLanguage, "Go", ContextSignalClassificationDetectedFact, ContextConfidenceHigh, goSource),
	}

	result := NewContextDiscoveryResult(signals, []ContextDiscoveryNote{{Message: "Multiple stack signals detected."}})

	all := result.Signals()
	want := []ContextSignalKind{
		ContextSignalKindLanguage,
		ContextSignalKindLanguage,
		ContextSignalKindStack,
		ContextSignalKindTestCommand,
	}
	if len(all) != len(want) {
		t.Fatalf("signal count = %d, want %d (%+v)", len(all), len(want), all)
	}
	for index, kind := range want {
		if all[index].Kind != kind {
			t.Fatalf("signal %d kind = %q, want %q (signals=%+v)", index, all[index].Kind, kind, all)
		}
	}

	confirmed := result.SignalsByClassification(ContextSignalClassificationUserConfirmedContext)
	if len(confirmed) != 1 || confirmed[0].Value != "Go" {
		t.Fatalf("confirmed signals = %+v, want confirmed Go stack", confirmed)
	}
	assumptions := result.SignalsByClassification(ContextSignalClassificationSuggestedAssumption)
	if len(assumptions) != 1 || assumptions[0].Classification != ContextSignalClassificationSuggestedAssumption {
		t.Fatalf("assumptions = %+v, want one suggested assumption", assumptions)
	}
	if got := result.Notes(); len(got) != 1 || got[0].Message != "Multiple stack signals detected." {
		t.Fatalf("Notes = %+v, want ambiguity note", got)
	}

	all[0].Value = "mutated"
	if result.Signals()[0].Value == "mutated" {
		t.Fatalf("Signals() returned mutable backing slice")
	}
}

func TestContextDiscoverySkipPolicyCoversSensitiveFilesAndGeneratedFolders(t *testing.T) {
	for _, path := range []string{
		".env",
		".env.local",
		"docs/private.pem",
		"docs/private.key",
		"id_rsa",
		"id_ed25519",
		"secrets.yml",
		"credentials.json",
		"docs/secrets.local",
	} {
		t.Run(path, func(t *testing.T) {
			if !ShouldSkipContextDiscoveryPath(path) {
				t.Fatalf("ShouldSkipContextDiscoveryPath(%q) = false, want true", path)
			}
		})
	}

	for _, path := range []string{
		".git/config",
		"node_modules/package/README.md",
		"dist/app.js",
		"build/output.txt",
		"target/debug/app",
		"vendor/module/file.go",
		"coverage/report.html",
		".tmp/file",
		".cache/file",
		".next/server",
		".nuxt/output",
		"out/app",
		"bin/app",
		"obj/debug",
	} {
		t.Run(path, func(t *testing.T) {
			if !ShouldSkipContextDiscoveryPath(path) {
				t.Fatalf("ShouldSkipContextDiscoveryPath(%q) = false, want true", path)
			}
		})
	}

	for _, path := range []string{"README.md", "docs/usage.md", "openspec/project.md"} {
		if ShouldSkipContextDiscoveryPath(path) {
			t.Fatalf("ShouldSkipContextDiscoveryPath(%q) = true, want false", path)
		}
	}
}

func mustContextSignal(
	t *testing.T,
	kind ContextSignalKind,
	value string,
	classification ContextSignalClassification,
	confidence ContextConfidence,
	source ContextSource,
) ContextSignal {
	t.Helper()

	signal, err := NewContextSignal(ContextSignalInput{
		Kind:           kind,
		Value:          value,
		Classification: classification,
		Confidence:     confidence,
		Source:         source,
	})
	if err != nil {
		t.Fatalf("NewContextSignal() error = %v", err)
	}
	return signal
}
