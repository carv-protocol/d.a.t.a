package helper

import (
	"os"
	"testing"

	"github.com/carv-protocol/d.a.t.a/src/pkg/llm"
	"github.com/carv-protocol/d.a.t.a/src/plugins/plugin-d.a.t.a/providers"
	"github.com/carv-protocol/d.a.t.a/src/plugins/plugin-d.a.t.a/types"
	"go.uber.org/zap"
)

// SetupTestProvider creates a test database provider
func SetupTestProvider(t *testing.T) types.DatabaseProvider {
	return SetupTestProviderWithLLM(t, nil)
}

// SetupTestProviderWithLLM creates a test database provider with a specific LLM client
func SetupTestProviderWithLLM(t *testing.T, llmClient llm.Client) types.DatabaseProvider {
	apiURL := os.Getenv("CARV_DATA_API_KEY")
	if apiURL == "" {
		t.Fatal("CARV_DATA_API_KEY environment variable is not set")
	}

	authToken := os.Getenv("CARV_DATA_AUTH_TOKEN")
	if authToken == "" {
		t.Fatal("CARV_DATA_AUTH_TOKEN environment variable is not set")
	}

	return providers.NewDatabaseProvider(
		"test_provider",
		apiURL,
		authToken,
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
