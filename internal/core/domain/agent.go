package domain

type Agent string

const (
	CodexAgent      Agent = "codex"
	ClaudeCodeAgent Agent = "claude"
	DevinAgent      Agent = "devin"
	CursorAgent     Agent = "cursor"
	CopilotAgent    Agent = "copilot"
	GeminiAgent     Agent = "gemini"
	RooAgent        Agent = "roo"
	WindsurfAgent   Agent = "windsurf"
	AiderAgent      Agent = "aider"
	GenericAgent    Agent = "generic"
)
