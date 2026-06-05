package prompt

type Agent string

const (
	CodexAgent Agent = "codex"
	ClaudeCodeAgent Agent = "claude-code"
	CursorAgent Agent = "cursor"
	GenericAgent Agent = "generic"
)
