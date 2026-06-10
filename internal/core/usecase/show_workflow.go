package usecase

import (
	"errors"

	"github.com/guferreira1/spec-harbor/internal/core/domain"
)

type ShowWorkflow struct{}

type ShowWorkflowResult struct {
	Workflow domain.Workflow
}

func NewShowWorkflow() *ShowWorkflow {
	return &ShowWorkflow{}
}

func (useCase *ShowWorkflow) Execute() (ShowWorkflowResult, error) {
	if useCase == nil {
		return ShowWorkflowResult{}, errors.New("show workflow use case is required")
	}

	workflow, err := domain.RecommendedWorkflow()
	if err != nil {
		return ShowWorkflowResult{}, err
	}

	return ShowWorkflowResult{Workflow: workflow}, nil
}
