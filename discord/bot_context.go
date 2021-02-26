package discord

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
)

type BotContext struct {
	Session   *discordgo.Session
	ChannelId string
}

func (c *BotContext) SendMessage(m string) error {
	_, err := c.Session.ChannelMessageSend(c.ChannelId, m)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	return nil
}
