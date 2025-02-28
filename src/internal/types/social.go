package types

import (
	"github.com/carv-protocol/d.a.t.a/src/pkg/clients"
)

type DiscordConfig struct {
	clients.DiscordConfig
	AdminIDs []string `mapstructure:"admin_ids"`
}

type TwitterConfig struct {
	clients.TwitterConfig
	AdminIDs []string `mapstructure:"admin_ids"`
}

type TelegramConfig struct {
	clients.TelegramConfig
	AdminIDs []string `mapstructure:"admin_ids"`
}

type SocialConfig struct {
	DiscordConfig  *DiscordConfig  `mapstructure:"discord_config"`
	TwitterConfig  *TwitterConfig  `mapstructure:"twitter_config"`
	TelegramConfig *TelegramConfig `mapstructure:"telegram_config"`
}
