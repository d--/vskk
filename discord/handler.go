package discord

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"strings"
)

func CreateMessageHandler(channel, commandPhrase string, commands *BotCommands) interface{} {
	return func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.ID == s.State.User.ID {
			return
		}

		c, err := s.Channel(m.ChannelID)
		if err != nil {
			fmt.Println("error getting channel:", err)
			return
		}

		if c.Name == channel {
			dc := BotContext{
				Session:   s,
				ChannelId: m.ChannelID,
			}

			if strings.HasPrefix(m.Content, commandPhrase) {
				str := strings.TrimPrefix(m.Content, commandPhrase)
				str = strings.TrimSpace(str)
				ch, ok := commands.Get(str)
				if ok {
					ch <- dc
				}
			}
		}
	}
}
