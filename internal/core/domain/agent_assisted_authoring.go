package domain

import (
	"fmt"
	"strings"
)

type AgentName string

func ParseAgentName(value string) (AgentName, error) {
	agentName := AgentName(strings.TrimSpace(value))
	if agentName == "" {
		return "", fmt.Errorf("agent name is required")
	}
	if strings.ContainsAny(string(agentName), "\r\n") {
		return "", fmt.Errorf("agent name must be safe display text")
	}
	if _, ok := RecognizedAgentTarget(agentName); !ok {
		return "", fmt.Errorf("unknown agent target: %s", agentName)
	}
	return agentName, nil
}

type RecognizedAgentTargetInfo struct {
	ID          AgentName
	DisplayName string
}

func RecognizedAgentTargets() []RecognizedAgentTargetInfo {
	return []RecognizedAgentTargetInfo{
		{ID: AgentName(CodexAgent), DisplayName: "Codex"},
		{ID: AgentName(ClaudeCodeAgent), DisplayName: "Claude Code"},
		{ID: AgentName(DevinAgent), DisplayName: "Devin"},
		{ID: AgentName(CursorAgent), DisplayName: "Cursor"},
		{ID: AgentName(CopilotAgent), DisplayName: "GitHub Copilot"},
		{ID: AgentName(GeminiAgent), DisplayName: "Gemini CLI"},
		{ID: AgentName(RooAgent), DisplayName: "Roo Code"},
		{ID: AgentName(WindsurfAgent), DisplayName: "Windsurf"},
		{ID: AgentName(AiderAgent), DisplayName: "Aider"},
		{ID: AgentName(GenericAgent), DisplayName: "Generic Agent"},
	}
}

func RecognizedAgentTarget(agentName AgentName) (RecognizedAgentTargetInfo, bool) {
	for _, target := range RecognizedAgentTargets() {
		if target.ID == agentName {
			return target, true
		}
	}
	return RecognizedAgentTargetInfo{}, false
}

type ResolvedAgentCommand struct {
	AgentID          AgentName
	AgentDisplayName string
	CommandName      string
	fixedArgs        []string
}

func NewResolvedAgentCommand(
	agentID AgentName,
	agentDisplayName string,
	commandName string,
	fixedArgs []string,
) ResolvedAgentCommand {
	return ResolvedAgentCommand{
		AgentID:          agentID,
		AgentDisplayName: agentDisplayName,
		CommandName:      commandName,
		fixedArgs:        append([]string(nil), fixedArgs...),
	}
}

func (command ResolvedAgentCommand) FixedArgs() []string {
	return append([]string(nil), command.fixedArgs...)
}

func ExecutableAgentCommands() []ResolvedAgentCommand {
	commands := make([]ResolvedAgentCommand, 0, 9)
	for _, agentName := range []AgentName{
		AgentName(CodexAgent),
		AgentName(ClaudeCodeAgent),
		AgentName(DevinAgent),
		AgentName(CursorAgent),
		AgentName(CopilotAgent),
		AgentName(GeminiAgent),
		AgentName(RooAgent),
		AgentName(WindsurfAgent),
		AgentName(AiderAgent),
	} {
		command, _ := ResolveExecutableAgentCommand(agentName)
		commands = append(commands, command)
	}
	return commands
}

func ResolveExecutableAgentCommand(agentName AgentName) (ResolvedAgentCommand, error) {
	target, ok := RecognizedAgentTarget(agentName)
	if !ok {
		return ResolvedAgentCommand{}, fmt.Errorf("unknown agent target: %s", agentName)
	}
	if agentName == AgentName(GenericAgent) {
		return ResolvedAgentCommand{}, fmt.Errorf("agent target has no executable local runner mapping in this change: %s", agentName)
	}

	return ResolvedAgentCommand{
		AgentID:          target.ID,
		AgentDisplayName: target.DisplayName,
		CommandName:      string(target.ID),
	}, nil
}

type AgentAssistedAuthoringType string

