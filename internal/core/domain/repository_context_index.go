package domain

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

const (
	RepositoryContextIndexPath          = ".specharbor/context-index.json"
	RepositoryContextIndexSchemaVersion = 1
)

type RepositoryContextIndexFileType string

const (
	RepositoryContextIndexFileTypeMarkdown   RepositoryContextIndexFileType = "markdown"
	RepositoryContextIndexFileTypeJSON       RepositoryContextIndexFileType = "json"
	RepositoryContextIndexFileTypeYAML       RepositoryContextIndexFileType = "yaml"
	RepositoryContextIndexFileTypeXML        RepositoryContextIndexFileType = "xml"
	RepositoryContextIndexFileTypeGradle     RepositoryContextIndexFileType = "gradle"
	RepositoryContextIndexFileTypeTOML       RepositoryContextIndexFileType = "toml"
	RepositoryContextIndexFileTypeText       RepositoryContextIndexFileType = "text"
	RepositoryContextIndexFileTypeDockerfile RepositoryContextIndexFileType = "dockerfile"
	RepositoryContextIndexFileTypeMakefile   RepositoryContextIndexFileType = "makefile"
	RepositoryContextIndexFileTypeGoModule   RepositoryContextIndexFileType = "go_module"
)

type RepositoryContextIndexClassificationHint string

const (
	RepositoryContextIndexHintUserConfirmedContext RepositoryContextIndexClassificationHint = "user_confirmed_context"
	RepositoryContextIndexHintDetectedFact         RepositoryContextIndexClassificationHint = "detected_fact"
	RepositoryContextIndexHintSuggestedAssumption  RepositoryContextIndexClassificationHint = "suggested_assumption"
	RepositoryContextIndexHintInventoryMetadata    RepositoryContextIndexClassificationHint = "inventory_metadata"
)

type RepositoryContextIndexSkipReason string

const (
	RepositoryContextIndexSkipSensitiveFile        RepositoryContextIndexSkipReason = "sensitive_file"
	RepositoryContextIndexSkipGeneratedDirectory   RepositoryContextIndexSkipReason = "generated_directory"
	RepositoryContextIndexSkipSymlink              RepositoryContextIndexSkipReason = "symlink"
	RepositoryContextIndexSkipUnsupportedSource    RepositoryContextIndexSkipReason = "unsupported_source"
	RepositoryContextIndexSkipFileTooLarge         RepositoryContextIndexSkipReason = "file_too_large"
	RepositoryContextIndexSkipMaxFilesReached      RepositoryContextIndexSkipReason = "max_files_reached"
	RepositoryContextIndexSkipMaxTotalBytesReached RepositoryContextIndexSkipReason = "max_total_bytes_reached"
	RepositoryContextIndexSkipUnsafePath           RepositoryContextIndexSkipReason = "unsafe_path"
)

type RepositoryContextIndexStatus string

const (
	RepositoryContextIndexStatusBuilt     RepositoryContextIndexStatus = "built"
	RepositoryContextIndexStatusWritten   RepositoryContextIndexStatus = "written"
	RepositoryContextIndexStatusCurrent   RepositoryContextIndexStatus = "current"
	RepositoryContextIndexStatusStale     RepositoryContextIndexStatus = "stale"
	RepositoryContextIndexStatusMissing   RepositoryContextIndexStatus = "missing"
	RepositoryContextIndexStatusInvalid   RepositoryContextIndexStatus = "invalid"
	RepositoryContextIndexStatusTruncated RepositoryContextIndexStatus = "truncated"
)

type RepositoryContextIndexMode string

const (
	RepositoryContextIndexModeReport RepositoryContextIndexMode = "report"
	RepositoryContextIndexModeWrite  RepositoryContextIndexMode = "write"
	RepositoryContextIndexModeCheck  RepositoryContextIndexMode = "check"
)

type RepositoryContextIndexLimits struct {
	MaxIndexedFiles   int   `json:"max_indexed_files"`
	MaxFileSizeBytes  int64 `json:"max_file_size_bytes"`
	MaxTotalFileBytes int64 `json:"max_total_file_bytes"`
	MaxSkippedRecords int   `json:"max_skipped_records"`
	MaxDirectoryDepth int   `json:"max_directory_depth"`
}

