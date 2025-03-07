package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/carv-protocol/d.a.t.a/src/characters"
	"github.com/carv-protocol/d.a.t.a/src/internal/core"
	"github.com/carv-protocol/d.a.t.a/src/internal/governance"
	"github.com/carv-protocol/d.a.t.a/src/pkg/llm"

	"github.com/google/uuid"
)

type ProposalCommand struct {
	governanceRegistry governance.Registry
	llmClient          llm.Client
	model              string
	messageSender      MessageSender
	character          *characters.Character
}

func NewProposalCommand(registry governance.Registry, llmClient llm.Client, model string, messageSender MessageSender, character *characters.Character) *ProposalCommand {
	return &ProposalCommand{
		governanceRegistry: registry,
		llmClient:          llmClient,
		model:              model,
		messageSender:      messageSender,
		character:          character,
	}
}

func (p *ProposalCommand) Name() string {
	return "proposal"
}

func (p *ProposalCommand) Description() string {
	return "Create and manage governance proposals"
}

func (p *ProposalCommand) Execute(ctx context.Context, msg *core.SocialMessage) error {
	args := ParseCommandArgs(msg.Content)
	log.Printf("ProposalCommand received message: %s", msg.Content)
	log.Printf("Parsed arguments: %#v", args)

	if len(args) < 1 {
		msg.Content = "Invalid command format. Available subcommands: create, list, get, admin"
		return nil
	}

	log.Printf("Subcommand: %s", args[0])

	switch args[0] {
	case "create":
		return p.handleCreate(ctx, msg, args[1:])
	case "list":
		return p.handleList(ctx, msg, args[1:])
	case "get":
		return p.handleGet(ctx, msg, args[1:])
	case "admin":
		return p.handleAdmin(ctx, msg, args[1:])
	case "force-end":
		return p.handleForceEnd(ctx, msg, args[1:])
	case "character":
		return p.handleCharacter(ctx, msg, args[1:])
	default:
		msg.Content = fmt.Sprintf("Unknown subcommand: '%s'. Available subcommands: create, list, get, admin, force-end, character", args[0])
		return nil
	}
}

func (p *ProposalCommand) handleCreate(ctx context.Context, msg *core.SocialMessage, args []string) error {
	log.Printf("handleCreate received args: %#v", args)
	log.Printf("User ID: %s, Platform: %s", msg.FromUser, msg.Platform)

	if len(args) < 3 {
		msg.Content = "Invalid create command format. Usage: /proposal create <title> <description> <duration_in_hours>"
		return nil
	}

	title := args[0]
	description := args[1]

	// Add debug logging
	log.Printf("Title: %s", title)
	log.Printf("Description: %s", description)

	// Parse duration
	durationHours, err := strconv.ParseInt(args[2], 10, 64)
	if err != nil {
		msg.Content = fmt.Sprintf("Invalid duration: %v. Please provide duration in hours.", err)
		return nil
	}

	// Add debug logging
	log.Printf("Duration: %d hours", durationHours)

	// Calculate start and end times
	now := time.Now()
	startTimeUnix := now.Unix()
	endTimeUnix := now.Add(time.Duration(durationHours) * time.Hour).Unix()

	log.Printf("Calling CreateProposal with user: %s, platform: %s", msg.FromUser, msg.Platform)

	// Create the proposal
	proposal, err := p.governanceRegistry.CreateProposal(
		ctx,
		title,
		description,
		msg.FromUser,
		msg.Platform,
		startTimeUnix,
		endTimeUnix,
	)

	if err != nil {
		log.Printf("CreateProposal failed with error: %v", err)
		msg.Content = fmt.Sprintf("Failed to create proposal: %v", err)
		return nil
	}

	log.Printf("Proposal created successfully with ID: %s", proposal.ID)

	// Set proposal status to active
	err = p.governanceRegistry.UpdateProposalStatus(ctx, proposal.ID, governance.ProposalStatusActive)
	if err != nil {
		log.Printf("UpdateProposalStatus failed with error: %v", err)
		msg.Content = fmt.Sprintf("Failed to activate proposal: %v", err)
		return nil
	}

	msg.Content = fmt.Sprintf("Proposal created successfully!\nID: %s\nTitle: %s\nDuration: %d hours\nStart Time: %s\nEnd Time: %s\nStatus: %s",
		proposal.ID.String(),
		proposal.Title,
		durationHours,
		proposal.StartTime.Format(time.RFC3339),
		proposal.EndTime.Format(time.RFC3339),
		governance.ProposalStatusActive,
	)
	return nil
}

