package actions

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// FetchTransactionParams , the parameters for the fetch transaction action
type FetchTransactionParams struct {
	Address        *string `json:"address,omitempty"`
	StartDate      *string `json:"startDate,omitempty"`
	EndDate        *string `json:"endDate,omitempty"`
	MinValue       *string `json:"minValue,omitempty"`
	MaxValue       *string `json:"maxValue,omitempty"`
	Limit          *int    `json:"limit,omitempty"`
	OrderBy        *string `json:"orderBy,omitempty"`
	OrderDirection *string `json:"orderDirection,omitempty"`
}

// BlockStats , the statistics of the blocks
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

// TransactionStats , the statistics of the transactions
type TransactionStats struct {
	UniqueFromAddresses int            `json:"uniqueFromAddresses"`
	UniqueToAddresses   int            `json:"uniqueToAddresses"`
	TxTypeDistribution  map[string]int `json:"txTypeDistribution"`
	GasStats            struct {
		TotalGasUsed    int64   `json:"totalGasUsed"`
		AverageGasUsed  float64 `json:"averageGasUsed"`
		MinGasUsed      int64   `json:"minGasUsed"`
		MaxGasUsed      int64   `json:"maxGasUsed"`
		AverageGasPrice float64 `json:"averageGasPrice"`
		TotalGasCost    string  `json:"totalGasCost"`
	} `json:"gasStats"`
	ValueStats struct {
		TotalValue     string `json:"totalValue"`
		AverageValue   string `json:"averageValue"`
		MinValue       string `json:"minValue"`
		MaxValue       string `json:"maxValue"`
		ZeroValueCount int    `json:"zeroValueCount"`
	} `json:"valueStats"`
	ContractStats struct {
		ContractTransactions int `json:"contractTransactions"`
		NormalTransactions   int `json:"normalTransactions"`
		ContractInteractions struct {
			UniqueContracts int `json:"uniqueContracts"`
			TopContracts    []struct {
				Address string `json:"address"`
				Count   int    `json:"count"`
			} `json:"topContracts"`
		} `json:"contractInteractions"`
	} `json:"contractStats"`
	AddressStats struct {
		TopSenders []struct {
			Address    string `json:"address"`
			Count      int    `json:"count"`
			TotalValue string `json:"totalValue"`
		} `json:"topSenders"`
		TopReceivers []struct {
			Address    string `json:"address"`
			Count      int    `json:"count"`
			TotalValue string `json:"totalValue"`
		} `json:"topReceivers"`
	} `json:"addressStats"`
}

// TransactionQueryResult , the result of the transaction query
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
			Params          FetchTransactionParams `json:"params"`
			Query           string                 `json:"query"`
			ParamValidation []string               `json:"paramValidation,omitempty"`
		} `json:"queryDetails,omitempty"`
		BlockStats       *BlockStats       `json:"blockStats,omitempty"`
		TransactionStats *TransactionStats `json:"transactionStats,omitempty"`
	} `json:"metadata"`
	Error *struct {
		Code    string      `json:"code"`
		Message string      `json:"message"`
		Details interface{} `json:"details,omitempty"`
	} `json:"error,omitempty"`
}

// DatabaseProvider , the interface of the database provider
type DatabaseProvider interface {
	ExecuteQuery(ctx context.Context, sql string) (*TransactionQueryResult, error)
	ProcessQuery(ctx context.Context, params FetchTransactionParams) (*TransactionQueryResult, error)
	AnalyzeQuery(ctx context.Context, result *TransactionQueryResult) (string, error)
	GenerateQuery(ctx context.Context, message string) (string, error)
}

// FetchTransactionAction , the action of the fetch transaction
type FetchTransactionAction struct {
	name        string
	description string
	dbProvider  DatabaseProvider
	examples    []string
	similes     []string
}

