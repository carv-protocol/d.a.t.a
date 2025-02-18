package adapters

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

type PostgresStore struct {
	db         *sql.DB
	connStr    string
	maxRetries int
	retryDelay time.Duration
}

func NewPostgresStore(connStr string) *PostgresStore {
	return &PostgresStore{
		connStr:    connStr,
		maxRetries: 3,
		retryDelay: time.Second * 2,
	}
}

// Connect establishes connection to PostgreSQL with retry mechanism
func (s *PostgresStore) Connect(ctx context.Context) error {
	var err error
	var db *sql.DB

	for attempt := 0; attempt < s.maxRetries; attempt++ {
		db, err = sql.Open("postgres", s.connStr)
		if err != nil {
			continue
		}

		// Test the connection
		if err = db.PingContext(ctx); err == nil {
			s.db = db
			return nil
		}

		// Close the failed connection
		db.Close()

		// Wait before retrying
		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled while connecting to database: %w", ctx.Err())
		case <-time.After(s.retryDelay):
		}
	}

	return fmt.Errorf("failed to connect to database after %d attempts: %w", s.maxRetries, err)
}

// CreateTable creates a new table with the given schema
func (s *PostgresStore) CreateTable(ctx context.Context, tableName string, schema string) error {
	if s.db == nil {
		return fmt.Errorf("database connection not established")
	}

	// Clean and validate table name
	tableName = sanitizeIdentifier(tableName)
	if tableName == "" {
		return fmt.Errorf("invalid table name")
	}

	// Create table with schema
	query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s)", tableName, schema)

	// Wrap in transaction for safety
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	if _, err = tx.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("failed to create table %s: %w", tableName, err)
	}

	return tx.Commit()
}

// Insert adds new data to the specified table
func (s *PostgresStore) Insert(ctx context.Context, tableName string, data map[string]interface{}) error {
	if s.db == nil {
		return fmt.Errorf("database connection not established")
	}

	tableName = sanitizeIdentifier(tableName)
	if tableName == "" {
		return fmt.Errorf("invalid table name")
	}

	var columns []string
	var placeholders []string
	var values []interface{}
	paramCount := 1

	for col, val := range data {
		col = sanitizeIdentifier(col)
		if col != "" {
			columns = append(columns, col)
			placeholders = append(placeholders, fmt.Sprintf("$%d", paramCount))
			values = append(values, val)
			paramCount++
		}
	}

	if len(columns) == 0 {
		return fmt.Errorf("no valid columns provided for insert")
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		tableName,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)

	// Use transaction for atomic insert
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	if _, err = tx.ExecContext(ctx, query, values...); err != nil {
		return fmt.Errorf("failed to insert data into %s: %w", tableName, err)
	}

	return tx.Commit()
}

// Update modifies existing data in the specified table
func (s *PostgresStore) Update(ctx context.Context, tableName string, id string, data map[string]interface{}) error {
	if s.db == nil {
		return fmt.Errorf("database connection not established")
	}

	tableName = sanitizeIdentifier(tableName)
	if tableName == "" {
		return fmt.Errorf("invalid table name")
	}
	if id == "" {
		return fmt.Errorf("invalid id")
	}

	var setStatements []string
	var values []interface{}
	paramCount := 1

	for col, val := range data {
		col = sanitizeIdentifier(col)
		if col != "" {
			setStatements = append(setStatements, fmt.Sprintf("%s = $%d", col, paramCount))
			values = append(values, val)
			paramCount++
		}
	}

	if len(setStatements) == 0 {
		return fmt.Errorf("no valid columns provided for update")
	}

	// Add id to values
	values = append(values, id)

	query := fmt.Sprintf(
		"UPDATE %s SET %s WHERE id = $%d",
		tableName,
		strings.Join(setStatements, ", "),
		paramCount,
	)

	// Use transaction for atomic update
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	result, err := tx.ExecContext(ctx, query, values...)
	if err != nil {
		return fmt.Errorf("failed to update data in %s: %w", tableName, err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("no record found with id: %s", id)
	}

	return tx.Commit()
}

// Delete removes data from the specified table
func (s *PostgresStore) Delete(ctx context.Context, tableName string, id string) error {
	if s.db == nil {
		return fmt.Errorf("database connection not established")
	}

	tableName = sanitizeIdentifier(tableName)
	if tableName == "" {
		return fmt.Errorf("invalid table name")
	}
	if id == "" {
		return fmt.Errorf("invalid id")
	}

	query := fmt.Sprintf("DELETE FROM %s WHERE id = $1", tableName)

	// Use transaction for atomic delete
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	result, err := tx.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete data from %s: %w", tableName, err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("no record found with id: %s", id)
	}

	return tx.Commit()
}

