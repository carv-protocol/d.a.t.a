package data

import (
	"context"

	"github.com/carv-protocol/d.a.t.a/src/internal/actions"
)

type CARVDataTool struct {
}

func (t *CARVDataTool) Initialize(ctx context.Context) error {
	return nil
}

func (t *CARVDataTool) Name() string {
	return "carv's d.a.t.a tool"
}

func (t *CARVDataTool) Description() string {
	return `CARV's d.a.t.a tool allows you to fetch data from both on-chain and off-chain.
	It can fetch data from current Ethereum network, Base network, Solana network etc.
	It can also help you identify a Twitter account owner's on-chain address and the corresponding balance for different token.
	It can also fetch the token info including the smart contract addresses and the token price.`
}

func (t *CARVDataTool) AvailableActions() []actions.IAction {
	return nil
}
