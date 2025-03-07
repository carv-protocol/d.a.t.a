package commands

import (
	"context"
	"errors"

	"github.com/carv-protocol/d.a.t.a/src/characters"
	"github.com/carv-protocol/d.a.t.a/src/internal/commands/handlers"
	"github.com/carv-protocol/d.a.t.a/src/internal/core"
	"github.com/carv-protocol/d.a.t.a/src/internal/governance"
	"github.com/carv-protocol/d.a.t.a/src/pkg/llm"
)

// Common errors
var (
	ErrCommandNotFound = errors.New("command not found")
	ErrInvalidCommand  = errors.New("invalid command format")
)

// BaseCommand the base command struct
type BaseCommand struct {
	name        string
	description string
	usage       string
	examples    []string
	handler     CommandHandler
}

// CommandHandler defines the function signature for command execution
type CommandHandler func(ctx context.Context, msg *core.SocialMessage) error

// NewBaseCommand creates a new base command
func NewBaseCommand(name, description, usage string, examples []string, handler CommandHandler) *BaseCommand {
	return &BaseCommand{
		name:        name,
		description: description,
		usage:       usage,
		examples:    examples,
		handler:     handler,
	}
}

func (c *BaseCommand) Name() string {
	return c.name
}

func (c *BaseCommand) Description() string {
	return c.description
}

func (c *BaseCommand) Usage() string {
	return c.usage
}

func (c *BaseCommand) Examples() []string {
	return c.examples
}

func (c *BaseCommand) Execute(ctx context.Context, msg *core.SocialMessage) error {
	return c.handler(ctx, msg)
}

// Initialize all commands
func InitializeCommands(
	registry *Registry,
	governanceRegistry governance.Registry,
	llmClient llm.Client,
	model string,
	messageSender handlers.MessageSender,
	character *characters.Character,
) {
	// register help command
	helpCmd := handlers.NewHelpCommand(registry.AsHandlerRegistry())
	registry.Register(helpCmd)

	// register settings command
	settingsCmd := &handlers.SettingsCommand{}
	registry.Register(settingsCmd)

	// register profile command
	profileCmd := &handlers.ProfileCommand{}
	registry.Register(profileCmd)

	// register balance command
	balanceCmd := &handlers.BalanceCommand{}
	registry.Register(balanceCmd)

	// register vote command
	voteCmd := handlers.NewVoteCommand(governanceRegistry, character, nil)
	registry.Register(voteCmd)

	// register proposal command
	proposalCmd := handlers.NewProposalCommand(governanceRegistry, llmClient, model, messageSender, character)
	registry.Register(proposalCmd)

	// register start command
	startCmd := &handlers.StartCommand{}
	registry.Register(startCmd)

	// register character command
	characterCmd := handlers.NewCharacterCommand(character)
	registry.Register(characterCmd)
}
