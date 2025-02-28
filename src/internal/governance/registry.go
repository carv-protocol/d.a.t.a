package governance

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/carv-protocol/d.a.t.a/src/internal/types"
	"github.com/google/uuid"
)

// Registry defines the interface for governance operations
type Registry interface {
	// Proposal management
	CreateProposal(ctx context.Context, title, description, creator, platform string, startTime, endTime int64) (*Proposal, error)
	GetProposal(ctx context.Context, id uuid.UUID) (*Proposal, error)
	ListProposals(ctx context.Context, status ProposalStatus) ([]*Proposal, error)
	UpdateProposalStatus(ctx context.Context, id uuid.UUID, status ProposalStatus) error
	ForceEndProposal(proposalID uuid.UUID, status ProposalStatus) error

	// Voting
	CastVote(ctx context.Context, proposalID uuid.UUID, voter string, platform string, option VoteOption) error
	TallyVotes(ctx context.Context, proposalID uuid.UUID) (map[VoteOption]float64, error)
	GetVotes(ctx context.Context, proposalID uuid.UUID) ([]Vote, error)
	GetProposalResult(ctx context.Context, proposalID uuid.UUID) (ProposalStatus, error)

	// Admin management
	IsAdmin(userID string) bool
	IsAdminForPlatform(userID string, platform string) bool
	SetAdmin(userID string) error
	RemoveAdmin(userID string) error

	// Token balance management
	GetMinTokenBalance() float64
	GetMinTokenBalanceForPlatform(platform string) float64
	SetMinTokenBalance(balance float64) error
	SetMinTokenBalanceForPlatform(platform string, balance float64) error
	GetTokenBalance(ctx context.Context, userID string, platform string) (*types.TokenBalance, error)

	// Admin config
	UpdateAdminConfig(config *AdminConfig) error
}

type registryImpl struct {
	db           *sql.DB
	adminConfig  *AdminConfig
	tokenManager TokenManager
}

type TokenManager interface {
	// GetBalance(ctx context.Context, userID string, platform string) (float64, error)
	FetchNativeTokenBalance(ctx context.Context, userID string, platform string) (*types.TokenBalance, error)
}

func NewRegistry(db *sql.DB, tokenManager TokenManager, adminConfig *AdminConfig) Registry {
	if adminConfig == nil {
		adminConfig = &AdminConfig{
			Admins:          []string{},
			MinTokenBalance: 1000, // Default minimum balance
		}
	}

	return &registryImpl{
		db:           db,
		tokenManager: tokenManager,
		adminConfig:  adminConfig,
	}
}

func (r *registryImpl) IsAdmin(userID string) bool {
	for _, admin := range r.adminConfig.Admins {
		if admin == userID {
			return true
		}
	}
	return false
}

func (r *registryImpl) SetAdmin(userID string) error {
	// Check if already admin
	if r.IsAdmin(userID) {
		return fmt.Errorf("user %s is already an admin", userID)
	}

	r.adminConfig.Admins = append(r.adminConfig.Admins, userID)
	return nil
}

func (r *registryImpl) RemoveAdmin(userID string) error {
	for i, admin := range r.adminConfig.Admins {
		if admin == userID {
			// Remove admin by replacing with last element and truncating
			r.adminConfig.Admins[i] = r.adminConfig.Admins[len(r.adminConfig.Admins)-1]
			r.adminConfig.Admins = r.adminConfig.Admins[:len(r.adminConfig.Admins)-1]
			return nil
		}
	}
	return fmt.Errorf("user %s is not an admin", userID)
}

func (r *registryImpl) GetMinTokenBalance() float64 {
	return r.adminConfig.MinTokenBalance
}

func (r *registryImpl) SetMinTokenBalance(balance float64) error {
	if balance < 0 {
		return fmt.Errorf("minimum balance cannot be negative")
	}
	r.adminConfig.MinTokenBalance = balance
	return nil
}

func (r *registryImpl) GetTokenBalance(ctx context.Context, userID string, platform string) (*types.TokenBalance, error) {
	return r.tokenManager.FetchNativeTokenBalance(ctx, userID, platform)
}

func (r *registryImpl) CreateProposal(ctx context.Context, title, description, creator, platform string, startTime, endTime int64) (*Proposal, error) {
	// Implementation
	return nil, fmt.Errorf("not implemented")
}

func (r *registryImpl) GetProposal(ctx context.Context, id uuid.UUID) (*Proposal, error) {
	// Implementation
	return nil, fmt.Errorf("not implemented")
}

func (r *registryImpl) ListProposals(ctx context.Context, status ProposalStatus) ([]*Proposal, error) {
	// Implementation
	return nil, fmt.Errorf("not implemented")
}

func (r *registryImpl) UpdateProposalStatus(ctx context.Context, id uuid.UUID, status ProposalStatus) error {
	// Implementation
	return fmt.Errorf("not implemented")
}

func (r *registryImpl) CastVote(ctx context.Context, proposalID uuid.UUID, voter string, platform string, option VoteOption) error {
	// Check if user has enough tokens to vote
	balance, err := r.GetTokenBalance(ctx, voter, platform)
	if err != nil {
		return fmt.Errorf("failed to check token balance: %v", err)
	}

	minBalance := r.GetMinTokenBalance()
	if balance.Balance < minBalance {
		return fmt.Errorf("insufficient token balance to vote. Required: %f, Current: %f", minBalance, balance.Balance)
	}

	// Implementation
	return fmt.Errorf("not implemented")
}

func (r *registryImpl) TallyVotes(ctx context.Context, proposalID uuid.UUID) (map[VoteOption]float64, error) {
	// Implementation
	return nil, fmt.Errorf("not implemented")
}

func (r *registryImpl) GetVotes(ctx context.Context, proposalID uuid.UUID) ([]Vote, error) {
	// Implementation
	return nil, fmt.Errorf("not implemented")
}

func (r *registryImpl) GetProposalResult(ctx context.Context, proposalID uuid.UUID) (ProposalStatus, error) {
	// Implementation
	return ProposalStatusPending, fmt.Errorf("not implemented")
}

func (r *registryImpl) ForceEndProposal(proposalID uuid.UUID, status ProposalStatus) error {
	return r.UpdateProposalStatus(context.Background(), proposalID, status)
}

func (r *registryImpl) UpdateAdminConfig(config *AdminConfig) error {
	if config == nil {
		return fmt.Errorf("admin config cannot be nil")
	}
	r.adminConfig = config
	return nil
}

func (r *registryImpl) IsAdminForPlatform(userID string, platform string) bool {
	// Implementation
	return false
}

func (r *registryImpl) GetMinTokenBalanceForPlatform(platform string) float64 {
	// Implementation
	return 0
}

func (r *registryImpl) SetMinTokenBalanceForPlatform(platform string, balance float64) error {
	// Implementation
	return nil
}
