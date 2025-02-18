package social

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/carv-protocol/d.a.t.a/src/internal/core"
	"github.com/carv-protocol/d.a.t.a/src/pkg/clients"
)

// SocialClientImpl handles social media interactions and error reporting
type SocialClientImpl struct {
	twitterClient    clients.ITwitter
	discordBot       *clients.DiscordBot
	telegramBot      *clients.TelegramClient
	socialMsgChannel chan core.SocialMessage
	errorChannel     chan error // Channel for reporting errors to agent
}

// NewSocialClient creates a new social client with error handling
func NewSocialClient(
	twitterConfig *clients.TwitterConfig,
	discordConfig *clients.DiscordConfig,
	telegramConfig *clients.TelegramConfig,
) *SocialClientImpl {
	cli := &SocialClientImpl{
		socialMsgChannel: make(chan core.SocialMessage),
		errorChannel:     make(chan error, 100), // Buffered channel to prevent blocking
	}
	if twitterConfig != nil {
		client, err := clients.NewTwitterClient(twitterConfig)
		if err != nil {
			panic(err)
		}
		cli.twitterClient = client
	}
	if discordConfig != nil && discordConfig.APIToken != "" {
		cli.discordBot = clients.NewDiscordBot(discordConfig.APIToken)
	}
	if telegramConfig != nil && telegramConfig.Token != "" {
		client, err := clients.NewTelegramClient(telegramConfig)
		if err != nil {
			panic(err)
		}
		cli.telegramBot = client
	}

	return cli
}

func (sc *SocialClientImpl) SendMessage(ctx context.Context, msg core.SocialMessage) error {
	switch msg.Platform {
	case "twitter":
		return sc.twitterClient.Tweet(ctx, msg.Content)
	case "discord":
		return sc.discordBot.SendMessage(ctx, &clients.DiscordMsg{
			AuthorID:  msg.FromUser,
			Content:   msg.Content,
			ChannelID: msg.Metadata["channel_id"].(string),
		})
	case "telegram":
		return sc.telegramBot.BroadcastMessage(ctx, msg.Content)
	case "all":
		// Send to all platforms
		var errs []error

		if sc.twitterClient != nil {
			if err := sc.twitterClient.Tweet(context.Background(), msg.Content); err != nil {
				errs = append(errs, fmt.Errorf("twitter: %w", err))
			}
		}

		if sc.discordBot != nil {
			if err := sc.discordBot.SendMessage(context.Background(), &clients.DiscordMsg{
				AuthorID:  msg.FromUser,
				Content:   msg.Content,
				ChannelID: msg.Metadata["channel_id"].(string),
			}); err != nil {
				errs = append(errs, fmt.Errorf("discord: %w", err))
			}
		}

		if sc.telegramBot != nil {
			if err := sc.telegramBot.BroadcastMessage(context.Background(), msg.Content); err != nil {
				errs = append(errs, fmt.Errorf("telegram: %w", err))
			}
		}

		// If any errors occurred, return them combined
		if len(errs) > 0 {
			return fmt.Errorf("failed to send to some platforms: %v", errs)
		}
	default:
		return fmt.Errorf("invalid platform: %s", msg.Platform)

	}

	return nil
}

func (sc *SocialClientImpl) GetMessageChannel() <-chan core.SocialMessage {
	return sc.socialMsgChannel
}

// GetErrorChannel returns the channel for monitoring errors
func (sc *SocialClientImpl) GetErrorChannel() <-chan error {
	return sc.errorChannel
}

// MonitorMessages starts monitoring messages from all configured platforms
func (sc *SocialClientImpl) MonitorMessages(ctx context.Context) {
	var wg sync.WaitGroup
	if sc.twitterClient != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sc.monitorTwitter(ctx)
		}()
	}

	if sc.discordBot != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sc.monitorDiscord(ctx)
		}()
	}

	if sc.telegramBot != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sc.monitorTelegram(ctx)
		}()
	}

	wg.Wait()
}

// monitorTwitter monitors Twitter mentions and reports errors through errorChannel
func (sc *SocialClientImpl) monitorTwitter(ctx context.Context) {
	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			tweets, err := sc.twitterClient.MonitorMentioned(context.Background())
			if err != nil {
				// Report error through channel and continue monitoring
				select {
				case sc.errorChannel <- fmt.Errorf("twitter monitor error: %w", err):
					// Error successfully reported
				default:
					// Channel is full, log locally
					fmt.Printf("Error channel full, dropping error: %v\n", err)
				}
				//not return here, continue monitoring
				continue
			}

			for _, tweet := range tweets {
				sc.socialMsgChannel <- core.SocialMessage{
					Type:        "mention",
					Content:     tweet.Text,
					Platform:    "twitter",
					FromUser:    tweet.UserID,
					TargetUsers: []string{sc.twitterClient.GetMe()},
				}
			}
		case <-ctx.Done():
			return
		}
	}
}

func (sc *SocialClientImpl) monitorDiscord(ctx context.Context) {
	channel := sc.discordBot.GetMessageChannel()

	for {
		select {
		case msg := <-channel:
			sc.socialMsgChannel <- core.SocialMessage{
				Type:     "message",
				Content:  msg.Content,
				Platform: "discord",
				FromUser: msg.AuthorID,
				Metadata: map[string]interface{}{"channel_id": msg.ChannelID},
			}
		case <-ctx.Done():
			return
		}
	}
}

// monitorTelegram monitors Telegram messages
func (sc *SocialClientImpl) monitorTelegram(ctx context.Context) {
	// Start the Telegram listener
	if err := sc.telegramBot.StartListener(ctx); err != nil {
		fmt.Printf("Failed to start Telegram listener: %v\n", err)
		return
	}

	// Get the message channel
	channel := sc.telegramBot.GetMessageChannel()

	// Monitor messages
	for {
		select {
		case msg := <-channel:
			// Convert TelegramMessage to core.SocialMessage
			socialMsg := core.SocialMessage{
				Type:     "message",
				Content:  msg.Text,
				Platform: "telegram",
				FromUser: msg.Username,
				Metadata: map[string]interface{}{
					"message_id": msg.MessageID,
					"chat_id":    msg.ChatID,
					"user_id":    msg.UserID,
					"is_command": msg.IsCommand,
					"command":    msg.Command,
					"reply_to":   msg.ReplyTo,
					"timestamp":  msg.Timestamp,
				},
			}

			// If it's a command, set the type accordingly
			if msg.IsCommand {
				socialMsg.Type = "command"
			}

			// Send to the social message channel
			sc.socialMsgChannel <- socialMsg

		case <-ctx.Done():
			// Context cancelled, stop monitoring
			return
		}
	}
}
