package llm

import (
	"context"

	"github.com/carv-protocol/d.a.t.a/src/pkg/llm/deepseek"
	"github.com/carv-protocol/d.a.t.a/src/pkg/llm/openai"
)

type LLMConfig struct {
	Provider string `mapstructure:"provider"`
	APIKey   string `mapstructure:"api_key"`
	BaseURL  string `mapstructure:"base_url"`
	Model    string `mapstructure:"model"`
}

type State struct {
	Prompt string
}

type Message struct {
	Role    string
	Content string
}

type CompletionRequest struct {
	Model    string
	Messages []Message
}

type Client interface {
	CreateCompletion(ctx context.Context, request CompletionRequest) (string, error)
}

type clientImpl struct {
	provider       string
	openaiClient   *openai.Client
	deepseekClient *deepseek.Client
}

func (c *clientImpl) CreateCompletion(ctx context.Context, request CompletionRequest) (string, error) {
	switch c.provider {
	case "openai":
		return c.openaiClient.CreateCompletion(ctx, openai.CompletionRequest{
			Model:    request.Model,
			Messages: toOpenAIMessage(request.Messages),
		})
	case "deepseek":
		return c.deepseekClient.CreateCompletion(ctx, deepseek.CompletionRequest{
			Model:    request.Model,
			Messages: toDeepseekMessage(request.Messages),
		})
	}
	return "", nil
}

func NewClient(conf *LLMConfig) Client {
	switch conf.Provider {
	case "openai":
		return &clientImpl{
			provider:     conf.Provider,
			openaiClient: openai.NewClient(conf.APIKey),
		}
	case "deepseek":
		return &clientImpl{
			provider:       conf.Provider,
			deepseekClient: deepseek.NewClient(conf.APIKey, conf.BaseURL),
		}
	}
	return &clientImpl{}
}

func toOpenAIMessage(messages []Message) []openai.Message {
	var openAIMessages []openai.Message
	for _, message := range messages {
		openAIMessages = append(openAIMessages, openai.Message{
			Role:    message.Role,
			Content: message.Content,
		})
	}
	return openAIMessages
}

func toDeepseekMessage(messages []Message) []deepseek.Message {
	var deepseekMessages []deepseek.Message
	for _, message := range messages {
		deepseekMessages = append(deepseekMessages, deepseek.Message{
			Role:    message.Role,
			Content: message.Content,
		})
	}
	return deepseekMessages
}
