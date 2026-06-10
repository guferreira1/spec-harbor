package domain

import (
	"fmt"
	"strings"
)

const (
	aiOutputFileStartPrefix = "---FILE: "
	aiOutputFileStartSuffix = "---"
	aiOutputFileEndLine     = "---END FILE---"
)

type AIOutputParseFindingSeverity string

const (
	AIOutputParseFindingSeverityError AIOutputParseFindingSeverity = "error"
)

type AIOutputParseFindingCode string

const (
	AIOutputParseFindingCodeUnknownFileBlock     AIOutputParseFindingCode = "unknown_file_block"
	AIOutputParseFindingCodeDuplicateFileBlock   AIOutputParseFindingCode = "duplicate_file_block"
	AIOutputParseFindingCodeMissingFileBlock     AIOutputParseFindingCode = "missing_file_block"
	AIOutputParseFindingCodeEmptyFileContent     AIOutputParseFindingCode = "empty_file_content"
	AIOutputParseFindingCodePathTraversalName    AIOutputParseFindingCode = "path_traversal_file_name"
	AIOutputParseFindingCodeAbsoluteFileName     AIOutputParseFindingCode = "absolute_file_name"
	AIOutputParseFindingCodeNestedFileName       AIOutputParseFindingCode = "nested_file_name"
	AIOutputParseFindingCodeMalformedFileBlock   AIOutputParseFindingCode = "malformed_file_block"
	AIOutputParseFindingCodeTextOutsideFileBlock AIOutputParseFindingCode = "text_outside_file_block"
	AIOutputParseFindingCodeUnclosedFileBlock    AIOutputParseFindingCode = "unclosed_file_block"
)

type AIOutputFileBlock struct {
	FileName string
	Content  string
}

type AIOutputParseFinding struct {
	Severity AIOutputParseFindingSeverity
	Code     AIOutputParseFindingCode
	Message  string
	FileName string
	Line     int
}

type AIOutputParseResult struct {
	files    []AIOutputFileBlock
	findings []AIOutputParseFinding
}

func NewAIOutputParseResult(files []AIOutputFileBlock, findings []AIOutputParseFinding) AIOutputParseResult {
	return AIOutputParseResult{
		files:    append([]AIOutputFileBlock(nil), files...),
		findings: append([]AIOutputParseFinding(nil), findings...),
	}
}

func (result AIOutputParseResult) Files() []AIOutputFileBlock {
	return append([]AIOutputFileBlock(nil), result.files...)
}

func (result AIOutputParseResult) Findings() []AIOutputParseFinding {
	return append([]AIOutputParseFinding(nil), result.findings...)
}

func (result AIOutputParseResult) HasErrors() bool {
	return len(result.findings) > 0
}

func (result AIOutputParseResult) ContentFor(fileName string) (string, bool) {
	for _, file := range result.files {
		if file.FileName == fileName {
			return file.Content, true
		}
	}
	return "", false
}

type parsedAIOutputBlock struct {
	fileName string
	content  string
	line     int
}

func RequiredAIGeneratedFileNames() []string {
	return RequiredOpenSpecChangeFiles()
}

func IsAllowedAIGeneratedFileName(fileName string) bool {
	for _, requiredFile := range RequiredAIGeneratedFileNames() {
		if fileName == requiredFile {
			return true
		}
	}
	return false
}

