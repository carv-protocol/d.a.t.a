package clients

import (
	"context"
	"fmt"
	"time"

	telegram "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// TelegramConfig holds the configuration for Telegram client
type TelegramConfig struct {
	Token     string `mapstructure:"bot_token"`  // Bot token from BotFather
	ChannelID int64  `mapstructure:"channel_id"` // Default channel ID for broadcasts
	Debug     bool   `mapstructure:"debug"`      // Enable debug mode
}

// TelegramMessage represents a message structure
type TelegramMessage struct {
	MessageID int64
	ChatID    int64
	UserID    int64
	Username  string
	Text      string
	IsCommand bool
	Command   string
	ReplyTo   int64
	Timestamp time.Time
}

// TelegramClient represents a Telegram bot client
type TelegramClient struct {
	bot     *telegram.BotAPI
	config  TelegramConfig
	msgChan chan TelegramMessage // msg channel
}

// NewTelegramClient creates a new Telegram client instance
func NewTelegramClient(config *TelegramConfig) (*TelegramClient, error) {
	bot, err := telegram.NewBotAPI(config.Token)
	if err != nil {
		return nil, fmt.Errorf("failed to create telegram bot: %w", err)
	}

	bot.Debug = config.Debug

	client := &TelegramClient{
		bot:     bot,
		config:  *config,
		msgChan: make(chan TelegramMessage),
	}

	return client, nil
}

// StartListener starts listening for incoming messages
func (c *TelegramClient) StartListener(ctx context.Context) error {
	u := telegram.NewUpdate(0)
	u.Timeout = 60

	updates := c.bot.GetUpdatesChan(u)

	go func() {
		for {
			select {
			case update := <-updates:
				if update.Message != nil {
					// Get ReplyToMessageID safely
					var replyToID int64
					if update.Message.ReplyToMessage != nil {
						replyToID = int64(update.Message.ReplyToMessage.MessageID)
					}

					msg := TelegramMessage{
						MessageID: int64(update.Message.MessageID),
						ChatID:    update.Message.Chat.ID,
						UserID:    int64(update.Message.From.ID),
						Username:  update.Message.From.UserName,
						Text:      update.Message.Text,
						IsCommand: update.Message.IsCommand(),
						Command:   update.Message.Command(),
						ReplyTo:   replyToID,
						Timestamp: time.Now(),
					}
					c.msgChan <- msg
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}

// GetMessageChannel returns channel for receiving messages
func (c *TelegramClient) GetMessageChannel() <-chan TelegramMessage {
	return c.msgChan
}

// SendMessage sends a message to specified chat
func (c *TelegramClient) SendMessage(ctx context.Context, chatID int64, text string) error {
	msg := telegram.NewMessage(chatID, text)
	msg.ParseMode = "HTML"

	_, err := c.bot.Send(msg)
	if err != nil {
		return fmt.Errorf("failed to send telegram message: %w", err)
	}

	return nil
}

// SendReply sends a reply to a specific message
func (c *TelegramClient) SendReply(ctx context.Context, chatID int64, replyToID int64, text string) error {
	msg := telegram.NewMessage(chatID, text)
	msg.ReplyToMessageID = int(replyToID)
	msg.ParseMode = "HTML"

	_, err := c.bot.Send(msg)
	if err != nil {
		return fmt.Errorf("failed to send reply message: %w", err)
	}

	return nil
}

// BroadcastMessage sends a message to the default channel
func (c *TelegramClient) BroadcastMessage(ctx context.Context, text string) error {
	return c.SendMessage(ctx, c.config.ChannelID, text)
}

// SendPhoto sends a photo with optional caption
func (c *TelegramClient) SendPhoto(ctx context.Context, chatID int64, photoPath string, caption string) error {
	photo := telegram.NewPhoto(chatID, telegram.FilePath(photoPath))
	photo.Caption = caption

	_, err := c.bot.Send(photo)
	if err != nil {
		return fmt.Errorf("failed to send photo: %w", err)
	}

	return nil
}

// SendDocument sends a document file
func (c *TelegramClient) SendDocument(ctx context.Context, chatID int64, filePath string, caption string) error {
	doc := telegram.NewDocument(chatID, telegram.FilePath(filePath))
	doc.Caption = caption

	_, err := c.bot.Send(doc)
	if err != nil {
		return fmt.Errorf("failed to send document: %w", err)
	}

	return nil
}

// HandleCommand registers a command handler
func (c *TelegramClient) HandleCommand(command string, handler func(TelegramMessage) error) {
	go func() {
		for msg := range c.msgChan {
			if msg.IsCommand && msg.Command == command {
				if err := handler(msg); err != nil {
					fmt.Printf("Error handling command %s: %v\n", command, err)
				}
			}
		}
	}()
}

// GetChatMember gets information about a chat member
func (c *TelegramClient) GetChatMember(chatID int64, userID int64) (*telegram.ChatMember, error) {
	var config = telegram.GetChatMemberConfig{
		ChatConfigWithUser: telegram.ChatConfigWithUser{
			ChatID: chatID,
			UserID: userID,
		},
	}

	member, err := c.bot.GetChatMember(config)
	if err != nil {
		return nil, fmt.Errorf("failed to get chat member: %w", err)
	}

	return &member, nil
}

// DeleteMessage deletes a message
func (c *TelegramClient) DeleteMessage(chatID int64, messageID int64) error {
	deleteConfig := telegram.NewDeleteMessage(chatID, int(messageID))

	_, err := c.bot.Request(deleteConfig)
	if err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}

	return nil
}
