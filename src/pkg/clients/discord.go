package clients

import (
	"context"

	"github.com/bwmarrin/discordgo"
)

type DiscordMsg struct {
	AuthorID  string
	Content   string
	ChannelID string
}

type DiscordBot struct {
	session    *discordgo.Session
	msgChannel chan DiscordMsg
}

func NewDiscordBot(token string) *DiscordBot {
	discord, err := discordgo.New("Bot " + token)
	if err != nil {
		// TODO: handle error
		panic(err)
	}

	msgChannel := make(chan DiscordMsg)
	discord.AddHandler(MessageListener(msgChannel))
	discord.Open()

	return &DiscordBot{
		session:    discord,
		msgChannel: msgChannel,
	}
}

func (dc *DiscordBot) GetMessageChannel() <-chan DiscordMsg {
	return dc.msgChannel
}

func (dc *DiscordBot) SendMessage(
	ctx context.Context,
	msg *DiscordMsg,
) error {
	_, err := dc.session.ChannelMessageSend(msg.ChannelID, msg.Content)
	return err
}

func MessageListener(
	msgChannel chan<- DiscordMsg,
) func(*discordgo.Session, *discordgo.MessageCreate) {
	return func(discord *discordgo.Session, message *discordgo.MessageCreate) {
		channel, err := discord.Channel(message.ChannelID)
		if err != nil {
			return
		}

		if shouldReact(discord.State.User, channel, message) {
			msgChannel <- DiscordMsg{
				AuthorID:  message.Author.ID,
				Content:   message.Content,
				ChannelID: message.ChannelID,
			}
		}
	}
}

func shouldReact(
	me *discordgo.User,
	channel *discordgo.Channel,
	message *discordgo.MessageCreate,
) bool {
	/* prevent bot responding to its own message
	this is achived by looking into the message author id
	if message.author.id is same as bot.author.id then just return
	*/
	if message.Author.ID == me.ID {
		return false
	}

	/* always respond to direct messages */
	if channel.Type == discordgo.ChannelTypeDM {
		return true
	}

	/* check if bot was mentioned in the message */
	for _, mention := range message.Mentions {
		if mention.ID == me.ID {
			return true
		}
	}

	return false
}
