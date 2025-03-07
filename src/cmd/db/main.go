package main

import (
	"context"

	"github.com/carv-protocol/d.a.t.a/src/pkg/database/adapters"
	"github.com/carv-protocol/d.a.t.a/src/pkg/logger"
)

func main() {
	logger.Init()

	ctx := context.Background()

	// Initialize database connection
	store := adapters.NewSQLiteStore("./data/agent.db") // You might want to make this configurable
	if err := store.Connect(ctx); err != nil {
		logger.GetLogger().Fatalf("Failed to connect to database: %v", err)
	}
	defer store.Close()
}
