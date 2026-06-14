package openai

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/guferreira1/spec-harbor/internal/core/domain"
)

const defaultOpenAIResponsesBaseURL = "https://api.openai.com/v1"

type HTTPContextRAGProvider struct {
	apiKey           string
	model            string
	baseURL          string
	client           *http.Client
	maxResponseBytes int64
}

func NewHTTPContextRAGProvider(
	apiKey string,
	model string,
	timeout time.Duration,
	maxResponseBytes int64,
) *HTTPContextRAGProvider {
	return NewHTTPContextRAGProviderWithClient(
		apiKey,
		model,
		defaultOpenAIResponsesBaseURL,
		&http.Client{Timeout: timeout},
		maxResponseBytes,
	)
}

func NewHTTPContextRAGProviderWithClient(
	apiKey string,
	model string,
	baseURL string,
	client *http.Client,
	maxResponseBytes int64,
) *HTTPContextRAGProvider {
	return &HTTPContextRAGProvider{
		apiKey:           strings.TrimSpace(apiKey),
		model:            strings.TrimSpace(model),
		baseURL:          strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		client:           client,
		maxResponseBytes: maxResponseBytes,
	}
}

func (provider *HTTPContextRAGProvider) GenerateContextAnswer(
	request domain.ContextRAGProviderRequest,
) (domain.ContextRAGProviderResponse, error) {
	if provider == nil {
		return domain.ContextRAGProviderResponse{}, domain.ContextRAGProviderError{
			Code:    domain.ContextRAGProviderErrorFailed,
			Message: "context rag provider adapter is required",
		}
	}
	if provider.client == nil {
		return domain.ContextRAGProviderResponse{}, domain.ContextRAGProviderError{
			Code:    domain.ContextRAGProviderErrorFailed,
			Message: "context rag provider HTTP client is required",
		}
	}
	if strings.TrimSpace(provider.apiKey) == "" {
		return domain.ContextRAGProviderResponse{}, domain.ContextRAGProviderError{
			Code:    domain.ContextRAGProviderErrorMissingCredentials,
			Message: fmt.Sprintf("missing OpenAI API key; set %s", domain.ContextRAGOpenAIAPIKeyEnvVar),
		}
	}
	if request.Provider != domain.ContextRAGProviderOpenAI {
		return domain.ContextRAGProviderResponse{}, domain.ContextRAGProviderError{
			Code:    domain.ContextRAGProviderErrorFailed,
			Message: fmt.Sprintf("unsupported context rag provider: %s", request.Provider),
		}
	}

	model := provider.model
	if model == "" {
		model = request.Model
	}
	payload := openAIResponsesRequest{
		Model:           model,
		Input:           openAIInputForRequest(request),
		MaxOutputTokens: maxOutputTokensForChars(request.MaxAnswerChars),
		Store:           false,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return domain.ContextRAGProviderResponse{}, err
	}

	url := provider.baseURL
	if url == "" {
		url = defaultOpenAIResponsesBaseURL
	}
	httpRequest, err := http.NewRequest(http.MethodPost, url+"/responses", bytes.NewReader(body))
	if err != nil {
		return domain.ContextRAGProviderResponse{}, err
	}
	httpRequest.Header.Set("Authorization", "Bearer "+provider.apiKey)
	httpRequest.Header.Set("Content-Type", "application/json")
	httpRequest.Header.Set("Accept", "application/json")
	httpRequest.Header.Set("User-Agent", "specharbor")

	response, err := provider.client.Do(httpRequest)
	if err != nil {
		return domain.ContextRAGProviderResponse{}, mapOpenAITransportError(err)
	}
	defer response.Body.Close()

	responseBody, err := readBoundedOpenAIResponse(response.Body, provider.maxResponseBytes)
	if err != nil {
		return domain.ContextRAGProviderResponse{}, err
	}
	if response.StatusCode >= 400 {
		return domain.ContextRAGProviderResponse{}, mapOpenAIStatus(response.StatusCode)
	}

	answer, finishReason, insufficient, err := parseOpenAIResponseText(responseBody)
	if err != nil {
		return domain.ContextRAGProviderResponse{}, err
	}
	return domain.ContextRAGProviderResponse{
		Provider:     domain.ContextRAGProviderOpenAI,
		Model:        model,
		Answer:       answer,
		Insufficient: insufficient,
		FinishReason: finishReason,
	}, nil
}

type openAIResponsesRequest struct {
	Model           string               `json:"model"`
	Input           []openAIInputMessage `json:"input"`
	MaxOutputTokens int                  `json:"max_output_tokens"`
	Store           bool                 `json:"store"`
}

