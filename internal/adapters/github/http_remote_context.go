package github

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	neturl "net/url"
	"path"
	"strings"
	"time"

	"github.com/guferreira1/spec-harbor/internal/core/domain"
)

const (
	defaultGitHubAPIBaseURL            = "https://api.github.com"
	defaultGitHubMetadataResponseBytes = 1024 * 1024
	defaultGitHubFileResponseBytes     = 256 * 1024
	defaultGitHubTimeout               = 10 * time.Second
)

type RemoteContextClientOptions struct {
	Token                    string
	BaseURL                  string
	HTTPClient               *http.Client
	Timeout                  time.Duration
	MaxMetadataResponseBytes int64
	MaxFileResponseBytes     int64
}

type HTTPRemoteContextClient struct {
	token                    string
	baseURL                  string
	httpClient               *http.Client
	maxMetadataResponseBytes int64
	maxFileResponseBytes     int64
}

func NewHTTPRemoteContextClient(token string) *HTTPRemoteContextClient {
	return NewHTTPRemoteContextClientWithOptions(RemoteContextClientOptions{Token: token})
}

func NewHTTPRemoteContextClientWithOptions(options RemoteContextClientOptions) *HTTPRemoteContextClient {
	baseURL := strings.TrimRight(strings.TrimSpace(options.BaseURL), "/")
	if baseURL == "" {
		baseURL = defaultGitHubAPIBaseURL
	}
	timeout := options.Timeout
	if timeout <= 0 {
		timeout = defaultGitHubTimeout
	}
	httpClient := options.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: timeout}
	} else if httpClient.Timeout == 0 {
		copyClient := *httpClient
		copyClient.Timeout = timeout
		httpClient = &copyClient
	}
	maxMetadataBytes := options.MaxMetadataResponseBytes
	if maxMetadataBytes <= 0 {
		maxMetadataBytes = defaultGitHubMetadataResponseBytes
	}
	maxFileBytes := options.MaxFileResponseBytes
	if maxFileBytes <= 0 {
		maxFileBytes = defaultGitHubFileResponseBytes
	}
	return &HTTPRemoteContextClient{
		token:                    strings.TrimSpace(options.Token),
		baseURL:                  baseURL,
		httpClient:               httpClient,
		maxMetadataResponseBytes: maxMetadataBytes,
		maxFileResponseBytes:     maxFileBytes,
	}
}

func (client *HTTPRemoteContextClient) ResolveRepository(
	locator domain.GitHubRepositoryLocator,
) (domain.GitHubRemoteRepository, error) {
	var response repositoryResponse
	if err := client.getJSON(client.apiPath("repos", locator.Owner(), locator.Name()), client.maxMetadataResponseBytes, &response); err != nil {
		return domain.GitHubRemoteRepository{}, err
	}
	if strings.TrimSpace(response.DefaultBranch) == "" {
		return domain.GitHubRemoteRepository{}, domain.GitHubRemoteContextError{
			Code:    domain.GitHubRemoteContextErrorInvalidResponse,
			Message: "GitHub repository response did not include a default branch",
		}
	}
	return domain.GitHubRemoteRepository{
		Locator:       locator,
		DefaultBranch: strings.TrimSpace(response.DefaultBranch),
	}, nil
}

func (client *HTTPRemoteContextClient) ResolveRef(
	locator domain.GitHubRepositoryLocator,
	ref domain.GitHubRemoteRef,
) (domain.GitHubRemoteResolvedRef, error) {
	var response commitResponse
	if err := client.getJSON(client.apiPath("repos", locator.Owner(), locator.Name(), "commits", ref.Value()), client.maxMetadataResponseBytes, &response); err != nil {
		return domain.GitHubRemoteResolvedRef{}, err
	}
	if strings.TrimSpace(response.SHA) == "" {
		return domain.GitHubRemoteResolvedRef{}, domain.GitHubRemoteContextError{
			Code:    domain.GitHubRemoteContextErrorInvalidResponse,
			Message: "GitHub ref response did not include a commit SHA",
		}
	}
	return domain.GitHubRemoteResolvedRef{
		RequestedRef: ref.Value(),
		ResolvedRef:  ref.Value(),
		CommitSHA:    strings.TrimSpace(response.SHA),
	}, nil
}

