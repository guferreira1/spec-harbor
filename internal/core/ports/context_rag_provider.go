package ports

import "github.com/guferreira1/spec-harbor/internal/core/domain"

// ContextRAGProvider generates one bounded answer from already-selected
// source-attributed context. It intentionally exposes no embedding, vector,
// file upload, tool execution, shell, agent, or source-control methods.
type ContextRAGProvider interface {
	GenerateContextAnswer(
		request domain.ContextRAGProviderRequest,
	) (domain.ContextRAGProviderResponse, error)
}
