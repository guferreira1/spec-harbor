package domain

import (
	"fmt"
	"regexp"
	"strings"
)

// ChangeFileContent carries the loaded content of one required change file so
// validation rules can run as pure functions over strings.
type ChangeFileContent struct {
	FileName     string
	RelativePath string
	Content      string
}

// canonicalStarterMarkers is the domain-owned source of truth for
// boilerplate-only detection. Adapter template files may share this wording,
// but they are never read at validation time; changing starter wording
// requires an intentional update here, guarded by domain tests and the
// adapter-layer drift test.
var canonicalStarterMarkers = map[string]struct{}{
	"Describe the problem this change should solve and who is affected.":                  {},
	"Describe the outcome this change should deliver.":                                    {},
	"List the behavior, files, or interfaces included in this change.":                    {},
	"List related work that should not be implemented in this change.":                    {},
	"Describe how reviewers can tell the change is complete.":                             {},
	"Describe the proposed approach at a high level.":                                     {},
	"Describe the affected layers, boundaries, and dependencies.":                         {},
	"Record important implementation choices and tradeoffs.":                              {},
	"Describe the tests needed to cover the change.":                                      {},
	"List the commands or checks that should pass before completion.":                     {},
	"Read the project context, architecture spec, and active OpenSpec change.":            {},
	"Keep implementation limited to the approved scope.":                                  {},
	"Add or update domain, use case, port, adapter, CLI, and test code as needed.":        {},
	"Run required formatting and verification commands.":                                  {},
	"Update this task list only after implementation work is complete.":                   {},
	"The requested behavior is implemented within the approved scope.":                    {},
	"Existing behavior outside the scope remains unchanged.":                              {},
	"Errors are clear and actionable for users.":                                          {},
	"Automated tests cover the important success and failure paths.":                      {},
	"Required verification commands pass.":                                                {},
	"Identify technical, product, security, or delivery risks introduced by this change.": {},
	"Describe how each risk will be reduced, tested, or monitored.":                       {},
}

var (
	placeholderTokenPattern  = regexp.MustCompile(`\b(TBD|TODO|FIXME)\b`)
	loremIpsumPattern        = regexp.MustCompile(`(?i)lorem ipsum`)
	orderedListItemPattern   = regexp.MustCompile(`^\d+\. (.*)$`)
	checkboxItemPattern      = regexp.MustCompile(`^[-*] \[[ xX]\] (.*)$`)
	checkboxCandidatePattern = regexp.MustCompile(`^[-*]\s*\[`)
	validCheckboxPattern     = regexp.MustCompile(`^[-*] \[[ xX]\] \S`)
	checkedCheckboxPattern   = regexp.MustCompile(`^[-*] \[[xX]\] \S`)
	whitespaceRunPattern     = regexp.MustCompile(`\s+`)
)

// ValidateChangeFileContents runs the ordered per-file rule chain over every
// loaded file, then the cross-file rules over all loaded files together.
func ValidateChangeFileContents(files []ChangeFileContent) []ValidationFinding {
	var findings []ValidationFinding
	for _, file := range files {
		findings = append(findings, validateChangeFileContent(file)...)
	}
	findings = append(findings, validateCrossFileRules(files)...)
	return findings
}

type changeFileContentRule func(file ChangeFileContent, lines []string) []ValidationFinding

// changeFileContentRules is the ordered per-file rule chain. New rules are
// added by appending an entry.
var changeFileContentRules = []changeFileContentRule{
	ruleFileMissingHeading,
	ruleFileMissingBody,
	rulePlaceholderContent,
	ruleBoilerplateOnlyContent,
	ruleProposalSections,
	ruleDesignSections,
	ruleTasksCheckboxes,
	ruleAcceptanceCriteriaItems,
	ruleRisksMitigation,
}

func validateChangeFileContent(file ChangeFileContent) []ValidationFinding {
	// file_empty suppresses every other content rule for the same file.
	if strings.TrimSpace(file.Content) == "" {
		return []ValidationFinding{newFileFinding(file, ValidationFindingSeverityError, ValidationFindingCodeFileEmpty, "File is empty.")}
	}

	lines := contentLines(file.Content)
	var findings []ValidationFinding
	for _, rule := range changeFileContentRules {
		findings = append(findings, rule(file, lines)...)
	}
	return findings
}

func ruleFileMissingHeading(file ChangeFileContent, lines []string) []ValidationFinding {
	for _, line := range lines {
		if headingLevel(line) > 0 {
			return nil
		}
	}
	return []ValidationFinding{newFileFinding(file, ValidationFindingSeverityError, ValidationFindingCodeFileMissingHeading, "No markdown heading found.")}
}

func ruleFileMissingBody(file ChangeFileContent, lines []string) []ValidationFinding {
	for _, line := range lines {
		if isMeaningfulLine(line) {
			return nil
		}
	}
	return []ValidationFinding{newFileFinding(file, ValidationFindingSeverityError, ValidationFindingCodeFileMissingBody, "No body content found beyond headings.")}
}

