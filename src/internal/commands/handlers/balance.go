package handlers

import (
	"context"

	"github.com/carv-protocol/d.a.t.a/src/internal/core"
)

type BalanceCommand struct{}

func (b *BalanceCommand) Name() string {
	return "balance"
}

func (b *BalanceCommand) Description() string {
	return "🚧 Under construction 🚧 - View user account balance"
}

func (b *BalanceCommand) Execute(ctx context.Context, msg *core.SocialMessage) error {
	msg.Content = "Balance feature is not implemented yet."
	return nil
}

func (b *BalanceCommand) Usage() string {
	return "/balance"
}

func (b *BalanceCommand) Examples() []string {
	return []string{"/balance"}
}
