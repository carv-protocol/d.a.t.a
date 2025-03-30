package handlers

import (
	"context"

	"github.com/carv-protocol/d.a.t.a/src/internal/core"
)

type SettingsCommand struct{}

func (s *SettingsCommand) Name() string {
	return "settings"
}

func (s *SettingsCommand) Description() string {
	return "🚧 Under construction 🚧 - Manage user settings"
}

func (s *SettingsCommand) Execute(ctx context.Context, msg *core.SocialMessage) error {
	msg.Content = "Settings preferences feature is not implemented yet."
	return nil
}

func (s *SettingsCommand) Usage() string {
	return "/settings"
}

func (s *SettingsCommand) Examples() []string {
	return []string{"/settings"}
}
