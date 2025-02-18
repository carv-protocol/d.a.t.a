package token

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"time"

	"github.com/carv-protocol/d.a.t.a/src/internal/core"
	"github.com/carv-protocol/d.a.t.a/src/internal/memory"
)

// StakeholderManager manages stakeholder interactions and influences
type StakeholderManager struct {
	memoryManager memory.Manager
	store         *StakeholderStore
}

func NewStakeholderManager(memoryManager memory.Manager) *StakeholderManager {
	return &StakeholderManager{
		memoryManager: memoryManager,
	}
}

// ProcessMessage handles new input from social media
func (sm *StakeholderManager) FetchOrCreateStakeholder(
	ctx context.Context,
	id string,
	platform string,
	stakeholderType core.StakeholderType,
) (*core.Stakeholder, error) {
	key := fmt.Sprintf("%s:%s", platform, id)
	var stakeholder *core.Stakeholder
	mem, err := sm.memoryManager.GetMemory(ctx, key)
	if err != nil {
		return nil, err
	}
	// stakeholder doesn't exist
	if mem == nil {
		stakeholder = &core.Stakeholder{
			Key:            key,
			ID:             id,
			CarvID:         "",
			Platform:       platform,
			Type:           stakeholderType,
			HistoricalMsgs: []string{},
		}

		res, err := json.Marshal(stakeholder)
		if err != nil {
			return nil, err
		}
		sm.memoryManager.CreateMemory(ctx, memory.Memory{
			MemoryID:  key,
			CreatedAt: time.Now(),
			Content:   res,
		})
	} else {
		err = json.Unmarshal(mem.Content, &stakeholder)
		if err != nil {
			return nil, err
		}
	}

	return stakeholder, nil
}

// AddHistoricalMsg adds a new historical message to a stakeholder's record
func (sm *StakeholderManager) AddHistoricalMsg(ctx context.Context, id, platform string, msgs []string) error {
	key := fmt.Sprintf("%s:%s", platform, id)
	var stakeholder *core.Stakeholder
	mem, err := sm.memoryManager.GetMemory(ctx, key)
	if err != nil {
		return err
	}
	if mem == nil {
		return fmt.Errorf("stakeholder doesn't exist")
	}

	err = json.Unmarshal(mem.Content, &stakeholder)
	if err != nil {
		return err
	}
	stakeholder.HistoricalMsgs = append(stakeholder.HistoricalMsgs, msgs...)

	res, err := json.Marshal(stakeholder)
	if err != nil {
		return err
	}

	return sm.memoryManager.SetMemory(ctx, &memory.Memory{
		MemoryID:  mem.MemoryID,
		CreatedAt: mem.CreatedAt,
		Content:   res,
	})
}

// GetAggregatedPreferences gets current preferences weighted by stake
func (sm *StakeholderManager) GetAggregatedPreferences(ctx context.Context) (map[string]interface{}, error) {
	// Get all stakeholder states
	states, err := sm.store.GetAllStates(ctx)
	if err != nil {
		return nil, err
	}

	// Aggregate preferences weighted by token holdings
	aggregated := make(map[string]interface{})
	for _, state := range states {
		weight := calculateWeight(state.TokenBalance)
		for k, pref := range state.Preferences {
			aggregated[k] = aggregatePreference(aggregated[k], pref, weight)
		}
	}

	return aggregated, nil
}

// aggregatePreference combines two preference values based on weight
// The exact implementation depends on the type of preference value
func aggregatePreference(existing, new interface{}, weight float64) interface{} {
	switch v := new.(type) {
	case float64:
		// For numeric preferences, do a weighted average
		if e, ok := existing.(float64); ok {
			return e*(1-weight) + v*weight
		}
		return v

	case bool:
		// For boolean preferences, use weight as probability threshold
		if e, ok := existing.(bool); ok {
			// If weights heavily favor the new value, use it
			if weight > 0.7 {
				return v
			}
			// Otherwise keep existing
			return e
		}
		return v

	case string:
		// For string preferences, keep the value from the higher weight
		if e, ok := existing.(string); ok {
			if weight > 0.5 {
				return v
			}
			return e
		}
		return v

	case map[string]interface{}:
		// For nested preferences, recursively aggregate each field
		result := make(map[string]interface{})
		if e, ok := existing.(map[string]interface{}); ok {
			// Start with existing values
			for k, val := range e {
				result[k] = val
			}
			// Update with new values using weight
			for k, val := range v {
				if existingVal, exists := result[k]; exists {
					result[k] = aggregatePreference(existingVal, val, weight)
				} else {
					result[k] = val
				}
			}
			return result
		}
		return v

	case []interface{}:
		// For array preferences, combine arrays with weight-based selection
		if e, ok := existing.([]interface{}); ok {
			// Calculate how many items to take from each array based on weight
			newCount := int(float64(len(v)) * weight)
			oldCount := len(e) - newCount
			if oldCount < 0 {
				oldCount = 0
			}

			// Create combined result
			result := make([]interface{}, 0, oldCount+newCount)
			result = append(result, e[:oldCount]...)
			result = append(result, v[:newCount]...)
			return result
		}
		return v

	default:
		// For unsupported types, just return the new value
		return v
	}
}

// calculateWeight determines a stakeholder's voting weight based on their token balance
// Returns a normalized weight between 0 and 1
func calculateWeight(balance *big.Int) float64 {
	if balance == nil || balance.Sign() <= 0 {
		return 0.0
	}

	// Convert balance to float64 for calculations
	balanceFloat := new(big.Float).SetInt(balance)
	weightFloat, _ := balanceFloat.Float64()

	// Apply logarithmic scaling to prevent large token holders from having too much influence
	// while still maintaining meaningful weight differences
	// Using log base 10 plus 1 to ensure positive weights and handle small balances
	weight := math.Log10(weightFloat + 1)

	// Normalize weight to be between 0 and 1
	// You might want to adjust these constants based on your token economics
	const maxLogWeight = 15.0 // log10 of 1 quadrillion (10^15) + 1
	normalizedWeight := weight / maxLogWeight

	// Ensure weight is between 0 and 1
	if normalizedWeight > 1.0 {
		return 1.0
	}
	return normalizedWeight
}