func (p *ProposalCommand) handleList(ctx context.Context, msg *core.SocialMessage, args []string) error {
	log.Printf("Starting handleList with args: %#v", args)
	log.Printf("Message metadata: %#v", msg.Metadata)

	var status governance.ProposalStatus
	if len(args) > 0 {
		// Convert status to uppercase for matching
		status = governance.ProposalStatus(strings.ToUpper(args[0]))
		log.Printf("Filtering by status: %s", status)
	}

	log.Printf("Fetching proposals from registry")
	proposals, err := p.governanceRegistry.ListProposals(ctx, status)
	if err != nil {
		log.Printf("Error listing proposals: %v", err)
		msg.Content = fmt.Sprintf("Failed to list proposals: %v", err)
		return nil
	}

	log.Printf("Found %d proposals", len(proposals))
	if len(proposals) == 0 {
		msg.Content = "No proposals found."
		return nil
	}

	// Build response message
	log.Printf("Building response message")
	var sb strings.Builder
	sb.WriteString("Proposals:\n\n")

	// Build a summary for AI analysis
	var summary strings.Builder
	summary.WriteString("Current Governance Proposals Summary:\n\n")

	for _, prop := range proposals {
		log.Printf("Processing proposal: %s", prop.ID)
		// Calculate remaining time for active proposals
		var timeInfo string
		if prop.Status == governance.ProposalStatusActive {
			if time.Now().Before(prop.EndTime) {
				remaining := time.Until(prop.EndTime).Round(time.Hour)
				timeInfo = fmt.Sprintf("(Ends in %d hours)", int(remaining.Hours()))
			} else {
				timeInfo = "(Voting ended)"
			}
		}

		// Add to main response
		sb.WriteString(fmt.Sprintf("ID: %s\nTitle: %s\nStatus: %s %s\nCreator: %s\nCreated At: %s\n\n",
			prop.ID.String(),
			prop.Title,
			prop.Status,
			timeInfo,
			prop.Creator,
			prop.CreatedAt.Format("2006-01-02 15:04:05"),
		))

		// Add to summary for AI
		summary.WriteString(fmt.Sprintf("- %s (Status: %s)\n", prop.Title, prop.Status))
		summary.WriteString(fmt.Sprintf("  Description: %s\n", prop.Description))
		if prop.Status == governance.ProposalStatusActive {
			if time.Now().Before(prop.EndTime) {
				remaining := time.Until(prop.EndTime)
				summary.WriteString(fmt.Sprintf("  Time Remaining: %.1f hours\n", remaining.Hours()))
			}
		}

		// Add voting results if available
		tally, err := p.governanceRegistry.TallyVotes(ctx, prop.ID)
		if err == nil {
			summary.WriteString(fmt.Sprintf("  Current Votes: Yes: %.2f, No: %.2f\n",
				tally[governance.VoteOptionYes],
				tally[governance.VoteOptionNo],
			))
		}
		summary.WriteString("\n")
	}

	// First set the basic list as content
	msg.Content = sb.String()

	// Start a goroutine to generate and send AI analysis
	go func() {
		// Generate AI analysis prompt
		log.Printf("Generating AI analysis")
		prompt := fmt.Sprintf(`As a governance assistant, please analyze the following proposals and provide insights:

%s

Please provide:
1. A brief overview of the current governance activity
2. Analysis of key proposals and their potential impact
3. Recommendations for stakeholders
4. Any concerning trends or notable patterns

Keep the analysis concise and focused on the most important points.`, summary.String())

		// Call LLM for analysis
		log.Printf("Calling LLM for analysis with model: %s", p.model)
		request := llm.CompletionRequest{
			Model: p.model,
			Messages: []llm.Message{
				{
					Role:    "user",
					Content: prompt,
				},
			},
		}
		response, err := p.llmClient.CreateCompletion(ctx, request)
		if err != nil {
			log.Printf("Error from LLM: %v", err)
			return
		}

		log.Printf("Got LLM response, sending analysis")
		analysisMsg := fmt.Sprintf("\nAI Analysis:\n══════════════════════\n\n%s", response)

		// Split long message into chunks
		const maxLength = 1900 // Leave some room for formatting
		chunks := splitMessage(analysisMsg, maxLength)

		// Send each chunk as a separate message
		for i, chunk := range chunks {
			if i > 0 {
				chunk = fmt.Sprintf("(Continued %d/%d)\n%s", i+1, len(chunks), chunk)
			}
			err = p.messageSender.SendMessage(ctx, core.SocialMessage{
				Platform: msg.Platform,
				Type:     "Response",
				Content:  chunk,
				Metadata: msg.Metadata,
			})
			if err != nil {
				log.Printf("Error sending analysis message chunk %d: %v", i+1, err)
			}
		}
	}()

	return nil
}