type openAIInputMessage struct {
	Role    string                 `json:"role"`
	Content []openAIInputTextChunk `json:"content"`
}

type openAIInputTextChunk struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func openAIInputForRequest(request domain.ContextRAGProviderRequest) []openAIInputMessage {
	return []openAIInputMessage{
		{
			Role: "developer",
			Content: []openAIInputTextChunk{{
				Type: "input_text",
				Text: request.Instructions,
			}},
		},
		{
			Role: "user",
			Content: []openAIInputTextChunk{{
				Type: "input_text",
				Text: strings.Join([]string{
					"Question:",
					request.Query.DisplayQuery,
					"",
					"Bounded source snippets:",
					request.SourceContext,
				}, "\n"),
			}},
		},
	}
}

func maxOutputTokensForChars(maxAnswerChars int) int {
	if maxAnswerChars <= 0 {
		return 1024
	}
	tokens := maxAnswerChars / 4
	if tokens < 256 {
		return 256
	}
	return tokens
}

func readBoundedOpenAIResponse(reader io.Reader, maxBytes int64) ([]byte, error) {
	if maxBytes <= 0 {
		return nil, domain.ContextRAGProviderError{
			Code:    domain.ContextRAGProviderErrorOversizedResponse,
			Message: "provider response byte limit is invalid",
		}
	}
	limited := io.LimitReader(reader, maxBytes+1)
	contents, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}
	if int64(len(contents)) > maxBytes {
		return nil, domain.ContextRAGProviderError{
			Code:    domain.ContextRAGProviderErrorOversizedResponse,
			Message: "provider response exceeded configured byte limit",
		}
	}
	return contents, nil
}

func mapOpenAITransportError(err error) error {
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return domain.ContextRAGProviderError{
			Code:    domain.ContextRAGProviderErrorTimeout,
			Message: "OpenAI request timed out",
		}
	}
	return domain.ContextRAGProviderError{
		Code:    domain.ContextRAGProviderErrorNetwork,
		Message: "OpenAI request failed",
	}
}

func mapOpenAIStatus(statusCode int) error {
	switch statusCode {
	case http.StatusUnauthorized, http.StatusForbidden:
		return domain.ContextRAGProviderError{
			Code:    domain.ContextRAGProviderErrorUnauthorized,
			Message: "OpenAI request was not authorized",
		}
	case http.StatusTooManyRequests:
		return domain.ContextRAGProviderError{
			Code:    domain.ContextRAGProviderErrorRateLimited,
			Message: "OpenAI rate limit exceeded",
		}
	default:
		return domain.ContextRAGProviderError{
			Code:    domain.ContextRAGProviderErrorFailed,
			Message: fmt.Sprintf("OpenAI request failed with status %d", statusCode),
		}
	}
}

type openAIResponse struct {
	OutputText string         `json:"output_text"`
	Output     []openAIOutput `json:"output"`
	Incomplete *struct {
		Reason string `json:"reason"`
	} `json:"incomplete_details"`
}

type openAIOutput struct {
	Content []openAIOutputContent `json:"content"`
}

type openAIOutputContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func parseOpenAIResponseText(contents []byte) (string, string, bool, error) {
	var parsed openAIResponse
	if err := json.Unmarshal(contents, &parsed); err != nil {
		return "", "", false, domain.ContextRAGProviderError{
			Code:    domain.ContextRAGProviderErrorInvalidResponse,
			Message: "OpenAI response was not valid JSON",
		}
	}
	if parsed.Incomplete != nil {
		message := "OpenAI response was incomplete"
		if strings.TrimSpace(parsed.Incomplete.Reason) != "" {
			message = message + ": " + strings.TrimSpace(parsed.Incomplete.Reason)
		}
		return "", "", false, domain.ContextRAGProviderError{
			Code:    domain.ContextRAGProviderErrorInvalidResponse,
			Message: message,
		}
	}
	text := strings.TrimSpace(parsed.OutputText)
	if text == "" {
		var parts []string
		for _, output := range parsed.Output {
			for _, content := range output.Content {
				if content.Type == "output_text" && strings.TrimSpace(content.Text) != "" {
					parts = append(parts, strings.TrimSpace(content.Text))
				}
			}
		}
		text = strings.Join(parts, "\n")
	}
	if strings.TrimSpace(text) == "" {
		return "", "", false, domain.ContextRAGProviderError{
			Code:    domain.ContextRAGProviderErrorInvalidResponse,
			Message: "OpenAI response did not include answer text",
		}
	}
	insufficient := strings.Contains(
		strings.ToLower(text),
		"provided sources are insufficient",
	)
	return text, "", insufficient, nil
}
