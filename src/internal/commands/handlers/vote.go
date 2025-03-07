package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/carv-protocol/d.a.t.a/src/characters"
	"github.com/carv-protocol/d.a.t.a/src/internal/core"
	"github.com/carv-protocol/d.a.t.a/src/internal/governance"

	"github.com/google/uuid"
)

type VoteCommand struct {
	governanceRegistry governance.Registry
	character          *characters.Character
	agent              CharacterUpdater
}

func NewVoteCommand(registry governance.Registry, character *characters.Character, agent CharacterUpdater) *VoteCommand {
	return &VoteCommand{
		governanceRegistry: registry,
		character:          character,
		agent:              agent,
	}
}

func (v *VoteCommand) Name() string {
	return "vote"
}

func (v *VoteCommand) Description() string {
	return "Participate in voting"
}

func (v *VoteCommand) Execute(ctx context.Context, msg *core.SocialMessage) error {
	// Parse command arguments
	args := ParseCommandArgs(msg.Content)
	log.Printf("VoteCommand received message: %s", msg.Content)
	log.Printf("Parsed arguments: %#v", args)

	if len(args) < 1 {
		msg.Content = "Invalid command format. Usage: /vote <proposal_id> <yes/no> or /vote tally <proposal_id> or /vote list <proposal_id>"
		return nil
	}

	// Check if first argument is a subcommand
	if args[0] == "tally" {
		if len(args) < 2 {
			msg.Content = "Invalid tally command format. Usage: /vote tally <proposal_id>"
			return nil
		}
		return v.handleTally(ctx, msg, args[1])
	}

	if args[0] == "list" {
		if len(args) < 2 {
			msg.Content = "Invalid list command format. Usage: /vote list <proposal_id>"
			return nil
		}
		return v.handleList(ctx, msg, args[1])
	}

	// Handle regular vote command
	if len(args) < 2 { // command <proposal_id> <yes/no>
		msg.Content = "Invalid command format. Usage: /vote <proposal_id> <yes/no>"
		return nil
	}

	// Parse proposal ID
	proposalID, err := uuid.Parse(args[0])
	if err != nil {
		msg.Content = fmt.Sprintf("Invalid proposal ID: %v", err)
		return nil
	}

	// Parse vote option
	var option governance.VoteOption
	switch strings.ToLower(args[1]) {
	case "yes":
		option = governance.VoteOptionYes
	case "no":
		option = governance.VoteOptionNo
	default:
		msg.Content = "Invalid vote option. Please use 'yes' or 'no'."
		return nil
	}

	// Cast the vote
	err = v.governanceRegistry.CastVote(ctx, proposalID, msg.FromUser, msg.Platform, option)
	if err != nil {
		msg.Content = fmt.Sprintf("Failed to cast vote: %v", err)
		return nil
	}

	// Get current voting tally
	tally, err := v.governanceRegistry.TallyVotes(ctx, proposalID)
	if err != nil {
		msg.Content = "Vote cast successfully. Unable to fetch current results."
		return nil
	}

	msg.Content = fmt.Sprintf("Vote cast successfully!\nCurrent Results:\nYes: %.2f\nNo: %.2f", tally[governance.VoteOptionYes], tally[governance.VoteOptionNo])
	return nil
}

func (v *VoteCommand) handleTally(ctx context.Context, msg *core.SocialMessage, proposalIDStr string) error {
	proposalID, err := uuid.Parse(proposalIDStr)
	if err != nil {
		msg.Content = fmt.Sprintf("Invalid proposal ID: %v", err)
		return nil
	}

	// Get proposal details
	proposal, err := v.governanceRegistry.GetProposal(ctx, proposalID)
	if err != nil {
		msg.Content = fmt.Sprintf("Failed to get proposal: %v", err)
		return nil
	}

	// Get voting results
	tally, err := v.governanceRegistry.TallyVotes(ctx, proposalID)
	if err != nil {
		msg.Content = fmt.Sprintf("Failed to tally votes: %v", err)
		return nil
	}

	// Check if proposal has ended
	if proposal.Status == governance.ProposalStatusActive {
		msg.Content = fmt.Sprintf("Current Results:\nYes: %.2f\nNo: %.2f",
			tally[governance.VoteOptionYes],
			tally[governance.VoteOptionNo],
		)
		return nil
	}

	// If this is a character modification proposal and it passed, apply the changes
	if proposal.Status == governance.ProposalStatusPassed && proposal.Type == governance.ProposalTypeCharacter {
		if err := v.applyCharacterModification(ctx, proposal); err != nil {
			msg.Content = fmt.Sprintf("Failed to apply character modification: %v", err)
			return nil
		}
	}

	msg.Content = fmt.Sprintf("Final Results:\nStatus: %s\nYes: %.2f\nNo: %.2f",
		proposal.Status,
		tally[governance.VoteOptionYes],
		tally[governance.VoteOptionNo],
	)
	return nil
}