// splitMessage splits a long message into chunks that fit within Discord's message limit
func splitMessage(msg string, maxLength int) []string {
	if len(msg) <= maxLength {
		return []string{msg}
	}

	var chunks []string
	var currentChunk strings.Builder
	lines := strings.Split(msg, "\n")

	for _, line := range lines {
		// If adding this line would exceed maxLength, start a new chunk
		if currentChunk.Len()+len(line)+1 > maxLength {
			if currentChunk.Len() > 0 {
				chunks = append(chunks, currentChunk.String())
				currentChunk.Reset()
			}
			// If a single line is longer than maxLength, split it
			if len(line) > maxLength {
				words := strings.Fields(line)
				for _, word := range words {
					if currentChunk.Len()+len(word)+1 > maxLength {
						chunks = append(chunks, currentChunk.String())
						currentChunk.Reset()
					}
					if currentChunk.Len() > 0 {
						currentChunk.WriteString(" ")
					}
					currentChunk.WriteString(word)
				}
				continue
			}
		}

		if currentChunk.Len() > 0 {
			currentChunk.WriteString("\n")
		}
		currentChunk.WriteString(line)
	}

	// Add the last chunk if not empty
	if currentChunk.Len() > 0 {
		chunks = append(chunks, currentChunk.String())
	}

	return chunks
}

func (p *ProposalCommand) handleGet(ctx context.Context, msg *core.SocialMessage, args []string) error {
	if len(args) < 1 {
		msg.Content = "Invalid get command format. Usage: /proposal get <proposal_id>"
		return nil
	}

	// Parse proposal ID
	proposalID, err := uuid.Parse(args[0])
	if err != nil {
		msg.Content = fmt.Sprintf("Invalid proposal ID: %v", err)
		return nil
	}

	// Get proposal
	proposal, err := p.governanceRegistry.GetProposal(ctx, proposalID)
	if err != nil {
		msg.Content = fmt.Sprintf("Failed to get proposal: %v", err)
		return nil
	}

	// Get voting results
	tally, err := p.governanceRegistry.TallyVotes(ctx, proposalID)
	if err != nil {
		tally = make(map[governance.VoteOption]float64)
	}

	// Calculate remaining time for active proposals
	var timeInfo string
	if proposal.Status == governance.ProposalStatusActive {
		if time.Now().Before(proposal.EndTime) {
			remaining := time.Until(proposal.EndTime).Round(time.Hour)
			timeInfo = fmt.Sprintf("Ends in %d hours", int(remaining.Hours()))
		} else {
			timeInfo = "Voting ended"
		}
	}

	msg.Content = fmt.Sprintf(`Proposal Details:
ID: %s
Title: %s
Description: %s
Status: %s
Time: %s
Creator: %s
Platform: %s
Created At: %s
Start Time: %s
End Time: %s

Current Results:
Yes: %.2f
No: %.2f`,
		proposal.ID.String(),
		proposal.Title,
		proposal.Description,
		proposal.Status,
		timeInfo,
		proposal.Creator,
		proposal.Platform,
		proposal.CreatedAt.Format("2006-01-02 15:04:05"),
		proposal.StartTime.Format("2006-01-02 15:04:05"),
		proposal.EndTime.Format("2006-01-02 15:04:05"),
		tally[governance.VoteOptionYes],
		tally[governance.VoteOptionNo],
	)

	return nil
}

