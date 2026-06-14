package cli

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/guferreira1/spec-harbor/internal/adapters/filesystem"
	githubadapter "github.com/guferreira1/spec-harbor/internal/adapters/github"
	"github.com/guferreira1/spec-harbor/internal/core/domain"
	"github.com/guferreira1/spec-harbor/internal/core/ports"
	"github.com/guferreira1/spec-harbor/internal/core/usecase"
)

var newGitHubRemoteContextReader = func(token string) ports.GitHubRemoteContextReader {
	return githubadapter.NewHTTPRemoteContextClient(token)
}

func contextCommand(ctx CommandContext) error {
	arguments, err := parseContextArguments(ctx.Args)
	if err != nil {
		return err
	}

	root, err := os.Getwd()
	if err != nil {
		return err
	}

	if arguments.subcommand == "index" {
		fileSystem := filesystem.NewRepositoryContextIndexFileSystem()
		buildIndex := usecase.NewBuildRepositoryContextIndex(fileSystem)
		result, err := buildIndex.Execute(usecase.RepositoryContextIndexInput{
			ProjectRoot: root,
			Mode:        arguments.indexMode,
		})
		if err != nil {
			return err
		}
		printRepositoryContextIndexReport(ctx.Output, result)
		if result.Status == domain.RepositoryContextIndexStatusMissing ||
			result.Status == domain.RepositoryContextIndexStatusStale ||
			result.Status == domain.RepositoryContextIndexStatusInvalid {
			return ExitError{Code: 1}
		}
		return nil
	}
	if arguments.subcommand == "retrieve" {
		fileSystem := filesystem.NewRepositoryContextIndexFileSystem()
		retrieveContext := usecase.NewRetrieveLocalContext(fileSystem)
		result, err := retrieveContext.Execute(usecase.RetrieveLocalContextInput{
			ProjectRoot: root,
			Query:       arguments.retrieveQuery,
		})
		if err != nil {
			return err
		}
		printLocalContextRetrievalReport(ctx.Output, result)
		if result.Status != domain.LocalContextRetrievalStatusCurrent &&
			result.Status != domain.LocalContextRetrievalStatusNoResults {
			return ExitError{Code: 1}
		}
		return nil
	}
	if arguments.subcommand == "github" {
		reader := newGitHubRemoteContextReader(os.Getenv(domain.GitHubRemoteContextTokenEnvVar))
		retrieveContext := usecase.NewRetrieveGitHubRemoteContext(reader)
		result, err := retrieveContext.Execute(usecase.RetrieveGitHubRemoteContextInput{
			Repository:  arguments.githubRepo,
			Ref:         arguments.githubRef,
			Query:       arguments.githubQuery,
			PathFilters: arguments.githubPaths,
		})
		if err != nil {
			return err
		}
		printGitHubRemoteContextReport(ctx.Output, result)
		if result.Status != domain.GitHubRemoteContextStatusCurrent &&
			result.Status != domain.GitHubRemoteContextStatusNoResults {
			return ExitError{Code: 1}
		}
		return nil
	}

	fileSystem := filesystem.NewContextDiscoveryFileSystem()
	discoverContext := usecase.NewDiscoverProjectContext(fileSystem)
	result, err := discoverContext.Execute(usecase.DiscoverProjectContextInput{ProjectRoot: root})
	if err != nil {
		return err
	}

	printContextDiscoveryReport(ctx.Output, result)
	return nil
}

type contextArguments struct {
	subcommand    string
	indexMode     domain.RepositoryContextIndexMode
	retrieveQuery string
	githubRepo    string
	githubRef     string
	githubQuery   string
	githubPaths   []string
}

func parseContextArguments(args []string) (contextArguments, error) {
	if len(args) == 0 {
		return contextArguments{}, fmt.Errorf("context subcommand is required: discover, index, retrieve, or github")
	}
	if strings.HasPrefix(args[0], "-") {
		return contextArguments{}, fmt.Errorf("unsupported flag: %s", args[0])
	}
	if args[0] == "discover" {
		for _, arg := range args[1:] {
			if strings.HasPrefix(arg, "-") {
				return contextArguments{}, fmt.Errorf("unsupported flag: %s", arg)
			}
			return contextArguments{}, fmt.Errorf("unexpected argument: %s", arg)
		}
		return contextArguments{subcommand: "discover"}, nil
	}
	if args[0] == "index" {
		return parseContextIndexArguments(args[1:])
	}
	if args[0] == "retrieve" {
		return parseContextRetrieveArguments(args[1:])
	}
	if args[0] == "github" {
		return parseContextGitHubArguments(args[1:])
	}
	return contextArguments{}, fmt.Errorf("unsupported context subcommand: %s", args[0])
}

