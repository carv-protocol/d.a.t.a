package tasks

import (
	"context"

	"github.com/carv-protocol/d.a.t.a/src/internal/core"
	"github.com/carv-protocol/d.a.t.a/src/pkg/database"
)

// Implement me
type TaskStore struct {
	db database.Store
}

func NewTaskStore(db database.Store) *TaskStore {
	return &TaskStore{
		db: db,
	}
}

func (t *TaskStore) AddTask(ctx context.Context, task core.Task) error {
	return nil
}

func (t *TaskStore) GetTask(ctx context.Context, taskID string) (core.Task, error) {
	return core.Task{}, nil
}

func (t *TaskStore) UpdateTask(ctx context.Context, task core.Task) error {
	return nil
}

func (t *TaskStore) GetAllTasks(ctx context.Context) ([]*core.Task, error) {
	return nil, nil
}
