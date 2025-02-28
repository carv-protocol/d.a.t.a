package token

import (
	"context"
	"fmt"

	"github.com/carv-protocol/d.a.t.a/src/internal/types"
	"github.com/carv-protocol/d.a.t.a/src/pkg/carv"
)

type TokenManager struct {
	// Implementation for token manager
	carvClient  *carv.Client
	nativeToken *types.TokenInfo
}

func NewTokenManager(carvClient *carv.Client, nativeToken *types.TokenInfo) *TokenManager {
	return &TokenManager{
		carvClient:  carvClient,
		nativeToken: nativeToken,
	}
}

// func (t *TokenManager) GetBalanceByDiscordID(ctx context.Context, discordID string, ticker string, network string) (*big.Int, error) {
// 	balance, err := t.carvClient.GetBalanceByDiscordID(ctx, discordID, network, ticker)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return balance.Amount, nil
// }

func (t *TokenManager) GetBalance(ctx context.Context, userID string, platform string) (float64, error) {
	if t.nativeToken == nil {
		return 0, fmt.Errorf("native token not set")
	}

	balance, err := t.FetchNativeTokenBalance(ctx, userID, platform)
	if err != nil {
		return 0, err
	}

	return balance.Balance, nil
}

func (t *TokenManager) FetchNativeTokenBalance(
	ctx context.Context,
	id string,
	platform string,
) (*types.TokenBalance, error) {
	if t.nativeToken == nil {
		return nil, fmt.Errorf("native token not set")
	}

	switch platform {
	case "discord":
		// Handle the numeric balance response from carv.Client
		balanceResp, err := t.carvClient.GetBalanceByDiscordID(ctx, id, t.nativeToken.Network, t.nativeToken.Ticker)
		if err != nil {
			return nil, fmt.Errorf("failed to get balance: %v", err)
		}

		// Create TokenBalance with the numeric value
		return &types.TokenBalance{
			TokenInfo: types.TokenInfo{
				Network: t.nativeToken.Network,
				Ticker:  t.nativeToken.Ticker,
			},
			Balance: balanceResp.Amount,
		}, nil

	case "twitter":
		// TODO: Implement Twitter balance check
		// For testing purposes, return a mock balance
		return &types.TokenBalance{
			TokenInfo: types.TokenInfo{
				Network: t.nativeToken.Network,
				Ticker:  t.nativeToken.Ticker,
			},
			Balance: 500, // Mock balance for testing
		}, nil

	case "telegram":
		// TODO: Implement Telegram balance check
		// For testing purposes, return a mock balance
		return &types.TokenBalance{
			TokenInfo: types.TokenInfo{
				Network: t.nativeToken.Network,
				Ticker:  t.nativeToken.Ticker,
			},
			Balance: 500, // Mock balance for testing
		}, nil

	default:
		return nil, fmt.Errorf("unsupported platform: %s", platform)
	}
}

func (t *TokenManager) NativeTokenInfo(
	ctx context.Context,
) (*types.TokenInfo, error) {
	if t.nativeToken == nil {
		return nil, nil
	}

	return t.nativeToken, nil
}
