package handlers

import (
	"context"
	"fmt"
	"strings"

	"github.com/carv-protocol/d.a.t.a/src/characters"
	"github.com/carv-protocol/d.a.t.a/src/internal/core"
)

// CharacterCommand represents the command to get agent character information
type CharacterCommand struct {
	character *characters.Character
}

// NewCharacterCommand creates a new character command
func NewCharacterCommand(character *characters.Character) *CharacterCommand {
	return &CharacterCommand{
		character: character,
	}
}

func (c *CharacterCommand) Name() string {
	return "character"
}

func (c *CharacterCommand) Description() string {
	return "Get current agent character information"
}

func (c *CharacterCommand) Execute(ctx context.Context, msg *core.SocialMessage) error {
	args := strings.Fields(msg.Content)
	if len(args) > 1 {
		return c.handleSubcommand(ctx, msg, args[1])
	}

	// Build the response using a string builder
	var sb strings.Builder
	sb.WriteString("```\n") // Discord markdown for code block
	sb.WriteString("🤖 Agent Character Information\n")
	sb.WriteString("══════════════════════\n\n")

	// Basic information
	sb.WriteString(fmt.Sprintf("Name: %s\n", c.character.Name))
	sb.WriteString(fmt.Sprintf("System: %s\n\n", c.character.System))

	// Bio
	if len(c.character.Bio) > 0 {
		sb.WriteString("Biography:\n")
		for _, bio := range c.character.Bio {
			sb.WriteString(fmt.Sprintf("- %s\n", bio))
		}
		sb.WriteString("\n")
	}

	// Style
	sb.WriteString("Style Guide:\n")
	if len(c.character.Style.Tone) > 0 {
		sb.WriteString("Tone:\n")
		for _, tone := range c.character.Style.Tone {
			sb.WriteString(fmt.Sprintf("- %s\n", tone))
		}
	}
	if len(c.character.Style.Constraints) > 0 {
		sb.WriteString("\nConstraints:\n")
		for _, constraint := range c.character.Style.Constraints {
			sb.WriteString(fmt.Sprintf("- %s\n", constraint))
		}
	}
	sb.WriteString("\n")

	// Topics
	if len(c.character.Topics) > 0 {
		sb.WriteString("Topics:\n")
		for _, topic := range c.character.Topics {
			sb.WriteString(fmt.Sprintf("- %s\n", topic))
		}
		sb.WriteString("\n")
	}

	// Footer
	sb.WriteString("\nUse /character [bio|style|topics|goals|examples] for detailed information\n")
	sb.WriteString("```")

	msg.Content = sb.String()
	return nil
}

func (c *CharacterCommand) handleSubcommand(ctx context.Context, msg *core.SocialMessage, subcommand string) error {
	var sb strings.Builder
	sb.WriteString("```\n")

	switch strings.ToLower(subcommand) {
	case "bio":
		sb.WriteString("🧬 Character Biography\n")
		sb.WriteString("══════════════════════\n\n")
		for _, bio := range c.character.Bio {
			sb.WriteString(fmt.Sprintf("%s\n", bio))
		}

	case "style":
		sb.WriteString("🎨 Character Style\n")
		sb.WriteString("══════════════════════\n\n")
		sb.WriteString("Tone:\n")
		for _, tone := range c.character.Style.Tone {
			sb.WriteString(fmt.Sprintf("- %s\n", tone))
		}
		sb.WriteString("\nConstraints:\n")
		for _, constraint := range c.character.Style.Constraints {
			sb.WriteString(fmt.Sprintf("- %s\n", constraint))
		}

	case "topics":
		sb.WriteString("📚 Character Topics\n")
		sb.WriteString("══════════════════════\n\n")
		for _, topic := range c.character.Topics {
			sb.WriteString(fmt.Sprintf("- %s\n", topic))
		}

	case "goals":
		sb.WriteString("🎯 Character Goals\n")
		sb.WriteString("══════════════════════\n\n")
		for _, goal := range c.character.Goals {
			sb.WriteString(fmt.Sprintf("Goal: %s\n", goal.Name))
			sb.WriteString(fmt.Sprintf("Description: %s\n", goal.Description))
			sb.WriteString(fmt.Sprintf("Priority: %.2f\n\n", goal.Priority))
		}

	case "examples":
		sb.WriteString("💭 Message Examples\n")
		sb.WriteString("══════════════════════\n\n")
		for i, example := range c.character.MessageExamples {
			sb.WriteString(fmt.Sprintf("%d. %s\n\n", i+1, example))
		}

	default:
		msg.Content = fmt.Sprintf("Unknown subcommand: %s\nAvailable subcommands: bio, style, topics, goals, examples", subcommand)
		return nil
	}

	sb.WriteString("```")
	msg.Content = sb.String()
	return nil
}

func (c *CharacterCommand) Usage() string {
	return "/character [bio|style|topics|goals|examples]"
}

func (c *CharacterCommand) Examples() []string {
	return []string{
		"/character",
		"/character bio",
		"/character style",
		"/character topics",
		"/character goals",
		"/character examples",
	}
}

func (c *CharacterCommand) GetSubcommands() map[string]string {
	return map[string]string{
		"bio":      "Show detailed character background information",
		"style":    "Show character's language style and constraints",
		"topics":   "Show character's expertise topics",
		"goals":    "Show character's goal settings",
		"examples": "Show character's message examples",
	}
}