func (client *HTTPRemoteContextClient) ListDirectory(
	locator domain.GitHubRepositoryLocator,
	ref domain.GitHubRemoteResolvedRef,
	relativePath string,
) ([]domain.GitHubRemoteEntry, error) {
	endpoint := client.contentsPath(locator, relativePath, ref.CommitSHA)
	var response []contentsResponse
	if err := client.getJSON(endpoint, client.maxMetadataResponseBytes, &response); err != nil {
		return nil, err
	}
	entries := make([]domain.GitHubRemoteEntry, 0, len(response))
	for _, item := range response {
		if strings.TrimSpace(item.Path) == "" {
			continue
		}
		entries = append(entries, domain.GitHubRemoteEntry{
			Path:      item.Path,
			Type:      item.Type,
			SizeBytes: item.Size,
		})
	}
	return entries, nil
}

func (client *HTTPRemoteContextClient) ReadFile(
	locator domain.GitHubRepositoryLocator,
	ref domain.GitHubRemoteResolvedRef,
	relativePath string,
	maxBytes int64,
) (domain.GitHubRemoteFile, error) {
	var response contentsResponse
	if err := client.getJSON(client.contentsPath(locator, relativePath, ref.CommitSHA), client.maxFileResponseBytes, &response); err != nil {
		return domain.GitHubRemoteFile{}, err
	}
	if response.Type != "file" {
		return domain.GitHubRemoteFile{}, domain.GitHubRemoteContextError{
			Code:    domain.GitHubRemoteContextErrorUnsupportedContent,
			Message: "GitHub content is not a file",
		}
	}
	if response.Size > maxBytes {
		return domain.GitHubRemoteFile{}, domain.GitHubRemoteContextError{
			Code:    domain.GitHubRemoteContextErrorOversizedResponse,
			Message: "GitHub file exceeds configured size limit",
		}
	}
	if strings.TrimSpace(response.Encoding) != "base64" {
		return domain.GitHubRemoteFile{}, domain.GitHubRemoteContextError{
			Code:    domain.GitHubRemoteContextErrorUnsupportedContent,
			Message: "GitHub file content uses unsupported encoding",
		}
	}
	encoded := strings.NewReplacer("\n", "", "\r", "", " ", "").Replace(response.Content)
	contents, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return domain.GitHubRemoteFile{}, domain.GitHubRemoteContextError{
			Code:    domain.GitHubRemoteContextErrorInvalidResponse,
			Message: "GitHub file content could not be decoded",
		}
	}
	if int64(len(contents)) > maxBytes {
		return domain.GitHubRemoteFile{}, domain.GitHubRemoteContextError{
			Code:    domain.GitHubRemoteContextErrorOversizedResponse,
			Message: "GitHub file exceeds configured size limit",
		}
	}
	return domain.GitHubRemoteFile{
		Path:      response.Path,
		SizeBytes: int64(len(contents)),
		Contents:  contents,
	}, nil
}

