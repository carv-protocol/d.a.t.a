package actions

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

	"go.uber.org/zap"
)

const (
	clientTimeout       = 30 * time.Second
	defaultRetryCount   = 3
	defaultRetryDelay   = 1 * time.Second
	maxIdleConns        = 100
	maxIdleConnsPerHost = 100
	idleConnTimeout     = 90 * time.Second
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

// DatabaseProviderImpl implements the DatabaseProvider interface
type DatabaseProviderImpl struct {
	APIURL     string
	AuthToken  string
	Chain      string
	DBSchema   string
	SQLExample string
	LLMClient  llm.Client
}

// NewDatabaseProvider creates a new database provider
func NewDatabaseProvider(apiURL, authToken, chain string, llmClient llm.Client) DatabaseProvider {
	return &DatabaseProviderImpl{
		APIURL:     apiURL,
		AuthToken:  authToken,
		Chain:      chain,
		DBSchema:   getDefaultDatabaseSchema(),
		SQLExample: getDefaultQueryExamples(),
		LLMClient:  llmClient,
	}
}

// ProcessQuery processes the query and returns the result
func (p *DatabaseProviderImpl) ProcessQuery(ctx context.Context, params FetchTransactionParams) (*TransactionQueryResult, error) {
	// 1. Build SQL query based on params
	sql, err := p.BuildSQLQuery(params)
	if err != nil {
		return nil, fmt.Errorf("failed to build SQL query: %w", err)
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
func (p *DatabaseProviderImpl) AnalyzeQuery(ctx context.Context, result *TransactionQueryResult) (string, error) {
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

// GenerateQuery generates a SQL query based on the user's message
func (p *DatabaseProviderImpl) GenerateQuery(ctx context.Context, prompt string) (string, error) {
	// Create completion request
	request := llm.CompletionRequest{
		Model: "deepseek-chat",
		Messages: []llm.Message{
			{
				Role: "system",
				Content: `You are an expert SQL query generator for Ethereum blockchain data analysis. 
				Generate only valid SQL queries based on user requests. 
				Return ONLY the SQL query, no other text.
				The query must:
				1. Use proper table names (eth.transactions or eth.token_transfers)
				2. Be properly formatted on a single line`,
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	// Call LLM with retry
	var response string
	var lastErr error
	maxRetries := 3

	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			logger.GetLogger().With(
				zap.Int("retry", i),
				zap.Error(lastErr),
			).Info("Retrying query generation")

			// Exponential backoff
			backoff := time.Duration(i*i) * 5 * time.Second
			if backoff > 30*time.Second {
				backoff = 30 * time.Second
			}
			time.Sleep(backoff)
		}

		// Create context with timeout
		timeoutCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
		defer cancel()

		response, lastErr = p.LLMClient.CreateCompletion(timeoutCtx, request)
		if lastErr == nil {
			break
		}
	}

	if lastErr != nil {
		return "", fmt.Errorf("failed to generate query after %d retries: %w", maxRetries, lastErr)
	}

	// Extract and format SQL query
	query := p.extractSQLQuery(response)
	if query == "" {
		return "", fmt.Errorf("no valid SQL query found in response")
	}

	logger.GetLogger().With(
		zap.String("query", query),
	).Info("Successfully extracted SQL")

	// Basic SQL injection prevention
	if strings.Contains(strings.ToUpper(query), "DROP") ||
		strings.Contains(strings.ToUpper(query), "DELETE") ||
		strings.Contains(strings.ToUpper(query), "UPDATE") ||
		strings.Contains(strings.ToUpper(query), "INSERT") ||
		strings.Contains(strings.ToUpper(query), "ALTER") ||
		strings.Contains(strings.ToUpper(query), "CREATE") {
		return "", fmt.Errorf("invalid query generated: contains forbidden keywords")
	}

	logger.GetLogger().With(
		zap.String("query", query),
	).Info("Successfully generated SQL query")

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

	return ""
}

// BuildSQLQuery builds a SQL query from the given parameters
func (p *DatabaseProviderImpl) BuildSQLQuery(params FetchTransactionParams) (string, error) {
	// Validate parameters
	if params.OrderBy != nil {
		validOrderBy := map[string]bool{
			"block_timestamp": true,
			"value":           true,
			"gas_price":       true,
		}
		if !validOrderBy[*params.OrderBy] {
			return "", fmt.Errorf("invalid orderBy parameter: %s", *params.OrderBy)
		}
	}

	if params.OrderDirection != nil {
		validDirection := map[string]bool{
			"ASC":  true,
			"DESC": true,
		}
		if !validDirection[*params.OrderDirection] {
			return "", fmt.Errorf("invalid orderDirection parameter: %s", *params.OrderDirection)
		}
	}

	if params.Address != nil {
		if !strings.HasPrefix(*params.Address, "0x") || len(*params.Address) != 42 {
			return "", fmt.Errorf("invalid ethereum address format: %s", *params.Address)
		}
	}

	if params.StartDate != nil {
		if _, err := time.Parse("2006-01-02", *params.StartDate); err != nil {
			return "", fmt.Errorf("invalid start date format: %s", *params.StartDate)
		}
	}

	if params.EndDate != nil {
		if _, err := time.Parse("2006-01-02", *params.EndDate); err != nil {
			return "", fmt.Errorf("invalid end date format: %s", *params.EndDate)
		}
	}

	// Build query
	query := strings.Builder{}
	query.WriteString("SELECT * FROM eth.transactions WHERE 1=1")

	// Add time range (default to last 3 months if not specified)
	if params.StartDate != nil {
		query.WriteString(fmt.Sprintf(" AND date_parse(date, '%%Y-%%m-%%d') >= date_parse('%s', '%%Y-%%m-%%d')", *params.StartDate))
	} else {
		// Use 90 days instead of 3 months for more precise date handling
		query.WriteString(" AND date >= date_format(date_add('day', -90, current_date), '%Y-%m-%d')")
	}
	if params.EndDate != nil {
		query.WriteString(fmt.Sprintf(" AND date_parse(date, '%%Y-%%m-%%d') <= date_parse('%s', '%%Y-%%m-%%d')", *params.EndDate))
	}

	// Add address filter
	if params.Address != nil {
		query.WriteString(fmt.Sprintf(" AND (from_address = '%s' OR to_address = '%s')", *params.Address, *params.Address))
	}

	// Add value range
	if params.MinValue != nil {
		query.WriteString(fmt.Sprintf(" AND value >= %s", *params.MinValue))
	}
	if params.MaxValue != nil {
		query.WriteString(fmt.Sprintf(" AND value <= %s", *params.MaxValue))
	}

	// Add order by
	if params.OrderBy != nil {
		direction := "DESC"
		if params.OrderDirection != nil {
			direction = *params.OrderDirection
		}
		query.WriteString(fmt.Sprintf(" ORDER BY %s %s", *params.OrderBy, direction))
	}

	// Add limit
	if params.Limit != nil {
		query.WriteString(fmt.Sprintf(" LIMIT %d", *params.Limit))
	} else {
		query.WriteString(" LIMIT 100") // Default limit
	}

	return query.String(), nil
}

// ExecuteQuery executes the SQL query with retries
func (p *DatabaseProviderImpl) ExecuteQuery(ctx context.Context, sql string) (*TransactionQueryResult, error) {
	// 1. Validate query
	if sql == "" || len(sql) > 5000 {
		return nil, fmt.Errorf("invalid SQL query length")
	}

	// 2. Prepare request
	url := fmt.Sprintf("%s/sql_query", p.APIURL)
	body := map[string]string{
		"sql_content": sql,
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	logger.GetLogger().Debug("Executing query",
		zap.String("url", url),
		zap.String("sql", sql))

	// 3. Create request
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 4. Set headers
	req.Header.Set("Content-Type", "application/json")
	if p.AuthToken != "" {
		req.Header.Set("Authorization", p.AuthToken)
	}

	// 5. Execute request with retries
	var resp *http.Response
	var lastErr error

	for retries := 0; retries < defaultRetryCount; retries++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		if retries > 0 {
			time.Sleep(defaultRetryDelay * time.Duration(retries))
		}

		resp, err = defaultClient.Do(req)
		if err == nil {
			break
		}

		lastErr = err
		logger.GetLogger().Warn("Request failed, retrying",
			zap.Int("attempt", retries+1),
			zap.Error(err))
	}

	if resp == nil {
		return nil, fmt.Errorf("failed after %d attempts, last error: %w", defaultRetryCount, lastErr)
	}
	defer resp.Body.Close()

	// 6. Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// 7. Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	logger.GetLogger().Debug("Received response",
		zap.Int("status", resp.StatusCode),
		zap.Int("body_size", len(respBody)))

	// 8. Parse response
	var apiResp APIResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// 9. Check API response status
	if apiResp.Code != 0 {
		return nil, fmt.Errorf("API error: %s", apiResp.Msg)
	}

	// 10. Transform data
	transformedData := p.TransformAPIResponse(&apiResp)

	// 11. Determine query type
	queryType := "transaction"
	if strings.Contains(strings.ToLower(sql), "token_transfers") {
		queryType = "token"
	} else if strings.Contains(strings.ToLower(sql), "count") {
		queryType = "aggregate"
	}

	// 12. Build result
	result := &TransactionQueryResult{
		Success: true,
		Data:    transformedData,
		Metadata: struct {
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
		}{
			Total:         len(transformedData),
			QueryTime:     time.Now().Format(time.RFC3339),
			QueryType:     queryType,
			ExecutionTime: 0,
			Cached:        false,
			QueryDetails: &struct {
				Params          FetchTransactionParams `json:"params"`
				Query           string                 `json:"query"`
				ParamValidation []string               `json:"paramValidation,omitempty"`
			}{
				Query: sql,
			},
		},
	}

	logger.GetLogger().With(
		zap.String("query_type", queryType),
		zap.Int("total_results", len(transformedData)),
		zap.String("query", sql),
		zap.Any("result", result),
	).Info("Query executed successfully")

	return result, nil
}

// TransformAPIResponse transforms the API response into a standard format
func (p *DatabaseProviderImpl) TransformAPIResponse(apiResp *APIResponse) []interface{} {
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

func (p *DatabaseProviderImpl) buildAnalysisTemplate(result *TransactionQueryResult) string {
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
	if p.LLMClient == nil {
		return "", fmt.Errorf("LLM client not initialized")
	}

	// Create completion request
	request := llm.CompletionRequest{
		Model: "deepseek-chat", // Use the default model
		Messages: []llm.Message{
			{
				Role: "system",
				Content: `You are an expert blockchain data analyst. Analyze the provided Ethereum transaction data and generate insights.
Focus on patterns, anomalies, and potential security concerns. Use clear and professional language.`,
			},
			{
				Role:    "user",
				Content: template,
			},
		},
	}

	// Call LLM
	response, err := p.LLMClient.CreateCompletion(ctx, request)
	logger.GetLogger().With(
		zap.String("response", response),
	).Info("@@@ Response from LLM")

	if err != nil {
		return "", fmt.Errorf("failed to generate analysis: %w", err)
	}

	logger.GetLogger().With(
		zap.Int("response_length", len(response)),
	).Debug("Generated analysis response")

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
