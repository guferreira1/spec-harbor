package remote

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/guferreira1/spec-harbor/internal/core/domain"
)

func TestHTTPTemplateFetcherSuccessfulGET(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method = %s, want GET", r.Method)
		}
		fmt.Fprint(w, "zip bytes")
	}))
	defer server.Close()

	result, err := fetchFromTestServer(t, server, domain.MaxRemoteTemplateHTTPResponseBytes)
	if err != nil {
		t.Fatalf("FetchRemoteTemplate() error = %v", err)
	}
	if result.StatusCode() != http.StatusOK {
		t.Fatalf("StatusCode = %d, want 200", result.StatusCode())
	}
	if string(result.Body()) != "zip bytes" {
		t.Fatalf("Body = %q, want zip bytes", string(result.Body()))
	}
}

func TestHTTPTemplateFetcherRejectsRedirects(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/other.zip", http.StatusFound)
	}))
	defer server.Close()

	_, err := fetchFromTestServer(t, server, domain.MaxRemoteTemplateHTTPResponseBytes)
	if err == nil || !strings.Contains(err.Error(), "remote template redirects are unsupported: HTTP 302") {
		t.Fatalf("FetchRemoteTemplate() error = %v, want redirect rejection", err)
	}
}

func TestHTTPTemplateFetcherRejectsNon200Responses(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "missing", http.StatusNotFound)
	}))
	defer server.Close()

	_, err := fetchFromTestServer(t, server, domain.MaxRemoteTemplateHTTPResponseBytes)
	if err == nil || !strings.Contains(err.Error(), "remote template fetch failed: HTTP 404") {
		t.Fatalf("FetchRemoteTemplate() error = %v, want non-200 rejection", err)
	}
}

func TestHTTPTemplateFetcherEnforcesMaxDownloadSize(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, "123456")
	}))
	defer server.Close()

	_, err := fetchFromTestServer(t, server, 5)
	if err == nil || !strings.Contains(err.Error(), "remote template response exceeds maximum size 5 bytes") {
		t.Fatalf("FetchRemoteTemplate() error = %v, want max size rejection", err)
	}
}

func TestHTTPTemplateFetcherReportsTimeout(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(100 * time.Millisecond)
		fmt.Fprint(w, "late")
	}))
	defer server.Close()

	client := server.Client()
	client.Timeout = 10 * time.Millisecond
	fetcher := NewHTTPTemplateFetcherWithClient(client, domain.MaxRemoteTemplateHTTPResponseBytes)
	remoteURL := mustRemoteTemplateURL(t, server.URL+"/template.zip")

	_, err := fetcher.FetchRemoteTemplate(domain.NewRemoteTemplateFetchRequest(remoteURL))
	if err == nil || !strings.Contains(err.Error(), "remote template fetch timeout") {
		t.Fatalf("FetchRemoteTemplate() error = %v, want timeout", err)
	}
}

func TestHTTPTemplateFetcherReportsNetworkError(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, "unused")
	}))
	remoteURL := mustRemoteTemplateURL(t, server.URL+"/template.zip")
	client := server.Client()
	server.Close()

	fetcher := NewHTTPTemplateFetcherWithClient(client, domain.MaxRemoteTemplateHTTPResponseBytes)
	_, err := fetcher.FetchRemoteTemplate(domain.NewRemoteTemplateFetchRequest(remoteURL))
	if err == nil || !strings.Contains(err.Error(), "remote template network error") {
		t.Fatalf("FetchRemoteTemplate() error = %v, want network error", err)
	}
}

func TestHTTPTemplateFetcherSendsNoCredentialsAuthHeadersOrCookies(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "" {
			t.Fatalf("Authorization header = %q, want empty", r.Header.Get("Authorization"))
		}
		if r.Header.Get("Cookie") != "" {
			t.Fatalf("Cookie header = %q, want empty", r.Header.Get("Cookie"))
		}
		if len(r.Cookies()) != 0 {
			t.Fatalf("cookies = %v, want none", r.Cookies())
		}
		fmt.Fprint(w, "ok")
	}))
	defer server.Close()

	if _, err := fetchFromTestServer(t, server, domain.MaxRemoteTemplateHTTPResponseBytes); err != nil {
		t.Fatalf("FetchRemoteTemplate() error = %v", err)
	}
}

func fetchFromTestServer(t *testing.T, server *httptest.Server, maxBytes int64) (domain.RemoteTemplateFetchResult, error) {
	t.Helper()

	fetcher := NewHTTPTemplateFetcherWithClient(server.Client(), maxBytes)
	remoteURL := mustRemoteTemplateURL(t, server.URL+"/template.zip")
	return fetcher.FetchRemoteTemplate(domain.NewRemoteTemplateFetchRequest(remoteURL))
}

func mustRemoteTemplateURL(t *testing.T, raw string) domain.RemoteTemplateURL {
	t.Helper()

	remoteURL, err := domain.NewRemoteTemplateURL(raw)
	if err != nil {
		t.Fatalf("NewRemoteTemplateURL(%q) error = %v", raw, err)
	}
	return remoteURL
}
