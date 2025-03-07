package memory

import (
	"context"
	"time"

	"github.com/carv-protocol/d.a.t.a/src/pkg/database"
	"github.com/carv-protocol/d.a.t.a/src/pkg/database/model"
)

type Memory struct {
	MemoryID  string
	Content   string
	CreatedAt time.Time
}

type Manager interface {
	CreateMemory(ctx context.Context, memory Memory) error
	GetMemory(ctx context.Context, memoryID string) (*Memory, error)
	SetMemory(ctx context.Context, mem *Memory) error
}

type ManagerImpl struct {
	store database.Store
}

func NewManager(store database.Store) (*ManagerImpl, error) {
	if err := store.MemoryTable().AutoMigrate(&model.Memory{}); err != nil {
		return nil, err
	}
	return &ManagerImpl{
		store: store,
	}, nil
}

func (m *ManagerImpl) CreateMemory(ctx context.Context, memory Memory) error {
	return m.store.MemoryTable().Create(&model.Memory{
		MemoryID:  memory.MemoryID,
		Content:   memory.Content,
		CreatedAt: memory.CreatedAt,
	}).Error
}

func (m *ManagerImpl) GetMemory(ctx context.Context, memoryID string) (*Memory, error) {
	var memory model.Memory
	if err := m.store.MemoryTable().Where("memory_id = ?", memoryID).Find(&memory).Error; err != nil {
		return nil, err
	}

	if memory.ID == 0 {
		return nil, nil
	}

	return &Memory{
		MemoryID:  memory.MemoryID,
		Content:   memory.Content,
		CreatedAt: memory.CreatedAt,
	}, nil
}

func (m *ManagerImpl) SetMemory(ctx context.Context, mem *Memory) error {
	return m.store.MemoryTable().Model(&model.Memory{}).Where("memory_id = ?", mem.MemoryID).Updates(map[string]interface{}{
		"created_at": mem.CreatedAt,
		"content":    mem.Content,
	}).Error
}