type RepositoryContextIndexGeneration struct {
	Mode string `json:"mode"`
	Tool string `json:"tool"`
}

type RepositoryContextIndexProject struct {
	RootMarker string `json:"root_marker"`
}

type RepositoryContextIndexEntry struct {
	Path                   string                                     `json:"path"`
	SourceCategory         ContextSourceCategory                      `json:"source_category"`
	FileType               RepositoryContextIndexFileType             `json:"file_type"`
	LanguageOrEcosystem    string                                     `json:"language_or_ecosystem,omitempty"`
	SizeBytes              int64                                      `json:"size_bytes"`
	ContentHash            string                                     `json:"content_hash"`
	ModifiedTime           string                                     `json:"modified_time"`
	SupportedForRetrieval  bool                                       `json:"supported_for_retrieval"`
	ClassificationHints    []RepositoryContextIndexClassificationHint `json:"classification_hints,omitempty"`
	SourceEvidenceCategory string                                     `json:"source_evidence_category"`
}

type RepositoryContextIndexSkipped struct {
	Path   string                           `json:"path"`
	Reason RepositoryContextIndexSkipReason `json:"reason"`
}

type RepositoryContextIndex struct {
	SchemaVersion int                              `json:"schema_version"`
	Generated     RepositoryContextIndexGeneration `json:"generated"`
	Project       RepositoryContextIndexProject    `json:"project"`
	Limits        RepositoryContextIndexLimits     `json:"limits"`
	Entries       []RepositoryContextIndexEntry    `json:"entries"`
	Skipped       []RepositoryContextIndexSkipped  `json:"skipped"`
	Truncated     bool                             `json:"truncated"`
}

type RepositoryContextIndexStaleReason struct {
	Code    string `json:"code"`
	Path    string `json:"path,omitempty"`
	Message string `json:"message"`
}

type RepositoryContextIndexReport struct {
	Mode         RepositoryContextIndexMode
	Status       RepositoryContextIndexStatus
	IndexPath    string
	Index        RepositoryContextIndex
	StaleReasons []RepositoryContextIndexStaleReason
	ErrorMessage string
}

type RepositoryContextIndexEntryInput struct {
	Path                   string
	SourceCategory         ContextSourceCategory
	FileType               RepositoryContextIndexFileType
	LanguageOrEcosystem    string
	SizeBytes              int64
	ContentHash            string
	ModifiedTime           string
	SupportedForRetrieval  bool
	ClassificationHints    []RepositoryContextIndexClassificationHint
	SourceEvidenceCategory string
}

func DefaultRepositoryContextIndexLimits() RepositoryContextIndexLimits {
	return RepositoryContextIndexLimits{
		MaxIndexedFiles:   500,
		MaxFileSizeBytes:  256 * 1024,
		MaxTotalFileBytes: 5 * 1024 * 1024,
		MaxSkippedRecords: 200,
		MaxDirectoryDepth: 4,
	}
}

