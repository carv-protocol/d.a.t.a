package carv

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

type Client struct {
	APIKey     string
	BaseURL    string
	httpClient *http.Client
}

type Balance struct {
	Amount       float64
	Network      string
	ContractAddr string
}

func NewClient(apiKey string, baseURL string) *Client {
	return &Client{
		APIKey:  apiKey,
		BaseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (d *Client) GetBalanceByDiscordID(
	ctx context.Context,
	discordID string,
	chainName string,
	tokenTicker string,
) (*Balance, error) {
	// Input validation
	if discordID == "" || chainName == "" || tokenTicker == "" {
		return nil, fmt.Errorf("discordID, chainName, and tokenTicker cannot be empty")
	}

	url := fmt.Sprintf(
		"%s/user_balance_by_discord_id?discord_user_id=%s&chain_name=%s&token_ticker=%s",
		d.BaseURL,
		discordID,
		chainName,
		tokenTicker,
	)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Add("Authorization", d.APIKey)

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	var balanceResponse struct {
		Data struct {
			Balance      string `json:"balance"`
			ContractAddr string `json:"contract_addr"`
		} `json:"data"`
		Code    int    `json:"code"`
		Message string `json:"msg"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&balanceResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Check API response status
	if resp.StatusCode != http.StatusOK || balanceResponse.Code != 0 {
		return nil, fmt.Errorf("API error: status=%d, code=%d, message=%s",
			resp.StatusCode, balanceResponse.Code, balanceResponse.Message)
	}

	floatValue, err := strconv.ParseFloat(balanceResponse.Data.Balance, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse balance value: %w", err)
	}

	return &Balance{
		Amount:       floatValue,
		Network:      chainName,
		ContractAddr: balanceResponse.Data.ContractAddr,
	}, nil
}
