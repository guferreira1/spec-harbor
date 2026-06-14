package github

import (
	"encoding/base64"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/guferreira1/spec-harbor/internal/core/domain"
)

func TestHTTPRemoteContextClientResolvesRepositoryWithSafeHeaders(t *testing.T) {
	transport := &fakeRoundTripper{
		response: `{"default_branch":"main"}`,
	}
	client := NewHTTPRemoteContextClientWithOptions(RemoteContextClientOptions{
		Token:      "secret-token",
		BaseURL:    "https://api.github.com",
		HTTPClient: &http.Client{Transport: transport},
	})
	locator := mustGitHubLocator(t, "owner/repo")

	repository, err := client.ResolveRepository(locator)
	if err != nil {
		t.Fatalf("ResolveRepository() error = %v", err)
	}
	if repository.DefaultBranch != "main" {
		t.Fatalf("DefaultBranch = %q, want main", repository.DefaultBranch)
	}
	if transport.method != http.MethodGet {
		t.Fatalf("method = %q, want GET", transport.method)
	}
	if transport.url != "https://api.github.com/repos/owner/repo" {
		t.Fatalf("url = %q", transport.url)
	}
	if transport.authorization != "Bearer secret-token" {
		t.Fatalf("authorization header missing")
	}
	if transport.userAgent != "specharbor" {
		t.Fatalf("user agent = %q", transport.userAgent)
	}
}

func TestHTTPRemoteContextClientResolvesRefAndEscapesSlash(t *testing.T) {
	transport := &fakeRoundTripper{response: `{"sha":"abc123"}`}
	client := NewHTTPRemoteContextClientWithOptions(RemoteContextClientOptions{
		BaseURL:    "https://api.github.com",
		HTTPClient: &http.Client{Transport: transport},
	})
	ref, err := domain.NewGitHubRemoteRef("feature/context", domain.DefaultGitHubRemoteContextLimits())
	if err != nil {
		t.Fatalf("NewGitHubRemoteRef() error = %v", err)
	}
	resolved, err := client.ResolveRef(mustGitHubLocator(t, "owner/repo"), ref)
	if err != nil {
		t.Fatalf("ResolveRef() error = %v", err)
	}
	if resolved.CommitSHA != "abc123" {
		t.Fatalf("CommitSHA = %q", resolved.CommitSHA)
	}
	if !strings.Contains(transport.url, "feature%2Fcontext") {
		t.Fatalf("url = %q, want escaped slash", transport.url)
	}
}

func TestHTTPRemoteContextClientListsDirectoryAndReadsFile(t *testing.T) {
	encoded := base64.StdEncoding.EncodeToString([]byte("architecture context"))
	transport := &sequenceRoundTripper{responses: []string{
		`[{"path":"docs/architecture.md","type":"file","size":20},{"path":"docs/nested","type":"dir","size":0}]`,
		`{"path":"docs/architecture.md","type":"file","size":20,"encoding":"base64","content":"` + encoded + `"}`,
	}}
	client := NewHTTPRemoteContextClientWithOptions(RemoteContextClientOptions{
		BaseURL:    "https://api.github.com",
		HTTPClient: &http.Client{Transport: transport},
	})
	ref := domain.GitHubRemoteResolvedRef{CommitSHA: "abc123"}
	locator := mustGitHubLocator(t, "owner/repo")

	entries, err := client.ListDirectory(locator, ref, "docs")
	if err != nil {
		t.Fatalf("ListDirectory() error = %v", err)
	}
	if len(entries) != 2 || entries[0].Path != "docs/architecture.md" || entries[1].Type != "dir" {
		t.Fatalf("entries = %#v", entries)
	}
	file, err := client.ReadFile(locator, ref, "docs/architecture.md", 128*1024)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(file.Contents) != "architecture context" {
		t.Fatalf("Contents = %q", string(file.Contents))
	}
	if !strings.Contains(transport.urls[0], "/contents/docs?ref=abc123") {
		t.Fatalf("list URL = %q", transport.urls[0])
	}
	if !strings.Contains(transport.urls[1], "/contents/docs/architecture.md?ref=abc123") {
		t.Fatalf("file URL = %q", transport.urls[1])
	}
}

