package adapters

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/carv-protocol/d.a.t.a/src/pkg/database"
	"github.com/carv-protocol/d.a.t.a/src/pkg/logger"

	_ "github.com/mattn/go-sqlite3"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type SQLiteStore struct {
	db       *gorm.DB
	connPath string
}

func NewSQLiteStore(connPath string) *SQLiteStore {
	return &SQLiteStore{
		connPath: connPath,
	}
}

func (s *SQLiteStore) Connect(ctx context.Context) error {
	logger.GetLogger().Infof("Connecting to SQLite database..., %s", s.connPath)

	// Ensure the directory exists
	dir := filepath.Dir(s.connPath)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	db, err := gorm.Open(sqlite.Open(s.connPath), &gorm.Config{
		Logger:         database.NewTracer(logger.GetLogger()),
		PrepareStmt:    false,
		TranslateError: true,
	})
	if err != nil {
		return err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	if err = sqlDB.Ping(); err != nil {
		return err
	}

	s.db = db
	return nil
}

func (s *SQLiteStore) DB() *gorm.DB {
	return s.db
}

func (s *SQLiteStore) MemoryTable() *gorm.DB {
	return s.db.Table("memory")
}

func (s *SQLiteStore) CharacterTable() *gorm.DB {
	return s.db.Table("character")
}

func (s *SQLiteStore) Close() error {
	if s.db != nil {
		sqlDB, err := s.db.DB()
		if err != nil {
			return err
		}
		if err = sqlDB.Close(); err != nil {
			return fmt.Errorf("failed to close database: %w", err)
		}
		s.db = nil
	}
	return nil
}
