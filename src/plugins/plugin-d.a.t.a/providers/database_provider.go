package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/carv-protocol/d.a.t.a/src/pkg/llm"
	"github.com/carv-protocol/d.a.t.a/src/pkg/logger"
	"github.com/carv-protocol/d.a.t.a/src/plugins/core"
	"github.com/carv-protocol/d.a.t.a/src/plugins/plugin-d.a.t.a/types"
	"go.uber.org/zap"
)

const (
	clientTimeout       = 30 * time.Second
	defaultRetryCount   = 3
	defaultRetryDelay   = 1 * time.Second
	maxIdleConns        = 100
	maxIdleConnsPerHost = 100
	idleConnTimeout     = 90 * time.Second
	maxRetries          = 3
	requestTimeout      = 2 * time.Minute
	maxQueryLength      = 5000
)

var defaultTransport = &http.Transport{
	MaxIdleConns:        maxIdleConns,
	MaxIdleConnsPerHost: maxIdleConnsPerHost,
	IdleConnTimeout:     idleConnTimeout,
}

var defaultClient = &http.Client{
	Timeout:   clientTimeout,
	Transport: defaultTransport,
}

// QueryMetadata represents the metadata for a query
type QueryMetadata struct {
	ExecutionTime time.Duration `json:"executionTime"`
	Cached        bool          `json:"cached"`
	QueryDetails  *struct {
		Query           string   `json:"query"`
		ParamValidation []string `json:"paramValidation,omitempty"`
	} `json:"queryDetails,omitempty"`
}

// DatabaseProviderImpl implements the DatabaseProvider interface
type DatabaseProviderImpl struct {
	llmClient  llm.Client
	logger     *zap.SugaredLogger
	config     *DatabaseConfig
	name       string
	lastQuery  string
	queryCount int
	model      string
	apiURL     string
	authToken  string
	chain      string
	dbSchema   string
	sqlExample string
}

// DatabaseConfig contains configuration for database connection
type DatabaseConfig struct {
	DatabaseType string
	DatabaseName string
	Host         string
	Port         int
	Username     string
	Password     string
}

// NewDatabaseProvider creates a new database provider instance
func NewDatabaseProvider(
	name string,
	apiURL string,
	authToken string,
	chain string,
	dbSchema string,
	sqlExample string,
	llmClient llm.Client,
	model string,
	logger *zap.SugaredLogger,
) *DatabaseProviderImpl {
	return &DatabaseProviderImpl{
		name:       name,
		apiURL:     apiURL,
		authToken:  authToken,
		chain:      chain,
		dbSchema:   dbSchema,
		sqlExample: sqlExample,
		llmClient:  llmClient,
		model:      model,
		logger:     logger,
	}
}

// ProcessQuery processes the query and returns the result
func (p *DatabaseProviderImpl) ProcessQuery(ctx context.Context, params map[string]interface{}) (*types.TransactionQueryResult, error) {
	// 1. Generate SQL query based on params
	sql, err := p.GenerateQuery(ctx, fmt.Sprintf("%+v", params))
	if err != nil {
		return nil, fmt.Errorf("failed to generate SQL query: %w", err)
	}

	// 2. Execute query
	result, err := p.ExecuteQuery(ctx, sql)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	// 3. Return result
	return result, nil
}

// AnalyzeQuery analyzes the query result and returns insights
func (p *DatabaseProviderImpl) AnalyzeQuery(ctx context.Context, result *types.TransactionQueryResult) (string, error) {
	if result == nil {
		return "", fmt.Errorf("nil result provided for analysis")
	}

	// 1. Build analysis template
	template := p.buildAnalysisTemplate(result)

	// 2. Generate analysis using LLM
	analysis, err := p.generateAnalysis(ctx, template)
	if err != nil {
		return "", fmt.Errorf("failed to generate analysis: %w", err)
	}

	logger.GetLogger().With(
		zap.Any("analysis", analysis),
	).Info("@@@ Analysis generated successfully")

	// 3. Format and return analysis
	return p.formatAnalysis(analysis)
}

