package handlers

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/carv-protocol/d.a.t.a/src/internal/core"
)

type HelpCommand struct {
	registry Registry
}

// NewHelpCommand creates a new help command
func NewHelpCommand(registry Registry) *HelpCommand {
	return &HelpCommand{
		registry: registry,
	}
}

func (h *HelpCommand) Name() string {
	return "help"
}

func (h *HelpCommand) Description() string {
	return "Display available commands and their descriptions"
}

func (h *HelpCommand) Execute(ctx context.Context, msg *core.SocialMessage) error {
	// Parse arguments to check if a specific command is requested
	args := strings.Fields(msg.Content)
	if len(args) > 1 {
		// Get specific command help
		return h.showCommandHelp(ctx, msg, args[1])
	}

	// Get all commands
	commands := h.registry.ListCommands()

	// Sort commands by name for consistent display
	sort.Slice(commands, func(i, j int) bool {
		return commands[i].Name() < commands[j].Name()
	})

	// Build the response using a string builder for better performance
	var sb strings.Builder

	// Add header with some decoration
	sb.WriteString("```\n") // Discord markdown for code block
	sb.WriteString("🔥 Available Commands 🔥\n")
	sb.WriteString("══════════════════════\n\n")

	// Format each command's information
	for _, cmd := range commands {
		// Command name and description with emoji based on type
		emoji := h.getCommandEmoji(cmd.Name())
		sb.WriteString(fmt.Sprintf("%s /%s\n", emoji, cmd.Name()))
		sb.WriteString(fmt.Sprintf("   %s\n\n", cmd.Description()))

		// Usage with formatting
		sb.WriteString("   Usage:\n")
		sb.WriteString(fmt.Sprintf("   └─ %s\n\n", cmd.Usage()))

		// Examples with formatting (limit to 2 examples in the main help)
		if examples := cmd.Examples(); len(examples) > 0 {
			sb.WriteString("   Examples:\n")
			exampleCount := len(examples)
			if exampleCount > 2 {
				exampleCount = 2
			}
			for i := 0; i < exampleCount; i++ {
				sb.WriteString(fmt.Sprintf("   └─ %s\n", examples[i]))
			}
			if len(examples) > 2 {
				sb.WriteString("   └─ ... (use /help " + cmd.Name() + " for more examples)\n")
			}
			sb.WriteString("\n")
		}

		// Add separator between commands
		sb.WriteString("──────────────────────\n\n")
	}

	// Add footer
	sb.WriteString("Type /help <command> for more details about a specific command\n")
	sb.WriteString("```") // Close Discord markdown code block

	msg.Content = sb.String()
	return nil
}

// showCommandHelp displays detailed help for a specific command
func (h *HelpCommand) showCommandHelp(ctx context.Context, msg *core.SocialMessage, commandName string) error {
	// Get the command
	cmd, found := h.registry.Get(commandName)
	if !found {
		msg.Content = fmt.Sprintf("Command '/%s' not found. Type /help to see all available commands.", commandName)
		return nil
	}

	// Build detailed help for the command
	var sb strings.Builder

	// Add header with some decoration
	sb.WriteString("```\n") // Discord markdown for code block
	emoji := h.getCommandEmoji(cmd.Name())
	sb.WriteString(fmt.Sprintf("%s Command: /%s\n", emoji, cmd.Name()))
	sb.WriteString("══════════════════════\n\n")

	// Description
	sb.WriteString("Description:\n")
	sb.WriteString(fmt.Sprintf("%s\n\n", cmd.Description()))

	// Usage
	sb.WriteString("Usage:\n")
	sb.WriteString(fmt.Sprintf("%s\n\n", cmd.Usage()))

	// Check if command has subcommands
	if subCmdProvider, ok := cmd.(SubcommandProvider); ok {
		sb.WriteString("Subcommands:\n")
		for subCmd, desc := range subCmdProvider.GetSubcommands() {
			sb.WriteString(fmt.Sprintf("└─ %s: %s\n", subCmd, desc))
		}
		sb.WriteString("\n")
	}

	// Examples with formatting
	if examples := cmd.Examples(); len(examples) > 0 {
		sb.WriteString("Examples:\n")
		for _, example := range examples {
			sb.WriteString(fmt.Sprintf("└─ %s\n", example))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("```") // Close Discord markdown code block

	msg.Content = sb.String()
	return nil
}

// getCommandEmoji returns an appropriate emoji for different command types
func (h *HelpCommand) getCommandEmoji(cmdName string) string {
	switch cmdName {
	case "help":
		return "❓"
	case "proposal":
		return "📜"
	case "vote":
		return "🗳️"
	case "balance":
		return "💰"
	case "profile":
		return "👤"
	case "settings":
		return "⚙️"
	case "start":
		return "🚀"
	default:
		return "📌"
	}
}

func (h *HelpCommand) Usage() string {
	return "/help [command]"
}

func (h *HelpCommand) Examples() []string {
	return []string{
		"/help",
		"/help settings",
		"/help vote",
	}
}
