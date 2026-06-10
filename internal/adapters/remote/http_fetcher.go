package remote

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/guferreira1/spec-harbor/internal/core/domain"
)

const defaultFetchTimeout = 15 * time.Second

type HTTPTemplateFetcher struct {
	client   *http.Client
	maxBytes int64
}

func NewHTTPTemplateFetcher() *HTTPTemplateFetcher {
	return NewHTTPTemplateFetcherWithOptions(defaultFetchTimeout, domain.MaxRemoteTemplateHTTPResponseBytes)
}

func NewHTTPTemplateFetcherWithOptions(timeout time.Duration, maxBytes int64) *HTTPTemplateFetcher {
	return &HTTPTemplateFetcher{
		client: &http.Client{
			Timeout: timeout,
			CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		maxBytes: maxBytes,
	}
}

func NewHTTPTemplateFetcherWithClient(client *http.Client, maxBytes int64) *HTTPTemplateFetcher {
	if client == nil {
		return NewHTTPTemplateFetcherWithOptions(defaultFetchTimeout, maxBytes)
	}
	client.CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
		return http.ErrUseLastResponse
	}
	return &HTTPTemplateFetcher{client: client, maxBytes: maxBytes}
}

func (fetcher *HTTPTemplateFetcher) FetchRemoteTemplate(
	request domain.RemoteTemplateFetchRequest,
) (domain.RemoteTemplateFetchResult, error) {
	if fetcher == nil {
		return domain.RemoteTemplateFetchResult{}, errors.New("remote template HTTP fetcher is required")
	}
	client := fetcher.client
	if client == nil {
		client = NewHTTPTemplateFetcher().client
	}
	maxBytes := fetcher.maxBytes
	if maxBytes <= 0 {
		maxBytes = domain.MaxRemoteTemplateHTTPResponseBytes
	}

	httpRequest, err := http.NewRequest(http.MethodGet, request.URL().String(), nil)
	if err != nil {
		return domain.RemoteTemplateFetchResult{}, fmt.Errorf("remote template request is invalid: %w", err)
	}

	response, err := client.Do(httpRequest)
	if err != nil {
		if isTimeoutError(err) {
			return domain.RemoteTemplateFetchResult{}, fmt.Errorf("remote template fetch timeout: %w", err)
		}
		return domain.RemoteTemplateFetchResult{}, fmt.Errorf("remote template network error: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode >= 300 && response.StatusCode <= 399 {
		return domain.RemoteTemplateFetchResult{}, fmt.Errorf("remote template redirects are unsupported: HTTP %d", response.StatusCode)
	}
	if response.StatusCode != http.StatusOK {
		return domain.RemoteTemplateFetchResult{}, fmt.Errorf("remote template fetch failed: HTTP %d", response.StatusCode)
	}
	if response.ContentLength > maxBytes {
		return domain.RemoteTemplateFetchResult{}, fmt.Errorf("remote template response exceeds maximum size %d bytes", maxBytes)
	}

	body, err := io.ReadAll(io.LimitReader(response.Body, maxBytes+1))
	if err != nil {
		if isTimeoutError(err) {
			return domain.RemoteTemplateFetchResult{}, fmt.Errorf("remote template fetch timeout: %w", err)
		}
		return domain.RemoteTemplateFetchResult{}, fmt.Errorf("read remote template response: %w", err)
	}
	if int64(len(body)) > maxBytes {
		return domain.RemoteTemplateFetchResult{}, fmt.Errorf("remote template response exceeds maximum size %d bytes", maxBytes)
	}

	return domain.NewRemoteTemplateFetchResult(response.StatusCode, body), nil
}

func isTimeoutError(err error) bool {
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	var netError net.Error
	return errors.As(err, &netError) && netError.Timeout()
}
