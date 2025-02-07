package core

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/carv-protocol/d.a.t.a/src/pkg/clients"
)

type SocialClient interface {
	SendMessage(msg SocialMessage) error
	GetMessageChannel() chan SocialMessage
	MonitorMessages(ctx context.Context) error
}

// IntentType defines different types of intents
type IntentType string

const (
	IntentQuestion    IntentType = "question"
	IntentFeedback    IntentType = "feedback"
	IntentComplaint   IntentType = "complaint"
	IntentSuggestion  IntentType = "suggestion"
	IntentGreeting    IntentType = "greeting"
	IntentInquiry     IntentType = "inquiry"
	IntentRequest     IntentType = "request"
	IntentAcknowledge IntentType = "acknowledge"
)

// EntityType defines different types of entities
type EntityType string

const (
	EntityPerson   EntityType = "person"
	EntityProduct  EntityType = "product"
	EntityCompany  EntityType = "company"
	EntityLocation EntityType = "location"
	EntityDateTime EntityType = "datetime"
	EntityCrypto   EntityType = "crypto"
	EntityWallet   EntityType = "wallet"
	EntityContract EntityType = "contract"
)

// EmotionType defines different types of emotions
type EmotionType string

const (
	EmotionPositive EmotionType = "positive"
	EmotionNegative EmotionType = "negative"
	EmotionNeutral  EmotionType = "neutral"
)

type ProcessedMessage struct {
	Intent               IntentType  `json:"intent"`
	Entity               EntityType  `json:"entity"`
	Emotion              EmotionType `json:"emotion"`
	Confidence           float64     `json:"confidence"`
	ShouldReply          bool        `json:"should_reply"`
	ResponseMsg          string      `json:"response_msg"`
	ShouldGenerateTask   bool        `json:"should_generate_task"`
	ShouldGenerateAction bool        `json:"should_generate_action"`
}

type SocialMessage struct {
	Type        string
	Content     string
	Platform    string
	FromUser    string
	TargetUsers []string
	Metadata    map[string]interface{}
}

// SocialClientImpl handles social media interactions
type SocialClientImpl struct {
	twitterClient    *clients.TwitterClient
	discordBot       *clients.DiscordBot
	telegramBot      *clients.TelegramClient
	socialMsgChannel chan SocialMessage
}

func NewSocialClient(
	twitterConfig *clients.TwitterConfig,
	discordConfig *clients.DiscordConfig,
	telegramConfig *clients.TelegramConfig,
) *SocialClientImpl {
	cli := &SocialClientImpl{
		socialMsgChannel: make(chan SocialMessage),
	}
	if twitterConfig != nil {
		client, err := clients.NewTwitterClient(twitterConfig)
		if err != nil {
			panic(err)
		}
		cli.twitterClient = client
	}
	if discordConfig != nil {
		cli.discordBot = clients.NewDiscordBot(discordConfig.APIToken)
	}
	if telegramConfig != nil {
		client, err := clients.NewTelegramClient(telegramConfig)
		if err != nil {
			panic(err)
		}
		cli.telegramBot = client
	}

	return cli
}

func (sc *SocialClientImpl) SendMessage(msg SocialMessage) error {
	switch msg.Platform {
	case "twitter":
		return sc.twitterClient.Tweet(context.Background(), msg.Content)
	case "discord":
		return sc.discordBot.SendMessage(context.Background(), &clients.DiscordMsg{
			AuthorID:  msg.FromUser,
			Content:   msg.Content,
			ChannelID: msg.Metadata["channel_id"].(string),
		})
	case "telegram":
		return sc.telegramBot.BroadcastMessage(context.Background(), msg.Content)
	case "all":
		// Send to all platforms
		if err := sc.twitterClient.Tweet(context.Background(), msg.Content); err != nil {
			return err
		}
		// return sc.discordBot.SendMessage(msg.Content)
	}

	return nil
}

func (sc *SocialClientImpl) GetMessageChannel() chan SocialMessage {
	return sc.socialMsgChannel
}

// MonitorMessages starts monitoring messages from all configured platforms
func (sc *SocialClientImpl) MonitorMessages(ctx context.Context) error {
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
	return nil
}

func (sc *SocialClientImpl) monitorTwitter(ctx context.Context) {
	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()

	// fmt.Println("try to monitor twitter")

	for {
		select {
		case <-ticker.C:
			tweets, err := sc.twitterClient.MonitorMentioned(context.Background())
			if err != nil {
				// TODO: handle error
				return
			}

			for _, tweet := range tweets {
				sc.socialMsgChannel <- SocialMessage{
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
			sc.socialMsgChannel <- SocialMessage{
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
			// Convert TelegramMessage to SocialMessage
			socialMsg := SocialMessage{
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

func parseAnalysis(response string) (*ProcessedMessage, error) {
	startTag := "```json\n"
	endTag := "\n```"

	startIndex := strings.Index(response, startTag)
	if startIndex == -1 {
		return nil, fmt.Errorf("start tag not found")
	}
	startIndex += len(startTag)

	endIndex := strings.Index(response[startIndex:], endTag)
	if endIndex == -1 {
		return nil, fmt.Errorf("end tag not found")
	}
	endIndex += startIndex

	jsonContent := response[startIndex:endIndex]

	var processedMsg ProcessedMessage
	if err := json.Unmarshal([]byte(jsonContent), &processedMsg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	return &processedMsg, nil
}
