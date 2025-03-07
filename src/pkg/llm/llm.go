package llm

import (
	"context"
	"fmt"

	"github.com/carv-protocol/d.a.t.a/src/internal/conf"
	"github.com/carv-protocol/d.a.t.a/src/pkg/llm/deepseek"
	"github.com/carv-protocol/d.a.t.a/src/pkg/llm/openai"
)

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
	model          string
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
	default:
		return "", fmt.Errorf("unsupported provider: %s", c.provider)
	}
}

func NewClient(conf *conf.LLMConfig) Client {
	client := &clientImpl{
		provider: conf.Provider,
		model:    conf.Model,
	}

	switch conf.Provider {
	case "openai":
		client.openaiClient = openai.NewClient(conf.APIKey)
	case "deepseek":
		client.deepseekClient = deepseek.NewClient(conf.APIKey, conf.BaseURL)
	}

	return client
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
