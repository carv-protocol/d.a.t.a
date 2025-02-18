package token

import (
	"context"
	"fmt"

	"github.com/carv-protocol/d.a.t.a/src/internal/core"
	"github.com/carv-protocol/d.a.t.a/src/pkg/carv"
)

type TokenManager struct {
	// Implementation for token manager
	carvClient  *carv.Client
	nativeToken *core.TokenInfo
}

func NewTokenManager(carvClient *carv.Client, nativeToken *core.TokenInfo) *TokenManager {
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

func (t *TokenManager) FetchNativeTokenBalance(
	ctx context.Context,
	id string,
	platform string,
) (*core.TokenBalance, error) {
	if t.nativeToken == nil {
		return nil, fmt.Errorf("native token not set")
	}
	if platform == "discord" {
		balance, err := t.carvClient.GetBalanceByDiscordID(ctx, id, t.nativeToken.Network, t.nativeToken.Ticker)
		if err != nil {
			return nil, err
		}

		return &core.TokenBalance{
			TokenInfo: core.TokenInfo{
				Network: t.nativeToken.Network,
				Ticker:  t.nativeToken.Ticker,
			},
			Balance: balance.Amount,
		}, nil
	}

	return nil, fmt.Errorf("not supported platform")
}

func (t *TokenManager) NativeTokenInfo(
	ctx context.Context,
) (*core.TokenInfo, error) {
	if t.nativeToken == nil {
		return nil, nil
	}

	return t.nativeToken, nil
}
