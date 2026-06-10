package remote

import (
	"archive/zip"
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/guferreira1/spec-harbor/internal/core/domain"
)

func TestZIPTemplateBundleReaderDecodesValidBundle(t *testing.T) {
	bundle, err := NewZIPTemplateBundleReader().ReadRemoteTemplateBundle(
		remoteTemplateZip(t, validZipEntries()),
		domain.DefaultRemoteTemplateArchivePolicy(),
	)
	if err != nil {
		t.Fatalf("ReadRemoteTemplateBundle() error = %v", err)
	}

	files := bundle.Files()
	for _, requiredFile := range domain.RequiredOpenSpecChangeFiles() {
		if !strings.Contains(files[requiredFile], requiredFile) {
			t.Fatalf("bundle file %s = %q, want file content", requiredFile, files[requiredFile])
		}
	}
}

func TestZIPTemplateBundleReaderRejectsMalformedZip(t *testing.T) {
	_, err := NewZIPTemplateBundleReader().ReadRemoteTemplateBundle([]byte("not a zip"), domain.DefaultRemoteTemplateArchivePolicy())
	if err == nil || !strings.Contains(err.Error(), "malformed remote template zip archive") {
		t.Fatalf("ReadRemoteTemplateBundle() error = %v, want malformed zip", err)
	}
}

func TestZIPTemplateBundleReaderRejectsUnsafeEntries(t *testing.T) {
	tests := []struct {
		name    string
		entries []zipEntry
		want    string
	}{
		{name: "traversal", entries: withZipEntry(validZipEntries(), zipEntry{name: "../proposal.md", contents: "bad"}), want: "remote template archive path must not contain traversal"},
		{name: "absolute", entries: withZipEntry(validZipEntries(), zipEntry{name: "/proposal.md", contents: "bad"}), want: "remote template archive path must not be absolute"},
		{name: "nested", entries: replaceZipEntry(validZipEntries(), "proposal.md", zipEntry{name: "nested/proposal.md", contents: "bad"}), want: "remote template archive path must be root-level"},
		{name: "windows drive", entries: replaceZipEntry(validZipEntries(), "proposal.md", zipEntry{name: `C:\proposal.md`, contents: "bad"}), want: "remote template archive path must not be a Windows drive path"},
		{name: "duplicate", entries: withZipEntry(validZipEntries(), zipEntry{name: "proposal.md", contents: "duplicate"}), want: "remote template archive contains duplicate file: proposal.md"},
		{name: "extra", entries: withZipEntry(validZipEntries(), zipEntry{name: "README.md", contents: "notes"}), want: "remote template archive contains unsupported file: README.md"},
		{name: "directory", entries: withZipEntry(validZipEntries(), zipEntry{name: "templates/", directory: true}), want: "remote template archive contains unsupported directory entry: templates/"},
		{name: "symlink", entries: replaceZipEntry(validZipEntries(), "proposal.md", zipEntry{name: "proposal.md", contents: "target", mode: os.ModeSymlink | 0o644}), want: "remote template archive entry is a symlink: proposal.md"},
		{name: "executable", entries: replaceZipEntry(validZipEntries(), "proposal.md", zipEntry{name: "proposal.md", contents: "#!/bin/sh\n", mode: 0o755}), want: "remote template archive entry is executable: proposal.md"},
		{name: "missing", entries: withoutZipEntry(validZipEntries(), "risks.md"), want: "remote template archive is missing required files: risks.md"},
		{name: "empty", entries: replaceZipEntry(validZipEntries(), "tasks.md", zipEntry{name: "tasks.md", contents: " \n\t"}), want: "remote template file tasks.md is empty"},
		{name: "file size", entries: replaceZipEntry(validZipEntries(), "proposal.md", zipEntry{name: "proposal.md", contents: strings.Repeat("x", domain.MaxRemoteTemplateFileUncompressedBytes+1)}), want: "remote template archive file proposal.md exceeds maximum size"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewZIPTemplateBundleReader().ReadRemoteTemplateBundle(
				remoteTemplateZip(t, test.entries),
				domain.DefaultRemoteTemplateArchivePolicy(),
			)
			if err == nil {
				t.Fatalf("ReadRemoteTemplateBundle() error = nil, want %q", test.want)
			}
			if !strings.Contains(err.Error(), test.want) {
				t.Fatalf("ReadRemoteTemplateBundle() error = %q, want %q", err.Error(), test.want)
			}
		})
	}
}

func TestZIPTemplateBundleReaderEnforcesTotalUncompressedSize(t *testing.T) {
	policy := domain.DefaultRemoteTemplateArchivePolicy()
	entries := []zipEntry{
		{name: "proposal.md", contents: strings.Repeat("a", 250)},
		{name: "design.md", contents: strings.Repeat("a", 250)},
		{name: "tasks.md", contents: strings.Repeat("a", 250)},
		{name: "acceptance-criteria.md", contents: strings.Repeat("a", 250)},
		{name: "risks.md", contents: strings.Repeat("a", 250)},
	}
	policy = domain.NewRemoteTemplateArchivePolicy(
		policy.RequiredFiles(),
		1000,
		domain.MaxRemoteTemplateFileUncompressedBytes,
	)

	_, err := NewZIPTemplateBundleReader().ReadRemoteTemplateBundle(remoteTemplateZip(t, entries), policy)
	if err == nil || !strings.Contains(err.Error(), "remote template archive exceeds maximum uncompressed size 1000 bytes") {
		t.Fatalf("ReadRemoteTemplateBundle() error = %v, want total size rejection", err)
	}
}

type zipEntry struct {
	name      string
	contents  string
	mode      os.FileMode
	directory bool
}

func validZipEntries() []zipEntry {
	entries := make([]zipEntry, 0, len(domain.RequiredOpenSpecChangeFiles()))
	for _, requiredFile := range domain.RequiredOpenSpecChangeFiles() {
		entries = append(entries, zipEntry{
			name:     requiredFile,
			contents: "# " + requiredFile + "\n",
			mode:     0o644,
		})
	}
	return entries
}

func remoteTemplateZip(t *testing.T, entries []zipEntry) []byte {
	t.Helper()

	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	for _, entry := range entries {
		header := &zip.FileHeader{Name: entry.name}
		if entry.directory {
			header.SetMode(os.ModeDir | 0o755)
		} else if entry.mode != 0 {
			header.SetMode(entry.mode)
		} else {
			header.SetMode(0o644)
		}
		fileWriter, err := writer.CreateHeader(header)
		if err != nil {
			t.Fatalf("CreateHeader(%q) error = %v", entry.name, err)
		}
		if !entry.directory {
			if _, err := fileWriter.Write([]byte(entry.contents)); err != nil {
				t.Fatalf("Write(%q) error = %v", entry.name, err)
			}
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	return buffer.Bytes()
}

func withZipEntry(entries []zipEntry, entry zipEntry) []zipEntry {
	return append(entries, entry)
}

func withoutZipEntry(entries []zipEntry, name string) []zipEntry {
	filtered := make([]zipEntry, 0, len(entries))
	for _, entry := range entries {
		if entry.name != name {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}

func replaceZipEntry(entries []zipEntry, name string, replacement zipEntry) []zipEntry {
	replaced := make([]zipEntry, 0, len(entries))
	for _, entry := range entries {
		if entry.name == name {
			replaced = append(replaced, replacement)
			continue
		}
		replaced = append(replaced, entry)
	}
	return replaced
}