func (p *ProposalCommand) handleAdmin(ctx context.Context, msg *core.SocialMessage, args []string) error {
	// Check if user is admin
	if !p.governanceRegistry.IsAdmin(msg.FromUser) {
		msg.Content = "Only administrators can use admin commands."
		return nil
	}

	if len(args) < 1 {
		msg.Content = "Invalid admin command format. Usage: /proposal admin [set-admin/remove-admin/set-min-balance] [args...]"
		return nil
	}

	switch args[0] {
	case "set-admin":
		if len(args) < 2 {
			msg.Content = "Usage: /proposal admin set-admin <user_id>"
			return nil
		}
		err := p.governanceRegistry.SetAdmin(args[1])
		if err != nil {
			msg.Content = fmt.Sprintf("Failed to set admin: %v", err)
			return nil
		}
		msg.Content = fmt.Sprintf("Successfully set %s as admin", args[1])

	case "remove-admin":
		if len(args) < 2 {
			msg.Content = "Usage: /proposal admin remove-admin <user_id>"
			return nil
		}
		err := p.governanceRegistry.RemoveAdmin(args[1])
		if err != nil {
			msg.Content = fmt.Sprintf("Failed to remove admin: %v", err)
			return nil
		}
		msg.Content = fmt.Sprintf("Successfully removed admin status from %s", args[1])

	case "set-min-balance":
		if len(args) < 2 {
			msg.Content = "Usage: /proposal admin set-min-balance <amount>"
			return nil
		}
		balance, err := strconv.ParseFloat(args[1], 64)
		if err != nil {
			msg.Content = fmt.Sprintf("Invalid balance amount: %v", err)
			return nil
		}
		err = p.governanceRegistry.SetMinTokenBalance(balance)
		if err != nil {
			msg.Content = fmt.Sprintf("Failed to set minimum balance: %v", err)
			return nil
		}
		msg.Content = fmt.Sprintf("Successfully set minimum token balance to %.2f", balance)

	default:
		msg.Content = "Unknown admin command. Available commands: set-admin, remove-admin, set-min-balance"
	}

	return nil
}

func (p *ProposalCommand) handleForceEnd(ctx context.Context, msg *core.SocialMessage, args []string) error {
	// Check if user is admin
	if !p.governanceRegistry.IsAdmin(msg.FromUser) {
		msg.Content = "Only administrators can force end proposals."
		return nil
	}

	if len(args) < 2 {
		msg.Content = "Usage: /proposal force-end <proposal_id> <status>"
		return nil
	}

	// Parse proposal ID
	proposalID, err := uuid.Parse(args[0])
	if err != nil {
		msg.Content = fmt.Sprintf("Invalid proposal ID: %v", err)
		return nil
	}

	// Parse status
	status := governance.ProposalStatus(strings.ToUpper(args[1]))
	if status != governance.ProposalStatusPassed && status != governance.ProposalStatusRejected {
		msg.Content = "Invalid status. Must be either 'passed' or 'rejected'"
		return nil
	}

	// Force end the proposal
	err = p.governanceRegistry.ForceEndProposal(proposalID, status)
	if err != nil {
		msg.Content = fmt.Sprintf("Failed to force end proposal: %v", err)
		return nil
	}

	msg.Content = fmt.Sprintf("Successfully force ended proposal %s with status %s", proposalID, status)
	return nil
}

