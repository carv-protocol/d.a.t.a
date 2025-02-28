package types

import (
	"context"
)

// StakeholderType defines the type of stakeholder
type StakeholderType string

const (
	StakeholderTypePriority StakeholderType = "priority"
	StakeholderTypeUser     StakeholderType = "user"
)

// SocialMessage represents a message from a social platform
type SocialMessage struct {
	Type        string                 // Type of message (e.g. "message", "command", "mention")
	Content     string                 // Message content
	Platform    string                 // Platform the message is from/to (e.g. "discord", "twitter")
	FromUser    string                 // User who sent the message
	TargetUsers []string               // Users the message is targeting (for mentions)
	Metadata    map[string]interface{} // Additional platform-specific metadata
}

// GetContent implements the plugin core SocialMessage interface
func (m *SocialMessage) GetContent() string {
	return m.Content
}

// GetFromUser implements the plugin core SocialMessage interface
func (m *SocialMessage) GetFromUser() string {
	return m.FromUser
}

// GetPlatform implements the plugin core SocialMessage interface
func (m *SocialMessage) GetPlatform() string {
	return m.Platform
}

// GetMetadata implements the plugin core SocialMessage interface
func (m *SocialMessage) GetMetadata() map[string]interface{} {
	return m.Metadata
}

// SocialClient defines the interface for social platform interactions
type SocialClient interface {
	SendMessage(ctx context.Context, msg SocialMessage) error
	GetMessageChannel() chan SocialMessage
	MonitorMessages(ctx context.Context)
}

// LLMClient defines the interface for LLM interactions
type LLMClient interface {
	GenerateResponse(ctx context.Context, prompt string) (string, error)
}

// StakeholderManager defines the interface for stakeholder management
type StakeholderManager interface {
	FetchOrCreateStakeholder(ctx context.Context, userID, platform string, stakeholderType StakeholderType) (*Stakeholder, error)
	GetAggregatedPreferences(ctx context.Context) (map[string]interface{}, error)
	AddHistoricalMsg(ctx context.Context, userID, platform string, messages []string) error
}

// Stakeholder represents a user in the system
type Stakeholder struct {
	Key            string
	ID             string
	Platform       string
	CarvID         string
	Type           StakeholderType
	TokenBalance   *TokenBalance
	HistoricalMsgs []string
}

// GetID implements the plugin core Stakeholder interface
func (s *Stakeholder) GetID() string {
	return s.ID
}

// GetType implements the plugin core Stakeholder interface
func (s *Stakeholder) GetType() string {
	return string(s.Type)
}

// GetHistoricalMsgs implements the plugin core Stakeholder interface
func (s *Stakeholder) GetHistoricalMsgs() []string {
	return s.HistoricalMsgs
}

// GetTokenBalance implements the plugin core Stakeholder interface
func (s *Stakeholder) GetTokenBalance() interface{} {
	if s.TokenBalance == nil {
		return float64(0)
	}
	return s.TokenBalance.Balance
}

// TokenInfo is a struct for token information
type TokenInfo struct {
	Network      string
	Ticker       string
	ContractAddr string
}

// TokenBalance represents a user's token balance
type TokenBalance struct {
	TokenInfo
	Balance float64
}

// MessageSender defines the interface for sending messages
type MessageSender interface {
	SendMessage(ctx context.Context, msg SocialMessage) error
}