func (v *VoteCommand) applyCharacterModification(ctx context.Context, proposal *governance.Proposal) error {
	if proposal.Modification == nil {
		return fmt.Errorf("no modification details found in proposal")
	}

	// Create a copy of the current character
	newCharacter := v.character

	// Apply the modification
	switch proposal.Modification.Field {
	case "name":
		newCharacter.Name = proposal.Modification.Value.(string)
	case "system":
		newCharacter.System = proposal.Modification.Value.(string)
	case "bio", "lore", "topics":
		values := strings.Split(proposal.Modification.Value.(string), ",")
		switch proposal.Modification.Field {
		case "bio":
			newCharacter.Bio = values
		case "lore":
			newCharacter.Lore = values
		case "topics":
			newCharacter.Topics = values
		}
	case "style_tone", "style_constraints":
		values := strings.Split(proposal.Modification.Value.(string), ",")
		switch proposal.Modification.Field {
		case "style_tone":
			newCharacter.Style.Tone = values
		case "style_constraints":
			newCharacter.Style.Constraints = values
		}
	case "goals":
		var goals []characters.Goal
		if err := json.Unmarshal([]byte(proposal.Modification.Value.(string)), &goals); err != nil {
			return fmt.Errorf("failed to parse goals: %w", err)
		}
		newCharacter.Goals = goals
	default:
		return fmt.Errorf("unknown field: %s", proposal.Modification.Field)
	}

	// // Save the new character to a proposal-specific file
	// proposalPath := filepath.Join("src", "config", fmt.Sprintf("character_data_agent_%s.json", proposal.ID))
	// if err := newCharacter.SaveToPath(proposalPath); err != nil {
	// 	return fmt.Errorf("failed to save character: %w", err)
	// }

	// // Update the agent's character
	// if err := v.agent.UpdateCharacter(newCharacter); err != nil {
	// 	return fmt.Errorf("failed to update agent character: %w", err)
	// }

	return nil
}

func (v *VoteCommand) handleList(ctx context.Context, msg *core.SocialMessage, proposalIDStr string) error {
	// Parse proposal ID
	proposalID, err := uuid.Parse(proposalIDStr)
	if err != nil {
		msg.Content = fmt.Sprintf("Invalid proposal ID: %v", err)
		return nil
	}

	// Get proposal details
	proposal, err := v.governanceRegistry.GetProposal(ctx, proposalID)
	if err != nil {
		msg.Content = fmt.Sprintf("Failed to get proposal: %v", err)
		return nil
	}

	// Get votes
	votes, err := v.governanceRegistry.GetVotes(ctx, proposalID)
	if err != nil {
		msg.Content = fmt.Sprintf("Failed to get votes: %v", err)
		return nil
	}

	if len(votes) == 0 {
		msg.Content = fmt.Sprintf("No votes found for proposal: %s", proposal.Title)
		return nil
	}

	// Build response
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Votes for Proposal: %s\n\n", proposal.Title))

	for i, vote := range votes {
		sb.WriteString(fmt.Sprintf("%d. Voter: %s\n   Option: %s\n   Time: %s\n\n",
			i+1,
			vote.Voter,
			vote.Option,
			vote.CreatedAt.Format("2006-01-02 15:04:05"),
		))
	}

	msg.Content = sb.String()
	return nil
}

func (v *VoteCommand) Usage() string {
	return "/vote <proposal_id> <yes/no> or /vote tally <proposal_id> or /vote list <proposal_id>"
}

func (v *VoteCommand) Examples() []string {
	return []string{
		"/vote 123e4567-e89b-12d3-a456-426614174000 yes",
		"/vote 123e4567-e89b-12d3-a456-426614174000 no",
		"/vote 123e4567-e89b-12d3-a456-426614174000 yes",
		"/vote 123e4567-e89b-12d3-a456-426614174000 no",
		"/vote tally 123e4567-e89b-12d3-a456-426614174000",
		"/vote list 123e4567-e89b-12d3-a456-426614174000",
	}
}

// GetSubcommands returns a list of available subcommands with descriptions
func (v *VoteCommand) GetSubcommands() map[string]string {
	return map[string]string{
		"<proposal_id> yes":   "Vote yes for the proposal",
		"<proposal_id> no":    "Vote no for the proposal",
		"tally <proposal_id>": "Tally the votes for the proposal",
		"list <proposal_id>":  "List all votes for the proposal",
	}
}