const (
	FeatureAgentAssistedAuthoringType  AgentAssistedAuthoringType = "feature"
	BugfixAgentAssistedAuthoringType   AgentAssistedAuthoringType = "bugfix"
	DocsAgentAssistedAuthoringType     AgentAssistedAuthoringType = "docs"
	RefactorAgentAssistedAuthoringType AgentAssistedAuthoringType = "refactor"
)

func ParseAgentAssistedAuthoringType(value string) (AgentAssistedAuthoringType, error) {
	authoringType := AgentAssistedAuthoringType(strings.TrimSpace(value))
	if authoringType == "" {
		return "", fmt.Errorf("agent-assisted authoring type is required")
	}
	if !authoringType.IsSupported() {
		return "", fmt.Errorf("unknown agent-assisted authoring type: %s", authoringType)
	}
	return authoringType, nil
}

func SupportedAgentAssistedAuthoringTypes() []AgentAssistedAuthoringType {
	return []AgentAssistedAuthoringType{
		FeatureAgentAssistedAuthoringType,
		BugfixAgentAssistedAuthoringType,
		DocsAgentAssistedAuthoringType,
		RefactorAgentAssistedAuthoringType,
	}
}

func (authoringType AgentAssistedAuthoringType) IsSupported() bool {
	switch authoringType {
	case FeatureAgentAssistedAuthoringType,
		BugfixAgentAssistedAuthoringType,
		DocsAgentAssistedAuthoringType,
		RefactorAgentAssistedAuthoringType:
		return true
	default:
		return false
	}
}

type AgentAssistedAuthoringPromptRequest struct {
	ChangeID      string
	AgentName     AgentName
	AuthoringType AgentAssistedAuthoringType
	Title         string
	Summary       string
	ChangePath    string
	RequiredFiles []string
}

func NewAgentAssistedAuthoringPromptRequest(
	changeID string,
	agentName AgentName,
	authoringType AgentAssistedAuthoringType,
	title string,
	summary string,
	changePath string,
	requiredFiles []string,
) AgentAssistedAuthoringPromptRequest {
	return AgentAssistedAuthoringPromptRequest{
		ChangeID:      changeID,
		AgentName:     agentName,
		AuthoringType: authoringType,
		Title:         title,
		Summary:       summary,
		ChangePath:    changePath,
		RequiredFiles: append([]string(nil), requiredFiles...),
	}
}

type AgentAssistedAuthoringResult struct {
	ChangeID                             string
	AgentName                            AgentName
	AgentDisplayName                     string
	AuthoringType                        AgentAssistedAuthoringType
	Title                                string
	Summary                              string
	ChangePath                           string
	DryRun                               bool
	Execute                              bool
	ResolvedCommandName                  string
	WorkingDirectory                     string
	PromptSentToRunner                   bool
	NoFilesWritten                       bool
	NoPromptFileWritten                  bool
	NoAgentExecuted                      bool
	NoExternalCommandExecuted            bool
	NoAgentOutputParsedOrApplied         bool
	NoOpenSpecFilesWrittenFromOutput     bool
	NoProductionCodeModifiedBySpecHarbor bool
	NoAutoCommitPushMerge                bool
	Prompt                               string
	requiredFiles                        []string
	plan                                 []string
	resolvedCommandFixedArgs             []string
	runnerResult                         *AgentRunResult
}

func NewAgentAssistedAuthoringResult(
	changeID string,
	agentName AgentName,
	authoringType AgentAssistedAuthoringType,
	title string,
	summary string,
	changePath string,
	requiredFiles []string,
	plan []string,
	prompt string,
) AgentAssistedAuthoringResult {
	return AgentAssistedAuthoringResult{
		ChangeID:                             changeID,
		AgentName:                            agentName,
		AgentDisplayName:                     displayNameForAgent(agentName),
		AuthoringType:                        authoringType,
		Title:                                title,
		Summary:                              summary,
		ChangePath:                           changePath,
		DryRun:                               true,
		NoOpenSpecFilesWrittenFromOutput:     true,
		NoProductionCodeModifiedBySpecHarbor: true,
		NoAutoCommitPushMerge:                true,
		NoFilesWritten:                       true,
		NoPromptFileWritten:                  true,
		NoAgentExecuted:                      true,
		NoExternalCommandExecuted:            true,
		NoAgentOutputParsedOrApplied:         true,
		Prompt:                               prompt,
		requiredFiles:                        append([]string(nil), requiredFiles...),
		plan:                                 append([]string(nil), plan...),
	}
}

