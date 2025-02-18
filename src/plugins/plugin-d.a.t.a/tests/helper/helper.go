package helper

import (
	"testing"

	"github.com/carv-protocol/d.a.t.a/src/pkg/llm"
	"github.com/carv-protocol/d.a.t.a/src/plugins/plugin-d.a.t.a/providers"
	"github.com/carv-protocol/d.a.t.a/src/plugins/plugin-d.a.t.a/types"
	"go.uber.org/zap"
)

const defaultModel = "gpt-4"

// SetupTestProvider creates a test database provider
func SetupTestProvider(t *testing.T) types.DatabaseProvider {
	return SetupTestProviderWithLLM(t, nil)
}

// SetupTestProviderWithLLM creates a test database provider with a specific LLM client
func SetupTestProviderWithLLM(t *testing.T, llmClient llm.Client) types.DatabaseProvider {
	return providers.NewDatabaseProvider(
		"test_provider",
		"http://test.api",
		"test-token",
		"ethereum",
		"test-schema",
		"test-examples",
		llmClient,
		"test-model",
		zap.NewNop().Sugar(),
	)
}

// StrPtr returns a pointer to the given string
func StrPtr(s string) *string {
	return &s
}

// IntPtr returns a pointer to the given integer
func IntPtr(i int) *int {
	return &i
}
