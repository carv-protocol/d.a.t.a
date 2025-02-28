package handlers

import (
	"context"

	"github.com/carv-protocol/d.a.t.a/src/internal/types"
)

type ProfileCommand struct{}

func (p *ProfileCommand) Name() string {
	return "profile"
}

func (p *ProfileCommand) Description() string {
	return "🚧 Under construction 🚧 - View and edit user profile"
}

func (p *ProfileCommand) Execute(ctx context.Context, msg *types.SocialMessage) error {
	msg.Content = "Profile feature is not implemented yet."
	return nil
}

func (p *ProfileCommand) Usage() string {
	return "/profile"
}

func (p *ProfileCommand) Examples() []string {
	return []string{"/profile"}
}