// NewFetchTransactionAction , the new action of the fetch transaction
func NewFetchTransactionAction(dbProvider DatabaseProvider) *FetchTransactionAction {
	return &FetchTransactionAction{
		name:        "fetch_transactions",
		description: "Fetch and analyze Ethereum transactions with comprehensive statistics",
		dbProvider:  dbProvider,
		examples: []string{
			"Show me the latest 10 Ethereum transactions",
			"Get transactions for address 0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
			"Analyze gas fees trend in the last 24 hours",
			"Find whale addresses with transactions over 100 ETH",
			"Show me USDT transactions with value over 100,000 USD",
		},
		similes: []string{
			"get transactions",
			"show transfers",
			"display eth transactions",
			"find transactions",
			"search transfers",
			"check transactions",
			"view transfers",
			"list transactions",
		},
	}
}

// Execute implements the IAction interface
func (a *FetchTransactionAction) Execute() error {
	// Default implementation without parameters
	params := FetchTransactionParams{}
	_, err := a.ExecuteWithParams(context.Background(), "", params)
	return err
}

// ExecuteWithParams executes the action with specific parameters
func (a *FetchTransactionAction) ExecuteWithParams(ctx context.Context, query string, params FetchTransactionParams) (*TransactionQueryResult, error) {
	// 1. execute the query
	result, err := a.dbProvider.ExecuteQuery(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	// 2. analyze the result
	analysis, err := a.dbProvider.AnalyzeQuery(ctx, result)
	if err != nil {
		// if the analysis failed, still return the original result
		return result, nil
	}

	// 3. add the analysis result
	result.Analysis = analysis

	// 4. add query details to metadata
	result.Metadata.QueryDetails = &struct {
		Params          FetchTransactionParams `json:"params"`
		Query           string                 `json:"query"`
		ParamValidation []string               `json:"paramValidation,omitempty"`
	}{
		Params: params,
		Query:  query,
	}

	return result, nil
}

func (a *FetchTransactionAction) Name() string {
	return a.name
}

func (a *FetchTransactionAction) Description() string {
	return a.description
}

func (a *FetchTransactionAction) Type() string {
	return "FETCH_TRANSACTIONS"
}

// GetExamples , the examples of the fetch transaction
func (a *FetchTransactionAction) GetExamples() []string {
	return a.examples
}

// GetSimiles , the similes of the fetch transaction
func (a *FetchTransactionAction) GetSimiles() []string {
	return a.similes
}

// ValidateParams , the validation of the parameters
func (a *FetchTransactionAction) ValidateParams(params FetchTransactionParams) error {
	// 1. validate the date format
	if params.StartDate != nil {
		if _, err := time.Parse(time.RFC3339, *params.StartDate); err != nil {
			return fmt.Errorf("invalid start date format: %w", err)
		}
	}
	if params.EndDate != nil {
		if _, err := time.Parse(time.RFC3339, *params.EndDate); err != nil {
			return fmt.Errorf("invalid end date format: %w", err)
		}
	}

	// 2. validate the address format
	if params.Address != nil && len(*params.Address) != 42 {
		return fmt.Errorf("invalid ethereum address format")
	}

	// 3. validate the orderBy parameter
	if params.OrderBy != nil {
		validOrderBy := map[string]bool{
			"block_timestamp": true,
			"value":           true,
			"gas_price":       true,
		}
		if !validOrderBy[*params.OrderBy] {
			return fmt.Errorf("invalid orderBy parameter")
		}
	}

	// 4. validate the orderDirection parameter
	if params.OrderDirection != nil {
		validDirection := map[string]bool{
			"ASC":  true,
			"DESC": true,
		}
		if !validDirection[*params.OrderDirection] {
			return fmt.Errorf("invalid orderDirection parameter")
		}
	}

	return nil
}

// getDatabaseSchema returns the database schema for prompting
func getDatabaseSchema() string {
	return `
	CREATE EXTERNAL TABLE transactions(
		hash string,
		nonce bigint,
		block_hash string,
		block_number bigint,
		block_timestamp timestamp,
		date string,
		transaction_index bigint,
		from_address string,
		to_address string,
		value double,
		gas bigint,
		gas_price bigint,
		input string,
		max_fee_per_gas bigint,
		max_priority_fee_per_gas bigint,
		transaction_type bigint
	) PARTITIONED BY (date string);
	`
}

// lasted transaction example: SELECT * FROM eth.transactions WHERE date >= date_format(date_add('day', -90, current_date), '%Y-%m-%d') ORDER BY block_timestamp DESC LIMIT 3;

// getQueryExamples returns example queries for prompting
func getQueryExamples() string {
	return `
	Common Query Examples:

	1. Find Most Active Addresses in Last 7 Days:
	SELECT from_address, COUNT(*) as tx_count 
	FROM eth.transactions 
	WHERE date >= date_format(date_add('day', -7, current_date), '%Y-%m-%d')
	GROUP BY from_address 
	ORDER BY tx_count DESC 
	LIMIT 10;

	2. Analyze Gas Price Trends:
	SELECT date_trunc('hour', block_timestamp) as hour,
		   avg(gas_price) as avg_gas_price
	FROM eth.transactions
	WHERE date >= date_sub(current_date(), 1)
	GROUP BY 1
	ORDER BY 1;

	3. Find latest transactions:
	SELECT * FROM eth.transactions 
	WHERE date >= date_format(date_add('day', -7, current_date), '%Y-%m-%d')
	ORDER BY block_timestamp DESC 
	LIMIT 3;
	`
}

// getQueryTemplate returns the template for generating SQL queries
func getQueryTemplate() string {
	return `
	# Database Schema
	{{databaseSchema}}

	# Query Examples
	{{queryExamples}}

	# User's Query
	{{userQuery}}

	# Query Guidelines:
	1. Time Range Requirements:
	   - ALWAYS include time range limitations in queries
	   - Default to last 3 months if no specific time range is mentioned
	   - Use date_parse(date, '%Y-%m-%d') >= date_format(date_add('month', -3, current_date), '%Y-%m-%d') for default time range
	   - Adjust time range based on user's specific requirements

	2. Query Optimization:
	   - Include appropriate LIMIT clauses
	   - Use proper indexing columns (date, address, block_number)
	   - Consider partitioning by date
	   - Add WHERE clauses for efficient filtering

	3. Response Format Requirements:
	   You MUST respond with ONLY the SQL query, no other text or explanation.
	   The query should be a valid SQL statement that can be executed directly.

	4. Safety Requirements:
	   - Only SELECT statements are allowed
	   - No modifications to the database
	   - No creation of new tables or views
	   - No execution of stored procedures
	`
}

// GenerateQuery generates a SQL query based on the user's message
func (a *FetchTransactionAction) GenerateQuery(ctx context.Context, message string) (string, error) {
	// Build the prompt by replacing placeholders in template
	prompt := getQueryTemplate()
	prompt = strings.ReplaceAll(prompt, "{{databaseSchema}}", getDatabaseSchema())
	prompt = strings.ReplaceAll(prompt, "{{queryExamples}}", getQueryExamples())
	prompt = strings.ReplaceAll(prompt, "{{userQuery}}", message)

	// Call provider to generate query
	return a.dbProvider.GenerateQuery(ctx, prompt)
}

// FormatQueryResult formats the transaction query result into a readable string
func FormatQueryResult(result *TransactionQueryResult) string {
	if !result.Success {
		return fmt.Sprintf("Query failed: %s", result.Error.Message)
	}

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("Found %d transactions\n", result.Metadata.Total))

	if len(result.Data) > 0 {
		builder.WriteString("\nTransactions:\n")
		for i, tx := range result.Data {
			if txMap, ok := tx.(map[string]interface{}); ok {
				builder.WriteString(fmt.Sprintf("%d. From: %v\n   To: %v\n   Value: %v ETH\n   Hash: %v\n\n",
					i+1,
					txMap["from_address"],
					txMap["to_address"],
					txMap["value"],
					txMap["hash"],
				))
			}
		}
	}

	if result.Metadata.QueryDetails != nil && result.Metadata.QueryDetails.Query != "" {
		builder.WriteString("\nAnalysis:\n")
		builder.WriteString(result.Metadata.QueryDetails.Query)
	}

	return builder.String()
}