func NewRepositoryContextIndexEntry(input RepositoryContextIndexEntryInput) (RepositoryContextIndexEntry, error) {
	relativePath, err := SafeRepositoryContextIndexPath(input.Path)
	if err != nil {
		return RepositoryContextIndexEntry{}, err
	}
	if !input.SourceCategory.IsSupported() {
		return RepositoryContextIndexEntry{}, fmt.Errorf("unsupported repository context index source category: %s", input.SourceCategory)
	}
	if !input.FileType.IsSupported() {
		return RepositoryContextIndexEntry{}, fmt.Errorf("unsupported repository context index file type: %s", input.FileType)
	}
	if input.SizeBytes < 0 {
		return RepositoryContextIndexEntry{}, errors.New("repository context index entry size must not be negative")
	}
	if !isSHA256ContentHash(input.ContentHash) {
		return RepositoryContextIndexEntry{}, fmt.Errorf("repository context index content hash must be sha256:<64 hex>: %s", input.ContentHash)
	}
	if strings.TrimSpace(input.ModifiedTime) == "" {
		return RepositoryContextIndexEntry{}, errors.New("repository context index modified time is required")
	}
	evidenceCategory := strings.TrimSpace(input.SourceEvidenceCategory)
	if evidenceCategory == "" {
		return RepositoryContextIndexEntry{}, errors.New("repository context index source evidence category is required")
	}

	hints := append([]RepositoryContextIndexClassificationHint(nil), input.ClassificationHints...)
	sort.SliceStable(hints, func(i, j int) bool { return hints[i] < hints[j] })
	for _, hint := range hints {
		if !hint.IsSupported() {
			return RepositoryContextIndexEntry{}, fmt.Errorf("unsupported repository context index classification hint: %s", hint)
		}
	}

	return RepositoryContextIndexEntry{
		Path:                   relativePath,
		SourceCategory:         input.SourceCategory,
		FileType:               input.FileType,
		LanguageOrEcosystem:    strings.TrimSpace(input.LanguageOrEcosystem),
		SizeBytes:              input.SizeBytes,
		ContentHash:            strings.TrimSpace(input.ContentHash),
		ModifiedTime:           strings.TrimSpace(input.ModifiedTime),
		SupportedForRetrieval:  input.SupportedForRetrieval,
		ClassificationHints:    hints,
		SourceEvidenceCategory: evidenceCategory,
	}, nil
}

func NewRepositoryContextIndexSkipped(
	relativePath string,
	reason RepositoryContextIndexSkipReason,
) (RepositoryContextIndexSkipped, error) {
	safePath, err := SafeRepositoryContextIndexPath(relativePath)
	if err != nil {
		return RepositoryContextIndexSkipped{}, err
	}
	if !reason.IsSupported() {
		return RepositoryContextIndexSkipped{}, fmt.Errorf("unsupported repository context index skip reason: %s", reason)
	}
	return RepositoryContextIndexSkipped{Path: safePath, Reason: reason}, nil
}

func NewRepositoryContextIndex(
	projectRootMarker string,
	limits RepositoryContextIndexLimits,
	entries []RepositoryContextIndexEntry,
	skipped []RepositoryContextIndexSkipped,
	truncated bool,
) (RepositoryContextIndex, error) {
	if err := limits.Validate(); err != nil {
		return RepositoryContextIndex{}, err
	}

	rootMarker := strings.TrimSpace(projectRootMarker)
	if rootMarker == "" {
		rootMarker = "none"
	} else if rootMarker != "none" {
		var err error
		rootMarker, err = SafeRepositoryContextIndexPath(rootMarker)
		if err != nil {
			return RepositoryContextIndex{}, fmt.Errorf("repository context index root marker: %w", err)
		}
	}

	copiedEntries := append([]RepositoryContextIndexEntry(nil), entries...)
	copiedSkipped := append([]RepositoryContextIndexSkipped(nil), skipped...)
	sortRepositoryContextIndexEntries(copiedEntries)
	sortRepositoryContextIndexSkipped(copiedSkipped)

	return RepositoryContextIndex{
		SchemaVersion: RepositoryContextIndexSchemaVersion,
		Generated: RepositoryContextIndexGeneration{
			Mode: "deterministic",
			Tool: "specharbor context index",
		},
		Project:   RepositoryContextIndexProject{RootMarker: rootMarker},
		Limits:    limits,
		Entries:   copiedEntries,
		Skipped:   copiedSkipped,
		Truncated: truncated,
	}, nil
}

