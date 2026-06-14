package cli

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/guferreira1/spec-harbor/internal/adapters/filesystem"
	openaiadapter "github.com/guferreira1/spec-harbor/internal/adapters/openai"
	"github.com/guferreira1/spec-harbor/internal/core/domain"
	"github.com/guferreira1/spec-harbor/internal/core/ports"
	"github.com/guferreira1/spec-harbor/internal/core/usecase"
)

var newContextRagProvider = func(
	apiKey string,
	model string,
	limits domain.ContextRAGLimits,
) ports.ContextRAGProvider {
	return openaiadapter.NewHTTPContextRAGProvider(
		apiKey,
		model,
		time.Duration(limits.ProviderTimeoutSeconds)*time.Second,
		limits.MaxProviderResponseBytes,
	)
}

func contextRagCommand(ctx CommandContext, root string, arguments contextArguments) error {
	limits, err := contextRagLimits(arguments)
	if err != nil {
		return err
	}
	providerName, err := domain.NewContextRAGProviderName(arguments.ragProvider)
	if err != nil {
		return err
	}
	query, err := domain.NewContextRAGQuery(arguments.ragQuery, limits)
	if err != nil {
		return err
	}
	sourceKinds, err := domain.NewContextRAGSourceKinds(arguments.ragSources)
	if err != nil {
		return err
	}

	model := strings.TrimSpace(os.Getenv(domain.ContextRAGOpenAIModelEnvVar))
	if model == "" {
		model = domain.DefaultContextRAGOpenAIModel
	}
	apiKey := ""
	if providerName == domain.ContextRAGProviderOpenAI {
		apiKey = os.Getenv(domain.ContextRAGOpenAIAPIKeyEnvVar)
		if strings.TrimSpace(apiKey) == "" {
			result := domain.NewContextRAGReport(
				domain.ContextRAGStatusMissingCredentials,
				providerName,
				model,
				query,
				"",
				nil,
				fmt.Sprintf("missing OpenAI API key; set %s", domain.ContextRAGOpenAIAPIKeyEnvVar),
				false,
			)
			printContextRagReport(ctx.Output, result)
			return ExitError{Code: 1}
		}
	}

	fileSystem := filesystem.NewRepositoryContextIndexFileSystem()
	localRetriever := usecase.NewRetrieveLocalContext(fileSystem)
	var githubRetriever *usecase.RetrieveGitHubRemoteContext
	if contextRagIncludesSource(sourceKinds, domain.ContextRAGSourceGitHub) {
		githubReader := newGitHubRemoteContextReader(os.Getenv(domain.GitHubRemoteContextTokenEnvVar))
		githubRetriever = usecase.NewRetrieveGitHubRemoteContext(githubReader)
	}
	provider := newContextRagProvider(apiKey, model, limits)
	generateAnswer := usecase.NewGenerateContextRAGAnswerWithLimits(
		localRetriever,
		githubRetriever,
		provider,
		limits,
	)
	result, err := generateAnswer.Execute(usecase.GenerateContextRAGAnswerInput{
		ProjectRoot:       root,
		Query:             arguments.ragQuery,
		Provider:          arguments.ragProvider,
		Model:             model,
		SourceKinds:       arguments.ragSources,
		GitHubRepository:  arguments.ragRepo,
		GitHubRef:         arguments.ragRef,
		GitHubPathFilters: arguments.ragPaths,
	})
	if err != nil {
		return err
	}
	printContextRagReport(ctx.Output, result)
	if !contextRagStatusSucceeded(result.Status) {
		return ExitError{Code: 1}
	}
	return nil
}

func contextRagLimits(arguments contextArguments) (domain.ContextRAGLimits, error) {
	limits := domain.DefaultContextRAGLimits()
	var err error
	if arguments.ragMaxSources > 0 {
		limits, err = limits.WithMaxSources(arguments.ragMaxSources)
		if err != nil {
			return domain.ContextRAGLimits{}, err
		}
	}
	if arguments.ragMaxAnswerChars > 0 {
		limits, err = limits.WithMaxAnswerChars(arguments.ragMaxAnswerChars)
		if err != nil {
			return domain.ContextRAGLimits{}, err
		}
	}
	return limits, nil
}

