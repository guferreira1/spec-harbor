package domain

type Provider string

const (
	OpenAICompatibleProvider Provider = "openai-compatible"
	AnthropicProvider        Provider = "anthropic"
	GeminiProvider           Provider = "gemini"
	OllamaProvider           Provider = "ollama"
	AzureOpenAIProvider      Provider = "azure-openai"
)
