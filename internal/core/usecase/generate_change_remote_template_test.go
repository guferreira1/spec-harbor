package usecase

import (
	"errors"
	"strings"
	"testing"

	"github.com/guferreira1/spec-harbor/internal/core/domain"
)

func TestGenerateChangeConfigTemplateRemoteAlias(t *testing.T) {
	changeID := "add-service"
	downloadedBytes := []byte("zip bytes")
	checksum := domain.NewRemoteTemplateChecksumFromBytes(downloadedBytes).String()
	fileSystem := newFakeGenerationFileSystem()
	seedGenerationOpenSpecProject(fileSystem)
	configFileSystem := newFakeConfigTemplateConfigFileSystem()
	seedConfigTemplateConfig(configFileSystem)
	parser := &fakeConfigTemplateParser{
		config: configWithRemoteTemplateAlias(t, "service-feature", "https://example.com/templates/service-feature.zip", checksum, "zip"),
	}
	fetcher := &fakeRemoteTemplateFetcher{
		result: domain.NewRemoteTemplateFetchResult(200, downloadedBytes),
	}
	reader := &fakeRemoteTemplateBundleReader{
		bundle: mustRemoteTemplateBundle(t, remoteUsecaseFiles("remote")),
	}
	useCase := newRemoteConfigTemplateGenerateChange(fileSystem, newFakeTemplateChangeContent(), newFakeCustomTemplateFileSystem(), configFileSystem, parser, fetcher, reader)

	result, err := useCase.Execute(GenerateChangeInput{
		ProjectRoot:         "/project",
		ChangeID:            changeID,
		Mode:                domain.TemplateMode,
		ConfigTemplate:      true,
		ConfigTemplateAlias: "service-feature",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if result.ConfigTemplateAlias != "service-feature" {
		t.Fatalf("ConfigTemplateAlias = %q, want service-feature", result.ConfigTemplateAlias)
	}
	if result.ConfigTemplateSource != domain.ConfigTemplateSourceRemote {
		t.Fatalf("ConfigTemplateSource = %q, want remote", result.ConfigTemplateSource)
	}
	if result.RemoteTemplateHost != "example.com" {
		t.Fatalf("RemoteTemplateHost = %q, want example.com", result.RemoteTemplateHost)
	}
	if result.RemoteTemplateFormat != domain.RemoteTemplateFormatZip {
		t.Fatalf("RemoteTemplateFormat = %q, want zip", result.RemoteTemplateFormat)
	}
	if result.ChecksumAlgorithm != domain.ChecksumAlgorithmSHA256 {
		t.Fatalf("ChecksumAlgorithm = %q, want sha256", result.ChecksumAlgorithm)
	}
	assertStringSlicesEqual(t, result.CreatedFiles(), domain.RequiredOpenSpecChangeFiles())
	assertStringSlicesEqual(t, result.SkippedExistingFiles(), nil)

	changePath := openspecChangesDirectory + "/" + changeID
	for _, requiredFile := range domain.RequiredOpenSpecChangeFiles() {
		path := changePath + "/" + requiredFile
		if fileSystem.files[path] != "remote:"+requiredFile {
			t.Fatalf("file %q content = %q, want remote content", path, fileSystem.files[path])
		}
	}
	if fetcher.calls != 1 {
		t.Fatalf("fetcher calls = %d, want 1", fetcher.calls)
	}
	if reader.calls != 1 {
		t.Fatalf("reader calls = %d, want 1", reader.calls)
	}
	if string(reader.contents) != string(downloadedBytes) {
		t.Fatalf("reader contents = %q, want downloaded bytes", string(reader.contents))
	}
}

func TestGenerateChangeConfigTemplateRemoteSkipsExistingFiles(t *testing.T) {
	changeID := "existing-service"
	changePath := openspecChangesDirectory + "/" + changeID
	downloadedBytes := []byte("zip bytes")
	checksum := domain.NewRemoteTemplateChecksumFromBytes(downloadedBytes).String()
	fileSystem := newFakeGenerationFileSystem()
	seedGenerationOpenSpecProject(fileSystem)
	fileSystem.directories[changePath] = true
	fileSystem.files[changePath+"/proposal.md"] = "existing proposal"
	configFileSystem := newFakeConfigTemplateConfigFileSystem()
	seedConfigTemplateConfig(configFileSystem)
	parser := &fakeConfigTemplateParser{
		config: configWithRemoteTemplateAlias(t, "service-feature", "https://example.com/templates/service-feature.zip", checksum, "zip"),
	}
	fetcher := &fakeRemoteTemplateFetcher{result: domain.NewRemoteTemplateFetchResult(200, downloadedBytes)}
	reader := &fakeRemoteTemplateBundleReader{bundle: mustRemoteTemplateBundle(t, remoteUsecaseFiles("remote"))}
	useCase := newRemoteConfigTemplateGenerateChange(fileSystem, newFakeTemplateChangeContent(), newFakeCustomTemplateFileSystem(), configFileSystem, parser, fetcher, reader)

	result, err := useCase.Execute(GenerateChangeInput{
		ProjectRoot:         "/project",
		ChangeID:            changeID,
		Mode:                domain.TemplateMode,
		ConfigTemplate:      true,
		ConfigTemplateAlias: "service-feature",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	assertStringSlicesEqual(t, result.SkippedExistingFiles(), []string{"proposal.md"})
	if fileSystem.files[changePath+"/proposal.md"] != "existing proposal" {
		t.Fatalf("proposal.md = %q, want preserved existing content", fileSystem.files[changePath+"/proposal.md"])
	}
	for _, writtenFile := range fileSystem.writtenFiles {
		if !strings.HasPrefix(writtenFile, changePath+"/") {
			t.Fatalf("write target %q is outside %q", writtenFile, changePath)
		}
	}
}

func TestGenerateChangeConfigTemplateRemoteValidationErrorsDoNotFetchOrWrite(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{name: "missing url", err: errors.New(`invalid config template alias "service-feature": remote template URL is required`), want: "remote template URL is required"},
		{name: "invalid url", err: errors.New(`invalid config template alias "service-feature": remote template URL is invalid`), want: "remote template URL is invalid"},
		{name: "non https", err: errors.New(`invalid config template alias "service-feature": remote template URL must use https`), want: "remote template URL must use https"},
		{name: "credentials", err: errors.New(`invalid config template alias "service-feature": remote template URL must not include credentials`), want: "remote template URL must not include credentials"},
		{name: "query", err: errors.New(`invalid config template alias "service-feature": remote template URL must not include query strings`), want: "remote template URL must not include query strings"},
		{name: "fragment", err: errors.New(`invalid config template alias "service-feature": remote template URL must not include fragments`), want: "remote template URL must not include fragments"},
		{name: "missing checksum", err: errors.New(`invalid config template alias "service-feature": remote template checksum is required`), want: "remote template checksum is required"},
		{name: "unsupported checksum", err: errors.New(`invalid config template alias "service-feature": unsupported remote template checksum algorithm: sha512`), want: "unsupported remote template checksum algorithm"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fileSystem := newFakeGenerationFileSystem()
			seedGenerationOpenSpecProject(fileSystem)
			configFileSystem := newFakeConfigTemplateConfigFileSystem()
			seedConfigTemplateConfig(configFileSystem)
			parser := &fakeConfigTemplateParser{err: test.err}
			fetcher := &fakeRemoteTemplateFetcher{}
			reader := &fakeRemoteTemplateBundleReader{}
			useCase := newRemoteConfigTemplateGenerateChange(fileSystem, newFakeTemplateChangeContent(), newFakeCustomTemplateFileSystem(), configFileSystem, parser, fetcher, reader)

			_, err := useCase.Execute(GenerateChangeInput{
				ProjectRoot:         "/project",
				ChangeID:            "new-change",
				Mode:                domain.TemplateMode,
				ConfigTemplate:      true,
				ConfigTemplateAlias: "service-feature",
			})
			if err == nil {
				t.Fatalf("Execute() error = nil, want %q", test.want)
			}
			if !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Execute() error = %q, want %q", err.Error(), test.want)
			}
			if fetcher.calls != 0 {
				t.Fatalf("fetcher calls = %d, want 0", fetcher.calls)
			}
			if reader.calls != 0 {
				t.Fatalf("reader calls = %d, want 0", reader.calls)
			}
			assertNoGenerationWrites(t, fileSystem)
		})
	}
}

func TestGenerateChangeConfigTemplateRemoteFailuresWriteNothing(t *testing.T) {
	downloadedBytes := []byte("zip bytes")
	checksum := domain.NewRemoteTemplateChecksumFromBytes(downloadedBytes).String()
	tests := []struct {
		name        string
		fetchBytes  []byte
		fetchErr    error
		readerErr   error
		want        string
		wantReader  bool
		wantFetcher bool
	}{
		{name: "network", fetchErr: errors.New("remote template network error: dial failed"), want: "fetch remote template for alias service-feature: remote template network error", wantFetcher: true},
		{name: "timeout", fetchErr: errors.New("remote template fetch timeout: deadline exceeded"), want: "remote template fetch timeout", wantFetcher: true},
		{name: "max size", fetchErr: errors.New("remote template response exceeds maximum size 5242880 bytes"), want: "remote template response exceeds maximum size", wantFetcher: true},
		{name: "checksum mismatch", fetchBytes: []byte("different"), want: "remote template checksum mismatch for alias service-feature", wantFetcher: true},
		{name: "malformed zip", fetchBytes: downloadedBytes, readerErr: errors.New("malformed remote template zip archive"), want: "malformed remote template zip archive", wantFetcher: true, wantReader: true},
		{name: "unsafe archive path", fetchBytes: downloadedBytes, readerErr: errors.New("remote template archive path must not contain traversal: ../proposal.md"), want: "remote template archive path must not contain traversal", wantFetcher: true, wantReader: true},
		{name: "symlink", fetchBytes: downloadedBytes, readerErr: errors.New("remote template archive entry is a symlink: proposal.md"), want: "remote template archive entry is a symlink", wantFetcher: true, wantReader: true},
		{name: "executable", fetchBytes: downloadedBytes, readerErr: errors.New("remote template archive entry is executable: proposal.md"), want: "remote template archive entry is executable", wantFetcher: true, wantReader: true},
		{name: "duplicate", fetchBytes: downloadedBytes, readerErr: errors.New("remote template archive contains duplicate file: proposal.md"), want: "remote template archive contains duplicate file", wantFetcher: true, wantReader: true},
		{name: "extra", fetchBytes: downloadedBytes, readerErr: errors.New("remote template archive contains unsupported file: README.md"), want: "remote template archive contains unsupported file", wantFetcher: true, wantReader: true},
		{name: "missing", fetchBytes: downloadedBytes, readerErr: errors.New("remote template archive is missing required files: risks.md"), want: "remote template archive is missing required files", wantFetcher: true, wantReader: true},
		{name: "empty", fetchBytes: downloadedBytes, readerErr: errors.New("remote template file risks.md is empty"), want: "remote template file risks.md is empty", wantFetcher: true, wantReader: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fileSystem := newFakeGenerationFileSystem()
			seedGenerationOpenSpecProject(fileSystem)
			configFileSystem := newFakeConfigTemplateConfigFileSystem()
			seedConfigTemplateConfig(configFileSystem)
			parser := &fakeConfigTemplateParser{
				config: configWithRemoteTemplateAlias(t, "service-feature", "https://example.com/templates/service-feature.zip", checksum, "zip"),
			}
			if test.fetchBytes == nil {
				test.fetchBytes = downloadedBytes
			}
			fetcher := &fakeRemoteTemplateFetcher{
				result: domain.NewRemoteTemplateFetchResult(200, test.fetchBytes),
				err:    test.fetchErr,
			}
			reader := &fakeRemoteTemplateBundleReader{
				bundle: mustRemoteTemplateBundle(t, remoteUsecaseFiles("remote")),
				err:    test.readerErr,
			}
			useCase := newRemoteConfigTemplateGenerateChange(fileSystem, newFakeTemplateChangeContent(), newFakeCustomTemplateFileSystem(), configFileSystem, parser, fetcher, reader)

			_, err := useCase.Execute(GenerateChangeInput{
				ProjectRoot:         "/project",
				ChangeID:            "new-change",
				Mode:                domain.TemplateMode,
				ConfigTemplate:      true,
				ConfigTemplateAlias: "service-feature",
			})
			if err == nil {
				t.Fatalf("Execute() error = nil, want %q", test.want)
			}
			if !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Execute() error = %q, want %q", err.Error(), test.want)
			}
			if (fetcher.calls > 0) != test.wantFetcher {
				t.Fatalf("fetcher calls = %d, want called %t", fetcher.calls, test.wantFetcher)
			}
			if (reader.calls > 0) != test.wantReader {
				t.Fatalf("reader calls = %d, want called %t", reader.calls, test.wantReader)
			}
			assertNoGenerationWrites(t, fileSystem)
			if fileSystem.directories[openspecChangesDirectory+"/new-change"] {
				t.Fatalf("change directory was created on remote failure")
			}
		})
	}
}

