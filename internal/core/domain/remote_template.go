package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"unicode"
)

const (
	MaxRemoteTemplateURLLength             = 2048
	MaxRemoteTemplateHTTPResponseBytes     = 5 * 1024 * 1024
	MaxRemoteTemplateUncompressedBytes     = 1 * 1024 * 1024
	MaxRemoteTemplateFileUncompressedBytes = 256 * 1024
)

type RemoteTemplateURL struct {
	value string
	host  string
}

func NewRemoteTemplateURL(raw string) (RemoteTemplateURL, error) {
	if raw == "" {
		return RemoteTemplateURL{}, errors.New("remote template URL is required")
	}
	if len(raw) > MaxRemoteTemplateURLLength {
		return RemoteTemplateURL{}, fmt.Errorf("remote template URL must be at most %d characters", MaxRemoteTemplateURLLength)
	}
	if containsWhitespaceOrControl(raw) {
		return RemoteTemplateURL{}, errors.New("remote template URL must not contain whitespace or control characters")
	}
	if isSCPStyleGitTarget(raw) {
		return RemoteTemplateURL{}, errors.New("remote template URL must use https")
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		return RemoteTemplateURL{}, fmt.Errorf("remote template URL is invalid: %w", err)
	}
	if parsed.Scheme != "https" {
		return RemoteTemplateURL{}, errors.New("remote template URL must use https")
	}
	if parsed.Host == "" || parsed.Hostname() == "" {
		return RemoteTemplateURL{}, errors.New("remote template URL host is required")
	}
	if parsed.User != nil {
		return RemoteTemplateURL{}, errors.New("remote template URL must not include credentials")
	}
	if parsed.RawQuery != "" || parsed.ForceQuery || strings.Contains(raw, "?") {
		return RemoteTemplateURL{}, errors.New("remote template URL must not include query strings")
	}
	if parsed.Fragment != "" || strings.Contains(raw, "#") {
		return RemoteTemplateURL{}, errors.New("remote template URL must not include fragments")
	}
	if parsed.Path == "" || parsed.Path == "/" {
		return RemoteTemplateURL{}, errors.New("remote template URL path is required")
	}

	return RemoteTemplateURL{value: raw, host: parsed.Host}, nil
}

func (remoteURL RemoteTemplateURL) String() string {
	return remoteURL.value
}

func (remoteURL RemoteTemplateURL) Host() string {
	return remoteURL.host
}

func containsWhitespaceOrControl(value string) bool {
	for _, character := range value {
		if unicode.IsSpace(character) || unicode.IsControl(character) {
			return true
		}
	}
	return false
}

func isSCPStyleGitTarget(value string) bool {
	if strings.Contains(value, "://") {
		return false
	}
	at := strings.Index(value, "@")
	colon := strings.Index(value, ":")
	return at > 0 && colon > at+1
}

type RemoteTemplateFormat string

const RemoteTemplateFormatZip RemoteTemplateFormat = "zip"

func ParseRemoteTemplateFormat(raw string) (RemoteTemplateFormat, error) {
	value := RemoteTemplateFormat(strings.TrimSpace(raw))
	if value == "" {
		return "", errors.New("remote template format is required")
	}
	if value != RemoteTemplateFormatZip {
		return "", fmt.Errorf("unsupported remote template format: %s", value)
	}
	return value, nil
}

type ChecksumAlgorithm string

const ChecksumAlgorithmSHA256 ChecksumAlgorithm = "sha256"

type RemoteTemplateChecksum struct {
	algorithm ChecksumAlgorithm
	digest    string
}

func ParseRemoteTemplateChecksum(raw string) (RemoteTemplateChecksum, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return RemoteTemplateChecksum{}, errors.New("remote template checksum is required")
	}
	if value != raw || containsWhitespaceOrControl(value) {
		return RemoteTemplateChecksum{}, errors.New("remote template checksum must be sha256:<64 hex>")
	}

	parts := strings.Split(value, ":")
	if len(parts) != 2 {
		return RemoteTemplateChecksum{}, errors.New("remote template checksum must be sha256:<64 hex>")
	}
	if parts[0] != string(ChecksumAlgorithmSHA256) {
		return RemoteTemplateChecksum{}, fmt.Errorf("unsupported remote template checksum algorithm: %s", parts[0])
	}

	digest := strings.ToLower(parts[1])
	if len(digest) != 64 {
		return RemoteTemplateChecksum{}, errors.New("remote template sha256 checksum must contain 64 hex characters")
	}
	for _, character := range digest {
		if !isHexCharacter(character) {
			return RemoteTemplateChecksum{}, errors.New("remote template sha256 checksum must contain only hex characters")
		}
	}

	return RemoteTemplateChecksum{
		algorithm: ChecksumAlgorithmSHA256,
		digest:    digest,
	}, nil
}

