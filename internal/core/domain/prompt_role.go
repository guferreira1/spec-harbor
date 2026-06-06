package domain

import "strings"

type PromptRole string

const (
	PromptRoleSpecAuthor           PromptRole = "spec-author"
	PromptRoleArchitectureReviewer PromptRole = "architecture-reviewer"
	PromptRoleImplementer          PromptRole = "implementer"
	PromptRoleTestEngineer         PromptRole = "test-engineer"
	PromptRoleChangeReviewer       PromptRole = "change-reviewer"
)

func SupportedPromptRoles() []PromptRole {
	return []PromptRole{
		PromptRoleSpecAuthor,
		PromptRoleArchitectureReviewer,
		PromptRoleImplementer,
		PromptRoleTestEngineer,
		PromptRoleChangeReviewer,
	}
}

func ParsePromptRole(value string) (PromptRole, bool) {
	role := PromptRole(strings.TrimSpace(value))
	for _, supported := range SupportedPromptRoles() {
		if role == supported {
			return role, true
		}
	}
	return "", false
}