func NormalizeRepositoryContextIndex(index RepositoryContextIndex) (RepositoryContextIndex, error) {
	if index.SchemaVersion != RepositoryContextIndexSchemaVersion {
		return RepositoryContextIndex{}, fmt.Errorf("unsupported repository context index schema version: %d", index.SchemaVersion)
	}
	if index.Generated.Mode != "deterministic" || index.Generated.Tool != "specharbor context index" {
		return RepositoryContextIndex{}, errors.New("unsupported repository context index generation marker")
	}
	entries := make([]RepositoryContextIndexEntry, 0, len(index.Entries))
	for _, entry := range index.Entries {
		normalized, err := NewRepositoryContextIndexEntry(RepositoryContextIndexEntryInput{
			Path:                   entry.Path,
			SourceCategory:         entry.SourceCategory,
			FileType:               entry.FileType,
			LanguageOrEcosystem:    entry.LanguageOrEcosystem,
			SizeBytes:              entry.SizeBytes,
			ContentHash:            entry.ContentHash,
			ModifiedTime:           entry.ModifiedTime,
			SupportedForRetrieval:  entry.SupportedForRetrieval,
			ClassificationHints:    entry.ClassificationHints,
			SourceEvidenceCategory: entry.SourceEvidenceCategory,
		})
		if err != nil {
			return RepositoryContextIndex{}, err
		}
		entries = append(entries, normalized)
	}
	skipped := make([]RepositoryContextIndexSkipped, 0, len(index.Skipped))
	for _, skip := range index.Skipped {
		normalized, err := NewRepositoryContextIndexSkipped(skip.Path, skip.Reason)
		if err != nil {
			return RepositoryContextIndex{}, err
		}
		skipped = append(skipped, normalized)
	}
	return NewRepositoryContextIndex(index.Project.RootMarker, index.Limits, entries, skipped, index.Truncated)
}

func CompareRepositoryContextIndexes(
	stored RepositoryContextIndex,
	current RepositoryContextIndex,
) []RepositoryContextIndexStaleReason {
	var reasons []RepositoryContextIndexStaleReason
	if stored.SchemaVersion != current.SchemaVersion {
		reasons = append(reasons, staleReason("schema_version_changed", "", "schema version changed"))
	}
	if stored.Generated != current.Generated {
		reasons = append(reasons, staleReason("generation_marker_changed", "", "generation marker changed"))
	}
	if stored.Project != current.Project {
		reasons = append(reasons, staleReason("project_marker_changed", "", "project root marker changed"))
	}
	if stored.Limits != current.Limits {
		reasons = append(reasons, staleReason("limits_changed", "", "index limits changed"))
	}
	if stored.Truncated != current.Truncated {
		reasons = append(reasons, staleReason("truncation_changed", "", "truncation state changed"))
	}

	storedEntries := repositoryContextIndexEntryMap(stored.Entries)
	currentEntries := repositoryContextIndexEntryMap(current.Entries)
	for path, storedEntry := range storedEntries {
		currentEntry, ok := currentEntries[path]
		if !ok {
			reasons = append(reasons, staleReason("entry_removed", path, "indexed source is no longer present"))
			continue
		}
		reasons = append(reasons, compareRepositoryContextIndexEntry(storedEntry, currentEntry)...)
	}
	for path := range currentEntries {
		if _, ok := storedEntries[path]; !ok {
			reasons = append(reasons, staleReason("entry_added", path, "new supported source is present"))
		}
	}

	storedSkipped := repositoryContextIndexSkippedMap(stored.Skipped)
	currentSkipped := repositoryContextIndexSkippedMap(current.Skipped)
	for key, skip := range storedSkipped {
		if _, ok := currentSkipped[key]; !ok {
			reasons = append(reasons, staleReason("skip_removed", skip.Path, "skip record changed"))
		}
	}
	for key, skip := range currentSkipped {
		if _, ok := storedSkipped[key]; !ok {
			reasons = append(reasons, staleReason("skip_added", skip.Path, "skip record changed"))
		}
	}

	sort.SliceStable(reasons, func(i, j int) bool {
		if reasons[i].Path != reasons[j].Path {
			return reasons[i].Path < reasons[j].Path
		}
		return reasons[i].Code < reasons[j].Code
	})
	return reasons
}

func (limits RepositoryContextIndexLimits) Validate() error {
	if limits.MaxIndexedFiles <= 0 {
		return errors.New("repository context index max indexed files must be positive")
	}
	if limits.MaxFileSizeBytes <= 0 {
		return errors.New("repository context index max file size must be positive")
	}
	if limits.MaxTotalFileBytes <= 0 {
		return errors.New("repository context index max total file bytes must be positive")
	}
	if limits.MaxSkippedRecords < 0 {
		return errors.New("repository context index max skipped records must not be negative")
	}
	if limits.MaxDirectoryDepth < 0 {
		return errors.New("repository context index max directory depth must not be negative")
	}
	return nil
}