func NewExecutedAgentAssistedAuthoringResult(
	changeID string,
	agentCommand ResolvedAgentCommand,
	authoringType AgentAssistedAuthoringType,
	title string,
	summary string,
	changePath string,
	workingDirectory string,
	requiredFiles []string,
	plan []string,
	prompt string,
	runnerResult AgentRunResult,
) AgentAssistedAuthoringResult {
	return AgentAssistedAuthoringResult{
		ChangeID:                             changeID,
		AgentName:                            agentCommand.AgentID,
		AgentDisplayName:                     agentCommand.AgentDisplayName,
		AuthoringType:                        authoringType,
		Title:                                title,
		Summary:                              summary,
		ChangePath:                           changePath,
		Execute:                              true,
		ResolvedCommandName:                  agentCommand.CommandName,
		WorkingDirectory:                     strings.TrimSpace(workingDirectory),
		PromptSentToRunner:                   true,
		NoFilesWritten:                       true,
		NoPromptFileWritten:                  true,
		NoAgentOutputParsedOrApplied:         true,
		NoOpenSpecFilesWrittenFromOutput:     true,
		NoProductionCodeModifiedBySpecHarbor: true,
		NoAutoCommitPushMerge:                true,
		Prompt:                               prompt,
		requiredFiles:                        append([]string(nil), requiredFiles...),
		plan:                                 append([]string(nil), plan...),
		resolvedCommandFixedArgs:             agentCommand.FixedArgs(),
		runnerResult:                         copyAgentRunResult(runnerResult),
	}
}

func (result AgentAssistedAuthoringResult) RequiredFiles() []string {
	return append([]string(nil), result.requiredFiles...)
}

func (result AgentAssistedAuthoringResult) Plan() []string {
	return append([]string(nil), result.plan...)
}

func (result AgentAssistedAuthoringResult) ResolvedCommandFixedArgs() []string {
	return append([]string(nil), result.resolvedCommandFixedArgs...)
}

func (result AgentAssistedAuthoringResult) RunnerResult() (AgentRunResult, bool) {
	if result.runnerResult == nil {
		return AgentRunResult{}, false
	}
	return *result.runnerResult, true
}

func displayNameForAgent(agentName AgentName) string {
	target, ok := RecognizedAgentTarget(agentName)
	if !ok {
		return ""
	}
	return target.DisplayName
}

func copyAgentRunResult(result AgentRunResult) *AgentRunResult {
	copy := result
	return &copy
}

type AgentRunStatus string

const (
	AgentRunStatusSuccess     AgentRunStatus = "success"
	AgentRunStatusNonZeroExit AgentRunStatus = "non_zero_exit"
)

type AgentRunRequest struct {
	AgentID          AgentName
	AgentDisplayName string
	CommandName      string
	prompt           string
	workingDirectory string
	fixedArgs        []string
}

func NewAgentRunRequest(
	command ResolvedAgentCommand,
	prompt string,
	workingDirectory string,
) AgentRunRequest {
	return AgentRunRequest{
		AgentID:          command.AgentID,
		AgentDisplayName: command.AgentDisplayName,
		CommandName:      command.CommandName,
		prompt:           prompt,
		workingDirectory: strings.TrimSpace(workingDirectory),
		fixedArgs:        command.FixedArgs(),
	}
}

func (request AgentRunRequest) Prompt() string {
	return request.prompt
}

func (request AgentRunRequest) WorkingDirectory() string {
	return request.workingDirectory
}

func (request AgentRunRequest) FixedArgs() []string {
	return append([]string(nil), request.fixedArgs...)
}

type AgentRunResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Status   AgentRunStatus
}

func NewAgentRunResult(stdout string, stderr string, exitCode int) AgentRunResult {
	status := AgentRunStatusSuccess
	if exitCode != 0 {
		status = AgentRunStatusNonZeroExit
	}
	return AgentRunResult{
		Stdout:   stdout,
		Stderr:   stderr,
		ExitCode: exitCode,
		Status:   status,
	}
}
