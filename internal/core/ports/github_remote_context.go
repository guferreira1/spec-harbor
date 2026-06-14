package ports

import "github.com/guferreira1/spec-harbor/internal/core/domain"

// GitHubRemoteContextReader provides read-only GitHub repository context access.
// It must not expose repository mutation, issue, pull request, release, label,
// workflow, branch creation, commit creation, or source-control automation methods.
type GitHubRemoteContextReader interface {
	ResolveRepository(locator domain.GitHubRepositoryLocator) (domain.GitHubRemoteRepository, error)
	ResolveRef(
		locator domain.GitHubRepositoryLocator,
		ref domain.GitHubRemoteRef,
	) (domain.GitHubRemoteResolvedRef, error)
	ListDirectory(
		locator domain.GitHubRepositoryLocator,
		ref domain.GitHubRemoteResolvedRef,
		relativePath string,
	) ([]domain.GitHubRemoteEntry, error)
	ReadFile(
		locator domain.GitHubRepositoryLocator,
		ref domain.GitHubRemoteResolvedRef,
		relativePath string,
		maxBytes int64,
	) (domain.GitHubRemoteFile, error)
}
