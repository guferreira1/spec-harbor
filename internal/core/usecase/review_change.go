package usecase

import (
	"errors"
	"fmt"
	"strings"

	"github.com/guferreira1/spec-harbor/internal/core/domain"
	"github.com/guferreira1/spec-harbor/internal/core/ports"
)

type ReviewChangeInput struct {
	ProjectRoot string
	ChangeID    string
}

type ReviewChange struct {
	fileSystem ports.ReviewFileSystem
}

func NewReviewChange(fileSystem ports.ReviewFileSystem) *ReviewChange {
	return &ReviewChange{fileSystem: fileSystem}
}

func (useCase *ReviewChange) Execute(input ReviewChangeInput) (domain.ReviewResult, error) {
	if useCase == nil {
		return domain.ReviewResult{}, errors.New("review change use case is required")
	}
	if useCase.fileSystem == nil {
		return domain.ReviewResult{}, errors.New("review filesystem is required")
	}

	projectRoot, changeID, err := validateReviewInput(input)
	if err != nil {
		return domain.ReviewResult{}, err
	}

	requiredFiles := domain.RequiredOpenSpecChangeFiles()
	checkedPath := openspecChangesDirectory + "/" + changeID

	if result, available, err := useCase.reviewProjectStructure(projectRoot, changeID, checkedPath, requiredFiles); err != nil {
		return domain.ReviewResult{}, err
	} else if !available {
		return result, nil
	}

	changeDirectoryExists, err := useCase.fileSystem.DirectoryExists(projectRoot, checkedPath)
	if err != nil {
		return domain.ReviewResult{}, fmt.Errorf("check directory %s: %w", checkedPath, err)
	}
	if !changeDirectoryExists {
		return domain.NewReviewResult(changeID, checkedPath, requiredFiles, domain.ReviewTaskSummary{}, []domain.ReviewFinding{
			reviewChangeDirectoryMissingFinding(checkedPath),
		}), nil
	}

	findings, err := useCase.reviewRequiredFiles(projectRoot, checkedPath, requiredFiles)
	if err != nil {
		return domain.ReviewResult{}, err
	}
	if len(findings) > 0 {
		return domain.NewReviewResult(changeID, checkedPath, requiredFiles, domain.ReviewTaskSummary{}, findings), nil
	}

	tasksPath := checkedPath + "/tasks.md"
	contents, err := useCase.fileSystem.ReadFile(projectRoot, tasksPath)
	if err != nil {
		return domain.NewReviewResult(changeID, checkedPath, requiredFiles, domain.ReviewTaskSummary{}, []domain.ReviewFinding{
			reviewTasksFileUnreadableFinding(tasksPath),
		}), nil
	}

	tasks := parseReviewTasks(contents)
	if tasks.summary.Total == 0 {
		return domain.NewReviewResult(changeID, checkedPath, requiredFiles, tasks.summary, []domain.ReviewFinding{
			reviewTasksNotFoundFinding(tasksPath),
		}), nil
	}

	incompleteFindings := incompleteTaskFindings(tasksPath, tasks.incompleteTaskTexts)
	return domain.NewReviewResult(changeID, checkedPath, requiredFiles, tasks.summary, incompleteFindings), nil
}

func validateReviewInput(input ReviewChangeInput) (string, string, error) {
	projectRoot := strings.TrimSpace(input.ProjectRoot)
	if projectRoot == "" {
		return "", "", errors.New("project root is required")
	}

	changeID := strings.TrimSpace(input.ChangeID)
	if changeID == "" {
		return "", "", errors.New("change id is required")
	}
	if err := validateReviewChangeID(changeID); err != nil {
		return "", "", err
	}

	return projectRoot, changeID, nil
}

func validateReviewChangeID(changeID string) error {
	if changeID == "." || changeID == ".." {
		return errors.New("change id must be a safe single path segment")
	}
	if strings.HasPrefix(changeID, "-") ||
		strings.Contains(changeID, "/") ||
		strings.Contains(changeID, "\\") ||
		strings.Contains(changeID, ":") ||
		strings.Contains(changeID, "..") {
		return errors.New("change id must be a safe single path segment")
	}
	return nil
}