func TestGenerateChangeConfigTemplateRemoteChecksumMismatchReportsDigestFacts(t *testing.T) {
	expectedBytes := []byte("expected zip bytes")
	actualBytes := []byte("actual zip bytes")
	expectedChecksum := domain.NewRemoteTemplateChecksumFromBytes(expectedBytes)
	actualChecksum := domain.NewRemoteTemplateChecksumFromBytes(actualBytes)
	fileSystem := newFakeGenerationFileSystem()
	seedGenerationOpenSpecProject(fileSystem)
	configFileSystem := newFakeConfigTemplateConfigFileSystem()
	seedConfigTemplateConfig(configFileSystem)
	parser := &fakeConfigTemplateParser{
		config: configWithRemoteTemplateAlias(
			t,
			"service-feature",
			"https://example.com/templates/service-feature.zip",
			expectedChecksum.String(),
			"zip",
		),
	}
	fetcher := &fakeRemoteTemplateFetcher{
		result: domain.NewRemoteTemplateFetchResult(200, actualBytes),
	}
	reader := &fakeRemoteTemplateBundleReader{
		bundle: mustRemoteTemplateBundle(t, remoteUsecaseFiles("remote")),
	}
	useCase := newRemoteConfigTemplateGenerateChange(fileSystem, newFakeTemplateChangeContent(), newFakeCustomTemplateFileSystem(), configFileSystem, parser, fetcher, reader)

	_, err := useCase.Execute(GenerateChangeInput{
		ProjectRoot:         "/project",
		ChangeID:            "new-change",
		Mode:                domain.TemplateMode,
		ConfigTemplate:      true,
		ConfigTemplateAlias: "service-feature",
	})
	if err == nil {
		t.Fatalf("Execute() error = nil, want checksum mismatch")
	}
	for _, want := range []string{
		"remote template checksum mismatch for alias service-feature",
		"expected " + expectedChecksum.String(),
		"got " + actualChecksum.String(),
		string(domain.ChecksumAlgorithmSHA256),
	} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("Execute() error = %q, want %q", err.Error(), want)
		}
	}
	if reader.calls != 0 {
		t.Fatalf("reader calls = %d, want 0 before archive parsing", reader.calls)
	}
	assertNoGenerationWrites(t, fileSystem)
	if fileSystem.directories[openspecChangesDirectory+"/new-change"] {
		t.Fatalf("change directory was created on checksum mismatch")
	}
}