func NewRemoteTemplateChecksumFromBytes(contents []byte) RemoteTemplateChecksum {
	sum := sha256.Sum256(contents)
	return RemoteTemplateChecksum{
		algorithm: ChecksumAlgorithmSHA256,
		digest:    hex.EncodeToString(sum[:]),
	}
}

func (checksum RemoteTemplateChecksum) Algorithm() ChecksumAlgorithm {
	return checksum.algorithm
}

func (checksum RemoteTemplateChecksum) Digest() string {
	return checksum.digest
}

func (checksum RemoteTemplateChecksum) String() string {
	if checksum.algorithm == "" || checksum.digest == "" {
		return ""
	}
	return string(checksum.algorithm) + ":" + checksum.digest
}

func (checksum RemoteTemplateChecksum) MatchesBytes(contents []byte) (RemoteTemplateChecksum, bool) {
	actual := NewRemoteTemplateChecksumFromBytes(contents)
	return actual, checksum.String() == actual.String()
}

func isHexCharacter(character rune) bool {
	return (character >= '0' && character <= '9') ||
		(character >= 'a' && character <= 'f')
}

type RemoteTemplateReference struct {
	remoteURL RemoteTemplateURL
	checksum  RemoteTemplateChecksum
	format    RemoteTemplateFormat
}

func NewRemoteTemplateReference(rawURL string, rawChecksum string, rawFormat string) (RemoteTemplateReference, error) {
	remoteURL, err := NewRemoteTemplateURL(rawURL)
	if err != nil {
		return RemoteTemplateReference{}, err
	}
	checksum, err := ParseRemoteTemplateChecksum(rawChecksum)
	if err != nil {
		return RemoteTemplateReference{}, err
	}
	format, err := ParseRemoteTemplateFormat(rawFormat)
	if err != nil {
		return RemoteTemplateReference{}, err
	}

	return RemoteTemplateReference{
		remoteURL: remoteURL,
		checksum:  checksum,
		format:    format,
	}, nil
}

func (reference RemoteTemplateReference) URL() RemoteTemplateURL {
	return reference.remoteURL
}

func (reference RemoteTemplateReference) Checksum() RemoteTemplateChecksum {
	return reference.checksum
}

func (reference RemoteTemplateReference) Format() RemoteTemplateFormat {
	return reference.format
}

type RemoteTemplateArchivePolicy struct {
	requiredFiles     []string
	maxTotalBytes     uint64
	maxBytesPerFile   uint64
	maxRegularEntries int
}

func DefaultRemoteTemplateArchivePolicy() RemoteTemplateArchivePolicy {
	return RemoteTemplateArchivePolicy{
		requiredFiles:     RequiredOpenSpecChangeFiles(),
		maxTotalBytes:     MaxRemoteTemplateUncompressedBytes,
		maxBytesPerFile:   MaxRemoteTemplateFileUncompressedBytes,
		maxRegularEntries: len(RequiredOpenSpecChangeFiles()),
	}
}

func NewRemoteTemplateArchivePolicy(
	requiredFiles []string,
	maxTotalBytes uint64,
	maxBytesPerFile uint64,
) RemoteTemplateArchivePolicy {
	return RemoteTemplateArchivePolicy{
		requiredFiles:     append([]string(nil), requiredFiles...),
		maxTotalBytes:     maxTotalBytes,
		maxBytesPerFile:   maxBytesPerFile,
		maxRegularEntries: len(requiredFiles),
	}
}

func (policy RemoteTemplateArchivePolicy) RequiredFiles() []string {
	return append([]string(nil), policy.requiredFiles...)
}

func (policy RemoteTemplateArchivePolicy) MaxTotalBytes() uint64 {
	return policy.maxTotalBytes
}

func (policy RemoteTemplateArchivePolicy) MaxBytesPerFile() uint64 {
	return policy.maxBytesPerFile
}

func (policy RemoteTemplateArchivePolicy) MaxRegularEntries() int {
	return policy.maxRegularEntries
}

func (policy RemoteTemplateArchivePolicy) RequiredFileSet() map[string]struct{} {
	required := make(map[string]struct{}, len(policy.requiredFiles))
	for _, file := range policy.requiredFiles {
		required[file] = struct{}{}
	}
	return required
}

type RemoteTemplateArchiveEntry struct {
	Name             string
	IsDirectory      bool
	IsSymlink        bool
	IsExecutable     bool
	UncompressedSize uint64
}