func ParseAIOutputBlocks(source string) AIOutputParseResult {
	lines := splitAIOutputLines(source)
	blocks := make(map[string]parsedAIOutputBlock)
	var findings []AIOutputParseFinding

	var currentFileName string
	var currentStartLine int
	var currentContent []string
	currentAccepted := false

	for index, line := range lines {
		lineNumber := index + 1

		if currentFileName == "" {
			fileName, isStart, ok := parseAIOutputStartLine(line)
			if isStart {
				if !ok {
					findings = append(findings, newAIParseFinding(
						AIOutputParseFindingCodeMalformedFileBlock,
						"Malformed file block start.",
						"",
						lineNumber,
					))
					continue
				}

				currentFileName = fileName
				currentStartLine = lineNumber
				currentContent = nil
				currentAccepted = validateAIOutputFileName(fileName, lineNumber, blocks, &findings)
				continue
			}

			if line == aiOutputFileEndLine {
				findings = append(findings, newAIParseFinding(
					AIOutputParseFindingCodeMalformedFileBlock,
					"End marker found outside a file block.",
					"",
					lineNumber,
				))
				continue
			}

			if strings.TrimSpace(line) != "" {
				findings = append(findings, newAIParseFinding(
					AIOutputParseFindingCodeTextOutsideFileBlock,
					"Text outside file blocks is not allowed.",
					"",
					lineNumber,
				))
			}
			continue
		}

		if line == aiOutputFileEndLine {
			if currentAccepted {
				blocks[currentFileName] = parsedAIOutputBlock{
					fileName: currentFileName,
					content:  strings.Join(currentContent, "\n"),
					line:     currentStartLine,
				}
			}

			currentFileName = ""
			currentStartLine = 0
			currentContent = nil
			currentAccepted = false
			continue
		}

		if _, isStart, _ := parseAIOutputStartLine(line); isStart {
			findings = append(findings, newAIParseFinding(
				AIOutputParseFindingCodeMalformedFileBlock,
				"File block start found before the previous block ended.",
				currentFileName,
				lineNumber,
			))
		}

		currentContent = append(currentContent, line)
	}

	if currentFileName != "" {
		findings = append(findings, newAIParseFinding(
			AIOutputParseFindingCodeUnclosedFileBlock,
			"File block is missing an end marker.",
			currentFileName,
			currentStartLine,
		))
	}

	for _, requiredFile := range RequiredAIGeneratedFileNames() {
		block, exists := blocks[requiredFile]
		if !exists {
			findings = append(findings, newAIParseFinding(
				AIOutputParseFindingCodeMissingFileBlock,
				"Missing required file block.",
				requiredFile,
				0,
			))
			continue
		}
		if strings.TrimSpace(block.content) == "" {
			findings = append(findings, newAIParseFinding(
				AIOutputParseFindingCodeEmptyFileContent,
				"Generated file content is empty.",
				requiredFile,
				block.line,
			))
		}
	}

	files := make([]AIOutputFileBlock, 0, len(blocks))
	for _, requiredFile := range RequiredAIGeneratedFileNames() {
		block, exists := blocks[requiredFile]
		if !exists {
			continue
		}
		files = append(files, AIOutputFileBlock{
			FileName: requiredFile,
			Content:  block.content,
		})
	}

	return NewAIOutputParseResult(files, findings)
}

func splitAIOutputLines(source string) []string {
	lines := strings.Split(source, "\n")
	for index, line := range lines {
		lines[index] = strings.TrimSuffix(line, "\r")
	}
	return lines
}

func parseAIOutputStartLine(line string) (string, bool, bool) {
	if !strings.HasPrefix(line, "---FILE") {
		return "", false, false
	}
	if !strings.HasPrefix(line, aiOutputFileStartPrefix) || !strings.HasSuffix(line, aiOutputFileStartSuffix) {
		return "", true, false
	}
	fileName := strings.TrimSuffix(strings.TrimPrefix(line, aiOutputFileStartPrefix), aiOutputFileStartSuffix)
	if fileName == "" {
		return "", true, false
	}
	return fileName, true, true
}