func rulePlaceholderContent(file ChangeFileContent, lines []string) []ValidationFinding {
	for index, line := range lines {
		marker, found := placeholderMarkerOnLine(line)
		if !found {
			continue
		}
		message := fmt.Sprintf("Placeholder marker %q found (line %d)", marker, index+1)
		return []ValidationFinding{newFileFinding(file, ValidationFindingSeverityWarning, ValidationFindingCodePlaceholderContent, message)}
	}
	return nil
}

func ruleBoilerplateOnlyContent(file ChangeFileContent, lines []string) []ValidationFinding {
	meaningfulLines := 0
	for _, line := range lines {
		if !isMeaningfulLine(line) {
			continue
		}
		meaningfulLines++
		if _, known := canonicalStarterMarkers[normalizeStarterLine(line)]; !known {
			return nil
		}
	}
	if meaningfulLines == 0 {
		return nil
	}
	return []ValidationFinding{newFileFinding(file, ValidationFindingSeverityWarning, ValidationFindingCodeBoilerplateOnlyContent, "Only starter boilerplate content found.")}
}

func ruleProposalSections(file ChangeFileContent, lines []string) []ValidationFinding {
	if file.FileName != "proposal.md" {
		return nil
	}
	if hasSectionHeadingWithPrefix(lines, "problem", "goal", "summary") {
		return nil
	}
	return []ValidationFinding{newFileFinding(file, ValidationFindingSeverityWarning, ValidationFindingCodeProposalSectionMissing, "No Problem, Goal, or Summary section found.")}
}

func ruleDesignSections(file ChangeFileContent, lines []string) []ValidationFinding {
	if file.FileName != "design.md" {
		return nil
	}
	if hasSectionHeadingWithPrefix(lines, "overview", "approach", "design", "architecture", "technical decisions", "decisions") {
		return nil
	}
	return []ValidationFinding{newFileFinding(file, ValidationFindingSeverityWarning, ValidationFindingCodeDesignSectionMissing, "No Overview, Approach, Design, Architecture, Technical Decisions, or Decisions section found.")}
}

func ruleTasksCheckboxes(file ChangeFileContent, lines []string) []ValidationFinding {
	if file.FileName != "tasks.md" {
		return nil
	}

	var findings []ValidationFinding
	validTasks := 0
	checkedTasks := 0
	hasPhaseHeading := false

	for index, line := range lines {
		if headingLevel(line) == 2 {
			hasPhaseHeading = true
		}

		trimmed := strings.TrimSpace(line)
		if !checkboxCandidatePattern.MatchString(trimmed) {
			continue
		}
		if !validCheckboxPattern.MatchString(trimmed) {
			message := fmt.Sprintf("Malformed checkbox task (line %d)", index+1)
			findings = append(findings, newFileFinding(file, ValidationFindingSeverityError, ValidationFindingCodeTasksCheckboxMalformed, message))
			continue
		}
		validTasks++
		if checkedCheckboxPattern.MatchString(trimmed) {
			checkedTasks++
		}
	}

	if validTasks == 0 {
		findings = append(findings, newFileFinding(file, ValidationFindingSeverityError, ValidationFindingCodeTasksCheckboxMissing, "No checkbox tasks found."))
		return findings
	}
	if !hasPhaseHeading {
		findings = append(findings, newFileFinding(file, ValidationFindingSeverityWarning, ValidationFindingCodeTasksPhaseHeadingMissing, "No level-2 phase heading found."))
	}
	if checkedTasks == validTasks {
		findings = append(findings, newFileFinding(file, ValidationFindingSeverityWarning, ValidationFindingCodeTasksAllCompleted, "All checkbox tasks are completed; confirm implementation evidence before review."))
	}
	return findings
}

func ruleAcceptanceCriteriaItems(file ChangeFileContent, lines []string) []ValidationFinding {
	if file.FileName != "acceptance-criteria.md" {
		return nil
	}
	for _, line := range lines {
		itemText, isItem := listItemText(line)
		if !isItem {
			continue
		}
		text := strings.TrimSpace(itemText)
		if text != "" && !isPlaceholderOnlyText(text) {
			return nil
		}
	}
	return []ValidationFinding{newFileFinding(file, ValidationFindingSeverityError, ValidationFindingCodeAcceptanceCriteriaItemMissing, "No meaningful acceptance criteria items found.")}
}

func ruleRisksMitigation(file ChangeFileContent, lines []string) []ValidationFinding {
	if file.FileName != "risks.md" {
		return nil
	}

	hasBody := false
	for _, line := range lines {
		if isMeaningfulLine(line) {
			hasBody = true
		}
		if strings.Contains(strings.ToLower(line), "mitigation") {
			return nil
		}
	}
	if !hasBody {
		return nil
	}
	return []ValidationFinding{newFileFinding(file, ValidationFindingSeverityWarning, ValidationFindingCodeRisksMitigationMissing, "Risks are listed without mitigation notes.")}
}

