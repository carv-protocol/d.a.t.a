package actions

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/carv-protocol/d.a.t.a/src/tools/wallet/actions/clients"
)

type TransferERC20Action struct {
	client     *clients.BaseClient
	actionType string
}

func NewTransferERC20Action(
	privateKey string,
	network string,
	rpcURL string,
	chainID int64,
	timeout time.Duration,
	actionType string,
) (*TransferERC20Action, error) {
	client, err := clients.NewBaseClient(clients.Config{
		RPC:        rpcURL,
		ChainID:    chainID,
		Timeout:    timeout,
		PrivateKey: privateKey,
	})
	if err != nil {
		return nil, err
	}

	return &TransferERC20Action{
		client:     client,
		actionType: actionType,
	}, nil
}

func (a *TransferERC20Action) Name() string {
	return "Transfer ERC20 Token on Base chain"
}

func (a *TransferERC20Action) Description() string {
	return "Transfer ERC20 tokens from one address to another on Base chain"
}

func (a *TransferERC20Action) Type() string {
	return a.actionType
}

/*
  Parameters:
    - toAddress: string
    - amount: string
		- erc20Address: string
    - network: string
*/

func (a *TransferERC20Action) Validate(params map[string]interface{}) error {
	erc20Address := params["erc20Address"].(string)
	if erc20Address == "" {
		return fmt.Errorf("erc20Address is required")
	}

	amount := params["amount"].(float64)
	if amount <= 0 {
		return fmt.Errorf("amount must be greater than 0")
	}

	toAddress := params["toAddress"].(string)
	if toAddress == "" {
		return fmt.Errorf("toAddress is required")
	}

	network := params["network"].(string)
	if network == "" {
		return fmt.Errorf("network is required")
	}

	if network != "base" {
		return fmt.Errorf("network must be base")
	}

	return nil
}

func (a *TransferERC20Action) ParametersPrompt() string {
	return ` Parameters in JSON format:
	{
		"erc20Address": "0x1234567890123456789012345678901234567890",
		"amount": 1.0,
		"toAddress": "0x1234567890123456789012345678901234567890",
		"network": "base"
	}
	`
}

func (a *TransferERC20Action) Execute(ctx context.Context, params map[string]interface{}) error {
	erc20Address := params["erc20Address"].(string)
	amount := params["amount"].(float64)
	toAddress := params["toAddress"].(string)

	_, err := a.client.TransferERC20Token(ctx, &clients.ERC20TokenTransferInput{
		TokenAddress: erc20Address,
		To:           toAddress,
		Amount:       big.NewFloat(amount),
	})

	if err != nil {
		return err
	}

	return nil
}