func TestGenerateChangeConfigTemplateRemoteRequiresOpenSpecProjectBeforeFetch(t *testing.T) {
	downloadedBytes := []byte("zip bytes")
	checksum := domain.NewRemoteTemplateChecksumFromBytes(downloadedBytes).String()
	fileSystem := newFakeGenerationFileSystem()
	configFileSystem := newFakeConfigTemplateConfigFileSystem()
	seedConfigTemplateConfig(configFileSystem)
	parser := &fakeConfigTemplateParser{
		config: configWithRemoteTemplateAlias(t, "service-feature", "https://example.com/templates/service-feature.zip", checksum, "zip"),
	}
	fetcher := &fakeRemoteTemplateFetcher{result: domain.NewRemoteTemplateFetchResult(200, downloadedBytes)}
	reader := &fakeRemoteTemplateBundleReader{bundle: mustRemoteTemplateBundle(t, remoteUsecaseFiles("remote"))}
	useCase := newRemoteConfigTemplateGenerateChange(fileSystem, newFakeTemplateChangeContent(), newFakeCustomTemplateFileSystem(), configFileSystem, parser, fetcher, reader)

	_, err := useCase.Execute(GenerateChangeInput{
		ProjectRoot:         "/project",
		ChangeID:            "new-change",
		Mode:                domain.TemplateMode,
		ConfigTemplate:      true,
		ConfigTemplateAlias: "service-feature",
	})
	if err == nil || err.Error() != "OpenSpec project structure is missing. Run specharbor init first." {
		t.Fatalf("Execute() error = %v, want missing OpenSpec project", err)
	}
	if fetcher.calls != 0 {
		t.Fatalf("fetcher calls = %d, want 0", fetcher.calls)
	}
	if reader.calls != 0 {
		t.Fatalf("reader calls = %d, want 0", reader.calls)
	}
	assertNoGenerationWrites(t, fileSystem)
}

