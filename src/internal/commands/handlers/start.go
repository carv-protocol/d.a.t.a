package handlers

import (
	"context"

	"github.com/carv-protocol/d.a.t.a/src/internal/core"
)

// StartHandler handles the /start command
func StartHandler(ctx context.Context, msg *core.SocialMessage) error {
	msg.Content = "Welcome to the command system! Type /help to see available commands."
	return nil
}

type StartCommand struct{}

func (s *StartCommand) Name() string {
	return "start"
}

func (s *StartCommand) Description() string {
	return "🚧 Under construction 🚧 - Start system or feature"
}

func (s *StartCommand) Execute(ctx context.Context, msg *core.SocialMessage) error {
	msg.Content = "Start feature is not implemented yet."
	return nil
}

func (s *StartCommand) Usage() string {
	return "/start"
}

func (s *StartCommand) Examples() []string {
	return []string{"/start"}
}
