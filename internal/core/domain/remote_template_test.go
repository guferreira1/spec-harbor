package domain

import (
	"strings"
	"testing"
)

func TestRemoteTemplateURLValidationAcceptsHTTPSHostAndPath(t *testing.T) {
	remoteURL, err := NewRemoteTemplateURL("https://example.com/templates/service-feature.zip")
	if err != nil {
		t.Fatalf("NewRemoteTemplateURL() error = %v", err)
	}
	if remoteURL.String() != "https://example.com/templates/service-feature.zip" {
		t.Fatalf("String() = %q, want original URL", remoteURL.String())
	}
	if remoteURL.Host() != "example.com" {
		t.Fatalf("Host() = %q, want example.com", remoteURL.Host())
	}
}

func TestRemoteTemplateURLValidationRejectsUnsafeValues(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{name: "missing", raw: "", want: "remote template URL is required"},
		{name: "invalid", raw: "https://[::1/template.zip", want: "remote template URL is invalid"},
		{name: "http", raw: "http://example.com/template.zip", want: "remote template URL must use https"},
		{name: "file", raw: "file:///tmp/template.zip", want: "remote template URL must use https"},
		{name: "ssh", raw: "ssh://example.com/template.zip", want: "remote template URL must use https"},
		{name: "git", raw: "git://example.com/template.zip", want: "remote template URL must use https"},
		{name: "git ssh", raw: "git+ssh://example.com/template.zip", want: "remote template URL must use https"},
		{name: "ftp", raw: "ftp://example.com/template.zip", want: "remote template URL must use https"},
		{name: "scp style", raw: "git@example.com:org/repo", want: "remote template URL must use https"},
		{name: "missing host", raw: "https:///template.zip", want: "remote template URL host is required"},
		{name: "missing path", raw: "https://example.com", want: "remote template URL path is required"},
		{name: "root path only", raw: "https://example.com/", want: "remote template URL path is required"},
		{name: "credentials", raw: "https://user:pass@example.com/template.zip", want: "remote template URL must not include credentials"},
		{name: "query", raw: "https://example.com/template.zip?token=secret", want: "remote template URL must not include query strings"},
		{name: "empty query marker", raw: "https://example.com/template.zip?", want: "remote template URL must not include query strings"},
		{name: "fragment", raw: "https://example.com/template.zip#section", want: "remote template URL must not include fragments"},
		{name: "whitespace", raw: "https://example.com/template zip", want: "remote template URL must not contain whitespace or control characters"},
		{name: "control", raw: "https://example.com/template.zip\n", want: "remote template URL must not contain whitespace or control characters"},
		{name: "over length", raw: "https://example.com/" + strings.Repeat("a", MaxRemoteTemplateURLLength), want: "remote template URL must be at most 2048 characters"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewRemoteTemplateURL(test.raw)
			if err == nil {
				t.Fatalf("NewRemoteTemplateURL(%q) error = nil, want %q", test.raw, test.want)
			}
			if !strings.Contains(err.Error(), test.want) {
				t.Fatalf("NewRemoteTemplateURL(%q) error = %q, want %q", test.raw, err.Error(), test.want)
			}
		})
	}
}

func TestRemoteTemplateFormatAcceptsOnlyZip(t *testing.T) {
	format, err := ParseRemoteTemplateFormat(" zip ")
	if err != nil {
		t.Fatalf("ParseRemoteTemplateFormat() error = %v", err)
	}
	if format != RemoteTemplateFormatZip {
		t.Fatalf("format = %q, want zip", format)
	}

	for _, test := range []struct {
		raw  string
		want string
	}{
		{raw: "", want: "remote template format is required"},
		{raw: "tar", want: "unsupported remote template format: tar"},
	} {
		_, err := ParseRemoteTemplateFormat(test.raw)
		if err == nil || err.Error() != test.want {
			t.Fatalf("ParseRemoteTemplateFormat(%q) error = %v, want %q", test.raw, err, test.want)
		}
	}
}