func (client *HTTPRemoteContextClient) getJSON(endpoint string, maxBytes int64, target any) error {
	request, err := http.NewRequestWithContext(context.Background(), http.MethodGet, endpoint, nil)
	if err != nil {
		return domain.GitHubRemoteContextError{
			Code:    domain.GitHubRemoteContextErrorInvalidResponse,
			Message: "GitHub request could not be created",
		}
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("User-Agent", "specharbor")
	if client.token != "" {
		request.Header.Set("Authorization", "Bearer "+client.token)
	}

	response, err := client.httpClient.Do(request)
	if err != nil {
		return mapGitHubTransportError(err)
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return client.mapStatusError(response)
	}

	contents, err := readBoundedGitHubResponse(response.Body, maxBytes)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(contents, target); err != nil {
		return domain.GitHubRemoteContextError{
			Code:    domain.GitHubRemoteContextErrorInvalidResponse,
			Message: "GitHub response was not valid JSON",
		}
	}
	return nil
}

func (client *HTTPRemoteContextClient) mapStatusError(response *http.Response) error {
	switch response.StatusCode {
	case http.StatusUnauthorized:
		if client.token != "" {
			return domain.GitHubRemoteContextError{
				Code:    domain.GitHubRemoteContextErrorInvalidToken,
				Message: "GitHub token is invalid or unauthorized",
			}
		}
		return domain.GitHubRemoteContextError{
			Code:    domain.GitHubRemoteContextErrorUnauthorized,
			Message: "GitHub request is unauthorized",
		}
	case http.StatusForbidden:
		if response.Header.Get("X-RateLimit-Remaining") == "0" {
			return domain.GitHubRemoteContextError{
				Code:    domain.GitHubRemoteContextErrorRateLimit,
				Message: "GitHub API rate limit exceeded",
			}
		}
		return domain.GitHubRemoteContextError{
			Code:    domain.GitHubRemoteContextErrorForbidden,
			Message: "GitHub request is forbidden",
		}
	case http.StatusNotFound:
		return domain.GitHubRemoteContextError{
			Code:    domain.GitHubRemoteContextErrorNotFound,
			Message: "GitHub resource was not found",
		}
	default:
		return domain.GitHubRemoteContextError{
			Code:    domain.GitHubRemoteContextErrorInvalidResponse,
			Message: fmt.Sprintf("GitHub request failed with status %d", response.StatusCode),
		}
	}
}

func (client *HTTPRemoteContextClient) apiPath(parts ...string) string {
	escaped := make([]string, 0, len(parts))
	for _, part := range parts {
		escaped = append(escaped, neturl.PathEscape(part))
	}
	return client.baseURL + "/" + strings.Join(escaped, "/")
}

func (client *HTTPRemoteContextClient) contentsPath(
	locator domain.GitHubRepositoryLocator,
	relativePath string,
	ref string,
) string {
	endpoint := client.baseURL + "/repos/" +
		neturl.PathEscape(locator.Owner()) + "/" +
		neturl.PathEscape(locator.Name()) + "/contents"
	if strings.TrimSpace(relativePath) != "" {
		endpoint += "/" + escapeGitHubPath(relativePath)
	}
	if strings.TrimSpace(ref) != "" {
		endpoint += "?ref=" + neturl.QueryEscape(ref)
	}
	return endpoint
}

func escapeGitHubPath(value string) string {
	normalized := strings.Trim(path.Clean(strings.ReplaceAll(value, "\\", "/")), "/")
	if normalized == "." {
		return ""
	}
	parts := strings.Split(normalized, "/")
	for index, part := range parts {
		parts[index] = neturl.PathEscape(part)
	}
	return strings.Join(parts, "/")
}

func readBoundedGitHubResponse(reader io.Reader, maxBytes int64) ([]byte, error) {
	if maxBytes <= 0 {
		return nil, domain.GitHubRemoteContextError{
			Code:    domain.GitHubRemoteContextErrorOversizedResponse,
			Message: "GitHub response size limit is invalid",
		}
	}
	limited := io.LimitReader(reader, maxBytes+1)
	contents, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}
	if int64(len(contents)) > maxBytes {
		return nil, domain.GitHubRemoteContextError{
			Code:    domain.GitHubRemoteContextErrorOversizedResponse,
			Message: "GitHub response exceeds configured size limit",
		}
	}
	return contents, nil
}

func mapGitHubTransportError(err error) error {
	var urlErr *neturl.Error
	if errors.As(err, &urlErr) && urlErr.Timeout() {
		return domain.GitHubRemoteContextError{
			Code:    domain.GitHubRemoteContextErrorTimeout,
			Message: "GitHub request timed out",
		}
	}
	return domain.GitHubRemoteContextError{
		Code:    domain.GitHubRemoteContextErrorNetwork,
		Message: "GitHub network request failed",
	}
}

type repositoryResponse struct {
	DefaultBranch string `json:"default_branch"`
}

type commitResponse struct {
	SHA string `json:"sha"`
}

type contentsResponse struct {
	Path     string `json:"path"`
	Type     string `json:"type"`
	Size     int64  `json:"size"`
	Encoding string `json:"encoding"`
	Content  string `json:"content"`
}