func TestGenerateChangeDoesNotFetchRemoteForNonRemoteModes(t *testing.T) {
	fileSystem := newFakeGenerationFileSystem()
	seedGenerationOpenSpecProject(fileSystem)
	templateContent := newFakeTemplateChangeContent()
	customTemplateFiles := newFakeCustomTemplateFileSystem()
	seedCustomTemplate(customTemplateFiles, "api-feature")
	configFileSystem := newFakeConfigTemplateConfigFileSystem()
	seedConfigTemplateConfig(configFileSystem)
	fetcher := &fakeRemoteTemplateFetcher{}
	reader := &fakeRemoteTemplateBundleReader{bundle: mustRemoteTemplateBundle(t, remoteUsecaseFiles("remote"))}
	parser := &fakeConfigTemplateParser{
		config: configWithTemplateAlias(t, "default-feature", "builtin", "feature"),
	}
	useCase := newRemoteConfigTemplateGenerateChange(fileSystem, templateContent, customTemplateFiles, configFileSystem, parser, fetcher, reader)

	requests := []GenerateChangeInput{
		{ProjectRoot: "/project", ChangeID: "direct-built-in", Mode: domain.TemplateMode, TemplateName: "feature"},
		{ProjectRoot: "/project", ChangeID: "direct-custom", Mode: domain.TemplateMode, TemplateSource: domain.CustomTemplateSource, CustomTemplateName: "api-feature"},
		{ProjectRoot: "/project", ChangeID: "config-built-in", Mode: domain.TemplateMode, ConfigTemplate: true, ConfigTemplateAlias: "default-feature"},
	}
	for _, request := range requests {
		if _, err := useCase.Execute(request); err != nil {
			t.Fatalf("Execute(%s) error = %v", request.ChangeID, err)
		}
	}

	parser.config = configWithTemplateAlias(t, "api-feature", "custom", "api-feature")
	if _, err := useCase.Execute(GenerateChangeInput{
		ProjectRoot:         "/project",
		ChangeID:            "config-custom",
		Mode:                domain.TemplateMode,
		ConfigTemplate:      true,
		ConfigTemplateAlias: "api-feature",
	}); err != nil {
		t.Fatalf("Execute(config custom) error = %v", err)
	}

	if fetcher.calls != 0 {
		t.Fatalf("fetcher calls = %d, want 0", fetcher.calls)
	}
	if reader.calls != 0 {
		t.Fatalf("reader calls = %d, want 0", reader.calls)
	}
}