func TestRemoteTemplateChecksumParsingAndVerification(t *testing.T) {
	checksum, err := ParseRemoteTemplateChecksum("sha256:ABCDEFabcdef0123456789abcdef0123456789abcdef0123456789abcdef0123")
	if err != nil {
		t.Fatalf("ParseRemoteTemplateChecksum() error = %v", err)
	}
	if checksum.String() != "sha256:abcdefabcdef0123456789abcdef0123456789abcdef0123456789abcdef0123" {
		t.Fatalf("checksum = %q, want normalized lowercase", checksum.String())
	}
	if checksum.Algorithm() != ChecksumAlgorithmSHA256 {
		t.Fatalf("Algorithm = %q, want sha256", checksum.Algorithm())
	}

	bytesChecksum := NewRemoteTemplateChecksumFromBytes([]byte("bundle"))
	actual, matches := bytesChecksum.MatchesBytes([]byte("bundle"))
	if !matches {
		t.Fatalf("MatchesBytes() = false, want true")
	}
	if actual.String() != bytesChecksum.String() {
		t.Fatalf("actual = %q, want %q", actual, bytesChecksum)
	}
	mismatchActual, matches := bytesChecksum.MatchesBytes([]byte("different"))
	if matches {
		t.Fatalf("MatchesBytes(different) = true, want false")
	}
	if bytesChecksum.Algorithm() != ChecksumAlgorithmSHA256 {
		t.Fatalf("expected checksum algorithm = %q, want sha256", bytesChecksum.Algorithm())
	}
	if mismatchActual.Algorithm() != ChecksumAlgorithmSHA256 {
		t.Fatalf("actual checksum algorithm = %q, want sha256", mismatchActual.Algorithm())
	}
	wantActual := NewRemoteTemplateChecksumFromBytes([]byte("different"))
	if mismatchActual.Digest() != wantActual.Digest() {
		t.Fatalf("actual mismatch digest = %q, want %q", mismatchActual.Digest(), wantActual.Digest())
	}
	if bytesChecksum.Digest() == mismatchActual.Digest() {
		t.Fatalf("expected digest %q should differ from actual digest %q", bytesChecksum.Digest(), mismatchActual.Digest())
	}
}

func TestRemoteTemplateChecksumRejectsInvalidValues(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{name: "missing", raw: "", want: "remote template checksum is required"},
		{name: "malformed", raw: "sha256", want: "remote template checksum must be sha256:<64 hex>"},
		{name: "short", raw: "sha256:abc", want: "remote template sha256 checksum must contain 64 hex characters"},
		{name: "long", raw: "sha256:" + strings.Repeat("a", 65), want: "remote template sha256 checksum must contain 64 hex characters"},
		{name: "non hex", raw: "sha256:" + strings.Repeat("g", 64), want: "remote template sha256 checksum must contain only hex characters"},
		{name: "unsupported", raw: "sha512:" + strings.Repeat("a", 64), want: "unsupported remote template checksum algorithm: sha512"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := ParseRemoteTemplateChecksum(test.raw)
			if err == nil {
				t.Fatalf("ParseRemoteTemplateChecksum(%q) error = nil, want %q", test.raw, test.want)
			}
			if err.Error() != test.want {
				t.Fatalf("ParseRemoteTemplateChecksum(%q) error = %q, want %q", test.raw, err.Error(), test.want)
			}
		})
	}
}

