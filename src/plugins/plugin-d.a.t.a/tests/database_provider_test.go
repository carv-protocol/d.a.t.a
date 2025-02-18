package tests

import (
	"context"
	"testing"
	"time"

	"github.com/carv-protocol/d.a.t.a/src/pkg/llm"
	"github.com/carv-protocol/d.a.t.a/src/plugins/d.a.t.a/providers"
	"github.com/carv-protocol/d.a.t.a/src/plugins/d.a.t.a/tests/helper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	defaultTimeout = 120 * time.Second
	maxRetries     = 2
)

// MockLLMClient is a mock implementation of llm.Client
type MockLLMClient struct {
	mock.Mock
}

func (m *MockLLMClient) CreateCompletion(ctx context.Context, request llm.CompletionRequest) (string, error) {
	args := m.Called(ctx, request)
	return args.String(0), args.Error(1)
}

func TestExecuteQuery(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	provider := helper.SetupTestProvider(t)
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	// Test simple query
	t.Run("Simple query", func(t *testing.T) {
		sql := `SELECT * FROM eth.transactions WHERE date >= date_format(date_add('day', -90, current_date), '%Y-%m-%d') ORDER BY block_timestamp DESC LIMIT 3;`

		t.Logf("Context: %+v", ctx)
		t.Logf("SQL Query: %s", sql)

		result, err := provider.ExecuteQuery(ctx, sql)
		if err != nil {
			t.Logf("Query execution error: %v", err)
			if result != nil {
				t.Logf("Partial result: %+v", result)
			}
			t.FailNow()
		}
		require.NoError(t, err)
		require.NotNil(t, result)
		require.True(t, result.Success)
		require.NotEmpty(t, result.Data)

		// Log the result details
		t.Logf("Result metadata: %+v", result.Metadata)
		t.Logf("Result data count: %d", len(result.Data))

		// Verify the returned data structure
		data := result.Data[0].(map[string]interface{})
		require.Contains(t, data, "hash")
		require.Contains(t, data, "block_number")
		require.Contains(t, data, "from_address")
		require.Contains(t, data, "to_address")
		require.Contains(t, data, "value")
	})

	// Test complex query
	t.Run("Complex query", func(t *testing.T) {
		sql := `
		WITH daily_stats AS (
			SELECT 
				date_format(block_timestamp, '%Y-%m-%d') as day,
				COUNT(*) as tx_count,
				AVG(CAST(gas_price AS DOUBLE)) as avg_gas_price,
				SUM(CAST(value AS DOUBLE)) as total_value
			FROM eth.transactions 
			WHERE date >= date_format(date_add('day', -7, current_date), '%Y-%m-%d')
			GROUP BY date_format(block_timestamp, '%Y-%m-%d')
			ORDER BY day DESC
		)
		SELECT * FROM daily_stats
		LIMIT 1`

		result, err := provider.ExecuteQuery(ctx, sql)
		if err != nil {
			t.Logf("Query execution error: %v", err)
			t.FailNow()
		}
		require.NoError(t, err)
		require.NotNil(t, result)
		require.True(t, result.Success)
		require.NotEmpty(t, result.Data)

		// Verify the returned data structure
		data := result.Data[0].(map[string]interface{})
		require.Contains(t, data, "day")
		require.Contains(t, data, "tx_count")
		require.Contains(t, data, "avg_gas_price")
		require.Contains(t, data, "total_value")
	})

	// Test error cases
	t.Run("Invalid SQL", func(t *testing.T) {
		sql := "INVALID SQL"
		result, err := provider.ExecuteQuery(ctx, sql)
		require.Error(t, err)
		require.Nil(t, result)
		require.Contains(t, err.Error(), "MALFORMED_QUERY")
	})

	t.Run("Empty SQL", func(t *testing.T) {
		sql := ""
		result, err := provider.ExecuteQuery(ctx, sql)
		require.Error(t, err)
		require.Nil(t, result)
		require.Contains(t, err.Error(), "invalid SQL query length")
	})
}

func TestAnalyzeQuery(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create a mock LLM client
	mockLLM := new(MockLLMClient)
	mockLLM.On("CreateCompletion", mock.Anything, mock.Anything).Return("Transaction Overview:\n- Found 3 transactions\n- Total value: 1.5 ETH\n\nValue Analysis:\n- Average value: 0.5 ETH\n\nGas and Network Analysis:\n- Average gas price: 50000 gwei\n\nAddress Activity:\n- 3 unique addresses\n\nTechnical Insights:\n- All transactions successful\n\nRisk and Security:\n- No suspicious patterns detected", nil)

	provider := helper.SetupTestProviderWithLLM(t, mockLLM)
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	// First execute a query to get real data
	sql := `SELECT * FROM eth.transactions WHERE date >= date_format(date_add('day', -90, current_date), '%Y-%m-%d') ORDER BY block_timestamp DESC LIMIT 3;`
	result, err := provider.ExecuteQuery(ctx, sql)
	if err != nil {
		t.Logf("Query execution error: %v", err)
		t.FailNow()
	}
	require.NoError(t, err)
	require.NotNil(t, result)
	require.True(t, result.Success)
	require.NotEmpty(t, result.Data)

	// Test analysis generation
	analysis, err := provider.AnalyzeQuery(ctx, result)
	require.NoError(t, err)
	require.NotEmpty(t, analysis)

	// Verify analysis contains key sections
	require.Contains(t, analysis, "Transaction Overview")
	require.Contains(t, analysis, "Value Analysis")
	require.Contains(t, analysis, "Gas and Network Analysis")
	require.Contains(t, analysis, "Address Activity")
	require.Contains(t, analysis, "Technical Insights")
	require.Contains(t, analysis, "Risk and Security")

	// Verify mock was called
	mockLLM.AssertExpectations(t)
}

func TestTransformAPIResponse(t *testing.T) {
	provider := helper.SetupTestProvider(t)
	providerImpl, ok := provider.(*providers.DatabaseProviderImpl)
	require.True(t, ok, "provider should be of type *DatabaseProviderImpl")

	tests := []struct {
		name     string
		response *providers.APIResponse
		want     []interface{}
	}{
		{
			name: "Empty response",
			response: &providers.APIResponse{
				Code: 0,
				Msg:  "success",
				Data: struct {
					ColumnInfos []string `json:"column_infos"`
					Rows        []struct {
						Items []interface{} `json:"items"`
					} `json:"rows"`
				}{
					ColumnInfos: []string{},
					Rows: []struct {
						Items []interface{} `json:"items"`
					}{},
				},
			},
			want: []interface{}{},
		},
		{
			name: "Single row",
			response: &providers.APIResponse{
				Code: 0,
				Msg:  "success",
				Data: struct {
					ColumnInfos []string `json:"column_infos"`
					Rows        []struct {
						Items []interface{} `json:"items"`
					} `json:"rows"`
				}{
					ColumnInfos: []string{"name", "age"},
					Rows: []struct {
						Items []interface{} `json:"items"`
					}{
						{Items: []interface{}{"Alice", 30}},
					},
				},
			},
			want: []interface{}{
				map[string]interface{}{
					"name": "Alice",
					"age":  30,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := providerImpl.TransformAPIResponse(tt.response)
			assert.Equal(t, tt.want, got)
		})
	}
}

// Helper functions for creating pointers to values
func strPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}
