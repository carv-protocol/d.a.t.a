package handlers

import (
	"context"

	"github.com/carv-protocol/d.a.t.a/src/internal/types"
)

// Command defines the basic information of a command
type Command interface {
	Name() string
	Description() string
	Usage() string
	Examples() []string
	Execute(ctx context.Context, msg *types.SocialMessage) error
}

// SubcommandProvider defines a command that has subcommands
type SubcommandProvider interface {
	GetSubcommands() map[string]string
}

// Registry defines the interface for command registry
type Registry interface {
	Register(cmd Command) error
	Get(name string) (Command, bool)
	ListCommands() []Command
}