func parseContextIndexArguments(args []string) (contextArguments, error) {
	writeProvided := false
	checkProvided := false
	for _, arg := range args {
		if arg == "--write" {
			if writeProvided {
				return contextArguments{}, fmt.Errorf("context index write flag specified more than once")
			}
			writeProvided = true
			continue
		}
		if arg == "--check" {
			if checkProvided {
				return contextArguments{}, fmt.Errorf("context index check flag specified more than once")
			}
			checkProvided = true
			continue
		}
		if strings.HasPrefix(arg, "-") {
			return contextArguments{}, fmt.Errorf("unsupported flag: %s", arg)
		}
		return contextArguments{}, fmt.Errorf("unexpected argument: %s", arg)
	}
	if writeProvided && checkProvided {
		return contextArguments{}, fmt.Errorf("context index write and check flags cannot be used together")
	}
	mode := domain.RepositoryContextIndexModeReport
	if writeProvided {
		mode = domain.RepositoryContextIndexModeWrite
	}
	if checkProvided {
		mode = domain.RepositoryContextIndexModeCheck
	}
	return contextArguments{subcommand: "index", indexMode: mode}, nil
}

func parseContextRetrieveArguments(args []string) (contextArguments, error) {
	queryProvided := false
	query := ""
	for index := 0; index < len(args); index++ {
		arg := args[index]
		if arg == "--query" {
			if queryProvided {
				return contextArguments{}, fmt.Errorf("context retrieve query flag specified more than once")
			}
			if index+1 >= len(args) {
				return contextArguments{}, fmt.Errorf("context retrieve query value is required")
			}
			if strings.HasPrefix(args[index+1], "-") {
				return contextArguments{}, fmt.Errorf("unsupported flag: %s", args[index+1])
			}
			queryProvided = true
			query = args[index+1]
			index++
			continue
		}
		if strings.HasPrefix(arg, "-") {
			return contextArguments{}, fmt.Errorf("unsupported flag: %s", arg)
		}
		return contextArguments{}, fmt.Errorf("unexpected argument: %s", arg)
	}
	if !queryProvided {
		return contextArguments{}, fmt.Errorf("context retrieve query is required")
	}
	return contextArguments{subcommand: "retrieve", retrieveQuery: query}, nil
}

func parseContextGitHubArguments(args []string) (contextArguments, error) {
	repoProvided := false
	queryProvided := false
	refProvided := false
	arguments := contextArguments{subcommand: "github"}
	for index := 0; index < len(args); index++ {
		arg := args[index]
		switch arg {
		case "--repo":
			if repoProvided {
				return contextArguments{}, fmt.Errorf("context github repo flag specified more than once")
			}
			value, nextIndex, err := contextFlagValue(args, index, "--repo")
			if err != nil {
				return contextArguments{}, err
			}
			repoProvided = true
			arguments.githubRepo = value
			index = nextIndex
		case "--query":
			if queryProvided {
				return contextArguments{}, fmt.Errorf("context github query flag specified more than once")
			}
			value, nextIndex, err := contextFlagValue(args, index, "--query")
			if err != nil {
				return contextArguments{}, err
			}
			queryProvided = true
			arguments.githubQuery = value
			index = nextIndex
		case "--ref":
			if refProvided {
				return contextArguments{}, fmt.Errorf("context github ref flag specified more than once")
			}
			value, nextIndex, err := contextFlagValue(args, index, "--ref")
			if err != nil {
				return contextArguments{}, err
			}
			refProvided = true
			arguments.githubRef = value
			index = nextIndex
		case "--path":
			value, nextIndex, err := contextFlagValue(args, index, "--path")
			if err != nil {
				return contextArguments{}, err
			}
			arguments.githubPaths = append(arguments.githubPaths, value)
			index = nextIndex
		default:
			if strings.HasPrefix(arg, "-") {
				return contextArguments{}, fmt.Errorf("unsupported flag: %s", arg)
			}
			return contextArguments{}, fmt.Errorf("unexpected argument: %s", arg)
		}
	}
	if !repoProvided {
		return contextArguments{}, fmt.Errorf("context github repo is required")
	}
	if !queryProvided {
		return contextArguments{}, fmt.Errorf("context github query is required")
	}
	return arguments, nil
}

