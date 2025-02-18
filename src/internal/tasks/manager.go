package tasks

import (
	"context"

	"github.com/carv-protocol/d.a.t.a/src/internal/core"
)

type Metric struct {
	Threshold float64
	Current   float64
}

type Manager struct {
	store *TaskStore
}

type TaskSettings struct {
	AutoSuggest            bool
	SuggestThreshold       float64
	MaxConcurrent          int
	MinStakeholderApproval float64
}

func NewManager(store *TaskStore) *Manager {
	return &Manager{
		store: store,
	}
}

func (m *Manager) AddTask(ctx context.Context, task core.Task) error {
	return m.store.AddTask(ctx, task)
}

func (m *Manager) GetTasks(ctx context.Context) ([]*core.Task, error) {
	return m.store.GetAllTasks(ctx)
}
