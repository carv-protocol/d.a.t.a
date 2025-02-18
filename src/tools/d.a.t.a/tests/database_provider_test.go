package tests

import (
	"context"
	"testing"
	"time"

	"github.com/carv-protocol/d.a.t.a/src/pkg/llm"
	"github.com/carv-protocol/d.a.t.a/src/plugins/plugin-d.a.t.a/providers"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
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

	provider := providers.NewDatabaseProvider(
		"test_provider",
		"http://test.api",
		"test-token",
		"ethereum",
		"test-schema",
		"test-examples",
		nil,
		"test-model",
		zap.NewNop().Sugar(),
	)

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
}

func TestGenerateQuery(t *testing.T) {
	mockLLM := new(MockLLMClient)
	provider := providers.NewDatabaseProvider(
		"test_provider",
		"http://test.api",
		"test-token",
		"ethereum",
		"test-schema",
		"test-examples",
		mockLLM,
		"test-model",
		zap.NewNop().Sugar(),
	)

	ctx := context.Background()

	testCases := []struct {
		name          string
		message       string
		mockResponse  string
		expectedQuery string
		expectedError bool
		errorContains string
	}{
		{
			name:    "Valid query generation",
			message: "Show me the latest transactions",
			mockResponse: `SELECT * FROM eth.transactions 
				WHERE date >= date_format(date_add('day', -7, current_date), '%Y-%m-%d') 
				ORDER BY block_timestamp DESC LIMIT 10;`,
			expectedQuery: `SELECT * FROM eth.transactions 
				WHERE date >= date_format(date_add('day', -7, current_date), '%Y-%m-%d') 
				ORDER BY block_timestamp DESC LIMIT 10;`,
			expectedError: false,
		},
		{
			name:          "Invalid query generation",
			message:       "Show me the latest transactions",
			mockResponse:  "DROP TABLE eth.transactions",
			expectedError: true,
			errorContains: "invalid query generated",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLLM.On("CreateCompletion", ctx, mock.Anything).Return(tc.mockResponse, nil).Once()

			query, err := provider.GenerateQuery(ctx, tc.message)

			if tc.expectedError {
				require.Error(t, err)
				if tc.errorContains != "" {
					require.Contains(t, err.Error(), tc.errorContains)
				}
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedQuery, query)
			}

			mockLLM.AssertExpectations(t)
		})
	}
}