func contextFlagValue(args []string, index int, flag string) (string, int, error) {
	if index+1 >= len(args) {
		return "", index, fmt.Errorf("context github %s value is required", strings.TrimPrefix(flag, "--"))
	}
	value := args[index+1]
	if strings.HasPrefix(value, "-") {
		return "", index, fmt.Errorf("unsupported flag: %s", value)
	}
	return value, index + 1, nil
}

func printContextDiscoveryReport(output io.Writer, result domain.ContextDiscoveryResult) {
	fmt.Fprintln(output, "Detected project context:")
	printContextSignalSection(
		output,
		"User-confirmed context:",
		result.SignalsByClassification(domain.ContextSignalClassificationUserConfirmedContext),
	)
	printContextSignalSection(
		output,
		"Detected facts:",
		result.SignalsByClassification(domain.ContextSignalClassificationDetectedFact),
	)
	printContextSignalSection(
		output,
		"Suggested assumptions:",
		result.SignalsByClassification(domain.ContextSignalClassificationSuggestedAssumption),
	)
	printContextNotesSection(output, result.Notes())
}

func printContextSignalSection(output io.Writer, heading string, signals []domain.ContextSignal) {
	fmt.Fprintln(output)
	fmt.Fprintln(output, heading)
	if len(signals) == 0 {
		fmt.Fprintln(output, "- none detected")
		return
	}
	for _, signal := range signals {
		fmt.Fprintf(output, "- %s: %s\n", signal.Kind.Label(), signal.Value)
		fmt.Fprintf(output, "  Source: %s\n", formatContextSource(signal.Source))
		fmt.Fprintf(output, "  Classification: %s\n", signal.Classification)
		fmt.Fprintf(output, "  Confidence: %s\n", signal.Confidence)
	}
}

func printContextNotesSection(output io.Writer, notes []domain.ContextDiscoveryNote) {
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Notes:")
	if len(notes) == 0 {
		fmt.Fprintln(output, "- none detected")
		return
	}
	for _, note := range notes {
		fmt.Fprintf(output, "- %s\n", note.Message)
	}
}

func formatContextSource(source domain.ContextSource) string {
	if source.Evidence == "" {
		return source.Path
	}
	return fmt.Sprintf("%s (%s)", source.Path, source.Evidence)
}

func printRepositoryContextIndexReport(output io.Writer, result domain.RepositoryContextIndexReport) {
	fmt.Fprintln(output, "Repository context index:")
	fmt.Fprintf(output, "Mode: %s\n", result.Mode)
	fmt.Fprintf(output, "Status: %s\n", result.Status)
	fmt.Fprintf(output, "Path: %s\n", result.IndexPath)
	fmt.Fprintf(output, "Schema version: %d\n", result.Index.SchemaVersion)
	fmt.Fprintf(output, "Indexed files: %d\n", len(result.Index.Entries))
	fmt.Fprintf(output, "Skipped records: %d\n", len(result.Index.Skipped))
	fmt.Fprintf(output, "Total indexed bytes: %d\n", repositoryContextIndexTotalBytes(result.Index))
	fmt.Fprintf(output, "Truncated: %s\n", yesNo(result.Index.Truncated))

	fmt.Fprintln(output)
	fmt.Fprintln(output, "Limits:")
	fmt.Fprintf(output, "- max indexed files: %d\n", result.Index.Limits.MaxIndexedFiles)
	fmt.Fprintf(output, "- max file size bytes: %d\n", result.Index.Limits.MaxFileSizeBytes)
	fmt.Fprintf(output, "- max total file bytes: %d\n", result.Index.Limits.MaxTotalFileBytes)
	fmt.Fprintf(output, "- max skipped records: %d\n", result.Index.Limits.MaxSkippedRecords)
	fmt.Fprintf(output, "- max directory depth: %d\n", result.Index.Limits.MaxDirectoryDepth)

	if result.ErrorMessage != "" {
		fmt.Fprintln(output)
		fmt.Fprintf(output, "Detail: %s\n", result.ErrorMessage)
	}

	if len(result.StaleReasons) > 0 {
		fmt.Fprintln(output)
		fmt.Fprintf(output, "Stale reasons: %d\n", len(result.StaleReasons))
		for index, reason := range result.StaleReasons {
			if index >= 10 {
				fmt.Fprintln(output, "- additional stale reasons omitted")
				break
			}
			if reason.Path != "" {
				fmt.Fprintf(output, "- %s: %s (%s)\n", reason.Code, reason.Message, reason.Path)
				continue
			}
			fmt.Fprintf(output, "- %s: %s\n", reason.Code, reason.Message)
		}
	}

	if len(result.Index.Skipped) > 0 {
		fmt.Fprintln(output)
		fmt.Fprintln(output, "Skipped:")
		for index, skipped := range result.Index.Skipped {
			if index >= 10 {
				fmt.Fprintln(output, "- additional skipped records omitted")
				break
			}
			fmt.Fprintf(output, "- %s: %s\n", skipped.Reason, skipped.Path)
		}
	}
}

