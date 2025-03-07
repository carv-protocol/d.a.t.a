package adapters

import (
	"context"
	"fmt"
	"time"

	"github.com/carv-protocol/d.a.t.a/src/pkg/database"
	"github.com/carv-protocol/d.a.t.a/src/pkg/logger"

	"github.com/gozelus/gormotel"
	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	MaxOpenConnections = 20
	MaxIdleConnections = 10
	ConnMaxLifetime    = time.Minute * 6
	ConnMaxIdleTime    = time.Minute * 3
)

type PostgresStore struct {
	db       *gorm.DB
	connPath string
}

func NewPostgresStore(connPath string) *PostgresStore {
	return &PostgresStore{
		connPath: connPath,
	}
}

func (s *PostgresStore) Connect(ctx context.Context) error {
	logger.GetLogger().Infof("Connecting to Postgres database..., %s", s.connPath)

	db, err := gorm.Open(postgres.Open(s.connPath), &gorm.Config{
		Logger:         database.NewTracer(logger.GetLogger()),
		PrepareStmt:    false,
		TranslateError: true,
	})
	if err != nil {
		return err
	}

	if err = db.Use(gormotel.Plugin); err != nil {
		return err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	sqlDB.SetMaxOpenConns(MaxOpenConnections)
	sqlDB.SetMaxIdleConns(MaxIdleConnections)
	sqlDB.SetConnMaxLifetime(ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(ConnMaxIdleTime)

	if err = sqlDB.Ping(); err != nil {
		return err
	}

	s.db = db
	return nil
}

func (s *PostgresStore) DB() *gorm.DB {
	return s.db
}

func (s *PostgresStore) MemoryTable() *gorm.DB {
	return s.db.Table("data_framework.memory")
}

func (s *PostgresStore) CharacterTable() *gorm.DB {
	return s.db.Table("data_framework.character")
}

func (s *PostgresStore) Close() error {
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