func contextRagStatusSucceeded(status domain.ContextRAGStatus) bool {
	return status == domain.ContextRAGStatusAnswered ||
		status == domain.ContextRAGStatusInsufficientSources
}

func parseContextRagArguments(args []string) (contextArguments, error) {
	providerProvided := false
	queryProvided := false
	repoProvided := false
	refProvided := false
	maxSourcesProvided := false
	maxAnswerCharsProvided := false
	arguments := contextArguments{subcommand: "rag"}

	for index := 0; index < len(args); index++ {
		arg := args[index]
		switch arg {
		case "--provider":
			if providerProvided {
				return contextArguments{}, fmt.Errorf("context rag provider flag specified more than once")
			}
			value, nextIndex, err := contextRagFlagValue(args, index, "--provider")
			if err != nil {
				return contextArguments{}, err
			}
			providerProvided = true
			arguments.ragProvider = value
			index = nextIndex
		case "--query":
			if queryProvided {
				return contextArguments{}, fmt.Errorf("context rag query flag specified more than once")
			}
			value, nextIndex, err := contextRagFlagValue(args, index, "--query")
			if err != nil {
				return contextArguments{}, err
			}
			queryProvided = true
			arguments.ragQuery = value
			index = nextIndex
		case "--from":
			value, nextIndex, err := contextRagFlagValue(args, index, "--from")
			if err != nil {
				return contextArguments{}, err
			}
			arguments.ragSources = append(arguments.ragSources, value)
			index = nextIndex
		case "--repo":
			if repoProvided {
				return contextArguments{}, fmt.Errorf("context rag repo flag specified more than once")
			}
			value, nextIndex, err := contextRagFlagValue(args, index, "--repo")
			if err != nil {
				return contextArguments{}, err
			}
			repoProvided = true
			arguments.ragRepo = value
			index = nextIndex
		case "--ref":
			if refProvided {
				return contextArguments{}, fmt.Errorf("context rag ref flag specified more than once")
			}
			value, nextIndex, err := contextRagFlagValue(args, index, "--ref")
			if err != nil {
				return contextArguments{}, err
			}
			refProvided = true
			arguments.ragRef = value
			index = nextIndex
		case "--path":
			value, nextIndex, err := contextRagFlagValue(args, index, "--path")
			if err != nil {
				return contextArguments{}, err
			}
			arguments.ragPaths = append(arguments.ragPaths, value)
			index = nextIndex
		case "--max-sources":
			if maxSourcesProvided {
				return contextArguments{}, fmt.Errorf("context rag max sources flag specified more than once")
			}
			value, nextIndex, err := contextRagFlagValue(args, index, "--max-sources")
			if err != nil {
				return contextArguments{}, err
			}
			parsed, err := contextRagPositiveInt(value, "--max-sources")
			if err != nil {
				return contextArguments{}, err
			}
			maxSourcesProvided = true
			arguments.ragMaxSources = parsed
			index = nextIndex
		case "--max-answer-chars":
			if maxAnswerCharsProvided {
				return contextArguments{}, fmt.Errorf("context rag max answer chars flag specified more than once")
			}
			value, nextIndex, err := contextRagFlagValue(args, index, "--max-answer-chars")
			if err != nil {
				return contextArguments{}, err
			}
			parsed, err := contextRagPositiveInt(value, "--max-answer-chars")
			if err != nil {
				return contextArguments{}, err
			}
			maxAnswerCharsProvided = true
			arguments.ragMaxAnswerChars = parsed
			index = nextIndex
		default:
			if strings.HasPrefix(arg, "-") {
				return contextArguments{}, fmt.Errorf("unsupported flag: %s", arg)
			}
			return contextArguments{}, fmt.Errorf("unexpected argument: %s", arg)
		}
	}

	if !providerProvided {
		return contextArguments{}, fmt.Errorf("context rag provider is required")
	}
	if _, err := domain.NewContextRAGProviderName(arguments.ragProvider); err != nil {
		return contextArguments{}, err
	}
	if !queryProvided {
		return contextArguments{}, fmt.Errorf("context rag query is required")
	}
	sourceKinds, err := domain.NewContextRAGSourceKinds(arguments.ragSources)
	if err != nil {
		return contextArguments{}, err
	}
	hasGitHubSource := contextRagIncludesSource(sourceKinds, domain.ContextRAGSourceGitHub)
	if hasGitHubSource && !repoProvided {
		return contextArguments{}, fmt.Errorf("context rag repo is required when github sources are selected")
	}
	if !hasGitHubSource {
		if repoProvided {
			return contextArguments{}, fmt.Errorf("context rag repo requires --from github")
		}
		if refProvided {
			return contextArguments{}, fmt.Errorf("context rag ref requires --from github")
		}
		if len(arguments.ragPaths) > 0 {
			return contextArguments{}, fmt.Errorf("context rag path requires --from github")
		}
	}
	return arguments, nil
}