func validateAIOutputFileName(
	fileName string,
	lineNumber int,
	blocks map[string]parsedAIOutputBlock,
	findings *[]AIOutputParseFinding,
) bool {
	if isAbsoluteAIOutputFileName(fileName) {
		*findings = append(*findings, newAIParseFinding(
			AIOutputParseFindingCodeAbsoluteFileName,
			"File block name must be a relative file name.",
			fileName,
			lineNumber,
		))
		return false
	}
	if isTraversalAIOutputFileName(fileName) {
		*findings = append(*findings, newAIParseFinding(
			AIOutputParseFindingCodePathTraversalName,
			"File block name must not contain path traversal.",
			fileName,
			lineNumber,
		))
		return false
	}
	if strings.ContainsAny(fileName, `/\`) {
		*findings = append(*findings, newAIParseFinding(
			AIOutputParseFindingCodeNestedFileName,
			"File block name must not contain path separators.",
			fileName,
			lineNumber,
		))
		return false
	}
	if !IsAllowedAIGeneratedFileName(fileName) {
		*findings = append(*findings, newAIParseFinding(
			AIOutputParseFindingCodeUnknownFileBlock,
			fmt.Sprintf("Unknown file block: %s.", fileName),
			fileName,
			lineNumber,
		))
		return false
	}
	if _, exists := blocks[fileName]; exists {
		*findings = append(*findings, newAIParseFinding(
			AIOutputParseFindingCodeDuplicateFileBlock,
			"Duplicate file block.",
			fileName,
			lineNumber,
		))
		return false
	}
	return true
}

func isAbsoluteAIOutputFileName(fileName string) bool {
	if strings.HasPrefix(fileName, "/") || strings.HasPrefix(fileName, `\`) {
		return true
	}
	return len(fileName) >= 3 && isASCIILetter(fileName[0]) && fileName[1] == ':' && (fileName[2] == '/' || fileName[2] == '\\')
}

func isTraversalAIOutputFileName(fileName string) bool {
	return fileName == ".." ||
		strings.HasPrefix(fileName, "../") ||
		strings.HasPrefix(fileName, `..\`) ||
		strings.Contains(fileName, "/../") ||
		strings.Contains(fileName, `\..\`) ||
		strings.HasSuffix(fileName, "/..") ||
		strings.HasSuffix(fileName, `\..`) ||
		strings.Contains(fileName, "..")
}

func isASCIILetter(value byte) bool {
	return (value >= 'a' && value <= 'z') || (value >= 'A' && value <= 'Z')
}

func newAIParseFinding(
	code AIOutputParseFindingCode,
	message string,
	fileName string,
	line int,
) AIOutputParseFinding {
	return AIOutputParseFinding{
		Severity: AIOutputParseFindingSeverityError,
		Code:     code,
		Message:  message,
		FileName: fileName,
		Line:     line,
	}
}

type AIAssistedGenerationResult struct {
	ChangeID               string
	SourcePath             string
	TargetPath             string
	ChangeDirectoryCreated bool
	Overwrite              bool

	ProviderAPIsCalled     bool
	RemoteAIServicesCalled bool
	AgentCommandsExecuted  bool
	ProductionCodeModified bool
	VCSCommandsRun         bool
	AutomationPerformed    bool

	generatedFiles   []string
	skippedFiles     []string
	overwrittenFiles []string
	validationResult ValidationResult
	validationRan    bool
}

func NewAIAssistedGenerationResult(
	changeID string,
	sourcePath string,
	targetPath string,
	changeDirectoryCreated bool,
	overwrite bool,
	generatedFiles []string,
	skippedFiles []string,
	overwrittenFiles []string,
	validationResult ValidationResult,
) AIAssistedGenerationResult {
	return AIAssistedGenerationResult{
		ChangeID:               changeID,
		SourcePath:             sourcePath,
		TargetPath:             targetPath,
		ChangeDirectoryCreated: changeDirectoryCreated,
		Overwrite:              overwrite,
		generatedFiles:         append([]string(nil), generatedFiles...),
		skippedFiles:           append([]string(nil), skippedFiles...),
		overwrittenFiles:       append([]string(nil), overwrittenFiles...),
		validationResult:       copyValidationResult(validationResult),
		validationRan:          true,
	}
}

func (result AIAssistedGenerationResult) GeneratedFiles() []string {
	return append([]string(nil), result.generatedFiles...)
}

func (result AIAssistedGenerationResult) SkippedFiles() []string {
	return append([]string(nil), result.skippedFiles...)
}

func (result AIAssistedGenerationResult) OverwrittenFiles() []string {
	return append([]string(nil), result.overwrittenFiles...)
}

func (result AIAssistedGenerationResult) ValidationResult() (ValidationResult, bool) {
	if !result.validationRan {
		return ValidationResult{}, false
	}
	return copyValidationResult(result.validationResult), true
}

func copyValidationResult(result ValidationResult) ValidationResult {
	return ValidationResult{
		ChangeID:      result.ChangeID,
		CheckedPath:   result.CheckedPath,
		Status:        result.Status,
		RequiredFiles: append([]string(nil), result.RequiredFiles...),
		Findings:      append([]ValidationFinding(nil), result.Findings...),
	}
}
