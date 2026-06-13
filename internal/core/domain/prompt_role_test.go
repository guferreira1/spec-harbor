package domain

import "testing"

func TestSupportedPromptRolesRemainScopedToAgentWorkflowRoles(t *testing.T) {
	roles := SupportedPromptRoles()
	want := []PromptRole{
		PromptRoleSpecAuthor,
		PromptRoleArchitectureReviewer,
		PromptRoleImplementer,
		PromptRoleTestEngineer,
		PromptRoleChangeReviewer,
	}
	if len(roles) != len(want) {
		t.Fatalf("SupportedPromptRoles() = %v, want %v", roles, want)
	}
	for index, role := range want {
		if roles[index] != role {
			t.Fatalf("SupportedPromptRoles()[%d] = %q, want %q", index, roles[index], role)
		}
	}

	for _, unsupported := range []string{"pull-request", "pr", "archive", "archive-housekeeping"} {
		if _, ok := ParsePromptRole(unsupported); ok {
			t.Fatalf("ParsePromptRole(%q) is supported; PR/archive prompt roles are out of scope", unsupported)
		}
	}
}
