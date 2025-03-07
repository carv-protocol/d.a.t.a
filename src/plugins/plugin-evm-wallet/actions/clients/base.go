package clients

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/params"
)

// Standard ERC20 ABI
const erc20ABI = `[
    {
        "constant": true,
        "inputs": [],
        "name": "name",
        "outputs": [{"name": "", "type": "string"}],
        "type": "function"
    },
    {
        "constant": true,
        "inputs": [],
        "name": "symbol",
        "outputs": [{"name": "", "type": "string"}],
        "type": "function"
    },
    {
        "constant": true,
        "inputs": [],
        "name": "decimals",
        "outputs": [{"name": "", "type": "uint8"}],
        "type": "function"
    },
    {
        "constant": true,
        "inputs": [{"name": "_owner", "type": "address"}],
        "name": "balanceOf",
        "outputs": [{"name": "balance", "type": "uint256"}],
        "type": "function"
    },
		{
      "inputs":[
         {
            "internalType":"address",
            "name":"to",
            "type":"address"
         },
         {
            "internalType":"uint256",
            "name":"value",
            "type":"uint256"
         }
      ],
      "name":"transfer",
      "outputs":[
         {
            "internalType":"bool",
            "name":"",
            "type":"bool"
         }
      ],
      "stateMutability":"nonpayable",
      "type":"function"
   }
]`

// BaseClient represents a client for interacting with Base chain
type BaseClient struct {
	client     *ethclient.Client
	chainID    *big.Int
	PrivateKey *ecdsa.PrivateKey
	address    string
}

// Config holds the configuration for Base client
type Config struct {
	RPC        string
	ChainID    int64
	Timeout    time.Duration
	PrivateKey string
}

// NewBaseClient creates a new Base chain client
func NewBaseClient(cfg Config) (*BaseClient, error) {
	if strings.TrimSpace(cfg.PrivateKey) == "" {
		return nil, fmt.Errorf("private key cannot be empty")
	}

	if strings.TrimSpace(cfg.RPC) == "" {
		return nil, fmt.Errorf("RPC URL cannot be empty")
	}

	client, err := ethclient.Dial(cfg.RPC)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Base chain: %w", err)
	}

	// Verify chain ID
	chainID, err := client.ChainID(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get chain ID: %w", err)
	}

	if chainID.Int64() != cfg.ChainID {
		return nil, fmt.Errorf("unexpected chain ID: got %d, want %d", chainID, cfg.ChainID)
	}

	// Parse private key
	key, err := crypto.HexToECDSA(strings.TrimPrefix(cfg.PrivateKey, "0x"))
	if err != nil {
		return nil, fmt.Errorf("invalid private key: %w", err)
	}

	address := crypto.PubkeyToAddress(key.PublicKey)

	return &BaseClient{
		client:     client,
		chainID:    chainID,
		PrivateKey: key,
		address:    address.Hex(),
	}, nil
}

// Balance represents an account balance
type Balance struct {
	Address string
	Amount  *big.Float
	Symbol  string
}

func (c *BaseClient) GetAddress(ctx context.Context) string {
	return c.address
}

// GetBalance fetches the ETH balance for a given address
func (c *BaseClient) GetBalance(ctx context.Context, address string) (*Balance, error) {
	if !common.IsHexAddress(address) {
		return nil, fmt.Errorf("invalid ethereum address: %s", address)
	}

	account := common.HexToAddress(address)
	balance, err := c.client.BalanceAt(ctx, account, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}

	// Convert Wei to ETH
	ethBalance := new(big.Float).Quo(
		new(big.Float).SetInt(balance),
		new(big.Float).SetFloat64(params.Ether),
	)

	return &Balance{
		Address: address,
		Amount:  ethBalance,
		Symbol:  "ETH",
	}, nil
}

// TransferInput represents the input for a transfer transaction
type TransferInput struct {
	To       string
	Amount   *big.Float // in ETH
	GasLimit uint64
	GasPrice *big.Int
	Nonce    uint64
}

// TransferResult represents the result of a transfer transaction
type TransferResult struct {
	TokenAddress string
	TxHash       string
	From         string
	To           string
	Amount       *big.Float
	GasUsed      uint64
	Status       bool
	BlockHash    string
}

