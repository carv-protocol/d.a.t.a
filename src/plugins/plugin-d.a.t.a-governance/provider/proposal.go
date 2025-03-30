package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/carv-protocol/d.a.t.a/src/internal/governance"
	"github.com/carv-protocol/d.a.t.a/src/internal/plugins"
	"github.com/carv-protocol/d.a.t.a/src/pkg/logger"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	// Maximum number of hot proposals to return
	MaxHotProposals = 5
	// Maximum number of recent proposals to return
	MaxRecentProposals = 5
	// Time threshold for ending soon proposals (in hours)
	EndingSoonThreshold = 24
)

// ProposalStats contains proposal statistics
type ProposalStats struct {
	TotalProposals      int            `json:"total_proposals"`
	ActiveProposals     int            `json:"active_proposals"`
	StatusCount         map[string]int `json:"status_count"`
	HotProposals        []ProposalInfo `json:"hot_proposals"`
	RecentProposals     []ProposalInfo `json:"recent_proposals"`
	EndingSoonProposals []ProposalInfo `json:"ending_soon_proposals"`
}

// ProposalInfo contains summary information for a single proposal
type ProposalInfo struct {
	ID         uuid.UUID `json:"id"`
	Title      string    `json:"title"`
	Status     string    `json:"status"`
	VoteCount  int       `json:"vote_count"`
	YesVotes   int       `json:"yes_votes"`
	NoVotes    int       `json:"no_votes"`
	CreateTime time.Time `json:"create_time"`
	EndTime    time.Time `json:"end_time"`
}

// ProposalInformationProvider implements the Provider interface
type ProposalInformationProvider struct {
	registry governance.Registry
	logger   *zap.SugaredLogger
}

// NewProposalInformationProvider creates a new proposal information provider
func NewProposalInformationProvider(registry governance.Registry) *ProposalInformationProvider {
	return &ProposalInformationProvider{
		registry: registry,
		logger:   logger.GetLogger().With(zap.String("provider", "proposal_information")),
	}
}

// Name returns the provider name
func (p *ProposalInformationProvider) Name() string {
	return "proposal_information"
}

// Type returns the provider type
func (p *ProposalInformationProvider) Type() string {
	return "proposal_information"
}

// GetProviderState implements the Provider interface
func (p *ProposalInformationProvider) GetProviderState(ctx context.Context) (*plugins.ProviderState, error) {
	stats, err := p.collectStats(ctx)
	if err != nil {
		return nil, err
	}

	stateJSON, err := json.Marshal(stats)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal stats: %w", err)
	}

	return &plugins.ProviderState{
		Name:  p.Name(),
		Type:  p.Type(),
		State: string(stateJSON),
	}, nil
}

// collectStats collects proposal statistics
func (p *ProposalInformationProvider) collectStats(ctx context.Context) (*ProposalStats, error) {
	// Get all proposals
	proposals, err := p.registry.ListProposals(ctx, "")
	if err != nil {
		return nil, err
	}

	stats := &ProposalStats{
		TotalProposals: len(proposals),
		StatusCount:    make(map[string]int),
	}

	var activeProposals []*governance.Proposal
	var allProposalInfos []ProposalInfo

	// Iterate through proposals to collect statistics
	for _, prop := range proposals {
		// Update status count
		stats.StatusCount[string(prop.Status)]++

		// Count active proposals
		if prop.Status == "active" {
			activeProposals = append(activeProposals, prop)
			stats.ActiveProposals++
		}

		// Get proposal votes
		votes, err := p.registry.GetVotes(ctx, prop.ID)
		if err != nil {
			p.logger.Warnw("Failed to get votes for proposal",
				"proposal_id", prop.ID,
				"error", err,
			)
			continue
		}

		// Count votes
		var yesVotes, noVotes int
		for _, vote := range votes {
			if vote.Option == "yes" {
				yesVotes++
			} else {
				noVotes++
			}
		}

		info := ProposalInfo{
			ID:         prop.ID,
			Title:      prop.Title,
			Status:     string(prop.Status),
			VoteCount:  len(votes),
			YesVotes:   yesVotes,
			NoVotes:    noVotes,
			CreateTime: prop.StartTime,
			EndTime:    prop.EndTime,
		}
		allProposalInfos = append(allProposalInfos, info)
	}

	// Get hot proposals (sorted by vote count)
	sort.Slice(allProposalInfos, func(i, j int) bool {
		return allProposalInfos[i].VoteCount > allProposalInfos[j].VoteCount
	})
	if len(allProposalInfos) > MaxHotProposals {
		stats.HotProposals = allProposalInfos[:MaxHotProposals]
	} else {
		stats.HotProposals = allProposalInfos
	}

	// Get recent proposals (sorted by creation time)
	sort.Slice(allProposalInfos, func(i, j int) bool {
		return allProposalInfos[i].CreateTime.After(allProposalInfos[j].CreateTime)
	})
	if len(allProposalInfos) > MaxRecentProposals {
		stats.RecentProposals = allProposalInfos[:MaxRecentProposals]
	} else {
		stats.RecentProposals = allProposalInfos
	}

	// Get proposals ending soon
	now := time.Now()
	threshold := now.Add(EndingSoonThreshold * time.Hour)
	for _, info := range allProposalInfos {
		if info.Status == "active" && info.EndTime.Before(threshold) {
			stats.EndingSoonProposals = append(stats.EndingSoonProposals, info)
		}
	}

	// Sort by remaining time
	sort.Slice(stats.EndingSoonProposals, func(i, j int) bool {
		return stats.EndingSoonProposals[i].EndTime.Before(stats.EndingSoonProposals[j].EndTime)
	})

	return stats, nil
}
