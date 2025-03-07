package model

import (
	"time"
)

type Character struct {
	ID               uint64 `gorm:"primarykey"`
	Name             string `gorm:"index"`
	System           string `gorm:"text"`
	Bio              string `gorm:"text"`
	Lore             string `gorm:"text"`
	Style            string `gorm:"text"`
	Topics           string `gorm:"text"`
	Goals            string `gorm:"text"`
	MessageExamples  string `gorm:"text"`
	TaskInstructions string `gorm:"text"`
	PriorityAccounts string `gorm:"text"`
	Preferences      string `gorm:"text"`
	CreatedAt        time.Time
}