func (fileType RepositoryContextIndexFileType) IsSupported() bool {
	switch fileType {
	case RepositoryContextIndexFileTypeMarkdown,
		RepositoryContextIndexFileTypeJSON,
		RepositoryContextIndexFileTypeYAML,
		RepositoryContextIndexFileTypeXML,
		RepositoryContextIndexFileTypeGradle,
		RepositoryContextIndexFileTypeTOML,
		RepositoryContextIndexFileTypeText,
		RepositoryContextIndexFileTypeDockerfile,
		RepositoryContextIndexFileTypeMakefile,
		RepositoryContextIndexFileTypeGoModule:
		return true
	default:
		return false
	}
}

func (hint RepositoryContextIndexClassificationHint) IsSupported() bool {
	switch hint {
	case RepositoryContextIndexHintUserConfirmedContext,
		RepositoryContextIndexHintDetectedFact,
		RepositoryContextIndexHintSuggestedAssumption,
		RepositoryContextIndexHintInventoryMetadata:
		return true
	default:
		return false
	}
}

func (reason RepositoryContextIndexSkipReason) IsSupported() bool {
	switch reason {
	case RepositoryContextIndexSkipSensitiveFile,
		RepositoryContextIndexSkipGeneratedDirectory,
		RepositoryContextIndexSkipSymlink,
		RepositoryContextIndexSkipUnsupportedSource,
		RepositoryContextIndexSkipFileTooLarge,
		RepositoryContextIndexSkipMaxFilesReached,
		RepositoryContextIndexSkipMaxTotalBytesReached,
		RepositoryContextIndexSkipUnsafePath:
		return true
	default:
		return false
	}
}

func SafeRepositoryContextIndexPath(relativePath string) (string, error) {
	trimmed := strings.TrimSpace(relativePath)
	if trimmed == "" {
		return "", errors.New("repository context index path is required")
	}
	if strings.ContainsRune(trimmed, 0) {
		return "", errors.New("repository context index path contains a null byte")
	}
	normalized := normalizeContextPath(trimmed)
	if strings.HasPrefix(normalized, "/") || isRepositoryContextIndexWindowsDrivePath(normalized) {
		return "", fmt.Errorf("repository context index path must be relative: %s", relativePath)
	}
	for _, segment := range strings.Split(normalized, "/") {
		if segment == ".." {
			return "", fmt.Errorf("repository context index path must not contain path traversal: %s", relativePath)
		}
	}
	cleaned := cleanRepositoryContextIndexPath(normalized)
	if cleaned == "." || cleaned == ".." || strings.HasPrefix(cleaned, "../") {
		return "", fmt.Errorf("repository context index path must be a safe relative path: %s", relativePath)
	}
	return cleaned, nil
}

func ShouldSkipRepositoryContextIndexPath(relativePath string) (RepositoryContextIndexSkipReason, bool) {
	normalized := normalizeContextPath(relativePath)
	if normalized == "" || normalized == "." {
		return "", false
	}
	segments := strings.Split(normalized, "/")
	for _, segment := range segments[:len(segments)-1] {
		if isGeneratedContextDirectoryName(segment) {
			return RepositoryContextIndexSkipGeneratedDirectory, true
		}
	}
	base := segments[len(segments)-1]
	if isSensitiveContextFileName(base) {
		return RepositoryContextIndexSkipSensitiveFile, true
	}
	if isGeneratedContextDirectoryName(base) {
		return RepositoryContextIndexSkipGeneratedDirectory, true
	}
	return "", false
}

func sortRepositoryContextIndexEntries(entries []RepositoryContextIndexEntry) {
	sort.SliceStable(entries, func(i, j int) bool {
		if entries[i].Path != entries[j].Path {
			return entries[i].Path < entries[j].Path
		}
		if entries[i].SourceCategory != entries[j].SourceCategory {
			return entries[i].SourceCategory < entries[j].SourceCategory
		}
		return entries[i].ContentHash < entries[j].ContentHash
	})
}