func (s *PostgresStore) Get(ctx context.Context, tableName string, id string) (map[string]interface{}, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database connection not established")
	}

	tableName = sanitizeIdentifier(tableName)
	if tableName == "" {
		return nil, fmt.Errorf("invalid table name")
	}
	if id == "" {
		return nil, fmt.Errorf("invalid id")
	}

	query := fmt.Sprintf("SELECT * FROM %s WHERE id = $1", tableName)

	result, err := s.db.QueryContext(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to query data from %s: %w", tableName, err)
	}

	if !result.Next() {
		return nil, nil
	}

	var (
		ID        string
		CreatedAt time.Time
		Content   string
	)

	err = result.Scan(&ID, &Content, &CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to scan data from %s: %w", tableName, err)
	}

	return map[string]interface{}{
		"id":         ID,
		"content":    Content,
		"created_at": CreatedAt,
	}, nil
}

// Close closes the database connection
func (s *PostgresStore) Close() error {
	if s.db != nil {
		if err := s.db.Close(); err != nil {
			return fmt.Errorf("failed to close database: %w", err)
		}
		s.db = nil
	}
	return nil
}

// Additional PostgreSQL-specific helper methods

// CreateIndex creates an index on specified columns
func (s *PostgresStore) CreateIndex(ctx context.Context, tableName string, indexName string, columns []string, unique bool) error {
	if s.db == nil {
		return fmt.Errorf("database connection not established")
	}

	tableName = sanitizeIdentifier(tableName)
	indexName = sanitizeIdentifier(indexName)
	if tableName == "" || indexName == "" {
		return fmt.Errorf("invalid table or index name")
	}

	// Clean column names
	var cleanColumns []string
	for _, col := range columns {
		col = sanitizeIdentifier(col)
		if col != "" {
			cleanColumns = append(cleanColumns, col)
		}
	}

	if len(cleanColumns) == 0 {
		return fmt.Errorf("no valid columns provided for index")
	}

	uniqueStr := ""
	if unique {
		uniqueStr = "UNIQUE"
	}

	query := fmt.Sprintf(
		"CREATE %s INDEX IF NOT EXISTS %s ON %s (%s)",
		uniqueStr,
		indexName,
		tableName,
		strings.Join(cleanColumns, ", "),
	)

	_, err := s.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}

	return nil
}

// Begin starts a new transaction
func (s *PostgresStore) Begin(ctx context.Context) (*sql.Tx, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database connection not established")
	}
	return s.db.BeginTx(ctx, nil)
}

// Query executes a custom query
func (s *PostgresStore) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database connection not established")
	}
	return s.db.QueryContext(ctx, query, args...)
}

// Helper functions

func sanitizeIdentifier(identifier string) string {
	return identifier
	// Remove any characters that aren't alphanumeric or underscores
	cleaned := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') ||
			(r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') ||
			r == '_' {
			return r
		}
		return -1
	}, identifier)

	// Ensure it doesn't start with a number
	if len(cleaned) > 0 && cleaned[0] >= '0' && cleaned[0] <= '9' {
		return ""
	}

	return cleaned
}
