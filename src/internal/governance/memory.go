package governance

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/carv-protocol/d.a.t.a/src/internal/types"
	"github.com/google/uuid"
)

type MemoryRegistry struct {
	sync.RWMutex
	proposals    map[uuid.UUID]*Proposal
	votes        map[uuid.UUID][]Vote
	adminConfig  *AdminConfig
	tokenManager TokenManager
}

func NewMemoryRegistry(tokenManager TokenManager) Registry {
	registry := &MemoryRegistry{
		proposals: make(map[uuid.UUID]*Proposal),
		votes:     make(map[uuid.UUID][]Vote),
		adminConfig: &AdminConfig{
			Admins:          []string{},
			MinTokenBalance: 0, // Default minimum balance for voting is 0
			PlatformConfigs: make(map[string]*PlatformConfig),
		},
		tokenManager: tokenManager,
	}

	return registry
}

func (r *MemoryRegistry) CreateProposal(ctx context.Context, title, description, creator, platform string, startTime, endTime int64) (*Proposal, error) {
	r.Lock()
	defer r.Unlock()

	// Check if user is admin
	isAdmin := r.isAdminInternal(creator, platform)

	// If not admin, check token balance
	if !isAdmin {
		// Get user's token balance
		balance, err := r.GetTokenBalance(ctx, creator, platform)
		if err != nil {
			return nil, fmt.Errorf("failed to check token balance: %v", err)
		}

		// Get minimum token balance for the platform
		minBalance := r.getMinTokenBalanceInternal(platform)

		// Check if user has enough tokens
		if balance.Balance < minBalance {
			return nil, fmt.Errorf("insufficient token balance. Required: %.2f, Current: %.2f",
				minBalance, balance.Balance)
		}
	}

	proposal := &Proposal{
		ID:          uuid.New(),
		Title:       title,
		Description: description,
		Creator:     creator,
		Platform:    platform,
		Status:      ProposalStatusPending,
		StartTime:   time.Unix(startTime, 0),
		EndTime:     time.Unix(endTime, 0),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	r.proposals[proposal.ID] = proposal
	return proposal, nil
}

func (r *MemoryRegistry) GetProposal(ctx context.Context, id uuid.UUID) (*Proposal, error) {
	r.RLock()
	defer r.RUnlock()

	proposal, ok := r.proposals[id]
	if !ok {
		return nil, fmt.Errorf("proposal not found")
	}
	return proposal, nil
}

func (r *MemoryRegistry) ListProposals(ctx context.Context, status ProposalStatus) ([]*Proposal, error) {
	r.RLock()
	defer r.RUnlock()

	var result []*Proposal
	for _, proposal := range r.proposals {
		if status == "" || proposal.Status == status {
			result = append(result, proposal)
		}
	}
	return result, nil
}

func (r *MemoryRegistry) UpdateProposalStatus(ctx context.Context, id uuid.UUID, status ProposalStatus) error {
	r.Lock()
	defer r.Unlock()

	proposal, ok := r.proposals[id]
	if !ok {
		return fmt.Errorf("proposal not found")
	}

	proposal.Status = status
	proposal.UpdatedAt = time.Now()
	return nil
}

func (r *MemoryRegistry) CastVote(ctx context.Context, proposalID uuid.UUID, voter string, platform string, option VoteOption) error {
	r.Lock()
	defer r.Unlock()

	// Check if proposal exists
	proposal, ok := r.proposals[proposalID]
	if !ok {
		return fmt.Errorf("proposal not found")
	}

	// Check if proposal is active
	if proposal.Status != ProposalStatusActive {
		return fmt.Errorf("proposal is not active")
	}

	// Check if user has already voted
	for _, vote := range r.votes[proposalID] {
		if vote.Voter == voter {
			if vote.Option == option {
				return fmt.Errorf("you have already voted %s on this proposal", option)
			}
			return fmt.Errorf("you have already voted on this proposal with option %s, cannot change to %s", vote.Option, option)
		}
	}

	// Check if user is admin
	isAdmin := r.isAdminInternal(voter, platform)

	// If not admin, check token balance
	if !isAdmin {
		// Check if user has enough tokens
		balance, err := r.GetTokenBalance(ctx, voter, platform)
		if err != nil {
			return fmt.Errorf("failed to check token balance: %v", err)
		}

		// Get minimum token balance for the platform
		minBalance := r.getMinTokenBalanceInternal(platform)

		if balance.Balance < minBalance {
			return fmt.Errorf("insufficient token balance to vote. Required: %f, Current: %f", minBalance, balance.Balance)
		}
	}

	// Add vote
	vote := Vote{
		ProposalID: proposalID,
		Voter:      voter,
		Option:     option,
		CreatedAt:  time.Now(),
	}

	r.votes[proposalID] = append(r.votes[proposalID], vote)
	return nil
}

func (r *MemoryRegistry) TallyVotes(ctx context.Context, proposalID uuid.UUID) (map[VoteOption]float64, error) {
	r.RLock()
	defer r.RUnlock()

	votes := r.votes[proposalID]
	result := make(map[VoteOption]float64)

	for _, vote := range votes {
		result[vote.Option]++
	}

	return result, nil
}

func (r *MemoryRegistry) GetProposalResult(ctx context.Context, proposalID uuid.UUID) (ProposalStatus, error) {
	r.RLock()
	defer r.RUnlock()

	proposal, ok := r.proposals[proposalID]
	if !ok {
		return "", fmt.Errorf("proposal not found")
	}

	return proposal.Status, nil
}

func (r *MemoryRegistry) IsAdmin(userID string) bool {
	r.RLock()
	defer r.RUnlock()

	// Check global admins
	for _, admin := range r.adminConfig.Admins {
		if admin == userID {
			return true
		}
	}

	// Check platform-specific admins
	for _, platformConfig := range r.adminConfig.PlatformConfigs {
		for _, admin := range platformConfig.Admins {
			if admin == userID {
				return true
			}
		}
	}

	return false
}

func (r *MemoryRegistry) IsAdminForPlatform(userID string, platform string) bool {
	r.RLock()
	defer r.RUnlock()

	// Check global admins
	for _, admin := range r.adminConfig.Admins {
		if admin == userID {
			return true
		}
	}

	// Check platform-specific admins
	if platformConfig, ok := r.adminConfig.PlatformConfigs[platform]; ok {
		for _, admin := range platformConfig.Admins {
			if admin == userID {
				return true
			}
		}
	}

	return false
}

func (r *MemoryRegistry) SetAdmin(userID string) error {
	fmt.Println("setting admin")
	r.Lock()
	defer r.Unlock()

	fmt.Println("checking if user is admin")
	// Check admin status directly without calling IsAdmin to avoid deadlock
	for _, admin := range r.adminConfig.Admins {
		if admin == userID {
			return fmt.Errorf("user is already an admin")
		}
	}

	fmt.Println("adding user to admin list")
	r.adminConfig.Admins = append(r.adminConfig.Admins, userID)
	return nil
}

func (r *MemoryRegistry) RemoveAdmin(userID string) error {
	r.Lock()
	defer r.Unlock()

	for i, admin := range r.adminConfig.Admins {
		if admin == userID {
			r.adminConfig.Admins = append(r.adminConfig.Admins[:i], r.adminConfig.Admins[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("user is not an admin")
}

func (r *MemoryRegistry) GetMinTokenBalance() float64 {
	r.RLock()
	defer r.RUnlock()

	return r.adminConfig.MinTokenBalance
}

func (r *MemoryRegistry) GetMinTokenBalanceForPlatform(platform string) float64 {
	r.RLock()
	defer r.RUnlock()

	if platformConfig, ok := r.adminConfig.PlatformConfigs[platform]; ok {
		return platformConfig.MinTokenBalance
	}

	return r.adminConfig.MinTokenBalance // Fall back to global setting
}

func (r *MemoryRegistry) GetTokenBalance(ctx context.Context, userID string, platform string) (*types.TokenBalance, error) {
	return r.tokenManager.FetchNativeTokenBalance(ctx, userID, platform)
}

func (r *MemoryRegistry) UpdateAdminConfig(config *AdminConfig) error {
	r.Lock()
	defer r.Unlock()

	if config == nil {
		return fmt.Errorf("admin config cannot be nil")
	}

	// Deep copy to avoid reference issues
	r.adminConfig = &AdminConfig{
		Admins:          make([]string, len(config.Admins)),
		MinTokenBalance: config.MinTokenBalance,
		PlatformConfigs: make(map[string]*PlatformConfig),
	}

	// Copy admins
	copy(r.adminConfig.Admins, config.Admins)

	// Copy platform configs
	for platform, platformConfig := range config.PlatformConfigs {
		r.adminConfig.PlatformConfigs[platform] = &PlatformConfig{
			Admins:          make([]string, len(platformConfig.Admins)),
			MinTokenBalance: platformConfig.MinTokenBalance,
		}
		copy(r.adminConfig.PlatformConfigs[platform].Admins, platformConfig.Admins)
	}

	return nil
}

func (r *MemoryRegistry) ForceEndProposal(proposalID uuid.UUID, status ProposalStatus) error {
	return r.UpdateProposalStatus(context.Background(), proposalID, status)
}

// Internal helper methods that don't acquire locks (to be used within methods that already have locks)
func (r *MemoryRegistry) isAdminInternal(userID string, platform string) bool {
	// Check global admins
	for _, admin := range r.adminConfig.Admins {
		if admin == userID {
			return true
		}
	}

	// Check platform-specific admins
	if platformConfig, ok := r.adminConfig.PlatformConfigs[platform]; ok {
		for _, admin := range platformConfig.Admins {
			if admin == userID {
				return true
			}
		}
	}

	return false
}

func (r *MemoryRegistry) getMinTokenBalanceInternal(platform string) float64 {
	if platformConfig, ok := r.adminConfig.PlatformConfigs[platform]; ok {
		return platformConfig.MinTokenBalance
	}

	return r.adminConfig.MinTokenBalance // Fall back to global setting
}

func (r *MemoryRegistry) SetMinTokenBalance(balance float64) error {
	r.Lock()
	defer r.Unlock()

	if balance < 0 {
		return fmt.Errorf("minimum balance cannot be negative")
	}

	r.adminConfig.MinTokenBalance = balance
	return nil
}

func (r *MemoryRegistry) SetMinTokenBalanceForPlatform(platform string, balance float64) error {
	r.Lock()
	defer r.Unlock()

	if balance < 0 {
		return fmt.Errorf("minimum balance cannot be negative")
	}

	// Create platform config if it doesn't exist
	if _, ok := r.adminConfig.PlatformConfigs[platform]; !ok {
		r.adminConfig.PlatformConfigs[platform] = &PlatformConfig{
			Admins:          []string{},
			MinTokenBalance: 0,
		}
	}

	r.adminConfig.PlatformConfigs[platform].MinTokenBalance = balance
	return nil
}

func (r *MemoryRegistry) GetVotes(ctx context.Context, proposalID uuid.UUID) ([]Vote, error) {
	r.RLock()
	defer r.RUnlock()

	// Check if proposal exists
	_, ok := r.proposals[proposalID]
	if !ok {
		return nil, fmt.Errorf("proposal not found")
	}

	// Return a copy of the votes to avoid concurrent modification issues
	votes := make([]Vote, len(r.votes[proposalID]))
	copy(votes, r.votes[proposalID])

	return votes, nil
}
