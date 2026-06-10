package remote

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/guferreira1/spec-harbor/internal/core/domain"
)

type ZIPTemplateBundleReader struct{}

func NewZIPTemplateBundleReader() *ZIPTemplateBundleReader {
	return &ZIPTemplateBundleReader{}
}

func (reader *ZIPTemplateBundleReader) ReadRemoteTemplateBundle(
	contents []byte,
	policy domain.RemoteTemplateArchivePolicy,
) (domain.RemoteTemplateBundle, error) {
	if policy.MaxRegularEntries() == 0 {
		policy = domain.DefaultRemoteTemplateArchivePolicy()
	}

	zipReader, err := zip.NewReader(bytes.NewReader(contents), int64(len(contents)))
	if err != nil {
		return domain.RemoteTemplateBundle{}, fmt.Errorf("malformed remote template zip archive: %w", err)
	}

	files := make(map[string]string, policy.MaxRegularEntries())
	seen := make(map[string]struct{}, policy.MaxRegularEntries())
	var totalUncompressed uint64

	for _, zipFile := range zipReader.File {
		entry, regular, err := remoteTemplateArchiveEntry(zipFile)
		if err != nil {
			return domain.RemoteTemplateBundle{}, err
		}
		if err := domain.ValidateRemoteTemplateArchiveEntry(policy, entry); err != nil {
			return domain.RemoteTemplateBundle{}, err
		}
		if !regular {
			return domain.RemoteTemplateBundle{}, fmt.Errorf("remote template archive entry is not a regular file: %s", entry.Name)
		}
		if _, exists := seen[entry.Name]; exists {
			return domain.RemoteTemplateBundle{}, fmt.Errorf("remote template archive contains duplicate file: %s", entry.Name)
		}
		seen[entry.Name] = struct{}{}

		contents, err := readZipFile(zipFile, policy.MaxBytesPerFile())
		if err != nil {
			return domain.RemoteTemplateBundle{}, err
		}
		totalUncompressed += uint64(len(contents))
		if totalUncompressed > policy.MaxTotalBytes() {
			return domain.RemoteTemplateBundle{}, fmt.Errorf("remote template archive exceeds maximum uncompressed size %d bytes", policy.MaxTotalBytes())
		}
		files[entry.Name] = string(contents)
	}

	return domain.NewRemoteTemplateBundle(files)
}

func remoteTemplateArchiveEntry(zipFile *zip.File) (domain.RemoteTemplateArchiveEntry, bool, error) {
	info := zipFile.FileInfo()
	mode := info.Mode()
	return domain.RemoteTemplateArchiveEntry{
		Name:             zipFile.Name,
		IsDirectory:      info.IsDir(),
		IsSymlink:        mode&os.ModeSymlink != 0,
		IsExecutable:     mode.Perm()&0o111 != 0,
		UncompressedSize: zipFile.UncompressedSize64,
	}, mode.IsRegular(), nil
}

func readZipFile(zipFile *zip.File, maxBytes uint64) ([]byte, error) {
	reader, err := zipFile.Open()
	if err != nil {
		return nil, fmt.Errorf("read remote template archive file %s: %w", zipFile.Name, err)
	}
	defer reader.Close()

	contents, err := io.ReadAll(io.LimitReader(reader, int64(maxBytes)+1))
	if err != nil {
		return nil, fmt.Errorf("read remote template archive file %s: %w", zipFile.Name, err)
	}
	if uint64(len(contents)) > maxBytes {
		return nil, fmt.Errorf("remote template archive file %s exceeds maximum size %d bytes", zipFile.Name, maxBytes)
	}
	return contents, nil
}
