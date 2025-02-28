package handlers

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/google/uuid"

	"github.com/carv-protocol/d.a.t.a/src/internal/governance"
	"github.com/carv-protocol/d.a.t.a/src/internal/types"
	"github.com/carv-protocol/d.a.t.a/src/internal/utils"
)

type VoteCommand struct {
	governanceRegistry governance.Registry
}

func NewVoteCommand(registry governance.Registry) *VoteCommand {
	return &VoteCommand{
		governanceRegistry: registry,
	}
}

func (v *VoteCommand) Name() string {
	return "vote"
}

func (v *VoteCommand) Description() string {
	return "Participate in voting"
}

func (v *VoteCommand) Execute(ctx context.Context, msg *types.SocialMessage) error {
	// Parse command arguments
	args := utils.ParseCommandArgs(msg.Content)
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

func (v *VoteCommand) handleTally(ctx context.Context, msg *types.SocialMessage, proposalIDStr string) error {
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

	// Get current tally
	tally, err := v.governanceRegistry.TallyVotes(ctx, proposalID)
	if err != nil {
		msg.Content = fmt.Sprintf("Failed to tally votes: %v", err)
		return nil
	}

	// Calculate total votes
	totalVotes := tally[governance.VoteOptionYes] + tally[governance.VoteOptionNo]

	// Calculate percentages
	var yesPercent, noPercent float64
	if totalVotes > 0 {
		yesPercent = (tally[governance.VoteOptionYes] / totalVotes) * 100
		noPercent = (tally[governance.VoteOptionNo] / totalVotes) * 100
	}

	msg.Content = fmt.Sprintf(`Voting Results for Proposal:
Title: %s
ID: %s
Status: %s

Total Votes: %.0f
Yes: %.0f (%.2f%%)
No: %.0f (%.2f%%)`,
		proposal.Title,
		proposal.ID.String(),
		proposal.Status,
		totalVotes,
		tally[governance.VoteOptionYes], yesPercent,
		tally[governance.VoteOptionNo], noPercent,
	)
	return nil
}

func (v *VoteCommand) handleList(ctx context.Context, msg *types.SocialMessage, proposalIDStr string) error {
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
		"/vote 123e4567-e89b-12d3-a456-426614174000 yes  # 投票支持提案",
		"/vote 123e4567-e89b-12d3-a456-426614174000 no   # 投票反对提案",
		"/vote tally 123e4567-e89b-12d3-a456-426614174000  # 统计投票结果",
		"/vote list 123e4567-e89b-12d3-a456-426614174000   # 列出所有投票",
	}
}

// GetSubcommands returns a list of available subcommands with descriptions
func (v *VoteCommand) GetSubcommands() map[string]string {
	return map[string]string{
		"<proposal_id> yes":   "对指定提案投赞成票",
		"<proposal_id> no":    "对指定提案投反对票",
		"tally <proposal_id>": "统计指定提案的投票结果",
		"list <proposal_id>":  "列出指定提案的所有投票",
	}
}
