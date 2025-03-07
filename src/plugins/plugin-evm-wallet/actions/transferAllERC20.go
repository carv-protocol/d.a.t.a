package actions

import (
	"context"
	"fmt"
	"time"

	"github.com/carv-protocol/d.a.t.a/src/plugins/plugin-evm-wallet/actions/clients"
)

type TransferAllERC20Action struct {
	client     *clients.BaseClient
	actionType string
}

func NewTransferAllERC20Action(
	privateKey string,
	network string,
	rpcURL string,
	chainID int64,
	timeout time.Duration,
	actionType string,
) (*TransferAllERC20Action, error) {
	client, err := clients.NewBaseClient(clients.Config{
		RPC:        rpcURL,
		ChainID:    chainID,
		Timeout:    timeout,
		PrivateKey: privateKey,
	})
	if err != nil {
		return nil, err
	}

	return &TransferAllERC20Action{
		client:     client,
		actionType: actionType,
	}, nil
}

func (a *TransferAllERC20Action) Name() string {
	return "TransferAllERC20Action"
}

func (a *TransferAllERC20Action) Description() string {
	return "Transfer all of a given ERC20 tokens from self to another on Base chain"
}

func (a *TransferAllERC20Action) Type() string {
	return a.actionType
}

/*
  Parameters:
    - toAddress: string
		- erc20Address: string
    - network: string
*/

func (a *TransferAllERC20Action) Validate(params map[string]interface{}) error {
	erc20Address := params["erc20Address"].(string)
	if erc20Address == "" {
		return fmt.Errorf("erc20Address is required")
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

func (a *TransferAllERC20Action) ParametersPrompt() string {
	return `
	{
		"erc20Address": <The address of the ERC20 token to transfer, this needs to be a valid ERC20 address>,
		"toAddress": <The address to transfer the ERC20 token to, this needs to be a valid address>,
		"network": <The network to transfer the ERC20 token on. Only support base now.>
	}
	`
}

func (a *TransferAllERC20Action) Execute(ctx context.Context, params map[string]interface{}) error {
	erc20Address := params["erc20Address"].(string)
	toAddress := params["toAddress"].(string)

	balance, err := a.client.GetERC20TokenBalance(ctx, erc20Address, a.client.GetAddress(ctx))
	if err != nil {
		return err
	}

	_, err = a.client.TransferERC20Token(ctx, &clients.ERC20TokenTransferInput{
		TokenAddress: erc20Address,
		To:           toAddress,
		Amount:       balance.Amount,
		GasLimit:     100_000,
	})

	if err != nil {
		return err
	}

	return nil
}
