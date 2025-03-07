package model

import "time"

type Memory struct {
	ID        uint64 `gorm:"primarykey"`
	MemoryID  string `gorm:"index"`
	Content   string `gorm:"text"`
	CreatedAt time.Time
}
