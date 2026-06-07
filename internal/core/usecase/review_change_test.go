package usecase

import (
	"errors"
	"strings"
	"testing"

	"github.com/guferreira1/spec-harbor/internal/core/domain"
)

func TestReviewChangeReturnsApprovedForCompleteChange(t *testing.T) {
	changeID := "implement-review-foundation"
	fileSystem := newFakeReviewFileSystem()
	seedReviewOpenSpecProject(fileSystem)
	seedReviewCompleteChange(fileSystem, changeID, "- [x] Write tests\n- [X] Run tests\n")

	result, err := NewReviewChange(fileSystem).Execute(ReviewChangeInput{
		ProjectRoot: "/project",
		ChangeID:    changeID,
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if result.Status != domain.ReviewStatusApproved {
		t.Fatalf("Status = %q, want %q", result.Status, domain.ReviewStatusApproved)
	}
	if result.ChangeID != changeID {
		t.Fatalf("ChangeID = %q, want %q", result.ChangeID, changeID)
	}
	if result.CheckedPath != "openspec/changes/"+changeID {
		t.Fatalf("CheckedPath = %q, want openspec/changes/%s", result.CheckedPath, changeID)
	}
	if result.Tasks.Total != 2 || result.Tasks.Completed != 2 || result.Tasks.Incomplete != 0 {
		t.Fatalf("Tasks = %+v, want 2 total, 2 completed, 0 incomplete", result.Tasks)
	}
	if len(result.RequiredFiles) != len(domain.RequiredOpenSpecChangeFiles()) {
		t.Fatalf("RequiredFiles count = %d, want %d", len(result.RequiredFiles), len(domain.RequiredOpenSpecChangeFiles()))
	}
	if len(result.Findings) != 0 {
		t.Fatalf("Findings count = %d, want 0", len(result.Findings))
	}
}

func TestReviewChangeReturnsNeedsWorkForIncompleteTasks(t *testing.T) {
	changeID := "needs-work"
	fileSystem := newFakeReviewFileSystem()
	seedReviewOpenSpecProject(fileSystem)
	seedReviewCompleteChange(fileSystem, changeID, "- [x] Done\n- [ ] Run go test ./...\n- [ ] Update tasks.md\n")

	result, err := NewReviewChange(fileSystem).Execute(ReviewChangeInput{
		ProjectRoot: "/project",
		ChangeID:    changeID,
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if result.Status != domain.ReviewStatusNeedsWork {
		t.Fatalf("Status = %q, want %q", result.Status, domain.ReviewStatusNeedsWork)
	}
	if result.Tasks.Total != 3 || result.Tasks.Completed != 1 || result.Tasks.Incomplete != 2 {
		t.Fatalf("Tasks = %+v, want 3 total, 1 completed, 2 incomplete", result.Tasks)
	}
	if len(result.Findings) != 2 {
		t.Fatalf("Findings count = %d, want 2", len(result.Findings))
	}
	for _, finding := range result.Findings {
		if finding.Severity != domain.ReviewFindingSeverityWarning {
			t.Fatalf("finding severity = %q, want %q", finding.Severity, domain.ReviewFindingSeverityWarning)
		}
		if finding.Code != domain.ReviewFindingCodeIncompleteTask {
			t.Fatalf("finding code = %q, want %q", finding.Code, domain.ReviewFindingCodeIncompleteTask)
		}
	}
	if result.Findings[0].Message != "Task is not completed: Run go test ./..." {
		t.Fatalf("first finding message = %q", result.Findings[0].Message)
	}
	if result.Findings[1].Message != "Task is not completed: Update tasks.md" {
		t.Fatalf("second finding message = %q", result.Findings[1].Message)
	}
}

func TestReviewChangeReturnsFallbackFindingMessageForIncompleteTaskWithoutText(t *testing.T) {
	changeID := "blank-incomplete-task"
	fileSystem := newFakeReviewFileSystem()
	seedReviewOpenSpecProject(fileSystem)
	seedReviewCompleteChange(fileSystem, changeID, "- [ ]\n")

	result, err := NewReviewChange(fileSystem).Execute(ReviewChangeInput{
		ProjectRoot: "/project",
		ChangeID:    changeID,
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if result.Status != domain.ReviewStatusNeedsWork {
		t.Fatalf("Status = %q, want %q", result.Status, domain.ReviewStatusNeedsWork)
	}
	if result.Tasks.Total != 1 || result.Tasks.Completed != 0 || result.Tasks.Incomplete != 1 {
		t.Fatalf("Tasks = %+v, want 1 total, 0 completed, 1 incomplete", result.Tasks)
	}
	assertReviewSingleFindingCode(t, result, domain.ReviewFindingCodeIncompleteTask)
	if result.Findings[0].Message != "Task is not completed." {
		t.Fatalf("finding message = %q, want fallback incomplete task message", result.Findings[0].Message)
	}
}

func TestReviewChangeReturnsInvalidForMissingOpenSpecProjectStructure(t *testing.T) {
	tests := []struct {
		name  string
		setup func(fileSystem *fakeReviewFileSystem)
	}{
		{
			name:  "missing project file",
			setup: func(fileSystem *fakeReviewFileSystem) { fileSystem.directories[openspecChangesDirectory] = true },
		},
		{
			name:  "missing changes directory",
			setup: func(fileSystem *fakeReviewFileSystem) { fileSystem.files[openspecProjectFile] = "project" },
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fileSystem := newFakeReviewFileSystem()
			test.setup(fileSystem)

			result, err := NewReviewChange(fileSystem).Execute(ReviewChangeInput{
				ProjectRoot: "/project",
				ChangeID:    "change",
			})
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			assertReviewSingleFindingCode(t, result, domain.ReviewFindingCodeProjectRootUnavailable)
			if result.Status != domain.ReviewStatusInvalid {
				t.Fatalf("Status = %q, want %q", result.Status, domain.ReviewStatusInvalid)
			}
			if fileSystem.readCount != 0 {
				t.Fatalf("ReadFile calls = %d, want 0", fileSystem.readCount)
			}
		})
	}
}

func TestReviewChangeReturnsInvalidForMissingChangeDirectory(t *testing.T) {
	fileSystem := newFakeReviewFileSystem()
	seedReviewOpenSpecProject(fileSystem)

	result, err := NewReviewChange(fileSystem).Execute(ReviewChangeInput{
		ProjectRoot: "/project",
		ChangeID:    "missing-change",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	assertReviewSingleFindingCode(t, result, domain.ReviewFindingCodeChangeDirectoryMissing)
	if result.Findings[0].RelativePath != "openspec/changes/missing-change" {
		t.Fatalf("RelativePath = %q, want openspec/changes/missing-change", result.Findings[0].RelativePath)
	}
	if fileSystem.readCount != 0 {
		t.Fatalf("ReadFile calls = %d, want 0", fileSystem.readCount)
	}
}

func TestReviewChangeReturnsInvalidForNonDirectoryChangePath(t *testing.T) {
	changeID := "file-change"
	checkedPath := openspecChangesDirectory + "/" + changeID
	fileSystem := newFakeReviewFileSystem()
	seedReviewOpenSpecProject(fileSystem)
	fileSystem.files[checkedPath] = "not a directory"

	result, err := NewReviewChange(fileSystem).Execute(ReviewChangeInput{
		ProjectRoot: "/project",
		ChangeID:    changeID,
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	assertReviewSingleFindingCode(t, result, domain.ReviewFindingCodeChangeDirectoryMissing)
}

func TestReviewChangeReturnsInvalidForMissingRequiredFiles(t *testing.T) {
	changeID := "missing-files"
	checkedPath := openspecChangesDirectory + "/" + changeID
	fileSystem := newFakeReviewFileSystem()
	seedReviewOpenSpecProject(fileSystem)
	fileSystem.directories[checkedPath] = true
	fileSystem.files[checkedPath+"/design.md"] = "design"
	fileSystem.files[checkedPath+"/tasks.md"] = "- [x] Done\n"
	fileSystem.files[checkedPath+"/acceptance-criteria.md"] = "acceptance"

	result, err := NewReviewChange(fileSystem).Execute(ReviewChangeInput{
		ProjectRoot: "/project",
		ChangeID:    changeID,
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if result.Status != domain.ReviewStatusInvalid {
		t.Fatalf("Status = %q, want %q", result.Status, domain.ReviewStatusInvalid)
	}
	if got := reviewFindingSubjectsByCode(result, domain.ReviewFindingCodeRequiredFileMissing); strings.Join(got, ",") != "proposal.md,risks.md" {
		t.Fatalf("missing file findings = %v, want [proposal.md risks.md]", got)
	}
	if fileSystem.readCount != 0 {
		t.Fatalf("ReadFile calls = %d, want 0", fileSystem.readCount)
	}
}

func TestReviewChangeChecksSharedRequiredOpenSpecChangeFiles(t *testing.T) {
	changeID := "required-policy"
	fileSystem := newFakeReviewFileSystem()
	seedReviewOpenSpecProject(fileSystem)
	seedReviewCompleteChange(fileSystem, changeID, "- [x] Done\n")

	_, err := NewReviewChange(fileSystem).Execute(ReviewChangeInput{
		ProjectRoot: "/project",
		ChangeID:    changeID,
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	checkedPath := openspecChangesDirectory + "/" + changeID
	for _, requiredFile := range domain.RequiredOpenSpecChangeFiles() {
		if !fileSystem.checkedFiles[checkedPath+"/"+requiredFile] {
			t.Fatalf("required file %q was not checked", requiredFile)
		}
	}
}

func TestReviewChangeReturnsInvalidWhenTasksMarkdownIsMissing(t *testing.T) {
	changeID := "missing-tasks"
	checkedPath := openspecChangesDirectory + "/" + changeID
	fileSystem := newFakeReviewFileSystem()
	seedReviewOpenSpecProject(fileSystem)
	fileSystem.directories[checkedPath] = true
	for _, requiredFile := range domain.RequiredOpenSpecChangeFiles() {
		if requiredFile == "tasks.md" {
			continue
		}
		fileSystem.files[checkedPath+"/"+requiredFile] = requiredFile
	}

	result, err := NewReviewChange(fileSystem).Execute(ReviewChangeInput{
		ProjectRoot: "/project",
		ChangeID:    changeID,
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	assertReviewSingleFindingCode(t, result, domain.ReviewFindingCodeRequiredFileMissing)
	if result.Findings[0].Subject != "tasks.md" {
		t.Fatalf("finding Subject = %q, want tasks.md", result.Findings[0].Subject)
	}
	if fileSystem.readCount != 0 {
		t.Fatalf("ReadFile calls = %d, want 0", fileSystem.readCount)
	}
}

func TestReviewChangeReturnsInvalidForUnreadableTasksMarkdown(t *testing.T) {
	changeID := "unreadable-tasks"
	checkedPath := openspecChangesDirectory + "/" + changeID
	tasksPath := checkedPath + "/tasks.md"
	fileSystem := newFakeReviewFileSystem()
	seedReviewOpenSpecProject(fileSystem)
	seedReviewCompleteChange(fileSystem, changeID, "- [x] Done\n")
	fileSystem.readErrors[tasksPath] = errors.New("permission denied")

	result, err := NewReviewChange(fileSystem).Execute(ReviewChangeInput{
		ProjectRoot: "/project",
		ChangeID:    changeID,
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	assertReviewSingleFindingCode(t, result, domain.ReviewFindingCodeTasksFileUnreadable)
	if result.Status != domain.ReviewStatusInvalid {
		t.Fatalf("Status = %q, want %q", result.Status, domain.ReviewStatusInvalid)
	}
}

func TestReviewChangeReturnsInvalidWhenNoTasksAreFound(t *testing.T) {
	changeID := "no-tasks"
	fileSystem := newFakeReviewFileSystem()
	seedReviewOpenSpecProject(fileSystem)
	seedReviewCompleteChange(fileSystem, changeID, "# Tasks\n\nNo checkbox lines here.\n")

	result, err := NewReviewChange(fileSystem).Execute(ReviewChangeInput{
		ProjectRoot: "/project",
		ChangeID:    changeID,
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	assertReviewSingleFindingCode(t, result, domain.ReviewFindingCodeTasksNotFound)
	if result.Tasks.Total != 0 {
		t.Fatalf("Tasks.Total = %d, want 0", result.Tasks.Total)
	}
}

func TestReviewChangeRejectsInvalidInputBeforeFilesystemOperations(t *testing.T) {
	tests := []struct {
		name  string
		input ReviewChangeInput
		want  string
	}{
		{
			name:  "empty project root",
			input: ReviewChangeInput{ProjectRoot: " ", ChangeID: "change"},
			want:  "project root is required",
		},
		{
			name:  "empty change id",
			input: ReviewChangeInput{ProjectRoot: "/project", ChangeID: " "},
			want:  "change id is required",
		},
		{
			name:  "dot id",
			input: ReviewChangeInput{ProjectRoot: "/project", ChangeID: "."},
			want:  "change id must be a safe single path segment",
		},
		{
			name:  "dotdot id",
			input: ReviewChangeInput{ProjectRoot: "/project", ChangeID: ".."},
			want:  "change id must be a safe single path segment",
		},
		{
			name:  "traversal id",
			input: ReviewChangeInput{ProjectRoot: "/project", ChangeID: "../unsafe"},
			want:  "change id must be a safe single path segment",
		},
		{
			name:  "absolute id",
			input: ReviewChangeInput{ProjectRoot: "/project", ChangeID: "/unsafe"},
			want:  "change id must be a safe single path segment",
		},
		{
			name:  "slash id",
			input: ReviewChangeInput{ProjectRoot: "/project", ChangeID: "bad/id"},
			want:  "change id must be a safe single path segment",
		},
		{
			name:  "backslash id",
			input: ReviewChangeInput{ProjectRoot: "/project", ChangeID: `bad\id`},
			want:  "change id must be a safe single path segment",
		},
		{
			name:  "colon id",
			input: ReviewChangeInput{ProjectRoot: "/project", ChangeID: "bad:id"},
			want:  "change id must be a safe single path segment",
		},
		{
			name:  "leading dash id",
			input: ReviewChangeInput{ProjectRoot: "/project", ChangeID: "-bad"},
			want:  "change id must be a safe single path segment",
		},
		{
			name:  "embedded dotdot id",
			input: ReviewChangeInput{ProjectRoot: "/project", ChangeID: "bad..id"},
			want:  "change id must be a safe single path segment",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fileSystem := newFakeReviewFileSystem()
			_, err := NewReviewChange(fileSystem).Execute(test.input)
			if err == nil {
				t.Fatalf("Execute() error = nil, want %q", test.want)
			}
			if !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Execute() error = %q, want %q", err.Error(), test.want)
			}
			if fileSystem.operationCount() != 0 {
				t.Fatalf("filesystem operations = %d, want 0", fileSystem.operationCount())
			}
		})
	}
}

func TestReviewChangeReturnsFilesystemExecutionErrors(t *testing.T) {
	wantErr := errors.New("filesystem unavailable")
	tests := []struct {
		name  string
		setup func(fileSystem *fakeReviewFileSystem)
	}{
		{
			name: "project file check",
			setup: func(fileSystem *fakeReviewFileSystem) {
				fileSystem.fileErrors[openspecProjectFile] = wantErr
			},
		},
		{
			name: "changes directory check",
			setup: func(fileSystem *fakeReviewFileSystem) {
				fileSystem.files[openspecProjectFile] = "project"
				fileSystem.directoryErrors[openspecChangesDirectory] = wantErr
			},
		},
		{
			name: "change directory check",
			setup: func(fileSystem *fakeReviewFileSystem) {
				seedReviewOpenSpecProject(fileSystem)
				fileSystem.directoryErrors[openspecChangesDirectory+"/change"] = wantErr
			},
		},
		{
			name: "required file check",
			setup: func(fileSystem *fakeReviewFileSystem) {
				seedReviewOpenSpecProject(fileSystem)
				fileSystem.directories[openspecChangesDirectory+"/change"] = true
				fileSystem.fileErrors[openspecChangesDirectory+"/change/proposal.md"] = wantErr
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fileSystem := newFakeReviewFileSystem()
			test.setup(fileSystem)

			_, err := NewReviewChange(fileSystem).Execute(ReviewChangeInput{
				ProjectRoot: "/project",
				ChangeID:    "change",
			})
			if err == nil {
				t.Fatalf("Execute() error = nil, want filesystem error")
			}
			if !errors.Is(err, wantErr) {
				t.Fatalf("Execute() error = %v, want wrapping %v", err, wantErr)
			}
		})
	}
}

func TestReviewChangeRejectsMissingDependencies(t *testing.T) {
	_, err := (*ReviewChange)(nil).Execute(ReviewChangeInput{})
	if err == nil || !strings.Contains(err.Error(), "review change use case is required") {
		t.Fatalf("nil use case error = %v, want review change use case is required", err)
	}

	_, err = NewReviewChange(nil).Execute(ReviewChangeInput{})
	if err == nil || !strings.Contains(err.Error(), "review filesystem is required") {
		t.Fatalf("nil filesystem error = %v, want review filesystem is required", err)
	}
}

func TestParseReviewTasksCountsSupportedTaskMarkers(t *testing.T) {
	tasks := parseReviewTasks(strings.Join([]string{
		"# Tasks",
		"- [x] completed lowercase",
		"- [X] completed uppercase",
		"- [ ] incomplete task",
		"  - [ ] nested incomplete",
		"- [o] ignored",
		"plain text ignored",
	}, "\n"))

	if tasks.summary.Total != 4 {
		t.Fatalf("Total = %d, want 4", tasks.summary.Total)
	}
	if tasks.summary.Completed != 2 {
		t.Fatalf("Completed = %d, want 2", tasks.summary.Completed)
	}
	if tasks.summary.Incomplete != 2 {
		t.Fatalf("Incomplete = %d, want 2", tasks.summary.Incomplete)
	}
	if strings.Join(tasks.incompleteTaskTexts, "|") != "incomplete task|nested incomplete" {
		t.Fatalf("incomplete task texts = %v", tasks.incompleteTaskTexts)
	}
}

func TestParseReviewTasksIgnoresMalformedCheckboxes(t *testing.T) {
	tasks := parseReviewTasks(strings.Join([]string{
		"- [] missing checkbox state",
		"- [x completed without closing bracket",
		"- [X completed uppercase without closing bracket",
		"- [ x] malformed completed state",
		"- [x ] malformed completed state",
		"* [x] wrong list marker",
		"1. [ ] ordered list marker",
		"- [o] unsupported checkbox state",
	}, "\n"))

	if tasks.summary.Total != 0 {
		t.Fatalf("Total = %d, want 0", tasks.summary.Total)
	}
	if tasks.summary.Completed != 0 {
		t.Fatalf("Completed = %d, want 0", tasks.summary.Completed)
	}
	if tasks.summary.Incomplete != 0 {
		t.Fatalf("Incomplete = %d, want 0", tasks.summary.Incomplete)
	}
	if len(tasks.incompleteTaskTexts) != 0 {
		t.Fatalf("incomplete task texts = %v, want none", tasks.incompleteTaskTexts)
	}
}

func TestParseReviewTasksIgnoresNonTaskMarkdownLines(t *testing.T) {
	tasks := parseReviewTasks("# Tasks\n\n- not a checkbox\n[x] not a task\ntext - [ ] not at start\n")
	if tasks.summary.Total != 0 {
		t.Fatalf("Total = %d, want 0", tasks.summary.Total)
	}
}

func assertReviewSingleFindingCode(t *testing.T, result domain.ReviewResult, code domain.ReviewFindingCode) {
	t.Helper()

	if len(result.Findings) != 1 {
		t.Fatalf("Findings count = %d, want 1", len(result.Findings))
	}
	if result.Findings[0].Code != code {
		t.Fatalf("finding code = %q, want %q", result.Findings[0].Code, code)
	}
}

func reviewFindingSubjectsByCode(result domain.ReviewResult, code domain.ReviewFindingCode) []string {
	var subjects []string
	for _, finding := range result.Findings {
		if finding.Code == code {
			subjects = append(subjects, finding.Subject)
		}
	}
	return subjects
}

func seedReviewOpenSpecProject(fileSystem *fakeReviewFileSystem) {
	fileSystem.files[openspecProjectFile] = "project"
	fileSystem.directories[openspecChangesDirectory] = true
}

func seedReviewCompleteChange(fileSystem *fakeReviewFileSystem, changeID string, tasksContents string) {
	checkedPath := openspecChangesDirectory + "/" + changeID
	fileSystem.directories[checkedPath] = true
	for _, requiredFile := range domain.RequiredOpenSpecChangeFiles() {
		fileSystem.files[checkedPath+"/"+requiredFile] = requiredFile
	}
	fileSystem.files[checkedPath+"/tasks.md"] = tasksContents
}

type fakeReviewFileSystem struct {
	directories     map[string]bool
	files           map[string]string
	directoryErrors map[string]error
	fileErrors      map[string]error
	readErrors      map[string]error
	checkedFiles    map[string]bool
	directoryChecks int
	fileChecks      int
	readCount       int
}

func newFakeReviewFileSystem() *fakeReviewFileSystem {
	return &fakeReviewFileSystem{
		directories:     make(map[string]bool),
		files:           make(map[string]string),
		directoryErrors: make(map[string]error),
		fileErrors:      make(map[string]error),
		readErrors:      make(map[string]error),
		checkedFiles:    make(map[string]bool),
	}
}

func (fileSystem *fakeReviewFileSystem) DirectoryExists(_ string, relativePath string) (bool, error) {
	fileSystem.directoryChecks++
	if err := fileSystem.directoryErrors[relativePath]; err != nil {
		return false, err
	}
	return fileSystem.directories[relativePath], nil
}

func (fileSystem *fakeReviewFileSystem) FileExists(_ string, relativePath string) (bool, error) {
	fileSystem.fileChecks++
	fileSystem.checkedFiles[relativePath] = true
	if err := fileSystem.fileErrors[relativePath]; err != nil {
		return false, err
	}
	_, exists := fileSystem.files[relativePath]
	return exists, nil
}

func (fileSystem *fakeReviewFileSystem) ReadFile(_ string, relativePath string) (string, error) {
	fileSystem.readCount++
	if err := fileSystem.readErrors[relativePath]; err != nil {
		return "", err
	}
	return fileSystem.files[relativePath], nil
}

func (fileSystem *fakeReviewFileSystem) operationCount() int {
	return fileSystem.directoryChecks + fileSystem.fileChecks + fileSystem.readCount
}
