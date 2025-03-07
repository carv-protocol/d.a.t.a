package actions

import (
	"context"
	"time"

	"github.com/carv-protocol/d.a.t.a/src/plugins/plugin-evm-wallet/actions/clients"
)

type TransferAction struct {
	client     *clients.BaseClient
	actionType string
}

func NewTransferAction(
	privateKey string,
	network string,
	rpcURL string,
	chainID int64,
	timeout time.Duration,
	actionType string,
) (*TransferAction, error) {
	client, err := clients.NewBaseClient(clients.Config{
		RPC:     rpcURL,
		ChainID: chainID,
		Timeout: timeout,
	})
	if err != nil {
		return nil, err
	}

	return &TransferAction{
		client:     client,
		actionType: actionType,
	}, nil
}

func (a *TransferAction) Name() string {
	return "Transfer Token on Base chain"
}

func (a *TransferAction) Description() string {
	return "Transfer tokens from one address to another on Base chain"
}

func (a *TransferAction) Type() string {
	return a.actionType
}

/*
  Parameters:
    - toAddress: string
    - amount: string
    - privateKey: string
    - network: string
    - rpcURL: string
*/

func (a *TransferAction) Validate(params map[string]interface{}) error {

	return nil
}

func (a *TransferAction) ParametersPrompt() string {
	return ""
}

func (a *TransferAction) Execute(ctx context.Context, params map[string]interface{}) error {
	return nil
}
