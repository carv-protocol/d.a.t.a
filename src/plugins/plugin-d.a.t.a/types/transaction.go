package types

import (
	"context"
)

// BlockStats represents the statistics of the blocks
type BlockStats struct {
	BlockRange struct {
		StartBlock string `json:"startBlock"`
		EndBlock   string `json:"endBlock"`
		BlockCount int    `json:"blockCount"`
	} `json:"blockRange"`
	TimeRange struct {
		StartTime       string `json:"startTime"`
		EndTime         string `json:"endTime"`
		TimeSpanSeconds int    `json:"timeSpanSeconds"`
	} `json:"timeRange"`
	UniqueBlocks                int     `json:"uniqueBlocks"`
	AverageTransactionsPerBlock float64 `json:"averageTransactionsPerBlock"`
}

// TransactionQueryResult represents the result of a transaction query
type TransactionQueryResult struct {
	Success  bool          `json:"success"`
	Data     []interface{} `json:"data"`
	Analysis string        `json:"analysis,omitempty"`
	Metadata struct {
		Total         int    `json:"total"`
		QueryTime     string `json:"queryTime"`
		QueryType     string `json:"queryType"`
		ExecutionTime int    `json:"executionTime"`
		Cached        bool   `json:"cached"`
		QueryDetails  *struct {
			Query           string   `json:"query"`
			ParamValidation []string `json:"paramValidation,omitempty"`
		} `json:"queryDetails,omitempty"`
		BlockStats *BlockStats `json:"blockStats,omitempty"`
	} `json:"metadata"`
	Error *struct {
		Code    string      `json:"code"`
		Message string      `json:"message"`
		Details interface{} `json:"details,omitempty"`
	} `json:"error,omitempty"`
}

// DatabaseProvider defines the interface for database operations
type DatabaseProvider interface {
	ExecuteQuery(ctx context.Context, sql string) (*TransactionQueryResult, error)
	ProcessQuery(ctx context.Context, params map[string]interface{}) (*TransactionQueryResult, error)
	AnalyzeQuery(ctx context.Context, result *TransactionQueryResult) (string, error)
	GenerateQuery(ctx context.Context, message string) (string, error)
}

// APIResponse represents the response from the API
type APIResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		ColumnInfos []string `json:"column_infos"`
		Rows        []struct {
			Items []interface{} `json:"items"`
		} `json:"rows"`
	} `json:"data"`
}