// validateCrossFileRules runs after all available files are loaded. Missing
// and empty files are skipped: file_empty already reports empty files.
func validateCrossFileRules(files []ChangeFileContent) []ValidationFinding {
	loaded := make(map[string]ChangeFileContent)
	for _, file := range files {
		if strings.TrimSpace(file.Content) == "" {
			continue
		}
		loaded[file.FileName] = file
	}

	var findings []ValidationFinding

	if design, ok := loaded["design.md"]; ok && anyFileMentionsInternalPackages(loaded) && !hasHeadingContaining(contentLines(design.Content), "architecture") {
		findings = append(findings, newFileFinding(design, ValidationFindingSeverityWarning, ValidationFindingCodeDesignArchitectureSectionMissing, "Change files mention internal architecture packages but design.md has no Architecture section."))
	}

	if tasks, ok := loaded["tasks.md"]; ok && anyFileMentionsCLICommand(loaded) && !hasDocumentationTaskLine(contentLines(tasks.Content)) {
		findings = append(findings, newFileFinding(tasks, ValidationFindingSeverityWarning, ValidationFindingCodeTasksDocumentationTaskMissing, "CLI behavior is referenced but tasks.md has no documentation task."))
	}

	return findings
}

func anyFileMentionsInternalPackages(loaded map[string]ChangeFileContent) bool {
	for _, file := range loaded {
		for _, packagePath := range []string{"internal/core", "internal/adapters", "internal/platform"} {
			if strings.Contains(file.Content, packagePath) {
				return true
			}
		}
	}
	return false
}

func anyFileMentionsCLICommand(loaded map[string]ChangeFileContent) bool {
	for _, fileName := range []string{"proposal.md", "design.md"} {
		if file, ok := loaded[fileName]; ok && strings.Contains(file.Content, "specharbor") {
			return true
		}
	}
	return false
}

func hasDocumentationTaskLine(lines []string) bool {
	for _, line := range lines {
		lowered := strings.ToLower(line)
		if strings.Contains(lowered, "doc") || strings.Contains(lowered, "readme") {
			return true
		}
	}
	return false
}

func placeholderMarkerOnLine(line string) (string, bool) {
	if headingLevel(line) > 0 {
		return "", false
	}
	if marker := placeholderTokenPattern.FindString(line); marker != "" {
		return marker, true
	}
	if loremIpsumPattern.MatchString(line) {
		return "lorem ipsum", true
	}
	if itemText, isItem := listItemText(line); isItem {
		text := strings.TrimSpace(itemText)
		if text == "N/A" || text == "..." || text == "?" {
			return text, true
		}
	}
	return "", false
}

func isPlaceholderOnlyText(text string) bool {
	switch text {
	case "N/A", "...", "?", "TBD", "TODO", "FIXME":
		return true
	default:
		return false
	}
}

func listItemText(line string) (string, bool) {
	trimmed := strings.TrimSpace(line)
	if match := checkboxItemPattern.FindStringSubmatch(trimmed); match != nil {
		return match[1], true
	}
	if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") {
		return trimmed[2:], true
	}
	if match := orderedListItemPattern.FindStringSubmatch(trimmed); match != nil {
		return match[1], true
	}
	return "", false
}

func normalizeStarterLine(line string) string {
	trimmed := strings.TrimSpace(line)
	if itemText, isItem := listItemText(trimmed); isItem {
		trimmed = strings.TrimSpace(itemText)
	}
	return whitespaceRunPattern.ReplaceAllString(trimmed, " ")
}

func contentLines(content string) []string {
	return strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
}

func headingLevel(line string) int {
	trimmed := strings.TrimSpace(line)
	level := 0
	for _, character := range trimmed {
		if character != '#' {
			break
		}
		level++
	}
	if level == 0 || level > 6 {
		return 0
	}
	rest := trimmed[level:]
	if rest != "" && !strings.HasPrefix(rest, " ") {
		return 0
	}
	return level
}

func isMeaningfulLine(line string) bool {
	return strings.TrimSpace(line) != "" && headingLevel(line) == 0
}

func hasSectionHeadingWithPrefix(lines []string, prefixes ...string) bool {
	for _, line := range lines {
		if headingLevel(line) < 2 {
			continue
		}
		text := strings.ToLower(headingText(line))
		for _, prefix := range prefixes {
			if strings.HasPrefix(text, prefix) {
				return true
			}
		}
	}
	return false
}

func hasHeadingContaining(lines []string, fragment string) bool {
	for _, line := range lines {
		if headingLevel(line) == 0 {
			continue
		}
		if strings.Contains(strings.ToLower(headingText(line)), fragment) {
			return true
		}
	}
	return false
}

func headingText(line string) string {
	trimmed := strings.TrimSpace(line)
	return strings.TrimSpace(strings.TrimLeft(trimmed, "#"))
}

func newFileFinding(file ChangeFileContent, severity ValidationFindingSeverity, code ValidationFindingCode, message string) ValidationFinding {
	return ValidationFinding{
		Severity:     severity,
		Code:         code,
		Message:      message,
		RelativePath: file.RelativePath,
		Subject:      file.FileName,
	}
}
