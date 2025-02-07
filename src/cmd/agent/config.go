package main

import (
	"github.com/carv-protocol/d.a.t.a/src/pkg/clients"
	"github.com/carv-protocol/d.a.t.a/src/pkg/llm"
)

type Config struct {
	Character struct {
		Path string `mapstructure:"path"`
	} `mapstructure:"character"`

	Database struct {
		Type string `mapstructure:"type"`
		Path string `mapstructure:"path"`
	} `mapstructure:"database"`

	llm.LLMConfig `mapstructure:"llm_config"`

	Data struct {
		CarvID struct {
			URL    string `mapstructure:"url"`
			APIKey string `mapstructure:"api_key"`
		} `mapstructure:"carvid"`
	} `mapstructure:"data"`

	Social struct {
		clients.TwitterConfig  `mapstructure:"twitter"`
		clients.DiscordConfig  `mapstructure:"discord"`
		clients.TelegramConfig `mapstructure:"telegram"`
	} `mapstructure:"social"`
}