func newRemoteConfigTemplateGenerateChange(
	fileSystem *fakeGenerationFileSystem,
	templateContent *fakeTemplateChangeContent,
	customTemplateFiles *fakeCustomTemplateFileSystem,
	configFileSystem *fakeConfigTemplateConfigFileSystem,
	parser *fakeConfigTemplateParser,
	fetcher *fakeRemoteTemplateFetcher,
	reader *fakeRemoteTemplateBundleReader,
) *GenerateChange {
	return NewGenerateChangeWithRemoteConfigTemplates(
		fileSystem,
		newFakeBlankChangeContent(),
		templateContent,
		newFakeGuidedChangeContent(),
		customTemplateFiles,
		configFileSystem,
		parser,
		fetcher,
		reader,
	)
}

func configWithRemoteTemplateAlias(t *testing.T, aliasValue string, rawURL string, checksum string, format string) domain.LocalConfig {
	t.Helper()

	alias, err := domain.NewConfigTemplateAlias(aliasValue)
	if err != nil {
		t.Fatalf("NewConfigTemplateAlias(%q) error = %v", aliasValue, err)
	}
	reference, err := domain.NewConfigTemplateReference(domain.ConfigTemplateReferenceInput{
		Alias:                 alias,
		Source:                "remote",
		URL:                   rawURL,
		URLFieldProvided:      true,
		Checksum:              checksum,
		ChecksumFieldProvided: true,
		Format:                format,
		FormatFieldProvided:   true,
	})
	if err != nil {
		t.Fatalf("NewConfigTemplateReference(remote) error = %v", err)
	}
	return domain.LocalConfig{
		Version:   domain.SupportedLocalConfigVersion,
		Templates: domain.NewConfigTemplates(domain.NewConfigTemplateAliases([]domain.ConfigTemplateReference{reference})),
	}
}