func (p *ProposalCommand) handleCharacter(ctx context.Context, msg *core.SocialMessage, args []string) error {
	// Check if user is admin
	if !p.governanceRegistry.IsAdmin(msg.FromUser) {
		msg.Content = "Only administrators can propose character modifications."
		return nil
	}

	if len(args) < 2 {
		msg.Content = `Usage: /proposal character <field> <new_value>

Available fields:
- name: Set the character's name (string)
  Example: /proposal character name "Love Oracle"

- system: Set the character's system prompt (string)
  Example: /proposal character system "A friendly AI assistant focused on relationship advice"

- bio: Set the character's biography (comma-separated list)
  Example: /proposal character bio "Expert in relationship counseling,Created by CARV team"

- lore: Set the character's background lore (comma-separated list)
  Example: /proposal character lore "Born in digital realm,Trained by relationship experts"

- style_tone: Set the character's tone of voice (comma-separated list)
  Example: /proposal character style_tone "friendly,empathetic,professional"

- style_constraints: Set the character's behavioral constraints (comma-separated list)
  Example: /proposal character style_constraints "Never judge,Always be supportive"

- topics: Set the character's expertise topics (comma-separated list)
  Example: /proposal character topics "relationships,dating,marriage"

- goals: Set the character's goals with priorities (format: description:priority,description:priority)
  Example: /proposal character goals "Help users improve relationships:1.0,Provide actionable advice:0.8"

Note: All proposals require voting to pass before changes take effect.
Use '/vote <proposal_id> yes/no' to vote on proposals.`
		return nil
	}

	field := strings.ToLower(args[0])
	value := args[1]

	// Validate field
	switch field {
	case "name", "system":
		// Simple string fields, no validation needed
	case "bio", "lore", "topics":
		// Array fields, split by comma
		value = strings.Join(strings.Split(value, ","), ",")
	case "style_tone", "style_constraints":
		// Style guide fields, split by comma
		value = strings.Join(strings.Split(value, ","), ",")
	case "goals":
		// Parse goals in format: description1:priority1,description2:priority2
		goals := make([]characters.Goal, 0)
		for _, goalStr := range strings.Split(value, ",") {
			parts := strings.Split(goalStr, ":")
			if len(parts) != 2 {
				msg.Content = "Invalid goal format. Use: description:priority"
				return nil
			}
			priority, err := strconv.ParseFloat(parts[1], 64)
			if err != nil {
				msg.Content = fmt.Sprintf("Invalid priority value: %v", err)
				return nil
			}
			goals = append(goals, characters.Goal{
				Description: parts[0],
				Priority:    priority,
			})
		}
		// Convert goals to JSON for storage
		goalsJSON, err := json.Marshal(goals)
		if err != nil {
			msg.Content = fmt.Sprintf("Failed to encode goals: %v", err)
			return nil
		}
		value = string(goalsJSON)
	default:
		msg.Content = fmt.Sprintf("Unknown field: '%s'. Available fields: name, system, bio, lore, style_tone, style_constraints, topics, goals", field)
		return nil
	}

	// Create a character modification proposal
	title := fmt.Sprintf("Character Modification: %s", field)
	description := fmt.Sprintf("Modify character field '%s' to: %s", field, value)

	// Calculate start and end times
	now := time.Now()
	startTimeUnix := now.Unix()
	endTimeUnix := now.Add(24 * time.Hour).Unix() // 24 hour voting period

	// Create the proposal
	proposal, err := p.governanceRegistry.CreateProposal(
		ctx,
		title,
		description,
		msg.FromUser,
		msg.Platform,
		startTimeUnix,
		endTimeUnix,
	)
	if err != nil {
		msg.Content = fmt.Sprintf("Failed to create proposal: %v", err)
		return nil
	}

	// Add character modification details
	proposal.Type = governance.ProposalTypeCharacter
	proposal.Modification = &governance.CharacterModification{
		Field:      field,
		Value:      value,
		ProposalID: proposal.ID,
	}

	// Update proposal with modification details
	err = p.governanceRegistry.UpdateProposalStatus(ctx, proposal.ID, governance.ProposalStatusActive)
	if err != nil {
		msg.Content = fmt.Sprintf("Failed to activate proposal: %v", err)
		return nil
	}

	msg.Content = fmt.Sprintf("Character modification proposal created successfully!\nID: %s\nField: %s\nValue: %s\nVoting Period: 24 hours\nUse '/vote <proposal_id> yes/no' to vote",
		proposal.ID.String(),
		field,
		value,
	)
	return nil
}

func (p *ProposalCommand) Usage() string {
	return "/proposal <subcommand> [arguments]"
}

func (p *ProposalCommand) Examples() []string {
	return []string{
		"/proposal create \"Increase Token Supply\" \"We should increase the token supply by 10%\" 72",
		"/proposal list",
		"/proposal list active",
		"/proposal get 123e4567-e89b-12d3-a456-426614174000",
		"/proposal admin set-admin user123",
		"/proposal admin set-min-balance 1000",
		"/proposal force-end 123e4567-e89b-12d3-a456-426614174000 passed",
		"/proposal character name \"Love Oracle\"",
		"/proposal character system \"A friendly AI assistant focused on relationship advice\"",
		"/proposal character bio \"Expert in relationship counseling,Created by CARV team,Has helped thousands of couples\"",
		"/proposal character style_tone \"friendly,empathetic,professional\"",
		"/proposal character style_constraints \"Never judge,Always be supportive,Maintain confidentiality\"",
		"/proposal character topics \"relationships,dating,marriage,communication,conflict resolution\"",
		"/proposal character goals \"Help users improve relationships:1.0,Provide actionable advice:0.8,Maintain user privacy:0.9\"",
	}
}

// GetSubcommands returns a list of available subcommands with descriptions
func (p *ProposalCommand) GetSubcommands() map[string]string {
	return map[string]string{
		"create":    "Create a new proposal with title, description and duration in hours",
		"list":      "List all proposals or filter by status (active, passed, rejected, pending)",
		"get":       "Get details of a specific proposal by ID",
		"admin":     "Admin commands for governance management",
		"force-end": "Force end a proposal with a specific status",
		"character": "Modify character settings (name, system, bio, lore, style_tone, style_constraints, topics, goals)",
	}
}
