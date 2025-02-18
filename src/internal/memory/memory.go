package memory

import (
	"context"
	"time"

	"github.com/carv-protocol/d.a.t.a/src/pkg/database"
)

type Memory struct {
	MemoryID  string
	CreatedAt time.Time
	Content   []byte
}

type Manager interface {
	CreateMemory(ctx context.Context, memory Memory) error
	GetMemory(ctx context.Context, memoryID string) (*Memory, error)
	SetMemory(ctx context.Context, mem *Memory) error
	ListMemories(ctx context.Context, filter MemoryFilter) ([]*Memory, error)
	DeleteMemory(ctx context.Context, memoryID string) error
	SearchByEmbedding(ctx context.Context, embedding []float64, opts SearchOptions) ([]*Memory, error)
}

type managerImpl struct {
	store database.Store
}

func NewManager(store database.Store) *managerImpl {
	store.CreateTable(context.Background(), "memory", "id TEXT PRIMARY KEY, created_at TIMESTAMP, content BLOB")
	return &managerImpl{
		store: store,
	}
}

func (m *managerImpl) CreateMemory(ctx context.Context, memory Memory) error {
	return m.store.Insert(ctx, "memory", map[string]interface{}{
		"id":         memory.MemoryID,
		"created_at": memory.CreatedAt,
		"content":    memory.Content,
	})
}

func (m *managerImpl) GetMemory(ctx context.Context, memoryID string) (*Memory, error) {
	mem, err := m.store.Get(ctx, "memory", memoryID)
	if err != nil {
		return nil, err
	}

	if mem == nil {
		return nil, nil
	}

	// TODO: handle type assertion errors
	return &Memory{
		MemoryID:  mem["id"].(string),
		CreatedAt: mem["created_at"].(time.Time),
		Content:   mem["content"].([]byte),
	}, nil
}

func (m *managerImpl) SetMemory(ctx context.Context, mem *Memory) error {
	return m.store.Update(ctx, "memory", mem.MemoryID, map[string]interface{}{
		"id":         mem.MemoryID,
		"created_at": mem.CreatedAt,
		"content":    mem.Content,
	})
}

// GetState is a generic function to retrieve any type of state
func (m *managerImpl) ListMemories(ctx context.Context, filter MemoryFilter) ([]*Memory, error) {
	return nil, nil
}

// GetState is a generic function to retrieve any type of state
func (m *managerImpl) DeleteMemory(ctx context.Context, memoryID string) error {
	return nil
}

// SaveState is a generic function to save any type of state
func (m *managerImpl) SearchByEmbedding(ctx context.Context, embedding []float64, opts SearchOptions) ([]*Memory, error) {
	return nil, nil
}