// Transfer sends ETH from one address to another
func (c *BaseClient) Transfer(ctx context.Context, input TransferInput) (*TransferResult, error) {
	// Validate addresses
	if !common.IsHexAddress(input.To) {
		return nil, fmt.Errorf("invalid ethereum address")
	}

	// Verify sender address matches private key
	address := crypto.PubkeyToAddress(c.PrivateKey.PublicKey)

	// Convert ETH to Wei
	amount := new(big.Float).Mul(input.Amount, new(big.Float).SetFloat64(params.Ether))
	amountWei, _ := amount.Int(new(big.Int))

	// Get nonce if not provided
	nonce := input.Nonce
	var err error
	if nonce == 0 {
		nonce, err = c.client.PendingNonceAt(ctx, address)
		if err != nil {
			return nil, fmt.Errorf("failed to get nonce: %w", err)
		}
	}

	// Get gas price if not provided
	gasPrice := input.GasPrice
	if gasPrice == nil {
		gasPrice, err = c.client.SuggestGasPrice(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get gas price: %w", err)
		}
	}

	// Create transaction
	tx := types.NewTransaction(
		nonce,
		common.HexToAddress(input.To),
		amountWei,
		input.GasLimit,
		gasPrice,
		nil,
	)

	// Sign transaction
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(c.chainID), c.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	// Send transaction
	err = c.client.SendTransaction(ctx, signedTx)
	if err != nil {
		return nil, fmt.Errorf("failed to send transaction: %w", err)
	}

	// Wait for transaction receipt
	receipt, err := c.waitForTransaction(ctx, signedTx.Hash())
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction receipt: %w", err)
	}

	return &TransferResult{
		TxHash:    receipt.TxHash.Hex(),
		From:      address.Hex(),
		To:        input.To,
		Amount:    input.Amount,
		GasUsed:   receipt.GasUsed,
		Status:    receipt.Status == 1,
		BlockHash: receipt.BlockHash.Hex(),
	}, nil
}

// ERC20

// GetERC20TokenBalance fetches the token balance for a given address and token
func (c *BaseClient) GetERC20TokenBalance(ctx context.Context, tokenAddress, address string) (*Balance, error) {
	// Validate addresses
	if !common.IsHexAddress(tokenAddress) || !common.IsHexAddress(address) {
		return nil, fmt.Errorf("invalid address: token=%s, holder=%s", tokenAddress, address)
	}

	// Get token info for decimals and symbol
	tokenInfo, err := c.GetTokenInfo(ctx, tokenAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get token info: %w", err)
	}

	// Parse ABI
	parsed, err := abi.JSON(strings.NewReader(erc20ABI))
	if err != nil {
		return nil, fmt.Errorf("failed to parse ABI: %w", err)
	}

	// Create contract binding
	contract := bind.NewBoundContract(common.HexToAddress(tokenAddress), parsed, c.client, c.client, c.client)

	// Call balanceOf
	var result []interface{}
	err = contract.Call(&bind.CallOpts{Context: ctx}, &result, "balanceOf", common.HexToAddress(address))
	if err != nil {
		return nil, fmt.Errorf("failed to get token balance: %w", err)
	}

	balance, ok := result[0].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("invalid balance type")
	}

	// Convert balance to token units
	decimals := new(big.Float).SetFloat64(math.Pow(10, float64(tokenInfo.Decimals)))
	tokenBalance := new(big.Float).Quo(new(big.Float).SetInt(balance), decimals)

	return &Balance{
		Address: address,
		Amount:  tokenBalance,
		Symbol:  tokenInfo.Symbol,
	}, nil
}

// ERC20TokenTransferInput represents the input for a token transfer
type ERC20TokenTransferInput struct {
	TokenAddress string
	To           string
	Amount       *big.Float
	GasLimit     uint64
	GasPrice     *big.Int
	Nonce        uint64
}

// TokenInfo represents ERC20 token information
type TokenInfo struct {
	Address  string
	Name     string
	Symbol   string
	Decimals uint8
}

// GetTokenInfo fetches information about an ERC20 token
func (c *BaseClient) GetTokenInfo(ctx context.Context, tokenAddress string) (*TokenInfo, error) {
	if !common.IsHexAddress(tokenAddress) {
		return nil, fmt.Errorf("invalid token address: %s", tokenAddress)
	}

	parsed, err := abi.JSON(strings.NewReader(erc20ABI))
	if err != nil {
		return nil, fmt.Errorf("failed to parse ABI: %w", err)
	}

	address := common.HexToAddress(tokenAddress)
	contract := bind.NewBoundContract(address, parsed, c.client, c.client, c.client)

	var (
		nameRes   []interface{}
		symbolRes []interface{}
		decRes    []interface{}
	)

	// Get token name
	err = contract.Call(&bind.CallOpts{Context: ctx}, &nameRes, "name")
	if err != nil {
		return nil, fmt.Errorf("failed to get token name: %w", err)
	}
	name, ok := nameRes[0].(string)
	if !ok {
		return nil, fmt.Errorf("invalid name type")
	}

	// Get token symbol
	err = contract.Call(&bind.CallOpts{Context: ctx}, &symbolRes, "symbol")
	if err != nil {
		return nil, fmt.Errorf("failed to get token symbol: %w", err)
	}
	symbol, ok := symbolRes[0].(string)
	if !ok {
		return nil, fmt.Errorf("invalid symbol type")
	}

	// Get token decimals
	err = contract.Call(&bind.CallOpts{Context: ctx}, &decRes, "decimals")
	if err != nil {
		return nil, fmt.Errorf("failed to get token decimals: %w", err)
	}
	decimals, ok := decRes[0].(uint8)
	if !ok {
		return nil, fmt.Errorf("invalid decimals type")
	}

	return &TokenInfo{
		Address:  tokenAddress,
		Name:     name,
		Symbol:   symbol,
		Decimals: decimals,
	}, nil
}