func repositoryContextIndexTotalBytes(index domain.RepositoryContextIndex) int64 {
	var total int64
	for _, entry := range index.Entries {
		total += entry.SizeBytes
	}
	return total
}

func printLocalContextRetrievalReport(output io.Writer, result domain.LocalContextRetrievalReport) {
	fmt.Fprintln(output, "Local context retrieval:")
	fmt.Fprintf(output, "Query: %s\n", result.Query.DisplayQuery)
	fmt.Fprintf(output, "Normalized terms: %s\n", strings.Join(result.Query.Terms, ", "))
	fmt.Fprintf(output, "Index: %s\n", result.IndexPath)
	fmt.Fprintf(output, "Index status: %s\n", localContextRetrievalIndexStatus(result.Status))
	fmt.Fprintf(output, "Results: %d\n", len(result.Results))
	if result.OutputTruncated {
		fmt.Fprintln(output, "Output truncated: yes")
	}

	if result.Message != "" {
		fmt.Fprintln(output)
		fmt.Fprintf(output, "Detail: %s\n", result.Message)
	}
	if len(result.StaleReasons) > 0 {
		fmt.Fprintln(output)
		fmt.Fprintf(output, "Stale reasons: %d\n", len(result.StaleReasons))
		for _, reason := range result.StaleReasons {
			if reason.Path != "" {
				fmt.Fprintf(output, "- %s: %s (%s)\n", reason.Code, reason.Message, reason.Path)
				continue
			}
			fmt.Fprintf(output, "- %s: %s\n", reason.Code, reason.Message)
		}
	}
	if result.Status == domain.LocalContextRetrievalStatusNoResults {
		fmt.Fprintln(output)
		fmt.Fprintln(output, "No matching local context found.")
		return
	}
	if len(result.Results) == 0 {
		return
	}

	fmt.Fprintln(output)
	fmt.Fprintln(output, "Results:")
	for _, retrievalResult := range result.Results {
		fmt.Fprintf(output, "%d. %s\n", retrievalResult.Rank, retrievalResult.Path)
		fmt.Fprintf(output, "   Category: %s\n", retrievalResult.SourceCategory)
		if retrievalResult.SourceEvidenceCategory != "" {
			fmt.Fprintf(output, "   Evidence: %s\n", retrievalResult.SourceEvidenceCategory)
		}
		fmt.Fprintf(output, "   Score: %d\n", retrievalResult.Score)
		if len(retrievalResult.ClassificationHints) > 0 {
			fmt.Fprintf(output, "   Classification: %s\n", formatRetrievalClassificationHints(retrievalResult.ClassificationHints))
		}
		if retrievalResult.Snippet.Text != "" {
			fmt.Fprintf(output, "   Lines: %d-%d\n", retrievalResult.Snippet.LineStart, retrievalResult.Snippet.LineEnd)
			fmt.Fprintln(output, "   Snippet:")
			for _, line := range strings.Split(retrievalResult.Snippet.Text, "\n") {
				fmt.Fprintf(output, "   %s\n", line)
			}
			continue
		}
		if retrievalResult.Summary != "" {
			fmt.Fprintf(output, "   Summary: %s\n", retrievalResult.Summary)
		}
	}
}

func localContextRetrievalIndexStatus(status domain.LocalContextRetrievalStatus) string {
	if status == domain.LocalContextRetrievalStatusCurrent ||
		status == domain.LocalContextRetrievalStatusNoResults {
		return "current"
	}
	return string(status)
}

