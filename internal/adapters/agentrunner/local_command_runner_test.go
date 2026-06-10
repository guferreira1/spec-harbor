package agentrunner

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/guferreira1/spec-harbor/internal/core/domain"
)

func TestLocalCommandRunnerCapturesStdoutStderrAndZeroExit(t *testing.T) {
	request := helperRunRequest(t, "zero", nil, "prompt through stdin")

	result, err := NewLocalCommandRunner().Run(request)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.Status != domain.AgentRunStatusSuccess {
		t.Fatalf("Status = %q, want success", result.Status)
	}
	if result.ExitCode != 0 {
		t.Fatalf("ExitCode = %d, want 0", result.ExitCode)
	}
	if result.Stdout != "stdout:prompt through stdin" {
		t.Fatalf("Stdout = %q, want prompt from stdin", result.Stdout)
	}
	if result.Stderr != "stderr:zero" {
		t.Fatalf("Stderr = %q, want captured stderr", result.Stderr)
	}
}

func TestLocalCommandRunnerCapturesNonZeroExitAsCompletedResult(t *testing.T) {
	request := helperRunRequest(t, "nonzero", nil, "prompt")

	result, err := NewLocalCommandRunner().Run(request)
	if err != nil {
		t.Fatalf("Run() error = %v, want completed result", err)
	}
	if result.Status != domain.AgentRunStatusNonZeroExit {
		t.Fatalf("Status = %q, want non_zero_exit", result.Status)
	}
	if result.ExitCode != 7 {
		t.Fatalf("ExitCode = %d, want 7", result.ExitCode)
	}
	if result.Stdout != "stdout:nonzero" {
		t.Fatalf("Stdout = %q, want captured stdout", result.Stdout)
	}
	if result.Stderr != "stderr:nonzero" {
		t.Fatalf("Stderr = %q, want captured stderr", result.Stderr)
	}
}

func TestLocalCommandRunnerReturnsStartupFailureWithoutNormalResult(t *testing.T) {
	request := domain.NewAgentRunRequest(
		domain.NewResolvedAgentCommand("codex", "Codex", "__specharbor_missing_agent_runner_command__", nil),
		"prompt",
		t.TempDir(),
	)

	result, err := NewLocalCommandRunner().Run(request)
	if err == nil {
		t.Fatalf("Run() error = nil, want startup failure")
	}
	if !strings.Contains(err.Error(), `start agent runner command "__specharbor_missing_agent_runner_command__"`) {
		t.Fatalf("Run() error = %q, want command context", err.Error())
	}
	if result != (domain.AgentRunResult{}) {
		t.Fatalf("Run() result = %#v, want zero result for startup failure", result)
	}
}

func TestLocalCommandRunnerPassesFixedArgsDirectlyWithoutShellInterpolation(t *testing.T) {
	request := helperRunRequest(t, "args", []string{"literal;echo hacked", "$(echo hacked)"}, "prompt")

	result, err := NewLocalCommandRunner().Run(request)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.Stdout != "literal;echo hacked|$(echo hacked)" {
		t.Fatalf("Stdout = %q, want literal args without shell interpretation", result.Stdout)
	}
	if result.Stderr != "" {
		t.Fatalf("Stderr = %q, want empty stderr", result.Stderr)
	}
}

func TestLocalCommandRunnerRejectsMissingCommandAndWorkingDirectory(t *testing.T) {
	tests := []struct {
		name    string
		request domain.AgentRunRequest
		want    string
	}{
		{
			name: "missing command",
			request: domain.NewAgentRunRequest(
				domain.NewResolvedAgentCommand("codex", "Codex", " ", nil),
				"prompt",
				t.TempDir(),
			),
			want: "agent runner command is required",
		},
		{
			name: "missing working directory",
			request: domain.NewAgentRunRequest(
				domain.NewResolvedAgentCommand("codex", "Codex", os.Args[0], nil),
				"prompt",
				" ",
			),
			want: `agent runner working directory is required for command "`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := NewLocalCommandRunner().Run(test.request)
			if err == nil {
				t.Fatalf("Run() error = nil, want %q", test.want)
			}
			if !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Run() error = %q, want %q", err.Error(), test.want)
			}
			if !reflect.DeepEqual(result, domain.AgentRunResult{}) {
				t.Fatalf("Run() result = %#v, want zero result", result)
			}
		})
	}
}

func helperRunRequest(t *testing.T, mode string, args []string, prompt string) domain.AgentRunRequest {
	t.Helper()
	t.Setenv("SPECHARBOR_AGENTRUNNER_HELPER", "1")

	fixedArgs := []string{"-test.run", "^TestLocalCommandRunnerHelperProcess$", "--", mode}
	fixedArgs = append(fixedArgs, args...)

	return domain.NewAgentRunRequest(
		domain.NewResolvedAgentCommand("codex", "Codex", os.Args[0], fixedArgs),
		prompt,
		t.TempDir(),
	)
}

func TestLocalCommandRunnerHelperProcess(t *testing.T) {
	if os.Getenv("SPECHARBOR_AGENTRUNNER_HELPER") != "1" {
		return
	}

	args := os.Args
	separator := -1
	for index, arg := range args {
		if arg == "--" {
			separator = index
			break
		}
	}
	if separator == -1 || separator+1 >= len(args) {
		os.Exit(2)
	}

	mode := args[separator+1]
	rest := args[separator+2:]
	stdin, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(2)
	}

	switch mode {
	case "zero":
		fmt.Printf("stdout:%s", stdin)
		fmt.Fprint(os.Stderr, "stderr:zero")
		os.Exit(0)
	case "nonzero":
		fmt.Fprint(os.Stdout, "stdout:nonzero")
		fmt.Fprint(os.Stderr, "stderr:nonzero")
		os.Exit(7)
	case "args":
		fmt.Fprint(os.Stdout, strings.Join(rest, "|"))
		os.Exit(0)
	default:
		os.Exit(2)
	}
}
