package core

import (
	"context"
	"time"

	"github.com/carv-protocol/d.a.t.a/src/internal/actions"
	"github.com/carv-protocol/d.a.t.a/src/internal/types"
)

// StakeholderType is an enum for stakeholder types
type StakeholderType string

const (
	// StakeholderTypeUser is a stakeholder type for users
	StakeholderTypeUser StakeholderType = "user"
	// StakeholderTypeStakeholder is a stakeholder type for stakeholders
	StakeholderTypePriority StakeholderType = "priority"
)

// // Stakeholder is a stakeholder of the agent
// type Stakeholder struct {
// 	Key            string
// 	ID             string
// 	Platform       string
// 	CarvID         string
// 	Type           StakeholderType
// 	TokenBalance   *TokenBalance
// 	HistoricalMsgs []string
// }

// // TokenInfo is a struct for token information
// type TokenInfo struct {
// 	Network      string
// 	Ticker       string
// 	ContractAddr string
// }

// // TokenBalance is a struct for token balance information
// type TokenBalance struct {
// 	TokenInfo
// 	Balance float64
// }

// TaskStatus is an enum for task status
type TaskStatus string

const (
	TaskStatusActive    TaskStatus = "active"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusFailed    TaskStatus = "failed"
)

// Task is a task that the agent can execute
type Task struct {
	ID                       string
	Name                     string
	Description              string
	Priority                 float64
	ExecutionSteps           []string
	Status                   TaskStatus
	Deadline                 *time.Time
	RequiresApproval         bool
	RequiresStakeholderInput bool
	Tools                    []string
	CreatedBy                string
	CreatedAt                time.Time
	UpdatedAt                time.Time
}

// StakeholderManager is an interface for managing stakeholders
type StakeholderManager interface {
	FetchOrCreateStakeholder(ctx context.Context, id, platform string, stakeholderType types.StakeholderType) (*types.Stakeholder, error)
	AddHistoricalMsg(ctx context.Context, id, platform string, msgs []string) error
	GetAggregatedPreferences(ctx context.Context) (map[string]interface{}, error)
}

// TokenManager is an interface for managing tokens
type TokenManager interface {
	FetchNativeTokenBalance(ctx context.Context, id, platform string) (*types.TokenBalance, error)
	NativeTokenInfo(ctx context.Context) (*types.TokenInfo, error)
}

// TaskManager is an interface for managing tasks
type TaskManager interface {
	AddTask(ctx context.Context, task Task) error
	GetTasks(ctx context.Context) ([]*Task, error)
}

// Tool interface
type Tool interface {
	Initialize(ctx context.Context) error
	Name() string
	Description() string
	AvailableActions() []actions.IAction
}

// ToolManager is an interface for managing tools
type ToolManager interface {
	Register(tool Tool) error
	AvailableTools() []Tool
	AvailableActions() []actions.IAction
	GetAction(actionType string, actionName string) (actions.IAction, error)
}

// IntentType defines different types of intents
type IntentType string

const (
	IntentQuestion    IntentType = "question"
	IntentFeedback    IntentType = "feedback"
	IntentComplaint   IntentType = "complaint"
	IntentSuggestion  IntentType = "suggestion"
	IntentGreeting    IntentType = "greeting"
	IntentInquiry     IntentType = "inquiry"
	IntentRequest     IntentType = "request"
	IntentAcknowledge IntentType = "acknowledge"
)

// EntityType defines different types of entities
type EntityType string

const (
	EntityPerson   EntityType = "person"
	EntityProduct  EntityType = "product"
	EntityCompany  EntityType = "company"
	EntityLocation EntityType = "location"
	EntityDateTime EntityType = "datetime"
	EntityCrypto   EntityType = "crypto"
	EntityWallet   EntityType = "wallet"
	EntityContract EntityType = "contract"
)

// EmotionType defines different types of emotions
type EmotionType string

const (
	EmotionPositive EmotionType = "positive"
	EmotionNegative EmotionType = "negative"
	EmotionNeutral  EmotionType = "neutral"
)

type ProcessedAction struct {
	ActionType string `json:"action_type"`
	ActionName string `json:"action_name"`
}

// SocialClient defines the interface for social media interactions
type SocialClient interface {
	SendMessage(ctx context.Context, msg types.SocialMessage) error
	GetMessageChannel() <-chan types.SocialMessage
	GetErrorChannel() <-chan error
	MonitorMessages(ctx context.Context)
}