func (useCase *ReviewChange) reviewProjectStructure(projectRoot string, changeID string, checkedPath string, requiredFiles []string) (domain.ReviewResult, bool, error) {
	projectFileExists, err := useCase.fileSystem.FileExists(projectRoot, openspecProjectFile)
	if err != nil {
		return domain.ReviewResult{}, false, fmt.Errorf("check file %s: %w", openspecProjectFile, err)
	}

	changesDirectoryExists, err := useCase.fileSystem.DirectoryExists(projectRoot, openspecChangesDirectory)
	if err != nil {
		return domain.ReviewResult{}, false, fmt.Errorf("check directory %s: %w", openspecChangesDirectory, err)
	}

	if projectFileExists && changesDirectoryExists {
		return domain.ReviewResult{}, true, nil
	}

	return domain.NewReviewResult(changeID, checkedPath, requiredFiles, domain.ReviewTaskSummary{}, []domain.ReviewFinding{
		reviewProjectRootUnavailableFinding(),
	}), false, nil
}

func (useCase *ReviewChange) reviewRequiredFiles(projectRoot string, checkedPath string, requiredFiles []string) ([]domain.ReviewFinding, error) {
	var findings []domain.ReviewFinding

	for _, requiredFile := range requiredFiles {
		relativePath := checkedPath + "/" + requiredFile
		exists, err := useCase.fileSystem.FileExists(projectRoot, relativePath)
		if err != nil {
			return nil, fmt.Errorf("check file %s: %w", relativePath, err)
		}
		if exists {
			continue
		}

		findings = append(findings, reviewRequiredFileMissingFinding(relativePath, requiredFile))
	}

	return findings, nil
}

func reviewProjectRootUnavailableFinding() domain.ReviewFinding {
	return domain.ReviewFinding{
		Severity:     domain.ReviewFindingSeverityError,
		Code:         domain.ReviewFindingCodeProjectRootUnavailable,
		Message:      "OpenSpec project structure is unavailable.",
		RelativePath: "openspec",
	}
}

func reviewChangeDirectoryMissingFinding(checkedPath string) domain.ReviewFinding {
	return domain.ReviewFinding{
		Severity:     domain.ReviewFindingSeverityError,
		Code:         domain.ReviewFindingCodeChangeDirectoryMissing,
		Message:      "Missing change directory: " + checkedPath,
		RelativePath: checkedPath,
		Subject:      checkedPath,
	}
}

func reviewRequiredFileMissingFinding(relativePath string, fileName string) domain.ReviewFinding {
	return domain.ReviewFinding{
		Severity:     domain.ReviewFindingSeverityError,
		Code:         domain.ReviewFindingCodeRequiredFileMissing,
		Message:      "Missing required file: " + fileName,
		RelativePath: relativePath,
		Subject:      fileName,
	}
}

func reviewTasksFileUnreadableFinding(relativePath string) domain.ReviewFinding {
	return domain.ReviewFinding{
		Severity:     domain.ReviewFindingSeverityError,
		Code:         domain.ReviewFindingCodeTasksFileUnreadable,
		Message:      "Unable to read tasks file: tasks.md",
		RelativePath: relativePath,
		Subject:      "tasks.md",
	}
}

func reviewTasksNotFoundFinding(relativePath string) domain.ReviewFinding {
	return domain.ReviewFinding{
		Severity:     domain.ReviewFindingSeverityError,
		Code:         domain.ReviewFindingCodeTasksNotFound,
		Message:      "No task checkboxes found in tasks.md.",
		RelativePath: relativePath,
		Subject:      "tasks.md",
	}
}

func incompleteTaskFindings(relativePath string, taskTexts []string) []domain.ReviewFinding {
	var findings []domain.ReviewFinding
	for _, taskText := range taskTexts {
		message := "Task is not completed."
		if taskText != "" {
			message = "Task is not completed: " + taskText
		}

		findings = append(findings, domain.ReviewFinding{
			Severity:     domain.ReviewFindingSeverityWarning,
			Code:         domain.ReviewFindingCodeIncompleteTask,
			Message:      message,
			RelativePath: relativePath,
			Subject:      taskText,
		})
	}
	return findings
}

type parsedReviewTasks struct {
	summary             domain.ReviewTaskSummary
	incompleteTaskTexts []string
}

func parseReviewTasks(contents string) parsedReviewTasks {
	var completed int
	var incompleteTexts []string

	for _, line := range strings.Split(contents, "\n") {
		trimmed := strings.TrimLeft(line, " \t")
		switch {
		case strings.HasPrefix(trimmed, "- [x]"):
			completed++
		case strings.HasPrefix(trimmed, "- [X]"):
			completed++
		case strings.HasPrefix(trimmed, "- [ ]"):
			incompleteTexts = append(incompleteTexts, strings.TrimSpace(strings.TrimPrefix(trimmed, "- [ ]")))
		}
	}

	return parsedReviewTasks{
		summary:             domain.NewReviewTaskSummary(completed, len(incompleteTexts)),
		incompleteTaskTexts: incompleteTexts,
	}
}
