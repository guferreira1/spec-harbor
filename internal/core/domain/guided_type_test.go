package domain

import "testing"

func TestParseGuidedTypeAcceptsSupportedTypes(t *testing.T) {
	tests := []struct {
		value string
		want  GuidedType
	}{
		{value: "feature", want: FeatureGuidedType},
		{value: "bugfix", want: BugfixGuidedType},
		{value: "docs", want: DocsGuidedType},
		{value: "refactor", want: RefactorGuidedType},
		{value: " feature ", want: FeatureGuidedType},
	}

	for _, test := range tests {
		t.Run(test.value, func(t *testing.T) {
			got, err := ParseGuidedType(test.value)
			if err != nil {
				t.Fatalf("ParseGuidedType(%q) error = %v", test.value, err)
			}
			if got != test.want {
				t.Fatalf("ParseGuidedType(%q) = %q, want %q", test.value, got, test.want)
			}
		})
	}
}

func TestParseGuidedTypeRejectsEmptyAndUnknownValues(t *testing.T) {
	tests := []struct {
		value string
		want  string
	}{
		{value: "", want: "guided type is required"},
		{value: " ", want: "guided type is required"},
		{value: "maintenance", want: "unknown guided type: maintenance"},
		{value: "Feature", want: "unknown guided type: Feature"},
	}

	for _, test := range tests {
		t.Run(test.value, func(t *testing.T) {
			_, err := ParseGuidedType(test.value)
			if err == nil {
				t.Fatalf("ParseGuidedType(%q) error = nil, want %q", test.value, test.want)
			}
			if err.Error() != test.want {
				t.Fatalf("ParseGuidedType(%q) error = %q, want %q", test.value, err.Error(), test.want)
			}
		})
	}
}

func TestSupportedGuidedTypesReturnsSupportedTypesInDeterministicOrder(t *testing.T) {
	got := SupportedGuidedTypes()
	want := []GuidedType{FeatureGuidedType, BugfixGuidedType, DocsGuidedType, RefactorGuidedType}

	if len(got) != len(want) {
		t.Fatalf("SupportedGuidedTypes() = %v, want %v", got, want)
	}
	for index := range got {
		if got[index] != want[index] {
			t.Fatalf("SupportedGuidedTypes() = %v, want %v", got, want)
		}
	}

	got[0] = GuidedType("mutated")
	if SupportedGuidedTypes()[0] != FeatureGuidedType {
		t.Fatalf("SupportedGuidedTypes() returned mutable policy")
	}
}