// GenerateQuery generates a SQL query based on the message
func (p *DatabaseProviderImpl) GenerateQuery(ctx context.Context, prompt string) (string, error) {
	// Create completion request
	request := llm.CompletionRequest{
		Model: p.model,
		Messages: []llm.Message{
			{
				Role:    "system",
				Content: "You are a SQL query generator. Generate only the SQL query without any explanation.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	var response string
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		timeoutCtx, cancel := context.WithTimeout(ctx, requestTimeout)
		defer cancel()

		response, lastErr = p.llmClient.CreateCompletion(timeoutCtx, request)
		if lastErr == nil {
			break
		}

		if i < maxRetries-1 {
			time.Sleep(time.Duration(i+1) * time.Second)
		}
	}

	if lastErr != nil {
		return "", fmt.Errorf("failed to generate query after %d retries: %w", maxRetries, lastErr)
	}

	// Extract SQL query from response
	query := p.extractSQLQuery(response)
	if query == "" {
		return "", fmt.Errorf("no valid SQL query found in response")
	}

	return query, nil
}

// extractSQLQuery extracts a valid SQL query from the response
func (p *DatabaseProviderImpl) extractSQLQuery(response string) string {
	// Clean the response
	response = strings.TrimSpace(response)

	// Split into lines and find the SQL query
	lines := strings.Split(response, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Check if line starts with SELECT and ends with semicolon
		upperLine := strings.ToUpper(line)
		if strings.HasPrefix(upperLine, "SELECT") {
			// Add semicolon if missing
			if !strings.HasSuffix(line, ";") {
				line += ";"
			}
			// Validate table names
			if strings.Contains(line, "eth.transactions") || strings.Contains(line, "eth.token_transfers") {
				return line
			}
		}
	}

	// If no valid query found, return default query
	return "SELECT * FROM eth.transactions WHERE date >= date_format(date_add('day', -7, current_date), '%Y-%m-%d') ORDER BY block_timestamp DESC LIMIT 3;"
}

// ExecuteQuery executes a SQL query and returns the result
func (p *DatabaseProviderImpl) ExecuteQuery(ctx context.Context, query string) (*types.TransactionQueryResult, error) {
	// Validate API URL and auth token
	if p.apiURL == "" {
		return nil, fmt.Errorf("API URL is not configured")
	}

	if p.authToken == "" {
		return nil, fmt.Errorf("auth token is not configured")
	}

	// Validate query
	if query == "" || len(query) > 5000 {
		return nil, fmt.Errorf("invalid SQL query length")
	}

	queryType := "transaction"
	if strings.Contains(strings.ToLower(query), "token_transfers") {
		queryType = "token"
	} else if strings.Contains(strings.ToLower(query), "count") {
		queryType = "aggregate"
	}

	// Execute query with retries
	var apiResponse *types.APIResponse
	var lastErr error
	var err error
	for retries := 0; retries < defaultRetryCount; retries++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		if retries > 0 {
			time.Sleep(defaultRetryDelay * time.Duration(retries))
		}

		apiResponse, err = p.executeAPIRequest(ctx, query)
		if err == nil {
			break
		}

		lastErr = err
		logger.GetLogger().Warn("Request failed, retrying",
			zap.Int("attempt", retries+1),
			zap.Error(err))
	}

	if apiResponse == nil {
		return nil, fmt.Errorf("failed after %d attempts, last error: %w", defaultRetryCount, lastErr)
	}

	// Check API response status
	if apiResponse.Code != 0 {
		return nil, fmt.Errorf("API error: %s", apiResponse.Msg)
	}

	// Transform data
	transformedData := p.TransformAPIResponse(apiResponse)

	// Create result
	result := &types.TransactionQueryResult{
		Success:  true,
		Data:     transformedData,
		Analysis: "",
		Metadata: struct {
			Total         int    `json:"total"`
			QueryTime     string `json:"queryTime"`
			QueryType     string `json:"queryType"`
			ExecutionTime int    `json:"executionTime"`
			Cached        bool   `json:"cached"`
			QueryDetails  *struct {
				Query           string   `json:"query"`
				ParamValidation []string `json:"paramValidation,omitempty"`
			} `json:"queryDetails,omitempty"`
			BlockStats *types.BlockStats `json:"blockStats,omitempty"`
		}{
			Total:         len(transformedData),
			QueryTime:     time.Now().Format(time.RFC3339),
			QueryType:     queryType,
			ExecutionTime: 0,
			Cached:        false,
			QueryDetails: &struct {
				Query           string   `json:"query"`
				ParamValidation []string `json:"paramValidation,omitempty"`
			}{
				Query: query,
			},
		},
	}

	return result, nil
}

// executeAPIRequest executes the API request with the given SQL query
func (p *DatabaseProviderImpl) executeAPIRequest(ctx context.Context, sql string) (*types.APIResponse, error) {
	logger.GetLogger().With(
		zap.String("sql", sql),
		zap.String("url", p.apiURL),
	).Info("Executing API request")

	// Prepare request
	url := fmt.Sprintf("%s/sql_query", p.apiURL)
	body := map[string]string{
		"sql_content": sql,
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		logger.GetLogger().With(
			zap.Error(err),
		).Error("Failed to marshal request body")
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		logger.GetLogger().With(
			zap.Error(err),
		).Error("Failed to create request")
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	if p.authToken != "" {
		req.Header.Set("Authorization", p.authToken)
	}

	// Execute request
	resp, err := defaultClient.Do(req)
	if err != nil {
		logger.GetLogger().With(
			zap.Error(err),
		).Error("Failed to execute request")
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.GetLogger().With(
			zap.Error(err),
		).Error("Failed to read response body")
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check response status
	if resp.StatusCode != http.StatusOK {
		logger.GetLogger().With(
			zap.Int("status_code", resp.StatusCode),
			zap.String("response", string(respBody)),
		).Error("API request failed")
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse response
	var apiResp types.APIResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		logger.GetLogger().With(
			zap.Error(err),
			zap.String("response", string(respBody)),
		).Error("Failed to unmarshal response")
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	logger.GetLogger().With(
		zap.Int("code", apiResp.Code),
		zap.String("message", apiResp.Msg),
		zap.Int("rows", len(apiResp.Data.Rows)),
	).Info("API request completed")

	return &apiResp, nil
}

// TransformAPIResponse transforms the API response into a standard format
func (p *DatabaseProviderImpl) TransformAPIResponse(apiResp *types.APIResponse) []interface{} {
	result := make([]interface{}, 0, len(apiResp.Data.Rows))

	for _, row := range apiResp.Data.Rows {
		rowData := make(map[string]interface{})
		for i, value := range row.Items {
			if i < len(apiResp.Data.ColumnInfos) {
				rowData[apiResp.Data.ColumnInfos[i]] = value
			}
		}
		result = append(result, rowData)
	}

	return result
}

func (p *DatabaseProviderImpl) buildAnalysisTemplate(result *types.TransactionQueryResult) string {
	return fmt.Sprintf(`
Please analyze the provided Ethereum blockchain data and generate a comprehensive analysis report:

Transaction Data:
%s

Query Metadata:
%s

Focus on:
1. Transaction Overview
2. Value Analysis  
3. Gas and Network Analysis
4. Address Activity
5. Technical Insights
6. Risk and Security
`, prettyJSON(result.Data), prettyJSON(result.Metadata))
}

func (p *DatabaseProviderImpl) generateAnalysis(ctx context.Context, template string) (string, error) {
	if p.llmClient == nil {
		return "", fmt.Errorf("LLM client not initialized")
	}

	request := llm.CompletionRequest{
		Model: p.model,
		Messages: []llm.Message{
			{
				Role:    "system",
				Content: "You are a data analyst providing insights from query results.",
			},
			{
				Role:    "user",
				Content: template,
			},
		},
	}

	response, err := p.llmClient.CreateCompletion(ctx, request)
	if err != nil {
		return "", fmt.Errorf("failed to analyze results: %w", err)
	}

	return response, nil
}

func (p *DatabaseProviderImpl) formatAnalysis(analysis string) (string, error) {
	// Clean and format the analysis text
	return strings.TrimSpace(analysis), nil
}

func prettyJSON(v interface{}) string {
	b, _ := json.MarshalIndent(v, "", "  ")
	return string(b)
}

func getDefaultDatabaseSchema() string {
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

func getDefaultQueryExamples() string {
	return `
Common Query Examples:

1. Find Most Active Addresses in Last 7 Days:
SELECT from_address, COUNT(*) as tx_count 
FROM eth.transactions 
WHERE date >= date_sub(current_date(), 7)
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
`
}

// Type returns the type of the provider
func (p *DatabaseProviderImpl) Type() string {
	return "database"
}

// GetProviderState returns the current state of the provider
func (p *DatabaseProviderImpl) GetProviderState(ctx context.Context) (*core.ProviderState, error) {
	state := &core.ProviderState{
		Name:  p.Name(),
		Type:  p.Type(),
		State: "connected", // Default state since we don't maintain persistent connections
		Metadata: map[string]interface{}{
			"api_url":     p.apiURL,
			"chain":       p.chain,
			"last_query":  p.lastQuery,
			"query_count": p.queryCount,
		},
	}

	return state, nil
}

// Name returns the name of the provider
func (p *DatabaseProviderImpl) Name() string {
	return p.name
}

func (p *DatabaseProviderImpl) AnalyzeResults(ctx context.Context, results interface{}) (string, error) {
	// Convert results to JSON for analysis
	resultsJSON, err := json.Marshal(results)
	if err != nil {
		return "", fmt.Errorf("failed to marshal results: %w", err)
	}

	// Create analysis prompt
	prompt := fmt.Sprintf(`Analyze the following query results and provide insights:
Results: %s

Please provide:
1. Summary of the data
2. Key patterns or trends
3. Notable anomalies
4. Recommendations based on the data`, string(resultsJSON))

	request := llm.CompletionRequest{
		Model: p.model,
		Messages: []llm.Message{
			{
				Role:    "system",
				Content: "You are a data analyst providing insights from query results.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	response, err := p.llmClient.CreateCompletion(ctx, request)
	if err != nil {
		return "", fmt.Errorf("failed to analyze results: %w", err)
	}

	return response, nil
}