// TransferERC20Token transfers ERC20 tokens
func (c *BaseClient) TransferERC20Token(ctx context.Context, input *ERC20TokenTransferInput) (*TransferResult, error) {
	// Validate addresses
	if !common.IsHexAddress(input.TokenAddress) || !common.IsHexAddress(input.To) {
		return nil, fmt.Errorf("invalid address")
	}

	// Get token info for decimals
	tokenInfo, err := c.GetTokenInfo(ctx, input.TokenAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get token info: %w", err)
	}

	// Verify sender address matches private key
	address := crypto.PubkeyToAddress(c.PrivateKey.PublicKey)

	// Convert amount to token units
	decimals := new(big.Float).SetFloat64(math.Pow(10, float64(tokenInfo.Decimals)))
	amount := new(big.Float).Mul(input.Amount, decimals)
	tokenAmount, _ := amount.Int(new(big.Int))

	// Get nonce if not provided
	nonce := input.Nonce
	if nonce == 0 {
		nonce, err = c.client.PendingNonceAt(ctx, address)
		if err != nil {
			return nil, fmt.Errorf("failed to get nonce: %w", err)
		}
	}

	// Get gas price if not provided
	gasPrice := input.GasPrice
	if gasPrice == nil {
		gasPrice, err = c.client.SuggestGasPrice(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get gas price: %w", err)
		}
	}

	// Parse ABI
	parsed, err := abi.JSON(strings.NewReader(erc20ABI))
	if err != nil {
		return nil, fmt.Errorf("failed to parse ABI: %w", err)
	}

	// Prepare transaction data
	data, err := parsed.Pack("transfer", common.HexToAddress(input.To), tokenAmount)
	if err != nil {
		return nil, fmt.Errorf("failed to pack transfer data: %w", err)
	}

	tokenAddress := common.HexToAddress(input.TokenAddress)
	if input.GasLimit == 0 {
		gasLimit, err := c.client.EstimateGas(ctx, ethereum.CallMsg{
			To:   &tokenAddress,
			Data: data,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to estimate gas: %w", err)
		}
		input.GasLimit = gasLimit
	}

	// Create transaction
	tx := types.NewTransaction(
		nonce,
		common.HexToAddress(input.TokenAddress),
		big.NewInt(0),
		input.GasLimit,
		gasPrice,
		data,
	)

	// Sign transaction
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(c.chainID), c.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	// Send transaction
	err = c.client.SendTransaction(ctx, signedTx)
	if err != nil {
		return nil, fmt.Errorf("failed to send transaction: %w", err)
	}

	// Wait for transaction receipt
	receipt, err := c.waitForTransaction(ctx, signedTx.Hash())
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction receipt: %w", err)
	}

	return &TransferResult{
		TokenAddress: input.TokenAddress,
		TxHash:       receipt.TxHash.Hex(),
		From:         address.Hex(),
		To:           input.To,
		Amount:       input.Amount,
		GasUsed:      receipt.GasUsed,
		Status:       receipt.Status == 1,
		BlockHash:    receipt.BlockHash.Hex(),
	}, nil
}

// Helper functions

func EncodeTransactionToHex(signedTx *types.Transaction) (string, error) {
	// Encode the signed transaction to bytes
	txBytes, err := signedTx.MarshalBinary()
	if err != nil {
		return "", fmt.Errorf("failed to marshal transaction: %w", err)
	}

	// Convert to hex with "0x" prefix
	hexStr := "0x" + hex.EncodeToString(txBytes)
	return hexStr, nil
}

// waitForTransaction waits for a transaction to be mined
func (c *BaseClient) waitForTransaction(ctx context.Context, hash common.Hash) (*types.Receipt, error) {
	for {
		receipt, err := c.client.TransactionReceipt(ctx, hash)
		if err == nil {
			return receipt, nil
		}

		if err != ethereum.NotFound {
			return nil, err
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(time.Second):
			continue
		}
	}
}

// Close closes the client connection
func (c *BaseClient) Close() {
	if c.client != nil {
		c.client.Close()
	}
}
