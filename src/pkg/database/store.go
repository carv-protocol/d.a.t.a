package database

import (
	"context"

	"gorm.io/gorm"
)

type Store interface {
	Connect(ctx context.Context) error
	DB() *gorm.DB
	MemoryTable() *gorm.DB
	CharacterTable() *gorm.DB
	Close() error
}