func formatRetrievalClassificationHints(hints []domain.RepositoryContextIndexClassificationHint) string {
	values := make([]string, 0, len(hints))
	for _, hint := range hints {
		values = append(values, string(hint))
	}
	return strings.Join(values, ", ")
}

func printGitHubRemoteContextReport(output io.Writer, result domain.GitHubRemoteContextReport) {
	fmt.Fprintln(output, "GitHub remote context:")
	fmt.Fprintf(output, "Repository: %s\n", result.Repository)
	fmt.Fprintf(output, "Query: %s\n", result.Query.DisplayQuery)
	fmt.Fprintf(output, "Normalized terms: %s\n", strings.Join(result.Query.Terms, ", "))
	if result.RequestedRef != "" {
		fmt.Fprintf(output, "Requested ref: %s\n", result.RequestedRef)
	}
	if result.DefaultBranch != "" {
		fmt.Fprintf(output, "Default branch: %s\n", result.DefaultBranch)
	}
	if result.ResolvedRef != "" {
		fmt.Fprintf(output, "Resolved ref: %s\n", result.ResolvedRef)
	}
	if result.CommitSHA != "" {
		fmt.Fprintf(output, "Resolved SHA: %s\n", result.CommitSHA)
	}
	if len(result.PathFilters) > 0 {
		fmt.Fprintf(output, "Path filters: %s\n", strings.Join(result.PathFilters, ", "))
	}
	fmt.Fprintln(output, "Remote: yes")
	fmt.Fprintf(output, "Status: %s\n", result.Status)
	fmt.Fprintf(output, "Results: %d\n", len(result.Results))
	if result.OutputTruncated {
		fmt.Fprintln(output, "Output truncated: yes")
	}
	if result.Message != "" {
		fmt.Fprintln(output)
		fmt.Fprintf(output, "Detail: %s\n", result.Message)
	}
	if len(result.Skipped) > 0 {
		fmt.Fprintln(output)
		fmt.Fprintln(output, "Skipped:")
		for index, skipped := range result.Skipped {
			if index >= 10 {
				fmt.Fprintln(output, "- additional skipped records omitted")
				break
			}
			fmt.Fprintf(output, "- %s: %s\n", skipped.Reason, skipped.Path)
		}
	}
	if result.Status == domain.GitHubRemoteContextStatusNoResults {
		fmt.Fprintln(output)
		fmt.Fprintln(output, "No matching GitHub remote context found.")
		return
	}
	if len(result.Results) == 0 {
		return
	}

	fmt.Fprintln(output)
	fmt.Fprintln(output, "Results:")
	for _, remoteResult := range result.Results {
		fmt.Fprintf(output, "%d. %s\n", remoteResult.Rank, remoteResult.Path)
		fmt.Fprintf(output, "   Repository: %s\n", remoteResult.Repository)
		if remoteResult.RequestedRef != "" {
			fmt.Fprintf(output, "   Requested ref: %s\n", remoteResult.RequestedRef)
		}
		if remoteResult.DefaultBranch != "" {
			fmt.Fprintf(output, "   Default branch: %s\n", remoteResult.DefaultBranch)
		}
		if remoteResult.ResolvedRef != "" {
			fmt.Fprintf(output, "   Resolved ref: %s\n", remoteResult.ResolvedRef)
		}
		if remoteResult.CommitSHA != "" {
			fmt.Fprintf(output, "   Resolved SHA: %s\n", remoteResult.CommitSHA)
		}
		fmt.Fprintln(output, "   Remote: yes")
		fmt.Fprintf(output, "   Category: %s\n", remoteResult.SourceCategory)
		if remoteResult.SourceEvidenceCategory != "" {
			fmt.Fprintf(output, "   Evidence: %s\n", remoteResult.SourceEvidenceCategory)
		}
		fmt.Fprintf(output, "   Score: %d\n", remoteResult.Score)
		if remoteResult.Snippet.Text != "" {
			fmt.Fprintf(output, "   Lines: %d-%d\n", remoteResult.Snippet.LineStart, remoteResult.Snippet.LineEnd)
			fmt.Fprintln(output, "   Snippet:")
			for _, line := range strings.Split(remoteResult.Snippet.Text, "\n") {
				fmt.Fprintf(output, "   %s\n", line)
			}
			continue
		}
		if remoteResult.Summary != "" {
			fmt.Fprintf(output, "   Summary: %s\n", remoteResult.Summary)
		}
	}
}