func TestRemoteTemplateArchiveEntryPolicy(t *testing.T) {
	policy := DefaultRemoteTemplateArchivePolicy()
	if err := ValidateRemoteTemplateArchiveEntry(policy, RemoteTemplateArchiveEntry{Name: "proposal.md", UncompressedSize: 10}); err != nil {
		t.Fatalf("ValidateRemoteTemplateArchiveEntry(valid) error = %v", err)
	}

	tests := []struct {
		name  string
		entry RemoteTemplateArchiveEntry
		want  string
	}{
		{name: "traversal", entry: RemoteTemplateArchiveEntry{Name: "../proposal.md"}, want: "remote template archive path must not contain traversal"},
		{name: "absolute", entry: RemoteTemplateArchiveEntry{Name: "/proposal.md"}, want: "remote template archive path must not be absolute"},
		{name: "nested", entry: RemoteTemplateArchiveEntry{Name: "nested/proposal.md"}, want: "remote template archive path must be root-level"},
		{name: "windows drive", entry: RemoteTemplateArchiveEntry{Name: `C:\proposal.md`}, want: "remote template archive path must not be a Windows drive path"},
		{name: "extra", entry: RemoteTemplateArchiveEntry{Name: "README.md"}, want: "remote template archive contains unsupported file: README.md"},
		{name: "directory", entry: RemoteTemplateArchiveEntry{Name: "templates/", IsDirectory: true}, want: "remote template archive contains unsupported directory entry"},
		{name: "symlink", entry: RemoteTemplateArchiveEntry{Name: "proposal.md", IsSymlink: true}, want: "remote template archive entry is a symlink: proposal.md"},
		{name: "executable", entry: RemoteTemplateArchiveEntry{Name: "proposal.md", IsExecutable: true}, want: "remote template archive entry is executable: proposal.md"},
		{name: "oversized", entry: RemoteTemplateArchiveEntry{Name: "proposal.md", UncompressedSize: MaxRemoteTemplateFileUncompressedBytes + 1}, want: "remote template archive file proposal.md exceeds maximum size"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateRemoteTemplateArchiveEntry(policy, test.entry)
			if err == nil {
				t.Fatalf("ValidateRemoteTemplateArchiveEntry() error = nil, want %q", test.want)
			}
			if !strings.Contains(err.Error(), test.want) {
				t.Fatalf("ValidateRemoteTemplateArchiveEntry() error = %q, want %q", err.Error(), test.want)
			}
		})
	}
}

func TestRemoteTemplateBundleValidatesExactRequiredFilesAndCopiesDefensively(t *testing.T) {
	files := validRemoteTemplateFiles()
	bundle, err := NewRemoteTemplateBundle(files)
	if err != nil {
		t.Fatalf("NewRemoteTemplateBundle() error = %v", err)
	}

	files["proposal.md"] = "mutated"
	got := bundle.Files()
	got["design.md"] = "mutated"

	reloaded := bundle.Files()
	if reloaded["proposal.md"] == "mutated" || reloaded["design.md"] == "mutated" {
		t.Fatalf("RemoteTemplateBundle did not copy files defensively: %#v", reloaded)
	}
}

func TestRemoteTemplateBundleRejectsInvalidFileSets(t *testing.T) {
	tests := []struct {
		name  string
		files map[string]string
		want  string
	}{
		{name: "extra", files: withRemoteTemplateFile(validRemoteTemplateFiles(), "README.md", "notes"), want: "remote template archive contains unsupported file: README.md"},
		{name: "missing", files: withoutRemoteTemplateFile(validRemoteTemplateFiles(), "risks.md"), want: "remote template archive is missing required files: risks.md"},
		{name: "empty", files: withRemoteTemplateFile(validRemoteTemplateFiles(), "tasks.md", " \n\t"), want: "remote template file tasks.md is empty"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewRemoteTemplateBundle(test.files)
			if err == nil || err.Error() != test.want {
				t.Fatalf("NewRemoteTemplateBundle() error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestRemoteTemplateFetchResultCopiesBodyDefensively(t *testing.T) {
	body := []byte("downloaded")
	result := NewRemoteTemplateFetchResult(200, body)
	body[0] = 'X'
	copied := result.Body()
	copied[0] = 'Y'
	if string(result.Body()) != "downloaded" {
		t.Fatalf("RemoteTemplateFetchResult did not copy body defensively")
	}
}

func validRemoteTemplateFiles() map[string]string {
	files := make(map[string]string)
	for _, file := range RequiredOpenSpecChangeFiles() {
		files[file] = "# " + file + "\n"
	}
	return files
}

func withRemoteTemplateFile(files map[string]string, name string, contents string) map[string]string {
	files[name] = contents
	return files
}

func withoutRemoteTemplateFile(files map[string]string, name string) map[string]string {
	delete(files, name)
	return files
}