func TestHTTPRemoteContextClientMapsStatusCodesWithoutLeakingToken(t *testing.T) {
	for _, test := range []struct {
		name       string
		statusCode int
		header     http.Header
		wantCode   domain.GitHubRemoteContextErrorCode
	}{
		{name: "invalid token", statusCode: http.StatusUnauthorized, wantCode: domain.GitHubRemoteContextErrorInvalidToken},
		{name: "rate limit", statusCode: http.StatusForbidden, header: http.Header{"X-Ratelimit-Remaining": []string{"0"}}, wantCode: domain.GitHubRemoteContextErrorRateLimit},
		{name: "forbidden", statusCode: http.StatusForbidden, wantCode: domain.GitHubRemoteContextErrorForbidden},
		{name: "not found", statusCode: http.StatusNotFound, wantCode: domain.GitHubRemoteContextErrorNotFound},
	} {
		t.Run(test.name, func(t *testing.T) {
			transport := &fakeRoundTripper{
				statusCode: test.statusCode,
				header:     test.header,
				response:   `{"message":"secret-token should not be included"}`,
			}
			client := NewHTTPRemoteContextClientWithOptions(RemoteContextClientOptions{
				Token:      "secret-token",
				BaseURL:    "https://api.github.com",
				HTTPClient: &http.Client{Transport: transport},
			})
			_, err := client.ResolveRepository(mustGitHubLocator(t, "owner/repo"))
			var remoteErr domain.GitHubRemoteContextError
			if !errors.As(err, &remoteErr) {
				t.Fatalf("error = %T %v, want GitHubRemoteContextError", err, err)
			}
			if remoteErr.Code != test.wantCode {
				t.Fatalf("Code = %q, want %q", remoteErr.Code, test.wantCode)
			}
			if strings.Contains(remoteErr.Error(), "secret-token") {
				t.Fatalf("error leaked token: %q", remoteErr.Error())
			}
		})
	}
}

func TestHTTPRemoteContextClientBoundsResponsesAndFileSize(t *testing.T) {
	transport := &fakeRoundTripper{response: strings.Repeat("x", 20)}
	client := NewHTTPRemoteContextClientWithOptions(RemoteContextClientOptions{
		BaseURL:                  "https://api.github.com",
		HTTPClient:               &http.Client{Transport: transport},
		MaxMetadataResponseBytes: 10,
	})
	_, err := client.ResolveRepository(mustGitHubLocator(t, "owner/repo"))
	var remoteErr domain.GitHubRemoteContextError
	if !errors.As(err, &remoteErr) || remoteErr.Code != domain.GitHubRemoteContextErrorOversizedResponse {
		t.Fatalf("error = %T %v, want oversized response", err, err)
	}

	encoded := base64.StdEncoding.EncodeToString([]byte(strings.Repeat("a", 20)))
	transport = &fakeRoundTripper{response: `{"path":"README.md","type":"file","size":20,"encoding":"base64","content":"` + encoded + `"}`}
	client = NewHTTPRemoteContextClientWithOptions(RemoteContextClientOptions{
		BaseURL:    "https://api.github.com",
		HTTPClient: &http.Client{Transport: transport},
	})
	_, err = client.ReadFile(mustGitHubLocator(t, "owner/repo"), domain.GitHubRemoteResolvedRef{CommitSHA: "abc123"}, "README.md", 10)
	if !errors.As(err, &remoteErr) || remoteErr.Code != domain.GitHubRemoteContextErrorOversizedResponse {
		t.Fatalf("error = %T %v, want oversized file", err, err)
	}
}

func TestHTTPRemoteContextClientMapsTransportFailure(t *testing.T) {
	client := NewHTTPRemoteContextClientWithOptions(RemoteContextClientOptions{
		BaseURL: "https://api.github.com",
		HTTPClient: &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			return nil, errors.New("dial secret-token")
		})},
	})
	_, err := client.ResolveRepository(mustGitHubLocator(t, "owner/repo"))
	var remoteErr domain.GitHubRemoteContextError
	if !errors.As(err, &remoteErr) || remoteErr.Code != domain.GitHubRemoteContextErrorNetwork {
		t.Fatalf("error = %T %v, want network", err, err)
	}
	if strings.Contains(remoteErr.Error(), "secret-token") {
		t.Fatalf("error leaked transport detail: %q", remoteErr.Error())
	}
}

type fakeRoundTripper struct {
	statusCode    int
	header        http.Header
	response      string
	method        string
	url           string
	authorization string
	userAgent     string
}

func (transport *fakeRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	transport.method = request.Method
	transport.url = request.URL.String()
	transport.authorization = request.Header.Get("Authorization")
	transport.userAgent = request.Header.Get("User-Agent")
	statusCode := transport.statusCode
	if statusCode == 0 {
		statusCode = http.StatusOK
	}
	header := transport.header
	if header == nil {
		header = make(http.Header)
	}
	return &http.Response{
		StatusCode: statusCode,
		Header:     header,
		Body:       io.NopCloser(strings.NewReader(transport.response)),
	}, nil
}

type sequenceRoundTripper struct {
	responses []string
	urls      []string
	index     int
}

func (transport *sequenceRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	if transport.index >= len(transport.responses) {
		return nil, errors.New("unexpected request")
	}
	response := transport.responses[transport.index]
	transport.index++
	transport.urls = append(transport.urls, request.URL.String())
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(response)),
	}, nil
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return fn(request)
}

func mustGitHubLocator(t *testing.T, raw string) domain.GitHubRepositoryLocator {
	t.Helper()
	locator, err := domain.NewGitHubRepositoryLocator(raw, domain.DefaultGitHubRemoteContextLimits())
	if err != nil {
		t.Fatalf("NewGitHubRepositoryLocator() error = %v", err)
	}
	return locator
}
