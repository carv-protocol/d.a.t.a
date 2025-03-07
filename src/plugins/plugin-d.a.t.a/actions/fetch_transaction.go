package actions

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/carv-protocol/d.a.t.a/src/internal/actions"
	"github.com/carv-protocol/d.a.t.a/src/plugins/plugin-d.a.t.a/types"
)

// Ensure FetchTransactionAction implements core.FetchTransactionAction
var _ actions.IAction = (*FetchTransactionAction)(nil)

// FetchTransactionAction represents the action for fetching transactions
type FetchTransactionAction struct {
	name        string
	description string
	dbProvider  types.DatabaseProvider
	examples    []string
	similes     []string
}

// NewFetchTransactionAction creates a new fetch transaction action
func NewFetchTransactionAction(dbProvider types.DatabaseProvider) *FetchTransactionAction {
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

func (a *FetchTransactionAction) ParametersPrompt() string {
	return `
	# Parameters:
	- startDate: string
	- endDate: string
	- address: string
	- orderBy: string
	- orderDirection: string
	- limit: int
	`
}

func (a *FetchTransactionAction) Validate(params map[string]interface{}) error {
	// TODO: validate the parameters

	return nil
}

// Execute implements the Action interface
// TODO: fix this function
func (a *FetchTransactionAction) Execute(ctx context.Context, params map[string]interface{}) error {
	// Get message content from params
	message, ok := params["message"].(string)
	if !ok {
		return fmt.Errorf("message parameter is required")
	}

	// Generate query from message
	query, err := a.GenerateQuery(ctx, message)
	if err != nil {
		return fmt.Errorf("failed to generate query: %w", err)
	}

	// Execute query with parameters
	err = a.ExecuteWithParams(ctx, query, params)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}

	return nil
}

// ExecuteWithParams executes the action with specific parameters
// TODO: fix this function
func (a *FetchTransactionAction) ExecuteWithParams(ctx context.Context, query string, params map[string]interface{}) error {
	// 1. execute the query
	result, err := a.dbProvider.ExecuteQuery(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}

	// 2. analyze the result
	analysis, err := a.dbProvider.AnalyzeQuery(ctx, result)
	if err != nil {
		// if the analysis failed, still return the original result
		return nil
	}

	// 3. add the analysis result
	result.Analysis = analysis

	// 4. add query details to metadata
	result.Metadata.QueryDetails = &struct {
		Query           string   `json:"query"`
		ParamValidation []string `json:"paramValidation,omitempty"`
	}{
		Query: query,
	}

	return nil
}

func (a *FetchTransactionAction) Name() string {
	return a.name
}

func (a *FetchTransactionAction) Description() string {
	return a.description
}

func (a *FetchTransactionAction) Type() string {
	return "fetch_transactions"
}

// GetExamples returns the examples of the fetch transaction
func (a *FetchTransactionAction) GetExamples() []string {
	return a.examples
}

// GetSimiles returns the similes of the fetch transaction
func (a *FetchTransactionAction) GetSimiles() []string {
	return a.similes
}

// ValidateParams validates the parameters
func (a *FetchTransactionAction) ValidateParams(params map[string]interface{}) error {
	// 1. validate the date format
	if startDate, ok := params["startDate"].(string); ok {
		if _, err := time.Parse(time.RFC3339, startDate); err != nil {
			return fmt.Errorf("invalid start date format: %w", err)
		}
	}
	if endDate, ok := params["endDate"].(string); ok {
		if _, err := time.Parse(time.RFC3339, endDate); err != nil {
			return fmt.Errorf("invalid end date format: %w", err)
		}
	}

	// 2. validate the address format
	if address, ok := params["address"].(string); ok && len(address) != 42 {
		return fmt.Errorf("invalid ethereum address format")
	}

	// 3. validate the orderBy parameter
	if orderBy, ok := params["orderBy"].(string); ok {
		validOrderBy := map[string]bool{
			"block_timestamp": true,
			"value":           true,
			"gas_price":       true,
		}
		if !validOrderBy[orderBy] {
			return fmt.Errorf("invalid orderBy parameter")
		}
	}

	// 4. validate the orderDirection parameter
	if orderDirection, ok := params["orderDirection"].(string); ok {
		validDirection := map[string]bool{
			"ASC":  true,
			"DESC": true,
		}
		if !validDirection[orderDirection] {
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

// GenerateQuery generates a SQL query based on the message
func (a *FetchTransactionAction) GenerateQuery(ctx context.Context, message string) (string, error) {
	return a.dbProvider.GenerateQuery(ctx, message)
}

// FormatQueryResult formats the transaction query result into a readable string
func FormatQueryResult(result *types.TransactionQueryResult) string {
	if !result.Success {
		return fmt.Sprintf("Query failed: %s", result.Error.Message)
	}

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("Found %d transactions\n", result.Metadata.Total))

	if len(result.Data) > 0 {
		builder.WriteString("\nTransactions:\n")
		for _, tx := range result.Data {
			if txMap, ok := tx.(map[string]interface{}); ok {
				builder.WriteString(fmt.Sprintf("From: %v\n", txMap["from_address"]))
				builder.WriteString(fmt.Sprintf("To: %v\n", txMap["to_address"]))
				builder.WriteString(fmt.Sprintf("Value: %v ETH\n", txMap["value"]))
				builder.WriteString(fmt.Sprintf("Hash: %v\n\n", txMap["hash"]))
			}
		}
	}

	if result.Analysis != "" {
		builder.WriteString("\nAnalysis:\n")
		builder.WriteString(result.Analysis)
	}

	return builder.String()
}