func contextRagFlagValue(args []string, index int, flag string) (string, int, error) {
	if index+1 >= len(args) {
		return "", index, fmt.Errorf("context rag %s value is required", strings.TrimPrefix(flag, "--"))
	}
	value := args[index+1]
	if strings.HasPrefix(value, "-") {
		return "", index, fmt.Errorf("unsupported flag: %s", value)
	}
	return value, index + 1, nil
}

func contextRagPositiveInt(raw string, flag string) (int, error) {
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return 0, fmt.Errorf("context rag %s must be a positive integer", strings.TrimPrefix(flag, "--"))
	}
	return value, nil
}

func contextRagIncludesSource(sources []domain.ContextRAGSourceKind, source domain.ContextRAGSourceKind) bool {
	for _, candidate := range sources {
		if candidate == source {
			return true
		}
	}
	return false
}

func printContextRagReport(output io.Writer, result domain.ContextRAGReport) {
	fmt.Fprintln(output, "Context RAG answer:")
	fmt.Fprintf(output, "Query: %s\n", result.Query.DisplayQuery)
	fmt.Fprintf(output, "Provider: %s\n", result.Provider)
	if result.Model != "" {
		fmt.Fprintf(output, "Model: %s\n", result.Model)
	}
	fmt.Fprintf(output, "Status: %s\n", result.Status)
	fmt.Fprintf(output, "Sources: %d\n", len(result.Sources))
	if result.OutputTruncated {
		fmt.Fprintln(output, "Output truncated: yes")
	}

	if result.Message != "" {
		fmt.Fprintln(output)
		fmt.Fprintf(output, "Detail: %s\n", result.Message)
	}
	if result.Answer != "" {
		fmt.Fprintln(output)
		fmt.Fprintln(output, "Answer:")
		for _, line := range strings.Split(result.Answer, "\n") {
			fmt.Fprintf(output, "%s\n", line)
		}
	}
	if len(result.Sources) == 0 {
		return
	}

	fmt.Fprintln(output)
	fmt.Fprintln(output, "Source list:")
	for _, source := range result.Sources {
		fmt.Fprintf(output, "%d. %s\n", source.ID, source.Path)
		fmt.Fprintf(output, "   Source: %s\n", source.Kind)
		fmt.Fprintf(output, "   Remote: %s\n", yesNo(source.Kind == domain.ContextRAGSourceGitHub))
		if source.Repository != "" {
			fmt.Fprintf(output, "   Repository: %s\n", source.Repository)
		}
		if source.RequestedRef != "" {
			fmt.Fprintf(output, "   Requested ref: %s\n", source.RequestedRef)
		}
		if source.DefaultBranch != "" {
			fmt.Fprintf(output, "   Default branch: %s\n", source.DefaultBranch)
		}
		if source.ResolvedRef != "" {
			fmt.Fprintf(output, "   Resolved ref: %s\n", source.ResolvedRef)
		}
		if source.CommitSHA != "" {
			fmt.Fprintf(output, "   Resolved SHA: %s\n", source.CommitSHA)
		}
		if source.SourceCategory != "" {
			fmt.Fprintf(output, "   Category: %s\n", source.SourceCategory)
		}
		if source.SourceEvidenceCategory != "" {
			fmt.Fprintf(output, "   Evidence: %s\n", source.SourceEvidenceCategory)
		}
		if source.LineStart > 0 {
			fmt.Fprintf(output, "   Lines: %d-%d\n", source.LineStart, source.LineEnd)
		}
		if source.Truncated {
			fmt.Fprintln(output, "   Truncated: yes")
		}
	}
}