func remoteUsecaseFiles(prefix string) map[string]string {
	files := make(map[string]string)
	for _, requiredFile := range domain.RequiredOpenSpecChangeFiles() {
		files[requiredFile] = prefix + ":" + requiredFile
	}
	return files
}

func mustRemoteTemplateBundle(t *testing.T, files map[string]string) domain.RemoteTemplateBundle {
	t.Helper()

	bundle, err := domain.NewRemoteTemplateBundle(files)
	if err != nil {
		t.Fatalf("NewRemoteTemplateBundle() error = %v", err)
	}
	return bundle
}

type fakeRemoteTemplateFetcher struct {
	result   domain.RemoteTemplateFetchResult
	err      error
	calls    int
	requests []domain.RemoteTemplateFetchRequest
}

func (fetcher *fakeRemoteTemplateFetcher) FetchRemoteTemplate(
	request domain.RemoteTemplateFetchRequest,
) (domain.RemoteTemplateFetchResult, error) {
	fetcher.calls++
	fetcher.requests = append(fetcher.requests, request)
	if fetcher.err != nil {
		return domain.RemoteTemplateFetchResult{}, fetcher.err
	}
	return fetcher.result, nil
}

type fakeRemoteTemplateBundleReader struct {
	bundle   domain.RemoteTemplateBundle
	err      error
	calls    int
	contents []byte
}

func (reader *fakeRemoteTemplateBundleReader) ReadRemoteTemplateBundle(
	contents []byte,
	policy domain.RemoteTemplateArchivePolicy,
) (domain.RemoteTemplateBundle, error) {
	reader.calls++
	reader.contents = append([]byte(nil), contents...)
	if len(policy.RequiredFiles()) != len(domain.RequiredOpenSpecChangeFiles()) {
		return domain.RemoteTemplateBundle{}, errors.New("unexpected policy")
	}
	if reader.err != nil {
		return domain.RemoteTemplateBundle{}, reader.err
	}
	return reader.bundle, nil
}
