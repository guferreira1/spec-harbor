package agentrunner

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/guferreira1/spec-harbor/internal/core/domain"
)

type LocalCommandRunner struct{}

func NewLocalCommandRunner() *LocalCommandRunner {
	return &LocalCommandRunner{}
}

func (runner *LocalCommandRunner) Run(request domain.AgentRunRequest) (domain.AgentRunResult, error) {
	if runner == nil {
		return domain.AgentRunResult{}, fmt.Errorf("agent runner is required")
	}
	commandName := strings.TrimSpace(request.CommandName)
	if commandName == "" {
		return domain.AgentRunResult{}, fmt.Errorf("agent runner command is required")
	}
	workingDirectory := request.WorkingDirectory()
	if workingDirectory == "" {
		return domain.AgentRunResult{}, fmt.Errorf("agent runner working directory is required for command %q", commandName)
	}

	command := exec.Command(commandName, request.FixedArgs()...)
	command.Dir = workingDirectory
	command.Stdin = strings.NewReader(request.Prompt())

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr

	if err := command.Run(); err != nil {
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			return domain.NewAgentRunResult(stdout.String(), stderr.String(), exitError.ExitCode()), nil
		}
		return domain.AgentRunResult{}, fmt.Errorf(
			"start agent runner command %q in working directory %q: %w",
			commandName,
			workingDirectory,
			err,
		)
	}

	return domain.NewAgentRunResult(stdout.String(), stderr.String(), 0), nil
}
