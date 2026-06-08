package domain

import "testing"

func TestParseTemplateNameAcceptsSupportedBuiltIns(t *testing.T) {
	tests := []struct {
		value string
		want  TemplateName
	}{
		{value: "feature", want: FeatureTemplate},
		{value: "bugfix", want: BugfixTemplate},
		{value: "docs", want: DocsTemplate},
		{value: "refactor", want: RefactorTemplate},
		{value: " feature ", want: FeatureTemplate},
	}

	for _, test := range tests {
		t.Run(test.value, func(t *testing.T) {
			got, err := ParseTemplateName(test.value)
			if err != nil {
				t.Fatalf("ParseTemplateName(%q) error = %v", test.value, err)
			}
			if got != test.want {
				t.Fatalf("ParseTemplateName(%q) = %q, want %q", test.value, got, test.want)
			}
		})
	}
}

func TestParseTemplateNameRejectsEmptyAndUnknownValues(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{name: "empty", value: "", want: "template name is required"},
		{name: "blank", value: " ", want: "template name is required"},
		{name: "unknown", value: "maintenance", want: "unknown template name: maintenance"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := ParseTemplateName(test.value)
			if err == nil {
				t.Fatalf("ParseTemplateName(%q) error = nil, want %q", test.value, test.want)
			}
			if err.Error() != test.want {
				t.Fatalf("ParseTemplateName(%q) error = %q, want %q", test.value, err.Error(), test.want)
			}
		})
	}
}

func TestSupportedTemplateNamesReturnsSupportedBuiltInsInDeterministicOrder(t *testing.T) {
	got := SupportedTemplateNames()
	want := []TemplateName{FeatureTemplate, BugfixTemplate, DocsTemplate, RefactorTemplate}

	if len(got) != len(want) {
		t.Fatalf("SupportedTemplateNames() = %v, want %v", got, want)
	}
	for index := range got {
		if got[index] != want[index] {
			t.Fatalf("SupportedTemplateNames() = %v, want %v", got, want)
		}
	}

	got[0] = "mutated"
	if SupportedTemplateNames()[0] != FeatureTemplate {
		t.Fatalf("SupportedTemplateNames() returned mutable policy")
	}
}
