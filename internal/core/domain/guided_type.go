package domain

import (
	"fmt"
	"strings"
)

type GuidedType string

const (
	FeatureGuidedType  GuidedType = "feature"
	BugfixGuidedType   GuidedType = "bugfix"
	DocsGuidedType     GuidedType = "docs"
	RefactorGuidedType GuidedType = "refactor"
)

func ParseGuidedType(value string) (GuidedType, error) {
	guidedType := GuidedType(strings.TrimSpace(value))
	if guidedType == "" {
		return "", fmt.Errorf("guided type is required")
	}
	if !guidedType.IsSupported() {
		return "", fmt.Errorf("unknown guided type: %s", guidedType)
	}
	return guidedType, nil
}

func SupportedGuidedTypes() []GuidedType {
	return []GuidedType{
		FeatureGuidedType,
		BugfixGuidedType,
		DocsGuidedType,
		RefactorGuidedType,
	}
}

func (guidedType GuidedType) IsSupported() bool {
	switch guidedType {
	case FeatureGuidedType, BugfixGuidedType, DocsGuidedType, RefactorGuidedType:
		return true
	default:
		return false
	}
}