func sortRepositoryContextIndexSkipped(skipped []RepositoryContextIndexSkipped) {
	sort.SliceStable(skipped, func(i, j int) bool {
		if skipped[i].Path != skipped[j].Path {
			return skipped[i].Path < skipped[j].Path
		}
		return skipped[i].Reason < skipped[j].Reason
	})
}

func repositoryContextIndexEntryMap(entries []RepositoryContextIndexEntry) map[string]RepositoryContextIndexEntry {
	result := make(map[string]RepositoryContextIndexEntry, len(entries))
	for _, entry := range entries {
		result[entry.Path] = entry
	}
	return result
}

func repositoryContextIndexSkippedMap(skipped []RepositoryContextIndexSkipped) map[string]RepositoryContextIndexSkipped {
	result := make(map[string]RepositoryContextIndexSkipped, len(skipped))
	for _, skip := range skipped {
		result[skip.Path+"\x00"+string(skip.Reason)] = skip
	}
	return result
}

func compareRepositoryContextIndexEntry(
	stored RepositoryContextIndexEntry,
	current RepositoryContextIndexEntry,
) []RepositoryContextIndexStaleReason {
	var reasons []RepositoryContextIndexStaleReason
	path := stored.Path
	if stored.SizeBytes != current.SizeBytes {
		reasons = append(reasons, staleReason("file_size_changed", path, "file size changed"))
	}
	if stored.ContentHash != current.ContentHash {
		reasons = append(reasons, staleReason("content_hash_changed", path, "content hash changed"))
	}
	if stored.ModifiedTime != current.ModifiedTime {
		reasons = append(reasons, staleReason("modified_time_changed", path, "modified freshness marker changed"))
	}
	if stored.SourceCategory != current.SourceCategory ||
		stored.FileType != current.FileType ||
		stored.LanguageOrEcosystem != current.LanguageOrEcosystem ||
		stored.SupportedForRetrieval != current.SupportedForRetrieval ||
		stored.SourceEvidenceCategory != current.SourceEvidenceCategory ||
		strings.Join(hintsAsStrings(stored.ClassificationHints), ",") != strings.Join(hintsAsStrings(current.ClassificationHints), ",") {
		reasons = append(reasons, staleReason("entry_metadata_changed", path, "entry metadata changed"))
	}
	return reasons
}

func staleReason(code string, path string, message string) RepositoryContextIndexStaleReason {
	return RepositoryContextIndexStaleReason{
		Code:    code,
		Path:    path,
		Message: message,
	}
}

func hintsAsStrings(hints []RepositoryContextIndexClassificationHint) []string {
	values := make([]string, 0, len(hints))
	for _, hint := range hints {
		values = append(values, string(hint))
	}
	sort.Strings(values)
	return values
}

func isSHA256ContentHash(value string) bool {
	trimmed := strings.TrimSpace(value)
	if !strings.HasPrefix(trimmed, "sha256:") || len(trimmed) != len("sha256:")+64 {
		return false
	}
	for _, char := range strings.TrimPrefix(trimmed, "sha256:") {
		if (char >= '0' && char <= '9') ||
			(char >= 'a' && char <= 'f') {
			continue
		}
		return false
	}
	return true
}

func isRepositoryContextIndexWindowsDrivePath(value string) bool {
	return len(value) >= 2 && isRepositoryContextIndexASCIIAlpha(value[0]) && value[1] == ':'
}

func isRepositoryContextIndexASCIIAlpha(value byte) bool {
	return (value >= 'a' && value <= 'z') || (value >= 'A' && value <= 'Z')
}

func cleanRepositoryContextIndexPath(value string) string {
	var segments []string
	for _, segment := range strings.Split(value, "/") {
		if segment == "" || segment == "." {
			continue
		}
		segments = append(segments, segment)
	}
	if len(segments) == 0 {
		return "."
	}
	return strings.Join(segments, "/")
}