func ValidateRemoteTemplateArchiveEntry(policy RemoteTemplateArchivePolicy, entry RemoteTemplateArchiveEntry) error {
	if strings.TrimSpace(entry.Name) == "" {
		return errors.New("remote template archive path is required")
	}
	if entry.IsDirectory {
		return fmt.Errorf("remote template archive contains unsupported directory entry: %s", entry.Name)
	}
	if entry.IsSymlink {
		return fmt.Errorf("remote template archive entry is a symlink: %s", entry.Name)
	}
	if entry.IsExecutable {
		return fmt.Errorf("remote template archive entry is executable: %s", entry.Name)
	}
	if entry.UncompressedSize > policy.MaxBytesPerFile() {
		return fmt.Errorf(
			"remote template archive file %s exceeds maximum size %d bytes",
			entry.Name,
			policy.MaxBytesPerFile(),
		)
	}
	if err := validateRemoteTemplateArchivePath(entry.Name); err != nil {
		return err
	}
	if _, exists := policy.RequiredFileSet()[entry.Name]; !exists {
		return fmt.Errorf("remote template archive contains unsupported file: %s", entry.Name)
	}
	return nil
}

func validateRemoteTemplateArchivePath(name string) error {
	normalized := strings.ReplaceAll(name, "\\", "/")
	if isWindowsDriveArchivePath(normalized) {
		return fmt.Errorf("remote template archive path must not be a Windows drive path: %s", name)
	}
	if strings.HasPrefix(name, "/") || strings.HasPrefix(name, "\\") {
		return fmt.Errorf("remote template archive path must not be absolute: %s", name)
	}
	for _, segment := range strings.Split(normalized, "/") {
		if segment == ".." {
			return fmt.Errorf("remote template archive path must not contain traversal: %s", name)
		}
	}
	if strings.Contains(normalized, "/") {
		return fmt.Errorf("remote template archive path must be root-level: %s", name)
	}
	return nil
}

func isWindowsDriveArchivePath(value string) bool {
	return len(value) >= 2 && isRemoteTemplateASCIIAlpha(value[0]) && value[1] == ':'
}

func isRemoteTemplateASCIIAlpha(value byte) bool {
	return (value >= 'a' && value <= 'z') || (value >= 'A' && value <= 'Z')
}

type RemoteTemplateBundle struct {
	files map[string]string
}

func NewRemoteTemplateBundle(files map[string]string) (RemoteTemplateBundle, error) {
	required := RequiredOpenSpecChangeFiles()
	requiredSet := make(map[string]struct{}, len(required))
	for _, file := range required {
		requiredSet[file] = struct{}{}
	}

	for file := range files {
		if _, exists := requiredSet[file]; !exists {
			return RemoteTemplateBundle{}, fmt.Errorf("remote template archive contains unsupported file: %s", file)
		}
	}

	var missingFiles []string
	for _, file := range required {
		if _, exists := files[file]; !exists {
			missingFiles = append(missingFiles, file)
		}
	}
	if len(missingFiles) > 0 {
		return RemoteTemplateBundle{}, fmt.Errorf("remote template archive is missing required files: %s", strings.Join(missingFiles, ", "))
	}

	copied := make(map[string]string, len(required))
	for _, file := range required {
		contents := files[file]
		if strings.TrimSpace(contents) == "" {
			return RemoteTemplateBundle{}, fmt.Errorf("remote template file %s is empty", file)
		}
		copied[file] = contents
	}

	return RemoteTemplateBundle{files: copied}, nil
}

func (bundle RemoteTemplateBundle) Files() map[string]string {
	copied := make(map[string]string, len(bundle.files))
	for file, contents := range bundle.files {
		copied[file] = contents
	}
	return copied
}

type RemoteTemplateFetchRequest struct {
	remoteURL RemoteTemplateURL
}

func NewRemoteTemplateFetchRequest(remoteURL RemoteTemplateURL) RemoteTemplateFetchRequest {
	return RemoteTemplateFetchRequest{remoteURL: remoteURL}
}

func (request RemoteTemplateFetchRequest) URL() RemoteTemplateURL {
	return request.remoteURL
}

type RemoteTemplateFetchResult struct {
	statusCode int
	body       []byte
}

func NewRemoteTemplateFetchResult(statusCode int, body []byte) RemoteTemplateFetchResult {
	return RemoteTemplateFetchResult{
		statusCode: statusCode,
		body:       append([]byte(nil), body...),
	}
}

func (result RemoteTemplateFetchResult) StatusCode() int {
	return result.statusCode
}

func (result RemoteTemplateFetchResult) Body() []byte {
	return append([]byte(nil), result.body...)
}
