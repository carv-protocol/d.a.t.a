package adapters

import (
	"context"
	"fmt"
	"testing"
	"time"
)

var store *PostgresStore

func init() {
	config := PostgresConfig{
		Host:     "127.0.0.1",
		Port:     57290,
		User:     "teleportwriter",
		Password: "",
		DBName:   "dev",
		SSLMode:  "disable",
	}

	store = NewPostgresStore(config)
}

func Test_PGConnect(t *testing.T) {
	fmt.Println(store.Connect(context.TODO()))
}

func Test_PGCreate(t *testing.T) {
	ctx := context.TODO()

	err := store.Connect(ctx)
	if err != nil {
		panic(err)
	}

	err = store.CreateTable(ctx, "ai_agent.test", `id     SERIAL       PRIMARY KEY,             -- 自增主键
    content    VARCHAR(50) , 
    created_at timestamp(6)`)
	if err != nil {
		panic(err)
	}
}

func Test_PGInsert(t *testing.T) {
	ctx := context.TODO()

	err := store.Connect(ctx)
	if err != nil {
		panic(err)
	}

	err = store.Insert(ctx, "ai_agent.test", map[string]interface{}{
		"id":         1,
		"content":    "test",
		"created_at": time.Now(),
	})
	if err != nil {
		panic(err)
	}
}

func Test_PGUpdate(t *testing.T) {
	ctx := context.TODO()

	err := store.Connect(ctx)
	if err != nil {
		panic(err)
	}

	err = store.Update(ctx, "ai_agent.token_price_config", "7", map[string]interface{}{
		"token_id": "test1",
		"stop":     true,
	})
	if err != nil {
		panic(err)
	}
}

func Test_PGDelete(t *testing.T) {
	ctx := context.TODO()

	err := store.Connect(ctx)
	if err != nil {
		panic(err)
	}

	err = store.Delete(ctx, "ai_agent.token_price_config", "7")
	if err != nil {
		panic(err)
	}
}

func Test_PGGet(t *testing.T) {
	ctx := context.TODO()

	err := store.Connect(ctx)
	if err != nil {
		panic(err)
	}

	result, err := store.Get(ctx, "ai_agent.test", "1")
	if err != nil {
		panic(err)
	}

	fmt.Println(result)
}
