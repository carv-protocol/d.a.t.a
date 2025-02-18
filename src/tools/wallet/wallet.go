package wallet

import (
	"context"
	"time"

	"github.com/carv-protocol/d.a.t.a/src/internal/actions"
	walletactions "github.com/carv-protocol/d.a.t.a/src/tools/wallet/actions"
)

const ActionTypeWallet string = "wallet"

// Only support Base chain now
type Config struct {
	PrivateKey string        `mapstructure:"private_key"`
	Network    string        `mapstructure:"network"`
	RPCURL     string        `mapstructure:"rpc_url"`
	ChainID    int64         `mapstructure:"chain_id"`
	Timeout    time.Duration `mapstructure:"timeout"`
}

type WalletTool struct {
	actions []actions.IAction
}

func NewWalletTool(config *Config) (*WalletTool, error) {
	transferAllERC20Action, err := walletactions.NewTransferAllERC20Action(
		config.PrivateKey,
		config.Network,
		config.RPCURL,
		config.ChainID,
		config.Timeout,
		ActionTypeWallet,
	)
	if err != nil {
		return nil, err
	}

	return &WalletTool{
		actions: []actions.IAction{transferAllERC20Action},
	}, nil
}

func (t *WalletTool) Initialize(ctx context.Context) error {
	return nil
}

func (t *WalletTool) Name() string {
	return "Wallet tool"
}

func (t *WalletTool) Description() string {
	return `The wallet tool allows you to manage your wallet. You can control your own wallet to send ETH or ERC20 tokens.`
}

func (t *WalletTool) AvailableActions() []actions.IAction {
	return t.actions
}
