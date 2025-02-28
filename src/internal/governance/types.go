package governance

import (
	"time"

	"github.com/google/uuid"
)

// ProposalStatus represents the current state of a proposal
type ProposalStatus string

const (
	ProposalStatusPending   ProposalStatus = "PENDING"
	ProposalStatusActive    ProposalStatus = "ACTIVE"
	ProposalStatusPassed    ProposalStatus = "PASSED"
	ProposalStatusRejected  ProposalStatus = "REJECTED"
	ProposalStatusCancelled ProposalStatus = "CANCELLED"
)

// VoteOption represents a voting choice
type VoteOption string

const (
	VoteOptionYes     VoteOption = "YES"
	VoteOptionNo      VoteOption = "NO"
	VoteOptionAbstain VoteOption = "ABSTAIN"
)

// Proposal represents a governance proposal
type Proposal struct {
	ID          uuid.UUID      `json:"id"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Creator     string         `json:"creator"`
	Platform    string         `json:"platform"`
	Status      ProposalStatus `json:"status"`
	StartTime   time.Time      `json:"start_time"`
	EndTime     time.Time      `json:"end_time"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// Vote represents a vote cast on a proposal
type Vote struct {
	ProposalID uuid.UUID  `json:"proposal_id"`
	Voter      string     `json:"voter"`
	Option     VoteOption `json:"option"`
	CreatedAt  time.Time  `json:"created_at"`
}

// VotingPower represents the voting power of a stakeholder
type VotingPower struct {
	Stakeholder string    `json:"stakeholder"`
	Platform    string    `json:"platform"`
	Power       float64   `json:"power"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type AdminConfig struct {
	Admins          []string
	MinTokenBalance float64
	// Platform specific configurations
	PlatformConfigs map[string]*PlatformConfig
}

type PlatformConfig struct {
	Admins          []string
	MinTokenBalance float64
}
